---
title: "Commands Reference"
permalink: /docs/commands/
excerpt: "Complete reference for all JUDO CLI commands"
---

# Commands Reference

Complete documentation for all JUDO CLI commands and their options.

## Core Commands

### `judo build`

Build the JUDO application using Maven.

```bash
judo build [options]
```

**Options:**
- `-a, --app` - Build app module only
- `-f, --frontend` - Build frontend only  
- `-q, --quick` - Quick mode (skip validations, use cache)

**Examples:**
```bash
# Full build
judo build

# App module only
judo build -a

# Frontend only
judo build -f

# Quick build
judo build -q
```

### `judo start`

Start the application and required services.

```bash
judo start [options]
```

**Options:**
- `-d, --detach` - Run in background
- `-w, --wait` - Wait for services to be ready

**Examples:**
```bash
# Start all services
judo start

# Start in background
judo start -d

# Start and wait for readiness
judo start -w
```

### `judo stop`

Stop all running services.

```bash
judo stop [options]
```

**Options:**
- `-f, --force` - Force stop containers
- `-t, --timeout` - Timeout in seconds (default: 10)

**Examples:**
```bash
# Graceful stop
judo stop

# Force stop
judo stop -f

# Stop with custom timeout
judo stop -t 30
```

### `judo status`

Check the status of all services.

```bash
judo status [options]
```

**Options:**
- `-v, --verbose` - Show detailed status
- `-j, --json` - Output in JSON format

**Examples:**
```bash
# Basic status
judo status

# Detailed status
judo status -v

# JSON output
judo status -j
```

### `judo clean`

Clean the environment by removing containers and volumes.

```bash
judo clean [options]
```

**Options:**
- `-a, --all` - Remove all containers and volumes
- `-v, --volumes` - Remove volumes only
- `-c, --containers` - Remove containers only

**Examples:**
```bash
# Clean everything
judo clean

# Remove all containers and volumes
judo clean -a

# Remove volumes only
judo clean -v
```

## Utility Commands

### `judo reckless`

Ultra-fast build and start for rapid development.

```bash
judo reckless
```

This command combines `judo build -q` and `judo start` for the fastest possible development cycle.

### `judo prune`

Remove untracked files and reset state.

```bash
judo prune [options]
```

**Options:**
- `-f, --force` - Skip confirmation prompt
- `-d, --dry-run` - Show what would be removed

**Examples:**
```bash
# Interactive prune
judo prune

# Force prune without confirmation
judo prune -f

# Dry run (show what would be removed)
judo prune -d
```

### `judo update`

Update JUDO dependency versions.

```bash
judo update [options]
```

**Options:**
- `-c, --check` - Check for available updates
- `-f, --force` - Force update even if no changes

**Examples:**
```bash
# Update dependencies
judo update

# Check for updates only
judo update -c
```

## Database Commands

### `judo db dump`

Export database to SQL file.

```bash
judo db dump [filename] [options]
```

**Arguments:**
- `filename` - Output filename (optional, defaults to timestamp)

**Options:**
- `-t, --table` - Dump specific table only
- `-c, --clean` - Include DROP statements

**Examples:**
```bash
# Dump entire database
judo db dump

# Dump to specific file
judo db dump backup.sql

# Dump specific table
judo db dump -t users

# Include DROP statements
judo db dump -c
```

### `judo db import`

Import SQL file to database.

```bash
judo db import <filename> [options]
```

**Arguments:**
- `filename` - SQL file to import (required)

**Options:**
- `-f, --force` - Force import without confirmation
- `-c, --clean` - Clean database before import

**Examples:**
```bash
# Import SQL file
judo db import backup.sql

# Force import without confirmation
judo db import backup.sql -f

# Clean database before import
judo db import backup.sql -c
```

## Environment Commands

### `judo -e <env>`

Run commands with specific environment profile.

```bash
judo -e <environment> <command>
```

**Examples:**
```bash
# Use compose-dev environment
judo -e compose-dev build start

# Use production environment
judo -e production status
```

**Common Environments:**
- `karaf` - Local Karaf runtime (default)
- `compose-dev` - Docker Compose development
- `compose-prod` - Docker Compose production

## Global Options

These options can be used with any command:

- `-e, --environment` - Specify environment profile
- `-v, --verbose` - Enable verbose output
- `-q, --quiet` - Suppress non-essential output
- `-h, --help` - Show help for command
- `--version` - Show version information

## Configuration Commands

### `judo config`

Manage configuration settings.

```bash
judo config <subcommand>
```

**Subcommands:**
- `list` - List all configuration values
- `get <key>` - Get specific configuration value
- `set <key> <value>` - Set configuration value

**Examples:**
```bash
# List all configuration
judo config list

# Get specific value
judo config get database.type

# Set configuration value
judo config set app.port 8090
```

## Logging and Debugging

### `judo logs`

View application and service logs.

```bash
judo logs [service] [options]
```

**Arguments:**
- `service` - Specific service (karaf, postgres, keycloak)

**Options:**
- `-f, --follow` - Follow log output
- `-t, --tail` - Number of lines to show from end
- `--since` - Show logs since timestamp

**Examples:**
```bash
# View all logs
judo logs

# Follow Karaf logs
judo logs karaf -f

# Show last 100 lines
judo logs -t 100

# Show logs since specific time
judo logs --since "2024-01-01T10:00:00"
```

## Exit Codes

JUDO CLI uses standard exit codes:

- `0` - Success
- `1` - General error
- `2` - Misuse of shell command
- `130` - Script terminated by Ctrl+C

## Command Chaining

You can chain multiple commands:

```bash
# Build and start
judo build start

# Stop, clean, build, and start
judo stop clean build start
```

## See Also

- [Configuration](../configuration/) - Environment and profile configuration
- [Examples](../examples/) - Common usage patterns
- [Getting Started](../getting-started/) - Installation and setup