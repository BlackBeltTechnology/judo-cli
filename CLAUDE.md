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
./build.sh                            # Comprehensive build (frontend + backend with embedded assets)
go test ./internal/...               # Run tests (limited coverage)
go test -v ./internal/...            # Run tests with verbose output
go test ./internal/db/...            # Run specific package tests
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
./judo doctor                       # Check system requirements and setup
./judo prune                        # Clean up Docker resources
./judo self-update                  # Update CLI tool itself
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
- **docs-release.yml**: Documentation deployment to GitHub Pages on `current-site` tags
- **docs.yml**: Documentation build and testing
- **manual-release.yml**: Manual release triggering workflow

## Documentation

### Local Development
```bash
cd docs && ./serve-docs.sh           # Enhanced server with livereload
cd docs && ./serve-docs-simple.sh    # Simple server
```

**Note**: Hugo framework used for documentation. Docs auto-deployed to https://judo.technology/

## Dependencies

- **External tools**: Docker, Maven/mvnd, SDKMAN, Git, Java
- **Go modules**: Cobra, testify, Docker client
- **Runtime**: PostgreSQL, Keycloak, Apache Karaf containers

## Module Structure

- **cmd/judo**: CLI entry point with Cobra setup and version info
- **internal/commands**: Core command implementations (build, start, stop, status, clean, etc.)
- **internal/config**: Configuration management (properties files, environment setup)
- **internal/docker**: Docker container operations and client management
- **internal/karaf**: Karaf runtime management and operations
- **internal/db**: PostgreSQL database operations (dump, import, schema upgrade)
- **internal/utils**: Common utilities and command execution helpers
- **internal/session**: Interactive session mode with command history
- **internal/selfupdate**: Self-update functionality
- **internal/help**: Command help text and documentation
- **internal/server**: Web server with browser-based interface

## Advanced Features

### Database Operations
- **Schema Management**: Automatic schema upgrades and migrations
- **Data Import/Export**: PostgreSQL dump/restore functionality with timestamped backups
- **Multi-database Support**: HSQLDB (default) and PostgreSQL with automatic configuration

### Build System Features
- **Incremental Builds**: Smart build targeting with module-specific flags
- **Parallel Builds**: Maven parallel execution support (-T 1C)
- **Quick Mode**: Skip validations and use cache for faster development cycles
- **Reckless Mode**: Ultra-fast builds prioritizing speed over completeness

### Container Management
- **Automatic Network Setup**: Docker network creation and management
- **Port Conflict Detection**: Smart port checking with service-specific validation
- **Health Checking**: Container status monitoring and readiness verification
- **Volume Management**: Persistent data volume handling for databases

### Development Tools
- **Interactive Session**: Command history, auto-completion, and persistent state
- **Web Interface**: Browser-based GUI accessible via `judo server`
- **System Diagnostics**: Comprehensive `doctor` command for dependency checking
- **Self-update Mechanism**: Automatic updates for snapshot versions

## Configuration System

### Property Files Hierarchy
1. **Profile-specific**: `{profile}.properties` (highest priority)
2. **Global**: `judo.properties`
3. **Version**: `judo-version.properties`
4. **Defaults**: Built-in sensible defaults

### Key Configuration Properties
- `runtime`: "karaf" or "compose"
- `dbtype`: "hsqldb" or "postgresql"
- `karaf_port`: Karaf console port (default: 8181)
- `postgres_port`: PostgreSQL port (default: 5432)
- `keycloak_port`: Keycloak port (default: 8080)
- `model_dir`: Custom model directory path

## Performance Optimization

### Build Optimizations
- **Maven Daemon**: mvnd support for faster build times
- **Selective Building**: Target specific modules (-a, -f flags)
- **Cache Utilization**: Smart caching strategies for repeated builds
- **Validation Skipping**: Optional validation bypass for development speed

### Runtime Optimizations
- **Container Reuse**: Existing container detection and reuse
- **Port Management**: Intelligent port allocation and conflict resolution
- **Network Optimization**: Efficient Docker network configuration
- **Resource Management**: Proper container lifecycle management

## Security Features

- **Keycloak Integration**: OAuth2/OpenID Connect authentication
- **SSL/TLS Support**: Certificate management for secure communications
- **Port Security**: Port availability checking and conflict prevention
- **Container Isolation**: Proper Docker network segmentation

## Testing and Quality

### Test Coverage
- **Unit Tests**: Core functionality testing across all packages
- **Integration Tests**: Docker and runtime integration testing
- **Command Testing**: Comprehensive CLI command validation
- **Error Handling**: Robust error condition testing

### Code Quality
- **Go Vet**: Static analysis and code validation
- **Go Fmt**: Consistent code formatting
- **Linting**: Code style and best practices enforcement
- **Dependency Management**: Clean module management with go mod tidy

## Deployment Strategies

### Local Development
- **Karaf Runtime**: Local Karaf instance with Docker services
- **Compose Runtime**: Full Docker Compose environment
- **Hybrid Mode**: Mixed local and containerized services

### Production Deployment
- **Docker Compose**: Full containerized deployment
- **Standalone Binaries**: Self-contained executable distribution
- **Versioned Releases**: Semantic versioning with GitHub releases
- **Homebrew Distribution**: macOS package management integration

## Monitoring and Logging

- **Container Logs**: Docker container log access and tailing
- **Karaf Console**: Real-time Karaf console output monitoring
- **Port Monitoring**: Service port availability and health checking
- **Build Logging**: Comprehensive build process logging and output

## Extensibility Points

### Custom Commands
- **Cobra Framework**: Easy addition of new CLI commands
- **Configuration Hooks**: Pre/post command execution hooks
- **Plugin System**: Potential for external command plugins

### Integration Points
- **Docker API**: Full Docker client integration
- **Maven Integration**: Deep Maven build system integration
- **Database Connectivity**: Multiple database backend support
- **Web Interface**: Extensible web server architecture

# important-instruction-reminders
Do what has been asked; nothing more, nothing less.
NEVER create files unless they're absolutely necessary for achieving your goal.
ALWAYS prefer editing an existing file to creating a new one.
