---
layout: page
title: Installation
nav_order: 3
description: "How to install JUDO CLI on your system"
permalink: /installation/
---

# Installation

Multiple installation methods are available for JUDO CLI.

## Homebrew (Recommended for macOS/Linux)

The easiest way to install JUDO CLI on macOS and Linux:

```bash
# Add the BlackBelt Technology tap
brew tap blackbelttechnology/tap

# Install JUDO CLI
brew install judo

# Or install directly
brew install blackbelttechnology/tap/judo
```

### Updating via Homebrew

```bash
brew upgrade blackbelttechnology/tap/judo
```

## Manual Installation

Download the latest release from the [GitHub releases page](https://github.com/BlackBeltTechnology/judo-cli/releases):

### Linux x86_64
```bash
curl -L https://github.com/BlackBeltTechnology/judo-cli/releases/latest/download/judo_Linux_x86_64.tar.gz | tar xz
sudo mv judo /usr/local/bin/
```

### macOS x86_64 (Intel)
```bash
curl -L https://github.com/BlackBeltTechnology/judo-cli/releases/latest/download/judo_Darwin_x86_64.tar.gz | tar xz
sudo mv judo /usr/local/bin/
```

### macOS arm64 (Apple Silicon)
```bash
curl -L https://github.com/BlackBeltTechnology/judo-cli/releases/latest/download/judo_Darwin_arm64.tar.gz | tar xz
sudo mv judo /usr/local/bin/
```

### Windows x86_64
Download and extract `judo_Windows_x86_64.zip` from the releases page, then add the executable to your PATH.

## Building from Source

To build from source, you need Go 1.21 or later:

```bash
git clone https://github.com/BlackBeltTechnology/judo-cli.git
cd judo-cli

# Simple build
go build -o judo ./cmd/judo

# Or use the comprehensive build script (recommended)
./build.sh
```

## Comprehensive Build Script

The project includes a `build.sh` script that provides a complete build process:

```bash
# Run the comprehensive build
./build.sh
```

This script:
1. Builds the React frontend application
2. Prepares embedded assets for Go embedding
3. Compiles the Go backend with embedded frontend
4. Runs tests to ensure everything works
5. Includes proper version metadata and git commit information

## Self-Update (Snapshot Versions)

For snapshot versions, you can update using the built-in command:

```bash
# Check for updates
judo self-update --check

# Update to latest snapshot
judo self-update
```

> **Note:** Self-update only works for snapshot versions. Stable versions should be updated via Homebrew or manual download for safety.

## Verification

After installation, verify it works:

```bash
judo --version
judo doctor
```

## System Requirements

JUDO CLI requires these tools to be installed:

- **Docker** - For container management
- **Maven/mvnd** - For building Java applications
- **Git** - For source control operations
- **Java** - For application runtime (auto-installed via SDKMAN)

Run `judo doctor` to check all dependencies and get installation guidance for missing tools.

## Next Steps

After installation:

1. **Check system health**: `judo doctor -v`
2. **Initialize a project**: `judo init`
3. **Start interactive session**: `judo session`

See the [Command Reference](/commands/) for complete usage documentation.