// Package handlers provides MCP tool handlers for Vikunja integration.
package handlers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// moveTaskToBucketHandler handles the move_task_to_bucket tool
func (h *Handlers) moveTaskToBucketHandler(ctx context.Context, _ *mcp.CallToolRequest, input MoveTaskToBucketInput) (*mcp.CallToolResult, MoveTaskToBucketOutput, error) {
	// Check readonly mode
	if h.isReadonly() {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Operation not available in readonly mode"},
			},
		}, MoveTaskToBucketOutput{}, fmt.Errorf("operation not available in readonly mode")
	}

	client, err := createVikunjaClient()
	if err != nil {
		return nil, MoveTaskToBucketOutput{}, fmt.Errorf("failed to create client: %w", err)
	}

	// Parse and validate task ID
	taskID, err := strconv.ParseInt(input.TaskID, 10, 64)
	if err != nil {
		return h.buildErrorResult(fmt.Sprintf("Invalid task_id: %s", input.TaskID)), MoveTaskToBucketOutput{}, fmt.Errorf("invalid task_id: %s", input.TaskID)
	}

	// Parse and validate project ID
	projectID, err := strconv.ParseInt(input.ProjectID, 10, 64)
	if err != nil {
		return h.buildErrorResult(fmt.Sprintf("Invalid project_id: %s", input.ProjectID)), MoveTaskToBucketOutput{}, fmt.Errorf("invalid project_id: %s", input.ProjectID)
	}

	// Parse and validate view ID
	viewID, err := strconv.ParseInt(input.ViewID, 10, 64)
	if err != nil {
		return h.buildErrorResult(fmt.Sprintf("Invalid view_id: %s", input.ViewID)), MoveTaskToBucketOutput{}, fmt.Errorf("invalid view_id: %s", input.ViewID)
	}

	// Parse and validate bucket ID
	bucketID, err := strconv.ParseInt(input.BucketID, 10, 64)
	if err != nil {
		return h.buildErrorResult(fmt.Sprintf("Invalid bucket_id: %s", input.BucketID)), MoveTaskToBucketOutput{}, fmt.Errorf("invalid bucket_id: %s", input.BucketID)
	}

	// Basic validation: verify task exists
	task, err := client.GetTask(ctx, taskID)
	if err != nil {
		return h.buildErrorResult(fmt.Sprintf("Task with ID %s not found: %v", input.TaskID, err)), MoveTaskToBucketOutput{}, fmt.Errorf("task not found: %w", err)
	}

	// Verify task belongs to specified project
	if task.ProjectID != projectID {
		return h.buildErrorResult(fmt.Sprintf("Task %d does not belong to project %d", taskID, projectID)), MoveTaskToBucketOutput{}, fmt.Errorf("task does not belong to specified project")
	}

	// Move task to bucket
	taskBucket, err := client.MoveTaskToBucket(ctx, projectID, viewID, bucketID, taskID)
	if err != nil {
		return h.buildErrorResult(fmt.Sprintf("Failed to move task: %v", err)), MoveTaskToBucketOutput{}, fmt.Errorf("failed to move task: %w", err)
	}

	output := MoveTaskToBucketOutput{
		TaskBucket: *taskBucket,
		Message:    fmt.Sprintf("Task %d successfully moved to bucket %d", taskID, bucketID),
	}

	data, err := h.deps.OutputFormatter.Format(output)
	if err != nil {
		return nil, MoveTaskToBucketOutput{}, fmt.Errorf("failed to format response: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(data)},
		},
	}, output, nil
}
