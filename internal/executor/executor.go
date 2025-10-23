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

type ExecutionResult struct {
	ProjectKey string
	Script     string
	Success    bool
	Output     string
	Error      string
	Duration   time.Duration
}

type Executor struct {
	projectConfig *config.ProjectConfig
	projects      map[string]*config.AppProject
}

func New(projectConfig *config.ProjectConfig, projects map[string]*config.AppProject) *Executor {
	return &Executor{
		projectConfig: projectConfig,
		projects:      projects,
	}
}

func (e *Executor) ExecuteScript(ctx context.Context, projectKey, scriptName string) (*ExecutionResult, error) {
	project, exists := e.projects[projectKey]
	if !exists {
		return nil, fmt.Errorf("project %s not found", projectKey)
	}

	script, exists := e.projectConfig.Scripts[scriptName]
	if !exists {
		return nil, fmt.Errorf("script %s not found", scriptName)
	}

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

	workingDir := project.Path
	if script.WorkingDir != "" {
		expandedWorkingDir := e.replaceVariables(script.WorkingDir, project, project.Path)

		if filepath.IsAbs(expandedWorkingDir) {
			workingDir = expandedWorkingDir
		} else {
			workingDir = filepath.Join(project.Path, expandedWorkingDir)
		}
	}

	command := e.replaceVariables(script.Command, project, workingDir)

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = workingDir

	cmd.Env = os.Environ()
	for key, value := range script.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}
	for key, value := range project.Config.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

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

		if !result.Success {
			break
		}
	}

	return results, nil
}

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

func copyOutput(reader io.Reader, writer io.Writer) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		fmt.Fprintln(writer, scanner.Text())
	}
}
