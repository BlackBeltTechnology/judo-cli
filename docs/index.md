---
layout: default
title: JUDO CLI Documentation
nav_order: 1
description: "Complete documentation for the JUDO CLI tool"
permalink: /
---

# JUDO CLI Documentation

Welcome to the JUDO CLI documentation! This tool helps you manage the complete lifecycle of JUDO applications.

## Quick Start

### Installation

#### Homebrew (Recommended for macOS/Linux)
```bash
brew tap blackbelttechnology/tap
brew install judo
```

#### Manual Installation
Download the latest release from the [GitHub releases page](https://github.com/BlackBeltTechnology/judo-cli/releases).

#### Building from Source
```bash
go build -o judo ./cmd/judo
```

### First Steps

1. **Check system requirements**:
   ```bash
   judo doctor
   ```

2. **Initialize a new project**:
   ```bash
   judo init
   ```

3. **Build and start your application**:
   ```bash
   judo build start
   ```

## Features

- **Interactive Session Mode**: Command history, auto-completion, and persistent state
- **Multi-Runtime Support**: Karaf and Docker Compose environments
- **Database Management**: PostgreSQL dump/import operations
- **Auto-update**: Self-update functionality for snapshot versions
- **Cross-platform**: macOS, Linux, and Windows support

## Command Reference

### System Commands
- [`judo doctor`](commands/doctor) - System health check
- [`judo init`](commands/init) - Initialize new JUDO project
- [`judo session`](commands/session) - Interactive session mode

### Build Commands
- [`judo build`](commands/build) - Build project with various options
- [`judo reckless`](commands/reckless) - Fast build & run mode

### Application Lifecycle
- [`judo start`](commands/start) - Start application
- [`judo stop`](commands/stop) - Stop application
- [`judo status`](commands/status) - Check service status
- [`judo log`](commands/log) - View application logs

### Database Operations
- [`judo dump`](commands/dump) - Database backup
- [`judo import`](commands/import) - Database restore
- [`judo schema-upgrade`](commands/schema-upgrade) - Schema migration

### Maintenance
- [`judo clean`](commands/clean) - Clean environment
- [`judo prune`](commands/prune) - Remove untracked files
- [`judo update`](commands/update) - Update dependencies
- [`judo self-update`](commands/self-update) - Update CLI tool

## Configuration

The JUDO CLI uses profile-based configuration:

- **Default**: `judo.properties`
- **Environment-specific**: `{env}.properties` (e.g., `compose-dev.properties`)
- **Version constraints**: `judo-version.properties`

## Runtime Modes

### Karaf Runtime
Local development with Apache Karaf + Docker services:
- Apache Karaf application server
- PostgreSQL database container
- Keycloak authentication container

### Compose Runtime
Full Docker Compose environment with all services containerized.

## Interactive Session

The interactive session provides:
- Command history persistence
- Tab completion for commands and flags
- Real-time service status indicators
- Context-aware suggestions
- Session management and statistics

## License

This project is licensed under the [Eclipse Public License 2.0 (EPL-2.0)](https://www.eclipse.org/legal/epl-2.0/).

Copyright © 2026 BlackBelt Meta Zrt.

## Support

- [GitHub Issues](https://github.com/BlackBeltTechnology/judo-cli/issues)
- [Documentation](https://blackbeltechnology.github.io/judo-cli/)
- [Releases](https://github.com/BlackBeltTechnology/judo-cli/releases)