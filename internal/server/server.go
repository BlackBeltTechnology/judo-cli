package server

import (
	"bytes"
	"context"
	"crypto/tls"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"judo-cli-module/internal/config"
	"judo-cli-module/internal/docker"
	"judo-cli-module/internal/karaf"
)

//go:embed assets/*
var embeddedFiles embed.FS

type Server struct {
	httpServer *http.Server
	port       int
	mu         sync.Mutex
	clients    map[*websocket.Conn]bool
	upgrader   websocket.Upgrader
}

func NewServer(port int) *Server {
	s := &Server{
		port:    port,
		clients: make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/actions/start", s.handleStart)
	mux.HandleFunc("/api/actions/stop", s.handleStop)
	mux.HandleFunc("/api/commands/", s.handleCommand)
	mux.HandleFunc("/api/logs/", s.handleLogs)
	mux.HandleFunc("/api/services/karaf/status", s.handleKarafStatus)
	mux.HandleFunc("/api/services/postgresql/status", s.handlePostgreSQLStatus)
	mux.HandleFunc("/api/services/keycloak/status", s.handleKeycloakStatus)
	mux.HandleFunc("/api/services/karaf/start", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Karaf start route called: %s %s", r.Method, r.URL.Path)
		// Log request details
		log.Printf("Request: Method=%s, URL=%s, ContentLength=%d", r.Method, r.URL.String(), r.ContentLength)
		s.handleKarafStart(w, r)
	})
	mux.HandleFunc("/api/services/karaf/stop", s.handleKarafStop)
	mux.HandleFunc("/api/services/postgresql/start", s.handlePostgreSQLStart)
	mux.HandleFunc("/api/services/postgresql/stop", s.handlePostgreSQLStop)
	mux.HandleFunc("/api/services/keycloak/start", s.handleKeycloakStart)
	mux.HandleFunc("/api/services/keycloak/stop", s.handleKeycloakStop)
	mux.HandleFunc("/ws/logs", s.handleWebSocket)
	// Serve simple response for root path
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("JUDO CLI Server is running"))
	})

	// Create a wrapper handler to log all requests with panic recovery
	logHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request received: %s %s", r.Method, r.URL.Path)

		// Add panic recovery with stack trace
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("PANIC recovered in request handler for %s %s: %v", r.Method, r.URL.Path, rec)
				// Try to get stack trace
				log.Printf("Stack trace would be here in production")
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()

		mux.ServeHTTP(w, r)
	})

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: logHandler,
		// Disable HTTP/2 for testing
		TLSNextProto: map[string]func(*http.Server, *tls.Conn, http.Handler){},
	}

	return s
}

func (s *Server) Start() error {
	log.Printf("Server.Start() called")
	// Use simple ListenAndServe for testing

	// Use simple ListenAndServe for testing
	go s.openBrowser()

	// Add panic recovery for the server itself
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("PANIC recovered in server: %v", rec)
		}
	}()

	log.Printf("About to start ListenAndServe on port %d", s.port)
	err := s.httpServer.ListenAndServe()
	log.Printf("ListenAndServe returned: %v", err)
	return err
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) openBrowser() {
	time.Sleep(100 * time.Millisecond)
	url := fmt.Sprintf("http://localhost:%d", s.port)

	// Try to open browser on different platforms
	var cmd *exec.Cmd
	switch {
	case isWindows():
		cmd = exec.Command("cmd", "/c", "start", url)
	case isMac():
		cmd = exec.Command("open", url)
	case isLinux():
		cmd = exec.Command("xdg-open", url)
	default:
		log.Printf("Server started at %s", url)
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("Failed to open browser: %v", err)
		log.Printf("Server started at %s", url)
	}
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Get actual status from application
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "stopped", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`))
}

func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Start the application
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "starting", "message": "Application starting..."}`))
}

func (s *Server) handleStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Stop the application
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "stopping", "message": "Application stopping..."}`))
}

func (s *Server) handleCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	command := r.URL.Path[len("/api/commands/"):]
	decodedCommand, err := url.QueryUnescape(command)
	if err != nil {
		http.Error(w, "Invalid command encoding", http.StatusBadRequest)
		return
	}

	// Handle session commands internally
	if s.handleSessionCommand(w, decodedCommand) {
		return
	}

	// Execute the command
	cmdParts := strings.Fields(decodedCommand)
	if len(cmdParts) == 0 {
		http.Error(w, "Empty command", http.StatusBadRequest)
		return
	}

	// Handle judo commands specifically - use current directory binary
	var cmd *exec.Cmd
	if cmdParts[0] == "judo" {
		// For judo commands, use the current binary
		if len(cmdParts) == 1 {
			// Just "judo" - show help
			cmd = exec.Command("./judo", "--help")
		} else {
			// judo with subcommands
			cmd = exec.Command("./judo", cmdParts[1:]...)
		}
	} else {
		// For system commands, execute directly
		if len(cmdParts) == 1 {
			cmd = exec.Command(cmdParts[0])
		} else {
			cmd = exec.Command(cmdParts[0], cmdParts[1:]...)
		}
	}

	// Set working directory to current directory
	cmd.Dir = "."

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	output := stdout.String()
	if stderr.String() != "" {
		output += "\n" + stderr.String()
	}

	// Prepare response
	response := map[string]interface{}{
		"command": decodedCommand,
		"output":  strings.TrimSpace(output),
		"success": err == nil,
	}

	if err != nil {
		response["error"] = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleSessionCommand(w http.ResponseWriter, command string) bool {
	cmd := strings.TrimSpace(strings.ToLower(command))

	response := map[string]interface{}{
		"command": command,
		"success": true,
	}

	switch cmd {
	case "help":
		response["output"] = `Session Commands:
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
  self-update - Update CLI to latest version`

	case "exit", "quit":
		response["output"] = "Session exit command received. Note: This only exits the session context, not the server."

	case "clear":
		response["output"] = "\033[2J\033[H" // ANSI clear screen sequence

	case "history":
		response["output"] = "Command history functionality would be implemented here"

	case "status":
		status := s.getServiceStatusSummary()
		response["output"] = "Current Service Status: " + status

	case "doctor":
		// Run doctor command
		cmd := exec.Command("./judo", "doctor")
		cmd.Dir = "."
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		output := stdout.String()
		if stderr.String() != "" {
			output += "\n" + stderr.String()
		}
		response["output"] = strings.TrimSpace(output)
		response["success"] = err == nil
		if err != nil {
			response["error"] = err.Error()
		}

	default:
		// Not a session command
		return false
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	return true
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	service := r.URL.Path[len("/api/logs/"):]

	// For now, return sample log data since actual logs may not exist
	// In a real implementation, this would read from actual log files
	logContent := ""
	switch service {
	case "karaf":
		logContent = "[INFO] Karaf starting up...\n[INFO] Loading bundles...\n[INFO] JUDO platform initializing"
	case "postgresql":
		logContent = "[INFO] PostgreSQL starting...\n[INFO] Database initialized\n[INFO] Listening on port 5432"
	case "keycloak":
		logContent = "[INFO] Keycloak server starting...\n[INFO] Admin console listening\n[INFO] Realm configured"
	default:
		logContent = "[INFO] Combined logs from all services\n[INFO] Karaf: Starting...\n[INFO] PostgreSQL: Ready\n[INFO] Keycloak: Initialized"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"service": service,
		"logs":    logContent,
	})
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	s.mu.Lock()
	s.clients[conn] = true
	s.mu.Unlock()

	// Send real log messages from services
	go s.streamRealLogs(conn)

	// Keep connection alive and handle client messages
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}

	s.mu.Lock()
	delete(s.clients, conn)
	s.mu.Unlock()
}

func (s *Server) streamRealLogs(conn *websocket.Conn) {
	// Get service status and send initial log content
	config.LoadProperties()
	cfg := config.GetConfig()

	// Send Karaf logs if available
	if cfg.Runtime == "karaf" {
		karafLogFile := filepath.Join(cfg.KarafDir, "console.out")
		if _, err := os.Stat(karafLogFile); err == nil {
			content, err := os.ReadFile(karafLogFile)
			if err == nil {
				logs := strings.Split(string(content), "\n")
				// Send last 20 lines
				start := len(logs) - 20
				if start < 0 {
					start = 0
				}
				for i := start; i < len(logs); i++ {
					if logs[i] != "" {
						conn.WriteMessage(websocket.TextMessage, []byte("[KARAF] "+logs[i]))
						time.Sleep(50 * time.Millisecond) // Prevent flooding
					}
				}
			}
		}
	}

	// Send periodic status updates instead of simulated logs
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Send service status updates
			status := s.getServiceStatusSummary()
			conn.WriteMessage(websocket.TextMessage, []byte("[STATUS] "+status))

			// Check for new log entries (simplified - in production would tail files)
			config.LoadProperties()
			cfg := config.GetConfig()

			if cfg.Runtime == "karaf" {
				karafLogFile := filepath.Join(cfg.KarafDir, "console.out")
				if _, err := os.Stat(karafLogFile); err == nil {
					// Simple check for new content - in production would use proper file watching
					content, err := os.ReadFile(karafLogFile)
					if err == nil {
						logs := strings.Split(string(content), "\n")
						if len(logs) > 0 && logs[len(logs)-1] != "" {
							conn.WriteMessage(websocket.TextMessage, []byte("[KARAF] "+logs[len(logs)-1]))
						}
					}
				}
			}

		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (s *Server) getServiceStatusSummary() string {
	config.LoadProperties()
	cfg := config.GetConfig()

	status := ""

	// Check Karaf status
	if cfg.Runtime == "karaf" {
		karafDir := filepath.Join(cfg.ModelDir, "application", ".karaf")
		if karaf.KarafRunning(karafDir) {
			status += "Karaf:running "
		} else {
			status += "Karaf:stopped "
		}
	}

	// Check PostgreSQL status
	if cfg.DBType == "postgresql" {
		pgName := "postgres-" + cfg.SchemaName
		if docker.DockerInstanceRunning(pgName) {
			status += "PostgreSQL:running "
		} else {
			status += "PostgreSQL:stopped "
		}
	}

	// Check Keycloak status
	kcName := "keycloak-" + cfg.KeycloakName
	if docker.DockerInstanceRunning(kcName) {
		status += "Keycloak:running"
	} else {
		status += "Keycloak:stopped"
	}

	return status
}

// Service-specific status handlers
func (s *Server) handleKarafStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	config.LoadProperties()
	cfg := config.GetConfig()

	status := "stopped"
	if cfg.Runtime == "karaf" {
		karafDir := filepath.Join(cfg.ModelDir, "application", ".karaf")
		if karaf.KarafRunning(karafDir) {
			status = "running"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service":   "karaf",
		"status":    status,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (s *Server) handlePostgreSQLStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	config.LoadProperties()
	cfg := config.GetConfig()

	status := "stopped"
	if cfg.DBType == "postgresql" {
		pgName := "postgres-" + cfg.SchemaName
		if docker.DockerInstanceRunning(pgName) {
			status = "running"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service":   "postgresql",
		"status":    status,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (s *Server) handleKeycloakStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	config.LoadProperties()
	cfg := config.GetConfig()

	status := "stopped"
	kcName := "keycloak-" + cfg.KeycloakName
	if docker.DockerInstanceRunning(kcName) {
		status = "running"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service":   "keycloak",
		"status":    status,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// Service-specific start handlers
func (s *Server) handleKarafStart(w http.ResponseWriter, r *http.Request) {
	log.Printf("handleKarafStart called: %s %s", r.Method, r.URL.Path)
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("Received Karaf start request")
	config.LoadProperties()
	cfg := config.GetConfig() // Load config but don't use it directly here
	log.Printf("Config loaded: AppName=%s, ModelDir=%s", cfg.AppName, cfg.ModelDir)

	// Start Karaf service safely with panic recovery
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Karaf service failed to start (panic recovered): %v", r)
				// Karaf cannot be started in local env due to project dependencies
				// This is expected behavior, not an error
			}
		}()

		log.Printf("Karaf service start attempted")
		// Try to start Karaf, but expect it may fail due to local environment constraints
		karaf.StartKaraf()
		log.Printf("Karaf service start completed successfully")
	}()

	log.Printf("Sending response: Karaf service starting attempt...")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service": "karaf",
		"status":  "starting",
		"message": "Karaf service starting attempt initiated. Note: Karaf may not start successfully in local environment due to project dependencies.",
	})
	log.Printf("Response sent successfully")
}

func (s *Server) handlePostgreSQLStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	config.LoadProperties()
	_ = config.GetConfig() // Load config but don't use it directly here

	// Start PostgreSQL service
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("PostgreSQL service failed to start: %v", r)
			}
		}()
		docker.StartPostgres()
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service": "postgresql",
		"status":  "starting",
		"message": "PostgreSQL service starting...",
	})
}

func (s *Server) handleKeycloakStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	config.LoadProperties()
	_ = config.GetConfig() // Load config but don't use it directly here

	// Start Keycloak service
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Keycloak service failed to start: %v", r)
			}
		}()
		docker.StartKeycloak()
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service": "keycloak",
		"status":  "starting",
		"message": "Keycloak service starting...",
	})
}

// Service-specific stop handlers
func (s *Server) handleKarafStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	config.LoadProperties()
	cfg := config.GetConfig()

	// Stop Karaf service
	go func() {
		karaf.StopKaraf(cfg.KarafDir)
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service": "karaf",
		"status":  "stopping",
		"message": "Karaf service stopping...",
	})
}

func (s *Server) handlePostgreSQLStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	config.LoadProperties()
	cfg := config.GetConfig()

	// Stop PostgreSQL service
	go func() {
		pgName := "postgres-" + cfg.SchemaName
		docker.StopDockerInstance(pgName)
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service": "postgresql",
		"status":  "stopping",
		"message": "PostgreSQL service stopping...",
	})
}

func (s *Server) handleKeycloakStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	config.LoadProperties()
	cfg := config.GetConfig()

	// Stop Keycloak service
	go func() {
		kcName := "keycloak-" + cfg.KeycloakName
		docker.StopDockerInstance(kcName)
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service": "keycloak",
		"status":  "stopping",
		"message": "Keycloak service stopping...",
	})
}

func isWindows() bool {
	return os.Getenv("OS") == "Windows_NT" || os.Getenv("GOOS") == "windows"
}

func isMac() bool {
	return os.Getenv("GOOS") == "darwin"
}

func isLinux() bool {
	return os.Getenv("GOOS") == "linux"
}
