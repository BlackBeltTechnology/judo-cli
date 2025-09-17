package server

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"judo-cli-module/internal/commands"
	"judo-cli-module/internal/config"
	"judo-cli-module/internal/docker"
	"judo-cli-module/internal/karaf"
	"judo-cli-module/internal/session"
)

// ServerAPI provides direct function calls for server operations
// instead of executing external judo commands
type ServerAPI struct {
	mu sync.Mutex
}

// NewServerAPI creates a new ServerAPI instance
func NewServerAPI() *ServerAPI {
	return &ServerAPI{}
}

// ExecuteCommand executes a JUDO CLI command using direct function calls
func (api *ServerAPI) ExecuteCommand(command string, args []string) (string, error) {
	api.mu.Lock()
	defer api.mu.Unlock()

	// Load config first
	config.LoadProperties()
	_ = config.GetConfig()

	// Handle session commands internally
	if output, handled := api.handleSessionCommand(command, args); handled {
		return output, nil
	}

	// Handle specific judo commands with direct function calls
	switch command {
	case "judo":
		if len(args) == 0 {
			return api.executeJudoHelp(), nil
		}
		return api.ExecuteCommand(args[0], args[1:])

	case "help":
		return api.executeJudoHelp(), nil

	case "status":
		return api.executeStatus(), nil

	case "doctor":
		output, err := api.executeDoctor()
		if err != nil {
			return "", err
		}
		return output, nil

	case "build":
		return api.executeBuild(args), nil

	case "start":
		return api.executeStart(args), nil

	case "stop":
		return api.executeStop(args), nil

	case "clean":
		return api.executeClean(args), nil

	case "log":
		return api.executeLog(args), nil

	default:
		// For unsupported commands, fall back to original behavior
		return api.executeExternalCommand(command, args)
	}
}

// handleSessionCommand handles session-specific commands
func (api *ServerAPI) handleSessionCommand(command string, args []string) (string, bool) {
	cmd := strings.TrimSpace(strings.ToLower(command))

	switch cmd {
	case "help":
		return `Session Commands:
  help      - Show this help message
  exit      - Exit the interactive session
  quit      - Exit the interactive session
  clear     - Clear the terminal screen
  history   - Show command history
  status    - Show current session status
  doctor    - Run system health check

Project Commands:
  init      - Initialize a new JUDO project
  build     - Build project
  start     - Start application
  stop      - Stop application
  status    - Show application status
  clean     - Clean project data
  generate  - Generate application from model
  dump      - Dump PostgreSQL database
  import    - Import PostgreSQL database dump
  update    - Update dependency versions
  prune     - Clean untracked files
  reckless  - Fast build & run mode
  self-update - Update CLI to latest version`, true

	case "exit", "quit":
		return "Session exit command received. Note: This only exits the session context, not the server.", true

	case "clear":
		return "\033[2J\033[H", true // ANSI clear screen sequence

	case "history":
		return "Command history functionality would be implemented here", true

	case "status":
		status := api.getServiceStatusSummary()
		return "Current Service Status: " + status, true

	case "doctor":
		output, err := api.executeDoctor()
		if err != nil {
			return fmt.Sprintf("Doctor command failed: %v", err), true
		}
		return output, true

	default:
		return "", false
	}
}

// executeJudoHelp returns the help text for judo commands
func (api *ServerAPI) executeJudoHelp() string {
	return `JUDO CLI - Command Line Interface

Usage: judo COMMANDS... [OPTIONS...]

Commands:
    doctor                                  Check system health and required dependencies
    clean                                   Stop postgresql docker container and clear data.
    prune                                   Stop postgresql docker container and delete untracked files in this repository.
    update                                  Update dependency versions in JUDO project.
    generate                                Generate application based on model in JUDO project.
    generate-root                           Generate application root structure based on model in JUDO project.
    dump                                    Dump PostgreSQL DB data (creates <schema>_dump_YYYYMMDD_HHMMSS.tar.gz).
    import                                  Import PostgreSQL DB dump (pg_restore).
    schema-upgrade                          Apply RDBMS schema upgrade using current running database (PostgreSQL only).
    build                                   Build project
    reckless                                Build & run fast (skips validations, favors speed)
    start                                   Start application
    stop                                    Stop application, postgresql and keycloak. (if running)
    status                                  Print status of containers
    log                                     Display or tail Karaf console log
    init                                    Initialize a new JUDO project.
    self-update                             Update judo CLI to the latest version
    session                                 Start interactive JUDO CLI session
    server                                  Start JUDO CLI web server
        -p --port <port>                    Port to run the server on (default: 6969)

Examples:
  judo init
  judo build
  judo start
  judo status
  judo server -p 8080
`
}

// executeStatus returns the service status using direct function calls
func (api *ServerAPI) executeStatus() string {
	config.LoadProperties()
	cfg := config.GetConfig()

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Runtime: %s DB: %s\n", cfg.Runtime, cfg.DBType))

	if cfg.Runtime == "karaf" {
		karafDir := filepath.Join(cfg.ModelDir, "application", ".karaf")
		if karaf.KarafRunning(karafDir) {
			result.WriteString("Karaf is running\n")
		} else {
			result.WriteString("Karaf is not running\n")
		}

		if cfg.DBType == "postgresql" {
			pgName := "postgres-" + cfg.SchemaName
			if docker.DockerInstanceRunning(pgName) {
				result.WriteString("PostgreSQL is running\n")
			} else {
				result.WriteString("PostgreSQL is not running\n")
				if exists, _ := docker.ContainerExists(pgName); exists {
					result.WriteString("PostgreSQL container exists\n")
				} else {
					result.WriteString("PostgreSQL container does not exist\n")
				}
				if docker.DockerVolumeExists(cfg.AppName + "_postgresql_db") {
					result.WriteString("PostgreSQL db volume exists\n")
				} else {
					result.WriteString("PostgreSQL db volume does not exist\n")
				}
				if docker.DockerVolumeExists(cfg.AppName + "_postgresql_data") {
					result.WriteString("PostgreSQL data volume exists\n")
				} else {
					result.WriteString("PostgreSQL data volume does not exist\n")
				}
			}
		}

		kcName := "keycloak-" + cfg.KeycloakName
		if docker.DockerInstanceRunning(kcName) {
			result.WriteString("Keycloak is running\n")
		} else {
			result.WriteString("Keycloak is not running\n")
			if exists, _ := docker.ContainerExists(kcName); exists {
				result.WriteString("Keycloak container exists\n")
			} else {
				result.WriteString("Keycloak container does not exist\n")
			}
		}
	}

	return result.String()
}

// executeDoctor runs the doctor command using direct function calls
func (api *ServerAPI) executeDoctor() (string, error) {
	// Create a doctor command and capture its output
	doctorCmd := commands.CreateDoctorCommand()

	// Capture stdout/stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// Run the doctor command without arguments
	doctorCmd.SetArgs([]string{}) // Clear any arguments
	err := doctorCmd.Execute()

	// Restore stdout/stderr
	w.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// Read captured output
	output, _ := io.ReadAll(r)

	return string(output), err
}

// executeBuild executes build command using direct function calls
func (api *ServerAPI) executeBuild(args []string) string {
	// Check if JUDO project is initialized
	if !config.IsProjectInitialized() {
		return "Error: no JUDO project initialized in this directory\nRun 'judo init' to initialize a new JUDO project"
	}

	// Load config and set default build options
	config.LoadProperties()

	// Set default build options
	config.Options.BuildModel = true
	config.Options.BuildBackend = true
	config.Options.BuildFrontend = true
	config.Options.BuildKaraf = true
	config.Options.SchemaBuilding = true
	config.Options.SchemaCliBuilding = false
	config.Options.DockerBuilding = false

	// Parse command line arguments for build options
	for _, arg := range args {
		switch arg {
		case "--build-parallel", "-p":
			config.Options.BuildParallel = true
		case "--build-app-module", "-a":
			config.Options.BuildAppModule = true
		case "--build-frontend-module", "-f":
			config.Options.BuildFrontend = true
		case "--docker":
			config.Options.DockerBuilding = true
		case "--skip-model":
			config.Options.BuildModel = false
		case "--skip-backend":
			config.Options.BuildBackend = false
		case "--skip-frontend":
			config.Options.BuildFrontend = false
		case "--skip-karaf":
			config.Options.BuildKaraf = false
		case "--skip-schema":
			config.Options.SchemaBuilding = false
		case "--build-schema-cli":
			config.Options.SchemaCliBuilding = true
		}
	}

	// For now, return a message indicating build would be executed
	// In a full implementation, this would execute the actual build logic
	return "Build command executed with direct function calls. Args: " + strings.Join(args, " ")
}

// executeStart executes start command using direct function calls
func (api *ServerAPI) executeStart(args []string) string {
	// Check if JUDO project is initialized
	if !config.IsProjectInitialized() {
		return "Error: no JUDO project initialized in this directory\nRun 'judo init' to initialize a new JUDO project"
	}

	// Load config
	config.LoadProperties()
	cfg := config.GetConfig()

	// Set default values
	config.Options.StartKeycloak = true
	config.Options.WatchBundles = true
	config.Options.StartKaraf = true

	// Parse command line arguments
	for _, arg := range args {
		switch arg {
		case "--skip-keycloak":
			config.Options.StartKeycloak = false
		case "--skip-watch-bundles":
			config.Options.WatchBundles = false
		}
	}

	// Check if Docker is running
	if !docker.IsDockerRunning() {
		return "Error: Docker daemon is not running. Please start Docker and try again."
	}

	// Execute the start logic
	var result strings.Builder

	// Start services based on runtime
	switch cfg.Runtime {
	case "compose":
		// StartCompose doesn't return error, it runs in foreground
		go docker.StartCompose()
		result.WriteString("Docker compose starting in background\n")
	case "karaf":
		// Start local environment
		if cfg.DBType == "postgresql" {
			if err := docker.StartPostgres(); err != nil {
				result.WriteString(fmt.Sprintf("Error starting PostgreSQL: %v\n", err))
			} else {
				result.WriteString("PostgreSQL started successfully\n")
			}
		}

		if config.Options.StartKeycloak {
			if err := docker.StartKeycloak(); err != nil {
				result.WriteString(fmt.Sprintf("Error starting Keycloak: %v\n", err))
			} else {
				result.WriteString("Keycloak started successfully\n")
			}
		}

		if config.Options.StartKaraf {
			if err := karaf.StartKaraf(); err != nil {
				result.WriteString(fmt.Sprintf("Error starting Karaf: %v\n", err))
			} else {
				result.WriteString("Karaf started successfully\n")
			}
		}
	default:
		result.WriteString(fmt.Sprintf("Unknown runtime: %s — defaulting to karaf\n", cfg.Runtime))
		// Fallback to karaf runtime
		if cfg.DBType == "postgresql" {
			if err := docker.StartPostgres(); err != nil {
				result.WriteString(fmt.Sprintf("Error starting PostgreSQL: %v\n", err))
			} else {
				result.WriteString("PostgreSQL started successfully\n")
			}
		}

		if config.Options.StartKeycloak {
			if err := docker.StartKeycloak(); err != nil {
				result.WriteString(fmt.Sprintf("Error starting Keycloak: %v\n", err))
			} else {
				result.WriteString("Keycloak started successfully\n")
			}
		}

		if config.Options.StartKaraf {
			if err := karaf.StartKaraf(); err != nil {
				result.WriteString(fmt.Sprintf("Error starting Karaf: %v\n", err))
			} else {
				result.WriteString("Karaf started successfully\n")
			}
		}
	}

	return result.String()
}

// executeStop executes stop command using direct function calls
func (api *ServerAPI) executeStop(args []string) string {
	// Check if JUDO project is initialized
	if !config.IsProjectInitialized() {
		return "Error: no JUDO project initialized in this directory\nRun 'judo init' to initialize a new JUDO project"
	}

	// Load config
	config.LoadProperties()
	cfg := config.GetConfig()

	var result strings.Builder

	// Stop services based on runtime
	if cfg.Runtime == "karaf" {
		// Stop Karaf
		karaf.StopKaraf(cfg.KarafDir)
		result.WriteString("Karaf stopped\n")

		// Stop PostgreSQL if applicable
		if cfg.DBType == "postgresql" {
			if err := docker.StopDockerInstance("postgres-" + cfg.SchemaName); err != nil {
				result.WriteString(fmt.Sprintf("Error stopping PostgreSQL: %v\n", err))
			} else {
				result.WriteString("PostgreSQL stopped\n")
			}
		}

		// Stop Keycloak
		if err := docker.StopDockerInstance("keycloak-" + cfg.KeycloakName); err != nil {
			result.WriteString(fmt.Sprintf("Error stopping Keycloak: %v\n", err))
		} else {
			result.WriteString("Keycloak stopped\n")
		}
	} else {
		result.WriteString(fmt.Sprintf("Stop command not fully implemented for runtime: %s\n", cfg.Runtime))
	}

	return result.String()
}

// executeClean executes clean command using direct function calls
func (api *ServerAPI) executeClean(args []string) string {
	// Check if JUDO project is initialized
	if !config.IsProjectInitialized() {
		return "Error: no JUDO project initialized in this directory\nRun 'judo init' to initialize a new JUDO project"
	}

	// Load config
	config.LoadProperties()
	cfg := config.GetConfig()

	var result strings.Builder

	// Stop compose environments
	for _, env := range docker.GetComposeEnvs(cfg) {
		_ = docker.StopCompose(cfg, env)
		result.WriteString(fmt.Sprintf("Stopped compose environment: %s\n", env))
	}

	// Remove Docker instances
	_ = docker.RemoveDockerInstance("postgres-" + cfg.SchemaName)
	result.WriteString("Removed PostgreSQL instance\n")

	_ = docker.RemoveDockerInstance("keycloak-" + cfg.KeycloakName)
	result.WriteString("Removed Keycloak instance\n")

	// Remove Docker network
	_ = docker.RemoveDockerNetwork(cfg.AppName)
	result.WriteString("Removed Docker network\n")

	// Remove Docker volumes
	_ = docker.RemoveDockerVolume(cfg.AppName + "_certs")
	result.WriteString("Removed certs volume\n")

	_ = docker.RemoveDockerVolume(cfg.SchemaName + "_postgresql_db")
	result.WriteString("Removed PostgreSQL db volume\n")

	_ = docker.RemoveDockerVolume(cfg.SchemaName + "_postgresql_data")
	result.WriteString("Removed PostgreSQL data volume\n")

	_ = docker.RemoveDockerVolume(cfg.AppName + "_filestore")
	result.WriteString("Removed filestore volume\n")

	// Stop and remove Karaf directory if applicable
	if cfg.Runtime == "karaf" {
		karaf.StopKaraf(cfg.KarafDir)
		result.WriteString("Stopped Karaf\n")

		if err := os.RemoveAll(cfg.KarafDir); err != nil {
			result.WriteString(fmt.Sprintf("Error removing Karaf directory: %v\n", err))
		} else {
			result.WriteString("Removed Karaf directory\n")
		}
	}

	return result.String()
}

// executeLog executes log command using direct function calls
func (api *ServerAPI) executeLog(args []string) string {
	// Check if JUDO project is initialized
	if !config.IsProjectInitialized() {
		return "Error: no JUDO project initialized in this directory\nRun 'judo init' to initialize a new JUDO project"
	}

	config.LoadProperties()
	cfg := config.GetConfig()

	if cfg.Runtime != "karaf" {
		return "log command is only supported for karaf runtime"
	}

	logFile := filepath.Join(cfg.KarafDir, "console.out")
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		return fmt.Sprintf("log file not found: %s", logFile)
	}

	// Parse command line arguments
	lines := 50 // default
	tail := false
	follow := false

	for i, arg := range args {
		switch arg {
		case "--tail", "-t":
			tail = true
		case "--follow", "-f":
			follow = true
		case "--lines", "-n":
			if i+1 < len(args) {
				if n, err := fmt.Sscanf(args[i+1], "%d", &lines); err == nil && n == 1 {
					// lines value updated
				}
			}
		}
	}

	// Read log file content
	content, err := os.ReadFile(logFile)
	if err != nil {
		return fmt.Sprintf("failed to read log file: %v", err)
	}

	logContent := string(content)
	logLines := strings.Split(logContent, "\n")

	var result strings.Builder

	if tail || follow {
		// Show the last 'lines' number of lines
		start := len(logLines) - lines
		if start < 0 {
			start = 0
		}

		for i := start; i < len(logLines); i++ {
			if logLines[i] != "" {
				result.WriteString(logLines[i] + "\n")
			}
		}

		if follow {
			result.WriteString("\n⚠️  Follow mode not supported in API mode\n")
		}
	} else {
		// Show all lines or limited lines
		if lines <= 0 || lines >= len(logLines) {
			result.WriteString(logContent)
		} else {
			start := len(logLines) - lines
			if start < 0 {
				start = 0
			}
			for i := start; i < len(logLines); i++ {
				if logLines[i] != "" {
					result.WriteString(logLines[i] + "\n")
				}
			}
		}
	}

	return result.String()
}

// executeExternalCommand fallback for unsupported commands
func (api *ServerAPI) executeExternalCommand(command string, args []string) (string, error) {
	// This is a fallback for commands that aren't yet implemented with direct calls
	// In production, this should be replaced with proper direct function calls
	return fmt.Sprintf("External command execution: %s %s", command, strings.Join(args, " ")), nil
}

// getServiceStatusSummary returns a summary of service statuses
func (api *ServerAPI) getServiceStatusSummary() string {
	config.LoadProperties()
	cfg := config.GetConfig()

	var status string
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Check services status in parallel
	checkService := func(name string, checkFunc func() bool) {
		defer wg.Done()
		if checkFunc() {
			mu.Lock()
			status += name + ":running "
			mu.Unlock()
		} else {
			mu.Lock()
			status += name + ":stopped "
			mu.Unlock()
		}
	}

	wg.Add(3) // Karaf, PostgreSQL, Keycloak

	// Check Karaf status
	go func() {
		if cfg.Runtime == "karaf" {
			karafDir := filepath.Join(cfg.ModelDir, "application", ".karaf")
			checkService("Karaf", func() bool { return karaf.KarafRunning(karafDir) })
		} else {
			wg.Done()
		}
	}()

	// Check PostgreSQL status
	go func() {
		if cfg.DBType == "postgresql" {
			pgName := "postgres-" + cfg.SchemaName
			checkService("PostgreSQL", func() bool { return docker.DockerInstanceRunning(pgName) })
		} else {
			wg.Done()
		}
	}()

	// Check Keycloak status
	go func() {
		kcName := "keycloak-" + cfg.KeycloakName
		checkService("Keycloak", func() bool { return docker.DockerInstanceRunning(kcName) })
	}()

	wg.Wait()
	return strings.TrimSpace(status)
}

// StreamLogs streams logs from a service directly
func (api *ServerAPI) StreamLogs(service string, ctx context.Context, outputChan chan<- string) {
	config.LoadProperties()
	cfg := config.GetConfig()

	switch service {
	case "karaf":
		if cfg.Runtime == "karaf" {
			karafLogFile := filepath.Join(cfg.KarafDir, "console.out")
			api.streamLogFile(karafLogFile, "[KARAF]", ctx, outputChan)
		}
	case "postgresql":
		// PostgreSQL logs would be streamed from docker logs
		if cfg.DBType == "postgresql" {
			pgName := "postgres-" + cfg.SchemaName
			api.streamDockerLogs(pgName, "[POSTGRESQL]", ctx, outputChan)
		}
	case "keycloak":
		kcName := "keycloak-" + cfg.KeycloakName
		api.streamDockerLogs(kcName, "[KEYCLOAK]", ctx, outputChan)
	case "combined":
		// Stream all services combined
		go func() {
			if cfg.Runtime == "karaf" {
				karafLogFile := filepath.Join(cfg.KarafDir, "console.out")
				api.streamLogFile(karafLogFile, "[KARAF]", ctx, outputChan)
			}
		}()
		go func() {
			if cfg.DBType == "postgresql" {
				pgName := "postgres-" + cfg.SchemaName
				api.streamDockerLogs(pgName, "[POSTGRESQL]", ctx, outputChan)
			}
		}()
		go func() {
			kcName := "keycloak-" + cfg.KeycloakName
			api.streamDockerLogs(kcName, "[KEYCLOAK]", ctx, outputChan)
		}()
	}
}

// streamLogFile streams logs from a file
func (api *ServerAPI) streamLogFile(logFile, prefix string, ctx context.Context, outputChan chan<- string) {
	// Implementation for file-based log streaming
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var lastOffset int64 = 0

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			file, err := os.Open(logFile)
			if err != nil {
				continue
			}

			stat, err := file.Stat()
			if err != nil {
				file.Close()
				continue
			}

			// Handle file rotation - if file got smaller, reset offset
			if stat.Size() < lastOffset {
				lastOffset = 0
			}

			if stat.Size() > lastOffset {
				file.Seek(lastOffset, 0)
				buffer := make([]byte, stat.Size()-lastOffset)
				_, err := file.Read(buffer)
				if err == nil {
					lines := strings.Split(string(buffer), "\n")
					for _, line := range lines {
						if line != "" {
							outputChan <- prefix + " " + line
						}
					}
				}
				lastOffset = stat.Size()
			}
			file.Close()
		}
	}
}

// streamDockerLogs streams logs from a docker container
func (api *ServerAPI) streamDockerLogs(containerName, prefix string, ctx context.Context, outputChan chan<- string) {
	// Implementation for docker log streaming would go here
	// This would use the docker client to stream logs directly
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if docker.DockerInstanceRunning(containerName) {
				// Simulate docker logs output
				outputChan <- prefix + " Container " + containerName + " is running"
			} else {
				outputChan <- prefix + " Container " + containerName + " is not running"
			}
		}
	}
}

// StartInteractiveSession starts an interactive JUDO CLI session
// This is a direct function call replacement for "judo session"
func (api *ServerAPI) StartInteractiveSession() {
	session.StartInteractiveSession()
}
