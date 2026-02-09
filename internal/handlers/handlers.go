// Package handlers provides MCP tool handlers for Vikunja integration.
package handlers

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/meschbach/mcp-vikunja/internal/config"
	"github.com/meschbach/mcp-vikunja/pkg/vikunja"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Global output formatter
var outputFormatter vikunja.OutputFormatter

type ListTasksInput struct {
	ProjectID    string `json:"project_id,omitempty" jsonschema:"Optional project ID to filter tasks"`
	ProjectTitle string `json:"project_title,omitempty" jsonschema:"Optional project title (defaults to 'Inbox')"`
	ViewID       string `json:"view_id,omitempty" jsonschema:"Optional project view ID"`
	ViewTitle    string `json:"view_title,omitempty" jsonschema:"Optional project view title (defaults to 'Kanban')"`
}

// TaskSummary is a minimal version of a task for listing
type TaskSummary struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	URI   string `json:"uri"`
}

// BucketSummary is a minimal version of a bucket for listing
type BucketSummary struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

// BucketTasksSummary represents a bucket and its associated tasks for listing
type BucketTasksSummary struct {
	Bucket BucketSummary `json:"bucket"`
	Tasks  []TaskSummary `json:"tasks,omitempty"`
}

// ViewTasksSummary represents all buckets and tasks in a view for listing
type ViewTasksSummary struct {
	ViewID    int64                `json:"view_id"`
	ViewTitle string               `json:"view_title"`
	Buckets   []BucketTasksSummary `json:"buckets,omitempty" jsonschema:"Buckets the tasks are organized into"`
}

type ListTasksOutput struct {
	Project *Project         `json:"project,omitempty" jsonschema:"Project hte tasks are related to"`
	View    ViewTasksSummary `json:"view" jsonschema:"tasks associated with this view"`
}

type GetTaskInput struct {
	TaskID         string `json:"task_id" jsonschema:"The ID of the task to retrieve"`
	IncludeBuckets bool   `json:"include_buckets,omitempty" jsonschema:"Whether to include bucket information across all project views (default: true)"`
}

type GetTaskOutput struct {
	Task    Task                    `json:"task"`
	Buckets *vikunja.TaskBucketInfo `json:"buckets,omitempty"`
}

type ListBucketsInput struct {
	ProjectID string `json:"project_id" jsonschema:"The ID of the project"`
	ViewID    string `json:"view_id" jsonschema:"The ID of the project view"`
}

type ListBucketsOutput struct {
	Buckets []Bucket `json:"buckets"`
}

type ListProjectsInput struct {
}

type ListProjectsOutput struct {
	Projects []Project `json:"projects"`
}

type CreateTaskInput struct {
	Title       string `json:"title" jsonschema:"The title of the task"`
	Description string `json:"description,omitempty" jsonschema:"Optional task description"`
	ProjectID   string `json:"project_id" jsonschema:"The project ID to create the task in"`
}

type CreateTaskOutput struct {
	Task Task `json:"task"`
}

type FindProjectByNameInput struct {
	Name string `json:"name" jsonschema:"The name/title of the project to find"`
}

type FindProjectByNameOutput struct {
	Project Project `json:"project"`
}

type FindViewInput struct {
	ProjectID    string `json:"project_id,omitempty" jsonschema:"Optional project ID to search in (overrides project_title)"`
	ProjectTitle string `json:"project_title,omitempty" jsonschema:"Optional project title to search in"`
	ViewName     string `json:"view_name" jsonschema:"The name/title of the view to find"`
	Fuzzy        bool   `json:"fuzzy,omitempty" jsonschema:"Enable fuzzy/partial matching for view names (default: false)"`
}

type FindViewOutput struct {
	Project Project `json:"project"`
	View    View    `json:"view"`
}

type ListViewsInput struct {
	ProjectID    string `json:"project_id,omitempty" jsonschema:"Optional project ID to list views for (overrides project_title)"`
	ProjectTitle string `json:"project_title,omitempty" jsonschema:"Optional project title to list views for"`
	ViewKind     string `json:"view_kind,omitempty" jsonschema:"Optional filter by view kind (list, kanban, gantt, table)"`
}

type ListViewsOutput struct {
	Project Project `json:"project"`
	Views   []View  `json:"views"`
}

// View is a simplified version of vikunja.ProjectView to avoid recursive cycles in JSON schema
type View struct {
	ID                      int64                           `json:"id"`
	ProjectID               int64                           `json:"project_id"`
	Title                   string                          `json:"title"`
	ViewKind                vikunja.ViewKind                `json:"view_kind"`
	Position                float64                         `json:"position"`
	BucketConfigurationMode vikunja.BucketConfigurationMode `json:"bucket_configuration_mode"`
	DefaultBucketID         int64                           `json:"default_bucket_id,omitempty"`
	DoneBucketID            int64                           `json:"done_bucket_id,omitempty"`
	URI                     string                          `json:"uri"`
}

// Task is a simplified version of vikunja.Task to avoid recursive cycles in JSON schema
type Task struct {
	ID          int64    `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	ProjectID   int64    `json:"project_id,omitempty"`
	Done        bool     `json:"done"`
	DueDate     string   `json:"due_date,omitempty"`
	Created     string   `json:"created"`
	Updated     string   `json:"updated"`
	Buckets     []Bucket `json:"buckets,omitempty"`
	Position    float64  `json:"position"`
}

// Bucket is a simplified version of vikunja.Bucket to avoid recursive cycles in JSON schema
type Bucket struct {
	ID            int64   `json:"id"`
	ProjectViewID int64   `json:"project_view_id"`
	Title         string  `json:"title"`
	Description   string  `json:"description,omitempty"`
	Limit         int     `json:"limit"`
	Position      float64 `json:"position"`
	IsDoneBucket  bool    `json:"is_done_bucket"`
}

// Project is a simplified version of vikunja.Project
type Project struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	URI   string `json:"uri"`
}

// BucketTasks represents a bucket and its associated tasks
type BucketTasks struct {
	Bucket Bucket `json:"bucket"`
	Tasks  []Task `json:"tasks,omitempty"`
}

// ViewTasks represents all buckets and tasks in a view
type ViewTasks struct {
	ViewID    int64         `json:"view_id"`
	ViewTitle string        `json:"view_title"`
	Buckets   []BucketTasks `json:"buckets,omitempty"`
}

// Discovery types for AI-friendly interface
type DiscoverInput struct {
	MaxProjects   int  `json:"max_projects,omitempty" jsonschema:"Maximum projects to return (default: 5)"`
	IncludeCounts bool `json:"include_counts,omitempty" jsonschema:"Include task counts (slower)"`
}

type DiscoverOutput struct {
	Projects        []ProjectFlat   `json:"projects"`
	Views           []ViewFlat      `json:"views"`
	QuickStart      QuickStart      `json:"quick_start"`
	Tools           ToolGuide       `json:"tools"`
	ErrorPrevention ErrorPrevention `json:"error_prevention"`
	SchemaInfo      SchemaInfo      `json:"schema_info"`
	ServerInfo      ServerInfo      `json:"server_info"`
}

type ProjectFlat struct {
	ID            int64  `json:"id"`
	Title         string `json:"title"`
	DefaultViewID int64  `json:"default_view_id"`
	ViewCount     int    `json:"view_count"`
	TaskCount     *int   `json:"task_count,omitempty"`
}

type ViewFlat struct {
	ID        int64  `json:"id"`
	ProjectID int64  `json:"project_id"`
	Title     string `json:"title"`
	ViewKind  string `json:"view_kind"`
	IsDefault bool   `json:"is_default"`
}

type QuickStart struct {
	DefaultProject ProjectReference   `json:"default_project"`
	CommonCalls    []OperationExample `json:"common_calls"`
}

type ProjectReference struct {
	ID          int64         `json:"id"`
	Title       string        `json:"title"`
	DefaultView ViewReference `json:"default_view"`
}

type ViewReference struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	ViewKind string `json:"view_kind"`
}

type OperationExample struct {
	When         string                 `json:"when"`
	Call         string                 `json:"call"`
	Alternatives []string               `json:"alternatives,omitempty"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
}

type ToolGuide struct {
	Tools map[string]ToolInfo `json:"tools"`
}

type ToolInfo struct {
	Purpose            string               `json:"purpose"`
	Parameters         map[string]ParamInfo `json:"parameters"`
	CommonCombinations []ParamCombo         `json:"common_combinations"`
	ExampleCalls       []string             `json:"example_calls"`
}

type ParamInfo struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Source      string   `json:"source"`
	Default     *string  `json:"default,omitempty"`
	Description string   `json:"description"`
	ValidValues []string `json:"valid_values,omitempty"`
}

type ParamCombo struct {
	UseCase    string            `json:"use_case"`
	Parameters map[string]string `json:"parameters"`
}

type ErrorPrevention struct {
	CommonMistakes  []CommonMistake  `json:"common_mistakes"`
	ValidationRules []ValidationRule `json:"validation_rules"`
}

type CommonMistake struct {
	Error      string `json:"error"`
	Prevention string `json:"prevention"`
}

type ValidationRule struct {
	Field  string   `json:"field"`
	Type   string   `json:"type"`
	Values []string `json:"values,omitempty"`
	Source string   `json:"source"`
}

type SchemaInfo struct {
	Version        string            `json:"version"`
	RequiredFields []string          `json:"required_fields"`
	FieldSources   map[string]string `json:"field_sources"`
}

type ServerInfo struct {
	APIVersion string   `json:"api_version"`
	Status     string   `json:"status"`
	Features   []string `json:"features"`
}

func toTaskSummary(t vikunja.Task) TaskSummary {
	return TaskSummary{
		ID:    t.ID,
		Title: t.Title,
		URI:   fmt.Sprintf("vikunja://task/%d", t.ID),
	}
}

func toTasksSummary(tasks []vikunja.Task) []TaskSummary {
	if tasks == nil {
		return nil
	}
	res := make([]TaskSummary, len(tasks))
	for i, t := range tasks {
		res[i] = toTaskSummary(t)
	}
	return res
}

func toBucketSummary(b vikunja.Bucket) BucketSummary {
	return BucketSummary{
		ID:    b.ID,
		Title: b.Title,
	}
}

func toTask(t vikunja.Task) Task {
	dueDate := ""
	if !t.DueDate.IsZero() {
		dueDate = t.DueDate.Format(time.RFC3339)
	}
	created := ""
	if !t.Created.IsZero() {
		created = t.Created.Format(time.RFC3339)
	}
	updated := ""
	if !t.Updated.IsZero() {
		updated = t.Updated.Format(time.RFC3339)
	}

	return Task{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		ProjectID:   t.ProjectID,
		Done:        t.Done,
		DueDate:     dueDate,
		Created:     created,
		Updated:     updated,
		Buckets:     toBuckets(t.Buckets),
		Position:    t.Position,
	}
}

func toTasks(tasks []vikunja.Task) []Task {
	if tasks == nil {
		return nil
	}
	res := make([]Task, len(tasks))
	for i, t := range tasks {
		res[i] = toTask(t)
	}
	return res
}

func toBucket(b vikunja.Bucket) Bucket {
	return Bucket{
		ID:            b.ID,
		ProjectViewID: b.ProjectViewID,
		Title:         b.Title,
		Description:   b.Description,
		Limit:         b.Limit,
		Position:      b.Position,
		IsDoneBucket:  b.IsDoneBucket,
	}
}

func toBuckets(buckets []vikunja.Bucket) []Bucket {
	if buckets == nil {
		return nil
	}
	res := make([]Bucket, len(buckets))
	for i, b := range buckets {
		res[i] = toBucket(b)
	}
	return res
}

func toView(v vikunja.ProjectView) View {
	return View{
		ID:                      v.ID,
		ProjectID:               v.ProjectID,
		Title:                   v.Title,
		ViewKind:                v.ViewKind,
		Position:                v.Position,
		BucketConfigurationMode: v.BucketConfigurationMode,
		DefaultBucketID:         v.DefaultBucketID,
		DoneBucketID:            v.DoneBucketID,
		URI:                     fmt.Sprintf("vikunja://project/%d/view/%d", v.ProjectID, v.ID),
	}
}

func toViews(views []vikunja.ProjectView) []View {
	if views == nil {
		return nil
	}
	res := make([]View, len(views))
	for i, v := range views {
		res[i] = toView(v)
	}
	return res
}

func toVikunjaTask(t TaskSummary) vikunja.Task {
	return vikunja.Task{
		ID:    t.ID,
		Title: t.Title,
		// Populate other fields with default/zero values as they are not available in TaskSummary
		Description: "",
		ProjectID:   0,
		Done:        false,
		DueDate:     time.Time{},
		Created:     time.Time{},
		Updated:     time.Time{},
		Buckets:     nil,
		Position:    0,
	}
}

func toVikunjaBucket(b BucketSummary) vikunja.Bucket {
	return vikunja.Bucket{
		ID:    b.ID,
		Title: b.Title,
		// Populate other fields with default/zero values
		ProjectViewID: 0,
		Description:   "",
		Limit:         0,
		Position:      0,
		IsDoneBucket:  false,
		Tasks:         nil,
	}
}

func toVikunjaBucketTasks(bts BucketTasksSummary) vikunja.BucketTasks {
	vikunjaTasks := make([]vikunja.Task, len(bts.Tasks))
	for i, taskSummary := range bts.Tasks {
		vikunjaTasks[i] = toVikunjaTask(taskSummary)
	}
	return vikunja.BucketTasks{
		Bucket: toVikunjaBucket(bts.Bucket),
		Tasks:  vikunjaTasks,
	}
}

func toVikunjaViewTasks(vts ViewTasksSummary) *vikunja.ViewTasks {
	vikunjaBuckets := make([]vikunja.BucketTasks, len(vts.Buckets))
	for i, bucketTaskSummary := range vts.Buckets {
		vikunjaBuckets[i] = toVikunjaBucketTasks(bucketTaskSummary)
	}
	return &vikunja.ViewTasks{
		ViewID:    vts.ViewID,
		ViewTitle: vts.ViewTitle,
		Buckets:   vikunjaBuckets,
	}
}

// createVikunjaClient creates a new Vikunja client using environment variables
func createVikunjaClient() (*vikunja.Client, error) {
	host := os.Getenv("VIKUNJA_HOST")
	token := os.Getenv("VIKUNJA_TOKEN")
	if host == "" || token == "" {
		return nil, fmt.Errorf("VIKUNJA_HOST and VIKUNJA_TOKEN environment variables required")
	}

	return vikunja.NewClient(host, token, false)
}

// findProjectByIDOrTitle finds a project by ID or title
func findProjectByIDOrTitle(ctx context.Context, client *vikunja.Client, projectID, projectTitle string) (*Project, error) {
	if projectID != "" {
		id, err := strconv.ParseInt(projectID, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid project_id: %s", projectID)
		}
		return &Project{
			ID:    id,
			Title: fmt.Sprintf("Project %d", id),
			URI:   fmt.Sprintf("vikunja://project/%d", id),
		}, nil
	}

	if projectTitle == "" {
		return nil, fmt.Errorf("either project_id or project_title must be specified")
	}

	projects, err := client.GetProjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	var projectTitles []string
	for _, p := range projects {
		projectTitles = append(projectTitles, p.Title)
		if p.Title == projectTitle {
			return &Project{
				ID:    p.ID,
				Title: p.Title,
				URI:   fmt.Sprintf("vikunja://project/%d", p.ID),
			}, nil
		}
	}

	return nil, enhancedProjectNotFoundError(projectTitle, projectTitles)
}

// findViewByName finds a view by name within a project's views
func findViewByName(views []vikunja.ProjectView, viewName string, fuzzy bool, projectName string) (*vikunja.ProjectView, error) {
	var viewTitles []string
	for _, v := range views {
		viewTitles = append(viewTitles, v.Title)
		if fuzzy && containsIgnoreCase(v.Title, viewName) {
			return &v, nil
		}
		if !fuzzy && v.Title == viewName {
			return &v, nil
		}
	}

	return nil, enhancedViewNotFoundError(viewName, projectName, viewTitles)
}

// discoverVikunjaHandler provides AI-friendly discovery of all available resources
func discoverVikunjaHandler(ctx context.Context, _ *mcp.CallToolRequest, input DiscoverInput) (*mcp.CallToolResult, DiscoverOutput, error) {
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

	// Convert to flat structure
	var flatProjects []ProjectFlat
	var allViews []ViewFlat
	var defaultProject *ProjectReference

	for i, p := range projects {
		// Get views for each project
		views, err := client.GetProjectViews(ctx, p.ID)
		if err != nil {
			return nil, DiscoverOutput{}, fmt.Errorf("failed to get project views for %d: %w", p.ID, err)
		}

		// Find default view (first one for now)
		var defaultViewID int64 = 0
		var defaultViewTitle, defaultViewKind string
		for _, v := range views {
			allViews = append(allViews, ViewFlat{
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

		// Get task count if requested
		var taskCount *int
		if input.IncludeCounts {
			tasks, err := client.GetTasks(ctx, p.ID)
			if err == nil {
				count := len(tasks)
				taskCount = &count
			}
		}

		flatProject := ProjectFlat{
			ID:            p.ID,
			Title:         p.Title,
			DefaultViewID: defaultViewID,
			ViewCount:     len(views),
			TaskCount:     taskCount,
		}
		flatProjects = append(flatProjects, flatProject)

		// Set default project (first one)
		if i == 0 {
			defaultProject = &ProjectReference{
				ID:    p.ID,
				Title: p.Title,
				DefaultView: ViewReference{
					ID:       defaultViewID,
					Title:    defaultViewTitle,
					ViewKind: defaultViewKind,
				},
			}
		}
	}

	// Build quick start
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
					fmt.Sprintf("list_projects() to see all available projects"),
				},
			},
		},
	}

	// Build tool guide
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

	// Error prevention
	errorPrevention := ErrorPrevention{
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

	// Schema info
	schemaInfo := SchemaInfo{
		Version:        "1.0",
		RequiredFields: []string{"projects", "tools"},
		FieldSources: map[string]string{
			"project_id":    "projects.id",
			"view_id":       "views.id",
			"view_kind":     "views.view_kind",
			"project_title": "projects.title",
		},
	}

	// Server info
	serverInfo := ServerInfo{
		APIVersion: "v1",
		Status:     "connected",
		Features:   []string{"list_tasks", "get_task", "list_projects", "readonly"},
	}

	output := DiscoverOutput{
		Projects:        flatProjects,
		Views:           allViews,
		QuickStart:      quickStart,
		Tools:           toolGuide,
		ErrorPrevention: errorPrevention,
		SchemaInfo:      schemaInfo,
		ServerInfo:      serverInfo,
	}

	data, err := outputFormatter.Format(output)
	if err != nil {
		return nil, DiscoverOutput{}, fmt.Errorf("failed to format response: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(data)},
		},
	}, output, nil
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}

// Register adds all Vikunja tool handlers to the MCP server.
func Register(s *mcp.Server, cfg *config.Config) {
	// Initialize the output formatter based on configuration
	outputFormatter = vikunja.GetFormatter(cfg.OutputFormat)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "discover_vikunja",
		Description: "Discover all available Vikunja resources with AI-friendly guidance",
	}, discoverVikunjaHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_tasks",
		Description: "List tasks from Vikunja",
	}, listTasksHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_task",
		Description: "Get details of a specific task",
	}, getTaskHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_buckets",
		Description: "List all buckets in a project view",
	}, listBucketsHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_projects",
		Description: "List all projects via this Vikunja connection.   Provides a list of projects including the ID, name, and URI",
	}, listProjectsHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_task",
		Description: "Create a new task in Vikunja",
	}, createTaskHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "find_project_by_name",
		Description: "Find a project by its name/title",
	}, findProjectByNameHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "find_view",
		Description: "Find a specific view by name within a project",
	}, findViewHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_views",
		Description: "List all views for a project, optionally filtered by view kind",
	}, listViewsHandler)
}

func listTasksHandler(ctx context.Context, _ *mcp.CallToolRequest, input ListTasksInput) (*mcp.CallToolResult, ListTasksOutput, error) {
	host := os.Getenv("VIKUNJA_HOST")
	token := os.Getenv("VIKUNJA_TOKEN")
	if host == "" || token == "" {
		return nil, ListTasksOutput{}, fmt.Errorf("VIKUNJA_HOST and VIKUNJA_TOKEN environment variables required")
	}

	client, err := vikunja.NewClient(host, token, false)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
		}, ListTasksOutput{}, fmt.Errorf("failed to create client: %w", err)
	}

	var project *Project
	var targetProjectID int64
	if input.ProjectID != "" {
		targetProjectID, err = strconv.ParseInt(input.ProjectID, 10, 64)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
			}, ListTasksOutput{}, fmt.Errorf("invalid project_id: %s", input.ProjectID)
		}
	} else {
		// Default project title to "Inbox" if not specified
		projectTitle := input.ProjectTitle
		if projectTitle == "" {
			projectTitle = "Inbox"
		}

		projects, err := client.GetProjects(ctx)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
			}, ListTasksOutput{}, fmt.Errorf("failed to list projects: %w", err)
		}

		found := false
		for _, p := range projects {
			if p.Title == projectTitle {
				project = &Project{
					ID:    p.ID,
					Title: p.Title,
					URI:   fmt.Sprintf("vikunja://project/%d", p.ID),
				}
				targetProjectID = p.ID
				found = true
				break
			}
		}
		if !found {
			return &mcp.CallToolResult{
				IsError: true,
			}, ListTasksOutput{}, fmt.Errorf("project with title %q not found", projectTitle)
		}
	}

	var targetViewID int64
	var targetViewTitle string
	views, err := client.GetProjectViews(ctx, targetProjectID)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
		}, ListTasksOutput{}, fmt.Errorf("failed to get project views: %w", err)
	}

	if input.ViewID != "" {
		targetViewID, err = strconv.ParseInt(input.ViewID, 10, 64)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
			}, ListTasksOutput{}, fmt.Errorf("invalid view_id: %s", input.ViewID)
		}
		for _, v := range views {
			if v.ID == targetViewID {
				targetViewTitle = v.Title
				break
			}
		}
	} else {
		// Default view title to "Kanban" if not specified
		viewTitle := input.ViewTitle
		if viewTitle == "" {
			viewTitle = "Kanban"
		}

		found := false
		for _, v := range views {
			if v.Title == viewTitle {
				targetViewID = v.ID
				targetViewTitle = v.Title
				found = true
				break
			}
		}
		if !found {
			return &mcp.CallToolResult{
				IsError: true,
			}, ListTasksOutput{}, fmt.Errorf("view with title %q not found in project %d", viewTitle, targetProjectID)
		}
	}

	viewTasksResp, err := client.GetViewTasks(ctx, targetProjectID, targetViewID)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
		}, ListTasksOutput{}, fmt.Errorf("failed to get view tasks: %w", err)
	}

	vt := ViewTasksSummary{
		ViewID:    targetViewID,
		ViewTitle: targetViewTitle,
		Buckets:   make([]BucketTasksSummary, 0),
	}

	if len(viewTasksResp.Buckets) > 0 {
		for _, b := range viewTasksResp.Buckets {
			vt.Buckets = append(vt.Buckets, BucketTasksSummary{
				Bucket: toBucketSummary(b),
				Tasks:  toTasksSummary(b.Tasks),
			})
		}
	} else {
		vt.Buckets = append(vt.Buckets, BucketTasksSummary{
			Bucket: BucketSummary{ID: 0, Title: "All Tasks"},
			Tasks:  toTasksSummary(viewTasksResp.Tasks),
		})
	}

	// Convert handlers.ViewTasksSummary to vikunja.ViewTasksSummary for formatting
	vikunjaVT := vikunja.ViewTasksSummary{
		ViewID:    vt.ViewID,
		ViewTitle: vt.ViewTitle,
		Buckets:   make([]vikunja.BucketTasksSummary, len(vt.Buckets)),
	}
	for i, bucket := range vt.Buckets {
		vikunjaVT.Buckets[i] = vikunja.BucketTasksSummary{
			Bucket: vikunja.BucketSummary{
				ID:    bucket.Bucket.ID,
				Title: bucket.Bucket.Title,
			},
			Tasks: make([]vikunja.TaskSummary, len(bucket.Tasks)),
		}
		for j, task := range bucket.Tasks {
			vikunjaVT.Buckets[i].Tasks[j] = vikunja.TaskSummary{
				ID:    task.ID,
				Title: task.Title,
			}
		}
	}

	data, err := outputFormatter.Format(vikunjaVT)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
		}, ListTasksOutput{}, fmt.Errorf("failed to format response: %w", err)
	}

	return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(data)},
			},
		}, ListTasksOutput{
			View:    vt,
			Project: project,
		}, nil
}

func getTaskHandler(ctx context.Context, _ *mcp.CallToolRequest, input GetTaskInput) (*mcp.CallToolResult, GetTaskOutput, error) {
	host := os.Getenv("VIKUNJA_HOST")
	token := os.Getenv("VIKUNJA_TOKEN")
	if host == "" || token == "" {
		return nil, GetTaskOutput{}, fmt.Errorf("VIKUNJA_HOST and VIKUNJA_TOKEN environment variables required")
	}

	client, err := vikunja.NewClient(host, token, false)
	if err != nil {
		return nil, GetTaskOutput{}, fmt.Errorf("failed to create client: %w", err)
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
	if input.IncludeBuckets != false { // Default to true, allow explicit false
		// Reuse the logic from GetTaskBuckets but with the already fetched task
		views, err := client.GetProjectViews(ctx, task.ProjectID)
		if err != nil {
			fmt.Printf("Warning: failed to get project views for task %d: %v\n", taskID, err)
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

	data, err := outputFormatter.Format(vikunjaOutput)
	if err != nil {
		return nil, GetTaskOutput{}, fmt.Errorf("failed to format response: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(data)},
		},
	}, output, nil
}

func listBucketsHandler(ctx context.Context, _ *mcp.CallToolRequest, input ListBucketsInput) (*mcp.CallToolResult, ListBucketsOutput, error) {
	host := os.Getenv("VIKUNJA_HOST")
	token := os.Getenv("VIKUNJA_TOKEN")
	if host == "" || token == "" {
		return nil, ListBucketsOutput{}, fmt.Errorf("VIKUNJA_HOST and VIKUNJA_TOKEN environment variables required")
	}

	client, err := vikunja.NewClient(host, token, false)
	if err != nil {
		return nil, ListBucketsOutput{}, fmt.Errorf("failed to create client: %w", err)
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

	data, err := outputFormatter.Format(buckets)
	if err != nil {
		return nil, ListBucketsOutput{}, fmt.Errorf("failed to format response: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(data)},
		},
	}, ListBucketsOutput{Buckets: toBuckets(buckets)}, nil
}

func listProjectsHandler(ctx context.Context, _ *mcp.CallToolRequest, _ ListProjectsInput) (*mcp.CallToolResult, ListProjectsOutput, error) {
	host := os.Getenv("VIKUNJA_HOST")
	token := os.Getenv("VIKUNJA_TOKEN")
	if host == "" || token == "" {
		return nil, ListProjectsOutput{}, fmt.Errorf("VIKUNJA_HOST and VIKUNJA_TOKEN environment variables required")
	}

	client, err := vikunja.NewClient(host, token, false)
	if err != nil {
		return nil, ListProjectsOutput{}, fmt.Errorf("failed to create client: %w", err)
	}

	projects, err := client.GetProjects(ctx)
	if err != nil {
		return nil, ListProjectsOutput{}, fmt.Errorf("failed to list projects: %w", err)
	}

	output := ListProjectsOutput{
		Projects: make([]Project, len(projects)),
	}

	for i, p := range projects {
		output.Projects[i] = Project{
			ID:    p.ID,
			Title: p.Title,
			URI:   fmt.Sprintf("vikunja://projects/%d", p.ID),
		}
	}

	data, err := outputFormatter.Format(output.Projects)
	if err != nil {
		return nil, ListProjectsOutput{}, fmt.Errorf("failed to format response: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(data)},
		},
	}, output, nil
}

func createTaskHandler(ctx context.Context, _ *mcp.CallToolRequest, _ CreateTaskInput) (*mcp.CallToolResult, CreateTaskOutput, error) {
	return nil, CreateTaskOutput{}, fmt.Errorf("create task not implemented in Phase 1 (read-only operations only)")
}

func findProjectByNameHandler(ctx context.Context, _ *mcp.CallToolRequest, input FindProjectByNameInput) (*mcp.CallToolResult, FindProjectByNameOutput, error) {
	host := os.Getenv("VIKUNJA_HOST")
	token := os.Getenv("VIKUNJA_TOKEN")
	if host == "" || token == "" {
		return nil, FindProjectByNameOutput{}, fmt.Errorf("VIKUNJA_HOST and VIKUNJA_TOKEN environment variables required")
	}

	client, err := vikunja.NewClient(host, token, false)
	if err != nil {
		return nil, FindProjectByNameOutput{}, fmt.Errorf("failed to create client: %w", err)
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
		return &mcp.CallToolResult{
			IsError: true,
		}, FindProjectByNameOutput{}, fmt.Errorf("project with title %q not found", input.Name)
	}

	data, err := outputFormatter.Format(project)
	if err != nil {
		return nil, FindProjectByNameOutput{}, fmt.Errorf("failed to format response: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(data)},
		},
	}, FindProjectByNameOutput{Project: project}, nil
}

func findViewHandler(ctx context.Context, _ *mcp.CallToolRequest, input FindViewInput) (*mcp.CallToolResult, FindViewOutput, error) {
	client, err := createVikunjaClient()
	if err != nil {
		return nil, FindViewOutput{}, fmt.Errorf("failed to create client: %w", err)
	}

	project, err := findProjectByIDOrTitle(ctx, client, input.ProjectID, input.ProjectTitle)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
		}, FindViewOutput{}, err
	}

	views, err := client.GetProjectViews(ctx, project.ID)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
		}, FindViewOutput{}, fmt.Errorf("failed to get project views: %w", err)
	}

	foundView, err := findViewByName(views, input.ViewName, input.Fuzzy, project.Title)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
		}, FindViewOutput{}, fmt.Errorf("%v in project %q", err, project.Title)
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

	data, err := outputFormatter.Format(vikunjaOutput)
	if err != nil {
		return nil, FindViewOutput{}, fmt.Errorf("failed to format response: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(data)},
		},
	}, output, nil
}

func listViewsHandler(ctx context.Context, _ *mcp.CallToolRequest, input ListViewsInput) (*mcp.CallToolResult, ListViewsOutput, error) {
	client, err := createVikunjaClient()
	if err != nil {
		return nil, ListViewsOutput{}, fmt.Errorf("failed to create client: %w", err)
	}

	project, err := findProjectByIDOrTitle(ctx, client, input.ProjectID, input.ProjectTitle)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
		}, ListViewsOutput{}, err
	}

	views, err := client.GetProjectViews(ctx, project.ID)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
		}, ListViewsOutput{}, fmt.Errorf("failed to get project views: %w", err)
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

	data, err := outputFormatter.Format(vikunjaOutput)
	if err != nil {
		return nil, ListViewsOutput{}, fmt.Errorf("failed to format response: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(data)},
		},
	}, output, nil
}

// enhancedProjectNotFoundError provides contextual error message with available options
func enhancedProjectNotFoundError(title string, availableProjects []string) error {
	var suggestion string
	if len(availableProjects) > 0 {
		if len(availableProjects) <= 3 {
			suggestion = fmt.Sprintf(" Available projects: %v", availableProjects)
		} else {
			suggestion = fmt.Sprintf(" Available projects include: %s, %s, and %d others",
				availableProjects[0], availableProjects[1], len(availableProjects)-2)
		}
	}
	return fmt.Errorf("project with title %q not found.%s Try: discover_vikunja() to see all options", title, suggestion)
}

// enhancedViewNotFoundError provides contextual error message with available options
func enhancedViewNotFoundError(viewName string, projectName string, availableViews []string) error {
	var suggestion string
	if len(availableViews) > 0 {
		if len(availableViews) <= 3 {
			suggestion = fmt.Sprintf(" Available views in project '%s': %v", projectName, availableViews)
		} else {
			suggestion = fmt.Sprintf(" Available views in project '%s' include: %s, %s, and %d others",
				projectName, availableViews[0], availableViews[1], len(availableViews)-2)
		}
	}
	return fmt.Errorf("view with title %q not found in project %q.%s Try: list_views() to see project views", viewName, projectName, suggestion)
}

// parseTime parses a time string or returns zero time
func parseTime(timeStr string) time.Time {
	if timeStr == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

// toVikunjaBuckets converts handlers.Bucket to vikunja.Bucket
func toVikunjaBuckets(buckets []Bucket) []vikunja.Bucket {
	if buckets == nil {
		return nil
	}
	result := make([]vikunja.Bucket, len(buckets))
	for i, b := range buckets {
		result[i] = vikunja.Bucket{
			ID:            b.ID,
			ProjectViewID: b.ProjectViewID,
			Title:         b.Title,
			Description:   b.Description,
			Limit:         b.Limit,
			Position:      b.Position,
			IsDoneBucket:  b.IsDoneBucket,
			// Note: Tasks field is not included as it's not part of handlers.Bucket
		}
	}
	return result
}

func containsIgnoreCase(s, substr string) bool {
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)
	return strings.Contains(s, substr)
}
