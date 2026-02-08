package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/fatih/color"
	"github.com/meschbach/mcp-vikunja/pkg/vikunja"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(projectsCmd)
	projectsCmd.AddCommand(projectsListCmd)
	projectsCmd.AddCommand(projectsGetCmd)
}

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Manage projects",
	Long:  `List and retrieve Vikunja projects.`,
}

var projectsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	Long:  `List all projects you have access to.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		logger.Debug("listing projects")
		projects, err := client.GetProjects(ctx)
		if err != nil {
			return fmt.Errorf("failed to list projects: %w", err)
		}

		formatter := vikunja.NewFormatter(!noColor, outputWriter)

		if jsonFmt {
			return formatter.FormatProjectsAsJSON(projects)
		}

		return formatter.FormatProjects(projects)
	},
}

var projectsGetCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Get a project by ID",
	Long:  `Retrieve detailed information about a specific project.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid project ID: %s (must be a number)", args[0])
		}

		ctx := context.Background()

		logger.Debug("getting project", "id", id)
		project, err := client.GetProject(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to get project %d: %w", id, err)
		}

		if project == nil {
			if !noColor {
				color.Red("Project not found: %d\n", id)
			} else {
				fmt.Fprintf(outputWriter, "Project not found: %d\n", id)
			}
			return nil
		}

		formatter := vikunja.NewFormatter(!noColor, outputWriter)

		if jsonFmt {
			return formatter.FormatProjectAsJSON(project)
		}

		return formatter.FormatProject(project)
	},
}
