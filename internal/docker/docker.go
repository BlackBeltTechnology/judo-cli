package docker

import (
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
		log.Fatalf("Failed to create Docker client: %v", err)
	}
}

func GetDockerClient() *client.Client {
	return cli
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

func pullImage(imageName string) {
	reader, err := cli.ImagePull(context.Background(), imageName, image.PullOptions{})
	if err != nil {
		log.Fatalf("Failed to pull image %s: %v", imageName, err)
	}
	defer reader.Close()
	io.Copy(os.Stdout, reader)
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
func CreateDockerNetwork(name string) {
	networks, err := cli.NetworkList(context.Background(), network.ListOptions{})
	if err != nil {
		log.Fatalf("Failed to list Docker networks: %v", err)
	}
	for _, network := range networks {
		if network.Name == name {
			return
		}
	}
	_, err = cli.NetworkCreate(context.Background(), name, network.CreateOptions{})
	if err != nil {
		log.Fatalf("Failed to create Docker network: %v", err)
	}
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
	containers, err := cli.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		log.Fatalf("Failed to list Docker containers: %v", err)
	}
	for _, c := range containers {
		for _, n := range c.Names {
			if strings.TrimPrefix(n, "/") == name {
				return true
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

func ContainerExists(name string) bool {
	containers, err := cli.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		log.Fatalf("Failed to list Docker containers: %v", err)
	}
	for _, c := range containers {
		for _, n := range c.Names {
			if strings.TrimPrefix(n, "/") == name {
				return true
			}
		}
	}
	return false
}

func StartContainer(name string) {
	if err := cli.ContainerStart(context.Background(), name, container.StartOptions{}); err != nil {
		log.Fatalf("Failed to start container %s: %v", name, err)
	}
}

func StartCompose() {
	cfg := config.GetConfig()
	fmt.Println("Starting Docker compose environment...")
	cmd := utils.ExecuteCommand("docker", "compose", "-f", fmt.Sprintf("%s/docker/%s/docker-compose.yml", cfg.ModelDir, cfg.ComposeEnv), "up")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	utils.CheckError(cmd.Run())
}

func StartPostgres() {
	cfg := config.GetConfig()
	fmt.Println("Starting PostgreSQL...")
	name := "postgres-" + cfg.SchemaName
	image := "postgres:16.2"

	if !ContainerExists(name) {
		pullImage(image)
		CreateDockerNetwork(cfg.AppName)
		resp, err := cli.ContainerCreate(context.Background(), &container.Config{
			Image: image,
			Env: []string{
				"PGDATA=/var/lib/postgresql/pgdata",
				fmt.Sprintf("POSTGRES_USER=%s", cfg.SchemaName),
				fmt.Sprintf("POSTGRES_PASSWORD=%s", cfg.SchemaName),
			},
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
		}, nil, nil, name)
		if err != nil {
			log.Fatalf("Failed to create PostgreSQL container: %v", err)
		}
		if err := cli.ContainerStart(context.Background(), resp.ID, container.StartOptions{}); err != nil {
			log.Fatalf("Failed to start PostgreSQL container: %v", err)
		}
	} else {
		StartContainer(name)
	}
	utils.WaitForPort("localhost", cfg.PostgresPort, 30*utils.TimeSecond)
}

func StartKeycloak() {
	cfg := config.GetConfig()
	fmt.Println("Starting Keycloak...")
	name := "keycloak-" + cfg.KeycloakName
	image := "quay.io/keycloak/keycloak:23.0"

	if !ContainerExists(name) {
		pullImage(image)
		if cfg.DBType == "postgresql" {
			CreateDockerNetwork(cfg.AppName)
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
		}, &container.HostConfig{
			NetworkMode: container.NetworkMode(cfg.AppName),
			PortBindings: nat.PortMap{
				nat.Port(fmt.Sprintf("%d/tcp", cfg.KeycloakPort)): []nat.PortBinding{
					{HostIP: "0.0.0.0", HostPort: fmt.Sprintf("%d", cfg.KeycloakPort)},
				},
			},
		}, nil, nil, name)
		if err != nil {
			log.Fatalf("Failed to create Keycloak container: %v", err)
		}
		if err := cli.ContainerStart(context.Background(), resp.ID, container.StartOptions{}); err != nil {
			log.Fatalf("Failed to start Keycloak container: %v", err)
		}
		// Verify the container is running
		time.Sleep(2 * time.Second) // Give it a moment to stabilize
		if !DockerInstanceRunning(name) {
			// If it's not running, get the logs to see why it failed.
			reader, err := cli.ContainerLogs(context.Background(), name, container.LogsOptions{ShowStdout: true, ShowStderr: true})
			if err != nil {
				log.Fatalf("Failed to get Keycloak container logs: %v", err)
			}
			defer reader.Close()
			logs, _ := io.ReadAll(reader)
			log.Fatalf("Keycloak container failed to start. Logs:\n%s", string(logs))
		}
	} else {
		StartContainer(name)
	}
	utils.WaitForPort("localhost", cfg.KeycloakPort, 30*utils.TimeSecond)
}
