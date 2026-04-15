// Package vikunja provides a client for interacting with the Vikunja API.
package vikunja

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

	"github.com/meschbach/vikunja-client-go/client/project"
	"github.com/meschbach/vikunja-client-go/client/task"
	"github.com/meschbach/vikunja-client-go/models"
)

// Client wraps the Vikunja API client for task and project operations.
type Client struct {
	transport runtime.ClientTransport
	projects  project.ClientService
	tasks     task.ClientService
	auth      runtime.ClientAuthInfoWriter
}

// NewClient creates a new Vikunja API client configured with the provided host and authentication token.
func NewClient(host, token string, insecure bool) (*Client, error) {
	scheme := "https"
	if insecure {
		scheme = "http"
	}

	if strings.HasPrefix(host, "http://") {
		scheme = "http"
		host = strings.TrimPrefix(host, "http://")
	} else if strings.HasPrefix(host, "https://") {
		scheme = "https"
		host = strings.TrimPrefix(host, "https://")
	}

	parsedURL, err := url.Parse(host)
	if err == nil && parsedURL.Host != "" {
		host = parsedURL.Host
	}

	httpTransport := httptransport.New(host, "/api/v1", []string{scheme})
	httpTransport.DefaultAuthentication = httptransport.BearerToken(token)
	httpTransport.Consumers[runtime.JSONMime] = runtime.JSONConsumer()
	httpTransport.Producers[runtime.JSONMime] = runtime.JSONProducer()

	formats := strfmt.Default

	return &Client{
		transport: httpTransport,
		projects:  project.New(httpTransport, formats),
		tasks:     task.New(httpTransport, formats),
		auth:      httptransport.BearerToken(token),
	}, nil
}

func (c *Client) httpClient() *http.Client {
	return &http.Client{Timeout: 30 * time.Second}
}

// GetTasks retrieves all tasks, optionally filtered by project ID.
func (c *Client) GetTasks(ctx context.Context, projectID int64) ([]*models.ModelsTask, error) {
	params := task.NewGetTasksParams()
	params.SetContext(ctx)
	params.SetHTTPClient(c.httpClient())

	if projectID > 0 {
		filter := fmt.Sprintf("project_id:%d", projectID)
		params.SetFilter(&filter)
	}

	result, err := c.tasks.GetTasks(params, c.auth)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	return result.Payload, nil
}

// GetTask retrieves a single task by its ID.
//
// Duplicates GetProject due to generated swagger client patterns. Each method uses
// a different resource client (tasks vs projects) with identical parameter handling.
// Refactoring would require interface gymnastics that obscure the straightforward API calls.
//
//nolint:dupl
func (c *Client) GetTask(ctx context.Context, id int64) (*models.ModelsTask, error) {
	params := task.NewGetTasksIDParams()
	params.SetContext(ctx)
	params.SetHTTPClient(c.httpClient())
	params.SetID(id)

	result, err := c.tasks.GetTasksID(params, c.auth)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return result.Payload, nil
}

// CreateTask creates a new task in the specified project.
func (c *Client) CreateTask(ctx context.Context, title string, projectID int64, description string, bucketID *int64, dueDate time.Time) (*models.ModelsTask, error) {
	taskModel := &models.ModelsTask{
		Title:     title,
		ProjectID: projectID,
	}

	if description != "" {
		taskModel.Description = description
	}

	if bucketID != nil {
		taskModel.BucketID = *bucketID
	}

	if !dueDate.IsZero() {
		taskModel.DueDate = dueDate.Format("2006-01-02")
	}

	params := task.NewPutProjectsIDTasksParams()
	params.SetContext(ctx)
	params.SetHTTPClient(c.httpClient())
	params.SetID(projectID)
	params.SetTask(taskModel)

	result, err := c.tasks.PutProjectsIDTasks(params, c.auth)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return result.Payload, nil
}

// MoveTaskToBucket moves a task to the specified bucket within a project's view.
func (c *Client) MoveTaskToBucket(ctx context.Context, projectID, viewID, bucketID, taskID int64) (*models.ModelsTaskBucket, error) {
	taskBucket := &models.ModelsTaskBucket{
		TaskID: taskID,
	}

	params := task.NewPostProjectsProjectViewsViewBucketsBucketTasksParams()
	params.SetContext(ctx)
	params.SetHTTPClient(c.httpClient())
	params.SetProject(projectID)
	params.SetView(viewID)
	params.SetBucket(bucketID)
	params.SetTaskBucket(taskBucket)

	result, err := c.tasks.PostProjectsProjectViewsViewBucketsBucketTasks(params, c.auth)
	if err != nil {
		return nil, fmt.Errorf("failed to move task to bucket: %w", err)
	}

	return result.Payload, nil
}

// GetProjects retrieves all projects.
func (c *Client) GetProjects(ctx context.Context) ([]*models.ModelsProject, error) {
	params := project.NewGetProjectsParams()
	params.SetContext(ctx)
	params.SetHTTPClient(c.httpClient())

	result, err := c.projects.GetProjects(params, c.auth)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}

	return result.Payload, nil
}

// GetProject retrieves a single project by its ID.
//
// Duplicates GetTask due to generated swagger client patterns. Each method uses
// a different resource client (projects vs tasks) with identical parameter handling.
// Refactoring would require interface gymnastics that obscure the straightforward API calls.
//
//nolint:dupl
func (c *Client) GetProject(ctx context.Context, id int64) (*models.ModelsProject, error) {
	params := project.NewGetProjectsIDParams()
	params.SetContext(ctx)
	params.SetHTTPClient(c.httpClient())
	params.SetID(id)

	result, err := c.projects.GetProjectsID(params, c.auth)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return result.Payload, nil
}

// GetProjectViews retrieves all views for the specified project.
func (c *Client) GetProjectViews(ctx context.Context, projectID int64) ([]*models.ModelsProjectView, error) {
	params := project.NewGetProjectsProjectViewsParams()
	params.SetContext(ctx)
	params.SetHTTPClient(c.httpClient())
	params.SetProject(projectID)

	result, err := c.projects.GetProjectsProjectViews(params, c.auth)
	if err != nil {
		return nil, fmt.Errorf("failed to get project views: %w", err)
	}

	return result.Payload, nil
}

// GetViewBuckets retrieves all buckets for the specified project and view.
//
// Duplicates GetViewTasks due to generated swagger client patterns. Each method uses
// a different resource client (projects vs tasks) with identical parameter handling.
// Refactoring would require interface gymnastics that obscure the straightforward API calls.
//
//nolint:dupl
func (c *Client) GetViewBuckets(ctx context.Context, projectID, viewID int64) ([]*models.ModelsBucket, error) {
	params := project.NewGetProjectsIDViewsViewBucketsParams()
	params.SetContext(ctx)
	params.SetHTTPClient(c.httpClient())
	params.SetID(projectID)
	params.SetView(viewID)

	result, err := c.projects.GetProjectsIDViewsViewBuckets(params, c.auth)
	if err != nil {
		return nil, fmt.Errorf("failed to get view buckets: %w", err)
	}

	return result.Payload, nil
}

// GetViewTasks retrieves all tasks for the specified project and view.
//
// Duplicates GetViewBuckets due to generated swagger client patterns. Each method uses
// a different resource client (tasks vs projects) with identical parameter handling.
// Refactoring would require interface gymnastics that obscure the straightforward API calls.
//
//nolint:dupl
func (c *Client) GetViewTasks(ctx context.Context, projectID, viewID int64) ([]*models.ModelsTask, error) {
	params := task.NewGetProjectsIDViewsViewTasksParams()
	params.SetContext(ctx)
	params.SetHTTPClient(c.httpClient())
	params.SetID(projectID)
	params.SetView(viewID)

	result, err := c.tasks.GetProjectsIDViewsViewTasks(params, c.auth)
	if err != nil {
		return nil, fmt.Errorf("failed to get view tasks: %w", err)
	}

	return result.Payload, nil
}
