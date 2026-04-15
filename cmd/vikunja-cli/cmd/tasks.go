package cmd

import (
	"context"
	"fmt"

	"github.com/meschbach/mcp-vikunja/pkg/vikunja"
	"github.com/spf13/cobra"
)

var (
	projectID int64
)

func init() {
	rootCmd.AddCommand(tasksCmd)
	tasksCmd.AddCommand(tasksListCmd)
	tasksCmd.AddCommand(tasksGetCmd)

	tasksListCmd.Flags().Int64VarP(&projectID, "project", "p", 0, "Filter tasks by project ID")
}

var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "Manage tasks",
	Long:  `List and retrieve Vikunja tasks.`,
}

var tasksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	Long:  `List all tasks, optionally filtered by project.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return listTasks(cmd.Context())
	},
}

var tasksGetCmd = newGetCmd("task", func(ctx context.Context, id int64) (interface{}, error) {
	return client.GetTask(ctx, id)
}, false)

func listTasks(ctx context.Context) error {
	if projectID > 0 {
		logger.Debug("listing tasks for project", "project_id", projectID)
	} else {
		logger.Debug("listing all tasks")
	}

	tasks, err := client.GetTasks(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	formatter := vikunja.NewFormatter(!noColor, outputWriter)

	if jsonFmt {
		return formatter.FormatTasksAsJSON(tasks)
	}

	if markdown {
		return writeAll(outputWriter, formatter.FormatTasksAsMarkdown(tasks))
	}

	return formatter.FormatTasks(tasks)
}
