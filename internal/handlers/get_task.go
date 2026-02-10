// Package handlers provides MCP tool handlers for Vikunja integration.
package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/meschbach/mcp-vikunja/pkg/vikunja"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// getTaskHandler handles the get_task tool
func (h *Handlers) getTaskHandler(ctx context.Context, _ *mcp.CallToolRequest, input GetTaskInput) (*mcp.CallToolResult, GetTaskOutput, error) {
	client, err := createVikunjaClient()
	if err != nil {
		return nil, GetTaskOutput{}, err
	}

	taskID, err := strconv.ParseInt(input.TaskID, 10, 64)
	if err != nil {
		return nil, GetTaskOutput{}, fmt.Errorf("invalid task_id: %s", input.TaskID)
	}

	task, err := client.GetTask(ctx, taskID) // This task already has buckets expanded
	if err != nil {
		return nil, GetTaskOutput{}, fmt.Errorf("failed to get task: %w", err)
	}

	var bucketInfo *vikunja.TaskBucketInfo
	if input.IncludeBuckets { // Default to true, allow explicit false
		// Reuse logic from GetTaskBuckets but with already fetched task
		views, err := client.GetProjectViews(ctx, task.ProjectID)
		if err != nil {
			h.deps.Logger.Warn("failed to get project views for task",
				slog.Int64("task_id", taskID),
				slog.Any("error", err))
		} else {
			var taskViews []vikunja.TaskViewInfo
			for _, view := range views {
				viewInfo := vikunja.TaskViewInfo{
					ViewID:    view.ID,
					ViewTitle: view.Title,
					ViewKind:  view.ViewKind,
				}

				for _, bucket := range task.Buckets {
					if bucket.ProjectViewID == view.ID {
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
				}
				taskViews = append(taskViews, viewInfo)
			}
			bucketInfo = &vikunja.TaskBucketInfo{
				TaskID: taskID,
				Views:  taskViews,
			}
		}
	}

	output := GetTaskOutput{
		Task: toTask(*task),
	}
	if bucketInfo != nil {
		output.Buckets = bucketInfo
	}

	// Convert handlers.GetTaskOutput to vikunja.TaskOutput for formatting
	vikunjaOutput := vikunja.TaskOutput{
		Task: vikunja.Task{
			ID:          output.Task.ID,
			Title:       output.Task.Title,
			Description: output.Task.Description,
			ProjectID:   output.Task.ProjectID,
			Done:        output.Task.Done,
			DueDate:     parseTime(output.Task.DueDate),
			Created:     parseTime(output.Task.Created),
			Updated:     parseTime(output.Task.Updated),
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

	projectID, err := strconv.ParseInt(input.ProjectID, 10, 64)
	if err != nil {
		return nil, ListBucketsOutput{}, fmt.Errorf("invalid project_id: %s", input.ProjectID)
	}

	viewID, err := strconv.ParseInt(input.ViewID, 10, 64)
	if err != nil {
		return nil, ListBucketsOutput{}, fmt.Errorf("invalid view_id: %s", input.ViewID)
	}

	buckets, err := client.GetViewBuckets(ctx, projectID, viewID)
	if err != nil {
		return nil, ListBucketsOutput{}, fmt.Errorf("failed to get buckets: %w", err)
	}

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
