package goscan

import (
	"fmt"
	"log"
)

// ExampleUsage demonstrates how to use the dependency scanner
// NOTE: This example uses the duck monorepo structure. Adapt projectDirs to your repository layout.
func ExampleUsage() {
	// Example: Scan a specific project
	fmt.Println("=== Example 1: Scanning a single project ===")
	scanner := NewGoScanner()

	if scanner.CanScan("apps/namespace1/app1") {
		deps, err := AnalyzeProjectDependencies("apps/namespace1/app1")
		if err != nil {
			log.Fatalf("Failed to analyze: %v", err)
		}

		fmt.Printf("Project: %s\n", deps.ProjectPath)
		fmt.Printf("Language: %s\n", deps.Language)
		fmt.Printf("Dependencies:\n")
		for _, dep := range deps.Dependencies {
			fmt.Printf("  - %s (%s)\n", dep.Target, dep.Version)
		}
	}

	fmt.Println()

	// Example: Build a complete dependency graph
	fmt.Println("=== Example 2: Building a complete dependency graph ===")
	builder := NewGraphBuilder()

	// NOTE: Adapt these paths to your repository structure
	// Use LoadProjectData() from CLI to auto-discover projects in production
	projectDirs := []string{
		"apps/namespace1/app1",
		"apps/namespace2/app2",
		"apps/namespace2/app3",
		"common",
		"httputils",
	}

	graph, err := builder.BuildGraph(".", projectDirs)
	if err != nil {
		log.Fatalf("Failed to build graph: %v", err)
	}

	fmt.Printf("Total projects scanned: %d\n", len(graph.Projects))
	fmt.Println()

	// Example: Find all projects that depend on a specific package
	fmt.Println("=== Example 3: Finding dependents of 'duck/common' ===")
	dependents := builder.FindProjectDependencies(graph, "duck/common")
	for _, dep := range dependents {
		fmt.Printf("  - %s\n", dep)
	}
}
