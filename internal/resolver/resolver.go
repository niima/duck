package resolver

import (
	"fmt"
	"sort"

	"duck/internal/config"
)

type DependencyResolver struct {
	projects map[string]*config.AppProject
}

func New(projects map[string]*config.AppProject) *DependencyResolver {
	return &DependencyResolver{
		projects: projects,
	}
}

type ResolutionResult struct {
	ExecutionOrder []string
	Dependencies   map[string][]string
}

func (r *DependencyResolver) ResolveExecutionOrder() (*ResolutionResult, error) {
	result := &ResolutionResult{
		Dependencies: make(map[string][]string),
	}

	graph := make(map[string][]string)
	inDegree := make(map[string]int)

	for key := range r.projects {
		graph[key] = []string{}
		inDegree[key] = 0
	}

	for key, project := range r.projects {
		for _, dep := range project.Config.Dependencies {
			if _, exists := r.projects[dep]; !exists {
				return nil, fmt.Errorf("project %s depends on %s, but %s was not found", key, dep, dep)
			}

			graph[dep] = append(graph[dep], key)
			inDegree[key]++

			result.Dependencies[key] = append(result.Dependencies[key], dep)
		}
	}

	queue := []string{}

	for key, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, key)
		}
	}

	sort.Strings(queue)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		result.ExecutionOrder = append(result.ExecutionOrder, current)

		dependents := graph[current]
		sort.Strings(dependents)

		for _, dependent := range dependents {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
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

	if len(result.ExecutionOrder) != len(r.projects) {
		return nil, fmt.Errorf("circular dependency detected")
	}

	return result, nil
}

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

func (r *DependencyResolver) ValidateDependencies() error {
	_, err := r.ResolveExecutionOrder()
	return err
}
