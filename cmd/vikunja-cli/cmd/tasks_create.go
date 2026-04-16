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
		return createTask(cmd.Context(), args)
	},
}

func createTask(ctx context.Context, args []string) error {
	title := args[0]
	description := getDescription(args)

	project, err := resolveProject(ctx)
	if err != nil {
		return err
	}

	bucketID, err := resolveBucket(ctx, project.ID)
	if err != nil {
		return err
	}

	task, err := executeCreateTask(ctx, title, project.ID, description, bucketID)
	if err != nil {
		return err
	}

	return formatTaskOutput(task)
}

func getDescription(args []string) string {
	if len(args) > 1 {
		return args[1]
	}
	return ""
}

func resolveProject(ctx context.Context) (*resolution.Project, error) {
	projectTitle := tasksCreateProjectFlag
	if projectTitle == "" {
		projectTitle = "Inbox"
	}
	project, err := resolution.ResolveProject(ctx, client, projectTitle)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve project: %w", err)
	}
	return project, nil
}

func resolveBucket(ctx context.Context, projectID int64) (*int64, error) {
	if tasksCreateBucketFlag == "" {
		return nil, nil
	}
	bucket, err := resolution.FindBucketByIDOrTitle(ctx, client, projectID, tasksCreateBucketFlag)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve bucket: %w", err)
	}
	return bucket, nil
}

func executeCreateTask(ctx context.Context, title string, projectID int64, description string, bucketID *int64) (*vikunja.Task, error) {
	logger.Debug("creating task", "title", title, "project_id", projectID, "bucket_id", bucketID)
	task, err := client.CreateTask(ctx, title, projectID, description, bucketID, time.Time{})
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}
	return task, nil
}

func formatTaskOutput(task *vikunja.Task) error {
	formatter := vikunja.NewFormatter(!noColor, outputWriter)
	if jsonFmt {
		return formatter.FormatTaskAsJSON(task)
	}
	if markdown {
		return writeAll(outputWriter, formatter.FormatTaskAsMarkdown(task))
	}
	return formatter.FormatTaskAsJSON(task)
}
