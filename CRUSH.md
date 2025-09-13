# CRUSH.md - Judo CLI Development Guide

## Build & Test Commands
```bash
go build -o judo ./cmd/judo           # Build CLI binary
go test ./internal/...               # Run all tests
go test -v ./internal/...            # Verbose test output
go test ./internal/db/...            # Run specific package tests
go test -run TestName ./internal/... # Run single test
go vet ./...                         # Static analysis
go fmt ./...                         # Format code
go mod tidy                          # Clean dependencies
```

## Code Style Guidelines
- **Imports**: Group stdlib, third-party, local imports with blank lines
- **Formatting**: Use `go fmt` standard, 4-space indentation
- **Naming**: camelCase for variables/functions, PascalCase for exports
- **Error Handling**: Use `checkError` helper for basic error handling
- **Comments**: Eclipse Public License header in all files
- **Testing**: testify framework for assertions, skip complex integration tests

## Project Structure
- **cmd/judo**: CLI entry point with Cobra setup
- **internal/**: Modular packages (commands, config, docker, karaf, db, utils)
- **scripts/**: Version management and installation scripts
- **VERSION**: Semantic version file for build injection

## Key Dependencies
- Cobra for CLI structure
- Docker client for container management
- Testify for testing assertions
- Readline for interactive sessions