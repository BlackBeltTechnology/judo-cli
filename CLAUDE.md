# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go-based CLI tool (`judo`) for managing JUDO application lifecycle. Orchestrates Docker containers, Maven builds, and Apache Karaf runtime environments.

## Core Architecture

- **Cobra-based CLI** with modular internal packages
- **Two runtime modes**: karaf (local dev + Docker services) and compose (full Docker Compose)
- **Profile-based configuration**: judo.properties, {env}.properties, judo-version.properties

## Essential Commands

### Building & Testing
```bash
go build -o judo ./cmd/judo           # Build CLI
go test ./internal/...               # Run tests (limited coverage)
go vet ./...                         # Vet code
go fmt ./...                         # Format code
go mod tidy                          # Clean modules
```

### Core CLI Operations
```bash
./judo build                         # Build application (Maven mvnd clean install)
./judo start                         # Start services
./judo stop                          # Stop services  
./judo status                        # Check service status
./judo clean                         # Clean environment
./judo -e compose-dev build start    # Use environment profile
```

### Development Workflows
```bash
./judo build -a                      # Build app module only
./judo build -f                      # Build frontend only  
./judo build -q                      # Quick mode (skip validations)
./judo reckless                      # Ultra-fast build and start
./judo session                       # Interactive session mode
```

## Release Management

### Version Control
```bash
CI=true scripts/version.sh get               # Get current version
CI=true scripts/version.sh increment patch   # Increment version
CI=true scripts/version.sh set 0.1.4 --commit # Set version with commit
```

### Local Release Builds
```bash
# Snapshot build (no tag required)
JUDO_VERSION=1.0.0-SNAPSHOT go build -o judo-snapshot ./cmd/judo

# Versioned build (requires git tag)
go build -ldflags "-X main.version=$(git describe --tags)" -o judo ./cmd/judo
```

### GitHub Actions
- **build.yml**: CI testing and snapshot releases from develop branch
- **release.yml**: Full releases from main branch with GoReleaser
- **docs.yml**: Documentation deployment to GitHub Pages

## Documentation

### Local Development
```bash
cd docs && ./serve-docs.sh           # Enhanced server with livereload
cd docs && ./serve-docs-simple.sh    # Simple server
```

**Note**: Ruby environment required for local Jekyll. Docs auto-deployed to https://judo.technology/

## Dependencies

- **External tools**: Docker, Maven/mvnd, SDKMAN, Git, Java
- **Go modules**: Cobra, testify, Docker client
- **Runtime**: PostgreSQL, Keycloak, Apache Karaf containers

## Module Structure

- **cmd/judo**: CLI entry point with Cobra setup
- **internal/commands**: Core command implementations
- **internal/config**: Configuration management (properties files)
- **internal/docker**: Docker container operations
- **internal/karaf**: Karaf runtime management
- **internal/db**: PostgreSQL database operations
- **internal/utils**: Common utilities
- **internal/session**: Interactive session mode
- **internal/selfupdate**: Self-update functionality

# important-instruction-reminders
Do what has been asked; nothing more, nothing less.
NEVER create files unless they're absolutely necessary for achieving your goal.
ALWAYS prefer editing an existing file to creating a new one.
