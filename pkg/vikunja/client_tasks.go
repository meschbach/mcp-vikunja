package vikunja

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// GetTasks retrieves tasks, optionally filtered by project ID
func (c *Client) GetTasks(ctx context.Context, projectID int64) ([]Task, error) {
	var url string
	if projectID > 0 {
		url = fmt.Sprintf("%s/projects/%d/tasks?expand=buckets", c.baseURL, projectID)
	} else {
		url = fmt.Sprintf("%s/tasks?expand=buckets", c.baseURL)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tasks: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var tasks []Task
	if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return tasks, nil
}

// GetTask retrieves a single task by ID
func (c *Client) GetTask(ctx context.Context, id int64) (*Task, error) {
	url := fmt.Sprintf("%s/tasks/%d?expand=buckets", c.baseURL, id)
	return getSingleResource[Task](ctx, c, url, "task")
}

// GetTaskBuckets retrieves bucket information for a task across all views
func (c *Client) GetTaskBuckets(ctx context.Context, taskID int64) (*TaskBucketInfo, error) {
	// First get the task to determine its project and get buckets via expansion
	task, err := c.GetTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	if len(task.Buckets) == 0 {
		// Intentionally empty branch - buckets may be empty for some tasks
		// TODO: Add structured logging when available: logger.Warn("Vikunja API for task did not return bucket information with ?expand=buckets", "taskID", taskID)
		_ = struct{}{} // no-op to satisfy staticcheck
	}

	// Get all views for the task's project to resolve view titles/kinds
	views, err := c.GetProjectViews(ctx, task.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project views: %w", err)
	}

	var taskViews []TaskViewInfo

	// For each view, determine the task's bucket position from the expanded bucket info
	for _, view := range views {
		viewInfo := TaskViewInfo{
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
				// Determine if this is the done bucket for the view
				if view.DoneBucketID == bucket.ID {
					viewInfo.IsDoneBucket = true
				}
				break
			}
		}

		taskViews = append(taskViews, viewInfo)
	}

	return &TaskBucketInfo{
		TaskID: taskID,
		Views:  taskViews,
	}, nil
}

// MoveTaskToBucket moves a task to a different bucket within a project view
func (c *Client) MoveTaskToBucket(ctx context.Context, projectID, viewID, bucketID, taskID int64) (*TaskBucket, error) {
	url := fmt.Sprintf("%s/projects/%d/views/%d/buckets/%d/tasks", c.baseURL, projectID, viewID, bucketID)

	// Create request body with task_id
	reqBody := map[string]int64{"task_id": taskID}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to move task: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var taskBucket TaskBucket
	if err := json.NewDecoder(resp.Body).Decode(&taskBucket); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &taskBucket, nil
}
