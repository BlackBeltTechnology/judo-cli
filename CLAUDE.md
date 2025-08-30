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

## Release Management

The project uses GoReleaser for automated releases with GitHub Actions:

### Version Management
```bash
# Get current version
CI=true scripts/version.sh get

# Increment version (patch/minor/major)
CI=true scripts/version.sh increment patch

# Set specific version with commit
CI=true scripts/version.sh set 0.1.4 --commit
```

### Release Workflow
- **Manual releases**: `.github/workflows/manual-release.yml` for on-demand releases
- **Automated releases**: `.github/workflows/release.yml` triggered by version tags
- **Build verification**: `.github/workflows/build.yml` for continuous integration
- **GoReleaser config**: `.goreleaser.yml` defines cross-platform builds and Homebrew tap

### Release Artifacts
- Multi-platform binaries (Linux, macOS, Windows)
- Homebrew formula auto-update
- GitHub release with changelog
- Checksum verification

### Building Releases Locally
```bash
# Build snapshot (no tag required)
JUDO_VERSION=1.0.0-SNAPSHOT go build -o judo-snapshot ./cmd/judo

# Build with version info (requires git tag)
go build -ldflags "-X main.version=$(git describe --tags)" -o judo ./cmd/judo
```

## Documentation Development

The project includes Jekyll-based documentation with automated GitHub Pages deployment:

### Local Documentation Server
```bash
# Enhanced server with livereload and error handling
./serve-docs.sh

# Simple server for compatibility
./serve-docs-simple.sh

# Manual Jekyll commands
bundle install
bundle exec jekyll serve --livereload
```

### Documentation Structure
- **Main docs**: `docs/` directory with Jekyll configuration
- **Content**: `_docs/` directory for documentation pages
- **Auto-deployment**: `.github/workflows/docs.yml` publishes to GitHub Pages
- **Local setup**: Ruby/Jekyll environment with Gemfile dependencies

## Interactive Session Mode

The CLI includes an advanced interactive session with enhanced features:

### Session Features
- **Real-time status**: Service indicators in prompt (⚙️karaf:✓ 🔐keycloak:✗ 🐘postgres:✓)
- **Command history**: Persistent across sessions
- **Auto-completion**: Tab completion for commands and flags
- **Context awareness**: Tracks project state and provides relevant suggestions

### Session Commands
```bash
# Start interactive mode
judo session

# Within session
help              # Show all available commands
status            # Show detailed session and service status
history           # Display command history
clear             # Clear terminal
exit/quit         # Exit session
```

## Advanced Development Workflows

### Code Generation and Linting
```bash
# Go code generation (if applicable)
go generate ./...

# Code formatting
go fmt ./...
gofmt -s -w .

# Go modules maintenance
go mod tidy

# Vet code for issues
go vet ./...
```

### Docker Integration Testing
```bash
# The CLI manages Docker containers for:
# - PostgreSQL database (judo-postgres)
# - Keycloak authentication (judo-keycloak)
# - Schema migration utilities

# Test Docker connectivity
docker ps
docker system info
```

## Module Architecture Details

The codebase uses internal Go modules with clear separation:

- **cmd/judo**: Main CLI entry point with Cobra command registration
- **internal/commands**: Command implementations with business logic
- **internal/config**: Profile-based configuration management
- **internal/docker**: Docker client wrapper and container management
- **internal/karaf**: Apache Karaf runtime control
- **internal/db**: PostgreSQL database operations
- **internal/utils**: Shared utilities for command execution
- **internal/session**: Interactive session mode implementation
- **internal/selfupdate**: Self-update functionality for snapshot versions
- **internal/help**: Command help text and documentation

# important-instruction-reminders
Do what has been asked; nothing more, nothing less.
NEVER create files unless they're absolutely necessary for achieving your goal.
ALWAYS prefer editing an existing file to creating a new one.
NEVER proactively create documentation files (*.md) or README files. Only create documentation files if explicitly requested by the User.