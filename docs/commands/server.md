---
layout: page
title: Server Command
parent: Command Reference
nav_order: 10
description: "Start JUDO CLI web server with browser-based interface"
permalink: /commands/server/
---

# Server Command

Start a web server that provides a graphical interface for JUDO CLI operations.

## Usage

```bash
judo server [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `-p, --port <port>` | Port to run the server on (default: 6969) |

## Description

The `server` command starts a web server that provides a browser-based interface for JUDO CLI operations. This is particularly useful for:

- **Visual Project Management**: Manage JUDO projects through an intuitive web interface
- **Interactive Command Execution**: Execute commands with visual feedback and progress indicators
- **Real-time Monitoring**: Monitor service status, logs, and resource usage in real-time
- **Database Operations**: Perform database backups, restores, and schema upgrades visually
- **Log Analysis**: View and filter application logs with search and filtering capabilities

## Features

### Dashboard
- Project overview with health status
- Service status indicators (Karaf, PostgreSQL, Keycloak)
- Quick access to common operations

### Command Interface
- Visual command builder with parameter assistance
- Command history and results
- Real-time output streaming

### Database Management
- Database backup and restore operations
- Schema upgrade visualization
- Data export/import tools

### Log Viewer
- Real-time log streaming
- Search and filtering capabilities
- Log level filtering
- Export functionality

## Examples

### Start server on default port
```bash
judo server
```

### Start server on custom port
```bash
judo server --port 8080
judo server -p 9000
```

### Accessing the Web Interface
Once the server is running, open your web browser and navigate to:
```
http://localhost:6969
```

Replace `6969` with your custom port if using the `-p` flag.

## Technical Details

- **Port**: Default port 6969, configurable via `-p` flag
- **Protocol**: HTTP with REST API endpoints
- **Frontend**: React-based single page application
- **Authentication**: None (runs on localhost only for security)
- **Persistence**: Session-based, no permanent data storage

## Security Considerations

The server runs on localhost only and does not require authentication. This is intentional for development use:

- Only accessible from the local machine
- No sensitive data is stored by the server
- All operations use the same permissions as the CLI
- Recommended for development environments only

For production use, consider implementing proper authentication and authorization mechanisms.

## Integration with Other Commands

The server integrates with all JUDO CLI commands, providing visual interfaces for:

- `judo build` - Visual build progress and results
- `judo start`/`judo stop` - Service control with status indicators
- `judo doctor` - System health dashboard
- `judo dump`/`judo import` - Database management tools
- `judo log` - Enhanced log viewing capabilities

## Troubleshooting

### Port Already in Use
If the default port is already in use:
```bash
# Use a different port
judo server -p 7070
```

### Browser Connection Issues
Ensure the server is running and check the port:
```bash
# Check if server is running
judo status

# Verify port availability
netstat -an | grep 6969
```

### Server Not Starting
Check for errors in the console output and ensure all dependencies are available.

## Related Commands

- [`session`]({{ site.baseurl }}/commands/session/) - Interactive terminal session
- [`status`]({{ site.baseurl }}/commands/status/) - Check service status
- [`log`]({{ site.baseurl }}/commands/log/) - View application logs