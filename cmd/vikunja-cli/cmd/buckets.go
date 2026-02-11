package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/meschbach/mcp-vikunja/pkg/vikunja"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(bucketsCmd)
	bucketsCmd.AddCommand(bucketsListCmd)
	bucketsCmd.AddCommand(bucketsTasksCmd)
}

var bucketsCmd = &cobra.Command{
	Use:   "buckets",
	Short: "Manage buckets",
	Long:  `List and retrieve Vikunja buckets for project views.`,
}

var bucketsListCmd = &cobra.Command{
	Use:   "list [projectID] [viewID]",
	Short: "List all buckets in a project view",
	Long:  "List all buckets for a specific project view. If viewID is not specified, all views for the project are listed.",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid project ID: %s (must be a number)", args[0])
		}

		ctx := context.Background()
		formatter := vikunja.NewFormatter(!noColor, outputWriter)

		if len(args) == 1 {
			logger.Debug("listing views", "projectID", projectID)
			views, err := client.GetProjectViews(ctx, projectID)
			if err != nil {
				return fmt.Errorf("failed to list project views: %w", err)
			}

			if jsonFmt {
				return formatter.FormatAsJSON(views)
			}

			if markdown {
				markdownOutput := formatter.FormatViewAsMarkdown(views[0]) // Format first view
				for i := 1; i < len(views); i++ {
					markdownOutput += "\n---\n\n" + formatter.FormatViewAsMarkdown(views[i])
				}
				_, _ = fmt.Fprintf(outputWriter, "%s", markdownOutput)
				return nil
			}

			return formatter.FormatProjectViews(views)
		}

		viewID, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid view ID: %s (must be a number)", args[1])
		}

		logger.Debug("listing buckets", "projectID", projectID, "viewID", viewID)
		buckets, err := client.GetViewBuckets(ctx, projectID, viewID)
		if err != nil {
			return fmt.Errorf("failed to list buckets: %w", err)
		}

		if jsonFmt {
			return formatter.FormatAsJSON(buckets)
		}

		if markdown {
			markdownOutput := formatter.FormatBucketsAsMarkdown(buckets)
			_, _ = fmt.Fprintf(outputWriter, "%s", markdownOutput)
			return nil
		}

		return formatter.FormatBuckets(buckets)
	},
}

var bucketsTasksCmd = &cobra.Command{
	Use:   "tasks [projectID] [viewID]",
	Short: "List all tasks in a project view, organized by bucket",
	Long:  "List all tasks for a specific project view, grouped by the buckets they belong to.",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid project ID: %s (must be a number)", args[0])
		}

		viewID, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid view ID: %s (must be a number)", args[1])
		}

		ctx := context.Background()
		formatter := vikunja.NewFormatter(!noColor, outputWriter)

		logger.Debug("fetching view details", "projectID", projectID, "viewID", viewID)
		views, err := client.GetProjectViews(ctx, projectID)
		if err != nil {
			return fmt.Errorf("failed to fetch project views: %w", err)
		}

		var currentView *vikunja.ProjectView
		for _, v := range views {
			if v.ID == viewID {
				currentView = &v
				break
			}
		}

		if currentView == nil {
			return fmt.Errorf("view %d not found in project %d", viewID, projectID)
		}

		// Fetch tasks for the view. For kanban views this returns buckets with their tasks.
		viewTasksResp, err := client.GetViewTasks(ctx, projectID, viewID)
		if err != nil {
			return fmt.Errorf("failed to fetch tasks: %w", err)
		}

		var buckets []vikunja.Bucket
		if len(viewTasksResp.Buckets) > 0 {
			buckets = viewTasksResp.Buckets
		} else {
			// Non-kanban: create a single pseudo-bucket containing all tasks
			buckets = []vikunja.Bucket{
				{ID: 0, ProjectViewID: viewID, Title: "All Tasks"},
			}
		}

		vt := &vikunja.ViewTasks{
			ViewID:    viewID,
			ViewTitle: currentView.Title,
			Buckets:   make([]vikunja.BucketTasks, 0, len(buckets)),
		}

		// Build bucket tasks based on response shape
		if len(viewTasksResp.Buckets) > 0 {
			for _, b := range buckets {
				vt.Buckets = append(vt.Buckets, vikunja.BucketTasks{
					Bucket: b,
					Tasks:  b.Tasks,
				})
			}
		} else {
			// Single pseudo-bucket with all tasks
			allTasks := viewTasksResp.Tasks
			vt.Buckets = append(vt.Buckets, vikunja.BucketTasks{
				Bucket: buckets[0],
				Tasks:  allTasks,
			})
		}

		if jsonFmt {
			return formatter.FormatAsJSON(vt)
		}

		if markdown {
			markdownOutput := formatter.FormatViewTasksAsMarkdown(vt)
			_, _ = fmt.Fprintf(outputWriter, "%s", markdownOutput)
			return nil
		}

		return formatter.FormatViewTasks(vt)
	},
}
