# JUDO CLI Constitution

## Core Principles

### I. Simplicity & Single Responsibility
- Keep the project as a single Go CLI application with cohesive internal packages (`internal/*`) and a separate Hugo documentation site in `docs/`.
- Avoid unnecessary abstraction and patterns. Prefer direct usage of standard library and well‑established libraries (Cobra) over custom frameworks.
- Changes must be narrowly scoped; do not couple CLI features with documentation pipeline changes unless required.

### II. Consistent CLI UX (Cobra)
- Every command must define `Use`, `Short`, `Long`, and follow kebab‑case flags.
- Commands return non‑zero exit codes on failure and human‑readable errors to stderr.
- Backward compatibility is prioritized: do not break existing flags/commands without a deprecation path and release notes.

### III. Test‑First Discipline (Non‑negotiable)
- Use Go’s `testing` with table‑driven tests and `testify` where appropriate.
- Enforce `go fmt` and `go vet`. CI runs `go test ./...`, vet, and formatting checks.
- Write focused tests close to the changed code. Prefer integration tests for behavior that spans packages.

### IV. Versioning & Releases
- Semantic Versioning (MAJOR.MINOR.PATCH).
- `scripts/version.sh` is the single source of truth for the working version in `VERSION`.
- Snapshot builds originate from `develop` with tags like `vX.Y.Z-snapshot-YYYYMMDDHHMMSS` and never publish Homebrew updates.
- Stable releases are cut from `master` with tags `vX.Y.Z` and are published via GoReleaser.
- Homebrew formula updates happen only for stable releases; snapshots must not touch the tap.

### V. Observability & Error Handling
- Use clear, contextual errors wrapped with `fmt.Errorf("context: %w", err)`.
- Prefer structured log lines for non‑interactive diagnostics; keep CLI output concise by default.
- Ensure deterministic exit codes: 0 success; non‑zero for categorized failures.

### VI. Documentation (Hugo) & Deployment
- Hugo site lives in `docs/` and builds with the pinned Hugo and Node versions in CI (Hugo 0.150.x, Node 18).
- GitHub Pages deploys via `.github/workflows/hugo.yml`; dependency caching uses `docs/package-lock.json`.
- Pages environment name: `pages`. The workflow must not be blocked by environment rules intended for `github-pages`.
- **Documentation Integrity**: When implementing features or changes, corresponding documentation and CI scripts MUST be kept intact and updated to reflect the changes. Documentation updates are non-negotiable for feature implementations.

### VII. Security & Secrets
- All tokens (e.g., `HOMEBREW_TAP_TOKEN`) are provided via GitHub Secrets in workflows; never commit secrets.
- Least privilege: use `GITHUB_TOKEN` unless repository scoping requires a PAT.
- Validate third‑party actions by pinning major versions and reviewing deprecations.

## Development Workflow
- Default branch: `develop`; release branch: `master`.
- CI:
  - Build/Test workflow validates formatting, vetting, and tests on pushes/PRs.
  - Snapshot builds (develop) may create snapshot tags but must not update Homebrew.
  - Release builds (master) create a GitHub Release and update the Homebrew tap (branch `main`).
- Documentation is deployed from the current default branch per `hugo.yml` triggers; baseURL configured via `configure-pages` output.
- **CI/CD Integrity**: CI scripts and workflows MUST be maintained alongside code changes. Any feature that affects build, test, or deployment processes MUST include corresponding updates to CI/CD configuration and documentation.

## Additional Constraints
- Supported platforms: darwin (amd64, arm64), linux (amd64, arm64), windows (amd64).
- Minimum Go toolchain per CI: 1.25.
- Avoid breaking CLI UX; when unavoidable, provide deprecations and update `internal/help` and docs.

## Governance
- This constitution governs engineering practice for judo‑cli. Amendments require:
  1) updating this file, 2) syncing templates and helper guides, 3) noting the change in the checklist,
  4) adjusting workflows and docs if affected.
- PR reviews verify compliance with CLI UX rules, testing discipline, versioning, and documentation updates.
- Use conventional, meaningful commit messages focused on the “why”.

**Version**: 2.3.0 | **Ratified**: 2025-09-15 | **Last Amended**: 2025-09-16
