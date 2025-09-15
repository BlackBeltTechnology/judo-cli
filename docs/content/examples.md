---
layout: page
title: "Examples"
permalink: /examples/
nav_order: 6
description: "Common usage patterns and workflows with JUDO CLI"
---

# Examples

This page provides common usage patterns, workflows, and real-world examples for using JUDO CLI effectively.

## Using the Included Test Model

A full sample project lives in `test-model/` and is ideal for end‑to‑end testing with real infrastructure (Karaf, PostgreSQL, Keycloak).

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

## Basic Workflows

### First Time Setup

```bash
# Navigate to your JUDO project
cd /path/to/your/judo-project

# Build the application (first time may take longer)
judo build

# Start all services
judo start

# Check everything is running
judo status

# View logs if needed
judo logs

# Access your application at http://localhost:8080
```

### Daily Development Workflow

```bash
# Quick morning startup
judo reckless  # Fast build + start

# Make code changes...

# Rebuild and restart
judo stop
judo build -a  # App module only for speed
judo start

# Or use quick mode for even faster iteration
judo stop
judo build -q  # Quick mode with caching
judo start
```

### End of Day Cleanup

```bash
# Stop all services
judo stop

# Clean up containers and volumes (optional)
judo clean

# Check nothing is running
judo status
```

## Environment-Specific Examples

### Local Development with Karaf

```bash
# Default karaf mode - fastest for development
judo build start

# Enable debug mode
judo -e debug start

# Custom configuration
echo "karaf.debug.port=9000" >> judo.properties
judo start
```

### Docker Compose Development

```bash
# Switch to compose mode for full containerization
judo -e compose-dev build start

# View all container logs
judo logs -f

# Restart specific service
docker-compose restart keycloak
judo status
```

### Production Deployment

```bash
# Build for production
judo -e production build

# Deploy with production settings
judo -e production start -d  # Detached mode

# Monitor production deployment
judo -e production status -v
judo -e production logs --since "1h"
```

## Database Operations

### Database Backup and Restore

```bash
# Create a backup before major changes
judo db dump backup_$(date +%Y%m%d_%H%M%S).sql

# Make your changes...

# Restore if needed
judo db import backup_20240130_143000.sql

# Quick backup with specific name
judo db dump before_migration.sql
```

### Development Data Reset

```bash
# Clean slate development
judo stop
judo clean -v  # Remove volumes (data)
judo start

# Or restore from a known good backup
judo db import dev_seed_data.sql
```

### Production Data Migration

```bash
# Production-safe migration
judo -e production db dump prod_backup_$(date +%Y%m%d).sql
judo -e production stop
judo -e production db import migration_script.sql
judo -e production start
judo -e production status
```

## Build Optimization Examples

### Fast Development Builds

```bash
# Skip tests for speed during development
echo "build.skip.tests=true" >> judo.properties
judo build

# App module only (fastest)
judo build -a

# Frontend only
judo build -f

# Ultimate speed for rapid iteration
judo reckless
```

### CI/CD Pipeline Build

```bash
# Full build with all validations
judo -e ci build

# With test coverage
echo "maven.profile=coverage" >> ci.properties
judo -e ci build

# Parallel build for faster CI
echo "build.parallel=true" >> ci.properties
judo -e ci build
```

### Production Build

```bash
# Production build with optimizations
judo -e production build

# With specific Maven profile
echo "maven.profile=production" >> production.properties
judo -e production build
```

## Troubleshooting Examples

### Port Conflicts

```bash
# Check what's using port 8080
lsof -i :8080

# Use different port
echo "karaf.port=9090" >> judo.properties
judo start

# Or for temporary override
JUDO_KARAF_PORT=9090 judo start
```

### Memory Issues

```bash
# Increase JVM memory
echo "jvm.memory.max=4g" >> judo.properties
judo restart

# Monitor memory usage
judo logs karaf | grep -i memory
```

### Build Failures

```bash
# Clean build
judo clean
judo build

# Verbose build for debugging
judo -v build

# Build specific module
judo build -a  # App only
judo build -f  # Frontend only
```

### Container Issues

```bash
# Check Docker status
docker info

# Restart Docker services
judo stop
docker system prune -f
judo start

# Check container logs
judo logs postgres
judo logs keycloak
```

## Advanced Workflows

### Multi-Environment Development

```bash
# Set up different environments
cat > local.properties << EOF
database.port=5433
karaf.port=8081
EOF

cat > integration.properties << EOF
database.host=integration-db.company.com
keycloak.port=8181
EOF

# Use different environments
judo -e local start        # Local with custom ports
judo -e integration test   # Integration testing
judo -e production deploy  # Production deployment
```

### Team Development Setup

```bash
# Shared team configuration
cat > team.properties << EOF
# Team shared settings
database.type=postgresql
keycloak.port=8180
maven.profile=team

# Skip frontend build for backend developers
frontend.build=false
EOF

# Individual developer overrides
cat > ${USER}.properties << EOF
# Personal settings
karaf.port=8080
database.port=5432
debug.enabled=true
EOF

# Use team settings with personal overrides
judo -e team -e ${USER} start
```

### Performance Testing

```bash
# Performance test environment
cat > perf.properties << EOF
# Performance test settings
jvm.memory.max=8g
database.pool.size=50
logging.level=WARN
monitoring.enabled=true
EOF

# Run performance tests
judo -e perf build start
judo -e perf status -v

# Monitor performance
judo -e perf logs --follow | grep -i performance
```

### Blue-Green Deployment

```bash
# Blue environment (current production)
judo -e blue status

# Green environment (new version)
judo -e green build
judo -e green start
judo -e green db import latest_prod_data.sql

# Test green environment
curl -f http://green.example.com/health

# Switch traffic (external load balancer)
# Then stop blue
judo -e blue stop
```

## Configuration Examples

### Development Team Settings

```properties
# dev-team.properties
app.name=myapp
runtime.mode=karaf
database.type=postgresql
build.skip.tests=true
debug.enabled=true
hot.reload=true
```

### Staging Environment

```properties
# staging.properties  
app.name=myapp-staging
runtime.mode=compose
database.host=staging-db.internal
keycloak.realm=staging
monitoring.enabled=true
logging.level=INFO
```

### Production Environment

```properties
# production.properties
app.name=myapp
runtime.mode=compose
database.host=prod-db.internal
database.pool.size=20
security.enabled=true
auth.strict=true
monitoring.enabled=true
logging.level=WARN
jvm.memory.max=4g
```

## Automation Examples

### Shell Scripts

```bash
#!/bin/bash
# dev-setup.sh - Quick development environment setup

set -e

echo "Setting up JUDO development environment..."

# Stop any running services
judo stop || true

# Clean environment
judo clean -f

# Build application
judo build

# Start services
judo start

# Wait for services to be ready
judo status --wait

echo "Development environment ready!"
echo "Application: http://localhost:8080"
echo "Keycloak: http://localhost:8180"
```

### CI/CD Pipeline

```yaml
# .github/workflows/ci.yml
name: CI Pipeline

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup JUDO CLI
        run: |
          curl -L https://github.com/BlackBeltTechnology/judo-cli/releases/latest/download/judo_Linux_x86_64.tar.gz | tar xz
          sudo mv judo /usr/local/bin/
          
      - name: Build and Test
        run: |
          judo -e ci build
          judo -e ci test
          
      - name: Cleanup
        run: |
          judo stop
          judo clean
```

### Docker Compose Integration

```yaml
# docker-compose.override.yml
version: '3.8'

services:
  app:
    environment:
      - JUDO_ENV=compose-dev
      - JUDO_DATABASE_HOST=postgres
      - JUDO_KEYCLOAK_URL=http://keycloak:8080
    depends_on:
      - postgres
      - keycloak
```

## See Also

- [Commands Reference](../commands/) - Complete command documentation
- [Configuration](../configuration/) - Environment and profile setup
- [Getting Started](../getting-started/) - Installation and basic usage