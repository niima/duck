package cli

import (
	"fmt"

	"duck/internal/config"
	"duck/internal/scanner"
)

// FilterOptions holds filter criteria
type FilterOptions struct {
	Namespace string
	Tags      []string
}

// LoadProjectData loads and scans projects
func LoadProjectData() (*config.ProjectConfig, map[string]*config.AppProject, error) {
	projectConfig, err := config.LoadProjectConfig("duck.yaml")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load project config: %w", err)
	}

	scanner := scanner.New(projectConfig)
	if err := scanner.ScanProjects(); err != nil {
		return nil, nil, fmt.Errorf("failed to scan projects: %w", err)
	}

	return projectConfig, scanner.GetProjects(), nil
}

// FilterProjects filters projects based on criteria
func FilterProjects(projects map[string]*config.AppProject, opts FilterOptions) map[string]*config.AppProject {
	filtered := make(map[string]*config.AppProject)

	for key, project := range projects {
		// Filter by namespace
		if opts.Namespace != "" && project.Config.Namespace != opts.Namespace {
			continue
		}

		// Filter by tags
		if len(opts.Tags) > 0 {
			hasAllTags := true
			for _, requiredTag := range opts.Tags {
				found := false
				for _, projectTag := range project.Config.Tags {
					if projectTag == requiredTag {
						found = true
						break
					}
				}
				if !found {
					hasAllTags = false
					break
				}
			}
			if !hasAllTags {
				continue
			}
		}

		filtered[key] = project
	}

	return filtered
}

// OrganizeByNamespace groups projects by namespace
func OrganizeByNamespace(projects map[string]*config.AppProject) map[string][]*config.AppProject {
	organized := make(map[string][]*config.AppProject)

	for _, project := range projects {
		namespace := project.Config.Namespace
		organized[namespace] = append(organized[namespace], project)
	}

	return organized
}
