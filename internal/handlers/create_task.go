// Package handlers provides MCP tool handlers for Vikunja integration.
package handlers

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/meschbach/mcp-vikunja/pkg/resolution"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// createTaskHandler handles the create_task tool
func (h *Handlers) createTaskHandler(ctx context.Context, _ *mcp.CallToolRequest, input CreateTaskInput) (*mcp.CallToolResult, CreateTaskOutput, error) {
	// Check readonly mode
	if h.isReadonly() {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Operation not available in readonly mode"},
			},
		}, CreateTaskOutput{}, fmt.Errorf("operation not available in readonly mode")
	}

	// Validate required fields
	if err := validateRequiredString("title", input.Title); err != nil {
		return h.buildErrorResult(err.Error()), CreateTaskOutput{}, err
	}
	if err := validateRequiredString("project_id", input.ProjectID); err != nil {
		return h.buildErrorResult(err.Error()), CreateTaskOutput{}, err
	}

	// Validate bucket_id format early if it's intended as an integer.
	// If bucket_id parses as an integer, it must be positive.
	// Non-numeric strings are allowed as bucket titles.
	if input.BucketID != "" {
		if id, err := strconv.ParseInt(input.BucketID, 10, 64); err == nil {
			// Parsed as integer - ensure positive
			if id <= 0 {
				valErr := ValidationError{Field: "bucket_id", Message: "must be a positive integer"}
				return h.buildErrorResult(valErr.Error()), CreateTaskOutput{}, valErr
			}
		}
		// If parsing failed, it's treated as a title - no early error
	}

	// Create Vikunja client
	client, err := createVikunjaClient()
	if err != nil {
		return h.buildErrorResult(fmt.Sprintf("failed to create Vikunja client: %v", err)), CreateTaskOutput{}, fmt.Errorf("failed to create client: %w", err)
	}

	// Resolve project (ID-first with duplicate detection)
	project, err := resolution.ResolveProject(ctx, client, input.ProjectID)
	if err != nil {
		return h.buildErrorResult(err.Error()), CreateTaskOutput{}, err
	}

	// Resolve optional bucket
	var bucketID *int64
	if input.BucketID != "" {
		bucket, err := resolution.FindBucketByIDOrTitle(ctx, client, project.ID, input.BucketID)
		if err != nil {
			return h.buildErrorResult(err.Error()), CreateTaskOutput{}, err
		}
		bucketID = bucket
	}

	// Create task (no due date support)
	task, err := client.CreateTask(ctx, input.Title, project.ID, input.Description, bucketID, time.Time{})
	if err != nil {
		return h.buildErrorResult(fmt.Sprintf("failed to create task: %v", err)), CreateTaskOutput{}, fmt.Errorf("failed to create task: %w", err)
	}

	// Convert to output format
	output := CreateTaskOutput{
		Task: toTask(*task),
	}

	// Format response
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
