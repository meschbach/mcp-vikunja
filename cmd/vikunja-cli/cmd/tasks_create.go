package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/meschbach/mcp-vikunja/pkg/resolution"
	"github.com/meschbach/mcp-vikunja/pkg/vikunja"
	"github.com/spf13/cobra"
)

var (
	tasksCreateProjectFlag string
	tasksCreateBucketFlag  string
)

func init() {
	tasksCmd.AddCommand(tasksCreateCmd)
	tasksCreateCmd.Flags().StringVarP(&tasksCreateProjectFlag, "project", "p", "", "Project ID or title (default: Inbox)")
	tasksCreateCmd.Flags().StringVarP(&tasksCreateBucketFlag, "bucket", "b", "", "Bucket ID or title")
}

var tasksCreateCmd = &cobra.Command{
	Use:   "create <title> [description]",
	Short: "Create a new task",
	Long:  `Create a new Vikunja task with optional description and project/bucket assignment.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		title := args[0]
		description := ""
		if len(args) > 1 {
			description = args[1]
		}

		// Resolve project (default to "Inbox")
		projectTitle := tasksCreateProjectFlag
		if projectTitle == "" {
			projectTitle = "Inbox"
		}

		project, err := resolution.ResolveProject(ctx, client, projectTitle)
		if err != nil {
			return fmt.Errorf("failed to resolve project: %w", err)
		}

		// Resolve bucket if specified
		var bucketID *int64
		if tasksCreateBucketFlag != "" {
			bucket, err := resolution.FindBucketByIDOrTitle(ctx, client, project.ID, tasksCreateBucketFlag)
			if err != nil {
				return fmt.Errorf("failed to resolve bucket: %w", err)
			}
			bucketID = bucket
		}

		// Create task
		logger.Debug("creating task", "title", title, "project_id", project.ID, "bucket_id", bucketID)
		task, err := client.CreateTask(ctx, title, project.ID, description, bucketID, time.Time{})
		if err != nil {
			return fmt.Errorf("failed to create task: %w", err)
		}

		// Always fetch bucket info for output
		bucketInfo, err := client.GetTaskBuckets(ctx, task.ID)
		if err != nil {
			logger.Debug("failed to get bucket info", "task_id", task.ID, "error", err)
		}

		// Format output
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

		if markdown {
			markdownOutput := formatter.FormatTaskAsMarkdown(*task)
			_, _ = fmt.Fprintf(outputWriter, "%s", markdownOutput)
			return nil
		}

		return formatter.FormatTaskWithBuckets(task, bucketInfo)
	},
}
