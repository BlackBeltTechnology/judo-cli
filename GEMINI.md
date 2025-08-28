# GEMINI.md

## Project Overview

This project is a command-line interface (CLI) tool named `judo-cli`, written in Go. It is designed to manage the lifecycle of JUDO applications. The tool utilizes the Cobra library for its command structure and interacts with several external tools and services:

*   **Docker:** Manages Docker containers for services like PostgreSQL and Keycloak.
*   **Maven:** Used for building the application. The tool can execute Maven commands with various flags.
*   **Karaf:** The application appears to run in an Apache Karaf container, which the CLI can start, stop, and configure.

The CLI reads its configuration from `judo.properties` files, with the ability to use different environments via a `--env` or `-e` flag.

## Building and Running

### Building the CLI

To build the `judo-cli` executable, run the following command:

```sh
go build judo.go
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

## Development Conventions

*   **Command Structure:** The CLI uses the `cobra` library to define and organize its commands. New commands should be added by creating new `cobra.Command` objects and adding them to the `rootCmd`.
*   **Configuration:** The application's configuration is loaded from `judo.properties` files. The tool looks for a profile-specific properties file (e.g., `develop.properties`) first, and then falls back to a default `judo.properties` file.
*   **Error Handling:** The `checkError` function is used for basic error handling, which logs and exits on error.
*   **External Commands:** The CLI executes external commands like `docker`, `mvnd`, and `git` using Go's `os/exec` package.
