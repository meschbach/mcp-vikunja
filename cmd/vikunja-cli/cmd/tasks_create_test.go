package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"log/slog"

	"github.com/meschbach/mcp-vikunja/pkg/vikunja"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	origClient       *vikunja.Client
	origOutputWriter io.Writer
	origLogger       *slog.Logger
	origJSONFmt      bool
	origMarkdown     bool
	origNoColor      bool
	origInsecure     bool
)

func resetGlobals() {
	client = origClient
	outputWriter = origOutputWriter
	logger = origLogger
	jsonFmt = origJSONFmt
	markdown = origMarkdown
	noColor = origNoColor
	insecure = origInsecure
	// Reset command-specific flags to avoid leakage between tests
	tasksCreateProjectFlag = ""
	tasksCreateBucketFlag = ""
}

func TestTasksCreateCmd_Structure(t *testing.T) {
	cmd := tasksCreateCmd
	assert.Contains(t, cmd.Use, "create")
	assert.Equal(t, "Create a new task", cmd.Short)
	assert.Equal(t, "Create a new Vikunja task with optional description and project/bucket assignment.", cmd.Long)
	assert.NotNil(t, cmd.Flags().Lookup("project"))
	assert.NotNil(t, cmd.Flags().Lookup("bucket"))
}

func TestTasksCreateCmd_RunE_SuccessDefaultInbox(t *testing.T) {
	// Save globals
	origClient = client
	origOutputWriter = outputWriter
	origLogger = logger
	origJSONFmt = jsonFmt
	origMarkdown = markdown
	origNoColor = noColor
	origInsecure = insecure
	defer resetGlobals()

	// Setup output capture
	buf := &bytes.Buffer{}
	outputWriter = buf
	logger = slog.New(slog.NewTextHandler(io.Discard, nil))

	var createTaskReq struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		ProjectID   int64  `json:"project_id"`
		BucketID    *int64 `json:"bucket_id,omitempty"`
	}
	createTaskCalled := false

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects":
			fmt.Fprint(w, `[{"id":1,"title":"Inbox"}]`)
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/projects/1/tasks":
			createTaskCalled = true
			if err := json.NewDecoder(r.Body).Decode(&createTaskReq); err != nil {
				t.Fatalf("decode create task request: %v", err)
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":          123,
				"title":       "My Task",
				"description": "",
				"project_id":  1,
				"done":        false,
				"due_date":    nil,
				"created":     "2024-01-01T00:00:00Z",
				"updated":     "2024-01-01T00:00:00Z",
				"position":    0,
				"bucket_ids":  []int64{},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/tasks/123":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":          123,
				"title":       "My Task",
				"description": "",
				"project_id":  1,
				"done":        false,
				"due_date":    nil,
				"created":     "2024-01-01T00:00:00Z",
				"updated":     "2024-01-01T00:00:00Z",
				"position":    0,
			})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/1/views":
			json.NewEncoder(w).Encode([]interface{}{
				map[string]interface{}{
					"id":                        1,
					"project_id":                1,
					"title":                     "Kanban",
					"view_kind":                 "kanban",
					"position":                  0,
					"bucket_configuration_mode": "none",
					"default_bucket_id":         0,
					"done_bucket_id":            0,
				},
			})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()
	host := ts.URL[7:] // strip http://

	// Prepare root command with host and token
	args := []string{
		"--host", host,
		"--token", "dummy",
		"--insecure",
		"tasks", "create", "My Task",
	}
	insecure = true
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "My Task")
	assert.Contains(t, output, "123")
	assert.True(t, createTaskCalled, "CreateTask should be called")
	assert.Equal(t, int64(1), createTaskReq.ProjectID)
	assert.Nil(t, createTaskReq.BucketID)
}

func TestTasksCreateCmd_RunE_SuccessWithNumericProjectAndBucket(t *testing.T) {
	origClient = client
	origOutputWriter = outputWriter
	origLogger = logger
	origJSONFmt = jsonFmt
	origMarkdown = markdown
	origNoColor = noColor
	origInsecure = insecure
	defer resetGlobals()

	buf := &bytes.Buffer{}
	outputWriter = buf
	logger = slog.New(slog.NewTextHandler(io.Discard, nil))

	var createTaskReq struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		ProjectID   int64  `json:"project_id"`
		BucketID    *int64 `json:"bucket_id,omitempty"`
	}
	createTaskCalled := false

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/projects/42/tasks":
			createTaskCalled = true
			if err := json.NewDecoder(r.Body).Decode(&createTaskReq); err != nil {
				t.Fatalf("decode create task request: %v", err)
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":          456,
				"title":       "Task with bucket",
				"description": "desc",
				"project_id":  42,
				"done":        false,
				"due_date":    nil,
				"created":     "2024-01-01T00:00:00Z",
				"updated":     "2024-01-01T00:00:00Z",
				"position":    0,
				"bucket_ids":  []int64{99},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/tasks/456":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":          456,
				"title":       "Task with bucket",
				"description": "desc",
				"project_id":  42,
				"done":        false,
				"due_date":    nil,
				"created":     "2024-01-01T00:00:00Z",
				"updated":     "2024-01-01T00:00:00Z",
				"position":    0,
			})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/42/views":
			json.NewEncoder(w).Encode([]interface{}{
				map[string]interface{}{
					"id":                        5,
					"project_id":                42,
					"title":                     "Kanban",
					"view_kind":                 "kanban",
					"position":                  0,
					"bucket_configuration_mode": "none",
					"default_bucket_id":         0,
					"done_bucket_id":            0,
				},
			})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()
	host := ts.URL[7:]

	args := []string{
		"--host", host,
		"--token", "dummy",
		"--insecure",
		"tasks", "create", "Task with bucket", "desc",
		"--project", "42",
		"--bucket", "99",
	}
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Task with bucket")
	assert.Contains(t, output, "456")
	assert.True(t, createTaskCalled)
	assert.Equal(t, int64(42), createTaskReq.ProjectID)
	assert.Equal(t, int64(99), *createTaskReq.BucketID)
}

func TestTasksCreateCmd_RunE_SuccessWithTitleProjectAndBucket(t *testing.T) {
	origClient = client
	origOutputWriter = outputWriter
	origLogger = logger
	origJSONFmt = jsonFmt
	origMarkdown = markdown
	origNoColor = noColor
	origInsecure = insecure
	defer resetGlobals()

	buf := &bytes.Buffer{}
	outputWriter = buf
	logger = slog.New(slog.NewTextHandler(io.Discard, nil))

	var createTaskReq struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		ProjectID   int64  `json:"project_id"`
		BucketID    *int64 `json:"bucket_id,omitempty"`
	}
	createTaskCalled := false

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects":
			fmt.Fprint(w, `[{"id":2,"title":"Work"}]`)
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/2/views":
			fmt.Fprint(w, `[{"id":5,"project_id":2,"title":"Kanban","view_kind":"kanban","position":0,"default_bucket_id":0,"done_bucket_id":0}]`)
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/2/views/5/buckets":
			fmt.Fprint(w, `[{"id":10,"project_view_id":5,"title":"Todo","description":"","position":0,"is_done_bucket":false},{"id":11,"project_view_id":5,"title":"In Progress","description":"","position":1,"is_done_bucket":false}]`)
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/projects/2/tasks":
			createTaskCalled = true
			if err := json.NewDecoder(r.Body).Decode(&createTaskReq); err != nil {
				t.Fatalf("decode create task request: %v", err)
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":          789,
				"title":       "Task with title bucket",
				"description": "desc2",
				"project_id":  2,
				"done":        false,
				"due_date":    nil,
				"created":     "2024-01-01T00:00:00Z",
				"updated":     "2024-01-01T00:00:00Z",
				"position":    0,
				"bucket_ids":  []int64{11},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/tasks/789":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":          789,
				"title":       "Task with title bucket",
				"description": "desc2",
				"project_id":  2,
				"done":        false,
				"due_date":    nil,
				"created":     "2024-01-01T00:00:00Z",
				"updated":     "2024-01-01T00:00:00Z",
				"position":    0,
			})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()
	host := ts.URL[7:]

	args := []string{
		"--host", host,
		"--token", "dummy",
		"--insecure",
		"tasks", "create", "Task with title bucket", "desc2",
		"--project", "Work",
		"--bucket", "In Progress",
	}
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Task with title bucket")
	assert.Contains(t, output, "789")
	assert.True(t, createTaskCalled)
	assert.Equal(t, int64(2), createTaskReq.ProjectID)
	assert.Equal(t, int64(11), *createTaskReq.BucketID)
}

func TestTasksCreateCmd_RunE_ErrorMissingTitle(t *testing.T) {
	// Save globals
	origClient = client
	origOutputWriter = outputWriter
	origLogger = logger
	origJSONFmt = jsonFmt
	origMarkdown = markdown
	origNoColor = noColor
	defer resetGlobals()

	buf := &bytes.Buffer{}
	outputWriter = buf
	logger = slog.New(slog.NewTextHandler(io.Discard, nil))

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))
	defer ts.Close()
	host := ts.URL[7:]

	args := []string{
		"--host", host,
		"--token", "dummy",
		"--insecure",
		"tasks", "create",
		"--bucket", "Inbox",
	}
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	// Cobra should return an error about needing at least 1 arg
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requires at least 1 arg")
}

func TestTasksCreateCmd_RunE_ErrorProjectNotFound(t *testing.T) {
	origClient = client
	origOutputWriter = outputWriter
	origLogger = logger
	origJSONFmt = jsonFmt
	origMarkdown = markdown
	origNoColor = noColor
	origInsecure = insecure
	defer resetGlobals()

	buf := &bytes.Buffer{}
	outputWriter = buf
	logger = slog.New(slog.NewTextHandler(io.Discard, nil))

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects" {
			fmt.Fprint(w, `[]`) // no projects
		} else {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()
	host := ts.URL[7:]

	args := []string{
		"--host", host,
		"--token", "dummy",
		"--insecure",
		"tasks", "create", "My Task",
	}
	insecure = true
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), `project with title "Inbox" not found`)
}

func TestTasksCreateCmd_RunE_ErrorBucketNotFound(t *testing.T) {
	origClient = client
	origOutputWriter = outputWriter
	origLogger = logger
	origJSONFmt = jsonFmt
	origMarkdown = markdown
	origNoColor = noColor
	origInsecure = insecure
	defer resetGlobals()

	buf := &bytes.Buffer{}
	outputWriter = buf
	logger = slog.New(slog.NewTextHandler(io.Discard, nil))

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects":
			fmt.Fprint(w, `[{"id":1,"title":"Inbox"}]`)
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/1":
			fmt.Fprint(w, `{"id":1,"title":"Inbox"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/1/views":
			fmt.Fprint(w, `[{"id":5,"project_id":1,"title":"Kanban","view_kind":"kanban","position":0}]`)
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/1/views/5/buckets":
			// Buckets: only "Todo" exists, request asks for "Inbox" implicitly later
			fmt.Fprint(w, `[{"id":10,"project_view_id":5,"title":"Todo","description":"","position":0,"is_done_bucket":false}]`)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()
	host := ts.URL[7:]
	// Simulate --bucket Inbox by setting the flag variable directly
	tasksCreateBucketFlag = "Inbox"
	args := []string{
		"--host", host,
		"--token", "dummy",
		"--insecure",
		"tasks", "create", "Table Task",
	}
	insecure = true
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	t.Logf("err.Error() = %q", err.Error())
	require.Error(t, err)
	assert.Contains(t, err.Error(), `bucket "Inbox" not found`)
}

func TestTasksCreateCmd_RunE_ErrorMissingKanbanView(t *testing.T) {
	origClient = client
	origOutputWriter = outputWriter
	origLogger = logger
	origJSONFmt = jsonFmt
	origMarkdown = markdown
	origNoColor = noColor
	origInsecure = insecure
	defer resetGlobals()

	buf := &bytes.Buffer{}
	outputWriter = buf
	logger = slog.New(slog.NewTextHandler(io.Discard, nil))

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects":
			fmt.Fprint(w, `[{"id":1,"title":"Inbox"}]`)
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/1/views":
			// No kanban view, only list view
			fmt.Fprint(w, `[{"id":2,"project_id":1,"title":"List","view_kind":"list","position":0}]`)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()
	host := ts.URL[7:]

	args := []string{
		"--host", host,
		"--token", "dummy",
		"--insecure",
		"tasks", "create", "My Task",
		"--bucket", "SomeBucket", // requires kanban view
	}
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "kanban view not found in project 1")
}

func TestTasksCreateCmd_RunE_OutputJSON(t *testing.T) {
	origClient = client
	origOutputWriter = outputWriter
	origLogger = logger
	origJSONFmt = jsonFmt
	origMarkdown = markdown
	origNoColor = noColor
	origInsecure = insecure
	defer resetGlobals()

	buf := &bytes.Buffer{}
	outputWriter = buf
	logger = slog.New(slog.NewTextHandler(io.Discard, nil))

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects":
			fmt.Fprint(w, `[{"id":1,"title":"Inbox"}]`)
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/projects/1/tasks":
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":          100,
				"title":       "JSON Task",
				"description": "json desc",
				"project_id":  1,
				"done":        false,
				"due_date":    nil,
				"created":     "2024-01-01T00:00:00Z",
				"updated":     "2024-01-01T00:00:00Z",
				"position":    0,
				"bucket_ids":  []int64{},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/tasks/100":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":          100,
				"title":       "JSON Task",
				"description": "json desc",
				"project_id":  1,
				"done":        false,
				"due_date":    nil,
				"created":     "2024-01-01T00:00:00Z",
				"updated":     "2024-01-01T00:00:00Z",
				"position":    0,
				"buckets":     []interface{}{},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/1/views":
			json.NewEncoder(w).Encode([]interface{}{
				map[string]interface{}{
					"id":                        1,
					"project_id":                1,
					"title":                     "Kanban",
					"view_kind":                 "kanban",
					"position":                  0,
					"bucket_configuration_mode": "none",
					"default_bucket_id":         0,
					"done_bucket_id":            0,
				},
			})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()
	host := ts.URL[7:]

	args := []string{
		"--host", host,
		"--token", "dummy",
		"--insecure",
		"--json",
		"tasks", "create", "JSON Task", "json desc",
	}
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	// JSON output includes task and buckets when bucket info is available
	var result struct {
		Task    map[string]interface{} `json:"task"`
		Buckets interface{}            `json:"buckets,omitempty"`
	}
	require.NoError(t, json.Unmarshal([]byte(output), &result))
	task := result.Task
	require.NotNil(t, task, "expected task object in JSON output")
	assert.Equal(t, float64(100), task["id"])
	assert.Equal(t, "JSON Task", task["title"])
}

func TestTasksCreateCmd_RunE_OutputMarkdown(t *testing.T) {
	origClient = client
	origOutputWriter = outputWriter
	origLogger = logger
	origJSONFmt = jsonFmt
	origMarkdown = markdown
	origNoColor = noColor
	origInsecure = insecure
	defer resetGlobals()

	buf := &bytes.Buffer{}
	outputWriter = buf
	logger = slog.New(slog.NewTextHandler(io.Discard, nil))

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects":
			fmt.Fprint(w, `[{"id":1,"title":"Inbox"}]`)
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/projects/1/tasks":
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":          200,
				"title":       "Markdown Task",
				"description": "md desc",
				"project_id":  1,
				"done":        false,
				"due_date":    nil,
				"created":     "2024-01-01T00:00:00Z",
				"updated":     "2024-01-01T00:00:00Z",
				"position":    0,
				"bucket_ids":  []int64{},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/tasks/200":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":          200,
				"title":       "Markdown Task",
				"description": "md desc",
				"project_id":  1,
				"done":        false,
				"due_date":    nil,
				"created":     "2024-01-01T00:00:00Z",
				"updated":     "2024-01-01T00:00:00Z",
				"position":    0,
				"buckets":     []interface{}{},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/1/views":
			json.NewEncoder(w).Encode([]interface{}{
				map[string]interface{}{
					"id":                        1,
					"project_id":                1,
					"title":                     "Kanban",
					"view_kind":                 "kanban",
					"position":                  0,
					"bucket_configuration_mode": "none",
					"default_bucket_id":         0,
					"done_bucket_id":            0,
				},
			})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()
	host := ts.URL[7:]

	args := []string{
		"--host", host,
		"--token", "dummy",
		"--insecure",
		"--markdown",
		"tasks", "create", "Markdown Task", "md desc",
	}
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	// Markdown should contain a title heading and details
	assert.Contains(t, output, "# Markdown Task")
	assert.Contains(t, output, "**ID**: 200")
}

func TestTasksCreateCmd_RunE_OutputTable(t *testing.T) {
	origClient = client
	origOutputWriter = outputWriter
	origLogger = logger
	origJSONFmt = jsonFmt
	origMarkdown = markdown
	origNoColor = noColor
	origInsecure = insecure
	defer resetGlobals()

	buf := &bytes.Buffer{}
	outputWriter = buf
	logger = slog.New(slog.NewTextHandler(io.Discard, nil))

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects":
			fmt.Fprint(w, `[{"id":1,"title":"Inbox"}]`)
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/projects/1/tasks":
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":          300,
				"title":       "Table Task",
				"description": "",
				"project_id":  1,
				"done":        false,
				"due_date":    nil,
				"created":     "2024-01-01T00:00:00Z",
				"updated":     "2024-01-01T00:00:00Z",
				"position":    0,
				"bucket_ids":  []int64{},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/tasks/300":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":          300,
				"title":       "Table Task",
				"description": "",
				"project_id":  1,
				"done":        false,
				"due_date":    nil,
				"created":     "2024-01-01T00:00:00Z",
				"updated":     "2024-01-01T00:00:00Z",
				"position":    0,
				"buckets":     []interface{}{},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/1/views":
			json.NewEncoder(w).Encode([]interface{}{
				map[string]interface{}{
					"id":                        1,
					"project_id":                1,
					"title":                     "Kanban",
					"view_kind":                 "kanban",
					"position":                  0,
					"bucket_configuration_mode": "none",
					"default_bucket_id":         0,
					"done_bucket_id":            0,
				},
			})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()
	host := ts.URL[7:]

	args := []string{
		"--host", host,
		"--token", "dummy",
		"--insecure",
		"tasks", "create", "Table Task",
	}
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	// Table format includes headers and tab-separated fields
	assert.Contains(t, output, "Table Task")
	assert.Contains(t, output, "300")
}
