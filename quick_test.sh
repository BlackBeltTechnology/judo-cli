#!/bin/bash

# Quick test script to verify judo server functionality

echo "=== JUDO Server Quick Test ==="

# Check if judo binary exists
if [ ! -f "judo" ]; then
    echo "❌ judo binary not found. Run ./build.sh first"
    exit 1
fi

# Check if we're in test-model directory
if [ "$(basename "$PWD")" != "test-model" ]; then
    echo "⚠️  Not in test-model directory. This test should be run from test-model/"
    echo "   Current directory: $PWD"
    echo "   Please run: cd test-model && ../quick_test.sh"
    exit 1
fi

echo "✅ In test-model directory"

# Check if judo.properties exists
if [ ! -f "judo.properties" ]; then
    echo "❌ judo.properties not found in current directory"
    exit 1
fi

echo "✅ judo.properties found"

# Start server in background
echo "Starting judo server on port 6969..."
../judo server > server.log 2>&1 &
SERVER_PID=$!

# Wait a bit for server to start
sleep 2

# Test HTTP API
echo "Testing HTTP API..."
if curl -s http://localhost:6969/api/status > /dev/null; then
    echo "✅ HTTP API is responding"
else
    echo "❌ HTTP API not responding"
    kill $SERVER_PID 2>/dev/null
    exit 1
fi

# Test WebSocket connectivity (simple handshake test)
echo "Testing WebSocket connectivity..."
if python3 -c "
import socket
s = socket.socket()
s.settimeout(2)
try:
    s.connect(('localhost', 6969))
    s.send(b'GET /ws/logs/combined HTTP/1.1\\r\\nHost: localhost:6969\\r\\nUpgrade: websocket\\r\\nConnection: Upgrade\\r\\nSec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\\r\\nSec-WebSocket-Version: 13\\r\\n\\r\\n')
    response = s.recv(1024).decode()
    if '101' in response:
        print('✅ WebSocket handshake successful')
    else:
        print('❌ WebSocket handshake failed')
        print('Response:', response)
except Exception as e:
    print('❌ WebSocket test failed:', e)
finally:
    s.close()
"; then
    echo "WebSocket test completed"
else
    echo "WebSocket test failed"
fi

# Stop server
kill $SERVER_PID 2>/dev/null
echo "Server stopped"

echo "=== Test completed ==="