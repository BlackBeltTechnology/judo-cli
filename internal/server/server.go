package server

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/gorilla/websocket"
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
	mux.HandleFunc("/ws/logs", s.handleWebSocket)
	// Serve embedded frontend files
	assetsFS, _ := fs.Sub(embeddedFiles, "assets")
	mux.Handle("/", http.FileServer(http.FS(assetsFS)))

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	return s
}

func (s *Server) Start() error {
	log.Printf("Starting server on port %d", s.port)
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

	// TODO: Execute the command and capture output
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"command": "` + command + `", "output": "Command execution not yet implemented", "success": true}`))
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	service := r.URL.Path[len("/api/logs/"):]

	// TODO: Return log content for the specified service
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"service": "` + service + `", "logs": "Log streaming not yet implemented"}`))
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

	// TODO: Send log messages to client
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}
		log.Printf("Received WebSocket message: %s", message)
	}

	s.mu.Lock()
	delete(s.clients, conn)
	s.mu.Unlock()
}

// handleStatic is no longer needed as we use embedded filesystem

func isWindows() bool {
	return os.Getenv("OS") == "Windows_NT" || os.Getenv("GOOS") == "windows"
}

func isMac() bool {
	return os.Getenv("GOOS") == "darwin"
}

func isLinux() bool {
	return os.Getenv("GOOS") == "linux"
}
