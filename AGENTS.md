# Agent Guidelines for judo-cli

## Build/Test Commands
- Run: `go run ./cmd/judo`
- Build: `go build -o judo ./cmd/judo`
- Test all: `go test ./... -cover`
- Test single: `go test ./internal/commands -v -run TestName`
- Lint: `go fmt ./... && go vet ./...`
- Versioned build: `go build -ldflags "-X main.version=$(scripts/version.sh get) -X main.commit=$(git rev-parse --short HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ) -X main.builtBy=agent" -o judo ./cmd/judo`

## Code Style
- Go fmt/vet compliance required
- Packages: lowercase, files: snake_case
- Exported: CamelCase, errors: `ErrX`, wrap with `fmt.Errorf("context: %w", err)`
- Cobra: Set `Use`, `Short`, `Long`; flags use kebab-case
- Tests: `testing` + `testify`, table-driven patterns
- Imports: stdlib, third-party, local (grouped)

## Structure
- Commands: `internal/commands/*.go`
- Config: `internal/config` 
- Help: `internal/help`
- Tests: alongside code (`*_test.go`)
- Avoid breaking CLI UX; update `internal/help` and docs

## Frontend Integration
- Frontend: React app in `/frontend/` built with `npm run build`
- Embedded assets: Frontend build embedded via `//go:embed assets/*` in `internal/server/server.go:79`
- Server serves embedded assets first, falls back to `frontend/build/` directory
- Embedded assets include: `index.html`, static files, React bundle, manifest files
- Build process: Frontend must be built before Go compilation for embedding
- Server connectivity: Ensure port 6969 is available, check for empty response issues
