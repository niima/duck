package dependencyscanner

import (
	"fmt"
)

// Dependency represents a single dependency
type Dependency struct {
	Source      string   // The project that has the dependency
	Target      string   // The dependency itself
	Version     string   // Version of the dependency (if available)
	IsDirect    bool     // Whether it's a direct or indirect dependency
	ImportPaths []string // Specific import paths used
}

// ProjectDependencies represents all dependencies for a project
type ProjectDependencies struct {
	ProjectPath  string       // Path to the project
	Language     string       // Programming language
	Dependencies []Dependency // List of dependencies
}

// Scanner is the interface that all language-specific scanners must implement
type Scanner interface {
	// ScanProject scans a single project directory and returns its dependencies
	ScanProject(projectPath string) (*ProjectDependencies, error)

	// GetLanguage returns the language this scanner supports
	GetLanguage() string

	// CanScan checks if this scanner can handle the given project
	CanScan(projectPath string) bool
}

// DependencyGraph represents the dependency graph for multiple projects
type DependencyGraph struct {
	Projects map[string]*ProjectDependencies
}

// NewDependencyGraph creates a new empty dependency graph
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		Projects: make(map[string]*ProjectDependencies),
	}
}

// AddProject adds a project's dependencies to the graph
func (dg *DependencyGraph) AddProject(deps *ProjectDependencies) {
	dg.Projects[deps.ProjectPath] = deps
}

// GetDependencies returns all dependencies for a specific project
func (dg *DependencyGraph) GetDependencies(projectPath string) (*ProjectDependencies, error) {
	deps, ok := dg.Projects[projectPath]
	if !ok {
		return nil, fmt.Errorf("project not found: %s", projectPath)
	}
	return deps, nil
}

// GetProjectsWithDependencies returns all projects in the graph
func (dg *DependencyGraph) GetProjectsWithDependencies() []*ProjectDependencies {
	result := make([]*ProjectDependencies, 0, len(dg.Projects))
	for _, deps := range dg.Projects {
		result = append(result, deps)
	}
	return result
}
