package goscan

import (
	"duck/internal/dependencyscanner"
	"fmt"
	"path/filepath"
)

// GraphBuilder builds a dependency graph for multiple Go projects
type GraphBuilder struct {
	scanner  *GoScanner
	registry *dependencyscanner.ScannerRegistry
}

// NewGraphBuilder creates a new graph builder
func NewGraphBuilder() *GraphBuilder {
	scanner := NewGoScanner()
	registry := dependencyscanner.NewScannerRegistry()
	registry.RegisterScanner(scanner)

	return &GraphBuilder{
		scanner:  scanner,
		registry: registry,
	}
}

// BuildGraph scans all projects in the workspace and builds a dependency graph
func (gb *GraphBuilder) BuildGraph(workspaceRoot string, projectDirs []string) (*dependencyscanner.DependencyGraph, error) {
	graph := dependencyscanner.NewDependencyGraph()

	for _, projectDir := range projectDirs {
		projectPath := filepath.Join(workspaceRoot, projectDir)

		if !gb.scanner.CanScan(projectPath) {
			continue
		}

		deps, err := AnalyzeProjectDependencies(projectPath)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze project %s: %w", projectPath, err)
		}

		// Store relative path for better readability
		deps.ProjectPath = projectDir
		graph.AddProject(deps)
	}

	return graph, nil
}

// FindProjectDependencies finds which projects depend on a specific package
func (gb *GraphBuilder) FindProjectDependencies(graph *dependencyscanner.DependencyGraph, packageName string) []string {
	dependents := make([]string, 0)

	for _, project := range graph.GetProjectsWithDependencies() {
		for _, dep := range project.Dependencies {
			if dep.Target == packageName {
				dependents = append(dependents, project.ProjectPath)
				break
			}
		}
	}

	return dependents
}
