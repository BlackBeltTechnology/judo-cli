---
layout: page
title: doctor
parent: Command Reference
nav_order: 1
description: "Check system health and required dependencies"
---

# judo doctor

Check system health and required dependencies.

## Synopsis

```bash
judo doctor [flags]
```

## Description

The `doctor` command performs comprehensive system health checks to ensure all required tools and dependencies are available for JUDO development. It verifies installation status, checks port availability, and can automatically install missing tools.

## Flags

| Flag | Description |
|------|-------------|
| `-v, --verbose` | Show detailed version information for all tools |
| `-h, --help` | Show help for this command |

## What it Checks

### Required Tools
- **Docker**: Verifies Docker daemon is running
- **Maven**: Checks for Maven or Maven Daemon (mvnd)
- **Git**: Ensures Git is available for source control
- **Java**: Checks Java installation (optional for some operations)

### Development Tools
- **Maven Daemon (mvnd)**: Preferred over regular Maven
- **SDKMAN**: SDK manager for Java toolchain (auto-installs if missing)

### Port Availability
- **8080**: Keycloak (default)
- **8181**: Karaf (default)  
- **5432**: PostgreSQL (default)

### Project Status
- **JUDO Project**: Checks if current directory is initialized
- **Configuration**: Validates presence of required config files

## Auto-Installation Features

The doctor command can automatically install missing tools:

### SDKMAN Installation
If SDKMAN is not found, doctor will automatically install it:
- Downloads and installs SDKMAN
- Sets up the environment
- Installs required Java and Maven versions

### Development Tools Setup
When in a JUDO project directory with SDKMAN available:
- Automatically installs required Java version
- Installs Maven if not present
- Sets up the complete development environment

## Examples

### Basic Health Check
```bash
judo doctor
```
Output:
```
🩺 JUDO CLI Doctor - Checking system health...

✅ Docker: Available and running
✅ Maven: Available
✅ Git: Available
⚠️  Java: Not found (optional for some operations)
✅ Maven Daemon (mvnd): Available
✅ SDKMAN: Available

🔌 Port availability checks:
✅ Port 8080 (Keycloak): Available
✅ Port 8181 (Karaf): Available
✅ Port 5432 (PostgreSQL): Available

✅ JUDO Project: Initialized

🎉 All essential tools are available! JUDO CLI should work properly.
```

### Verbose Output
```bash
judo doctor -v
```
Shows detailed version information:
```
✅ Docker: Available and running
   Docker version: Docker version 24.0.7, build afdd53b

✅ Maven Daemon (mvnd): Available
   Maven Daemon version: Apache Maven Daemon 1.0.1

✅ Git: Available
   Git version: git version 2.42.0
```

### Missing Dependencies
```bash
judo doctor
```
When tools are missing:
```
🩺 JUDO CLI Doctor - Checking system health...

❌ Docker: Not available or not running
✅ Maven: Available
❌ Git: Not found
⚠️  SDKMAN: Not found - installing now...

   Installing SDKMAN...
   ✅ SDKMAN installed successfully

🚨 Some essential tools are missing. Please install them before using JUDO CLI.
```

## Port Conflict Detection

When ports are in use, doctor provides intelligent feedback:

### JUDO Services Using Ports
```bash
⚠️  Port 8181 (Karaf): In use by current Karaf instance
   Note: This port is used by your running JUDO application
```

### External Services Using Ports
```bash
❌ Port 8080 (Keycloak): In use by another process
   Warning: This port is occupied by another application, which will cause conflicts
```

## Project Detection

### Initialized Project
```bash
✅ JUDO Project: Initialized

🔧 Installing required development tools via SDKMAN...
✅ Development tools installed successfully
```

### Uninitialized Directory
```bash
ℹ️  JUDO Project: Not initialized in this directory
   Run 'judo init' to initialize a new JUDO project
```

## Common Issues and Solutions

### Docker Not Running
**Issue**: `❌ Docker: Not available or not running`

**Solutions**:
- Start Docker Desktop
- Check Docker daemon status: `systemctl status docker` (Linux)
- Verify Docker installation

### Maven Not Found
**Issue**: `❌ Maven: Not found`

**Solutions**:
- Install via SDKMAN: `sdk install maven`
- Install via package manager: `brew install maven`
- Download from Apache Maven website

### Port Conflicts
**Issue**: `❌ Port 8080: In use by another process`

**Solutions**:
- Stop the conflicting service
- Use alternate ports via `--options` flag
- Configure different ports in properties files

### SDKMAN Installation Fails
**Issue**: `❌ Failed to install SDKMAN`

**Solutions**:
- Check internet connectivity
- Verify curl/wget availability
- On Windows, ensure WSL is properly configured
- Manual installation: https://sdkman.io/install

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All essential tools available |
| 1 | Some essential tools missing |

## Integration with Other Commands

The doctor command is automatically run by some other commands:
- `judo init`: Runs basic health checks before project creation
- Interactive session: Shows condensed health status

## Best Practices

1. **Run Before Starting**: Always run `doctor` in new environments
2. **Use Verbose Mode**: Use `-v` flag when troubleshooting
3. **Auto-Install**: Let doctor auto-install SDKMAN and tools when possible
4. **Regular Checks**: Run periodically to catch environment changes
5. **Project Setup**: Run in project directory for complete environment setup

## See Also

- [`judo init`](init) - Initialize new JUDO project
- [`judo session`](session) - Interactive session with health indicators
- [Installation Guide](/installation/) - Detailed installation instructions