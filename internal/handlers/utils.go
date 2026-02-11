// Package handlers provides MCP tool handlers for Vikunja integration.
package handlers

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/meschbach/mcp-vikunja/pkg/vikunja"
)

// Conversion functions

func toTaskSummary(t vikunja.Task) TaskSummary {
	return TaskSummary{
		ID:    t.ID,
		Title: t.Title,
		URI:   fmt.Sprintf("vikunja://task/%d", t.ID),
	}
}

func toTasksSummary(tasks []vikunja.Task) []TaskSummary {
	if tasks == nil {
		return nil
	}
	res := make([]TaskSummary, len(tasks))
	for i, t := range tasks {
		res[i] = toTaskSummary(t)
	}
	return res
}

func toBucketSummary(b vikunja.Bucket) BucketSummary {
	return BucketSummary{
		ID:    b.ID,
		Title: b.Title,
	}
}

func toTask(t vikunja.Task) Task {
	dueDate := ""
	if !t.DueDate.IsZero() {
		dueDate = t.DueDate.Format(time.RFC3339)
	}
	created := ""
	if !t.Created.IsZero() {
		created = t.Created.Format(time.RFC3339)
	}
	updated := ""
	if !t.Updated.IsZero() {
		updated = t.Updated.Format(time.RFC3339)
	}

	return Task{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		ProjectID:   t.ProjectID,
		Done:        t.Done,
		DueDate:     dueDate,
		Created:     created,
		Updated:     updated,
		Buckets:     toBuckets(t.Buckets),
		Position:    t.Position,
	}
}

func toBucket(b vikunja.Bucket) Bucket {
	return Bucket{
		ID:            b.ID,
		ProjectViewID: b.ProjectViewID,
		Title:         b.Title,
		Description:   b.Description,
		Limit:         b.Limit,
		Position:      b.Position,
		IsDoneBucket:  b.IsDoneBucket,
	}
}

func toBuckets(buckets []vikunja.Bucket) []Bucket {
	if buckets == nil {
		return nil
	}
	res := make([]Bucket, len(buckets))
	for i, b := range buckets {
		res[i] = toBucket(b)
	}
	return res
}

func toView(v vikunja.ProjectView) View {
	return View{
		ID:                      v.ID,
		ProjectID:               v.ProjectID,
		Title:                   v.Title,
		ViewKind:                v.ViewKind,
		Position:                v.Position,
		BucketConfigurationMode: v.BucketConfigurationMode,
		DefaultBucketID:         v.DefaultBucketID,
		DoneBucketID:            v.DoneBucketID,
		URI:                     fmt.Sprintf("vikunja://project/%d/view/%d", v.ProjectID, v.ID),
	}
}

func toViews(views []vikunja.ProjectView) []View {
	if views == nil {
		return nil
	}
	res := make([]View, len(views))
	for i, v := range views {
		res[i] = toView(v)
	}
	return res
}

// toVikunjaBuckets converts handlers.Bucket to vikunja.Bucket
func toVikunjaBuckets(buckets []Bucket) []vikunja.Bucket {
	if buckets == nil {
		return nil
	}
	result := make([]vikunja.Bucket, len(buckets))
	for i, b := range buckets {
		result[i] = vikunja.Bucket{
			ID:            b.ID,
			ProjectViewID: b.ProjectViewID,
			Title:         b.Title,
			Description:   b.Description,
			Limit:         b.Limit,
			Position:      b.Position,
			IsDoneBucket:  b.IsDoneBucket,
			// Note: Tasks field is not included as it's not part of handlers.Bucket
		}
	}
	return result
}

// Client creation and utility functions

func createVikunjaClient() (*vikunja.Client, error) {
	host := os.Getenv("VIKUNJA_HOST")
	token := os.Getenv("VIKUNJA_TOKEN")
	if host == "" || token == "" {
		return nil, fmt.Errorf("VIKUNJA_HOST and VIKUNJA_TOKEN environment variables required")
	}

	insecure := os.Getenv("VIKUNJA_INSECURE") == "true"
	return vikunja.NewClient(host, token, insecure)
}

// findProjectByIDOrTitle finds a project by ID or title
func findProjectByIDOrTitle(ctx context.Context, client *vikunja.Client, projectID, projectTitle string) (*Project, error) {
	if projectID != "" {
		id, err := strconv.ParseInt(projectID, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid project_id: %s", projectID)
		}
		return &Project{
			ID:    id,
			Title: fmt.Sprintf("Project %d", id),
			URI:   fmt.Sprintf("vikunja://project/%d", id),
		}, nil
	}

	if projectTitle == "" {
		return nil, fmt.Errorf("either project_id or project_title must be specified")
	}

	projects, err := client.GetProjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	var projectTitles []string
	for _, p := range projects {
		projectTitles = append(projectTitles, p.Title)
		if p.Title == projectTitle {
			return &Project{
				ID:    p.ID,
				Title: p.Title,
				URI:   fmt.Sprintf("vikunja://project/%d", p.ID),
			}, nil
		}
	}

	return nil, enhancedProjectNotFoundError(projectTitle, projectTitles)
}

// findViewByName finds a view by name within a project's views
func findViewByName(views []vikunja.ProjectView, viewName string, fuzzy bool, projectName string) (*vikunja.ProjectView, error) {
	var viewTitles []string
	for _, v := range views {
		viewTitles = append(viewTitles, v.Title)
		if fuzzy && containsIgnoreCase(v.Title, viewName) {
			return &v, nil
		}
		if !fuzzy && v.Title == viewName {
			return &v, nil
		}
	}

	return nil, enhancedViewNotFoundError(viewName, projectName, viewTitles)
}

// enhancedProjectNotFoundError provides contextual error message with available options
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

// enhancedViewNotFoundError provides contextual error message with available options
func enhancedViewNotFoundError(viewName string, projectName string, availableViews []string) error {
	var suggestion string
	if len(availableViews) > 0 {
		if len(availableViews) <= 3 {
			suggestion = fmt.Sprintf(" Available views in project '%s': %v", projectName, availableViews)
		} else {
			suggestion = fmt.Sprintf(" Available views in project '%s' include: %s, %s, and %d others",
				projectName, availableViews[0], availableViews[1], len(availableViews)-2)
		}
	}
	return fmt.Errorf("view with title %q not found in project %q.%s Try: list_views() to see project views", viewName, projectName, suggestion)
}

// parseTime parses a time string or returns zero time
func parseTime(timeStr string) time.Time {
	if timeStr == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

// containsIgnoreCase checks if a string contains a substring (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)
	return strings.Contains(s, substr)
}

// ValidationError represents a validation error with field name and message
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// validateRequiredString checks if a required string field is non-empty
func validateRequiredString(fieldName, value string) error {
	if value == "" {
		return ValidationError{Field: fieldName, Message: "is required"}
	}
	return nil
}

// parseID parses a string ID and validates it's a positive integer
func parseID(fieldName, value string) (int64, error) {
	if value == "" {
		return 0, ValidationError{Field: fieldName, Message: "is required"}
	}
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, ValidationError{Field: fieldName, Message: fmt.Sprintf("must be a valid integer, got: %s", value)}
	}
	if id <= 0 {
		return 0, ValidationError{Field: fieldName, Message: fmt.Sprintf("must be a positive integer, got: %d", id)}
	}
	return id, nil
}

// validateViewKind checks if a view kind is valid
func validateViewKind(kind string) error {
	if kind == "" {
		return nil // Optional field
	}
	validKinds := map[string]bool{
		"list":   true,
		"kanban": true,
		"gantt":  true,
		"table":  true,
	}
	if !validKinds[kind] {
		return ValidationError{Field: "view_kind", Message: fmt.Sprintf("must be one of: list, kanban, gantt, table. Got: %s", kind)}
	}
	return nil
}
