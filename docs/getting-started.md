---
layout: page
title: "Getting Started"
permalink: /getting-started/
excerpt: "Installation and first steps with JUDO CLI"
---

# Getting Started with JUDO CLI

This guide will help you install and start using JUDO CLI to manage your JUDO applications.

## Prerequisites

Before using JUDO CLI, ensure you have the following installed:

- **Docker**: Required for PostgreSQL and Keycloak containers
- **Java 11+**: Required for JUDO applications
- **Maven** or **mvnd**: For building Java applications
- **Git**: For version control operations

### Optional Tools
- **SDKMAN**: For managing Java versions
- **Node.js**: If your application includes frontend components

## Installation

### Homebrew (Recommended for macOS/Linux)

```bash
# Add the BlackBelt Technology tap
brew tap blackbelttechnology/tap

# Install JUDO CLI
brew install judo

# Verify installation
judo version
```

### Manual Installation

#### Linux (x86_64)
```bash
curl -L https://github.com/BlackBeltTechnology/judo-cli/releases/latest/download/judo_Linux_x86_64.tar.gz | tar xz
sudo mv judo /usr/local/bin/
```

#### macOS (Intel)
```bash
curl -L https://github.com/BlackBeltTechnology/judo-cli/releases/latest/download/judo_Darwin_x86_64.tar.gz | tar xz
sudo mv judo /usr/local/bin/
```

#### macOS (Apple Silicon)
```bash
curl -L https://github.com/BlackBeltTechnology/judo-cli/releases/latest/download/judo_Darwin_arm64.tar.gz | tar xz
sudo mv judo /usr/local/bin/
```

#### Windows
```powershell
Invoke-WebRequest -Uri "https://github.com/BlackBeltTechnology/judo-cli/releases/latest/download/judo_Windows_x86_64.zip" -OutFile "judo.zip"
Expand-Archive judo.zip -DestinationPath .
# Move judo.exe to a directory in your PATH
```

## First Steps

### 1. Verify Installation

```bash
judo version
```

This should display the JUDO CLI version and build information.

### 2. Check Available Commands

```bash
judo help
```

### 3. Navigate to Your JUDO Project

```bash
cd /path/to/your/judo-project
```

### 4. Build Your Application

```bash
# Full build (recommended for first time)
judo build

# Quick build (app module only)
judo build -a

# Frontend only
judo build -f
```

### 5. Start Your Application

```bash
# Start all services
judo start

# Check status
judo status
```

### 6. Access Your Application

Once started, your application will typically be available at:
- **Application**: http://localhost:8080
- **Keycloak Admin**: http://localhost:8180 (if using Keycloak)

## Try With the Included Test Model

A fully featured sample project is included in this repository under `test-model/`. It is suitable for end‑to‑end testing of JUDO CLI with real infrastructure (Karaf, PostgreSQL, Keycloak).

- Location: `test-model/`
- Works with: generate, build, start, stop, dump, import, export
- How to use:

```bash
cd test-model
judo doctor -v
judo build
judo start --options "runtime=karaf,dbtype=postgresql"
# Database workflows
judo dump
judo import
# Stop and clean
judo stop
judo clean
```

## Configuration

JUDO CLI uses configuration files to manage different environments:

- `judo.properties` - Default configuration
- `{env}.properties` - Environment-specific overrides
- `judo-version.properties` - Version constraints

### Example Configuration

Create a `judo.properties` file in your project root:

```properties
# Application settings
app.name=my-judo-app
app.schema=myapp

# Database settings
database.type=postgresql
database.host=localhost
database.port=5432
database.name=myapp_db

# Runtime mode
runtime.mode=karaf

# Ports
karaf.port=8080
keycloak.port=8180
postgres.port=5432
```

## Next Steps

- Learn about [Commands](../commands/) available in JUDO CLI
- Explore [Configuration](../configuration/) options
- Check out [Examples](../examples/) for common workflows

## Troubleshooting

### Common Issues

**Docker not running**
```bash
# Check Docker status
docker info

# Start Docker service (Linux)
sudo systemctl start docker
```

**Port conflicts**
```bash
# Check what's using a port
lsof -i :8080

# Stop conflicting services
judo stop
```

**Build failures**
```bash
# Clean and rebuild
judo clean
judo build
```

## Getting Help

- Run `judo help` for command-specific help
- Check the [Commands Reference](../commands/) for detailed documentation
- Visit [GitHub Issues](https://github.com/BlackBeltTechnology/judo-cli/issues) for bug reports