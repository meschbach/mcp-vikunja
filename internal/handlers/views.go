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
		View:    toView(foundView),
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
	if err := validateViewKind(input.ViewKind); err != nil {
		return h.buildErrorResult(err.Error()), ListViewsOutput{}, err
	}

	client, err := createVikunjaClient()
	if err != nil {
		return nil, ListViewsOutput{}, fmt.Errorf("failed to create client: %w", err)
	}

	project, views, err := h.resolveProjectAndViews(ctx, client, input)
	if err != nil {
		return h.buildErrorResult(err.Error()), ListViewsOutput{}, err
	}

	filteredViews := h.filterViewsByKind(views, input.ViewKind)

	return h.formatListViewsOutput(filteredViews, project)
}

func (h *Handlers) resolveProjectAndViews(ctx context.Context, client *vikunja.Client, input ListViewsInput) (*Project, []*vikunja.ProjectView, error) {
	project, err := findProjectByIDOrTitle(ctx, client, input.ProjectID, input.ProjectTitle)
	if err != nil {
		return nil, nil, err
	}

	views, err := client.GetProjectViews(ctx, project.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get project views: %w", err)
	}

	return project, views, nil
}

func (h *Handlers) filterViewsByKind(views []*vikunja.ProjectView, viewKind string) []*vikunja.ProjectView {
	if viewKind == "" {
		return views
	}

	kind := vikunja.ViewKind(viewKind)
	filtered := make([]*vikunja.ProjectView, 0, len(views))
	for _, v := range views {
		if v.ViewKind == kind {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

func (h *Handlers) formatListViewsOutput(views []*vikunja.ProjectView, project *Project) (*mcp.CallToolResult, ListViewsOutput, error) {
	output := ListViewsOutput{
		Project: *project,
		Views:   toViews(views),
	}

	vikunjaOutput := vikunja.ViewsOutput{
		Project: vikunja.Project{
			ID:    output.Project.ID,
			Title: output.Project.Title,
		},
		Views: views,
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
