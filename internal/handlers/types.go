package handlers

import (
	"github.com/meschbach/mcp-vikunja/pkg/vikunja"
)

// Input/Output types for handlers

// ListTasksInput defines input for listing tasks.
type ListTasksInput struct {
	Project string `json:"project,omitempty" jsonschema:"Optional project ID (integer) or title (string). Defaults to 'Inbox'"`
	View    string `json:"view,omitempty" jsonschema:"Optional view ID (integer) or title (string). Defaults to 'Kanban'"`
	Bucket  string `json:"bucket,omitempty" jsonschema:"Optional bucket ID (integer) or title (string)"`
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
	Buckets   []BucketTasksSummary `json:"buckets,omitempty" jsonschema:"Buckets tasks are organized into"`
}

// ListTasksOutput defines output for listing tasks.
type ListTasksOutput struct {
	Project *Project         `json:"project,omitempty" jsonschema:"Project the tasks are related to"`
	View    ViewTasksSummary `json:"view" jsonschema:"tasks associated with this view"`
}

// GetTaskInput defines input for retrieving a task.
type GetTaskInput struct {
	TaskID         string `json:"task_id" jsonschema:"The ID of task to retrieve"`
	IncludeBuckets bool   `json:"include_buckets,omitempty" jsonschema:"Whether to include bucket information across all project views (default: true)"`
}

// GetTaskOutput defines output for retrieving a task.
type GetTaskOutput struct {
	Task    Task                    `json:"task"`
	Buckets *vikunja.TaskBucketInfo `json:"buckets,omitempty"`
}

// ListBucketsInput defines input for listing buckets.
type ListBucketsInput struct {
	ProjectTitle string `json:"project_title,omitempty" jsonschema:"Optional project title to list buckets for (defaults to 'Inbox')"`
	ViewTitle    string `json:"view_title,omitempty" jsonschema:"Optional view title to list buckets for (defaults to 'Kanban')"`
}

// ListBucketsOutput defines output for listing buckets.
type ListBucketsOutput struct {
	Buckets []Bucket `json:"buckets"`
}

// ListProjectsInput defines input for listing projects.
type ListProjectsInput struct {
}

// ListProjectsOutput defines output for listing projects.
type ListProjectsOutput struct {
	Projects []*vikunja.Project `json:"projects"`
}

// CreateTaskInput defines input for creating a task.
type CreateTaskInput struct {
	Title       string `json:"title" jsonschema:"The title of task"`
	Description string `json:"description,omitempty" jsonschema:"Optional task description"`
	ProjectID   string `json:"project_id" jsonschema:"Project ID (numeric) or project title to create task in"`
	BucketID    string `json:"bucket_id,omitempty" jsonschema:"Optional bucket ID (numeric) or bucket title to assign task to. Bucket must be in the project's Kanban view."`
}

// CreateTaskOutput defines output for creating a task.
type CreateTaskOutput struct {
	Task Task `json:"task"`
}

// FindProjectByNameInput defines input for finding a project by name.
type FindProjectByNameInput struct {
	Name string `json:"name" jsonschema:"The name/title of project to find"`
}

// FindProjectByNameOutput defines output for finding a project by name.
type FindProjectByNameOutput struct {
	Project Project `json:"project"`
}

// FindViewInput defines input for finding a view.
type FindViewInput struct {
	ProjectID    string `json:"project_id,omitempty" jsonschema:"Optional project ID to search in (overrides project_title)"`
	ProjectTitle string `json:"project_title,omitempty" jsonschema:"Optional project title to search in"`
	ViewName     string `json:"view_name" jsonschema:"The name/title of view to find"`
	Fuzzy        bool   `json:"fuzzy,omitempty" jsonschema:"Enable fuzzy/partial matching for view names (default: false)"`
}

// FindViewOutput defines output for finding a view.
type FindViewOutput struct {
	Project Project `json:"project"`
	View    View    `json:"view"`
}

// ListViewsInput defines input for listing views.
type ListViewsInput struct {
	ProjectID    string `json:"project_id,omitempty" jsonschema:"Optional project ID to list views for (overrides project_title)"`
	ProjectTitle string `json:"project_title,omitempty" jsonschema:"Optional project title to list views for"`
	ViewKind     string `json:"view_kind,omitempty" jsonschema:"Optional filter by view kind (list, kanban, gantt, table)"`
}

// ListViewsOutput defines output for listing views.
type ListViewsOutput struct {
	Project Project `json:"project"`
	Views   []View  `json:"views"`
}

// MoveTaskToBucketInput defines input for moving a task to a bucket.
type MoveTaskToBucketInput struct {
	TaskID    string `json:"task_id" jsonschema:"The ID of task to move"`
	ProjectID string `json:"project_id" jsonschema:"The project ID containing task"`
	ViewID    string `json:"view_id" jsonschema:"The view ID containing task"`
	BucketID  string `json:"bucket_id" jsonschema:"The bucket ID to move task to"`
}

// MoveTaskToBucketOutput defines output for moving a task to a bucket.
type MoveTaskToBucketOutput struct {
	TaskBucket vikunja.TaskBucket `json:"task_bucket"`
	Message    string             `json:"message"`
}

// Core types

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
	Position      float64 `json:"position"`
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
