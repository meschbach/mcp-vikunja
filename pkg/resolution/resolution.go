// Package resolution provides functions for resolving Vikunja projects and buckets by ID or title.
// It is used by both the CLI and MCP handlers to avoid code duplication.
//
// Error cases documented:
//   - ResolveProject: returns errors like "project identifier is required", "invalid project_id",
//     "project with title ... not found", "multiple projects found with title ... please use project ID".
//   - FindBucketByIDOrTitle: returns errors like "bucket ... not found in Kanban view of project ...",
//     "multiple buckets found with title ... please use bucket ID", or "kanban view not found in project ...".
//   - FindKanbanView: returns "kanban view not found in project ...".
package resolution

import (
	"context"
	"fmt"
	"strconv"

	"github.com/meschbach/mcp-vikunja/pkg/vikunja"
)

// Client is the subset of vikunja.Client methods required for resolution.
type Client interface {
	GetProjects(ctx context.Context) ([]*vikunja.Project, error)
	GetProjectViews(ctx context.Context, projectID int64) ([]*vikunja.ProjectView, error)
	GetViewBuckets(ctx context.Context, projectID int64, viewID int64) ([]*vikunja.Bucket, error)
	GetProject(ctx context.Context, projectID int64) (*vikunja.Project, error)
}

// Project is a minimal representation of a Vikunja project, containing only ID and Title.
type Project struct {
	ID    int64
	Title string
}

// ResolveProject resolves a project by an identifier which can be either a numeric ID or a title.
// Numeric IDs are parsed first; if parsing fails, the identifier is treated as a title.
// For numeric identifiers, a Project is returned without verifying existence (to avoid extra API calls).
// For title identifiers, the project is looked up via GetProjects and must match exactly one.
func ResolveProject(ctx context.Context, client Client, identifier string) (*Project, error) {
	if identifier == "" {
		return nil, fmt.Errorf("project identifier is required")
	}
	// Try as numeric ID first
	if _, err := strconv.ParseInt(identifier, 10, 64); err == nil {
		return findProjectByIDOrTitle(ctx, client, identifier, "")
	}
	// Otherwise treat as title
	return findProjectByIDOrTitle(ctx, client, "", identifier)
}

// findProjectByIDOrTitle finds a project either by numeric ID or by title.
// If projectID is non-empty, it must be a valid integer; returns a Project with a default title.
// If projectTitle is non-empty, it searches all projects for an exact title match.
func findProjectByIDOrTitle(ctx context.Context, client Client, projectID, projectTitle string) (*Project, error) {
	if projectID != "" {
		id, err := strconv.ParseInt(projectID, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid project_id: %s", projectID)
		}
		return &Project{
			ID:    id,
			Title: fmt.Sprintf("Project %d", id),
		}, nil
	}

	if projectTitle == "" {
		return nil, fmt.Errorf("either projectID or projectTitle must be specified")
	}

	return findProjectByTitle(ctx, client, projectTitle)
}

func findProjectByTitle(ctx context.Context, client Client, title string) (*Project, error) {
	projects, err := client.GetProjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	var matches []Project
	for _, p := range projects {
		if p.Title == title {
			matches = append(matches, Project{
				ID:    p.ID,
				Title: p.Title,
			})
		}
	}

	if len(matches) == 0 {
		var projectTitles []string
		for _, p := range projects {
			projectTitles = append(projectTitles, p.Title)
		}
		return nil, enhancedProjectNotFoundError(title, projectTitles)
	}
	if len(matches) > 1 {
		return nil, fmt.Errorf("multiple projects found with title %q, please use project ID", title)
	}
	return &matches[0], nil
}

// FindKanbanView finds the Kanban (bucket) view for a given project.
// Returns an error if the project has no Kanban view.
func FindKanbanView(ctx context.Context, client Client, projectID int64) (*vikunja.ProjectView, error) {
	views, err := client.GetProjectViews(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project views: %w", err)
	}

	for _, view := range views {
		if view.ViewKind == string(vikunja.ViewKindKanban) {
			return view, nil
		}
	}

	return nil, fmt.Errorf("kanban view not found in project %d. Project must have a Kanban view to use buckets", projectID)
}

// FindBucketByIDOrTitle finds a bucket by ID (if numeric) or title (via Kanban view).
// If bucketInput is numeric, returns that ID without validation.
// If bucketInput is non-numeric, it must match exactly one bucket title in the project's Kanban view.
func FindBucketByIDOrTitle(ctx context.Context, client Client, projectID int64, bucketInput string) (*int64, error) {
	if id, err := strconv.ParseInt(bucketInput, 10, 64); err == nil && id > 0 {
		return &id, nil
	}

	bucket, err := findBucketByTitle(ctx, client, projectID, bucketInput)
	if err != nil {
		return nil, err
	}
	return &bucket.ID, nil
}

func findBucketByTitle(ctx context.Context, client Client, projectID int64, title string) (*vikunja.Bucket, error) {
	kanbanView, err := FindKanbanView(ctx, client, projectID)
	if err != nil {
		return nil, err
	}

	buckets, err := client.GetViewBuckets(ctx, projectID, kanbanView.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get buckets for kanban view: %w", err)
	}

	var matches []*vikunja.Bucket
	for _, b := range buckets {
		if b.Title == title {
			matches = append(matches, b)
		}
	}

	if len(matches) == 0 {
		bucketNames := make([]string, len(buckets))
		for i, b := range buckets {
			bucketNames[i] = b.Title
		}
		projTitle := getProjectTitle(ctx, client, projectID)
		return nil, enhancedBucketNotFoundError(title, projTitle, bucketNames)
	}
	if len(matches) > 1 {
		return nil, fmt.Errorf("multiple buckets found with title %q in Kanban view, please use bucket ID", title)
	}

	return matches[0], nil
}

func getProjectTitle(ctx context.Context, client Client, projectID int64) string {
	proj, err := client.GetProject(ctx, projectID)
	if err != nil {
		return fmt.Sprintf("Project %d", projectID)
	}
	return proj.Title
}

// enhancedProjectNotFoundError provides contextual error message with available options.
func enhancedProjectNotFoundError(title string, availableProjects []string) error {
	var suggestion string
	if len(availableProjects) > 0 {
		if len(availableProjects) <= 3 {
			suggestion = fmt.Sprintf(" Available projects: %v", availableProjects)
		} else {
			suggestion = fmt.Sprintf(" Available projects include: %s, %s, and %d others",
				availableProjects[0], availableProjects[1], len(availableProjects)-2)
		}
	}
	return fmt.Errorf("project with title %q not found.%s Try: list_projects() to see all available projects", title, suggestion)
}

// enhancedBucketNotFoundError provides contextual error message with available buckets in the Kanban view.
func enhancedBucketNotFoundError(bucket, projectTitle string, availableBuckets []string) error {
	var suggestion string
	if len(availableBuckets) > 0 {
		if len(availableBuckets) <= 3 {
			suggestion = fmt.Sprintf(" Available buckets in project '%s': %v", projectTitle, availableBuckets)
		} else {
			suggestion = fmt.Sprintf(" Available buckets in project '%s' include: %s, %s, and %d others",
				projectTitle, availableBuckets[0], availableBuckets[1], len(availableBuckets)-2)
		}
	}
	return fmt.Errorf("bucket %q not found in Kanban view of project %q.%s Try: list_buckets(project_title=%q) to see available buckets",
		bucket, projectTitle, suggestion, projectTitle)
}
