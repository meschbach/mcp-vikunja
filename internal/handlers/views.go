// Package handlers provides MCP tool handlers for Vikunja integration.
package handlers

import (
	"context"
	"fmt"

	"github.com/meschbach/mcp-vikunja/pkg/vikunja"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// findViewHandler handles the find_view tool
func (h *Handlers) findViewHandler(ctx context.Context, _ *mcp.CallToolRequest, input FindViewInput) (*mcp.CallToolResult, FindViewOutput, error) {
	// Validate required field
	if err := validateRequiredString("view_name", input.ViewName); err != nil {
		return h.buildErrorResult(err.Error()), FindViewOutput{}, err
	}

	client, err := createVikunjaClient()
	if err != nil {
		return nil, FindViewOutput{}, fmt.Errorf("failed to create client: %w", err)
	}

	project, err := findProjectByIDOrTitle(ctx, client, input.ProjectID, input.ProjectTitle)
	if err != nil {
		return h.buildErrorResult(err.Error()), FindViewOutput{}, err
	}

	views, err := client.GetProjectViews(ctx, project.ID)
	if err != nil {
		return h.buildErrorResult(err.Error()), FindViewOutput{}, fmt.Errorf("failed to get project views: %w", err)
	}

	foundView, err := findViewByName(views, input.ViewName, input.Fuzzy, project.Title)
	if err != nil {
		return h.buildErrorResult(err.Error()), FindViewOutput{}, fmt.Errorf("%v in project %q", err, project.Title)
	}

	output := FindViewOutput{
		Project: *project,
		View:    toView(*foundView),
	}

	// Convert handlers.FindViewOutput to vikunja.ViewOutput for formatting
	vikunjaOutput := vikunja.ViewOutput{
		Project: vikunja.Project{
			ID:    output.Project.ID,
			Title: output.Project.Title,
			// Note: handlers.Project has limited fields compared to vikunja.Project
		},
		View: vikunja.ProjectView{
			ID:                      output.View.ID,
			ProjectID:               output.View.ProjectID,
			Title:                   output.View.Title,
			ViewKind:                output.View.ViewKind,
			Position:                output.View.Position,
			BucketConfigurationMode: output.View.BucketConfigurationMode,
			DefaultBucketID:         output.View.DefaultBucketID,
			DoneBucketID:            output.View.DoneBucketID,
		},
	}

	data, err := h.deps.OutputFormatter.Format(vikunjaOutput)
	if err != nil {
		return nil, FindViewOutput{}, fmt.Errorf("failed to format response: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(data)},
		},
	}, output, nil
}

// listViewsHandler handles the list_views tool
func (h *Handlers) listViewsHandler(ctx context.Context, _ *mcp.CallToolRequest, input ListViewsInput) (*mcp.CallToolResult, ListViewsOutput, error) {
	// Validate view_kind if provided
	if err := validateViewKind(input.ViewKind); err != nil {
		return h.buildErrorResult(err.Error()), ListViewsOutput{}, err
	}

	client, err := createVikunjaClient()
	if err != nil {
		return nil, ListViewsOutput{}, fmt.Errorf("failed to create client: %w", err)
	}

	project, err := findProjectByIDOrTitle(ctx, client, input.ProjectID, input.ProjectTitle)
	if err != nil {
		return h.buildErrorResult(err.Error()), ListViewsOutput{}, err
	}

	views, err := client.GetProjectViews(ctx, project.ID)
	if err != nil {
		return h.buildErrorResult(err.Error()), ListViewsOutput{}, fmt.Errorf("failed to get project views: %w", err)
	}

	var filteredViews []vikunja.ProjectView
	if input.ViewKind != "" {
		viewKind := vikunja.ViewKind(input.ViewKind)
		for _, v := range views {
			if v.ViewKind == viewKind {
				filteredViews = append(filteredViews, v)
			}
		}
	} else {
		filteredViews = views
	}

	output := ListViewsOutput{
		Project: *project,
		Views:   toViews(filteredViews),
	}

	// Convert handlers.ListViewsOutput to vikunja.ViewsOutput for formatting
	vikunjaOutput := vikunja.ViewsOutput{
		Project: vikunja.Project{
			ID:    output.Project.ID,
			Title: output.Project.Title,
			// Note: handlers.Project has limited fields
		},
		Views: filteredViews, // Already in vikunja format
	}

	data, err := h.deps.OutputFormatter.Format(vikunjaOutput)
	if err != nil {
		return nil, ListViewsOutput{}, fmt.Errorf("failed to format response: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(data)},
		},
	}, output, nil
}
