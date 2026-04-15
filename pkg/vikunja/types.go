package vikunja

import (
	"github.com/meschbach/vikunja-client-go/models"
)

// Project represents a Vikunja project.
type Project = models.ModelsProject

// Task represents a Vikunja task.
type Task = models.ModelsTask

// ViewKind represents the type of view for a project.
type ViewKind = string

// View kind constants.
const (
	ViewKindList   ViewKind = "list"
	ViewKindKanban ViewKind = "kanban"
	ViewKindGantt  ViewKind = "gantt"
	ViewKindTable  ViewKind = "table"
)

// BucketConfigurationMode represents how buckets are configured in a view.
type BucketConfigurationMode = string

// BucketConfigurationMode constants.
const (
	BucketConfigurationModeNone   BucketConfigurationMode = "none"
	BucketConfigurationModeManual BucketConfigurationMode = "manual"
	BucketConfigurationModeFilter BucketConfigurationMode = "filter"
)

// ProjectView represents a view within a Vikunja project.
type ProjectView = models.ModelsProjectView

// Bucket represents a bucket within a Vikunja project view.
type Bucket = models.ModelsBucket

// TaskBucket represents the association between a task and a bucket.
type TaskBucket = models.ModelsTaskBucket

// TaskViewInfo provides details about a task's position within a specific view.
type TaskViewInfo struct {
	ViewID       int64    `json:"view_id"`
	ViewTitle    string   `json:"view_title"`
	ViewKind     ViewKind `json:"view_kind"`
	BucketID     *int64   `json:"bucket_id,omitempty"`
	BucketTitle  *string  `json:"bucket_title,omitempty"`
	Position     float64  `json:"position"`
	IsDoneBucket bool     `json:"is_done_bucket"`
}

// TaskBucketInfo provides bucket information for a task across all views.
type TaskBucketInfo struct {
	TaskID int64          `json:"task_id"`
	Views  []TaskViewInfo `json:"views"`
}

// BucketTasks represents a bucket and its associated tasks.
type BucketTasks struct {
	Bucket Bucket  `json:"bucket"`
	Tasks  []*Task `json:"tasks"`
}

// ViewTasks represents tasks organized by buckets within a view.
type ViewTasks struct {
	ViewID    int64         `json:"view_id"`
	ViewTitle string        `json:"view_title"`
	Buckets   []BucketTasks `json:"buckets"`
}

// ViewTasksResponse represents the API response for view tasks.
type ViewTasksResponse struct {
	Buckets []*Bucket `json:"buckets,omitempty"`
	Tasks   []*Task   `json:"tasks,omitempty"`
}

// TaskOutput represents a task with its associated bucket information.
type TaskOutput struct {
	Task    Task            `json:"task"`
	Buckets *TaskBucketInfo `json:"buckets,omitempty"`
}

// ViewOutput represents a project with a single view.
type ViewOutput struct {
	Project Project     `json:"project"`
	View    ProjectView `json:"view"`
}

// ViewsOutput represents a project with all its views.
type ViewsOutput struct {
	Project Project        `json:"project"`
	Views   []*ProjectView `json:"views"`
}

// TaskSummary provides a minimal representation of a task.
type TaskSummary struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	URI   string `json:"uri"`
}

// BucketSummary provides a minimal representation of a bucket.
type BucketSummary struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

// BucketTasksSummary represents a bucket with its task summaries.
type BucketTasksSummary struct {
	Bucket BucketSummary `json:"bucket"`
	Tasks  []TaskSummary `json:"tasks,omitempty"`
}

// ViewTasksSummary provides a summarized view of tasks organized by buckets.
type ViewTasksSummary struct {
	ViewID    int64                `json:"view_id"`
	ViewTitle string               `json:"view_title"`
	Buckets   []BucketTasksSummary `json:"buckets,omitempty"`
}
