# 🦆 Duck.go - Monorepo Management Tool

A simple, fast, and dependency-aware build tool for Go monorepos. Duck scans your project structure and runs scripts across multiple applications while respecting dependencies.

## Features

- **Automatic Project Discovery** - Scans directories for `app.yaml` or `project.json` files
- **Namespace Organization** - Projects organized by `apps/namespace/app-name` structure
- **Dependency Resolution** - Respects project dependencies and execution order
- **Tag-based Filtering** - Run scripts on projects with specific tags
- **Flexible Targeting** - Run on all projects, specific namespaces, or individual projects
- **Dry Run Support** - Preview what would be executed
- **Nx Compatibility** - Full support for Nx monorepo configuration format (lift-and-shift migration)

## Installation

TBD

## Quick Start

1. Create a `duck.yaml` file in your repository root:

```yaml
---
# Duck Monorepo Configuration
targetDirectory: "./apps"

scripts:
  build:
    command: "go build ."
    description: "Build the Go application"
    workingDir: "{projectRoot}"

  test:
    command: "go test -v ./..."
    description: "Run tests with verbose output"
    workingDir: "{projectRoot}"

  lint:
    command: "golangci-lint run --fix"
    description: "Run linter and fix issues"
    workingDir: "{projectRoot}"
```

2. Create `app.yaml` files in your project directories:

```yaml
---
name: user-service
namespace: core
description: "User management service"
dependencies:
  - "shared/database"
tags:
  - go
  - microservice
  - api
scripts:
  build: true
  test: true
  lint: true
environment:
  SERVICE_NAME: "user-service"
```

3. Start using Duck:

```bash
# List all projects
./duck list

# Run build on all projects
./duck run --script build --all

# Run tests on specific namespace
./duck run --script test --namespace core
```

## Nx Compatibility

Duck supports Nx's `project.json` format, allowing you to use Duck alongside Nx or migrate from Nx without changing your project structure.

### Switching to Nx Format

Switch between Duck's `app.yaml` format and Nx's `project.json` format:

```bash
# Switch to Nx format
./duck config format --set nx

# Switch back to Duck format
./duck config format --set duck

# Check current format
./duck config format
```

### Using Nx project.json Files

When using Nx format, Duck automatically:

- Scans for `project.json` files instead of `app.yaml`
- Converts Nx targets to Duck scripts
- Respects Nx project structure and dependencies
- Supports Nx variable substitution (`{projectRoot}`, `{workspaceRoot}`, etc.)

**Example Nx `project.json`:**

```json
{
  "name": "domain-configurator-api",
  "$schema": "../../../node_modules/nx/schemas/project-schema.json",
  "projectType": "application",
  "sourceRoot": "apps/core-shared/domain-configurator-api",
  "tags": ["api", "go", "backend"],
  "targets": {
    "build": {
      "executor": "@pangaea/nix:develop",
      "options": {
        "command": "go build -o bin/{projectName} ."
      }
    },
    "test": {
      "executor": "@pangaea/nix:develop",
      "options": {
        "command": "go test -v ./..."
      }
    },
    "lint": {
      "executor": "@pangaea/nix:develop",
      "options": {
        "command": "golangci-lint run --fix"
      }
    },
    "seed-test-data": {
      "executor": "@pangaea/nix:develop",
      "options": {
        "command": "cd {projectRoot}/scripts/seed_test_data && go run seed_test_data.go"
      }
    }
  }
}
```

### Duck Configuration for Nx

Update your `duck.yaml` to use Nx format:

```yaml
---
# Duck Monorepo Configuration
targetDirectory: "./apps"

# Set format to 'nx' to use project.json files
projectConfigFormat: "nx"

# Scripts are automatically discovered from Nx targets
# You can optionally override them here
scripts:
  build:
    command: "go build ."
    description: "Build the Go application"
    workingDir: "{projectRoot}"
```

### Using Duck with Nx

Once configured for Nx format, use Duck commands as normal:

```bash
# List all Nx projects
./duck list

# Run Nx targets using Duck
./duck run --script build --all
./duck run --script test --namespace core-shared
./duck run --script seed-test-data --project core-shared/domain-configurator-api

# List all available Nx targets
./duck scripts
```

### Migration Guide: Nx to Duck

Duck makes it easy to migrate from Nx or use both tools simultaneously:

1. **Keep using Nx format:**

   ```bash
   ./duck config format --set nx
   ```

2. **Or gradually migrate to Duck format:**

   - Set format to `duck`
   - Create `app.yaml` files alongside `project.json`
   - Duck will use the simpler YAML format

3. **Use both tools:**
   - Keep `project.json` files for Nx
   - Set Duck to Nx format
   - Use Duck for build orchestration, Nx for other features

### Benefits of Using Duck with Nx

- **Simpler CLI**: Lighter weight alternative to Nx CLI
- **Faster startup**: No Node.js dependency for simple operations
- **Familiar commands**: Similar to Nx but with Duck's simplicity
- **No lock-in**: Switch between formats anytime
- **Lift-and-shift**: Use existing Nx projects without modification

## Commands

### `duck list` - List Projects

Show all discovered projects with optional filtering.

```bash
# List all projects
./duck list

# List with detailed information
./duck list --verbose

# Filter by namespace
./duck list --namespace core

# Filter by tags
./duck list --tag microservice --tag api
```

**Example Output:**

```
📁 core
  🦆 user-service
  🦆 auth-service

📁 shared
  🦆 database
  🦆 logging
```

### `duck run` - Execute Scripts

Run scripts on selected projects with various targeting options.

```bash
# Run on all projects (respects dependency order)
./duck run --script build --all

# Run on specific project
./duck run --script test --project core/user-service

# Run on entire namespace
./duck run --script lint --namespace core

# Run on projects with specific tags
./duck run --script build --tag microservice

# Dry run (preview without execution)
./duck run --script build --all --dry-run

# Verbose output
./duck run --script test --all --verbose
```

**Example Output:**

```
Running script 'build' on 3 project(s)...

[1/3] Running on database (shared)... ✅ SUCCESS (150ms)
[2/3] Running on user-service (core)... ✅ SUCCESS (300ms)
[3/3] Running on auth-service (core)... ✅ SUCCESS (250ms)

✅ Script 'build' completed successfully on all projects!
```

### `duck scripts` - List Available Scripts

Show all available scripts defined in `project.yaml`.

```bash
# List all scripts
./duck scripts

# List with command details
./duck scripts --verbose
```

**Example Output:**

```
Available scripts:
  build - Build the Go application
  test - Run tests with verbose output
  lint - Run linter and fix issues
  format - Format Go code
```

## Configuration

### Project Configuration (`project.yaml`)

Global configuration file that defines scripts and target directory.

```yaml
---
# Directory to scan for applications
targetDirectory: "./apps"

# Global scripts that can be run on projects
scripts:
  build:
    command: "go build ."
    description: "Build the Go application"
    workingDir: "{projectRoot}"
    environment:
      CGO_ENABLED: "0"

  test:
    command: "go test -v ./..."
    description: "Run tests with verbose output"
    workingDir: "{projectRoot}"

  lint:
    command: "golangci-lint run --fix"
    description: "Run linter and fix issues"
    workingDir: "{projectRoot}"

  docker-build:
    command: "docker build -t {projectName}:latest ."
    description: "Build Docker image"
    workingDir: "{projectRoot}"
```

### Application Configuration (`app.yaml`)

Individual project configuration in each `apps/namespace/app-name/app.yaml`.

```yaml
---
name: user-service
namespace: core
description: "User management and authentication service"

# Dependencies (will be built first)
dependencies:
  - "shared/database"
  - "shared/logging"

# Tags for filtering
tags:
  - go
  - microservice
  - api
  - authentication

# Script enablement (inherits from duck.yaml)
scripts:
  build: true
  test: true
  lint: true
  docker-build: false # Disable this script for this project

# Project-specific environment variables
environment:
  SERVICE_NAME: "user-service"
  SERVICE_PORT: "8080"
  LOG_LEVEL: "info"
```

### Variable Substitution

Duck supports variable substitution in script commands and working directories:

- `{projectRoot}` - Full path to the project directory
- `{projectName}` - Name of the project
- `{namespace}` - Namespace of the project
- `{workingDir}` - Current working directory

## Project Structure Example

```
my-monorepo/
├── duck.yaml              # Global configuration
├── duck                      # Built CLI tool
├── apps/
│   ├── core/
│   │   ├── user-service/
│   │   │   ├── app.yaml
│   │   │   ├── main.go
│   │   │   └── go.mod
│   │   └── auth-service/
│   │       ├── app.yaml
│   │       ├── main.go
│   │       └── go.mod
│   ├── api/
│   │   └── gateway/
│   │       ├── app.yaml
│   │       ├── main.go
│   │       └── go.mod
│   └── shared/
│       ├── database/
│       │   ├── app.yaml
│       │   └── db.go
│       └── logging/
│           ├── app.yaml
│           └── logger.go
```

## Advanced Usage

### Complex Filtering

Combine multiple filters for precise targeting:

```bash
# Run tests on Go microservices in core namespace
./duck run --script test --namespace core --tag microservice --tag go

# Build all API projects
./duck run --script build --tag api

# Lint everything except disabled projects
./duck run --script lint --all
```

### Dependency Management

Duck automatically resolves dependencies and runs projects in the correct order:

```bash
# This will build shared/database first, then core services
./duck run --script build --all
```

### Dry Run for Safety

Always preview complex operations:

```bash
# See what would be built
./duck run --script build --all --dry-run

# Preview namespace rebuild
./duck run --script build --namespace core --dry-run
```

### Verbose Debugging

Get detailed output for troubleshooting:

```bash
# See all output and errors
./duck run --script test --all --verbose

# Debug specific project
./duck run --script build --project core/user-service --verbose
```

## Common Workflows

### Full Build Pipeline

```bash
# 1. Format all code
./duck run --script format --all

# 2. Run linting
./duck run --script lint --all

# 3. Run tests
./duck run --script test --all

# 4. Build all projects
./duck run --script build --all
```

### Selective Development

```bash
# Work on specific namespace
./duck run --script test --namespace core --verbose

# Build dependencies of a service
./duck list --verbose  # Find dependencies
./duck run --script build --project shared/database
./duck run --script build --project core/user-service
```

### CI/CD Integration

```bash
# CI Pipeline example
./duck run --script lint --all
./duck run --script test --all
./duck run --script build --all
./duck run --script docker-build --tag deployable
```

## Help and Documentation

Get help for any command:

```bash
./duck --help
./duck list --help
./duck run --help
./duck scripts --help
```

## License

MIT License - see LICENSE file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

---

## Roadmap

- [x] Initial release
- [x] Service discovery feature
- [x] Run commands against services
- [x] Nx config format compatibility
- [ ] Auto generate dependency graph
- [ ] Build only dependent services
- [ ] support for shared packages
- [ ] Build only changed services
- [ ] Golang build support
- [ ] Javascript build support
- [ ] Support for docker remote cache (artifact registry)
