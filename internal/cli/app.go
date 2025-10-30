package cli

import (
	"github.com/urfave/cli/v2"
)

func CreateApp() *cli.App {
	return &cli.App{
		Name:  "duck",
		Usage: "A powerful monorepo management tool",
		Description: "Duck is a build tool and dependency resolver for Go monorepos. " +
			"It scans your project structure and runs scripts across multiple applications " +
			"while respecting dependencies.",
		Version: "1.0.0",
		Commands: []*cli.Command{
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Usage:   "List all discovered projects",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "namespace",
						Aliases: []string{"ns"},
						Usage:   "Filter projects by namespace",
					},
					&cli.StringSliceFlag{
						Name:    "tag",
						Aliases: []string{"t"},
						Usage:   "Filter projects by tag (can be used multiple times)",
					},
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Show detailed project information",
					},
				},
				Action: ListProjects,
			},
			{
				Name:    "run",
				Aliases: []string{"r"},
				Usage:   "Run a script on projects",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "script",
						Aliases:  []string{"s"},
						Usage:    "Script name to run (required)",
						Required: true,
					},
					&cli.StringSliceFlag{
						Name:    "project",
						Aliases: []string{"p"},
						Usage:   "Run on specific projects (namespace/name format)",
					},
					&cli.StringFlag{
						Name:    "namespace",
						Aliases: []string{"ns"},
						Usage:   "Run on all projects in namespace",
					},
					&cli.StringSliceFlag{
						Name:    "tag",
						Aliases: []string{"t"},
						Usage:   "Run on projects with specific tags",
					},
					&cli.BoolFlag{
						Name:    "all",
						Aliases: []string{"a"},
						Usage:   "Run on all projects (respects dependency order)",
					},
					&cli.BoolFlag{
						Name:    "dry-run",
						Aliases: []string{"n"},
						Usage:   "Show what would be executed without running",
					},
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Show detailed execution output",
					},
					&cli.BoolFlag{
						Name:  "parallel",
						Usage: "Run on independent projects in parallel",
					},
				},
				Action: RunScript,
			},
			{
				Name:    "scripts",
				Aliases: []string{"sc"},
				Usage:   "List available scripts",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Show detailed script information",
					},
				},
				Action: ListScripts,
			},
			{
				Name:  "config",
				Usage: "Manage Duck configuration",
				Subcommands: []*cli.Command{
					{
						Name:  "format",
						Usage: "Get or set the project configuration format",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "set",
								Aliases: []string{"s"},
								Usage:   "Set format to 'duck' or 'nx'",
							},
						},
						Action: ConfigFormat,
					},
				},
			},
			{
				Name:    "deps",
				Aliases: []string{"dependencies"},
				Usage:   "Analyze project dependencies",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "workspace",
						Aliases: []string{"w"},
						Usage:   "Workspace root directory",
						Value:   ".",
					},
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Show detailed dependency information including import paths",
					},
					&cli.BoolFlag{
						Name:  "show-indirect",
						Usage: "Show indirect dependencies",
					},
					&cli.BoolFlag{
						Name:  "sync",
						Usage: "Sync discovered dependencies to app.yaml/project.json files",
					},
				},
				Action: AnalyzeDependencies,
			},
		},
	}
}
