package vikunja

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

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
	defer func() {
		_ = resp.Body.Close()
	}()

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
	return getSingleResource[Project](ctx, c, url, "project")
}
