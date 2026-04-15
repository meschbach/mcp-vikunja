package cmd

import (
	"context"
	"fmt"

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
	RunE: func(cmd *cobra.Command, _ []string) error {
		return listProjects(cmd.Context())
	},
}

var projectsGetCmd = newGetCmd("project", func(ctx context.Context, id int64) (interface{}, error) {
	return client.GetProject(ctx, id)
}, true)

func listProjects(ctx context.Context) error {
	logger.Debug("listing projects")
	projects, err := client.GetProjects(ctx)
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	formatter := vikunja.NewFormatter(!noColor, outputWriter)

	if jsonFmt {
		return formatter.FormatProjectsAsJSON(projects)
	}

	if markdown {
		return writeAll(outputWriter, formatter.FormatProjectsAsMarkdown(projects))
	}

	return formatter.FormatProjects(projects)
}
