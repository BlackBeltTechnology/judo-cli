# WebSocket Testing Solution

## Problem Identified

The judo server doesn't work when run from the root directory because:
- Configuration loads based on current working directory (`os.Getwd()`)
- When run from root, it looks for config in `/Users/robson/Project/judo-cli`
- When run from `test-model`, it finds proper `judo.properties` and application structure

## Solution Created

### 1. Proper Test Scripts

**`test_websocket_proper.js`** - Node.js test using `ws` library (recommended)
```bash
npm install ws
node test_websocket_proper.js
```

**`test_websocket_simple.py`** - Python test using only standard library
```bash
python3 test_websocket_simple.py
```

### 2. Quick Test Script

**`quick_test.sh`** - Comprehensive test script
```bash
cd test-model
../quick_test.sh
```

### 3. Documentation

**`TESTING.md`** - Complete testing guide
**`WEBSOCKET_TESTING_SUMMARY.md`** - This summary

## Testing Procedure

1. **Build the binary**: `./build.sh`
2. **Change to test-model**: `cd test-model`
3. **Run server**: `../judo server`
4. **Test WebSockets**: Use one of the test scripts

## Why curl doesn't work

curl doesn't support WebSocket protocol:
- Requires HTTP Upgrade handshake
- Protocol switching
- Frame-based messaging
- curl is designed for HTTP/1.1 only

## Alternative Tools

- `wscat` (`npm install -g wscat`)
- Browser developer tools
- Postman (WebSocket support)
- Proper WebSocket clients like the test scripts provided

## Key Endpoints

- `ws://localhost:6969/ws/logs/combined` - Combined logs
- `ws://localhost:6969/ws/logs/service/karaf` - Karaf logs
- `ws://localhost:6969/ws/session` - Interactive session

## Configuration Dependency

The server requires:
- `judo.properties` in current directory
- `application/` directory structure
- Proper service configuration

This is why running from `test-model` directory is essential.