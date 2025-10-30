package dependencyscanner

import (
	"fmt"
	"path/filepath"
)

// ScannerRegistry manages all available language scanners
type ScannerRegistry struct {
	scanners []Scanner
}

// NewScannerRegistry creates a new scanner registry
func NewScannerRegistry() *ScannerRegistry {
	return &ScannerRegistry{
		scanners: make([]Scanner, 0),
	}
}

// RegisterScanner adds a scanner to the registry
func (sr *ScannerRegistry) RegisterScanner(scanner Scanner) {
	sr.scanners = append(sr.scanners, scanner)
}

// FindScanner finds the appropriate scanner for a given project
func (sr *ScannerRegistry) FindScanner(projectPath string) (Scanner, error) {
	for _, scanner := range sr.scanners {
		if scanner.CanScan(projectPath) {
			return scanner, nil
		}
	}
	return nil, fmt.Errorf("no scanner found for project: %s", projectPath)
}

// ScanProjects scans multiple projects and builds a dependency graph
func (sr *ScannerRegistry) ScanProjects(projectPaths []string) (*DependencyGraph, error) {
	graph := NewDependencyGraph()

	for _, projectPath := range projectPaths {
		scanner, err := sr.FindScanner(projectPath)
		if err != nil {
			// Skip projects we can't scan
			continue
		}

		deps, err := scanner.ScanProject(projectPath)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project %s: %w", projectPath, err)
		}

		graph.AddProject(deps)
	}

	return graph, nil
}

// ScanProjectsRecursive scans projects recursively from a base directory
func (sr *ScannerRegistry) ScanProjectsRecursive(baseDir string, projectDirs []string) (*DependencyGraph, error) {
	graph := NewDependencyGraph()

	for _, projectDir := range projectDirs {
		projectPath := filepath.Join(baseDir, projectDir)
		scanner, err := sr.FindScanner(projectPath)
		if err != nil {
			// Skip projects we can't scan
			continue
		}

		deps, err := scanner.ScanProject(projectPath)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project %s: %w", projectPath, err)
		}

		graph.AddProject(deps)
	}

	return graph, nil
}
