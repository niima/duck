package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"duck/internal/config"
)

// Scanner handles discovering and managing application projects
type Scanner struct {
	projectConfig *config.ProjectConfig
	projects      map[string]*config.AppProject
}

// New creates a new scanner instance
func New(projectConfig *config.ProjectConfig) *Scanner {
	return &Scanner{
		projectConfig: projectConfig,
		projects:      make(map[string]*config.AppProject),
	}
}

// ScanProjects discovers all applications with app.yaml files
func (s *Scanner) ScanProjects() error {
	targetDir := s.projectConfig.TargetDirectory

	return filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip directories we can't read
			if os.IsPermission(err) {
				return nil
			}
			return err
		}

		// Look for app.yaml files
		if info.Name() == "app.yaml" {
			projectDir := filepath.Dir(path)

			// Load the app configuration
			appConfig, err := config.LoadAppConfig(path)
			if err != nil {
				fmt.Printf("Warning: Failed to load app config at %s: %v\n", path, err)
				return nil
			}

			// Create project key from namespace and app name
			projectKey := fmt.Sprintf("%s/%s", appConfig.Namespace, appConfig.Name)

			s.projects[projectKey] = &config.AppProject{
				Config: appConfig,
				Path:   projectDir,
			}
		}

		return nil
	})
}

// GetProjects returns all discovered projects
func (s *Scanner) GetProjects() map[string]*config.AppProject {
	return s.projects
}

// GetProject returns a specific project by key (namespace/name)
func (s *Scanner) GetProject(key string) (*config.AppProject, bool) {
	project, exists := s.projects[key]
	return project, exists
}

// GetProjectsByNamespace returns all projects in a specific namespace
func (s *Scanner) GetProjectsByNamespace(namespace string) []*config.AppProject {
	var projects []*config.AppProject

	for key, project := range s.projects {
		if strings.HasPrefix(key, namespace+"/") {
			projects = append(projects, project)
		}
	}

	return projects
}

// GetAvailableScripts returns all scripts that are enabled for a project
func (s *Scanner) GetAvailableScripts(project *config.AppProject) []string {
	var availableScripts []string

	for scriptName := range s.projectConfig.Scripts {
		// Check if the script is enabled for this project
		if enabled, exists := project.Config.Scripts[scriptName]; !exists || enabled {
			availableScripts = append(availableScripts, scriptName)
		}
	}

	return availableScripts
}
