# Release Process

This document describes the automated release process for JUDO CLI using GoReleaser and GitHub Actions.

## Overview

The release process is fully automated and follows semantic versioning. There are three workflows:

1. **Continuous Integration** (`build.yml`) - Runs on every push to `develop` and `master`
2. **Release** (`release.yml`) - Triggers on tags (`vX.Y.Z`), generates release notes and publishes assets
3. **Manual Release** (`manual-release.yml`) - Bumps version on `develop` and pushes tag to trigger the release workflow

## Version Management

### Version File
The project uses a `VERSION` file in the root directory to track the current version. This file contains a single line with the semantic version number (e.g., `1.2.3`).

### Version Script
The `scripts/version.sh` script provides version management functionality:

```bash
# Get current version
./scripts/version.sh get

# Get build version (handles snapshot versions for develop branch)
./scripts/version.sh build

# Increment version
./scripts/version.sh increment patch    # 1.2.3 -> 1.2.4
./scripts/version.sh increment minor    # 1.2.3 -> 1.3.0
./scripts/version.sh increment major    # 1.2.3 -> 2.0.0

# Set specific version
./scripts/version.sh set 1.5.0

# Generate snapshot version
./scripts/version.sh snapshot           # 1.2.3 -> 1.2.3-snapshot-20231201120000
```

## Branching Strategy

### develop Branch
- All development happens on the `develop` branch
- Pushes to `develop` trigger snapshot builds with version format: `{version}-snapshot-{timestamp}`
- Only the latest snapshot release is kept (older snapshots are automatically deleted)
- After a release, the version is automatically incremented for the next development cycle

### master Branch
- The `master` branch contains stable, released code
- Builds from `master` use the version from the `VERSION` file
- No automatic releases are created from `master` branch pushes

### Tags
- Releases are triggered by pushing tags in the format `v{major}.{minor}.{patch}` (e.g., `v1.2.3`)
- Tags must follow semantic versioning exactly
- When a tag is pushed, it triggers the full release process

## Release Workflows

### 1. Develop Branch Builds

**Trigger:** Push to `develop` branch

**Process:**
1. Run tests and code quality checks
2. Generate snapshot version with timestamp
3. Build binaries using GoReleaser with `--snapshot` flag
4. Upload artifacts as GitHub Actions artifacts
5. Clean up old snapshot releases (keep only the latest)

**Artifacts:**
- Snapshot binaries for all platforms
- Available as GitHub Actions artifacts
- Not published as GitHub releases

### 2. Master Branch Builds

**Trigger:** Push to `master` branch

**Process:**
1. Run tests and code quality checks
2. Use version from `VERSION` file
3. Build binaries using GoReleaser with `build` command
4. Upload artifacts as GitHub Actions artifacts

**Artifacts:**
- Release-ready binaries for all platforms
- Available as GitHub Actions artifacts
- Not published as GitHub releases

### 3. Tag-Based Releases

**Trigger:** Push tag matching `v*` pattern

**Process:**
1. Validate tag format (must be `v{major}.{minor}.{patch}`)
2. Extract version from tag
3. Update `VERSION` file with tag version
4. Build and release using GoReleaser
5. Create GitHub release with:
   - Cross-platform binaries (Linux, macOS, Windows)
   - Both `amd64` and `arm64` architectures
   - Compressed archives (`.tar.gz` for Unix, `.zip` for Windows)
   - Checksums file
   - Release notes generated from git log and used by GoReleaser

**Post-Release:**
1. Checkout `develop` branch
2. Set `VERSION` file to the released version
3. Increment patch version for next development cycle
4. Commit version increment back to `develop`

### 4. Manual Releases

**Trigger:** Manual workflow dispatch from GitHub Actions UI

**Process:**
1. Choose version increment type (patch/minor/major) or specify custom version
2. Update `VERSION` file on `develop` branch
3. Commit version change
4. Create and push tag
5. Trigger automatic release workflow

## Supported Platforms

GoReleaser builds binaries for the following platforms:

- **Linux**: amd64, arm64
- **macOS**: amd64, arm64 (Intel and Apple Silicon)
- **Windows**: amd64

## Artifacts

Each release includes:

- `judo_Linux_x86_64.tar.gz` - Linux AMD64
- `judo_Linux_arm64.tar.gz` - Linux ARM64
- `judo_Darwin_x86_64.tar.gz` - macOS Intel
- `judo_Darwin_arm64.tar.gz` - macOS Apple Silicon
- `judo_Windows_x86_64.zip` - Windows AMD64
- `checksums.txt` - SHA256 checksums for all artifacts

## Making a Release

### Option 1: Manual Release (Recommended)

1. Go to GitHub Actions in the repository
2. Select the "Manual Release" workflow
3. Click "Run workflow"
4. Choose the version increment type or specify a custom version
5. Click "Run workflow"

The workflow will automatically:
- Update the version
- Create the tag
- Trigger the release process
- Update the develop branch

### Option 2: Tag-Based Release

1. Ensure you're on the `develop` branch with latest changes
2. Update the version manually:
   ```bash
   ./scripts/version.sh increment minor --commit
   ```
3. Create and push the tag:
   ```bash
   git tag v1.2.0
   git push origin v1.2.0
   ```

### Option 3: Command Line Release

1. Increment version and create tag:
   ```bash
   # Increment version and commit
   NEW_VERSION=$(./scripts/version.sh increment patch --commit)

   # Push version update
   git push origin develop

   # Create and push tag
   git tag "v${NEW_VERSION}"
   git push origin "v${NEW_VERSION}"
   ```

## Version Information

The built binaries include version information that can be displayed with:

```bash
judo version
```

This shows:
- Version number
- Git commit hash
- Build date
- Built by (goreleaser)
- Go version used

## Snapshot Versions

For development builds from the `develop` branch, snapshot versions are created with the format:
`{next_version}-snapshot-{timestamp}`

For example: `1.2.4-snapshot-20231201120000`

This ensures that:
- Development builds are clearly identified
- Each build has a unique version
- Snapshot builds don't interfere with actual releases

## Troubleshooting

### Release Failed
- Check that the tag follows the exact format `v{major}.{minor}.{patch}`
- Ensure the repository has proper permissions for the GitHub token
- Verify that GoReleaser configuration is valid

### Version Conflicts
- If there are version conflicts, manually update the `VERSION` file
- Ensure the develop branch is up to date before creating releases

### Missing Artifacts
- Check that all required platforms build successfully
- Verify GoReleaser configuration includes all desired platforms
- Check GitHub Actions logs for build errors

## Configuration Files

- `.goreleaser.yml` - GoReleaser configuration
- `VERSION` - Current version file
- `scripts/version.sh` - Version management script
- `.github/workflows/build.yml` - CI/CD workflow
- `.github/workflows/release.yml` - Release workflow
- `.github/workflows/manual-release.yml` - Manual release workflow
