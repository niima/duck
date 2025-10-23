package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type NxProjectConfig struct {
	Name        string                 `json:"name"`
	Schema      string                 `json:"$schema,omitempty"`
	ProjectType string                 `json:"projectType,omitempty"`
	SourceRoot  string                 `json:"sourceRoot,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Targets     map[string]NxTarget    `json:"targets,omitempty"`
	Metadata    map[string]interface{} `json:"-"`
}

type NxTarget struct {
	Executor    string                 `json:"executor,omitempty"`
	Options     map[string]interface{} `json:"options,omitempty"`
	Inputs      []interface{}          `json:"inputs,omitempty"`
	Outputs     []string               `json:"outputs,omitempty"`
	DependsOn   []interface{}          `json:"dependsOn,omitempty"`
	Description string                 `json:"description,omitempty"`
}

func LoadNxProjectConfig(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read nx project config: %w", err)
	}

	var nxConfig NxProjectConfig
	if err := json.Unmarshal(data, &nxConfig); err != nil {
		return nil, fmt.Errorf("failed to parse nx project config: %w", err)
	}

	if nxConfig.Name == "" {
		return nil, fmt.Errorf("project name is required")
	}

	appConfig := &AppConfig{
		Name:         nxConfig.Name,
		Description:  fmt.Sprintf("%s project", nxConfig.ProjectType),
		Tags:         nxConfig.Tags,
		Scripts:      make(map[string]bool),
		Dependencies: extractDependencies(nxConfig.Targets),
		Environment:  make(map[string]string),
	}

	dir := filepath.Dir(path)
	parentDir := filepath.Dir(dir)
	appConfig.Namespace = filepath.Base(parentDir)

	for targetName := range nxConfig.Targets {
		appConfig.Scripts[targetName] = true
	}

	return appConfig, nil
}

func extractDependencies(targets map[string]NxTarget) []string {
	dependencySet := make(map[string]bool)

	for _, target := range targets {
		if target.DependsOn != nil {
			for _, dep := range target.DependsOn {
				switch v := dep.(type) {
				case string:
					if strings.Contains(v, "^") {
						cleanDep := strings.TrimPrefix(v, "^")
						dependencySet[cleanDep] = true
					}
				case map[string]interface{}:
					if projects, ok := v["projects"].([]interface{}); ok {
						for _, proj := range projects {
							if projStr, ok := proj.(string); ok {
								dependencySet[projStr] = true
							}
						}
					}
					if target, ok := v["target"].(string); ok {
						if strings.Contains(target, "^") {
							cleanDep := strings.TrimPrefix(target, "^")
							dependencySet[cleanDep] = true
						}
					}
				}
			}
		}
	}

	var dependencies []string
	for dep := range dependencySet {
		dependencies = append(dependencies, dep)
	}

	return dependencies
}

func ConvertNxTargetsToScripts(nxConfig *NxProjectConfig, projectRoot string) map[string]Script {
	scripts := make(map[string]Script)

	for targetName, target := range nxConfig.Targets {
		script := Script{
			Description: target.Description,
			WorkingDir:  "{projectRoot}",
			Environment: make(map[string]string),
		}

		if target.Options != nil {
			if command, ok := target.Options["command"].(string); ok {
				script.Command = replaceNxVariables(command, projectRoot)
			} else if commands, ok := target.Options["commands"].([]interface{}); ok {
				var cmdParts []string
				for _, cmd := range commands {
					if cmdStr, ok := cmd.(string); ok {
						cmdParts = append(cmdParts, replaceNxVariables(cmdStr, projectRoot))
					}
				}
				script.Command = strings.Join(cmdParts, " && ")
			}
		}

		if script.Command == "" {
			script.Command = fmt.Sprintf("echo 'Target %s has no command defined'", targetName)
			if script.Description == "" {
				script.Description = fmt.Sprintf("Nx target: %s", targetName)
			}
		}

		scripts[targetName] = script
	}

	return scripts
}

func replaceNxVariables(command string, projectRoot string) string {
	replacements := map[string]string{
		"{projectRoot}":   "{projectRoot}",
		"{workspaceRoot}": ".",
		"{projectName}":   "{projectName}",
	}

	result := command
	for nxVar, duckVar := range replacements {
		result = strings.ReplaceAll(result, nxVar, duckVar)
	}

	return result
}

func ScanNxTargets(targetDirectory string) (map[string]Script, error) {
	scriptsMap := make(map[string]Script)
	targetNames := make(map[string]bool)

	err := filepath.Walk(targetDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsPermission(err) {
				return nil
			}
			return err
		}

		if info.Name() == "project.json" {
			data, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			var nxConfig NxProjectConfig
			if err := json.Unmarshal(data, &nxConfig); err != nil {
				return nil
			}

			projectDir := filepath.Dir(path)
			for targetName, target := range nxConfig.Targets {
				targetNames[targetName] = true

				if _, exists := scriptsMap[targetName]; !exists {
					script := Script{
						Description: target.Description,
						WorkingDir:  "{projectRoot}",
						Environment: make(map[string]string),
					}

					if target.Options != nil {
						if command, ok := target.Options["command"].(string); ok {
							script.Command = replaceNxVariables(command, projectDir)
						} else if commands, ok := target.Options["commands"].([]interface{}); ok {
							var cmdParts []string
							for _, cmd := range commands {
								if cmdStr, ok := cmd.(string); ok {
									cmdParts = append(cmdParts, replaceNxVariables(cmdStr, projectDir))
								}
							}
							script.Command = strings.Join(cmdParts, " && ")
						}
					}

					if script.Description == "" {
						script.Description = fmt.Sprintf("Nx target: %s", targetName)
					}

					scriptsMap[targetName] = script
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan nx targets: %w", err)
	}

	return scriptsMap, nil
}
