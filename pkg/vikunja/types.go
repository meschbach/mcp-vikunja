package vikunja

import (
	"time"
)

// Project represents a Vikunja project
type Project struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Identifier  string    `json:"identifier,omitempty"`
	OwnerID     int64     `json:"owner_id"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
	Position    float64   `json:"position"`
}

// Task represents a Vikunja task
type Task struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	ProjectID   int64     `json:"project_id,omitempty"`
	Done        bool      `json:"done"`
	DueDate     time.Time `json:"due_date,omitempty"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
	Buckets     []Bucket  `json:"buckets,omitempty"`
	Position    float64   `json:"position"`
}

// ViewKind represents the type of project view
type ViewKind string

const (
	ViewKindList   ViewKind = "list"
	ViewKindKanban ViewKind = "kanban"
	ViewKindGantt  ViewKind = "gantt"
	ViewKindTable  ViewKind = "table"
)

// BucketConfigurationMode represents how buckets are configured in a view
type BucketConfigurationMode string

const (
	BucketConfigurationModeNone   BucketConfigurationMode = "none"
	BucketConfigurationModeManual BucketConfigurationMode = "manual"
	BucketConfigurationModeFilter BucketConfigurationMode = "filter"
)

// ProjectViewFilter represents the structure of the 'filter' object in a ProjectView
type ProjectViewFilter struct {
	S                  string  `json:"s"`
	SortBy             *string `json:"sort_by,omitempty"`
	OrderBy            *string `json:"order_by,omitempty"`
	Filter             string  `json:"filter"`
	FilterIncludeNulls bool    `json:"filter_include_nulls"`
}

// ProjectView represents a Vikunja project view
type ProjectView struct {
	ID                      int64                   `json:"id"`
	ProjectID               int64                   `json:"project_id"`
	Title                   string                  `json:"title"`
	ViewKind                ViewKind                `json:"view_kind"`
	Position                float64                 `json:"position"`
	Filter                  *ProjectViewFilter      `json:"filter,omitempty"`
	BucketConfigurationMode BucketConfigurationMode `json:"bucket_configuration_mode"`
	DefaultBucketID         int64                   `json:"default_bucket_id,omitempty"`
	DoneBucketID            int64                   `json:"done_bucket_id,omitempty"`
}

// Bucket represents a Vikunja bucket (column in Kanban view)
type Bucket struct {
	ID            int64   `json:"id"`
	ProjectViewID int64   `json:"project_view_id"`
	Title         string  `json:"title"`
	Description   string  `json:"description,omitempty"`
	Limit         int     `json:"limit"`
	Position      float64 `json:"position"`
	IsDoneBucket  bool    `json:"is_done_bucket"`
	// Tasks is only present when fetching tasks via the view tasks endpoint for kanban views
	Tasks []Task `json:"tasks,omitempty"`
}

// TaskBucket represents the relationship between a task and a bucket
type TaskBucket struct {
	TaskID   int64 `json:"task_id"`
	BucketID int64 `json:"bucket_id"`
}

// TaskViewInfo contains information about a task's position in a specific view
type TaskViewInfo struct {
	ViewID       int64    `json:"view_id"`
	ViewTitle    string   `json:"view_title"`
	ViewKind     ViewKind `json:"view_kind"`
	BucketID     *int64   `json:"bucket_id,omitempty"`
	BucketTitle  *string  `json:"bucket_title,omitempty"`
	Position     float64  `json:"position"`
	IsDoneBucket bool     `json:"is_done_bucket"`
}

// TaskBucketInfo contains bucket information for a task across all views
type TaskBucketInfo struct {
	TaskID int64          `json:"task_id"`
	Views  []TaskViewInfo `json:"views"`
}

// BucketTasks represents a bucket and its associated tasks
type BucketTasks struct {
	Bucket Bucket `json:"bucket"`
	Tasks  []Task `json:"tasks"`
}

// ViewTasks represents all buckets and tasks in a view
type ViewTasks struct {
	ViewID    int64         `json:"view_id"`
	ViewTitle string        `json:"view_title"`
	Buckets   []BucketTasks `json:"buckets"`
}

// ViewTasksResponse represents the polymorphic response of /projects/{id}/views/{view}/tasks
// For kanban views, the API returns buckets with their tasks.
// For other views, it returns a flat list of tasks.
// Exactly one of Buckets or Tasks will be non-empty.
type ViewTasksResponse struct {
	Buckets []Bucket `json:"buckets,omitempty"`
	Tasks   []Task   `json:"tasks,omitempty"`
}

// ErrorResponse represents a Vikunja API error response
type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}
