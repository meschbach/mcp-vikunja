// Package handlers provides MCP tool handlers for Vikunja integration.
package handlers

import (
	"github.com/meschbach/mcp-vikunja/pkg/vikunja"
)

// Input/Output types for handlers

type ListTasksInput struct {
	ProjectID    string `json:"project_id,omitempty" jsonschema:"Optional project ID to filter tasks"`
	ProjectTitle string `json:"project_title,omitempty" jsonschema:"Optional project title (defaults to 'Inbox')"`
	ViewID       string `json:"view_id,omitempty" jsonschema:"Optional project view ID"`
	ViewTitle    string `json:"view_title,omitempty" jsonschema:"Optional project view title (defaults to 'Kanban')"`
	BucketID     string `json:"bucket_id,omitempty" jsonschema:"Optional bucket ID to filter tasks"`
	BucketTitle  string `json:"bucket_title,omitempty" jsonschema:"Optional bucket title to filter tasks"`
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

type ListTasksOutput struct {
	Project *Project         `json:"project,omitempty" jsonschema:"Project hte tasks are related to"`
	View    ViewTasksSummary `json:"view" jsonschema:"tasks associated with this view"`
}

type GetTaskInput struct {
	TaskID         string `json:"task_id" jsonschema:"The ID of task to retrieve"`
	IncludeBuckets bool   `json:"include_buckets,omitempty" jsonschema:"Whether to include bucket information across all project views (default: true)"`
}

type GetTaskOutput struct {
	Task    Task                    `json:"task"`
	Buckets *vikunja.TaskBucketInfo `json:"buckets,omitempty"`
}

type ListBucketsInput struct {
	ProjectID string `json:"project_id" jsonschema:"The ID of project"`
	ViewID    string `json:"view_id" jsonschema:"The ID of project view"`
}

type ListBucketsOutput struct {
	Buckets []Bucket `json:"buckets"`
}

type ListProjectsInput struct {
}

type ListProjectsOutput struct {
	Projects []vikunja.Project `json:"projects"`
}

type CreateTaskInput struct {
	Title       string `json:"title" jsonschema:"The title of task"`
	Description string `json:"description,omitempty" jsonschema:"Optional task description"`
	ProjectID   string `json:"project_id" jsonschema:"The project ID to create task in"`
}

type CreateTaskOutput struct {
	Task Task `json:"task"`
}

type FindProjectByNameInput struct {
	Name string `json:"name" jsonschema:"The name/title of project to find"`
}

type FindProjectByNameOutput struct {
	Project Project `json:"project"`
}

type FindViewInput struct {
	ProjectID    string `json:"project_id,omitempty" jsonschema:"Optional project ID to search in (overrides project_title)"`
	ProjectTitle string `json:"project_title,omitempty" jsonschema:"Optional project title to search in"`
	ViewName     string `json:"view_name" jsonschema:"The name/title of view to find"`
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

type MoveTaskToBucketInput struct {
	TaskID    string `json:"task_id" jsonschema:"The ID of task to move"`
	ProjectID string `json:"project_id" jsonschema:"The project ID containing task"`
	ViewID    string `json:"view_id" jsonschema:"The view ID containing task"`
	BucketID  string `json:"bucket_id" jsonschema:"The bucket ID to move task to"`
}

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
