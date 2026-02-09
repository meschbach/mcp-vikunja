package handlers

import (
	"encoding/json"
	"testing"
)

func TestDiscoverVikunjaHandler(t *testing.T) {
	// This test would require a real Vikunja instance to work fully
	// For now, we test the structure compilation and basic flow
	t.Run("discover_input_structure", func(t *testing.T) {
		input := DiscoverInput{
			MaxProjects:   3,
			IncludeCounts: false,
		}

		// Test that input marshals correctly
		data, err := json.Marshal(input)
		if err != nil {
			t.Fatalf("Failed to marshal input: %v", err)
		}

		// Basic structure validation
		var unmarshaled DiscoverInput
		err = json.Unmarshal(data, &unmarshaled)
		if err != nil {
			t.Fatalf("Failed to unmarshal input: %v", err)
		}

		if unmarshaled.MaxProjects != 3 {
			t.Errorf("Expected MaxProjects=3, got %d", unmarshaled.MaxProjects)
		}
	})

	t.Run("enhanced_error_messages", func(t *testing.T) {
		// Test enhanced project error message
		err := enhancedProjectNotFoundError("NonExistent", []string{"Inbox", "Work", "Personal"})

		if !contains(err.Error(), "project with title \"NonExistent\" not found") {
			t.Errorf("Expected error to contain project not found message\nGot: %s", err.Error())
		}
		if !contains(err.Error(), "Available projects") {
			t.Errorf("Expected error to contain available projects suggestion\nGot: %s", err.Error())
		}
		if !contains(err.Error(), "discover_vikunja() to see all options") {
			t.Errorf("Expected error to contain discover suggestion\nGot: %s", err.Error())
		}

		// Test enhanced view error message
		viewErr := enhancedViewNotFoundError("NonExistent", "Inbox", []string{"Kanban", "List", "Gantt"})

		if !contains(viewErr.Error(), "view with title \"NonExistent\" not found in project \"Inbox\"") {
			t.Errorf("Expected error to contain view not found message\nGot: %s", viewErr.Error())
		}
		if !contains(viewErr.Error(), "Available views in project 'Inbox'") {
			t.Errorf("Expected error to contain available views suggestion\nGot: %s", viewErr.Error())
		}
		if !contains(viewErr.Error(), "list_views() to see project views") {
			t.Errorf("Expected error to contain list_views suggestion\nGot: %s", viewErr.Error())
		}
	})

	t.Run("output_structure_validation", func(t *testing.T) {
		// Validate that the output structure is complete and correctly typed
		projects := []ProjectFlat{
			{ID: 1, Title: "Test Project", DefaultViewID: 1, ViewCount: 2},
		}

		views := []ViewFlat{
			{ID: 1, ProjectID: 1, Title: "Kanban", ViewKind: "kanban", IsDefault: true},
		}

		toolGuide := ToolGuide{
			Tools: map[string]ToolInfo{
				"list_tasks": {
					Purpose: "get_project_tasks",
					Parameters: map[string]ParamInfo{
						"project_id": {
							Name:   "project_id",
							Type:   "int64",
							Source: "projects.id",
						},
					},
					ExampleCalls: []string{"list_tasks()", "list_tasks(project_id=1)"},
				},
			},
		}

		output := DiscoverOutput{
			Projects: projects,
			Views:    views,
			Tools:    toolGuide,
			ServerInfo: ServerInfo{
				APIVersion: "v1",
				Status:     "connected",
				Features:   []string{"list_tasks", "get_task", "list_projects", "readonly"},
			},
			SchemaInfo: SchemaInfo{
				Version:        "1.0",
				RequiredFields: []string{"projects", "tools"},
				FieldSources:   map[string]string{"project_id": "projects.id"},
			},
		}

		// Test that output marshals correctly
		data, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal output: %v", err)
		}

		// Basic validation that key fields are present
		if len(output.Projects) == 0 {
			t.Error("Expected at least one project")
		}

		if len(output.Tools.Tools) == 0 {
			t.Error("Expected at least one tool in guide")
		}

		if output.ServerInfo.APIVersion == "" {
			t.Error("Expected server info to have API version")
		}

		// Validate that the JSON contains expected structure markers
		jsonStr := string(data)
		if !contains(jsonStr, "projects") {
			t.Error("Output should contain 'projects' field")
		}
		if !contains(jsonStr, "quick_start") {
			t.Error("Output should contain 'quick_start' field")
		}
		if !contains(jsonStr, "error_prevention") {
			t.Error("Output should contain 'error_prevention' field")
		}
	})
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(len(s) >= len(substr) &&
			(s == substr || len(s) > len(substr) &&
				(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
					indexOf(s, substr) >= 0)))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
