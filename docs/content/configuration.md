---
title: "Configuration"
description: "Environment and profile configuration for JUDO CLI"
---

# Configuration

JUDO CLI uses a flexible configuration system based on properties files to manage different environments and deployment scenarios.

## Configuration Files

### Primary Configuration Files

#### `judo.properties`
The main configuration file that defines default settings for your JUDO application.

```properties
# Application Information
app.name=my-judo-app
app.schema=myapp
app.version=1.0.0

# Database Configuration
database.type=postgresql
database.host=localhost
database.port=5432
database.name=myapp_db
database.username=judo
database.password=judo

# Runtime Configuration
runtime.mode=karaf
karaf.port=8080
karaf.debug.port=8000

# Service Ports
keycloak.port=8180
postgres.port=5432

# Build Configuration
maven.profile=default
build.skip.tests=false
```

#### Environment-Specific Files
Create environment-specific configuration files that override the defaults:

- `compose-dev.properties` - Docker Compose development
- `compose-prod.properties` - Docker Compose production  
- `local.properties` - Local development overrides

#### `judo-version.properties`
Defines version constraints for JUDO dependencies:

```properties
# JUDO Framework versions
judo.version=1.0.0
judo.runtime.version=1.0.0
judo.frontend.version=1.0.0

# Dependency versions
karaf.version=4.4.3
postgresql.version=14
keycloak.version=21.1.1
```

## Configuration Properties

### Application Settings

| Property | Description | Default | Example |
|----------|-------------|---------|---------|
| `app.name` | Application name | - | `my-judo-app` |
| `app.schema` | Database schema name | - | `myapp` |
| `app.version` | Application version | `1.0.0` | `2.1.0` |
| `app.description` | Application description | - | `My JUDO Application` |

### Database Configuration

| Property | Description | Default | Example |
|----------|-------------|---------|---------|
| `database.type` | Database type | `postgresql` | `postgresql`, `hsqldb` |
| `database.host` | Database hostname | `localhost` | `db.example.com` |
| `database.port` | Database port | `5432` | `5432` |
| `database.name` | Database name | `{app.name}_db` | `myapp_db` |
| `database.username` | Database username | `judo` | `myuser` |
| `database.password` | Database password | `judo` | `secretpass` |
| `database.schema` | Database schema | `{app.schema}` | `myapp` |

### Runtime Configuration

| Property | Description | Default | Example |
|----------|-------------|---------|---------|
| `runtime.mode` | Runtime mode | `karaf` | `karaf`, `compose` |
| `karaf.port` | Karaf HTTP port | `8080` | `9090` |
| `karaf.debug.port` | Karaf debug port | `8000` | `9000` |
| `karaf.ssh.port` | Karaf SSH port | `8101` | `9101` |

### Service Ports

| Property | Description | Default | Example |
|----------|-------------|---------|---------|
| `keycloak.port` | Keycloak port | `8180` | `9180` |
| `postgres.port` | PostgreSQL port | `5432` | `5433` |
| `pgadmin.port` | PgAdmin port | `8080` | `8081` |

### Build Configuration

| Property | Description | Default | Example |
|----------|-------------|---------|---------|
| `maven.profile` | Maven build profile | `default` | `development`, `production` |
| `build.skip.tests` | Skip tests during build | `false` | `true`, `false` |
| `build.parallel` | Enable parallel builds | `true` | `true`, `false` |
| `frontend.build` | Build frontend | `true` | `true`, `false` |

## Environment Profiles

### Karaf Mode (Default)
Local development with Karaf runtime and Docker services.

```properties
# karaf.properties (or defaults in judo.properties)
runtime.mode=karaf
karaf.port=8080
database.type=postgresql
postgres.port=5432
keycloak.port=8180
```

### Compose Development
Full Docker Compose environment for development.

```properties
# compose-dev.properties
runtime.mode=compose
app.port=8080
database.type=postgresql
postgres.port=5432
keycloak.port=8180

# Docker-specific settings
docker.network=judo-dev
docker.compose.file=docker-compose-dev.yml
```

### Production Configuration
Production-ready settings with security considerations.

```properties
# compose-prod.properties
runtime.mode=compose
app.port=80
database.type=postgresql
database.host=prod-db.example.com
database.port=5432

# Security settings
keycloak.realm=production
auth.enabled=true

# Performance settings
jvm.memory.max=2g
database.pool.size=20
```

## Using Environment Profiles

### Command Line
Specify environment using the `-e` flag:

```bash
# Use compose-dev environment
judo -e compose-dev build start

# Use production environment  
judo -e production status

# Multiple environments
judo -e local build
judo -e compose-dev start
```

### Environment Variables
Set default environment via environment variable:

```bash
export JUDO_ENV=compose-dev
judo build start
```

### Profile Selection Priority
1. Command line `-e` flag
2. `JUDO_ENV` environment variable
3. `judo.properties` (default)

## Configuration Validation

JUDO CLI validates configuration at startup:

```bash
# Check configuration
judo config validate

# Show resolved configuration
judo config show

# Test specific environment
judo -e compose-dev config validate
```

## Advanced Configuration

### Property Interpolation
Use variables within configuration files:

```properties
# Base configuration
app.name=myapp
database.name=${app.name}_db
database.schema=${app.name}
keycloak.realm=${app.name}_realm
```

### Environment Variable Override
Override any property using environment variables:

```bash
# Override database host
export JUDO_DATABASE_HOST=remote-db.example.com

# Override application port
export JUDO_APP_PORT=9090

judo start
```

Convention: `JUDO_` + property path in uppercase with dots replaced by underscores.

### Conditional Configuration
Include configuration based on conditions:

```properties
# Include production security only in prod environment
[@prod]
security.enabled=true
auth.strict=true

[@dev]
debug.enabled=true
logging.level=DEBUG
```

## Configuration Examples

### Local Development
```properties
# judo.properties
app.name=myapp
runtime.mode=karaf
database.type=postgresql
build.skip.tests=true
```

### CI/CD Pipeline
```properties
# ci.properties  
app.name=myapp
runtime.mode=compose
database.type=postgresql
build.skip.tests=false
build.parallel=true
```

### Production Deployment
```properties
# production.properties
app.name=myapp
runtime.mode=compose
database.type=postgresql
database.host=prod-db.internal
security.enabled=true
monitoring.enabled=true
```

## Troubleshooting Configuration

### Common Issues

**Port conflicts:**
```bash
# Check what's using a port
lsof -i :8080

# Change port in configuration
echo "karaf.port=9090" >> judo.properties
```

**Database connection issues:**
```bash
# Test database connection
judo db test

# Check database configuration
judo config get database.*
```

**Environment not loading:**
```bash
# Verify environment file exists
ls -la *.properties

# Check configuration resolution
judo -e myenv config show
```

## See Also

- [Commands Reference](../commands/) - Command-line interface
- [Examples](../examples/) - Configuration examples and patterns
- [Getting Started](../getting-started/) - Basic setup and installation