package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// AppConfig represents the app.yaml configuration for individual applications
type AppConfig struct {
	Name         string            `yaml:"name"`
	Namespace    string            `yaml:"namespace"`
	Description  string            `yaml:"description,omitempty"`
	Dependencies []string          `yaml:"dependencies,omitempty"`
	Scripts      map[string]bool   `yaml:"scripts,omitempty"`
	Tags         []string          `yaml:"tags,omitempty"`
	Environment  map[string]string `yaml:"environment,omitempty"`
}

// AppProject represents a discovered application project
type AppProject struct {
	Config *AppConfig
	Path   string // Full path to the project directory
}

// LoadAppConfig loads and parses an app.yaml file
func LoadAppConfig(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read app config: %w", err)
	}

	var config AppConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse app config: %w", err)
	}

	// Validate required fields
	if config.Name == "" {
		return nil, fmt.Errorf("app name is required")
	}

	// Extract namespace from directory structure if not specified
	if config.Namespace == "" {
		dir := filepath.Dir(path)
		parentDir := filepath.Dir(dir)
		config.Namespace = filepath.Base(parentDir)
	}

	return &config, nil
}
