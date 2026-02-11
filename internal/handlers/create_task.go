// Package handlers provides MCP tool handlers for Vikunja integration.
package handlers

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// createTaskHandler handles the create_task tool
func (h *Handlers) createTaskHandler(_ context.Context, _ *mcp.CallToolRequest, input CreateTaskInput) (*mcp.CallToolResult, CreateTaskOutput, error) {
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

	if _, err := parseID("project_id", input.ProjectID); err != nil {
		return h.buildErrorResult(err.Error()), CreateTaskOutput{}, err
	}

	return nil, CreateTaskOutput{}, fmt.Errorf("create task not implemented in Phase 1 (read-only operations only)")
}
