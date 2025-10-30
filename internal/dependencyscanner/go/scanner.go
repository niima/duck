package goscan

import (
	"bufio"
	"duck/internal/dependencyscanner"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GoScanner implements the Scanner interface for Go projects
type GoScanner struct{}

// NewGoScanner creates a new Go scanner instance
func NewGoScanner() *GoScanner {
	return &GoScanner{}
}

// GetLanguage returns the language this scanner supports
func (gs *GoScanner) GetLanguage() string {
	return "go"
}

// CanScan checks if this scanner can handle the given project
func (gs *GoScanner) CanScan(projectPath string) bool {
	goModPath := filepath.Join(projectPath, "go.mod")
	_, err := os.Stat(goModPath)
	return err == nil
}

// ScanProject scans a Go project and returns its dependencies
func (gs *GoScanner) ScanProject(projectPath string) (*dependencyscanner.ProjectDependencies, error) {
	goModPath := filepath.Join(projectPath, "go.mod")

	file, err := os.Open(goModPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open go.mod: %w", err)
	}
	defer file.Close()

	deps := &dependencyscanner.ProjectDependencies{
		ProjectPath:  projectPath,
		Language:     "go",
		Dependencies: make([]dependencyscanner.Dependency, 0),
	}

	scanner := bufio.NewScanner(file)
	inRequireBlock := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// Check for require block
		if strings.HasPrefix(line, "require (") {
			inRequireBlock = true
			continue
		} else if strings.HasPrefix(line, "require ") {
			// Single line require
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				dep := gs.parseDependency(parts[1:])
				if dep != nil {
					deps.Dependencies = append(deps.Dependencies, *dep)
				}
			}
			continue
		}

		// Check for replace block
		if strings.HasPrefix(line, "replace (") {
			continue
		} else if strings.HasPrefix(line, "replace ") {
			// We'll track replaces but not add them as separate dependencies
			continue
		}

		// End of block
		if line == ")" {
			inRequireBlock = false
			continue
		}

		// Parse dependencies in require block
		if inRequireBlock {
			parts := strings.Fields(line)
			if len(parts) >= 1 {
				dep := gs.parseDependency(parts)
				if dep != nil {
					deps.Dependencies = append(deps.Dependencies, *dep)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading go.mod: %w", err)
	}

	return deps, nil
}

// parseDependency parses a dependency from go.mod line parts
func (gs *GoScanner) parseDependency(parts []string) *dependencyscanner.Dependency {
	if len(parts) < 1 {
		return nil
	}

	target := parts[0]
	version := ""
	isDirect := true

	if len(parts) >= 2 {
		version = parts[1]
	}

	// Check if it's an indirect dependency
	if len(parts) >= 3 && parts[2] == "//indirect" {
		isDirect = false
	}

	return &dependencyscanner.Dependency{
		Target:      target,
		Version:     version,
		IsDirect:    isDirect,
		ImportPaths: []string{target},
	}
}

// ScanImports scans all Go files in a project and returns actual import statements
// This is useful for finding which dependencies are actually used
func (gs *GoScanner) ScanImports(projectPath string) ([]string, error) {
	imports := make(map[string]bool)

	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-Go files
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip test files if desired
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		fileImports, err := gs.parseImportsFromFile(path)
		if err != nil {
			return err
		}

		for _, imp := range fileImports {
			imports[imp] = true
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Convert map to slice
	result := make([]string, 0, len(imports))
	for imp := range imports {
		result = append(result, imp)
	}

	return result, nil
}

// parseImportsFromFile extracts import statements from a Go file
func (gs *GoScanner) parseImportsFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	imports := make([]string, 0)
	scanner := bufio.NewScanner(file)
	inImportBlock := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// Check for import block
		if strings.HasPrefix(line, "import (") {
			inImportBlock = true
			continue
		} else if strings.HasPrefix(line, "import ") {
			// Single line import
			imp := gs.parseImportLine(line[7:])
			if imp != "" {
				imports = append(imports, imp)
			}
			continue
		}

		// End of import block
		if inImportBlock && line == ")" {
			inImportBlock = false
			continue
		}

		// Parse imports in block
		if inImportBlock {
			imp := gs.parseImportLine(line)
			if imp != "" {
				imports = append(imports, imp)
			}
		}

		// Stop parsing after imports (optimization)
		if !inImportBlock && !strings.HasPrefix(line, "import") && line != "package main" && !strings.HasPrefix(line, "package ") {
			break
		}
	}

	return imports, scanner.Err()
}

// parseImportLine parses a single import line and returns the import path
func (gs *GoScanner) parseImportLine(line string) string {
	line = strings.TrimSpace(line)

	// Remove quotes
	line = strings.Trim(line, "\"")

	// Handle aliased imports (e.g., alias "package")
	parts := strings.Fields(line)
	if len(parts) > 1 {
		return strings.Trim(parts[len(parts)-1], "\"")
	}

	return line
}
