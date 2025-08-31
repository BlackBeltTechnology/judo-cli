---
title: "API Reference"
permalink: /api/
excerpt: "Technical reference for JUDO CLI internals and extension points"
---

# API Reference

Technical reference for JUDO CLI internals, configuration schemas, and extension points.

## Configuration Schema

### Application Configuration

```yaml
# judo.properties schema
app:
  name: string          # Application name (required)
  schema: string        # Database schema name
  version: string       # Application version
  description: string   # Application description

database:
  type: enum           # postgresql | hsqldb
  host: string         # Database hostname
  port: integer        # Database port (1-65535)
  name: string         # Database name
  username: string     # Database username
  password: string     # Database password
  schema: string       # Database schema

runtime:
  mode: enum           # karaf | compose
  
karaf:
  port: integer        # HTTP port (1-65535)
  debug:
    port: integer      # Debug port
    enabled: boolean   # Enable debug mode
  ssh:
    port: integer      # SSH port
    enabled: boolean   # Enable SSH access

services:
  keycloak:
    port: integer      # Keycloak port
    realm: string      # Keycloak realm name
    enabled: boolean   # Enable Keycloak
  postgres:
    port: integer      # PostgreSQL port
    version: string    # PostgreSQL version

build:
  skip:
    tests: boolean     # Skip tests during build
  parallel: boolean    # Enable parallel build
  maven:
    profile: string    # Maven profile to use
```

## Environment Variables

All configuration properties can be overridden using environment variables with the pattern:
`JUDO_` + property path in uppercase with dots replaced by underscores.

### Examples

| Property | Environment Variable | Example |
|----------|---------------------|---------|
| `app.name` | `JUDO_APP_NAME` | `JUDO_APP_NAME=myapp` |
| `database.host` | `JUDO_DATABASE_HOST` | `JUDO_DATABASE_HOST=localhost` |
| `karaf.port` | `JUDO_KARAF_PORT` | `JUDO_KARAF_PORT=8080` |
| `build.skip.tests` | `JUDO_BUILD_SKIP_TESTS` | `JUDO_BUILD_SKIP_TESTS=true` |

## Command Exit Codes

| Code | Meaning | Description |
|------|---------|-------------|
| 0 | Success | Command completed successfully |
| 1 | General Error | Command failed with error |
| 2 | Misuse | Invalid command usage or arguments |
| 3 | Configuration Error | Invalid or missing configuration |
| 4 | Service Error | Service start/stop/status error |
| 5 | Build Error | Build or compilation error |
| 6 | Database Error | Database operation error |
| 130 | Interrupted | Command interrupted by user (Ctrl+C) |

## Status Codes

### Service Status

```json
{
  "karaf": {
    "status": "running|stopped|error",
    "port": 8080,
    "pid": 12345,
    "uptime": "2h 15m",
    "health": "healthy|unhealthy|unknown"
  },
  "postgres": {
    "status": "running|stopped|error", 
    "port": 5432,
    "container_id": "abc123",
    "health": "healthy|unhealthy|unknown"
  },
  "keycloak": {
    "status": "running|stopped|error",
    "port": 8180,
    "container_id": "def456", 
    "health": "healthy|unhealthy|unknown"
  }
}
```

### Build Status

```json
{
  "build": {
    "status": "success|failed|in_progress",
    "duration": "45s",
    "modules": [
      {
        "name": "app",
        "status": "success|failed|skipped",
        "duration": "30s"
      },
      {
        "name": "frontend", 
        "status": "success|failed|skipped",
        "duration": "15s"
      }
    ]
  }
}
```

## File Locations

### Configuration Files

| File | Location | Purpose |
|------|----------|---------|
| `judo.properties` | Project root | Main configuration |
| `{env}.properties` | Project root | Environment overrides |
| `judo-version.properties` | Project root | Version constraints |

### Runtime Files

| File | Location | Purpose |
|------|----------|---------|
| `.judo/` | Project root | Runtime data directory |
| `.judo/state.json` | Project root | Current state tracking |
| `.judo/logs/` | Project root | Application logs |
| `.judo/temp/` | Project root | Temporary files |

### System Files

| File | Location | Purpose |
|------|----------|---------|
| `~/.judo/config` | User home | Global configuration |
| `~/.judo/cache/` | User home | Download cache |
| `/usr/local/bin/judo` | System | Binary installation |

## Plugin System

### Plugin Interface

```go
type Plugin interface {
    Name() string
    Version() string
    Description() string
    
    // Lifecycle hooks
    Init(config Config) error
    PreBuild(context BuildContext) error
    PostBuild(context BuildContext) error
    PreStart(context StartContext) error
    PostStart(context StartContext) error
    PreStop(context StopContext) error
    PostStop(context StopContext) error
    
    // Commands
    Commands() []Command
}
```

### Plugin Configuration

```yaml
# Plugin configuration in judo.properties
plugins:
  enabled: true
  directory: .judo/plugins
  
  custom-plugin:
    enabled: true
    config:
      option1: value1
      option2: value2
```

## Docker Integration

### Container Management

```json
{
  "containers": {
    "postgres": {
      "image": "postgres:14",
      "name": "judo-postgres",
      "ports": ["5432:5432"],
      "environment": {
        "POSTGRES_DB": "judo",
        "POSTGRES_USER": "judo",
        "POSTGRES_PASSWORD": "judo"
      },
      "volumes": ["judo-postgres-data:/var/lib/postgresql/data"]
    },
    "keycloak": {
      "image": "quay.io/keycloak/keycloak:21.1.1",
      "name": "judo-keycloak",
      "ports": ["8180:8080"],
      "environment": {
        "KEYCLOAK_ADMIN": "admin",
        "KEYCLOAK_ADMIN_PASSWORD": "admin"
      },
      "depends_on": ["postgres"]
    }
  }
}
```

### Volume Management

```json
{
  "volumes": {
    "judo-postgres-data": {
      "driver": "local",
      "labels": {
        "app": "judo",
        "service": "postgres"
      }
    },
    "judo-app-logs": {
      "driver": "local",
      "labels": {
        "app": "judo",
        "service": "app"
      }
    }
  }
}
```

## Maven Integration

### Build Profiles

```xml
<!-- Default build profiles used by JUDO CLI -->
<profiles>
  <profile>
    <id>development</id>
    <properties>
      <skipTests>true</skipTests>
      <frontend.build>true</frontend.build>
    </properties>
  </profile>
  
  <profile>
    <id>production</id>
    <properties>
      <skipTests>false</skipTests>
      <optimization.enabled>true</optimization.enabled>
    </properties>
  </profile>
  
  <profile>
    <id>ci</id>
    <properties>
      <skipTests>false</skipTests>
      <parallel.build>true</parallel.build>
    </properties>
  </profile>
</profiles>
```

### Property Injection

JUDO CLI injects properties into Maven builds:

```xml
<properties>
  <!-- Injected by JUDO CLI -->
  <judo.app.name>${app.name}</judo.app.name>
  <judo.app.version>${app.version}</judo.app.version>
  <judo.database.type>${database.type}</judo.database.type>
  <judo.runtime.mode>${runtime.mode}</judo.runtime.mode>
</properties>
```

## Health Checks

### Service Health Endpoints

| Service | Endpoint | Expected Response |
|---------|----------|-------------------|
| Karaf | `http://localhost:8080/health` | `200 OK` |
| Keycloak | `http://localhost:8180/health` | `200 OK` |
| PostgreSQL | `pg_isready -h localhost -p 5432` | Exit code 0 |

### Health Check Configuration

```properties
# Health check settings
health.check.enabled=true
health.check.timeout=30s
health.check.interval=5s
health.check.retries=6
```

## Logging Configuration

### Log Levels

| Level | Usage | Example |
|-------|-------|---------|
| ERROR | Error conditions | Service startup failures |
| WARN | Warning conditions | Port conflicts, deprecated usage |
| INFO | Informational | Command execution, status changes |
| DEBUG | Debug information | Detailed execution flow |
| TRACE | Trace information | Very detailed debugging |

### Log Configuration

```properties
# Logging configuration
logging.level=INFO
logging.file=.judo/logs/judo.log
logging.max.size=10MB
logging.max.files=5
logging.pattern=%d{yyyy-MM-dd HH:mm:ss} [%level] %logger{36} - %msg%n
```

## Error Handling

### Error Response Format

```json
{
  "error": {
    "code": "SERVICE_START_FAILED",
    "message": "Failed to start Karaf service",
    "details": {
      "service": "karaf",
      "port": 8080,
      "cause": "Port already in use"
    },
    "timestamp": "2024-01-30T10:15:30Z",
    "suggestions": [
      "Check if port 8080 is already in use",
      "Try using a different port with --port flag",
      "Stop other services using 'judo stop'"
    ]
  }
}
```

### Common Error Codes

| Code | Description | Resolution |
|------|-------------|------------|
| `CONFIG_NOT_FOUND` | Configuration file missing | Create judo.properties |
| `PORT_IN_USE` | Port already occupied | Change port or stop conflicting service |
| `DOCKER_NOT_RUNNING` | Docker daemon not available | Start Docker service |
| `BUILD_FAILED` | Maven build failed | Check build logs and fix compilation errors |
| `SERVICE_TIMEOUT` | Service failed to start within timeout | Increase timeout or check service logs |

## See Also

- [Commands Reference](../commands/) - Command-line interface
- [Configuration](../configuration/) - Configuration options
- [Examples](../examples/) - Usage examples and patterns