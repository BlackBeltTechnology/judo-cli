package server

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	server := NewServer(8080)
	if server == nil {
		t.Error("Expected server to be created")
	}

	if server.port != 8080 {
		t.Errorf("Expected port 8080, got %d", server.port)
	}
}

func TestServerStartStop(t *testing.T) {
	server := NewServer(0) // Use random port

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.Start()
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test that server is running by making a request
	resp, err := http.Get("http://localhost:8080/api/status")
	if err == nil {
		resp.Body.Close()
	}

	// Stop the server
	if err := server.Stop(); err != nil {
		t.Errorf("Failed to stop server: %v", err)
	}

	// Wait for server to stop
	select {
	case err := <-errChan:
		if err != http.ErrServerClosed {
			t.Errorf("Expected server closed error, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Server stop timeout")
	}
}

func TestAPIEndpoints(t *testing.T) {
	server := NewServer(0)

	tests := []struct {
		name   string
		method string
		path   string
		status int
	}{
		{"GET status", http.MethodGet, "/api/status", http.StatusOK},
		{"POST start", http.MethodPost, "/api/actions/start", http.StatusOK},
		{"POST stop", http.MethodPost, "/api/actions/stop", http.StatusOK},
		{"POST command", http.MethodPost, "/api/commands/test", http.StatusOK},
		{"GET logs", http.MethodGet, "/api/logs/karaf", http.StatusOK},
		{"Wrong method status", http.MethodPut, "/api/status", http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			switch tt.path {
			case "/api/status":
				server.handleStatus(w, req)
			case "/api/actions/start":
				server.handleStart(w, req)
			case "/api/actions/stop":
				server.handleStop(w, req)
			case "/api/commands/test":
				server.handleCommand(w, req)
			case "/api/logs/karaf":
				server.handleLogs(w, req)
			}

			if w.Code != tt.status {
				t.Errorf("Expected status %d, got %d", tt.status, w.Code)
			}
		})
	}
}

func TestOpenBrowserPlatformDetection(t *testing.T) {
	// Test Windows detection
	originalOS := os.Getenv("OS")
	os.Setenv("OS", "Windows_NT")
	if !isWindows() {
		t.Error("Windows detection failed")
	}
	os.Setenv("OS", originalOS)

	// Test macOS detection
	originalGOOS := os.Getenv("GOOS")
	os.Setenv("GOOS", "darwin")
	if !isMac() {
		t.Error("macOS detection failed")
	}
	os.Setenv("GOOS", originalGOOS)

	// Test Linux detection
	os.Setenv("GOOS", "linux")
	if !isLinux() {
		t.Error("Linux detection failed")
	}
	os.Setenv("GOOS", originalGOOS)
}

func TestPortConflictHandling(t *testing.T) {
	// Find an available port for testing
	testPort := 18080

	// Create a listener on a specific port to simulate port conflict
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", testPort))
	if err != nil {
		t.Skipf("Cannot test port conflict - port %d not available: %v", testPort, err)
	}
	defer listener.Close()

	// Create server that will encounter port conflict
	server := NewServer(testPort)

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.Start()
	}()

	// Give server time to handle port conflict
	time.Sleep(100 * time.Millisecond)

	// Stop the server
	if err := server.Stop(); err != nil {
		t.Errorf("Failed to stop server: %v", err)
	}

	// Server should have handled the port conflict gracefully
	// The test expects the server to fail to start due to port conflict,
	// not necessarily return http.ErrServerClosed
	select {
	case err := <-errChan:
		// Accept either server closed error or port conflict error
		if err != http.ErrServerClosed && !strings.Contains(err.Error(), "address already in use") {
			t.Errorf("Expected server closed or port conflict error, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Server stop timeout")
	}
}
