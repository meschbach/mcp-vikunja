// Package handlers provides MCP tool handlers for Vikunja integration.
package handlers

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// listProjectsHandler handles the list_projects tool
func (h *Handlers) listProjectsHandler(ctx context.Context, _ *mcp.CallToolRequest, _ ListProjectsInput) (*mcp.CallToolResult, ListProjectsOutput, error) {
	client, err := createVikunjaClient()
	if err != nil {
		return nil, ListProjectsOutput{}, err
	}

	projects, err := client.GetProjects(ctx)
	if err != nil {
		return nil, ListProjectsOutput{}, fmt.Errorf("failed to list projects: %w", err)
	}

	output := ListProjectsOutput{
		Projects: projects,
	}

	data, err := h.deps.OutputFormatter.Format(output.Projects)
	if err != nil {
		return nil, ListProjectsOutput{}, fmt.Errorf("failed to format response: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(data)},
		},
	}, output, nil
}

// findProjectByNameHandler handles the find_project_by_name tool
func (h *Handlers) findProjectByNameHandler(ctx context.Context, _ *mcp.CallToolRequest, input FindProjectByNameInput) (*mcp.CallToolResult, FindProjectByNameOutput, error) {
	client, err := createVikunjaClient()
	if err != nil {
		return nil, FindProjectByNameOutput{}, err
	}

	projects, err := client.GetProjects(ctx)
	if err != nil {
		return nil, FindProjectByNameOutput{}, fmt.Errorf("failed to list projects: %w", err)
	}

	found := false
	var project Project
	for _, p := range projects {
		if p.Title == input.Name {
			project = Project{
				ID:    p.ID,
				Title: p.Title,
				URI:   fmt.Sprintf("vikunja://project/%d", p.ID),
			}
			found = true
			break
		}
	}
	if !found {
		return h.buildErrorResult(fmt.Sprintf("project with title %q not found", input.Name)), FindProjectByNameOutput{}, fmt.Errorf("project with title %q not found", input.Name)
	}

	data, err := h.deps.OutputFormatter.Format(project)
	if err != nil {
		return nil, FindProjectByNameOutput{}, fmt.Errorf("failed to format response: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(data)},
		},
	}, FindProjectByNameOutput{Project: project}, nil
}
