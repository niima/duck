package cli

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"duck/internal/executor"
	"duck/internal/resolver"

	"github.com/urfave/cli/v2"
)

// ListProjects implements the list command
func ListProjects(c *cli.Context) error {
	_, projects, err := LoadProjectData()
	if err != nil {
		return err
	}

	// Apply filters
	filtered := FilterProjects(projects, FilterOptions{
		Namespace: c.String("namespace"),
		Tags:      c.StringSlice("tag"),
	})

	if len(filtered) == 0 {
		fmt.Println("No projects found matching the criteria.")
		return nil
	}

	// Organize by namespace
	organized := OrganizeByNamespace(filtered)

	// Sort namespaces for consistent output
	var namespaces []string
	for ns := range organized {
		namespaces = append(namespaces, ns)
	}
	sort.Strings(namespaces)

	verbose := c.Bool("verbose")

	for _, namespace := range namespaces {
		fmt.Printf("📁 %s\n", namespace)

		projects := organized[namespace]
		// Sort projects within namespace
		sort.Slice(projects, func(i, j int) bool {
			return projects[i].Config.Name < projects[j].Config.Name
		})

		for _, project := range projects {
			fmt.Printf("  🦆 %s\n", project.Config.Name)

			if verbose {
				if project.Config.Description != "" {
					fmt.Printf("     Description: %s\n", project.Config.Description)
				}
				if len(project.Config.Dependencies) > 0 {
					fmt.Printf("     Dependencies: %s\n", strings.Join(project.Config.Dependencies, ", "))
				}
				if len(project.Config.Tags) > 0 {
					fmt.Printf("     Tags: %s\n", strings.Join(project.Config.Tags, ", "))
				}
				fmt.Printf("     Path: %s\n", project.Path)
			}
		}
		fmt.Println()
	}

	return nil
}

// RunScript implements the run command
func RunScript(c *cli.Context) error {
	projectConfig, projects, err := LoadProjectData()
	if err != nil {
		return err
	}

	scriptName := c.String("script")
	if _, exists := projectConfig.Scripts[scriptName]; !exists {
		return fmt.Errorf("script '%s' not found", scriptName)
	}

	// Determine which projects to run on
	var targetProjects []string

	if c.Bool("all") {
		// Resolve execution order for all projects
		resolver := resolver.New(projects)
		resolution, err := resolver.ResolveExecutionOrder()
		if err != nil {
			return fmt.Errorf("failed to resolve dependencies: %w", err)
		}
		targetProjects = resolution.ExecutionOrder
	} else if projectNames := c.StringSlice("project"); len(projectNames) > 0 {
		// Specific projects
		for _, name := range projectNames {
			if _, exists := projects[name]; !exists {
				return fmt.Errorf("project '%s' not found", name)
			}
			targetProjects = append(targetProjects, name)
		}
	} else if namespace := c.String("namespace"); namespace != "" {
		// All projects in namespace
		for key, project := range projects {
			if project.Config.Namespace == namespace {
				targetProjects = append(targetProjects, key)
			}
		}
		sort.Strings(targetProjects)
	} else if tags := c.StringSlice("tag"); len(tags) > 0 {
		// Projects with specific tags
		filtered := FilterProjects(projects, FilterOptions{Tags: tags})
		for key := range filtered {
			targetProjects = append(targetProjects, key)
		}
		sort.Strings(targetProjects)
	} else {
		return fmt.Errorf("must specify --all, --project, --namespace, or --tag")
	}

	if len(targetProjects) == 0 {
		fmt.Println("No projects match the selection criteria.")
		return nil
	}

	if c.Bool("dry-run") {
		fmt.Printf("Would run script '%s' on the following projects:\n", scriptName)
		for _, key := range targetProjects {
			project := projects[key]
			fmt.Printf("  - %s (%s)\n", project.Config.Name, project.Config.Namespace)
		}
		return nil
	}

	// Execute the script
	executor := executor.New(projectConfig, projects)
	ctx := context.Background()

	verbose := c.Bool("verbose")

	fmt.Printf("Running script '%s' on %d project(s)...\n\n", scriptName, len(targetProjects))

	for i, projectKey := range targetProjects {
		project := projects[projectKey]
		fmt.Printf("[%d/%d] Running on %s (%s)...", i+1, len(targetProjects), project.Config.Name, project.Config.Namespace)

		start := time.Now()
		result, err := executor.ExecuteScript(ctx, projectKey, scriptName)
		duration := time.Since(start)

		if err != nil {
			fmt.Printf(" ❌ ERROR\n")
			return fmt.Errorf("execution failed: %w", err)
		}

		if result.Success {
			fmt.Printf(" ✅ SUCCESS (%v)\n", duration.Truncate(time.Millisecond))
		} else {
			fmt.Printf(" ❌ FAILED (%v)\n", duration.Truncate(time.Millisecond))
		}

		if verbose || !result.Success {
			if result.Output != "" {
				fmt.Println("Output:")
				lines := strings.Split(strings.TrimSpace(result.Output), "\n")
				for _, line := range lines {
					fmt.Printf("  │ %s\n", line)
				}
			}
			if result.Error != "" && !result.Success {
				fmt.Println("Error:")
				lines := strings.Split(strings.TrimSpace(result.Error), "\n")
				for _, line := range lines {
					fmt.Printf("  │ %s\n", line)
				}
			}
		}
		fmt.Println()

		if !result.Success {
			return fmt.Errorf("script failed on %s", project.Config.Name)
		}
	}

	fmt.Printf("✅ Script '%s' completed successfully on all projects!\n", scriptName)
	return nil
}

// ListScripts implements the scripts command
func ListScripts(c *cli.Context) error {
	projectConfig, _, err := LoadProjectData()
	if err != nil {
		return err
	}

	fmt.Println("Available scripts:")

	var scriptNames []string
	for name := range projectConfig.Scripts {
		scriptNames = append(scriptNames, name)
	}
	sort.Strings(scriptNames)

	for _, name := range scriptNames {
		script := projectConfig.Scripts[name]
		fmt.Printf("  %s", name)
		if script.Description != "" {
			fmt.Printf(" - %s", script.Description)
		}
		fmt.Println()
		if c.Bool("verbose") {
			fmt.Printf("    Command: %s\n", script.Command)
		}
	}

	return nil
}
