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

	project, targetProjectID, err := h.resolveProjectByValue(ctx, client, input.Project)
	if err != nil {
		return h.buildErrorResult(err.Error()), ListTasksOutput{}, err
	}

	targetViewID, targetViewTitle, err := h.resolveViewByValue(ctx, client, targetProjectID, input.View)
	if err != nil {
		return h.buildErrorResult(err.Error()), ListTasksOutput{}, err
	}

	targetBucketID, targetBucketTitle, err := h.resolveBucketByValue(ctx, client, targetProjectID, targetViewID, input.Bucket)
	if err != nil {
		return h.buildErrorResult(err.Error()), ListTasksOutput{}, err
	}

	viewTasksResp, err := h.getViewTasks(ctx, client, targetProjectID, targetViewID, targetBucketID, targetBucketTitle, targetViewTitle)
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

// resolveProjectByValue resolves project from ID (integer string) or title
func (h *Handlers) resolveProjectByValue(ctx context.Context, client *vikunja.Client, value string) (*Project, int64, error) {
	if value == "" {
		return h.findProjectByTitle(ctx, client, "Inbox")
	}

	if id, err := strconv.ParseInt(value, 10, 64); err == nil && id > 0 {
		project, err := client.GetProject(ctx, id)
		if err != nil {
			return nil, 0, fmt.Errorf("project with ID %d not found: %w", id, err)
		}
		return &Project{
			ID:    project.ID,
			Title: project.Title,
			URI:   fmt.Sprintf("vikunja://project/%d", project.ID),
		}, id, nil
	}

	return h.findProjectByTitle(ctx, client, value)
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

// resolveViewByValue resolves view from ID (integer string) or title
func (h *Handlers) resolveViewByValue(ctx context.Context, client *vikunja.Client, projectID int64, value string) (int64, string, error) {
	views, err := client.GetProjectViews(ctx, projectID)
	if err != nil {
		return 0, "", fmt.Errorf("failed to get project views: %w", err)
	}

	if value == "" {
		return h.resolveViewByTitle("Kanban", views, projectID)
	}

	if id, err := strconv.ParseInt(value, 10, 64); err == nil && id > 0 {
		for _, v := range views {
			if v.ID == id {
				return id, v.Title, nil
			}
		}
		return 0, "", fmt.Errorf("view with ID %d not found in project %d", id, projectID)
	}

	return h.resolveViewByTitle(value, views, projectID)
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

// resolveBucketByValue resolves bucket from ID (integer string) or title
func (h *Handlers) resolveBucketByValue(ctx context.Context, client *vikunja.Client, projectID, viewID int64, value string) (int64, string, error) {
	if value == "" {
		return 0, "", nil
	}

	buckets, err := client.GetViewBuckets(ctx, projectID, viewID)
	if err != nil {
		return 0, "", fmt.Errorf("failed to get view buckets: %w", err)
	}

	if id, err := strconv.ParseInt(value, 10, 64); err == nil && id > 0 {
		for _, b := range buckets {
			if b.ID == id {
				return id, b.Title, nil
			}
		}
		return 0, "", fmt.Errorf("bucket with ID %d not found in view %d", id, viewID)
	}

	for _, b := range buckets {
		if b.Title == value {
			return b.ID, b.Title, nil
		}
	}
	return 0, "", fmt.Errorf("bucket with title %q not found in view %d", value, viewID)
}

// getViewTasks gets view tasks with optional bucket filtering
func (h *Handlers) getViewTasks(ctx context.Context, client *vikunja.Client, targetProjectID, targetViewID int64, targetBucketID int64, targetBucketTitle, targetViewTitle string) (*vikunja.ViewTasksResponse, error) {
	viewTasksResp, err := client.GetViewTasks(ctx, targetProjectID, targetViewID)
	if err != nil {
		return nil, fmt.Errorf("failed to get view tasks: %w", err)
	}

	if targetBucketID != 0 || targetBucketTitle != "" {
		return h.filterViewTasksByBucket(viewTasksResp, targetBucketID, targetBucketTitle, targetViewTitle)
	}

	return viewTasksResp, nil
}

// filterViewTasksByBucket filters view tasks by bucket
func (h *Handlers) filterViewTasksByBucket(viewTasksResp *vikunja.ViewTasksResponse, targetBucketID int64, targetBucketTitle, targetViewTitle string) (*vikunja.ViewTasksResponse, error) {
	if len(viewTasksResp.Buckets) == 0 && len(viewTasksResp.Tasks) > 0 {
		return nil, fmt.Errorf("bucket filtering not supported for non-kanban views")
	}

	foundBucket, err := h.findBucket(viewTasksResp.Buckets, targetBucketID, targetBucketTitle, targetViewTitle)
	if err != nil {
		return nil, err
	}

	viewTasksResp.Buckets = []vikunja.Bucket{*foundBucket}
	return viewTasksResp, nil
}

// findBucket finds a bucket by ID or title
func (h *Handlers) findBucket(buckets []vikunja.Bucket, targetBucketID int64, targetBucketTitle, targetViewTitle string) (*vikunja.Bucket, error) {
	if targetBucketID != 0 {
		for i := range buckets {
			if buckets[i].ID == targetBucketID {
				return &buckets[i], nil
			}
		}
		return nil, fmt.Errorf("bucket with ID %d not found in view %q", targetBucketID, targetViewTitle)
	}

	if targetBucketTitle != "" {
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
