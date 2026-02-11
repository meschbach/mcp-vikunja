package vikunja

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

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
	defer func() { _ = resp.Body.Close() }()

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
	defer func() { _ = resp.Body.Close() }()

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
	defer func() { _ = resp.Body.Close() }()

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
