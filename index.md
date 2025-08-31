---
layout: home
title: "JUDO CLI"
subtitle: "Low Code Command Line Interface"
description: "Command-line tool for managing the lifecycle of JUDO applications"
hero_image: /assets/img/hero-placeholder.png
show_install_tabs: true
---

# JUDO CLI

A powerful command-line tool for managing the complete lifecycle of JUDO applications, from development to deployment.

## Key Features

- **Application Management**: Build, start, stop, and clean JUDO applications
- **Environment Profiles**: Support for multiple configuration profiles (karaf, compose-dev, etc.)
- **Container Orchestration**: Automated Docker container management for PostgreSQL and Keycloak
- **Maven Integration**: Seamless integration with Maven builds using mvnd
- **Database Operations**: Database dump and import functionality
- **Status Monitoring**: Real-time status checking of all services

## Quick Start

### Installation

#### Homebrew (macOS/Linux)
```bash
brew install blackbelttechnology/tap/judo
```

#### Manual Installation
Download the latest release from [GitHub Releases](https://github.com/BlackBeltTechnology/judo-cli/releases/latest) and extract to your PATH.

#### Building from Source
```bash
go build -o judo ./cmd/judo
```

### Basic Usage

```bash
# Build your JUDO application
judo build

# Start the application and services
judo start

# Check status of all services
judo status

# Stop all services
judo stop

# Clean environment
judo clean
```

## Documentation

- [Getting Started](docs/getting-started/) - Installation and first steps
- [Commands Reference](docs/commands/) - Complete command documentation
- [Configuration](docs/configuration/) - Environment and profile configuration
- [Examples](docs/examples/) - Common usage patterns and workflows

## Support

- [GitHub Issues](https://github.com/BlackBeltTechnology/judo-cli/issues) - Bug reports and feature requests
- [Discussions](https://github.com/BlackBeltTechnology/judo-cli/discussions) - Community support and questions

## Contributing

We welcome contributions! Please see our [Contributing Guide](https://github.com/BlackBeltTechnology/judo-cli/blob/develop/CONTRIBUTING.md) for details.
