package handlers

import (
	"context"
	"fmt"

	"github.com/meschbach/mcp-vikunja/pkg/vikunja"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// moveTaskToBucketHandler handles the move_task_to_bucket tool
func (h *Handlers) moveTaskToBucketHandler(ctx context.Context, _ *mcp.CallToolRequest, input MoveTaskToBucketInput) (*mcp.CallToolResult, MoveTaskToBucketOutput, error) {
	if h.isReadonly() {
		return h.buildErrorResult("Operation not available in readonly mode"), MoveTaskToBucketOutput{}, fmt.Errorf("operation not available in readonly mode")
	}

	client, err := createVikunjaClient()
	if err != nil {
		return nil, MoveTaskToBucketOutput{}, fmt.Errorf("failed to create client: %w", err)
	}

	taskID, projectID, viewID, bucketID, err := h.parseMoveTaskIDs(input)
	if err != nil {
		return h.buildErrorResult(err.Error()), MoveTaskToBucketOutput{}, err
	}

	if err := h.verifyTaskExists(ctx, client, taskID, projectID); err != nil {
		return h.buildErrorResult(err.Error()), MoveTaskToBucketOutput{}, err
	}

	taskBucket, err := h.moveTask(ctx, client, projectID, viewID, bucketID, taskID)
	if err != nil {
		return h.buildErrorResult(fmt.Sprintf("Failed to move task: %v", err)), MoveTaskToBucketOutput{}, fmt.Errorf("failed to move task: %w", err)
	}

	return h.formatMoveTaskOutput(taskBucket, taskID, bucketID)
}

func (h *Handlers) parseMoveTaskIDs(input MoveTaskToBucketInput) (taskID, projectID, viewID, bucketID int64, err error) {
	taskID, err = parseID("task_id", input.TaskID)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	projectID, err = parseID("project_id", input.ProjectID)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	viewID, err = parseID("view_id", input.ViewID)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	bucketID, err = parseID("bucket_id", input.BucketID)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	return taskID, projectID, viewID, bucketID, nil
}

func (h *Handlers) verifyTaskExists(ctx context.Context, client *vikunja.Client, taskID, projectID int64) error {
	task, err := client.GetTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("task with ID %d not found: %w", taskID, err)
	}

	if task.ProjectID != projectID {
		return fmt.Errorf("task %d does not belong to project %d", taskID, projectID)
	}

	return nil
}

func (h *Handlers) moveTask(ctx context.Context, client *vikunja.Client, projectID, viewID, bucketID, taskID int64) (*vikunja.TaskBucket, error) {
	return client.MoveTaskToBucket(ctx, projectID, viewID, bucketID, taskID)
}

func (h *Handlers) formatMoveTaskOutput(taskBucket *vikunja.TaskBucket, taskID, bucketID int64) (*mcp.CallToolResult, MoveTaskToBucketOutput, error) {
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
