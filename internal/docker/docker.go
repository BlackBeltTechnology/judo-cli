package docker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"judo-cli-module/internal/config"
	"judo-cli-module/internal/utils"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

var cli *client.Client

func init() {
	var err error
	//cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	cli, err = newDockerClient()
	if err != nil {
		// Don't fatal here - just log and continue
		log.Printf("Warning: Failed to create Docker client: %v", err)
	}
}

func GetDockerClient() *client.Client {
	return cli
}

// CloseDockerClient closes the global Docker client if it exists
func CloseDockerClient() {
	if cli != nil {
		cli.Close()
	}
}

func newDockerClient() (*client.Client, error) {
	// 1) Respect env (DOCKER_HOST etc.)
	if cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation()); err == nil {
		if pingOK(cli) {
			return cli, nil
		}
	}

	home, _ := os.UserHomeDir()
	candidates := []string{
		"unix:///var/run/docker.sock",                     // Linux / some Desktop setups
		"unix://" + home + "/.docker/run/docker.sock",     // Docker Desktop macOS
		"unix://" + home + "/.colima/default/docker.sock", // Colima
	}

	for _, h := range candidates {
		cli, err := client.NewClientWithOpts(client.WithHost(h), client.WithAPIVersionNegotiation())
		if err != nil {
			continue
		}
		if pingOK(cli) {
			return cli, nil
		}
		_ = cli.Close()
	}
	return nil, errors.New("could not reach Docker daemon; set DOCKER_HOST or start Docker Desktop/Colima")
}

func pingOK(cli *client.Client) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := cli.Ping(ctx)
	return err == nil
}

func pullImage(imageName string) error {
	reader, err := cli.ImagePull(context.Background(), imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %v", imageName, err)
	}
	defer reader.Close()
	io.Copy(os.Stdout, reader)
	return nil
}

// IsDockerRunning checks if the Docker daemon is responsive.
func IsDockerRunning() bool {
	_, err := cli.Ping(context.Background())
	return err == nil
}

func GetComposeEnvs(cfg *config.Config) []string {
	root := filepath.Join(cfg.AppDir, "docker")
	var envs []string
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if filepath.Base(path) == "docker-compose.yml" {
			envs = append(envs, filepath.Base(filepath.Dir(path)))
		}
		return nil
	})
	return envs
}

func StopCompose(cfg *config.Config, env string) error {
	composeFile := filepath.Join(cfg.AppDir, "docker", env, "docker-compose.yml")
	cmd := utils.ExecuteCommand("docker", "compose", "-f", composeFile, "down", "--volumes")
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

func RemoveDockerInstance(name string) error {
	if name == "" {
		return nil
	}
	return cli.ContainerRemove(context.Background(), name, container.RemoveOptions{Force: true})
}
func CreateDockerNetwork(name string) error {
	networks, err := cli.NetworkList(context.Background(), network.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list Docker networks: %v", err)
	}
	for _, network := range networks {
		if network.Name == name {
			return nil
		}
	}
	_, err = cli.NetworkCreate(context.Background(), name, network.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create Docker network: %v", err)
	}
	return nil
}

func RemoveDockerNetwork(name string) error {
	if name == "" {
		return nil
	}
	return cli.NetworkRemove(context.Background(), name)
}

func RemoveDockerVolume(name string) error {
	if name == "" {
		return nil
	}
	return cli.VolumeRemove(context.Background(), name, true)
}

func DockerVolumeExists(name string) bool {
	if name == "" {
		return false
	}
	_, err := cli.VolumeInspect(context.Background(), name)
	return err == nil
}

// Docker stop helper (no-op if not running)
func DockerInstanceRunning(name string) bool {
	// Create a fresh Docker client to avoid stale connections
	client, err := newDockerClient()
	if err != nil {
		log.Printf("Failed to create Docker client: %v", err)
		return false
	}
	defer client.Close()

	containers, err := client.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		log.Printf("Failed to list Docker containers: %v", err)
		return false
	}
	for _, c := range containers {
		for _, n := range c.Names {
			if strings.TrimPrefix(n, "/") == name {
				// Check if container is actually running (not just exists)
				return strings.HasPrefix(c.State, "running")
			}
		}
	}
	return false
}

func StopDockerInstance(name string) error {
	if DockerInstanceRunning(name) {
		return cli.ContainerStop(context.Background(), name, container.StopOptions{})
	}
	return nil
}

func ContainerExists(name string) (bool, error) {
	containers, err := cli.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return false, fmt.Errorf("failed to list Docker containers: %v", err)
	}
	for _, c := range containers {
		for _, n := range c.Names {
			if strings.TrimPrefix(n, "/") == name {
				return true, nil
			}
		}
	}
	return false, nil
}

func StartContainer(name string) error {
	if err := cli.ContainerStart(context.Background(), name, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container %s: %v", name, err)
	}
	return nil
}

func StartCompose() {
	cfg := config.GetConfig()
	fmt.Println("Starting Docker compose environment...")
	cmd := utils.ExecuteCommand("docker", "compose", "-f", fmt.Sprintf("%s/docker/%s/docker-compose.yml", cfg.ModelDir, cfg.ComposeEnv), "up")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	utils.CheckError(cmd.Run())
}

func StartPostgres() error {
	cfg := config.GetConfig()
	fmt.Println("Starting PostgreSQL...")
	name := "postgres-" + cfg.SchemaName
	image := "postgres:16.2"

	exists, err := ContainerExists(name)
	if err != nil {
		return fmt.Errorf("failed to check if container exists: %v", err)
	}
	if !exists {
		if err := pullImage(image); err != nil {
			return fmt.Errorf("failed to pull image: %v", err)
		}
		if err := CreateDockerNetwork(cfg.AppName); err != nil {
			return fmt.Errorf("failed to create network: %v", err)
		}
		resp, err := cli.ContainerCreate(context.Background(), &container.Config{
			Image: image,
			Env: []string{
				"PGDATA=/var/lib/postgresql/pgdata",
				fmt.Sprintf("POSTGRES_USER=%s", cfg.SchemaName),
				fmt.Sprintf("POSTGRES_PASSWORD=%s", cfg.SchemaName),
			},
			AttachStdin:  false,
			AttachStdout: false,
			AttachStderr: false,
			Tty:          false,
		}, &container.HostConfig{
			Binds: []string{
				fmt.Sprintf("%s_postgresql_db:/var/lib/postgresql/pgdata", cfg.SchemaName),
				fmt.Sprintf("%s_postgresql_data:/var/lib/postgresql/data", cfg.SchemaName),
			},
			NetworkMode: container.NetworkMode(cfg.AppName),
			PortBindings: nat.PortMap{
				"5432/tcp": []nat.PortBinding{
					{
						HostIP:   "0.0.0.0",
						HostPort: fmt.Sprintf("%d", cfg.PostgresPort),
					},
				},
			},
			RestartPolicy: container.RestartPolicy{
				Name: "unless-stopped",
			},
		}, &network.NetworkingConfig{}, nil, name)
		if err != nil {
			return fmt.Errorf("failed to create PostgreSQL container: %v", err)
		}
		if err := cli.ContainerStart(context.Background(), resp.ID, container.StartOptions{
			// Ensure the container starts detached
		}); err != nil {
			return fmt.Errorf("failed to start PostgreSQL container: %v", err)
		}
	} else {
		if err := StartContainer(name); err != nil {
			return fmt.Errorf("failed to start existing container: %v", err)
		}
	}
	utils.WaitForPort("localhost", cfg.PostgresPort, 30*utils.TimeSecond)
	return nil
}

func StartKeycloak() error {
	cfg := config.GetConfig()
	fmt.Println("Starting Keycloak...")
	name := "keycloak-" + cfg.KeycloakName
	image := "quay.io/keycloak/keycloak:23.0"

	exists, err := ContainerExists(name)
	if err != nil {
		return fmt.Errorf("failed to check if container exists: %v", err)
	}
	if !exists {
		if err := pullImage(image); err != nil {
			return fmt.Errorf("failed to pull image: %v", err)
		}
		if cfg.DBType == "postgresql" {
			if err := CreateDockerNetwork(cfg.AppName); err != nil {
				return fmt.Errorf("failed to create network: %v", err)
			}
		}
		env := []string{
			"KEYCLOAK_ADMIN=admin",
			"KEYCLOAK_ADMIN_PASSWORD=judo",
		}
		if cfg.DBType == "postgresql" {
			env = append(env,
				"KC_DB=postgres",
				"KC_DB_URL_HOST=postgres-"+cfg.SchemaName,
				"KC_DB_URL_DATABASE="+cfg.SchemaName,
				"KC_DB_PASSWORD="+cfg.SchemaName,
				"KC_DB_USERNAME="+cfg.SchemaName,
				"KC_DB_SCHEMA=public",
			)
		}
		resp, err := cli.ContainerCreate(context.Background(), &container.Config{
			Image: image,
			Env:   env,
			Cmd: []string{
				"start-dev",
				fmt.Sprintf("--http-port=%d", cfg.KeycloakPort),
				"--http-relative-path", "/auth",
			},
			AttachStdin:  false,
			AttachStdout: false,
			AttachStderr: false,
			Tty:          false,
		}, &container.HostConfig{
			NetworkMode: container.NetworkMode(cfg.AppName),
			PortBindings: nat.PortMap{
				nat.Port(fmt.Sprintf("%d/tcp", cfg.KeycloakPort)): []nat.PortBinding{
					{HostIP: "0.0.0.0", HostPort: fmt.Sprintf("%d", cfg.KeycloakPort)},
				},
			},
			RestartPolicy: container.RestartPolicy{
				Name: "unless-stopped",
			},
		}, &network.NetworkingConfig{}, nil, name)
		if err != nil {
			return fmt.Errorf("failed to create Keycloak container: %v", err)
		}
		if err := cli.ContainerStart(context.Background(), resp.ID, container.StartOptions{
			// Ensure the container starts detached
		}); err != nil {
			return fmt.Errorf("failed to start Keycloak container: %v", err)
		}
		// Verify the container is running
		time.Sleep(2 * time.Second) // Give it a moment to stabilize
		if !DockerInstanceRunning(name) {
			// If it's not running, get the logs to see why it failed.
			reader, err := cli.ContainerLogs(context.Background(), name, container.LogsOptions{ShowStdout: true, ShowStderr: true})
			if err != nil {
				return fmt.Errorf("failed to get Keycloak container logs: %v", err)
			}
			defer reader.Close()
			logs, _ := io.ReadAll(reader)
			return fmt.Errorf("Keycloak container failed to start. Logs:\n%s", string(logs))
		}
	} else {
		if err := StartContainer(name); err != nil {
			return fmt.Errorf("failed to start existing container: %v", err)
		}
	}
	utils.WaitForPort("localhost", cfg.KeycloakPort, 30*utils.TimeSecond)
	return nil
}

// IsPortUsedByKeycloak checks if a port is being used by the current Keycloak Docker container
func IsPortUsedByKeycloak(port int) bool {
	cfg := config.GetConfig()
	if cfg == nil {
		return false
	}

	keycloakName := "keycloak-" + cfg.KeycloakName

	// List all running containers
	containers, err := cli.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		return false
	}

	// Find the Keycloak container
	for _, c := range containers {
		for _, n := range c.Names {
			if strings.TrimPrefix(n, "/") == keycloakName {
				// Check if this container is using the specified port
				return isContainerUsingPort(c, port)
			}
		}
	}

	return false
}

// isContainerUsingPort checks if a container is using a specific host port
func isContainerUsingPort(container types.Container, port int) bool {
	for _, p := range container.Ports {
		if int(p.PublicPort) == port && p.Type == "tcp" {
			return true
		}
	}
	return false
}

// IsPortUsedByPostgres checks if a port is being used by the current PostgreSQL Docker container
func IsPortUsedByPostgres(port int) bool {
	cfg := config.GetConfig()
	if cfg == nil {
		return false
	}

	if cfg.DBType != "postgresql" {
		return false
	}

	postgresName := "postgres-" + cfg.SchemaName

	// List all running containers
	containers, err := cli.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		return false
	}

	// Find the PostgreSQL container
	for _, c := range containers {
		for _, n := range c.Names {
			if strings.TrimPrefix(n, "/") == postgresName {
				// Check if this container is using the specified port
				return isContainerUsingPort(c, port)
			}
		}
	}

	return false
}

// StreamContainerLogs streams logs from a Docker container in real-time
func StreamContainerLogs(containerName string, ctx context.Context, logChan chan<- string) error {
	if cli == nil {
		return errors.New("Docker client not initialized")
	}

	// Check if container exists
	exists, err := ContainerExists(containerName)
	if err != nil {
		return fmt.Errorf("failed to check container existence: %v", err)
	}
	if !exists {
		return fmt.Errorf("container %s does not exist", containerName)
	}

	// Stream logs with follow option
	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       "100",                                                 // Last 100 lines initially
		Since:      time.Now().Add(-5 * time.Minute).Format(time.RFC3339), // Last 5 minutes
	}

	reader, err := cli.ContainerLogs(ctx, containerName, options)
	if err != nil {
		return fmt.Errorf("failed to get container logs: %v", err)
	}
	defer reader.Close()

	// Stream logs line by line
	buffer := make([]byte, 4096)
	var partialLine []byte

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			n, err := reader.Read(buffer)
			if err != nil {
				if err == io.EOF {
					// Container might have stopped, wait a bit and try to reconnect
					time.Sleep(1 * time.Second)
					continue
				}
				return fmt.Errorf("error reading logs: %v", err)
			}

			if n > 0 {
				data := buffer[:n]
				lines := bytes.Split(data, []byte("\n"))

				// Handle partial lines
				if len(partialLine) > 0 {
					lines[0] = append(partialLine, lines[0]...)
					partialLine = nil
				}

				// If last line doesn't end with newline, it's partial
				if len(lines) > 0 && !bytes.HasSuffix(data, []byte("\n")) {
					partialLine = lines[len(lines)-1]
					lines = lines[:len(lines)-1]
				}

				// Send complete lines to channel
				for _, line := range lines {
					if len(line) > 0 {
						// Docker logs include 8-byte header, skip it for clean output
						if len(line) > 8 {
							logChan <- string(line[8:])
						} else {
							logChan <- string(line)
						}
					}
				}
			}
		}
	}
}
