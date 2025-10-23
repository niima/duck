package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ProjectConfigFormat string

const (
	FormatDuck ProjectConfigFormat = "duck"
	FormatNx   ProjectConfigFormat = "nx"
	FormatAll  ProjectConfigFormat = "all"
)

type ProjectConfig struct {
	TargetDirectory     string              `yaml:"targetDirectory"`
	ProjectConfigFormat ProjectConfigFormat `yaml:"projectConfigFormat"`
	Scripts             map[string]Script   `yaml:"scripts"`
}

type Script struct {
	Command     string            `yaml:"command"`
	Description string            `yaml:"description"`
	WorkingDir  string            `yaml:"workingDir,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty"`
}

func LoadProjectConfig(path string) (*ProjectConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read project config: %w", err)
	}

	var config ProjectConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse project config: %w", err)
	}

	if config.TargetDirectory == "" {
		config.TargetDirectory = "./apps"
	}

	if config.ProjectConfigFormat == "" {
		config.ProjectConfigFormat = FormatDuck
	}

	if config.ProjectConfigFormat != FormatDuck && config.ProjectConfigFormat != FormatNx && config.ProjectConfigFormat != FormatAll {
		return nil, fmt.Errorf("invalid projectConfigFormat: must be 'duck', 'nx', or 'all', got '%s'", config.ProjectConfigFormat)
	}

	if !filepath.IsAbs(config.TargetDirectory) {
		absPath, err := filepath.Abs(config.TargetDirectory)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path for target directory: %w", err)
		}
		config.TargetDirectory = absPath
	}

	if config.ProjectConfigFormat == FormatNx || config.ProjectConfigFormat == FormatAll {
		nxScripts, err := ScanNxTargets(config.TargetDirectory)
		if err != nil {
			fmt.Printf("Warning: Failed to scan Nx targets: %v\n", err)
		} else {
			if len(config.Scripts) == 0 {
				config.Scripts = nxScripts
			} else {
				for targetName, targetScript := range nxScripts {
					if _, exists := config.Scripts[targetName]; !exists {
						config.Scripts[targetName] = targetScript
					}
				}
			}
		}
	}

	return &config, nil
}
