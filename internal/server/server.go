package server

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
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

// connectionPool manages a pool of WebSocket connections
type connectionPool struct {
	mu       sync.Mutex
	conns    map[*websocket.Conn]bool
	maxConns int
}

func newConnectionPool(maxConns int) *connectionPool {
	return &connectionPool{
		conns:    make(map[*websocket.Conn]bool),
		maxConns: maxConns,
	}
}

func (p *connectionPool) add(conn *websocket.Conn) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.conns) >= p.maxConns {
		return false
	}

	p.conns[conn] = true
	return true
}

func (p *connectionPool) remove(conn *websocket.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.conns, conn)
}

func (p *connectionPool) broadcast(message []byte) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for conn := range p.conns {
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			// Remove broken connections
			delete(p.conns, conn)
		}
	}
}

func (p *connectionPool) count() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.conns)
}

//go:embed assets/*
var embeddedFiles embed.FS

// Static frontend is served from embedded assets or frontend/build at runtime

type Server struct {
	httpServer *http.Server
	port       int
	mu         sync.Mutex
	clients    map[*websocket.Conn]bool
	upgrader   websocket.Upgrader

	// Connection pools
	logPool     *connectionPool
	sessionPool *connectionPool

	// Session input handling
	sessionStdin io.WriteCloser
	sessionConn  *websocket.Conn

	// API for direct function calls
	api *ServerAPI
}

func NewServer(port int) *Server {
	log.Printf("Creating new server on port %d", port)

	// Add panic recovery for server creation
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("PANIC recovered in NewServer: %v", rec)
			log.Printf("Stack trace would be here in production")
		}
	}()

	s := &Server{
		port: port,
		// Minimal initialization for testing
		clients: make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		// Skip connection pools and API for now
		logPool:     &connectionPool{conns: make(map[*websocket.Conn]bool), maxConns: 100},
		sessionPool: &connectionPool{conns: make(map[*websocket.Conn]bool), maxConns: 50},
		api:         NewServerAPI(),
	}

	mux := http.NewServeMux()

	// Serve static frontend files - try embedded assets first, then frontend/build
	embeddedFS, err := fs.Sub(embeddedFiles, "assets")
	var useEmbedded bool

	if err == nil {
		// Test if embedded assets actually contain files
		entries, err := fs.ReadDir(embeddedFS, ".")
		if err == nil && len(entries) > 0 {
			mux.Handle("/", http.FileServer(http.FS(embeddedFS)))
			useEmbedded = true
		} else {
			log.Printf("Embedded assets not available or empty, trying frontend/build")
		}
	} else {
		log.Printf("Embedded assets not available: %v", err)
	}

	// If we didn't set up embedded assets or they're empty, try frontend/build directory
	if !useEmbedded {
		if _, err := os.Stat("frontend/build"); err == nil {
			log.Printf("Serving frontend from frontend/build")
			mux.Handle("/", http.FileServer(http.Dir("frontend/build")))
		} else {
			log.Printf("Frontend build directory not found, using fallback landing page")
			// Final fallback: simple landing page
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/" {
					log.Printf("DEBUG: Path not found: %s", r.URL.Path)
					http.NotFound(w, r)
					return
				}
				log.Printf("DEBUG: Serving landing page")
				w.Header().Set("Content-Type", "text/html")
				fmt.Fprintf(w, `<html><body><h1>JUDO CLI Server</h1><p>Server is running on port %d.</p><p><a href="/api/status">API Status</a></p></body></html>`, s.port)
			})
		}
	}
	// REST API endpoints
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/commands/", s.handleCommand)
	mux.HandleFunc("/api/project/init/status", s.handleProjectInitStatus)
	mux.HandleFunc("/api/services/status", s.handleServicesStatus)
	mux.HandleFunc("/api/services/karaf/status", s.handleKarafStatus)
	mux.HandleFunc("/api/services/postgresql/status", s.handlePostgreSQLStatus)
	mux.HandleFunc("/api/services/keycloak/status", s.handleKeycloakStatus)
	mux.HandleFunc("/api/services/karaf/start", s.handleKarafStart)
	mux.HandleFunc("/api/services/postgresql/start", s.handlePostgreSQLStart)
	mux.HandleFunc("/api/services/keycloak/start", s.handleKeycloakStart)
	mux.HandleFunc("/api/services/karaf/stop", s.handleKarafStop)
	mux.HandleFunc("/api/services/postgresql/stop", s.handlePostgreSQLStop)
	mux.HandleFunc("/api/services/keycloak/stop", s.handleKeycloakStop)
	mux.HandleFunc("/api/services/start", s.handleServicesStart)
	mux.HandleFunc("/api/services/stop", s.handleServicesStop)
	// Optional simple logs HTTP endpoint
	mux.HandleFunc("/api/logs/", s.handleLogs)

	// WebSocket endpoints
	mux.HandleFunc("/ws/logs/combined", s.handleCombinedLogsWebSocket)
	mux.HandleFunc("/ws/logs/service/", s.handleServiceLogsWebSocket)
	mux.HandleFunc("/ws/session", s.handleSessionWebSocket)

	// Create a simple wrapper handler to log all requests
	logHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mux.ServeHTTP(w, r)
	})

	log.Printf("Creating HTTP server on port %d", port)
	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", port),
		Handler: logHandler,
	}
	log.Printf("HTTP server created successfully")

	return s
}

func (s *Server) Start() error {
	log.Printf("Server.Start() called on port %d", s.port)
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("PANIC recovered in server: %v", rec)
		}
	}()

	// Open browser automatically
	go s.openBrowser()

	return s.httpServer.ListenAndServe()
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

	status := s.getServiceStatusSummary()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    status,
		"timestamp": time.Now().Format(time.RFC3339),
	})
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

	// Execute the command using ServerAPI direct function calls
	cmdParts := strings.Fields(decodedCommand)
	if len(cmdParts) == 0 {
		http.Error(w, "Empty command", http.StatusBadRequest)
		return
	}

	// Handle judo commands using ServerAPI
	var result string
	var execErr error

	if cmdParts[0] == "judo" {
		// Handle "judo" command with subcommands
		if len(cmdParts) == 1 {
			// Just "judo" - show help
			result = s.api.executeJudoHelp()
		} else {
			// judo with subcommands - execute the subcommand
			result, execErr = s.api.ExecuteCommand(cmdParts[1], cmdParts[2:])
		}
	} else {
		// For non-judo commands, try to execute via ServerAPI
		result, execErr = s.api.ExecuteCommand(cmdParts[0], cmdParts[1:])
	}

	// Prepare response
	response := map[string]interface{}{
		"command": decodedCommand,
		"output":  strings.TrimSpace(result),
		"success": execErr == nil,
	}

	if execErr != nil {
		response["error"] = execErr.Error()
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
		// Run doctor command using ServerAPI
		output, err := s.api.executeDoctor()
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

	config.LoadProperties()
	cfg := config.GetConfig()

	var logContent string

	switch service {
	case "karaf":
		if cfg.Runtime == "karaf" {
			karafLogFile := filepath.Join(cfg.KarafDir, "console.out")
			if _, statErr := os.Stat(karafLogFile); statErr == nil {
				content, readErr := os.ReadFile(karafLogFile)
				if readErr == nil {
					logContent = string(content)
				} else {
					logContent = fmt.Sprintf("Error reading Karaf logs: %v", readErr)
				}
			} else {
				logContent = "Karaf log file not found"
			}
		} else {
			logContent = "Karaf runtime not configured"
		}
	case "postgresql":
		if cfg.DBType == "postgresql" {
			pgName := "postgres-" + cfg.SchemaName
			if exists, _ := docker.ContainerExists(pgName); exists {
				// Get recent logs from Docker container
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				logChan := make(chan string, 100)
				go s.api.streamDockerLogs(pgName, "", ctx, logChan)

				// Collect logs for a brief period
				var logs []string
				timer := time.NewTimer(2 * time.Second)
				defer timer.Stop()

				for {
					select {
					case logLine := <-logChan:
						logs = append(logs, logLine)
					case <-timer.C:
						logContent = strings.Join(logs, "\n")
						if logContent == "" {
							logContent = "No recent PostgreSQL logs available"
						}
						goto donePostgres
					}
				}
			donePostgres:
			} else {
				logContent = "PostgreSQL container does not exist"
			}
		} else {
			logContent = "PostgreSQL not configured"
		}
	case "keycloak":
		kcName := "keycloak-" + cfg.KeycloakName
		if exists, _ := docker.ContainerExists(kcName); exists {
			// Get recent logs from Docker container
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			logChan := make(chan string, 100)
			go s.api.streamDockerLogs(kcName, "", ctx, logChan)

			// Collect logs for a brief period
			var logs []string
			timer := time.NewTimer(2 * time.Second)
			defer timer.Stop()

			for {
				select {
				case logLine := <-logChan:
					logs = append(logs, logLine)
				case <-timer.C:
					logContent = strings.Join(logs, "\n")
					if logContent == "" {
						logContent = "No recent Keycloak logs available"
					}
					goto doneKeycloak
				}
			}
		doneKeycloak:
		} else {
			logContent = "Keycloak container does not exist"
		}
	default:
		// Combined logs - show recent logs from all available services
		var allLogs []string

		// Karaf logs
		if cfg.Runtime == "karaf" {
			karafLogFile := filepath.Join(cfg.KarafDir, "console.out")
			if _, statErr := os.Stat(karafLogFile); statErr == nil {
				content, readErr := os.ReadFile(karafLogFile)
				if readErr == nil {
					karafLines := strings.Split(string(content), "\n")
					// Get last 20 lines
					start := len(karafLines) - 20
					if start < 0 {
						start = 0
					}
					for i := start; i < len(karafLines); i++ {
						if karafLines[i] != "" {
							allLogs = append(allLogs, "[KARAF] "+karafLines[i])
						}
					}
				}
			}
		}

		// PostgreSQL logs
		if cfg.DBType == "postgresql" {
			pgName := "postgres-" + cfg.SchemaName
			if exists, _ := docker.ContainerExists(pgName); exists {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()

				logChan := make(chan string, 100)
				go s.api.streamDockerLogs(pgName, "[POSTGRESQL]", ctx, logChan)

				timer := time.NewTimer(1 * time.Second)
				defer timer.Stop()

				for {
					select {
					case logLine := <-logChan:
						allLogs = append(allLogs, logLine)
					case <-timer.C:
						goto doneCombinedPostgres
					}
				}
			doneCombinedPostgres:
			}
		}

		// Keycloak logs
		kcName := "keycloak-" + cfg.KeycloakName
		if exists, _ := docker.ContainerExists(kcName); exists {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			logChan := make(chan string, 100)
			go s.api.streamDockerLogs(kcName, "[KEYCLOAK]", ctx, logChan)

			timer := time.NewTimer(1 * time.Second)
			defer timer.Stop()

			for {
				select {
				case logLine := <-logChan:
					allLogs = append(allLogs, logLine)
				case <-timer.C:
					goto doneCombinedKeycloak
				}
			}
		doneCombinedKeycloak:
		}

		logContent = strings.Join(allLogs, "\n")
		if logContent == "" {
			logContent = "No recent logs available from any service"
		}
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
	// Use API to stream real logs instead of simulated ones
	config.LoadProperties()
	cfg := config.GetConfig()

	// Create a channel for log output
	logChan := make(chan string, 100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Stream logs from various services based on configuration
	if cfg.Runtime == "karaf" {
		karafLogFile := filepath.Join(cfg.KarafDir, "console.out")
		if _, err := os.Stat(karafLogFile); err == nil {
			go s.api.streamLogFile(karafLogFile, "[KARAF]", ctx, logChan)
		}
	}

	// Stream Docker container logs
	if cfg.DBType == "postgresql" {
		pgName := "postgres-" + cfg.SchemaName
		go s.api.streamDockerLogs(pgName, "[POSTGRESQL]", ctx, logChan)
	}

	kcName := "keycloak-" + cfg.KeycloakName
	go s.api.streamDockerLogs(kcName, "[KEYCLOAK]", ctx, logChan)

	// Send connection established message
	conn.WriteMessage(websocket.TextMessage, []byte("[SERVER] Real log streaming connected"))

	// Stream logs to WebSocket
	for {
		select {
		case <-ctx.Done():
			return
		case logLine := <-logChan:
			if err := conn.WriteMessage(websocket.TextMessage, []byte(logLine)); err != nil {
				return
			}
		case <-time.After(30 * time.Second):
			// Send empty heartbeat to keep connection alive (no message content)
			conn.WriteMessage(websocket.TextMessage, []byte(""))
		}
	}
}

func (s *Server) getServiceStatusSummary() string {
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

	// Start Karaf service with proper error handling
	go func() {
		log.Printf("Karaf service start attempted")
		// Try to start Karaf, but expect it may fail due to local environment constraints
		if err := s.safeStartKaraf(); err != nil {
			log.Printf("Karaf service failed to start: %v", err)
		} else {
			log.Printf("Karaf service start completed successfully")
		}
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
		if err := s.safeStartPostgreSQL(); err != nil {
			log.Printf("PostgreSQL service failed to start: %v", err)
		} else {
			log.Printf("PostgreSQL service start completed successfully")
		}
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
		if err := s.safeStartKeycloak(); err != nil {
			log.Printf("Keycloak service failed to start: %v", err)
		} else {
			log.Printf("Keycloak service start completed successfully")
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service": "keycloak",
		"status":  "starting",
		"message": "Keycloak service starting...",
	})
}

// Parallel service operations
func (s *Server) handleServicesStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	config.LoadProperties()
	_ = config.GetConfig()

	// Start all services in parallel
	go func() {
		var wg sync.WaitGroup
		services := []string{"karaf", "postgresql", "keycloak"}
		results := make(map[string]error)
		var mu sync.Mutex

		for _, service := range services {
			wg.Add(1)
			go func(svc string) {
				defer wg.Done()
				var err error

				switch svc {
				case "karaf":
					err = s.safeStartKaraf()
				case "postgresql":
					err = s.safeStartPostgreSQL()
				case "keycloak":
					err = s.safeStartKeycloak()
				}

				mu.Lock()
				results[svc] = err
				mu.Unlock()

				if err != nil {
					log.Printf("%s service failed to start: %v", svc, err)
				} else {
					log.Printf("%s service start completed successfully", svc)
				}
			}(service)
		}

		wg.Wait()

		// Log overall results
		for svc, err := range results {
			if err != nil {
				log.Printf("Parallel start: %s failed: %v", svc, err)
			} else {
				log.Printf("Parallel start: %s succeeded", svc)
			}
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service": "all",
		"status":  "starting",
		"message": "All services starting in parallel...",
	})
}

func (s *Server) handleServicesStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	config.LoadProperties()
	_ = config.GetConfig()

	// Stop all services in parallel
	go func() {
		var wg sync.WaitGroup
		services := []string{"karaf", "postgresql", "keycloak"}
		results := make(map[string]error)
		var mu sync.Mutex

		for _, service := range services {
			wg.Add(1)
			go func(svc string) {
				defer wg.Done()
				var err error

				switch svc {
				case "karaf":
					err = s.safeStopKaraf()
				case "postgresql":
					err = s.safeStopPostgreSQL()
				case "keycloak":
					err = s.safeStopKeycloak()
				}

				mu.Lock()
				results[svc] = err
				mu.Unlock()

				if err != nil {
					log.Printf("%s service failed to stop: %v", svc, err)
				} else {
					log.Printf("%s service stop completed successfully", svc)
				}
			}(service)
		}

		wg.Wait()

		// Log overall results
		for svc, err := range results {
			if err != nil {
				log.Printf("Parallel stop: %s failed: %v", svc, err)
			} else {
				log.Printf("Parallel stop: %s succeeded", svc)
			}
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service": "all",
		"status":  "stopping",
		"message": "All services stopping in parallel...",
	})
}

// Concurrent service status handler
func (s *Server) handleProjectInitStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if project is initialized by looking for common JUDO project files
	isInitialized := s.checkProjectInitialized()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"initialized": isInitialized,
		"timestamp":   time.Now().Format(time.RFC3339),
	})
}

func (s *Server) checkProjectInitialized() bool {
	// Check for common JUDO project files that indicate initialization
	projectFiles := []string{
		"judo.properties",
		"pom.xml",
		"model/TestProject.model", // Example model file
		".generated-files",
	}

	for _, file := range projectFiles {
		if _, err := os.Stat(file); err == nil {
			return true
		}
	}

	// Also check if we're in a directory that looks like a JUDO project
	if _, err := os.Stat("application"); err == nil {
		return true
	}
	if _, err := os.Stat("model"); err == nil {
		return true
	}

	return false
}

func (s *Server) handleServicesStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	config.LoadProperties()
	cfg := config.GetConfig()

	type serviceResult struct {
		Service   string `json:"service"`
		Status    string `json:"status"`
		Timestamp string `json:"timestamp"`
		Error     string `json:"error,omitempty"`
	}

	var results []serviceResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Check each service in parallel
	checkService := func(service string, checkFunc func() (string, error)) {
		defer wg.Done()
		status, err := checkFunc()
		result := serviceResult{
			Service:   service,
			Status:    status,
			Timestamp: time.Now().Format(time.RFC3339),
		}
		if err != nil {
			result.Error = err.Error()
		}
		mu.Lock()
		results = append(results, result)
		mu.Unlock()
	}

	wg.Add(3) // Karaf, PostgreSQL, Keycloak

	// Check Karaf status
	go func() {
		if cfg.Runtime == "karaf" {
			checkService("karaf", func() (string, error) {
				karafDir := filepath.Join(cfg.ModelDir, "application", ".karaf")
				if karaf.KarafRunning(karafDir) {
					return "running", nil
				}
				return "stopped", nil
			})
		} else {
			wg.Done()
		}
	}()

	// Check PostgreSQL status
	go func() {
		if cfg.DBType == "postgresql" {
			checkService("postgresql", func() (string, error) {
				pgName := "postgres-" + cfg.SchemaName
				if docker.DockerInstanceRunning(pgName) {
					return "running", nil
				}
				return "stopped", nil
			})
		} else {
			wg.Done()
		}
	}()

	// Check Keycloak status
	go func() {
		checkService("keycloak", func() (string, error) {
			kcName := "keycloak-" + cfg.KeycloakName
			if docker.DockerInstanceRunning(kcName) {
				return "running", nil
			}
			return "stopped", nil
		})
	}()

	wg.Wait()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// Service-specific stop handlers
func (s *Server) handleKarafStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	config.LoadProperties()
	_ = config.GetConfig() // Load config but don't use directly

	// Stop Karaf service
	go func() {
		if err := s.safeStopKaraf(); err != nil {
			log.Printf("Karaf service failed to stop: %v", err)
		} else {
			log.Printf("Karaf service stop completed successfully")
		}
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
	_ = config.GetConfig() // Load config but don't use directly

	// Stop PostgreSQL service
	go func() {
		if err := s.safeStopPostgreSQL(); err != nil {
			log.Printf("PostgreSQL service failed to stop: %v", err)
		} else {
			log.Printf("PostgreSQL service stop completed successfully")
		}
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
	_ = config.GetConfig() // Load config but don't use directly

	// Stop Keycloak service
	go func() {
		if err := s.safeStopKeycloak(); err != nil {
			log.Printf("Keycloak service failed to stop: %v", err)
		} else {
			log.Printf("Keycloak service stop completed successfully")
		}
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

// Safe wrapper functions for service operations
func (s *Server) safeStartKaraf() error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in safeStartKaraf: %v", r)
		}
	}()

	log.Printf("Attempting to start Karaf service...")
	if err := karaf.StartKaraf(); err != nil {
		log.Printf("Karaf service failed to start: %v", err)
		return fmt.Errorf("karaf service failed to start: %v", err)
	}
	log.Printf("Karaf service started successfully")
	return nil
}

func (s *Server) safeStartPostgreSQL() error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in safeStartPostgreSQL: %v", r)
		}
	}()

	log.Printf("Attempting to start PostgreSQL service...")
	if err := docker.StartPostgres(); err != nil {
		log.Printf("PostgreSQL service failed to start: %v", err)
		return fmt.Errorf("postgresql service failed to start: %v", err)
	}
	log.Printf("PostgreSQL service started successfully")
	return nil
}

func (s *Server) safeStartKeycloak() error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in safeStartKeycloak: %v", r)
		}
	}()

	log.Printf("Attempting to start Keycloak service...")
	if err := docker.StartKeycloak(); err != nil {
		log.Printf("Keycloak service failed to start: %v", err)
		return fmt.Errorf("keycloak service failed to start: %v", err)
	}
	log.Printf("Keycloak service started successfully")
	return nil
}

func (s *Server) safeStopKaraf() error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in safeStopKaraf: %v", r)
		}
	}()

	config.LoadProperties()
	cfg := config.GetConfig()
	log.Printf("Attempting to stop Karaf service...")
	karaf.StopKaraf(cfg.KarafDir)
	log.Printf("Karaf service stop attempted")
	return nil
}

func (s *Server) safeStopPostgreSQL() error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in safeStopPostgreSQL: %v", r)
		}
	}()

	config.LoadProperties()
	cfg := config.GetConfig()
	log.Printf("Attempting to stop PostgreSQL service...")
	pgName := "postgres-" + cfg.SchemaName
	if err := docker.StopDockerInstance(pgName); err != nil {
		log.Printf("PostgreSQL service failed to stop: %v", err)
		return fmt.Errorf("postgresql service failed to stop: %v", err)
	}
	log.Printf("PostgreSQL service stopped successfully")
	return nil
}

func (s *Server) safeStopKeycloak() error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in safeStopKeycloak: %v", r)
		}
	}()

	config.LoadProperties()
	cfg := config.GetConfig()
	log.Printf("Attempting to stop Keycloak service...")
	kcName := "keycloak-" + cfg.KeycloakName
	if err := docker.StopDockerInstance(kcName); err != nil {
		log.Printf("Keycloak service failed to stop: %v", err)
		return fmt.Errorf("keycloak service failed to stop: %v", err)
	}
	log.Printf("Keycloak service stopped successfully")
	return nil
}

// WebSocket handlers for specific log streams
func (s *Server) handleCombinedLogsWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Add to connection pool
	if !s.logPool.add(conn) {
		log.Printf("Log connection pool full, rejecting connection")
		conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "Connection pool full"))
		return
	}
	defer s.logPool.remove(conn)

	// Stream combined logs from all services
	go s.streamCombinedLogs(conn)

	// Keep connection alive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (s *Server) handleKarafLogsWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Add to connection pool
	if !s.logPool.add(conn) {
		log.Printf("Log connection pool full, rejecting connection")
		conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "Connection pool full"))
		return
	}
	defer s.logPool.remove(conn)

	// Stream Karaf logs only
	go s.streamServiceLogs(conn, "karaf")

	// Keep connection alive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (s *Server) handlePostgreSQLLogsWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Add to connection pool
	if !s.logPool.add(conn) {
		log.Printf("Log connection pool full, rejecting connection")
		conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "Connection pool full"))
		return
	}
	defer s.logPool.remove(conn)

	// Stream PostgreSQL logs only
	go s.streamServiceLogs(conn, "postgresql")

	// Keep connection alive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (s *Server) handleKeycloakLogsWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Add to connection pool
	if !s.logPool.add(conn) {
		log.Printf("Log connection pool full, rejecting connection")
		conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "Connection pool full"))
		return
	}
	defer s.logPool.remove(conn)

	// Stream Keycloak logs only
	go s.streamServiceLogs(conn, "keycloak")

	// Keep connection alive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// handleServiceLogsWebSocket handles dynamic service log streaming at /ws/logs/service/{service}
func (s *Server) handleServiceLogsWebSocket(w http.ResponseWriter, r *http.Request) {
	service := strings.TrimPrefix(r.URL.Path, "/ws/logs/service/")
	service = strings.TrimSuffix(service, "/")
	if service == "" {
		http.Error(w, "service not specified", http.StatusBadRequest)
		return
	}
	// Validate known services
	switch service {
	case "karaf", "postgresql", "keycloak":
		// ok
	default:
		http.Error(w, "unknown service", http.StatusNotFound)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	if !s.logPool.add(conn) {
		log.Printf("Log connection pool full, rejecting connection")
		conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "Connection pool full"))
		return
	}
	defer s.logPool.remove(conn)

	go s.streamServiceLogs(conn, service)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (s *Server) handleSessionWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Session WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Add to session connection pool
	if !s.sessionPool.add(conn) {
		log.Printf("Session connection pool full, rejecting connection")
		conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "Session pool full"))
		return
	}
	defer s.sessionPool.remove(conn)

	// Send handshake message with session information
	handshake := map[string]interface{}{
		"type":    "handshake",
		"version": "1.0",
		"session": "judo",
		"welcome": "JUDO Interactive Session - Connected\r\nType 'help' for available commands\r\n",
	}
	handshakeMsg, _ := json.Marshal(handshake)
	conn.WriteMessage(websocket.TextMessage, handshakeMsg)

	// Start interactive session after a brief delay to ensure handshake is processed
	time.Sleep(100 * time.Millisecond)
	go s.handleInteractiveSession(conn)

	// Handle client messages
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		if messageType == websocket.TextMessage {
			// Handle session input
			go s.handleSessionInput(conn, string(message))
		}
	}
}

func (s *Server) streamCombinedLogs(conn *websocket.Conn) {
	// Use API to stream real combined logs
	config.LoadProperties()
	cfg := config.GetConfig()

	// Create a channel for log output
	logChan := make(chan string, 100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Stream logs from various services based on configuration
	if cfg.Runtime == "karaf" {
		karafLogFile := filepath.Join(cfg.KarafDir, "console.out")
		if _, err := os.Stat(karafLogFile); err == nil {
			go s.api.streamLogFile(karafLogFile, "[KARAF]", ctx, logChan)
		}
	}

	// Stream Docker container logs
	if cfg.DBType == "postgresql" {
		pgName := "postgres-" + cfg.SchemaName
		go s.api.streamDockerLogs(pgName, "[POSTGRESQL]", ctx, logChan)
	}

	kcName := "keycloak-" + cfg.KeycloakName
	go s.api.streamDockerLogs(kcName, "[KEYCLOAK]", ctx, logChan)

	// Send connection established message as JSON
	init := map[string]interface{}{
		"ts":      time.Now().Format(time.RFC3339),
		"service": "combined",
		"line":    "Log stream connected",
	}
	msg, _ := json.Marshal(init)
	conn.WriteMessage(websocket.TextMessage, msg)

	// Stream logs to WebSocket
	for {
		select {
		case logLine := <-logChan:
			logMessage := map[string]interface{}{
				"ts":      time.Now().Format(time.RFC3339),
				"service": "combined",
				"line":    logLine,
			}
			message, _ := json.Marshal(logMessage)
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-time.After(5 * time.Second):
			// Send empty heartbeat to keep connection alive (no message content)
			heartbeat := map[string]interface{}{
				"ts":      time.Now().Format(time.RFC3339),
				"service": "heartbeat",
				"line":    "",
			}
			message, _ := json.Marshal(heartbeat)
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		}
	}
}

func (s *Server) streamServiceLogs(conn *websocket.Conn, service string) {
	// Use API to stream real service logs
	config.LoadProperties()
	cfg := config.GetConfig()

	// Create a channel for log output
	logChan := make(chan string, 100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Stream logs based on service type
	switch service {
	case "karaf":
		if cfg.Runtime == "karaf" {
			karafLogFile := filepath.Join(cfg.KarafDir, "console.out")
			if _, err := os.Stat(karafLogFile); err == nil {
				go s.api.streamLogFile(karafLogFile, "[KARAF]", ctx, logChan)
			}
		}
	case "postgresql":
		// PostgreSQL logs streamed from docker logs
		config.LoadProperties()
		cfg := config.GetConfig()
		if cfg.DBType == "postgresql" {
			pgName := "postgres-" + cfg.SchemaName
			go s.api.streamDockerLogs(pgName, "[POSTGRESQL]", ctx, logChan)
		}
	case "keycloak":
		// Keycloak logs streamed from docker logs
		config.LoadProperties()
		cfg := config.GetConfig()
		kcName := "keycloak-" + cfg.KeycloakName
		go s.api.streamDockerLogs(kcName, "[KEYCLOAK]", ctx, logChan)
	}

	// Send connection established message as JSON
	init := map[string]interface{}{
		"ts":      time.Now().Format(time.RFC3339),
		"service": service,
		"line":    fmt.Sprintf("%s log stream connected", strings.ToUpper(service)),
	}
	msg, _ := json.Marshal(init)
	conn.WriteMessage(websocket.TextMessage, msg)

	// Stream logs to WebSocket
	for {
		select {
		case logLine := <-logChan:
			logMessage := map[string]interface{}{
				"ts":      time.Now().Format(time.RFC3339),
				"service": service,
				"line":    logLine,
			}
			message, _ := json.Marshal(logMessage)
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-time.After(5 * time.Second):
			// Send heartbeat to keep connection alive
			heartbeat := map[string]interface{}{
				"ts":      time.Now().Format(time.RFC3339),
				"service": service,
				"line":    fmt.Sprintf("%s log stream active", strings.ToUpper(service)),
			}
			message, _ := json.Marshal(heartbeat)
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		}
	}
}

func (s *Server) handleInteractiveSession(conn *websocket.Conn) {
	// Send welcome message for interactive session
	welcomeMsg := map[string]interface{}{
		"type": "output",
		"data": "🚀 JUDO CLI Interactive Session\nType 'help' for available commands, 'exit' to quit\n\n",
	}
	msg, _ := json.Marshal(welcomeMsg)
	conn.WriteMessage(websocket.TextMessage, msg)

	// Send initial prompt
	promptMsg := map[string]interface{}{
		"type": "prompt",
		"data": "judo> ",
	}
	msg, _ = json.Marshal(promptMsg)
	conn.WriteMessage(websocket.TextMessage, msg)

	// Store connection for session input handling
	s.mu.Lock()
	s.sessionConn = conn
	s.mu.Unlock()

	// Send session ready message
	readyMsg := map[string]interface{}{
		"type":  "status",
		"state": "ready",
	}
	msg, _ = json.Marshal(readyMsg)
	conn.WriteMessage(websocket.TextMessage, msg)
}

func (s *Server) handleSessionInput(conn *websocket.Conn, input string) {
	// Parse input message
	var message map[string]interface{}
	if err := json.Unmarshal([]byte(input), &message); err != nil {
		log.Printf("Error parsing session input: %v", err)
		return
	}

	switch message["type"] {
	case "input":
		// Handle user input for interactive session
		if data, ok := message["data"].(string); ok {
			s.handleWebSocketSessionCommand(conn, data)
		}

	case "resize":
		// Handle terminal resize
		cols, ok1 := message["cols"].(float64)
		rows, ok2 := message["rows"].(float64)
		if ok1 && ok2 {
			log.Printf("Session resize: %dx%d", int(cols), int(rows))
		}

	case "control":
		// Handle control messages like Ctrl+C
		action, ok := message["action"].(string)
		if ok && action == "interrupt" {
			log.Printf("Session interrupt received")
			// Send interrupt message to client
			interruptMsg := map[string]interface{}{
				"type": "output",
				"data": "^C\n",
			}
			msg, _ := json.Marshal(interruptMsg)
			conn.WriteMessage(websocket.TextMessage, msg)
		}
	}
}

// handleWebSocketSessionCommand handles session commands for WebSocket connections
func (s *Server) handleWebSocketSessionCommand(conn *websocket.Conn, command string) {
	// Trim and handle empty commands
	command = strings.TrimSpace(command)
	if command == "" {
		return
	}

	// Handle session-specific commands
	switch command {
	case "exit", "quit":
		// Send exit message
		exitMsg := map[string]interface{}{
			"type": "output",
			"data": "👋 Session ended\n",
		}
		msg, _ := json.Marshal(exitMsg)
		conn.WriteMessage(websocket.TextMessage, msg)

		// Close connection
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		conn.Close()
		return

	case "help":
		helpMsg := map[string]interface{}{
			"type": "output",
			"data": "📋 Available commands:\n  help     - Show this help\n  exit     - Exit session\n  status   - Show status\n  doctor   - Run system check\n  clear    - Clear screen\n\nType any JUDO command to execute it directly\n",
		}
		msg, _ := json.Marshal(helpMsg)
		conn.WriteMessage(websocket.TextMessage, msg)

	case "clear":
		// Send clear screen command
		clearMsg := map[string]interface{}{
			"type": "output",
			"data": "\033[H\033[2J",
		}
		msg, _ := json.Marshal(clearMsg)
		conn.WriteMessage(websocket.TextMessage, msg)

	case "status":
		// Execute status command via API
		result := s.api.executeStatus()
		statusMsg := map[string]interface{}{
			"type": "output",
			"data": result + "\n",
		}
		msg, _ := json.Marshal(statusMsg)
		conn.WriteMessage(websocket.TextMessage, msg)

	case "doctor":
		// Execute doctor command via API
		result, _ := s.api.executeDoctor()
		doctorMsg := map[string]interface{}{
			"type": "output",
			"data": result + "\n",
		}
		msg, _ := json.Marshal(doctorMsg)
		conn.WriteMessage(websocket.TextMessage, msg)

	default:
		// Try to execute as a JUDO command
		args := strings.Fields(command)
		if len(args) > 0 {
			result, err := s.api.ExecuteCommand(args[0], args[1:])
			if err != nil {
				result = "Error: " + err.Error()
			}
			outputMsg := map[string]interface{}{
				"type": "output",
				"data": result + "\n",
			}
			msg, _ := json.Marshal(outputMsg)
			conn.WriteMessage(websocket.TextMessage, msg)
		}
	}

	// Send prompt after command execution
	promptMsg := map[string]interface{}{
		"type": "prompt",
		"data": "judo> ",
	}
	msg, _ := json.Marshal(promptMsg)
	conn.WriteMessage(websocket.TextMessage, msg)
}
