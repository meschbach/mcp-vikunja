// Package handlers provides MCP tool handlers for Vikunja integration.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/meschbach/mcp-vikunja/pkg/vikunja"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

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
	Buckets   []BucketTasksSummary `json:"buckets,omitempty"`
}

type ListTasksOutput struct {
	Project *Project         `json:"project,omitempty"`
	View    ViewTasksSummary `json:"view"`
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
	return Task{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		ProjectID:   t.ProjectID,
		Done:        t.Done,
		DueDate:     t.DueDate.String(),
		Created:     t.Created.String(),
		Updated:     t.Updated.String(),
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

// Register adds all Vikunja tool handlers to the MCP server.
func Register(s *mcp.Server) {
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

	data, err := json.MarshalIndent(vt, "", "")
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
		}, ListTasksOutput{}, fmt.Errorf("failed to marshal response: %w", err)
	}

	return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(data)},
				&mcp.TextContent{Text: fmt.Sprintf("Results for project %s and board %s", project.Title, targetViewTitle)},
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

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return nil, GetTaskOutput{}, fmt.Errorf("failed to marshal response: %w", err)
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

	data, err := json.MarshalIndent(buckets, "", "  ")
	if err != nil {
		return nil, ListBucketsOutput{}, fmt.Errorf("failed to marshal response: %w", err)
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

	data, err := json.MarshalIndent(output.Projects, "", "  ")
	if err != nil {
		return nil, ListProjectsOutput{}, fmt.Errorf("failed to marshal response: %w", err)
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
