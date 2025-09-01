---
layout: page
title: build
parent: Command Reference
nav_order: 2
description: "Build project components with configurable options"
---

# judo build

Build project components with configurable options.

## Synopsis

```bash
judo build [flags]
```

## Description

The `build` command executes a Maven build with configurable components and options. By default, it builds all components (model, backend, frontend, karaf) but can be customized to build specific parts or use different build strategies.

## Flags

### Version and Build Control
| Flag | Description |
|------|-------------|
| `-v, --version <version>` | Use given version as model and application version |
| `-p, --build-parallel` | Parallel maven build (log output can be chaotic) |
| `-q, --quick` | Quick mode: cache + skip validations |
| `-i, --ignore-checksum` | Ignore checksum errors and update checksums |

### Component Selection
| Flag | Description |
|------|-------------|
| `-a, --build-app-module` | Build app module only (backend) |
| `-f, --build-frontend-module` | Build frontend module only |
| `--docker` | Build Docker images |

### Skip Options
| Flag | Description |
|------|-------------|
| `--skip-model` | Skip model building |
| `--skip-backend` | Skip backend building |
| `--skip-frontend` | Skip frontend building |
| `--skip-karaf` | Skip Backend Karaf building |
| `--skip-schema` | Skip building schema migration image |

### Additional Options
| Flag | Description |
|------|-------------|
| `--build-schema-cli` | Build schema CLI standalone JAR file |
| `-m, --maven-argument <args>` | Add extra maven arguments (quoted) |

## Build Components

### Default Build Matrix
When run without flags, builds:
- ✅ Model generation
- ✅ Backend compilation
- ✅ Frontend compilation  
- ✅ Karaf packaging
- ✅ Schema migration image
- ❌ Docker images (requires `--docker`)
- ❌ Schema CLI (requires `--build-schema-cli`)

### Component Details

**Model Building**
- Generates Java classes from JUDO models
- Creates database schema definitions
- Validates model consistency

**Backend Building**
- Compiles Java application code
- Packages OSGi bundles
- Builds REST APIs and business logic

**Frontend Building**
- Compiles TypeScript/React application
- Bundles assets and dependencies
- Generates optimized production build

**Karaf Building**
- Creates Apache Karaf distribution
- Packages application as runnable archive
- Includes all required OSGi bundles

## Examples

### Basic Build
```bash
judo build
```
Builds all components with default settings.

### Quick Development Build
```bash
judo build -q
```
Fast build using cache and skipping validations.

### Parallel Build
```bash
judo build -p
```
Faster build using multiple CPU cores (output may be chaotic).

### Component-Specific Builds

#### Backend Only
```bash
judo build -a
```
Builds only the application module (backend + interceptors).

#### Frontend Only
```bash
judo build -f
```
Builds only the React frontend application.

#### Frontend Quick Mode
```bash
judo build -f -q
```
Fast frontend build with cache and minimal validation.

### Skip Components

#### Build Without Frontend
```bash
judo build --skip-frontend
```
Builds everything except the React frontend.

#### Build Without Karaf
```bash
judo build --skip-karaf
```
Builds code but skips Karaf packaging (useful for pure library projects).

#### Minimal Build
```bash
judo build --skip-frontend --skip-karaf --skip-schema
```
Builds only model and backend code.

### Docker and Schema Builds

#### Include Docker Images
```bash
judo build --docker
```
Builds Docker images in addition to standard components.

#### Build Schema CLI
```bash
judo build --build-schema-cli
```
Creates standalone schema management tool.

#### Full Production Build
```bash
judo build --docker --build-schema-cli
```
Builds everything including optional components.

### Version Control

#### Custom Version
```bash
judo build -v 1.2.0
```
Builds with specific version instead of SNAPSHOT.

#### Release Build
```bash
judo build -v 1.0.0 --docker --build-schema-cli
```
Complete release build with version and all components.

### Advanced Maven Options

#### Custom Maven Arguments
```bash
judo build -m "-rf :myproject-application-karaf-offline"
```
Resume build from specific module.

#### Multiple Maven Options
```bash
judo build -m "-X -Dmaven.test.skip=true"
```
Debug mode with skipped tests.

#### Ignore Checksum Issues
```bash
judo build -i
```
Bypass checksum validation (useful during development).

## Build Strategies

### Development Workflow
```bash
# Initial build
judo build

# Quick iterations
judo build -f -q          # Frontend changes
judo build -a             # Backend changes
judo build -q             # Full quick rebuild
```

### Testing Workflow
```bash
# Full clean build
judo clean && judo build

# Test specific components
judo build -a && judo start
judo build -f && # refresh browser
```

### Production Workflow
```bash
# Complete production build
judo build -v 1.0.0 --docker --build-schema-cli

# Verification build
judo build -p             # Parallel for speed
```

## Build Output

### Standard Output
```bash
$ judo build -f
Building frontend only...
[INFO] Scanning for projects...
[INFO] Building MyProject Frontend React 1.0.0-SNAPSHOT
[INFO] BUILD SUCCESS
[INFO] Total time: 45.234 s
```

### Parallel Build Output
```bash
$ judo build -p
[INFO] Using the MultiThreadedBuilder implementation with a thread count of 4
[WARNING] Build output may be chaotic due to parallel execution
```

### Quick Mode Output
```bash
$ judo build -q
[INFO] Using cache for faster builds
[INFO] Skipping model validation
[INFO] Skipping frontend preparation
```

## Performance Tips

1. **Use Quick Mode**: `-q` for development iterations
2. **Parallel Builds**: `-p` on multi-core systems
3. **Component-Specific**: Build only what changed
4. **Cache Maven**: Use `mvnd` (Maven Daemon) for faster restarts
5. **Skip Unnecessary**: Use `--skip-*` flags to avoid rebuilding unchanged components

## Common Issues

### Build Failures

#### Maven Dependency Issues
```bash
judo build -i           # Ignore checksums
judo build -m "-U"      # Force update dependencies
```

#### Disk Space Issues
```bash
judo clean              # Clean old artifacts
judo prune              # Remove untracked files
```

#### Memory Issues
```bash
# Reduce parallel threads
judo build -p -m "-T 2"

# Increase memory
judo build -m "-Xmx4g"
```

### Frontend Build Issues

#### Node.js Version Problems
```bash
# Let JUDO manage Node.js
judo build -f

# Skip Node.js preparation if already set up
judo build -f -q
```

#### TypeScript Errors
```bash
# Skip frontend temporarily
judo build --skip-frontend

# Build with detailed output
judo build -f -m "-X"
```

## Integration with Other Commands

### Reckless Mode
```bash
judo reckless
```
Equivalent to:
```bash
judo build -q --skip-karaf && judo start
```

### Build and Start
```bash
judo build && judo start
```
Common development pattern.

### Clean Build
```bash
judo clean && judo build
```
Fresh environment build.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Build successful |
| 1 | Build failed |
| 130 | Build interrupted (Ctrl+C) |

## See Also

- [`judo reckless`](reckless) - Fast build and run
- [`judo clean`](clean) - Clean build environment  
- [`judo start`](start) - Start built application
- [`judo prune`](prune) - Clean untracked files