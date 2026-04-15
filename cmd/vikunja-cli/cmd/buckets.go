// Package cmd provides CLI commands for the vikunja-cli tool.
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
		return runBucketsList(cmd.Context(), args)
	},
}

var bucketsTasksCmd = &cobra.Command{
	Use:   "tasks [projectID] [viewID]",
	Short: "List all tasks in a project view, organized by bucket",
	Long:  "List all tasks for a specific project view, grouped by the buckets they belong to.",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBucketsTasks(cmd.Context(), args)
	},
}

func runBucketsList(ctx context.Context, args []string) error {
	projectID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid project ID: %s (must be a number)", args[0])
	}

	formatter := vikunja.NewFormatter(!noColor, outputWriter)

	if len(args) == 1 {
		return listViews(ctx, projectID, formatter)
	}
	return listViewBuckets(ctx, projectID, args[1], formatter)
}

func listViews(ctx context.Context, projectID int64, formatter *vikunja.Formatter) error {
	logger.Debug("listing views", "projectID", projectID)
	views, err := client.GetProjectViews(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to list project views: %w", err)
	}

	if jsonFmt {
		return formatter.FormatAsJSON(views)
	}
	if markdown {
		return formatViewsMarkdown(formatter, views)
	}
	return formatter.FormatProjectViews(views)
}

func listViewBuckets(ctx context.Context, projectID int64, viewArg string, formatter *vikunja.Formatter) error {
	viewID, err := strconv.ParseInt(viewArg, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid view ID: %s (must be a number)", viewArg)
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
		return writeOutput(formatter.FormatBucketsAsMarkdown(buckets))
	}
	return formatter.FormatBuckets(buckets)
}

func runBucketsTasks(ctx context.Context, args []string) error {
	projectID, viewID, err := parseProjectAndViewIDs(args)
	if err != nil {
		return err
	}

	formatter := vikunja.NewFormatter(!noColor, outputWriter)

	logger.Debug("fetching view details", "projectID", projectID, "viewID", viewID)
	views, err := client.GetProjectViews(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to fetch project views: %w", err)
	}

	currentView := findViewByID(views, viewID)
	if currentView == nil {
		return fmt.Errorf("view %d not found in project %d", viewID, projectID)
	}

	vt, err := buildViewTasks(ctx, projectID, viewID, currentView)
	if err != nil {
		return err
	}

	return formatViewTasks(formatter, vt)
}

func parseProjectAndViewIDs(args []string) (projectID, viewID int64, err error) {
	projectID, err = strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid project ID: %s (must be a number)", args[0])
	}

	viewID, err = strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid view ID: %s (must be a number)", args[1])
	}
	return projectID, viewID, nil
}

func findViewByID(views []*vikunja.ProjectView, viewID int64) *vikunja.ProjectView {
	for _, v := range views {
		if v.ID == viewID {
			return v
		}
	}
	return nil
}

func buildViewTasks(ctx context.Context, projectID, viewID int64, currentView *vikunja.ProjectView) (*vikunja.ViewTasks, error) {
	buckets, err := client.GetViewBuckets(ctx, projectID, viewID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch buckets: %w", err)
	}

	vt := &vikunja.ViewTasks{
		ViewID:    viewID,
		ViewTitle: currentView.Title,
		Buckets:   make([]vikunja.BucketTasks, 0, len(buckets)),
	}

	for _, b := range buckets {
		vt.Buckets = append(vt.Buckets, vikunja.BucketTasks{
			Bucket: *b,
			Tasks:  b.Tasks,
		})
	}
	return vt, nil
}

func formatViewTasks(formatter *vikunja.Formatter, vt *vikunja.ViewTasks) error {
	if jsonFmt {
		return formatter.FormatAsJSON(vt)
	}
	if markdown {
		return writeOutput(formatter.FormatViewTasksAsMarkdown(vt))
	}
	return formatter.FormatViewTasks(vt)
}

func formatViewsMarkdown(formatter *vikunja.Formatter, views []*vikunja.ProjectView) error {
	if len(views) == 0 {
		return nil
	}
	markdownOutput := formatter.FormatViewAsMarkdown(views[0])
	for i := 1; i < len(views); i++ {
		markdownOutput += "\n---\n\n" + formatter.FormatViewAsMarkdown(views[i])
	}
	return writeOutput(markdownOutput)
}

func writeOutput(s string) error {
	_, err := outputWriter.Write([]byte(s))
	return err
}
