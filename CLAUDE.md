# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based CLI tool called `judo` for managing the lifecycle of JUDO applications. The tool orchestrates Docker containers, Maven builds, and Apache Karaf runtime environments for Java-based enterprise applications.

## Core Architecture

The CLI is built using the Cobra library and follows this structure:

- **cmd/judo/main.go**: Entry point that sets up the root command and registers all subcommands
- **internal/commands/commands.go**: Core command implementations (build, start, stop, clean, etc.)
- **internal/config/config.go**: Configuration management for judo.properties files and environment profiles
- **internal/docker/docker.go**: Docker container management (PostgreSQL, Keycloak)
- **internal/karaf/karaf.go**: Apache Karaf runtime management
- **internal/db/db.go**: Database operations (dump/import for PostgreSQL)
- **internal/utils/utils.go**: Common utilities for command execution and system checks

The tool operates in two main runtime modes:
- **karaf**: Local development with Karaf runtime + Docker services
- **compose**: Full Docker Compose environment

## Essential Commands

### Building the CLI
```bash
go build -o judo ./cmd/judo
```

### Running Tests
```bash
go test ./...
```

### Core Application Management
- `./judo build` - Build the JUDO application using Maven (mvnd clean install)
- `./judo start` - Start the application and required services
- `./judo stop` - Stop all running services
- `./judo clean` - Clean environment (remove containers/volumes)
- `./judo status` - Check status of all services

### Build Variations
- `./judo build -a` - Build app module only
- `./judo build -f` - Build frontend only
- `./judo build -q` - Quick mode (skip validations, use cache)
- `./judo reckless` - Ultra-fast build and start

### Environment Management
- `./judo -e compose-dev build start` - Use alternate environment profile
- `./judo prune` - Remove untracked files and reset state
- `./judo update` - Update JUDO dependency versions

## Configuration System

The tool uses a profile-based configuration system:
- **judo.properties**: Default configuration
- **{env}.properties**: Environment-specific overrides (e.g., compose-dev.properties)
- **judo-version.properties**: Version constraints

Key configuration aspects:
- Database type (postgresql/hsqldb)
- Runtime mode (karaf/compose)
- Port assignments for services
- Schema and application names

## Development Dependencies

The application requires several external tools:
- **Docker**: For PostgreSQL, Keycloak containers
- **Maven/mvnd**: For building Java applications
- **SDKMAN**: For managing Java toolchain versions
- **Git**: For source control and clean operations

## Test Framework

Uses testify for Go unit tests. Current test coverage is limited, with most tests being integration-style or requiring external dependencies (Docker, file system).

Run tests with: `go test ./internal/...`