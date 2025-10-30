# Dependency Scanner

A flexible dependency scanner service for analyzing project dependencies across multiple programming languages.

## Architecture

The dependency scanner uses the **Adapter Pattern** to support multiple programming languages. Each language has its own scanner implementation that adheres to a common interface.

### Structure

```
internal/dependencyscanner/
├── scanner.go          # Core interfaces and data structures
├── registry.go         # Scanner registry for managing multiple scanners
├── go/                 # Go language scanner
│   ├── scanner.go      # Go-specific scanner implementation
│   ├── analyzer.go     # Deep analysis utilities
│   ├── graph.go        # Dependency graph builder
│   └── example_usage.go # Usage examples
└── js/                 # (Future) JavaScript scanner
```

## Core Interfaces

### Scanner Interface

All language-specific scanners must implement this interface:

```go
type Scanner interface {
    ScanProject(projectPath string) (*ProjectDependencies, error)
    GetLanguage() string
    CanScan(projectPath string) bool
}
```

### Data Structures

```go
type Dependency struct {
    Source      string   // The project that has the dependency
    Target      string   // The dependency itself
    Version     string   // Version of the dependency (if available)
    IsDirect    bool     // Whether it's a direct or indirect dependency
    ImportPaths []string // Specific import paths used
}

type ProjectDependencies struct {
    ProjectPath  string       // Path to the project
    Language     string       // Programming language
    Dependencies []Dependency // List of dependencies
}

type DependencyGraph struct {
    Projects map[string]*ProjectDependencies
}
```

## Go Scanner

The Go scanner analyzes Go projects by:

1. **Parsing go.mod files** to extract dependency declarations
2. **Scanning Go source files** to find actual import statements
3. **Combining both** to provide a complete dependency picture

### Features

- ✅ Parses `go.mod` files to extract dependencies
- ✅ Identifies direct vs indirect dependencies
- ✅ Scans Go source files for actual imports
- ✅ Tracks specific import paths used
- ✅ Builds complete dependency graphs
- ✅ Finds reverse dependencies (which projects use a package)

### Usage

#### CLI

```bash
# Scan all projects and show dependencies
./duck deps

# Show detailed information including import paths
./duck deps --verbose

# Include indirect dependencies
./duck deps --show-indirect
```

#### Programmatic

```go
import (
    goscan "duck/internal/dependencyscanner/go"
)

// Scan a single project
scanner := goscan.NewGoScanner()
deps, err := goscan.AnalyzeProjectDependencies("apps/namespace1/app1")

// Build a complete dependency graph
builder := goscan.NewGraphBuilder()
projectDirs := []string{
    "apps/namespace1/app1",
    "apps/namespace2/app2",
    "common",
    "httputils",
}
graph, err := builder.BuildGraph(".", projectDirs)

// Find all projects that depend on a specific package
dependents := builder.FindProjectDependencies(graph, "duck/common")
```

## Example Output

```
🔍 Scanning Go projects for dependencies...

Found 5 Go projects:

📦 apps/namespace1/app1
   Dependencies (2 direct):
     → duck/common (v0.0.0)
     → duck/httputils (v0.0.0)

📦 apps/namespace2/app2
   Dependencies (1 direct):
     → duck/common (v0.0.0)

📊 Dependency Summary:

  duck/common is used by:
    • apps/namespace1/app1
    • apps/namespace2/app2
    • apps/namespace2/app3
    • httputils

  duck/httputils is used by:
    • apps/namespace1/app1
    • apps/namespace2/app3
```

## Adding New Language Scanners

To add support for a new language (e.g., JavaScript/TypeScript):

1. Create a new directory: `internal/dependencyscanner/js/`
2. Implement the `Scanner` interface
3. Register the scanner in your application:

```go
import (
    "duck/internal/dependencyscanner"
    goscan "duck/internal/dependencyscanner/go"
    jsscan "duck/internal/dependencyscanner/js"
)

registry := dependencyscanner.NewScannerRegistry()
registry.RegisterScanner(goscan.NewGoScanner())
registry.RegisterScanner(jsscan.NewJsScanner())
```

## Future Enhancements

- [ ] JavaScript/TypeScript scanner
- [ ] Python scanner
- [ ] Rust scanner
- [ ] Dependency visualization (graph output)
- [ ] Circular dependency detection
- [ ] Dependency version conflict detection
- [ ] Export to various formats (JSON, GraphML, DOT)
- [ ] Integration with build tools
