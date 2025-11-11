package cli

import (
	"fmt"
	"os"
	"strings"

	"duck/internal/config"
	"duck/internal/scanner"
)

type FilterOptions struct {
	Namespace string
	Tags      []string
}

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

func FilterProjects(projects map[string]*config.AppProject, opts FilterOptions) map[string]*config.AppProject {
	filtered := make(map[string]*config.AppProject)

	for key, project := range projects {
		if opts.Namespace != "" && project.Config.Namespace != opts.Namespace {
			continue
		}

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

func OrganizeByNamespace(projects map[string]*config.AppProject) map[string][]*config.AppProject {
	organized := make(map[string][]*config.AppProject)

	for _, project := range projects {
		namespace := project.Config.Namespace
		organized[namespace] = append(organized[namespace], project)
	}

	return organized
}

func LoadProjectConfig(configPath string) (*config.ProjectConfig, error) {
	return config.LoadProjectConfig(configPath)
}

func UpdateProjectConfigFormat(configPath string, format string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	updated := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "projectConfigFormat:") {
			lines[i] = fmt.Sprintf("projectConfigFormat: \"%s\"", format)
			updated = true
			break
		}
	}

	if !updated {
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "targetDirectory:") {
				newLines := make([]string, 0, len(lines)+3)
				newLines = append(newLines, lines[:i+1]...)
				newLines = append(newLines, "")
				newLines = append(newLines, "# Project configuration format: \"duck\" or \"nx\"")
				newLines = append(newLines, fmt.Sprintf("projectConfigFormat: \"%s\"", format))
				newLines = append(newLines, lines[i+1:]...)
				lines = newLines
				break
			}
		}
	}

	output := strings.Join(lines, "\n")
	if err := os.WriteFile(configPath, []byte(output), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ResolveProjectKey resolves a project name or key to the actual project key
// This allows users to reference projects by their name (e.g., "sending-api")
// or by their path (e.g., "core-event/sending-api")
func ResolveProjectKey(projectIdentifier string, projects map[string]*config.AppProject) (string, bool) {
	// First, check if it's a direct key match (path-based)
	if _, exists := projects[projectIdentifier]; exists {
		return projectIdentifier, true
	}

	// If not found, try to find by project name
	for key, project := range projects {
		if project.Config.Name == projectIdentifier {
			return key, true
		}
	}

	return "", false
}
