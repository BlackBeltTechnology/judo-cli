---
title: "API Reference"
permalink: /api/
excerpt: "Technical reference for JUDO CLI internals and extension points"
---

# API Reference

Technical reference for JUDO CLI internals, configuration schemas, and extension points.

## Configuration Schema

### Application Configuration

```properties
# judo.properties schema
app_name: string          # Application name (required)
schema_name: string       # Database schema name
app_version: string       # Application version
app_description: string   # Application description

dbtype: enum              # postgresql | hsqldb
database_host: string     # Database hostname
database_port: integer    # Database port (1-65535)
database_name: string     # Database name
database_username: string # Database username
database_password: string # Database password

runtime: enum             # karaf | compose

karaf_port: integer       # HTTP port (1-65535)
karaf_debug_port: integer # Debug port
karaf_debug_enabled: boolean # Enable debug mode
karaf_ssh_port: integer   # SSH port
karaf_ssh_enabled: boolean # Enable SSH access

keycloak_port: integer    # Keycloak port
keycloak_realm: string    # Keycloak realm name
keycloak_enabled: boolean # Enable Keycloak
postgres_port: integer    # PostgreSQL port
postgres_version: string  # PostgreSQL version

build_skip_tests: boolean # Skip tests during build
build_parallel: boolean   # Enable parallel build
maven_profile: string     # Maven profile to use
```

## Environment Variables

All configuration properties can be overridden using environment variables with the pattern:
`JUDO_` + property path in uppercase with dots replaced by underscores.

### Examples

| Property | Environment Variable | Example |
|----------|---------------------|---------|
| `app_name` | `JUDO_APP_NAME` | `JUDO_APP_NAME=myapp` |
| `database_host` | `JUDO_DATABASE_HOST` | `JUDO_DATABASE_HOST=localhost` |
| `karaf_port` | `JUDO_KARAF_PORT` | `JUDO_KARAF_PORT=8080` |
| `build_skip_tests` | `JUDO_BUILD_SKIP_TESTS` | `JUDO_BUILD_SKIP_TESTS=true` |

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

**Note**: All commands output simple text messages rather than structured JSON formats.


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


## Docker Integration

### Container Management

JUDO CLI manages Docker containers for PostgreSQL and Keycloak services. The actual container configuration is handled internally and outputs simple text status messages rather than structured JSON.






### Common Error Conditions

| Condition | Description | Resolution |
|-----------|-------------|------------|
| Configuration file missing | judo.properties not found | Create judo.properties file |
| Port already occupied | Service port conflict | Change port or stop conflicting service |
| Docker not running | Docker daemon unavailable | Start Docker service |
| Build failed | Maven compilation errors | Check build logs and fix errors |
| Service timeout | Service failed to start | Increase timeout or check service logs |

## See Also

- [Commands Reference](../commands/) - Command-line interface
- [Configuration](../configuration/) - Configuration options
- [Examples](../examples/) - Usage examples and patterns