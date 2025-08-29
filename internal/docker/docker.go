package docker

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"judo-cli-module/internal/config"
	"judo-cli-module/internal/utils"
)

// IsDockerRunning checks if the Docker daemon is responsive.
func IsDockerRunning() bool {
	cmd := exec.Command("docker", "info")
	err := cmd.Run()
	return err == nil
}

func GetComposeEnvs(cfg *config.Config) []string {
	root := filepath.Join(cfg.AppDir, "docker")
	envs := []string{}
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
	cmd := utils.ExecuteCommand("docker", "rm", "-f", name)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	_ = cmd.Run() // ignore if it doesn't exist
	return nil
}
func CreateDockerNetwork(name string) {
	out, _ := utils.RunCapture("docker", "network", "ls", "--format", "{{.Name}}")
	for _, n := range strings.Split(out, "\n") {
		if n == name {
			return
		}
	}
	_ = utils.Run("docker", "network", "create", name)
}

func RemoveDockerNetwork(name string) error {
	if name == "" {
		return nil
	}
	cmd := utils.ExecuteCommand("docker", "network", "rm", name)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	_ = cmd.Run() // ignore if it doesn't exist
	return nil
}

func RemoveDockerVolume(name string) error {
	if name == "" {
		return nil
	}
	cmd := utils.ExecuteCommand("docker", "volume", "rm", name)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	_ = cmd.Run() // ignore if it doesn't exist
	return nil
}

func DockerVolumeExists(name string) bool {
	if name == "" {
		return false
	}
	out, _ := utils.RunCapture("docker", "volume", "ls", "--format", "{{.Name}}")
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == name {
			return true
		}
	}
	return false
}

// Docker stop helper (no-op if not running)
func DockerInstanceRunning(name string) bool {
	out, _ := utils.RunCapture("docker", "ps", "--format", "{{.Names}}")
	for _, n := range strings.Split(out, "\n") {
		if n == name {
			return true
		}
	}
	return false
}

func StopDockerInstance(name string) error {
	if DockerInstanceRunning(name) {
		return utils.Run("docker", "stop", name)
	}
	return nil
}

func ContainerExists(name string) bool {
	cmd := utils.ExecuteCommand("docker", "ps", "-a", "-f", fmt.Sprintf("name=%s", name))
	output, _ := cmd.Output()
	return strings.Contains(string(output), name)
}

func StartContainer(name string) {
	utils.CheckError(utils.ExecuteCommand("docker", "start", name).Run())
}

func StartCompose() {
	fmt.Println("Starting Docker compose environment...")
	cmd := utils.ExecuteCommand("docker", "compose", "-f", fmt.Sprintf("%s/docker/%s/docker-compose.yml", config.ModelDir, config.ComposeEnv), "up")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	utils.CheckError(cmd.Run())
}

func StartPostgres() {
	fmt.Println("Starting PostgreSQL...")
	name := "postgres-" + config.SchemaName

	if !ContainerExists(name) {
		CreateDockerNetwork(config.AppName)
		cmd := utils.ExecuteCommand(
			"docker", "run", "-d",
			"-v", fmt.Sprintf("%s_postgresql_db:/var/lib/postgresql/pgdata", config.SchemaName),
			"-v", fmt.Sprintf("%s_postgresql_data:/var/lib/postgresql/data", config.SchemaName),
			"--network", config.AppName,
			"--name", name,
			"-e", "PGDATA=/var/lib/postgresql/pgdata",
			"-e", fmt.Sprintf("POSTGRES_USER=%s", config.SchemaName),
			"-e", fmt.Sprintf("POSTGRES_PASSWORD=%s", config.SchemaName),
			"-p", fmt.Sprintf("%d:5432", config.PostgresPort),
			"postgres:16.2",
		)
		if err := cmd.Run(); err != nil {
			log.Fatalf("Failed to start PostgreSQL container: %v", err)
		}
	} else {
		StartContainer(name)
	}
	utils.WaitForPort("localhost", config.PostgresPort, 30*utils.TimeSecond)
}

func StartKeycloak() {
	fmt.Println("Starting Keycloak...")
	name := "keycloak-" + config.KeycloakName

	if !ContainerExists(name) {
		if config.DBType == "postgresql" {
			CreateDockerNetwork(config.AppName)
		}
		args := []string{
			"run", "-d",
			"--name", name,
			"-e", "KEYCLOAK_ADMIN=admin",
			"-e", "KEYCLOAK_ADMIN_PASSWORD=judo",
			"-p", fmt.Sprintf("%d:%d", config.KeycloakPort, config.KeycloakPort),
		}
		// DB wiring like in the bash script
		if config.DBType == "postgresql" {
			args = append(args,
				"--network", config.AppName,
				"-e", "KC_DB=postgres",
				"-e", "KC_DB_URL_HOST=postgres-"+config.SchemaName,
				"-e", "KC_DB_URL_DATABASE="+config.SchemaName,
				"-e", "KC_DB_PASSWORD="+config.SchemaName,
				"-e", "KC_DB_USERNAME="+config.SchemaName,
				"-e", "KC_DB_SCHEMA=public",
			)
		}
		args = append(args,
			"quay.io/keycloak/keycloak:23.0",
			"start-dev",
			fmt.Sprintf("--http-port=%d", config.KeycloakPort),
			"--http-relative-path", "/auth",
		)
		cmd := utils.ExecuteCommand("docker", args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatalf("Failed to start Keycloak container: %v\nOutput: %s", err, string(output))
		}

		// Verify the container is running
		time.Sleep(2 * time.Second) // Give it a moment to stabilize
		if !DockerInstanceRunning(name) {
			// If it's not running, get the logs to see why it failed.
			logsCmd := utils.ExecuteCommand("docker", "logs", name)
			logs, _ := logsCmd.CombinedOutput()
			log.Fatalf("Keycloak container failed to start. Logs:\n%s", string(logs))
		}
	} else {
		StartContainer(name)
	}
	utils.WaitForPort("localhost", config.KeycloakPort, 30*utils.TimeSecond)
}
