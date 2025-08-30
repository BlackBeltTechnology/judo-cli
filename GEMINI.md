# GEMINI.md

## Project Overview

This project is a command-line interface (CLI) tool named `judo-cli`, written in Go. It is designed to manage the lifecycle of JUDO applications. The tool utilizes the Cobra library for its command structure and interacts with several external tools and services:

*   **Docker:** Manages Docker containers for services like PostgreSQL and Keycloak
*   **Maven:** Used for building the application with Maven Daemon (mvnd)
*   **Karaf:** Apache Karaf runtime management for local development
*   **GitHub API:** For self-update functionality to fetch latest releases

The CLI uses a profile-based configuration system with `judo.properties` files and supports different environments via the `--env` or `-e` flag.

## Project Structure

```
judo-cli/
├── cmd/judo/main.go          # CLI entry point with version info and command registration
├── internal/
│   ├── commands/commands.go  # Core command implementations
│   ├── config/config.go      # Configuration management
│   ├── docker/docker.go      # Docker container management
│   ├── karaf/karaf.go        # Apache Karaf runtime management
│   ├── db/db.go             # Database operations (PostgreSQL)
│   ├── utils/utils.go       # Common utilities and command execution
│   ├── selfupdate/          # Self-update functionality
│   ├── session/             # Interactive session management
│   └── help/                # Help text generation
├── scripts/version.sh       # Version management script
├── VERSION                  # Current version file
└── .goreleaser.yml         # GoReleaser configuration for releases
```

## Building and Running

### Building the CLI

To build the `judo-cli` executable, run the following command:

```sh
# Standard build
go build -o judo ./cmd/judo

# Build with specific version (for testing)
go build -ldflags="-X main.version=1.0.0-SNAPSHOT" -o judo-snapshot ./cmd/judo

# Build with version from VERSION file
VERSION=$(cat VERSION)
go build -ldflags="-X main.version=$VERSION" -o judo ./cmd/judo
```

### Managing the Application

Once the CLI is built, you can use it to manage your JUDO application. Here are some of the key commands:

*   **Build the application:**
    ```sh
    ./judo build
    ```
    This command executes a Maven build (`mvnd clean install`). It supports various flags to customize the build process, such as `--build-parallel` (`-p`) for parallel builds and `--version` (`-v`) to specify a version number.

*   **Start the application:**
    ```sh
    ./judo start
    ```
    This command starts the application. It can start a Docker Compose environment or a local environment with Karaf, PostgreSQL, and Keycloak, depending on the configuration.

*   **Stop the application:**
    ```sh
    ./judo stop
    ```
    This command stops the running application and associated services.

*   **Clean the environment:**
    ```sh
    ./judo clean
    ```
    This command stops and removes Docker containers, networks, and volumes associated with the application.

*   **Update dependencies:**
    ```sh
    ./judo update
    ```
    This command updates the dependency versions in the JUDO project.

*   **Self-update the CLI:**
    ```sh
    ./judo self-update
    ```
    Updates the judo CLI to the latest version from GitHub releases. Only works for snapshot versions.
    
    ```sh
    ./judo self-update --check    # Check for updates without installing
    ./judo self-update --force    # Force update even if already up to date
    ```

## Version Management

The CLI uses semantic versioning with build-time version injection:

*   **Version Source**: The `VERSION` file contains the current semantic version
*   **Build Flags**: Version is injected using `-ldflags="-X main.version=VERSION"`
*   **Snapshot Detection**: Self-update only works for versions containing "SNAPSHOT"
*   **Release Process**: Uses GoReleaser for multi-platform binary releases

## Self-Update Functionality

The self-update feature provides automated CLI updates:

*   **GitHub Integration**: Fetches releases from GitHub API
*   **Platform Detection**: Automatically downloads correct binary for OS/architecture
*   **Safe Replacement**: Uses temporary files and platform-specific scripts
*   **Compression Support**: Handles `.gz` and `.zip` compressed binaries
*   **Error Handling**: Graceful handling of network issues and missing releases

### Update Process:
1. Checks if current version is a snapshot
2. Fetches latest prerelease from GitHub
3. Downloads appropriate binary for platform
4. Creates temporary file (`judo.selfupdate` or `judo.selfupdate.exe`)
5. Sets executable permissions
6. Executes replacement script and restarts

## Development Conventions

*   **Command Structure:** The CLI uses the `cobra` library to define and organize its commands. New commands should be added by creating new `cobra.Command` objects and adding them to the `rootCmd`.
*   **Configuration:** The application's configuration is loaded from `judo.properties` files. The tool looks for a profile-specific properties file (e.g., `develop.properties`) first, and then falls back to a default `judo.properties` file.
*   **Error Handling:** The `checkError` function is used for basic error handling, which logs and exits on error.
*   **External Commands:** The CLI executes external commands like `docker`, `mvnd`, and `git` using Go's `os/exec` package.
*   **New Commands:** Add commands to `internal/commands/commands.go` and register in `cmd/judo/main.go`
