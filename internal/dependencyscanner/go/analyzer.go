package goscan

import (
	"duck/internal/dependencyscanner"
	"fmt"
	"path/filepath"
	"strings"
)

// AnalyzeProjectDependencies performs a deep analysis of Go project dependencies
// It combines go.mod parsing with actual import usage
func AnalyzeProjectDependencies(projectPath string) (*dependencyscanner.ProjectDependencies, error) {
	scanner := NewGoScanner()

	// First, get dependencies from go.mod
	deps, err := scanner.ScanProject(projectPath)
	if err != nil {
		return nil, err
	}

	// Then, scan actual imports to enrich the data
	imports, err := scanner.ScanImports(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to scan imports: %w", err)
	}

	// Create a map of used imports
	usedImports := make(map[string][]string)
	for _, imp := range imports {
		// Extract the base package (first part of the import path)
		basePkg := extractBasePackage(imp)
		usedImports[basePkg] = append(usedImports[basePkg], imp)
	}

	// Enrich dependencies with actual import paths
	for i := range deps.Dependencies {
		dep := &deps.Dependencies[i]
		if paths, ok := usedImports[dep.Target]; ok {
			dep.ImportPaths = paths
		}
	}

	return deps, nil
}

// extractBasePackage extracts the base package name from an import path
// For example: "github.com/user/repo/pkg" -> "github.com/user/repo"
func extractBasePackage(importPath string) string {
	// For local imports (e.g., "duck/common"), return as is
	if !strings.Contains(importPath, ".") {
		return importPath
	}

	// For external imports, take first 3 parts
	parts := strings.Split(importPath, "/")
	if len(parts) >= 3 {
		return filepath.Join(parts[0], parts[1], parts[2])
	}

	return importPath
}
