package server

import (
	"bytes"
	"context"
	"crypto/tls"
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
		logPool:     newConnectionPool(100), // Max 100 log connections
		sessionPool: newConnectionPool(50),  // Max 50 session connections
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
	mux.HandleFunc("/api/services/start", s.handleServicesStart)
	mux.HandleFunc("/api/services/stop", s.handleServicesStop)
	mux.HandleFunc("/api/services/status", s.handleServicesStatus)
	mux.HandleFunc("/api/project/init/status", s.handleProjectInitStatus)
	mux.HandleFunc("/ws/logs", s.handleWebSocket)
	mux.HandleFunc("/ws/logs/combined", s.handleCombinedLogsWebSocket)
	mux.HandleFunc("/ws/logs/service/karaf", s.handleKarafLogsWebSocket)
	mux.HandleFunc("/ws/logs/service/postgresql", s.handlePostgreSQLLogsWebSocket)
	mux.HandleFunc("/ws/logs/service/keycloak", s.handleKeycloakLogsWebSocket)
	mux.HandleFunc("/ws/session", s.handleSessionWebSocket)
	// Serve embedded frontend files
	assetsFS, _ := fs.Sub(embeddedFiles, "assets")
	mux.Handle("/", http.FileServer(http.FS(assetsFS)))

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

	// Start interactive session
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
	// Implementation for streaming combined logs from all services
	// This would tail logs from Karaf, PostgreSQL, and Keycloak
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Simulate combined log streaming
		logMessage := map[string]interface{}{
			"ts":      time.Now().Format(time.RFC3339),
			"service": "combined",
			"line":    fmt.Sprintf("Combined log message at %s", time.Now().Format(time.RFC3339)),
		}

		message, _ := json.Marshal(logMessage)

		// Broadcast to all connections in the log pool
		s.logPool.broadcast(message)

		// Also send to this specific connection
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			break
		}
	}
}

func (s *Server) streamServiceLogs(conn *websocket.Conn, service string) {
	// Implementation for streaming logs from a specific service
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Simulate service-specific log streaming
		logMessage := map[string]interface{}{
			"ts":      time.Now().Format(time.RFC3339),
			"service": service,
			"line":    fmt.Sprintf("[%s] Log message at %s", strings.ToUpper(service), time.Now().Format(time.RFC3339)),
		}

		message, _ := json.Marshal(logMessage)
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			break
		}
	}
}

func (s *Server) handleInteractiveSession(conn *websocket.Conn) {
	// Start an interactive judo session
	cmd := exec.Command("./judo", "session")
	cmd.Dir = "."

	// Create pipes for input/output
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Printf("Error creating stdin pipe: %v", err)
		return
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("Error creating stdout pipe: %v", err)
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("Error creating stderr pipe: %v", err)
		return
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		log.Printf("Error starting session: %v", err)
		return
	}

	// Stream stdout to WebSocket
	go func() {
		buffer := make([]byte, 1024)
		for {
			n, err := stdout.Read(buffer)
			if err != nil {
				break
			}
			if n > 0 {
				message := map[string]interface{}{
					"type": "output",
					"data": string(buffer[:n]),
				}
				msg, _ := json.Marshal(message)
				conn.WriteMessage(websocket.TextMessage, msg)
			}
		}
	}()

	// Stream stderr to WebSocket
	go func() {
		buffer := make([]byte, 1024)
		for {
			n, err := stderr.Read(buffer)
			if err != nil {
				break
			}
			if n > 0 {
				message := map[string]interface{}{
					"type": "output",
					"data": string(buffer[:n]),
				}
				msg, _ := json.Marshal(message)
				conn.WriteMessage(websocket.TextMessage, msg)
			}
		}
	}()

	// Handle input from WebSocket to stdin
	go func() {
		// Store stdin for later use by handleSessionInput
		s.mu.Lock()
		s.sessionStdin = stdin
		s.mu.Unlock()

		// Keep this goroutine alive to prevent stdin from being closed
		select {}
	}()

	// Wait for command to complete
	go func() {
		err := cmd.Wait()
		exitCode := 0
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			}
		}

		message := map[string]interface{}{
			"type":     "status",
			"state":    "exited",
			"exitCode": exitCode,
		}
		msg, _ := json.Marshal(message)
		conn.WriteMessage(websocket.TextMessage, msg)

		// Clean up stdin reference
		s.mu.Lock()
		s.sessionStdin = nil
		s.mu.Unlock()
	}()
}

func (s *Server) handleSessionInput(conn *websocket.Conn, input string) {
	// Parse input message
	var message map[string]interface{}
	if err := json.Unmarshal([]byte(input), &message); err != nil {
		log.Printf("Error parsing session input: %v", err)
		return
	}

	// Handle different message types
	s.mu.Lock()
	stdin := s.sessionStdin
	s.mu.Unlock()

	if stdin == nil {
		log.Printf("No active session stdin available")
		return
	}

	switch message["type"] {
	case "input":
		// Send input to session process
		if data, ok := message["data"].(string); ok {
			if _, err := stdin.Write([]byte(data)); err != nil {
				log.Printf("Error writing to session stdin: %v", err)
			}
		}

	case "resize":
		// Handle terminal resize (would need PTY for this)
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
			// Send Ctrl+C equivalent
			if _, err := stdin.Write([]byte{0x03}); err != nil {
				log.Printf("Error sending interrupt: %v", err)
			}
		}
	}
}
