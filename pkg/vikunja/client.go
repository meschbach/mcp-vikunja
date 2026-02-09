package vikunja

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client wraps the Vikunja API client
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient creates a new Vikunja API client
func NewClient(host, token string, insecure bool) (*Client, error) {
	scheme := "https"
	if insecure {
		scheme = "http"
	}

	baseURL := fmt.Sprintf("%s://%s/api/v1", scheme, host)

	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// GetProjects retrieves all projects
func (c *Client) GetProjects(ctx context.Context) ([]Project, error) {
	url := fmt.Sprintf("%s/projects", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch projects: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var projects []Project
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return projects, nil
}

// GetProject retrieves a single project by ID
func (c *Client) GetProject(ctx context.Context, id int64) (*Project, error) {
	url := fmt.Sprintf("%s/projects/%d", c.baseURL, id)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch project: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var project Project
	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &project, nil
}

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
	defer resp.Body.Close()

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
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch task: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var task Task
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &task, nil
}

// GetProjectViews retrieves all views for a project
func (c *Client) GetProjectViews(ctx context.Context, projectID int64) ([]ProjectView, error) {
	url := fmt.Sprintf("%s/projects/%d/views", c.baseURL, projectID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch project views: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var views []ProjectView
	if err := json.Unmarshal(bodyBytes, &views); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return views, nil
}

// GetViewBuckets retrieves all buckets for a specific view
func (c *Client) GetViewBuckets(ctx context.Context, projectID, viewID int64) ([]Bucket, error) {
	url := fmt.Sprintf("%s/projects/%d/views/%d/buckets", c.baseURL, projectID, viewID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch view buckets: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var buckets []Bucket
	if err := json.Unmarshal(bodyBytes, &buckets); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return buckets, nil
}

// GetViewTasks retrieves tasks for a specific view.
// For kanban views, the endpoint returns buckets with their tasks.
// For other views, it returns a flat list of tasks.
func (c *Client) GetViewTasks(ctx context.Context, projectID, viewID int64) (*ViewTasksResponse, error) {
	url := fmt.Sprintf("%s/projects/%d/views/%d/tasks", c.baseURL, projectID, viewID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch view tasks: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Inspect the payload to decide if it's buckets-with-tasks or flat tasks
	var raw any
	if err := json.Unmarshal(bodyBytes, &raw); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	respData := &ViewTasksResponse{}

	slice, ok := raw.([]any)
	if !ok {
		// Unexpected, but try to unmarshal as tasks and return
		var tasks []Task
		if err := json.Unmarshal(bodyBytes, &tasks); err != nil {
			return nil, fmt.Errorf("unexpected response format: %w", err)
		}
		respData.Tasks = tasks
		return respData, nil
	}

	isBucketShape := false
	for _, el := range slice {
		m, ok := el.(map[string]any)
		if !ok {
			continue
		}
		if _, hasTasks := m["tasks"]; hasTasks {
			isBucketShape = true
			break
		}
	}

	if isBucketShape {
		var buckets []Bucket
		if err := json.Unmarshal(bodyBytes, &buckets); err != nil {
			return nil, fmt.Errorf("failed to decode buckets response: %w", err)
		}
		respData.Buckets = buckets
		return respData, nil
	}

	// Fallback: flat tasks
	var tasks []Task
	if err := json.Unmarshal(bodyBytes, &tasks); err != nil {
		return nil, fmt.Errorf("failed to decode tasks response: %w", err)
	}
	respData.Tasks = tasks
	return respData, nil
}

// GetTaskBuckets retrieves bucket information for a task across all views
func (c *Client) GetTaskBuckets(ctx context.Context, taskID int64) (*TaskBucketInfo, error) {
	// First get the task to determine its project and get buckets via expansion
	task, err := c.GetTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	if len(task.Buckets) == 0 {
		// logger.Warn("Vikunja API for task did not return bucket information with ?expand=buckets", "taskID", taskID)
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

// handleErrorResponse processes error responses from the API
func (c *Client) handleErrorResponse(resp *http.Response) error {
	var errResp ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return fmt.Errorf("API error (status %d): failed to decode error response", resp.StatusCode)
	}
	return fmt.Errorf("API error (status %d): %s", resp.StatusCode, errResp.Message)
}
