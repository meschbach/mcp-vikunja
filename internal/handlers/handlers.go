// Package handlers provides MCP tool handlers for Vikunja integration.
package handlers

import (
	"log/slog"

	"github.com/meschbach/mcp-vikunja/internal/config"
	"github.com/meschbach/mcp-vikunja/pkg/vikunja"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// HandlerDependencies holds all external dependencies for handlers
type HandlerDependencies struct {
	Client          *vikunja.Client
	OutputFormatter vikunja.OutputFormatter
	Config          *config.Config
	Logger          *slog.Logger
}

// Handlers provides all MCP tool handlers
type Handlers struct {
	deps *HandlerDependencies
}

// NewHandlers creates a new Handlers instance with dependency injection
func NewHandlers(deps *HandlerDependencies) *Handlers {
	if deps.Logger == nil {
		deps.Logger = slog.Default()
	}
	return &Handlers{deps: deps}
}

// TODO: These will be replaced with proper handler methods after file splitting
// For now, we need to import the handlers from split files

// Register adds all Vikunja tool handlers to MCP server.
func Register(s *mcp.Server, cfg *config.Config) {
	// Initialize dependencies
	deps := &HandlerDependencies{
		Config:          cfg,
		OutputFormatter: vikunja.GetFormatter(cfg.OutputFormat),
		Logger:          slog.Default(),
	}

	handlers := NewHandlers(deps)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_tasks",
		Description: "List tasks from Vikunja filtering by criteria. Use 'project', 'view', and 'bucket' parameters with either ID (integer) or title (string). Defaults: project=Inbox, view=Kanban",
	}, handlers.listTasksHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_task",
		Description: "Get details of a specific task",
	}, handlers.getTaskHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_buckets",
		Description: "List all buckets in a project view",
	}, handlers.listBucketsHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_projects",
		Description: "List all projects via this Vikunja connection.   Provides a list of projects including ID, name, and URI",
	}, handlers.listProjectsHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_task",
		Description: "Create a new task in Vikunja",
	}, handlers.createTaskHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "find_project_by_name",
		Description: "Find a project by its name/title",
	}, handlers.findProjectByNameHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "find_view",
		Description: "Find a specific view by name within a project",
	}, handlers.findViewHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_views",
		Description: "List all views for a project, optionally filtered by view kind",
	}, handlers.listViewsHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "move_task_to_bucket",
		Description: "Move a task to a different bucket within a project view",
	}, handlers.moveTaskToBucketHandler)
}

// isReadonly returns true if server is in readonly mode
func (h *Handlers) isReadonly() bool {
	if h.deps.Config != nil {
		return h.deps.Config.Readonly
	}
	return false
}
