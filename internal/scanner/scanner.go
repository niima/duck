package scanner

import (
	"fmt"
	"os"
	"path/filepath"

	"duck/internal/config"
)

type Scanner struct {
	projectConfig *config.ProjectConfig
	projects      map[string]*config.AppProject
	workspaceRoot string // Cache the workspace root to avoid repeated os.Getwd() calls
}

func New(projectConfig *config.ProjectConfig) *Scanner {
	return &Scanner{
		projectConfig: projectConfig,
		projects:      make(map[string]*config.AppProject),
	}
}

func (s *Scanner) ScanProjects() error {
	// Get workspace root once for performance
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	s.workspaceRoot = cwd

	targetDir := s.projectConfig.TargetDirectory

	var scanAll bool
	var configFileNames []string

	switch s.projectConfig.ProjectConfigFormat {
	case config.FormatDuck:
		configFileNames = []string{"app.yaml"}
	case config.FormatNx:
		configFileNames = []string{"project.json"}
	case config.FormatAll:
		configFileNames = []string{"app.yaml", "project.json"}
		scanAll = true
	default:
		return fmt.Errorf("unsupported project config format: %s", s.projectConfig.ProjectConfigFormat)
	}

	// Scan target directory
	if err := s.scanDirectory(targetDir, configFileNames, scanAll); err != nil {
		return err
	}

	// Scan additional directories
	for _, additionalDir := range s.projectConfig.AdditionalDirectories {
		if err := s.scanDirectory(additionalDir, configFileNames, scanAll); err != nil {
			return err
		}
	}

	return nil
}

func (s *Scanner) scanDirectory(targetDir string, configFileNames []string, scanAll bool) error {
	return filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsPermission(err) {
				return nil
			}
			return err
		}

		for _, configFileName := range configFileNames {
			if info.Name() == configFileName {
				projectDir := filepath.Dir(path)

				if scanAll {
					var hasAppYaml, hasProjectJson bool
					if configFileName == "app.yaml" {
						hasAppYaml = true
						if _, err := os.Stat(filepath.Join(projectDir, "project.json")); err == nil {
							hasProjectJson = true
						}
					} else if configFileName == "project.json" {
						hasProjectJson = true
						if _, err := os.Stat(filepath.Join(projectDir, "app.yaml")); err == nil {
							hasAppYaml = true
						}
					}

					if hasAppYaml && hasProjectJson && configFileName == "project.json" {
						return nil
					}
				}

				var appConfig *config.AppConfig
				var loadErr error

				if configFileName == "app.yaml" {
					appConfig, loadErr = config.LoadAppConfig(path)
				} else if configFileName == "project.json" {
					appConfig, loadErr = config.LoadNxProjectConfig(path)
				}

				if loadErr != nil {
					fmt.Printf("Warning: Failed to load project config at %s: %v\n", path, loadErr)
					return nil
				}

				// Use relative path from workspace root as project key for consistency
				// Use cached workspace root for performance
				relPath, err := filepath.Rel(s.workspaceRoot, projectDir)
				if err != nil {
					// Fallback to namespace/name if relative path fails
					relPath = fmt.Sprintf("%s/%s", appConfig.Namespace, appConfig.Name)
				}

				projectKey := relPath

				s.projects[projectKey] = &config.AppProject{
					Config: appConfig,
					Path:   projectDir,
				}

				break
			}
		}

		return nil
	})
}

func (s *Scanner) GetProjects() map[string]*config.AppProject {
	return s.projects
}

func (s *Scanner) GetProject(key string) (*config.AppProject, bool) {
	project, exists := s.projects[key]
	return project, exists
}

func (s *Scanner) GetProjectsByNamespace(namespace string) []*config.AppProject {
	var projects []*config.AppProject

	for _, project := range s.projects {
		// Match by namespace field in config, not by key anymore
		if project.Config.Namespace == namespace {
			projects = append(projects, project)
		}
	}

	return projects
}

func (s *Scanner) GetAvailableScripts(project *config.AppProject) []string {
	var availableScripts []string

	for scriptName := range s.projectConfig.Scripts {
		if enabled, exists := project.Config.Scripts[scriptName]; !exists || enabled {
			availableScripts = append(availableScripts, scriptName)
		}
	}

	return availableScripts
}
