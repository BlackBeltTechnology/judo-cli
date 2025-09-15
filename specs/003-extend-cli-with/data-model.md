# Data Model: Browser-Based Interactive CLI Server

This document defines the data structures used for communication between the Go backend and the React frontend.

## Interactive Session (Terminal B)

Terminal B provides a browser-based interactive `judo session` terminal over WebSocket.

### Endpoint
- `WS /ws/session`

### Client → Server Messages
```json
{ "type": "input", "data": "string (UTF-8)" }
{ "type": "resize", "cols": 120, "rows": 30 }
{ "type": "control", "action": "interrupt" } // e.g., Ctrl+C
```

### Server → Client Messages
```json
{ "type": "output", "data": "string (UTF-8)" }
{ "type": "status", "state": "running|exited", "exitCode": 0 }
```

Notes:
- One session per WS connection. Session starts on connect and ends on process exit or disconnect.
- Input and output are streamed; no buffering beyond minimal batching.
- The server maps `control.interrupt` to an interrupt signal for the session process.

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
  "ts": "ISO-8601 string (UTC)",
  "service": "karaf | postgresql | keycloak",
  "line": "string"
}
```

Notes:
- `ts` is in UTC and represents when the line was emitted.
- `service` identifies the log source; values are strictly one of the three supported services.
- `line` contains the raw message text; if the source already includes a timestamp, the server SHOULD strip it to avoid duplication.
