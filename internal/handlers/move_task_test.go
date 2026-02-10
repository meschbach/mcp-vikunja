package handlers

import (
	"context"
	"os"
	"testing"

	"github.com/meschbach/mcp-vikunja/internal/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMoveTaskToBucketHandler_ReadOnlyMode(t *testing.T) {
	// Setup readonly config
	cfg := &config.Config{Readonly: true}
	deps := &HandlerDependencies{Config: cfg}
	handlers := NewHandlers(deps)

	input := MoveTaskToBucketInput{
		TaskID:    "123",
		ProjectID: "1",
		ViewID:    "2",
		BucketID:  "3",
	}

	result, output, err := handlers.moveTaskToBucketHandler(context.Background(), &mcp.CallToolRequest{}, input)

	// Should return error in readonly mode
	require.Error(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "Operation not available in readonly mode")
	assert.Equal(t, MoveTaskToBucketOutput{}, output)
}

func TestMoveTaskToBucketHandler_InvalidTaskID(t *testing.T) {
	// Setup test environment variables to avoid panics in createVikunjaClient
	os.Setenv("VIKUNJA_HOST", "test.example.com")
	os.Setenv("VIKUNJA_TOKEN", "test-token")
	defer func() {
		os.Unsetenv("VIKUNJA_HOST")
		os.Unsetenv("VIKUNJA_TOKEN")
	}()

	// Setup non-readonly config
	cfg := &config.Config{Readonly: false}
	deps := &HandlerDependencies{Config: cfg}
	handlers := NewHandlers(deps)

	input := MoveTaskToBucketInput{
		TaskID:    "invalid",
		ProjectID: "1",
		ViewID:    "2",
		BucketID:  "3",
	}

	result, output, err := handlers.moveTaskToBucketHandler(context.Background(), &mcp.CallToolRequest{}, input)

	// Should return error for invalid task ID during validation
	require.Error(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsError)

	// Safely check content
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
			assert.Contains(t, textContent.Text, "Invalid task_id")
		}
	}
	assert.Equal(t, MoveTaskToBucketOutput{}, output)
}

func TestIsReadonly(t *testing.T) {
	// Test with nil config
	deps := &HandlerDependencies{Config: nil}
	handlers := NewHandlers(deps)
	assert.False(t, handlers.isReadonly(), "Should return false when config is nil")

	// Test with readonly config
	cfg := &config.Config{Readonly: true}
	deps = &HandlerDependencies{Config: cfg}
	handlers = NewHandlers(deps)
	assert.True(t, handlers.isReadonly(), "Should return true when config is readonly")

	// Test with non-readonly config
	cfg = &config.Config{Readonly: false}
	deps = &HandlerDependencies{Config: cfg}
	handlers = NewHandlers(deps)
	assert.False(t, handlers.isReadonly(), "Should return false when config is not readonly")
}
