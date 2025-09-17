package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestHandleWebSocket(t *testing.T) {
	server := NewServer(0)
	s := httptest.NewServer(http.HandlerFunc(server.handleWebSocket))
	defer s.Close()

	// Convert http:// to ws://
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// Connect to the server
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer ws.Close()

	// This test just checks that the connection is established.
	// More advanced tests could check for messages.
	assert.NotNil(t, ws)
}

func TestHandleCombinedLogsWebSocket(t *testing.T) {
	server := NewServer(0)
	s := httptest.NewServer(http.HandlerFunc(server.handleCombinedLogsWebSocket))
	defer s.Close()

	// Convert http:// to ws://
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// Connect to the server
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer ws.Close()

	// This test just checks that the connection is established.
	// More advanced tests could check for messages.
	assert.NotNil(t, ws)
}

func TestHandleServiceLogsWebSocket(t *testing.T) {
	server := NewServer(0)
	mux := http.NewServeMux()
	mux.HandleFunc("/ws/logs/service/", server.handleServiceLogsWebSocket)
	s := httptest.NewServer(mux)
	defer s.Close()

	// Convert http:// to ws://
	u := "ws" + strings.TrimPrefix(s.URL, "http") + "/ws/logs/service/karaf"

	// Connect to the server
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer ws.Close()

	// This test just checks that the connection is established.
	// More advanced tests could check for messages.
	assert.NotNil(t, ws)
}

func TestHandleSessionWebSocket(t *testing.T) {
	server := NewServer(0)
	s := httptest.NewServer(http.HandlerFunc(server.handleSessionWebSocket))
	defer s.Close()

	// Convert http:// to ws://
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// Connect to the server
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer ws.Close()

	// This test just checks that the connection is established.
	// More advanced tests could check for messages.
	assert.NotNil(t, ws)
}
