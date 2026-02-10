// Package handlers provides MCP tool handlers for Vikunja integration.
package handlers

import (
	"context"
	"fmt"

	"github.com/meschbach/mcp-vikunja/pkg/vikunja"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// discoverVikunjaHandler provides AI-friendly discovery of all available resources
func (h *Handlers) discoverVikunjaHandler(ctx context.Context, _ *mcp.CallToolRequest, input DiscoverInput) (*mcp.CallToolResult, DiscoverOutput, error) {
	client, err := createVikunjaClient()
	if err != nil {
		return nil, DiscoverOutput{}, fmt.Errorf("failed to create client: %w", err)
	}

	// Set defaults
	maxProjects := 5
	if input.MaxProjects > 0 {
		maxProjects = input.MaxProjects
	}

	// Get all projects
	projects, err := client.GetProjects(ctx)
	if err != nil {
		return nil, DiscoverOutput{}, fmt.Errorf("failed to list projects: %w", err)
	}

	// Limit projects if requested
	if len(projects) > maxProjects {
		projects = projects[:maxProjects]
	}

	output, err := h.buildDiscoveryOutput(ctx, client, projects, input.IncludeCounts)
	if err != nil {
		return nil, DiscoverOutput{}, fmt.Errorf("failed to build discovery output: %w", err)
	}

	data, err := h.deps.OutputFormatter.Format(output)
	if err != nil {
		return nil, DiscoverOutput{}, fmt.Errorf("failed to format response: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(data)},
		},
	}, output, nil
}

// buildDiscoveryOutput constructs the complete discovery output
func (h *Handlers) buildDiscoveryOutput(ctx context.Context, client *vikunja.Client, projects []vikunja.Project, includeCounts bool) (DiscoverOutput, error) {
	projectsLen := len(projects)
	flatProjects := make([]ProjectFlat, 0, projectsLen)
	allViews := make([]ViewFlat, 0, projectsLen*3) // Estimate average 3 views per project
	var defaultProject *ProjectReference

	for i, p := range projects {
		// Get views for each project
		views, err := client.GetProjectViews(ctx, p.ID)
		if err != nil {
			h.deps.Logger.Warn("failed to get project views", "project_id", p.ID, "error", err)
			continue
		}

		flatProject, projectViews, defaultView := h.buildProjectInfo(p, views)
		flatProjects = append(flatProjects, flatProject)
		allViews = append(allViews, projectViews...)

		// Get task count if requested
		if includeCounts {
			tasks, err := client.GetTasks(ctx, p.ID)
			if err == nil {
				count := len(tasks)
				flatProjects[i].TaskCount = &count
			}
		}

		// Set default project (first one)
		if i == 0 {
			defaultProject = &ProjectReference{
				ID:          p.ID,
				Title:       p.Title,
				DefaultView: defaultView,
			}
		}
	}

	quickStart := h.buildQuickStart(defaultProject)
	toolGuide := h.buildToolGuide(defaultProject)
	errorPrevention := h.buildErrorPrevention()
	schemaInfo := h.buildSchemaInfo()
	serverInfo := h.buildServerInfo()

	return DiscoverOutput{
		Projects:        flatProjects,
		Views:           allViews,
		QuickStart:      quickStart,
		Tools:           toolGuide,
		ErrorPrevention: errorPrevention,
		SchemaInfo:      schemaInfo,
		ServerInfo:      serverInfo,
	}, nil
}

// buildProjectInfo creates flat project information and views
func (h *Handlers) buildProjectInfo(p vikunja.Project, views []vikunja.ProjectView) (ProjectFlat, []ViewFlat, ViewReference) {
	viewsLen := len(views)
	projectViews := make([]ViewFlat, 0, viewsLen)
	var defaultViewID int64 = 0
	var defaultViewTitle, defaultViewKind string

	for _, v := range views {
		projectViews = append(projectViews, ViewFlat{
			ID:        v.ID,
			ProjectID: v.ProjectID,
			Title:     v.Title,
			ViewKind:  string(v.ViewKind),
			IsDefault: v.ID == defaultViewID,
		})
		if defaultViewID == 0 {
			defaultViewID = v.ID
			defaultViewTitle = v.Title
			defaultViewKind = string(v.ViewKind)
		}
	}

	flatProject := ProjectFlat{
		ID:            p.ID,
		Title:         p.Title,
		DefaultViewID: defaultViewID,
		ViewCount:     len(views),
	}

	defaultView := ViewReference{
		ID:       defaultViewID,
		Title:    defaultViewTitle,
		ViewKind: defaultViewKind,
	}

	return flatProject, projectViews, defaultView
}

// buildQuickStart creates quick start examples
func (h *Handlers) buildQuickStart(defaultProject *ProjectReference) QuickStart {
	quickStart := QuickStart{
		DefaultProject: *defaultProject,
		CommonCalls: []OperationExample{
			{
				When:       "list default project tasks",
				Call:       "list_tasks()",
				Parameters: map[string]interface{}{},
			},
			{
				When: "list specific project tasks",
				Call: fmt.Sprintf("list_tasks(project_id=%d, view_id=%d)", defaultProject.ID, defaultProject.DefaultView.ID),
				Parameters: map[string]interface{}{
					"project_id": defaultProject.ID,
					"view_id":    defaultProject.DefaultView.ID,
				},
			},
			{
				When: "find specific project",
				Call: fmt.Sprintf("find_project_by_name(name='%s')", defaultProject.Title),
				Parameters: map[string]interface{}{
					"name": defaultProject.Title,
				},
				Alternatives: []string{
					"list_projects() to see all available projects",
				},
			},
		},
	}

	// Add task moving example if not in readonly mode
	if !h.isReadonly() {
		quickStart.CommonCalls = append(quickStart.CommonCalls, OperationExample{
			When: "move task between buckets",
			Call: "move_task_to_bucket(task_id=123, project_id=1, view_id=2, bucket_id=5)",
			Parameters: map[string]interface{}{
				"task_id":    123,
				"project_id": 1,
				"view_id":    2,
				"bucket_id":  5,
			},
		})
	}

	return quickStart
}

// buildToolGuide creates comprehensive tool documentation
func (h *Handlers) buildToolGuide(defaultProject *ProjectReference) ToolGuide {
	toolGuide := ToolGuide{
		Tools: map[string]ToolInfo{
			"list_tasks": {
				Purpose: "get_project_tasks",
				Parameters: map[string]ParamInfo{
					"project_id": {
						Name:        "project_id",
						Type:        "int64",
						Source:      "projects.id",
						Default:     stringPtr(fmt.Sprintf("%d", defaultProject.ID)),
						Description: "Project ID from discovery",
					},
					"view_id": {
						Name:        "view_id",
						Type:        "int64",
						Source:      "views.id",
						Default:     stringPtr(fmt.Sprintf("%d", defaultProject.DefaultView.ID)),
						Description: "View ID from project's views",
					},
					"project_title": {
						Name:        "project_title",
						Type:        "string",
						Source:      "projects.title",
						Default:     stringPtr("Inbox"),
						Description: "Exact project title from discovery",
					},
				},
				CommonCombinations: []ParamCombo{
					{
						UseCase:    "Default project tasks",
						Parameters: map[string]string{},
					},
					{
						UseCase: "Specific project tasks",
						Parameters: map[string]string{
							"project_id": "Use projects[0].id",
							"view_id":    "Use views where project_id matches project.id",
						},
					},
					{
						UseCase: "By project title",
						Parameters: map[string]string{
							"project_title": "Use exact title from projects array",
						},
					},
				},
				ExampleCalls: []string{
					"list_tasks() - uses defaults",
					fmt.Sprintf("list_tasks(project_id=%d, view_id=%d) - specific project/view", defaultProject.ID, defaultProject.DefaultView.ID),
					"list_tasks(project_title='Inbox') - by name",
				},
			},
		},
	}

	// Add move task tool if not in readonly mode
	if !h.isReadonly() {
		toolGuide.Tools["move_task_to_bucket"] = ToolInfo{
			Purpose: "move_task_between_buckets",
			Parameters: map[string]ParamInfo{
				"task_id": {
					Name:        "task_id",
					Type:        "int64",
					Source:      "tasks.id",
					Description: "Task ID from task listings",
				},
				"project_id": {
					Name:        "project_id",
					Type:        "int64",
					Source:      "projects.id",
					Description: "Project ID containing task",
				},
				"view_id": {
					Name:        "view_id",
					Type:        "int64",
					Source:      "views.id",
					Description: "View ID where task buckets are located",
				},
				"bucket_id": {
					Name:        "bucket_id",
					Type:        "int64",
					Source:      "buckets.id",
					Description: "Target bucket ID to move task to",
				},
			},
			CommonCombinations: []ParamCombo{
				{
					UseCase: "Move task between buckets",
					Parameters: map[string]string{
						"task_id":    "From task list or get_task",
						"project_id": "From task's project_id",
						"view_id":    "From available views list",
						"bucket_id":  "From bucket list",
					},
				},
			},
			ExampleCalls: []string{
				"move_task_to_bucket(task_id=123, project_id=1, view_id=2, bucket_id=5)",
				"move_task_to_bucket(task_id='task-id', project_id='project-id', view_id='view-id', bucket_id='bucket-id')",
			},
		}
	}

	return toolGuide
}

// buildErrorPrevention creates error prevention guidelines
func (h *Handlers) buildErrorPrevention() ErrorPrevention {
	return ErrorPrevention{
		CommonMistakes: []CommonMistake{
			{
				Error:      "invalid project_id",
				Prevention: "Use exact IDs from projects array in discovery response",
			},
			{
				Error:      "view not found",
				Prevention: "Verify view_id exists in project's views array",
			},
			{
				Error:      "project not found",
				Prevention: "Use exact project_title from projects array or call find_project_by_name",
			},
		},
		ValidationRules: []ValidationRule{
			{
				Field:  "project_id",
				Type:   "int64",
				Source: "projects.id",
			},
			{
				Field:  "view_id",
				Type:   "int64",
				Source: "views.id",
			},
			{
				Field:  "view_kind",
				Type:   "enum",
				Values: []string{"kanban", "list", "gantt", "table"},
				Source: "views.view_kind",
			},
		},
	}
}

// buildSchemaInfo creates schema information
func (h *Handlers) buildSchemaInfo() SchemaInfo {
	return SchemaInfo{
		Version:        "1.0",
		RequiredFields: []string{"projects", "tools"},
		FieldSources: map[string]string{
			"project_id":    "projects.id",
			"view_id":       "views.id",
			"view_kind":     "views.view_kind",
			"project_title": "projects.title",
		},
	}
}

// buildServerInfo creates server information
func (h *Handlers) buildServerInfo() ServerInfo {
	features := []string{"list_tasks", "get_task", "list_projects", "readonly"}
	if !h.isReadonly() {
		features = append(features, "move_task_to_bucket", "create_task")
	}

	return ServerInfo{
		APIVersion: "v1",
		Status:     "connected",
		Features:   features,
	}
}

// isReadonly returns true if server is in readonly mode
func (h *Handlers) isReadonly() bool {
	if h.deps.Config != nil {
		return h.deps.Config.Readonly
	}
	return false
}
