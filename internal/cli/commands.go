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

func ListProjects(c *cli.Context) error {
	_, projects, err := LoadProjectData()
	if err != nil {
		return err
	}

	filtered := FilterProjects(projects, FilterOptions{
		Namespace: c.String("namespace"),
		Tags:      c.StringSlice("tag"),
	})

	if len(filtered) == 0 {
		fmt.Println("No projects found matching the criteria.")
		return nil
	}

	organized := OrganizeByNamespace(filtered)

	// Sort for consistent output
	var namespaces []string
	for ns := range organized {
		namespaces = append(namespaces, ns)
	}
	sort.Strings(namespaces)

	verbose := c.Bool("verbose")

	for _, namespace := range namespaces {
		fmt.Printf("📁 %s\n", namespace)

		projects := organized[namespace]
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

func RunScript(c *cli.Context) error {
	projectConfig, projects, err := LoadProjectData()
	if err != nil {
		return err
	}

	scriptName := c.String("script")
	if _, exists := projectConfig.Scripts[scriptName]; !exists {
		return fmt.Errorf("script '%s' not found", scriptName)
	}

	var targetProjects []string

	if c.Bool("all") {
		resolver := resolver.New(projects)
		resolution, err := resolver.ResolveExecutionOrder()
		if err != nil {
			return fmt.Errorf("failed to resolve dependencies: %w", err)
		}
		targetProjects = resolution.ExecutionOrder
	} else if projectNames := c.StringSlice("project"); len(projectNames) > 0 {
		for _, name := range projectNames {
			if _, exists := projects[name]; !exists {
				return fmt.Errorf("project '%s' not found", name)
			}
			targetProjects = append(targetProjects, name)
		}
	} else if namespace := c.String("namespace"); namespace != "" {
		for key, project := range projects {
			if project.Config.Namespace == namespace {
				targetProjects = append(targetProjects, key)
			}
		}
		sort.Strings(targetProjects)
	} else if tags := c.StringSlice("tag"); len(tags) > 0 {
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

func ConfigFormat(c *cli.Context) error {
	configPath := "duck.yaml"

	setFormat := c.String("set")

	if setFormat != "" {
		if setFormat != "duck" && setFormat != "nx" && setFormat != "all" {
			return fmt.Errorf("invalid format: must be 'duck', 'nx', or 'all'")
		}

		if err := UpdateProjectConfigFormat(configPath, setFormat); err != nil {
			return fmt.Errorf("failed to update config format: %w", err)
		}

		fmt.Printf("✅ Project configuration format updated to '%s'\n", setFormat)

		if setFormat == "nx" {
			fmt.Println("\n📝 Note: Duck will now look for 'project.json' files instead of 'app.yaml'")
			fmt.Println("   All Nx targets will be automatically available as scripts")
		} else if setFormat == "all" {
			fmt.Println("\n📝 Note: Duck will now look for both 'app.yaml' AND 'project.json' files")
			fmt.Println("   If both exist in the same directory, 'app.yaml' takes precedence")
			fmt.Println("   All Nx targets will be automatically available as scripts")
		} else {
			fmt.Println("\n📝 Note: Duck will now look for 'app.yaml' files")
		}

		return nil
	}

	projectConfig, err := LoadProjectConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Printf("Current project configuration format: %s\n", projectConfig.ProjectConfigFormat)

	if projectConfig.ProjectConfigFormat == "duck" {
		fmt.Println("Using Duck's app.yaml format")
	} else if projectConfig.ProjectConfigFormat == "nx" {
		fmt.Println("Using Nx's project.json format")
	} else if projectConfig.ProjectConfigFormat == "all" {
		fmt.Println("Using both Duck's app.yaml and Nx's project.json formats")
		fmt.Println("(app.yaml takes precedence when both exist in same directory)")
	}

	return nil
}
