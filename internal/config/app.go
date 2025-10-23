package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Name         string            `yaml:"name"`
	Namespace    string            `yaml:"namespace"`
	Description  string            `yaml:"description,omitempty"`
	Dependencies []string          `yaml:"dependencies,omitempty"`
	Scripts      map[string]bool   `yaml:"scripts,omitempty"`
	Tags         []string          `yaml:"tags,omitempty"`
	Environment  map[string]string `yaml:"environment,omitempty"`
}

type AppProject struct {
	Config *AppConfig
	Path   string
}

func LoadAppConfig(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read app config: %w", err)
	}

	var config AppConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse app config: %w", err)
	}

	if config.Name == "" {
		return nil, fmt.Errorf("app name is required")
	}

	if config.Namespace == "" {
		dir := filepath.Dir(path)
		parentDir := filepath.Dir(dir)
		config.Namespace = filepath.Base(parentDir)
	}

	return &config, nil
}
