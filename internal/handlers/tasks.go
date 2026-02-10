// Package handlers provides MCP tool handlers for Vikunja integration.
package handlers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/meschbach/mcp-vikunja/pkg/vikunja"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// listTasksHandler handles the list_tasks tool
func (h *Handlers) listTasksHandler(ctx context.Context, _ *mcp.CallToolRequest, input ListTasksInput) (*mcp.CallToolResult, ListTasksOutput, error) {
	client, err := createVikunjaClient()
	if err != nil {
		return nil, ListTasksOutput{}, err
	}

	// Validate bucket filtering parameters early to fail fast
	if input.BucketID != "" && input.BucketTitle != "" {
		return &mcp.CallToolResult{
			IsError: true,
		}, ListTasksOutput{}, fmt.Errorf("cannot specify both bucket_id and bucket_title")
	}

	targetBucketID, targetBucketTitle, err := h.validateBucketFiltering(input)
	if err != nil {
		return h.buildErrorResult(err.Error()), ListTasksOutput{}, err
	}

	project, targetProjectID, err := h.resolveProject(ctx, client, input)
	if err != nil {
		return h.buildErrorResult(err.Error()), ListTasksOutput{}, err
	}

	targetViewID, targetViewTitle, err := h.resolveView(ctx, client, targetProjectID, input)
	if err != nil {
		return h.buildErrorResult(err.Error()), ListTasksOutput{}, err
	}

	viewTasksResp, err := h.getViewTasks(ctx, client, targetProjectID, targetViewID, input, targetBucketID, targetBucketTitle, targetViewTitle)
	if err != nil {
		return h.buildErrorResult(err.Error()), ListTasksOutput{}, err
	}

	vt := h.buildViewTasksSummary(targetViewID, targetViewTitle, viewTasksResp)

	vikunjaVT := h.convertToVikunjaViewTasksSummary(vt)

	data, err := h.deps.OutputFormatter.Format(vikunjaVT)
	if err != nil {
		return nil, ListTasksOutput{}, fmt.Errorf("failed to format response: %w", err)
	}

	return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(data)},
			},
		}, ListTasksOutput{
			View:    vt,
			Project: project,
		}, nil
}

// validateBucketFiltering validates bucket filtering parameters
func (h *Handlers) validateBucketFiltering(input ListTasksInput) (*int64, string, error) {
	if input.BucketID != "" {
		bucketID, err := strconv.ParseInt(input.BucketID, 10, 64)
		if err != nil {
			return nil, "", fmt.Errorf("invalid bucket_id: %s", input.BucketID)
		}
		return &bucketID, "", nil
	}
	if input.BucketTitle != "" {
		return nil, input.BucketTitle, nil
	}
	return nil, "", nil
}

// resolveProject resolves the project from input parameters
func (h *Handlers) resolveProject(ctx context.Context, client *vikunja.Client, input ListTasksInput) (*Project, int64, error) {
	if input.ProjectID != "" {
		targetProjectID, err := strconv.ParseInt(input.ProjectID, 10, 64)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid project_id: %s", input.ProjectID)
		}
		return nil, targetProjectID, nil
	}

	// Default project title to "Inbox" if not specified
	projectTitle := input.ProjectTitle
	if projectTitle == "" {
		projectTitle = "Inbox"
	}

	return h.findProjectByTitle(ctx, client, projectTitle)
}

// findProjectByTitle finds a project by its title
func (h *Handlers) findProjectByTitle(ctx context.Context, client *vikunja.Client, projectTitle string) (*Project, int64, error) {
	projects, err := client.GetProjects(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list projects: %w", err)
	}

	for _, p := range projects {
		if p.Title == projectTitle {
			project := &Project{
				ID:    p.ID,
				Title: p.Title,
				URI:   fmt.Sprintf("vikunja://project/%d", p.ID),
			}
			return project, p.ID, nil
		}
	}

	return nil, 0, fmt.Errorf("project with title %q not found", projectTitle)
}

// resolveView resolves the view from input parameters
func (h *Handlers) resolveView(ctx context.Context, client *vikunja.Client, targetProjectID int64, input ListTasksInput) (int64, string, error) {
	views, err := client.GetProjectViews(ctx, targetProjectID)
	if err != nil {
		return 0, "", fmt.Errorf("failed to get project views: %w", err)
	}

	if input.ViewID != "" {
		return h.resolveViewByID(input.ViewID, views)
	}

	// Default view title to "Kanban" if not specified
	viewTitle := input.ViewTitle
	if viewTitle == "" {
		viewTitle = "Kanban"
	}

	return h.resolveViewByTitle(viewTitle, views, targetProjectID)
}

// resolveViewByID resolves view by ID
func (h *Handlers) resolveViewByID(viewIDStr string, views []vikunja.ProjectView) (int64, string, error) {
	targetViewID, err := strconv.ParseInt(viewIDStr, 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("invalid view_id: %s", viewIDStr)
	}

	for _, v := range views {
		if v.ID == targetViewID {
			return targetViewID, v.Title, nil
		}
	}

	return 0, "", fmt.Errorf("view with ID %d not found", targetViewID)
}

// resolveViewByTitle resolves view by title
func (h *Handlers) resolveViewByTitle(viewTitle string, views []vikunja.ProjectView, targetProjectID int64) (int64, string, error) {
	for _, v := range views {
		if v.Title == viewTitle {
			return v.ID, v.Title, nil
		}
	}

	return 0, "", fmt.Errorf("view with title %q not found in project %d", viewTitle, targetProjectID)
}

// getViewTasks gets view tasks with optional bucket filtering
func (h *Handlers) getViewTasks(ctx context.Context, client *vikunja.Client, targetProjectID, targetViewID int64, input ListTasksInput, targetBucketID *int64, targetBucketTitle, targetViewTitle string) (*vikunja.ViewTasksResponse, error) {
	viewTasksResp, err := client.GetViewTasks(ctx, targetProjectID, targetViewID)
	if err != nil {
		return nil, fmt.Errorf("failed to get view tasks: %w", err)
	}

	// Handle bucket filtering if specified
	if input.BucketID != "" || input.BucketTitle != "" {
		return h.filterViewTasksByBucket(viewTasksResp, targetBucketID, targetBucketTitle, targetViewTitle)
	}

	return viewTasksResp, nil
}

// filterViewTasksByBucket filters view tasks by bucket
func (h *Handlers) filterViewTasksByBucket(viewTasksResp *vikunja.ViewTasksResponse, targetBucketID *int64, targetBucketTitle, targetViewTitle string) (*vikunja.ViewTasksResponse, error) {
	// Check if this is a kanban view (has buckets)
	if len(viewTasksResp.Buckets) == 0 && len(viewTasksResp.Tasks) > 0 {
		return nil, fmt.Errorf("bucket filtering not supported for non-kanban views")
	}

	foundBucket, err := h.findBucket(viewTasksResp.Buckets, targetBucketID, targetBucketTitle, targetViewTitle)
	if err != nil {
		return nil, err
	}

	// Filter to only include found bucket
	viewTasksResp.Buckets = []vikunja.Bucket{*foundBucket}
	return viewTasksResp, nil
}

// findBucket finds a bucket by ID or title
func (h *Handlers) findBucket(buckets []vikunja.Bucket, targetBucketID *int64, targetBucketTitle, targetViewTitle string) (*vikunja.Bucket, error) {
	if targetBucketID != nil {
		// Search by bucket ID
		for i := range buckets {
			if buckets[i].ID == *targetBucketID {
				return &buckets[i], nil
			}
		}
		return nil, fmt.Errorf("bucket with ID %q not found in view %q", *targetBucketID, targetViewTitle)
	}

	if targetBucketTitle != "" {
		// Search by bucket title
		for i := range buckets {
			if buckets[i].Title == targetBucketTitle {
				return &buckets[i], nil
			}
		}
		return nil, fmt.Errorf("bucket with title %q not found in view %q", targetBucketTitle, targetViewTitle)
	}

	return nil, fmt.Errorf("no bucket filter specified")
}

// buildViewTasksSummary builds the view tasks summary
func (h *Handlers) buildViewTasksSummary(targetViewID int64, targetViewTitle string, viewTasksResp *vikunja.ViewTasksResponse) ViewTasksSummary {
	vt := ViewTasksSummary{
		ViewID:    targetViewID,
		ViewTitle: targetViewTitle,
		Buckets:   make([]BucketTasksSummary, 0),
	}

	if len(viewTasksResp.Buckets) > 0 {
		for _, b := range viewTasksResp.Buckets {
			vikunjaBucket := b // Explicitly use vikunja.Bucket type
			vt.Buckets = append(vt.Buckets, BucketTasksSummary{
				Bucket: toBucketSummary(vikunjaBucket),
				Tasks:  toTasksSummary(vikunjaBucket.Tasks),
			})
		}
	} else {
		vt.Buckets = append(vt.Buckets, BucketTasksSummary{
			Bucket: BucketSummary{ID: 0, Title: "All Tasks"},
			Tasks:  toTasksSummary(viewTasksResp.Tasks),
		})
	}

	return vt
}

// convertToVikunjaViewTasksSummary converts handlers ViewTasksSummary to vikunja.ViewTasksSummary
func (h *Handlers) convertToVikunjaViewTasksSummary(vt ViewTasksSummary) vikunja.ViewTasksSummary {
	vikunjaVT := vikunja.ViewTasksSummary{
		ViewID:    vt.ViewID,
		ViewTitle: vt.ViewTitle,
		Buckets:   make([]vikunja.BucketTasksSummary, len(vt.Buckets)),
	}
	for i, bucket := range vt.Buckets {
		vikunjaVT.Buckets[i] = vikunja.BucketTasksSummary{
			Bucket: vikunja.BucketSummary{
				ID:    bucket.Bucket.ID,
				Title: bucket.Bucket.Title,
			},
			Tasks: make([]vikunja.TaskSummary, len(bucket.Tasks)),
		}
		for j, task := range bucket.Tasks {
			vikunjaVT.Buckets[i].Tasks[j] = vikunja.TaskSummary{
				ID:    task.ID,
				Title: task.Title,
			}
		}
	}
	return vikunjaVT
}

// buildErrorResult builds an error result
func (h *Handlers) buildErrorResult(message string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			&mcp.TextContent{Text: message},
		},
	}
}
