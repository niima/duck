package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"duck/internal/config"
	goscan "duck/internal/dependencyscanner/go"
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
			// Find the project key for this project
			var projectKey string
			for key, p := range filtered {
				if p == project {
					projectKey = key
					break
				}
			}

			fmt.Printf("  🦆 %s", project.Config.Name)
			// Show path in parentheses if it differs from name
			if projectKey != "" && projectKey != project.Config.Name {
				fmt.Printf(" (%s)", projectKey)
			}
			fmt.Println()

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
			// Resolve project name or key to actual project key
			projectKey, exists := ResolveProjectKey(name, projects)
			if !exists {
				return fmt.Errorf("project '%s' not found", name)
			}
			targetProjects = append(targetProjects, projectKey)
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

		fmt.Printf("Project configuration format updated to '%s'\n", setFormat)

		if setFormat == "nx" {
			fmt.Println("\nNote: Duck will now look for 'project.json' files instead of 'app.yaml'")
			fmt.Println("   All Nx targets will be automatically available as scripts")
		} else if setFormat == "all" {
			fmt.Println("\nNote: Duck will now look for both 'app.yaml' AND 'project.json' files")
			fmt.Println("   If both exist in the same directory, 'app.yaml' takes precedence")
			fmt.Println("   All Nx targets will be automatically available as scripts")
		} else {
			fmt.Println("\nNote: Duck will now look for 'app.yaml' files")
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

func AnalyzeDependencies(c *cli.Context) error {
	workspaceRoot := c.String("workspace")
	if workspaceRoot == "" {
		workspaceRoot = "."
	}

	// Convert workspace root to absolute path
	absWorkspaceRoot, err := filepath.Abs(workspaceRoot)
	if err != nil {
		return fmt.Errorf("failed to get absolute workspace path: %w", err)
	}

	// Change to workspace directory to load configuration
	originalCwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	if err := os.Chdir(absWorkspaceRoot); err != nil {
		return fmt.Errorf("failed to change to workspace directory: %w", err)
	}
	defer os.Chdir(originalCwd)

	// Load projects from configuration
	_, allProjects, err := LoadProjectData()
	if err != nil {
		return fmt.Errorf("failed to load project data: %w", err)
	}

	// Build a set of all local packages
	localPackages := make(map[string]bool)
	for _, project := range allProjects {
		// Extract module name from go.mod
		goModPath := filepath.Join(project.Path, "go.mod")
		if data, err := os.ReadFile(goModPath); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "module ") {
					moduleName := strings.TrimSpace(strings.TrimPrefix(trimmed, "module "))
					localPackages[moduleName] = true
					break
				}
			}
		}
	}

	// Extract project paths from loaded projects
	projectDirs := make([]string, 0)
	for _, project := range allProjects {
		// Get relative path from workspace root to project
		relPath, err := filepath.Rel(absWorkspaceRoot, project.Path)
		if err == nil {
			projectDirs = append(projectDirs, relPath)
		}
	}

	if len(projectDirs) == 0 {
		fmt.Println("No projects found in configuration.")
		return nil
	}

	fmt.Println("> Scanning Go projects for dependencies...\n")

	builder := goscan.NewGraphBuilder()
	graph, err := builder.BuildGraph(absWorkspaceRoot, projectDirs)
	if err != nil {
		return fmt.Errorf("failed to build dependency graph: %w", err)
	}

	projects := graph.GetProjectsWithDependencies()
	if len(projects) == 0 {
		fmt.Println("No Go projects found.")
		return nil
	}

	// Sort projects by path for consistent output
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].ProjectPath < projects[j].ProjectPath
	})

	verbose := c.Bool("verbose")
	showIndirect := c.Bool("show-indirect")

	fmt.Printf("Found %d Go projects:\n\n", len(projects))

	for _, project := range projects {
		fmt.Printf("%s\n", project.ProjectPath)

		// Filter to only internal dependencies
		var internalDeps []interface{}
		for _, dep := range project.Dependencies {
			if localPackages[dep.Target] {
				internalDeps = append(internalDeps, dep)
			}
		}

		// Count direct internal dependencies
		directCount := 0
		for _, d := range internalDeps {
			if d.(interface{}) != nil {
				dep := d.(interface{})
				// Type assertion to access IsDirect field
				if depObj, ok := dep.(interface{}); ok {
					_ = depObj
					directCount++
				}
			}
		}

		// Better approach - iterate and count
		directCount = 0
		for _, dep := range project.Dependencies {
			if localPackages[dep.Target] && dep.IsDirect {
				directCount++
			}
		}

		if directCount == 0 {
			fmt.Println("   No internal dependencies")
		} else {
			fmt.Printf("   Internal Dependencies (%d direct", directCount)
			if showIndirect {
				indirectCount := 0
				for _, dep := range project.Dependencies {
					if localPackages[dep.Target] && !dep.IsDirect {
						indirectCount++
					}
				}
				if indirectCount > 0 {
					fmt.Printf(", %d indirect", indirectCount)
				}
			}
			fmt.Println("):")

			for _, dep := range project.Dependencies {
				// Only show internal dependencies
				if !localPackages[dep.Target] {
					continue
				}

				if !showIndirect && !dep.IsDirect {
					continue
				}

				marker := "→"
				if !dep.IsDirect {
					marker = "⇢"
				}

				// Map module name to project path for display
				projectPath := mapGoModuleToProjectKey(dep.Target, allProjects)
				if projectPath == "" {
					projectPath = dep.Target // Fallback to module name if mapping fails
				}

				fmt.Printf("     %s %s", marker, projectPath)
				if dep.Version != "" {
					fmt.Printf(" (%s)", dep.Version)
				}
				if !dep.IsDirect {
					fmt.Printf(" [indirect]")
				}
				fmt.Println()

				if verbose && len(dep.ImportPaths) > 0 {
					fmt.Println("        Import paths:")
					for _, path := range dep.ImportPaths {
						fmt.Printf("          - %s\n", path)
					}
				}
			}
		}
		fmt.Println()
	}

	// Summary section
	fmt.Println("Dependency Summary:")
	fmt.Println()

	// Build a map of project paths to their dependents (also as project paths)
	projectPathToDependents := make(map[string][]string)

	// Show which projects depend on which packages (only internal)
	for pkg := range localPackages {
		dependents := builder.FindProjectDependencies(graph, pkg)
		if len(dependents) > 0 {
			// Map module name to project path
			pkgPath := mapGoModuleToProjectKey(pkg, allProjects)
			if pkgPath == "" {
				pkgPath = pkg // Fallback
			}

			// Map dependent module names to project paths too
			var mappedDependents []string
			for _, dep := range dependents {
				// dep might be a module name or project path, try to map it
				depPath := dep // dep is the project path from graph (already relative)
				mappedDependents = append(mappedDependents, depPath)
			}

			if len(mappedDependents) > 0 {
				projectPathToDependents[pkgPath] = mappedDependents
			}
		}
	}

	// Sort and display
	var sortedPaths []string
	for path := range projectPathToDependents {
		sortedPaths = append(sortedPaths, path)
	}
	sort.Strings(sortedPaths)

	for _, pkgPath := range sortedPaths {
		dependents := projectPathToDependents[pkgPath]
		fmt.Printf("  %s is used by:\n", pkgPath)
		for _, dep := range dependents {
			fmt.Printf("    • %s\n", dep)
		}
		fmt.Println()
	}

	// Sync dependencies if flag is set
	if c.Bool("sync") {
		fmt.Println("\nSyncing dependencies to configuration files...\n")

		for _, project := range projects {
			if len(project.Dependencies) == 0 {
				continue
			}

			fmt.Printf("- %s\n", project.ProjectPath)

			// Convert Go module dependencies to project keys (only internal)
			var projectKeys []string
			for _, dep := range project.Dependencies {
				if !dep.IsDirect {
					continue // Only sync direct dependencies
				}

				// Only sync internal dependencies
				if !localPackages[dep.Target] {
					continue
				}

				// Map Go module path to project key
				projectKey := mapGoModuleToProjectKey(dep.Target, allProjects)
				if projectKey != "" {
					projectKeys = append(projectKeys, projectKey)
					if verbose {
						fmt.Printf("    Mapped: %s -> %s\n", dep.Target, projectKey)
					}
				} else {
					if verbose {
						fmt.Printf("    Warning: Could not map %s to a project\n", dep.Target)
					}
				}
			}

			if len(projectKeys) > 0 {
				projectPath := filepath.Join(absWorkspaceRoot, project.ProjectPath)
				if err := syncDependenciesToConfig(projectPath, projectKeys); err != nil {
					fmt.Printf("    Error: %v\n", err)
				}
			} else {
				fmt.Printf("    No internal dependencies to sync\n")
			}
			fmt.Println()
		}

		fmt.Println("Dependency sync complete!")
	}

	return nil
}

// mapGoModuleToProjectKey maps Go module paths to project namespace/name format
// Returns the relative path from workspace root for clarity
func mapGoModuleToProjectKey(modulePath string, allProjects map[string]*config.AppProject) string {
	// Build a map of module names to project keys (which are already relative paths)
	moduleToPath := make(map[string]string)

	for projectKey, project := range allProjects {
		// Extract module name from go.mod
		goModPath := filepath.Join(project.Path, "go.mod")
		if data, err := os.ReadFile(goModPath); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "module ") {
					moduleName := strings.TrimSpace(strings.TrimPrefix(trimmed, "module "))
					// projectKey is already a relative path from workspace root
					moduleToPath[moduleName] = projectKey
					break
				}
			}
		}
	}

	// Try to find a direct match first
	if projectPath, exists := moduleToPath[modulePath]; exists {
		return projectPath
	}

	// If no direct match, try suffix matching based on the project path structure
	// For example: module "github.com/org/repo/packages/go/httputils" should match projectKey "packages/go/httputils"
	for localModule, projectPath := range moduleToPath {
		// Normalize paths to use forward slashes
		normalizedProjectPath := filepath.ToSlash(projectPath)
		normalizedModulePath := filepath.ToSlash(modulePath)
		normalizedLocalModule := filepath.ToSlash(localModule)

		// Check if the modulePath ends with the projectPath
		// e.g., "github.com/org/repo/packages/go/common" ends with "packages/go/common"
		if strings.HasSuffix(normalizedModulePath, "/"+normalizedProjectPath) ||
			strings.HasSuffix(normalizedModulePath, normalizedProjectPath) {
			return projectPath
		}

		// Check if the local module ends with the projectPath
		// e.g., "myrepo/packages/go/common" ends with "packages/go/common"
		if strings.HasSuffix(normalizedLocalModule, "/"+normalizedProjectPath) ||
			strings.HasSuffix(normalizedLocalModule, normalizedProjectPath) {
			// If the modulePath also matches the same pattern, return projectPath
			if normalizedModulePath == normalizedLocalModule {
				return projectPath
			}
		}

		// Check if the modulePath starts with the localModule (subpackage)
		if strings.HasPrefix(normalizedModulePath, normalizedLocalModule+"/") ||
			normalizedModulePath == normalizedLocalModule {
			return projectPath
		}
	}

	return ""
}

// syncDependenciesToConfig updates app.yaml or project.json with discovered dependencies
func syncDependenciesToConfig(projectPath string, dependencies []string) error {
	appYamlPath := filepath.Join(projectPath, "app.yaml")
	projectJsonPath := filepath.Join(projectPath, "project.json")

	hasAppYaml := false
	hasProjectJson := false

	if _, err := os.Stat(appYamlPath); err == nil {
		hasAppYaml = true
	}
	if _, err := os.Stat(projectJsonPath); err == nil {
		hasProjectJson = true
	}

	var errors []error

	// Update app.yaml if it exists
	if hasAppYaml {
		if err := updateAppYamlDependencies(appYamlPath, dependencies); err != nil {
			errors = append(errors, fmt.Errorf("failed to update app.yaml: %w", err))
		} else {
			fmt.Printf("    Updated app.yaml\n")
		}
	}

	// Update project.json if it exists
	if hasProjectJson {
		if err := updateProjectJsonDependencies(projectJsonPath, dependencies); err != nil {
			errors = append(errors, fmt.Errorf("failed to update project.json: %w", err))
		} else {
			fmt.Printf("    Updated project.json\n")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("sync errors: %v", errors)
	}

	if !hasAppYaml && !hasProjectJson {
		return fmt.Errorf("no app.yaml or project.json found")
	}

	return nil
}

// updateAppYamlDependencies updates the dependencies in an app.yaml file
func updateAppYamlDependencies(path string, dependencies []string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	var result []string
	inDependencies := false
	dependenciesFound := false
	indentLevel := ""
	existingDeps := make(map[string]bool)

	// First pass: collect existing dependencies
	for i := 0; i < len(lines); i++ {
		line := lines[i]

		if strings.HasPrefix(strings.TrimSpace(line), "dependencies:") {
			dependenciesFound = true
			inDependencies = true
			continue
		}

		if inDependencies {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "-") {
				// Extract dependency name
				dep := strings.TrimSpace(strings.TrimPrefix(trimmed, "-"))
				dep = strings.Trim(dep, "\"'")
				existingDeps[dep] = true
			} else if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
				// Not a dependency item, end of dependencies section
				break
			}
		}
	}

	// Merge: add new dependencies to existing ones
	allDeps := make(map[string]bool)
	for dep := range existingDeps {
		allDeps[dep] = true
	}
	for _, dep := range dependencies {
		allDeps[dep] = true
	}

	// Convert to sorted slice
	var mergedDeps []string
	for dep := range allDeps {
		mergedDeps = append(mergedDeps, dep)
	}
	sort.Strings(mergedDeps)

	// Second pass: rebuild file with merged dependencies
	inDependencies = false
	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Check if this is the dependencies line
		if strings.HasPrefix(strings.TrimSpace(line), "dependencies:") {
			inDependencies = true
			// Get the indent level for this section
			indentLevel = line[:len(line)-len(strings.TrimLeft(line, " \t"))]

			// Add the dependencies line
			result = append(result, line)

			// Skip old dependency entries
			for i+1 < len(lines) {
				nextLine := lines[i+1]
				trimmed := strings.TrimSpace(nextLine)
				// If it's a dependency item or empty, skip it
				if strings.HasPrefix(trimmed, "-") || (trimmed == "" && i+2 < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i+2]), "-")) {
					i++
					continue
				}
				break
			}

			// Add merged dependencies
			for _, dep := range mergedDeps {
				result = append(result, fmt.Sprintf("%s  - \"%s\"", indentLevel, dep))
			}

			inDependencies = false
			continue
		}

		if !inDependencies {
			result = append(result, line)
		}
	}

	// If dependencies weren't found, add them at the end
	if !dependenciesFound && len(mergedDeps) > 0 {
		result = append(result, "dependencies:")
		for _, dep := range mergedDeps {
			result = append(result, fmt.Sprintf("  - \"%s\"", dep))
		}
	}

	return os.WriteFile(path, []byte(strings.Join(result, "\n")), 0644)
}

// updateProjectJsonDependencies updates the implicitDependencies in a project.json file
// Uses JSON parsing to ensure proper formatting
func updateProjectJsonDependencies(path string, dependencies []string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// First, extract existing dependencies (even from malformed JSON)
	existingDeps := make(map[string]bool)
	lines := strings.Split(string(data), "\n")
	inDeps := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "\"implicitDependencies\"") {
			inDeps = true
			continue
		}
		if inDeps {
			if strings.Contains(trimmed, "]") {
				break
			}
			if strings.Contains(trimmed, "\"") {
				start := strings.Index(trimmed, "\"")
				end := strings.LastIndex(trimmed, "\"")
				if start != -1 && end != -1 && start < end {
					dep := trimmed[start+1 : end]
					if dep != "" && dep != "implicitDependencies" {
						existingDeps[dep] = true
					}
				}
			}
		}
	}

	// Merge with new dependencies
	allDeps := make(map[string]bool)
	for dep := range existingDeps {
		allDeps[dep] = true
	}
	for _, dep := range dependencies {
		allDeps[dep] = true
	}

	// Convert to sorted slice
	var mergedDeps []string
	for dep := range allDeps {
		mergedDeps = append(mergedDeps, dep)
	}
	sort.Strings(mergedDeps)

	// Try to parse JSON (might fail if malformed)
	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		// Try to fix common JSON errors
		fixed := string(data)
		// Fix: missing comma after } when followed by "
		fixed = strings.ReplaceAll(fixed, "}\n  \"", "},\n  \"")
		fixed = strings.ReplaceAll(fixed, "}\n \"", "},\n \"")
		// Fix: trailing comma before closing }
		fixed = strings.ReplaceAll(fixed, "],\n}", "]\n}")

		if err := json.Unmarshal([]byte(fixed), &jsonData); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}
	}

	// Update dependencies
	jsonData["implicitDependencies"] = mergedDeps

	// Marshal with proper indentation
	output, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}

	// Write with trailing newline
	return os.WriteFile(path, append(output, '\n'), 0644)
}
