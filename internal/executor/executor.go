package executor

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"duck/internal/config"
)

// ExecutionResult represents the result of executing a script
type ExecutionResult struct {
	ProjectKey string
	Script     string
	Success    bool
	Output     string
	Error      string
	Duration   time.Duration
}

// Executor handles script execution across projects
type Executor struct {
	projectConfig *config.ProjectConfig
	projects      map[string]*config.AppProject
}

// New creates a new executor instance
func New(projectConfig *config.ProjectConfig, projects map[string]*config.AppProject) *Executor {
	return &Executor{
		projectConfig: projectConfig,
		projects:      projects,
	}
}

// ExecuteScript runs a script on a specific project
func (e *Executor) ExecuteScript(ctx context.Context, projectKey, scriptName string) (*ExecutionResult, error) {
	project, exists := e.projects[projectKey]
	if !exists {
		return nil, fmt.Errorf("project %s not found", projectKey)
	}

	script, exists := e.projectConfig.Scripts[scriptName]
	if !exists {
		return nil, fmt.Errorf("script %s not found", scriptName)
	}

	// Check if script is enabled for this project
	if enabled, exists := project.Config.Scripts[scriptName]; exists && !enabled {
		return &ExecutionResult{
			ProjectKey: projectKey,
			Script:     scriptName,
			Success:    false,
			Error:      "script disabled for this project",
		}, nil
	}

	result := &ExecutionResult{
		ProjectKey: projectKey,
		Script:     scriptName,
	}

	start := time.Now()
	defer func() {
		result.Duration = time.Since(start)
	}()

	// Determine working directory
	workingDir := project.Path
	if script.WorkingDir != "" {
		// Replace variables in working directory first
		expandedWorkingDir := e.replaceVariables(script.WorkingDir, project, project.Path)

		if filepath.IsAbs(expandedWorkingDir) {
			workingDir = expandedWorkingDir
		} else {
			workingDir = filepath.Join(project.Path, expandedWorkingDir)
		}
	}

	// Replace variables in command
	command := e.replaceVariables(script.Command, project, workingDir)

	// Prepare command
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = workingDir

	// Set environment variables
	cmd.Env = os.Environ()
	for key, value := range script.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}
	for key, value := range project.Config.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	// Capture output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		result.Error = fmt.Sprintf("failed to create stdout pipe: %v", err)
		return result, nil
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		result.Error = fmt.Sprintf("failed to create stderr pipe: %v", err)
		return result, nil
	}

	if err := cmd.Start(); err != nil {
		result.Error = fmt.Sprintf("failed to start command: %v", err)
		return result, nil
	}

	// Read output
	var outputBuilder, errorBuilder strings.Builder
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		copyOutput(stdout, &outputBuilder)
	}()

	go func() {
		defer wg.Done()
		copyOutput(stderr, &errorBuilder)
	}()

	wg.Wait()

	// Wait for command to complete
	if err := cmd.Wait(); err != nil {
		result.Success = false
		result.Error = errorBuilder.String()
		if result.Error == "" {
			result.Error = err.Error()
		}
	} else {
		result.Success = true
	}

	result.Output = outputBuilder.String()
	if result.Error == "" {
		result.Error = errorBuilder.String()
	}

	return result, nil
}

// ExecuteScriptOnProjects runs a script on multiple projects in order
func (e *Executor) ExecuteScriptOnProjects(ctx context.Context, projectKeys []string, scriptName string) ([]*ExecutionResult, error) {
	var results []*ExecutionResult

	for _, projectKey := range projectKeys {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}

		result, err := e.ExecuteScript(ctx, projectKey, scriptName)
		if err != nil {
			return results, err
		}
		results = append(results, result)

		// Stop on first failure if desired (could be configurable)
		if !result.Success {
			break
		}
	}

	return results, nil
}

// replaceVariables replaces template variables in command strings
func (e *Executor) replaceVariables(command string, project *config.AppProject, workingDir string) string {
	replacements := map[string]string{
		"{projectRoot}": project.Path,
		"{projectName}": project.Config.Name,
		"{namespace}":   project.Config.Namespace,
		"{workingDir}":  workingDir,
	}

	result := command
	for variable, value := range replacements {
		result = strings.ReplaceAll(result, variable, value)
	}

	return result
}

// copyOutput copies data from reader to writer
func copyOutput(reader io.Reader, writer io.Writer) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		fmt.Fprintln(writer, scanner.Text())
	}
}
