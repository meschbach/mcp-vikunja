package handlers

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/meschbach/mcp-vikunja/pkg/resolution"
	"github.com/meschbach/mcp-vikunja/pkg/vikunja"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// createTaskHandler handles the create_task tool
func (h *Handlers) createTaskHandler(ctx context.Context, _ *mcp.CallToolRequest, input CreateTaskInput) (*mcp.CallToolResult, CreateTaskOutput, error) {
	if h.isReadonly() {
		return h.buildErrorResult("Operation not available in readonly mode"), CreateTaskOutput{}, fmt.Errorf("operation not available in readonly mode")
	}

	if err := validateCreateTaskInput(input); err != nil {
		return h.buildErrorResult(err.Error()), CreateTaskOutput{}, err
	}

	client := h.deps.Client
	if client == nil {
		return h.buildErrorResult("Vikunja client not configured"), CreateTaskOutput{}, fmt.Errorf("client not configured")
	}

	project, err := resolution.ResolveProject(ctx, client, input.ProjectID)
	if err != nil {
		return h.buildErrorResult(err.Error()), CreateTaskOutput{}, err
	}

	bucketID, err := h.resolveBucketForTask(ctx, client, project.ID, input.BucketID)
	if err != nil {
		return h.buildErrorResult(err.Error()), CreateTaskOutput{}, err
	}

	task, err := h.createTask(ctx, client, input, project.ID, bucketID)
	if err != nil {
		return h.buildErrorResult(err.Error()), CreateTaskOutput{}, err
	}

	return h.formatTaskOutput(task)
}

func validateCreateTaskInput(input CreateTaskInput) error {
	if err := validateRequiredString("title", input.Title); err != nil {
		return err
	}
	if err := validateRequiredString("project_id", input.ProjectID); err != nil {
		return err
	}
	if input.BucketID != "" {
		if id, err := strconv.ParseInt(input.BucketID, 10, 64); err == nil && id <= 0 {
			return ValidationError{Field: "bucket_id", Message: "must be a positive integer"}
		}
	}
	return nil
}

func (h *Handlers) resolveBucketForTask(ctx context.Context, client *vikunja.Client, projectID int64, bucketID string) (*int64, error) {
	if bucketID == "" {
		return nil, nil
	}
	bucket, err := resolution.FindBucketByIDOrTitle(ctx, client, projectID, bucketID)
	if err != nil {
		return nil, err
	}
	return bucket, nil
}

func (h *Handlers) createTask(ctx context.Context, client *vikunja.Client, input CreateTaskInput, projectID int64, bucketID *int64) (*vikunja.Task, error) {
	return client.CreateTask(ctx, input.Title, projectID, input.Description, bucketID, time.Time{})
}

func (h *Handlers) formatTaskOutput(task *vikunja.Task) (*mcp.CallToolResult, CreateTaskOutput, error) {
	output := CreateTaskOutput{
		Task: toTask(task),
	}

	data, err := h.deps.OutputFormatter.Format(output)
	if err != nil {
		return nil, CreateTaskOutput{}, fmt.Errorf("failed to format response: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(data)},
		},
	}, output, nil
}
