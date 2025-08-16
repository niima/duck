package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ProjectConfig represents the project.yaml configuration
type ProjectConfig struct {
	TargetDirectory string            `yaml:"targetDirectory"`
	Scripts         map[string]Script `yaml:"scripts"`
}

// Script represents a script command that can be run
type Script struct {
	Command     string            `yaml:"command"`
	Description string            `yaml:"description"`
	WorkingDir  string            `yaml:"workingDir,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty"`
}

// LoadProjectConfig loads and parses the project.yaml file
func LoadProjectConfig(path string) (*ProjectConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read project config: %w", err)
	}

	var config ProjectConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse project config: %w", err)
	}

	// Set default target directory if not specified
	if config.TargetDirectory == "" {
		config.TargetDirectory = "./apps"
	}

	// Convert to absolute path
	if !filepath.IsAbs(config.TargetDirectory) {
		absPath, err := filepath.Abs(config.TargetDirectory)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path for target directory: %w", err)
		}
		config.TargetDirectory = absPath
	}

	return &config, nil
}
