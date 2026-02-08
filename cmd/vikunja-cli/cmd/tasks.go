package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/fatih/color"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

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

		return formatter.FormatTasks(tasks)
	},
}

var tasksGetCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Get a task by ID",
	Long:  `Retrieve detailed information about a specific task.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid task ID: %s (must be a number)", args[0])
		}

		ctx := context.Background()

		logger.Debug("getting task", "id", id)
		task, err := client.GetTask(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to get task %d: %w", id, err)
		}

		if task == nil {
			if !noColor {
				color.Red("Task not found: %d\n", id)
			} else {
				fmt.Fprintf(outputWriter, "Task not found: %d\n", id)
			}
			return nil
		}

		// Get bucket information by default for CLI
		bucketInfo, err := client.GetTaskBuckets(ctx, id)
		if err != nil {
			logger.Debug("failed to get bucket info", "id", id, "error", err)
		}

		formatter := vikunja.NewFormatter(!noColor, outputWriter)

		if jsonFmt {
			if bucketInfo != nil {
				taskWithBuckets := struct {
					Task    vikunja.Task            `json:"task"`
					Buckets *vikunja.TaskBucketInfo `json:"buckets,omitempty"`
				}{
					Task:    *task,
					Buckets: bucketInfo,
				}
				return formatter.FormatAsJSON(taskWithBuckets)
			}
			return formatter.FormatTaskAsJSON(task)
		}

		return formatter.FormatTaskWithBuckets(task, bucketInfo)
	},
}
