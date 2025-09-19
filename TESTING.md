# Testing JUDO CLI WebSocket Functionality

## Problem Description

The judo server doesn't work properly when run from the root directory because:
1. Configuration is loaded based on current working directory (`os.Getwd()`)
2. When run from root, it looks for configuration in `/Users/robson/Project/judo-cli`
3. When run from `test-model`, it finds the proper `judo.properties` and configuration

## Proper Testing Procedure

### 1. Build the Binary

First, build the complete binary with embedded frontend:

```bash
./build.sh
```

This will:
- Build the React frontend
- Copy assets for embedding
- Build the Go binary with version information
- Run tests

### 2. Run from test-model Directory

Change to the test-model directory and run the server:

```bash
cd test-model
../judo server
```

### 3. Test WebSocket Connections

Use one of the following test methods:

#### Option A: Node.js with ws library (Recommended)

```bash
# Install ws library if not already installed
npm install ws

# Run the test
node ../test_websocket_proper.js
```

#### Option B: Python with websockets library

```bash
# Install websockets library if not already installed
pip install websockets

# Run the test
python3 ../test_websocket_proper.py
```

#### Option C: Simple connectivity test (no dependencies)

```bash
# Test basic WebSocket connectivity
python3 ../test_websocket_simple.py
```

### 4. Expected Endpoints

- `ws://localhost:6969/ws/logs/combined` - Combined logs from all services
- `ws://localhost:6969/ws/logs/service/karaf` - Karaf-specific logs
- `ws://localhost:6969/ws/session` - Interactive session

## Why curl doesn't work

curl doesn't support the WebSocket protocol properly. WebSocket requires:
1. HTTP Upgrade handshake
2. Protocol switching
3. Frame-based messaging

curl is designed for HTTP/1.1, not WebSocket protocol.

## Alternative Testing Tools

- `wscat` - WebSocket cat utility (`npm install -g wscat`)
- Browser developer tools - Open browser and use WebSocket console
- Postman - Has WebSocket support in newer versions

## Troubleshooting

If you see empty responses or connection issues:

1. **Check server is running**: `curl http://localhost:6969/api/status`
2. **Verify port 6969 is available**: `lsof -i :6969`
3. **Check test-model directory**: Ensure you're in `test-model` when running server
4. **Verify build**: Make sure `./build.sh` completed successfully

## Configuration Dependency

The server depends on finding proper configuration files:
- `judo.properties` - Main configuration
- `application/` directory structure
- Karaf logs and Docker container names

This is why running from the correct directory (`test-model`) is essential.