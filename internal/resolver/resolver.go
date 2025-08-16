package resolver

import (
	"fmt"
	"sort"

	"duck/internal/config"
)

// DependencyResolver handles dependency resolution and execution ordering
type DependencyResolver struct {
	projects map[string]*config.AppProject
}

// New creates a new dependency resolver
func New(projects map[string]*config.AppProject) *DependencyResolver {
	return &DependencyResolver{
		projects: projects,
	}
}

// ResolutionResult represents the result of dependency resolution
type ResolutionResult struct {
	ExecutionOrder []string            // Project keys in execution order
	Dependencies   map[string][]string // Map of project -> its dependencies
}

// ResolveExecutionOrder determines the order in which projects should be executed
// based on their dependencies
func (r *DependencyResolver) ResolveExecutionOrder() (*ResolutionResult, error) {
	result := &ResolutionResult{
		Dependencies: make(map[string][]string),
	}

	// Build dependency graph
	graph := make(map[string][]string)
	inDegree := make(map[string]int)

	// Initialize all projects
	for key := range r.projects {
		graph[key] = []string{}
		inDegree[key] = 0
	}

	// Build the dependency relationships
	for key, project := range r.projects {
		for _, dep := range project.Config.Dependencies {
			// Check if dependency exists
			if _, exists := r.projects[dep]; !exists {
				return nil, fmt.Errorf("project %s depends on %s, but %s was not found", key, dep, dep)
			}

			// Add edge from dependency to dependent
			graph[dep] = append(graph[dep], key)
			inDegree[key]++

			// Store in result
			result.Dependencies[key] = append(result.Dependencies[key], dep)
		}
	}

	// Topological sort using Kahn's algorithm
	queue := []string{}

	// Find all nodes with no incoming edges
	for key, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, key)
		}
	}

	// Sort the initial queue for consistent ordering
	sort.Strings(queue)

	for len(queue) > 0 {
		// Remove node from queue
		current := queue[0]
		queue = queue[1:]

		result.ExecutionOrder = append(result.ExecutionOrder, current)

		// Process all dependent nodes
		dependents := graph[current]
		sort.Strings(dependents) // For consistent ordering

		for _, dependent := range dependents {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				// Insert in sorted order
				inserted := false
				for i, item := range queue {
					if dependent < item {
						queue = append(queue[:i], append([]string{dependent}, queue[i:]...)...)
						inserted = true
						break
					}
				}
				if !inserted {
					queue = append(queue, dependent)
				}
			}
		}
	}

	// Check for circular dependencies
	if len(result.ExecutionOrder) != len(r.projects) {
		return nil, fmt.Errorf("circular dependency detected")
	}

	return result, nil
}

// GetDependents returns all projects that depend on the given project
func (r *DependencyResolver) GetDependents(projectKey string) []string {
	var dependents []string

	for key, project := range r.projects {
		for _, dep := range project.Config.Dependencies {
			if dep == projectKey {
				dependents = append(dependents, key)
				break
			}
		}
	}

	sort.Strings(dependents)
	return dependents
}

// ValidateDependencies checks if all dependencies exist and there are no circular dependencies
func (r *DependencyResolver) ValidateDependencies() error {
	_, err := r.ResolveExecutionOrder()
	return err
}
