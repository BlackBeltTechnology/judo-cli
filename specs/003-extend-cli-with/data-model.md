# Data Model: Browser-Based Interactive CLI Server

This document defines the data structures used for communication between the Go backend and the React frontend.

## Command Execution

### Command Request
Sent from the frontend to the backend to execute a command.

**Endpoint**: `POST /api/commands`

**Body**:
```json
{
  "command": "string"
}
```

### Command Response
Sent from the backend to the frontend after a command has finished executing.

**Structure**:
```json
{
  "output": "string",
  "error": "string"
}
```

## Application Status

### Status Response
Sent from the backend to the frontend in response to a status request.

**Endpoint**: `GET /api/status`

**Structure**:
```json
{
  "status": "string"
}
```

## Log Streaming

### WebSocket Message
Sent from the backend to the frontend over the WebSocket connection.

**Structure**:
```json
{
  "type": "log",
  "payload": "string"
}
```
