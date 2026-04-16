package handlers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/meschbach/mcp-vikunja/pkg/vikunja"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// getTaskHandler handles the get_task tool
func (h *Handlers) getTaskHandler(ctx context.Context, _ *mcp.CallToolRequest, input GetTaskInput) (*mcp.CallToolResult, GetTaskOutput, error) {
	client, err := createVikunjaClient()
	if err != nil {
		return nil, GetTaskOutput{}, err
	}

	taskID, err := parseID("task_id", input.TaskID)
	if err != nil {
		return h.buildErrorResult(err.Error()), GetTaskOutput{}, err
	}

	task, err := client.GetTask(ctx, taskID)
	if err != nil {
		return nil, GetTaskOutput{}, fmt.Errorf("failed to get task: %w", err)
	}

	var bucketInfo *vikunja.TaskBucketInfo
	if input.IncludeBuckets {
		bucketInfo, err = h.buildTaskBucketInfo(ctx, client, task)
		if err != nil {
			h.deps.Logger.Warn("failed to get bucket info for task",
				slog.Int64("task_id", taskID),
				slog.Any("error", err))
		}
	}

	return h.formatGetTaskOutput(task, bucketInfo)
}

func (h *Handlers) buildTaskBucketInfo(ctx context.Context, client *vikunja.Client, task *vikunja.Task) (*vikunja.TaskBucketInfo, error) {
	views, err := client.GetProjectViews(ctx, task.ProjectID)
	if err != nil {
		return nil, err
	}

	taskViews := make([]vikunja.TaskViewInfo, 0, len(views))
	for _, view := range views {
		viewInfo := h.buildViewInfoForTask(view, task)
		taskViews = append(taskViews, viewInfo)
	}

	return &vikunja.TaskBucketInfo{
		TaskID: task.ID,
		Views:  taskViews,
	}, nil
}

func (h *Handlers) buildViewInfoForTask(view *vikunja.ProjectView, task *vikunja.Task) vikunja.TaskViewInfo {
	viewInfo := vikunja.TaskViewInfo{
		ViewID:    view.ID,
		ViewTitle: view.Title,
		ViewKind:  view.ViewKind,
	}

	for _, bucket := range task.Buckets {
		if bucket.ProjectViewID != view.ID {
			continue
		}
		bID := bucket.ID
		bTitle := bucket.Title
		viewInfo.BucketID = &bID
		viewInfo.BucketTitle = &bTitle
		viewInfo.Position = bucket.Position
		if view.DoneBucketID == bucket.ID {
			viewInfo.IsDoneBucket = true
		}
		break
	}
	return viewInfo
}

func (h *Handlers) formatGetTaskOutput(task *vikunja.Task, bucketInfo *vikunja.TaskBucketInfo) (*mcp.CallToolResult, GetTaskOutput, error) {
	output := GetTaskOutput{
		Task: toTask(task),
	}
	if bucketInfo != nil {
		output.Buckets = bucketInfo
	}

	vikunjaOutput := vikunja.TaskOutput{
		Task: vikunja.Task{
			ID:          output.Task.ID,
			Title:       output.Task.Title,
			Description: output.Task.Description,
			ProjectID:   output.Task.ProjectID,
			Done:        output.Task.Done,
			DueDate:     output.Task.DueDate,
			Created:     output.Task.Created,
			Updated:     output.Task.Updated,
			Buckets:     toVikunjaBuckets(output.Task.Buckets),
			Position:    output.Task.Position,
		},
		Buckets: output.Buckets,
	}

	data, err := h.deps.OutputFormatter.Format(vikunjaOutput)
	if err != nil {
		return nil, GetTaskOutput{}, fmt.Errorf("failed to format response: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(data)},
		},
	}, output, nil
}

// listBucketsHandler handles the list_buckets tool
func (h *Handlers) listBucketsHandler(ctx context.Context, _ *mcp.CallToolRequest, input ListBucketsInput) (*mcp.CallToolResult, ListBucketsOutput, error) {
	client, err := createVikunjaClient()
	if err != nil {
		return nil, ListBucketsOutput{}, err
	}

	project, view, buckets, err := h.resolveBucketParams(ctx, client, input)
	if err != nil {
		return h.buildErrorResult(err.Error()), ListBucketsOutput{}, err
	}

	return h.formatBucketsOutput(buckets, project, view)
}

func (h *Handlers) resolveBucketParams(ctx context.Context, client *vikunja.Client, input ListBucketsInput) (project *Project, v *vikunja.ProjectView, buckets []*vikunja.Bucket, err error) {
	projectTitle := coalesceString(input.ProjectTitle, "Inbox")
	viewTitle := coalesceString(input.ViewTitle, "Kanban")

	project, err = findProjectByIDOrTitle(ctx, client, "", projectTitle)
	if err != nil {
		return nil, nil, nil, err
	}

	views, err := client.GetProjectViews(ctx, project.ID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get project views: %w", err)
	}

	v, err = findViewByName(views, viewTitle, false, project.Title)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%v in project %q", err, project.Title)
	}

	buckets, err = client.GetViewBuckets(ctx, project.ID, v.ID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get buckets: %w", err)
	}

	return project, v, buckets, nil
}

func (h *Handlers) formatBucketsOutput(buckets []*vikunja.Bucket, _ *Project, _ *vikunja.ProjectView) (*mcp.CallToolResult, ListBucketsOutput, error) {
	data, err := h.deps.OutputFormatter.Format(buckets)
	if err != nil {
		return nil, ListBucketsOutput{}, fmt.Errorf("failed to format response: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(data)},
		},
	}, ListBucketsOutput{Buckets: toBuckets(buckets)}, nil
}

func coalesceString(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
