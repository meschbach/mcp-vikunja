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

func writeTaskCreateResponse(w http.ResponseWriter, taskID int64, taskTitle, description string, projectID int64, bucketIDs []int64) {
	w.WriteHeader(http.StatusCreated)                 //nolint:errcheck,gosec
	json.NewEncoder(w).Encode(map[string]interface{}{ //nolint:errcheck,gosec
		"id":          taskID,
		"title":       taskTitle,
		"description": description,
		"project_id":  projectID,
		"done":        false,
		"due_date":    nil,
		"created":     "2024-01-01T00:00:00Z",
		"updated":     "2024-01-01T00:00:00Z",
		"position":    0,
		"bucket_ids":  bucketIDs,
	})
}

func writeTaskGetResponse(w http.ResponseWriter, taskID int64, taskTitle, description string, projectID int64) {
	json.NewEncoder(w).Encode(map[string]interface{}{ //nolint:errcheck,gosec
		"id":          taskID,
		"title":       taskTitle,
		"description": description,
		"project_id":  projectID,
		"done":        false,
		"due_date":    nil,
		"created":     "2024-01-01T00:00:00Z",
		"updated":     "2024-01-01T00:00:00Z",
		"position":    0,
		"buckets":     []interface{}{},
	})
}

type outputTestCase struct {
	taskID      int64
	taskTitle   string
	description string
	projectID   int64
	bucketIDs   []int64
	useJSON     bool
	useMarkdown bool
}

//nolint:gocyclo
func setupOutputTestServer(t *testing.T, tc *outputTestCase) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects":
			fmt.Fprint(w, `[{"id":1,"title":"Inbox"}]`) //nolint:errcheck
		case r.Method == http.MethodPut && r.URL.Path == "/api/v1/projects/1/tasks":
			writeTaskCreateResponse(w, tc.taskID, tc.taskTitle, tc.description, tc.projectID, tc.bucketIDs)
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf("/api/v1/tasks/%d", tc.taskID):
			writeTaskGetResponse(w, tc.taskID, tc.taskTitle, tc.description, tc.projectID)
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/1/views":
			fmt.Fprint(w, `[{"id":1,"project_id":1,"title":"Kanban","view_kind":"kanban","position":0,"bucket_configuration_mode":"none","default_bucket_id":0,"done_bucket_id":0}]`) //nolint:errcheck
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
}

var (
	origClient       *vikunja.Client
	origOutputWriter io.Writer
	origLogger       *slog.Logger
	origJSONFmt      bool
	origMarkdown     bool
	origNoColor      bool
	origInsecure     bool
)

func saveAndResetGlobals() {
	origClient = client
	origOutputWriter = outputWriter
	origLogger = logger
	origJSONFmt = jsonFmt
	origMarkdown = markdown
	origNoColor = noColor
	origInsecure = insecure
}

func resetGlobals() {
	client = origClient
	outputWriter = origOutputWriter
	logger = origLogger
	jsonFmt = origJSONFmt
	markdown = origMarkdown
	noColor = origNoColor
	insecure = origInsecure
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

//nolint:gocyclo
func TestTasksCreateCmd_RunE_SuccessDefaultInbox(t *testing.T) {
	saveAndResetGlobals()
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
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects":
			fmt.Fprint(w, `[{"id":1,"title":"Inbox"}]`) //nolint:errcheck
		case r.Method == http.MethodPut && r.URL.Path == "/api/v1/projects/1/tasks":
			createTaskCalled = true
			if err := json.NewDecoder(r.Body).Decode(&createTaskReq); err != nil {
				t.Fatalf("decode create task request: %v", err)
			}
			writeTaskCreateResponse(w, 123, "My Task", "", 1, []int64{})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/tasks/123":
			writeTaskGetResponse(w, 123, "My Task", "", 1)
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/1/views":
			fmt.Fprint(w, `[{"id":1,"project_id":1,"title":"Kanban","view_kind":"kanban","position":0,"bucket_configuration_mode":"none","default_bucket_id":0,"done_bucket_id":0}]`) //nolint:errcheck
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
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
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "My Task")
	assert.Contains(t, output, "123")
	assert.True(t, createTaskCalled, "CreateTask should be called")
	assert.Equal(t, int64(1), createTaskReq.ProjectID)
	assert.Nil(t, createTaskReq.BucketID)
}

func TestTasksCreateCmd_RunE_SuccessWithNumericProjectAndBucket(t *testing.T) {
	saveAndResetGlobals()
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
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPut && r.URL.Path == "/api/v1/projects/42/tasks":
			createTaskCalled = true
			if err := json.NewDecoder(r.Body).Decode(&createTaskReq); err != nil {
				t.Fatalf("decode create task request: %v", err)
			}
			writeTaskCreateResponse(w, 456, "Task with bucket", "desc", 42, []int64{99})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/tasks/456":
			writeTaskGetResponse(w, 456, "Task with bucket", "desc", 42)
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/42/views":
			fmt.Fprint(w, `[{"id":5,"project_id":42,"title":"Kanban","view_kind":"kanban","position":0,"bucket_configuration_mode":"none","default_bucket_id":0,"done_bucket_id":0}]`) //nolint:errcheck
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
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

//nolint:gocyclo
func TestTasksCreateCmd_RunE_SuccessWithTitleProjectAndBucket(t *testing.T) {
	saveAndResetGlobals()
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
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects":
			fmt.Fprint(w, `[{"id":2,"title":"Work"}]`) //nolint:errcheck
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/2/views":
			fmt.Fprint(w, `[{"id":5,"project_id":2,"title":"Kanban","view_kind":"kanban","position":0,"default_bucket_id":0,"done_bucket_id":0}]`) //nolint:errcheck
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/2/views/5/buckets":
			fmt.Fprint(w, `[{"id":10,"project_view_id":5,"title":"Todo","description":"","position":0,"is_done_bucket":false},{"id":11,"project_view_id":5,"title":"In Progress","description":"","position":1,"is_done_bucket":false}]`) //nolint:errcheck
		case r.Method == http.MethodPut && r.URL.Path == "/api/v1/projects/2/tasks":
			createTaskCalled = true
			if err := json.NewDecoder(r.Body).Decode(&createTaskReq); err != nil {
				t.Fatalf("decode create task request: %v", err)
			}
			writeTaskCreateResponse(w, 789, "Task with title bucket", "desc2", 2, []int64{11})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/tasks/789":
			writeTaskGetResponse(w, 789, "Task with title bucket", "desc2", 2)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
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
	saveAndResetGlobals()
	defer resetGlobals()

	buf := &bytes.Buffer{}
	outputWriter = buf
	logger = slog.New(slog.NewTextHandler(io.Discard, nil))

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
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
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires at least 1 arg")
}

func TestTasksCreateCmd_RunE_ErrorProjectNotFound(t *testing.T) {
	saveAndResetGlobals()
	defer resetGlobals()

	buf := &bytes.Buffer{}
	outputWriter = buf
	logger = slog.New(slog.NewTextHandler(io.Discard, nil))

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects" {
			fmt.Fprint(w, `[]`) //nolint:errcheck
		} else {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
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

//nolint:gocyclo
func TestTasksCreateCmd_RunE_ErrorBucketNotFound(t *testing.T) {
	saveAndResetGlobals()
	defer resetGlobals()

	buf := &bytes.Buffer{}
	outputWriter = buf
	logger = slog.New(slog.NewTextHandler(io.Discard, nil))

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects":
			fmt.Fprint(w, `[{"id":1,"title":"Inbox"}]`) //nolint:errcheck
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/1":
			fmt.Fprint(w, `{"id":1,"title":"Inbox"}`) //nolint:errcheck
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/1/views":
			fmt.Fprint(w, `[{"id":5,"project_id":1,"title":"Kanban","view_kind":"kanban","position":0}]`) //nolint:errcheck
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/1/views/5/buckets":
			fmt.Fprint(w, `[{"id":10,"project_view_id":5,"title":"Todo","description":"","position":0,"is_done_bucket":false}]`) //nolint:errcheck
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
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
	saveAndResetGlobals()
	defer resetGlobals()

	buf := &bytes.Buffer{}
	outputWriter = buf
	logger = slog.New(slog.NewTextHandler(io.Discard, nil))

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects":
			fmt.Fprint(w, `[{"id":1,"title":"Inbox"}]`) //nolint:errcheck
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects/1/views":
			fmt.Fprint(w, `[{"id":2,"project_id":1,"title":"List","view_kind":"list","position":0}]`) //nolint:errcheck
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
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
	saveAndResetGlobals()
	defer resetGlobals()

	buf := &bytes.Buffer{}
	outputWriter = buf
	logger = slog.New(slog.NewTextHandler(io.Discard, nil))

	ts := setupOutputTestServer(t, &outputTestCase{
		taskID:      100,
		taskTitle:   "JSON Task",
		description: "json desc",
		projectID:   1,
		bucketIDs:   []int64{},
		useJSON:     true,
	})
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
	var task map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &task))
	require.NotNil(t, task, "expected task object in JSON output")
	assert.InDelta(t, 100.0, task["id"], 0)
	assert.Equal(t, "JSON Task", task["title"])
}

func TestTasksCreateCmd_RunE_OutputMarkdown(t *testing.T) {
	saveAndResetGlobals()
	defer resetGlobals()

	buf := &bytes.Buffer{}
	outputWriter = buf
	logger = slog.New(slog.NewTextHandler(io.Discard, nil))

	ts := setupOutputTestServer(t, &outputTestCase{
		taskID:      200,
		taskTitle:   "Markdown Task",
		description: "md desc",
		projectID:   1,
		bucketIDs:   []int64{},
		useMarkdown: true,
	})
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
	assert.Contains(t, output, "# Markdown Task")
	assert.Contains(t, output, "**ID**: 200")
}

func TestTasksCreateCmd_RunE_OutputTable(t *testing.T) {
	saveAndResetGlobals()
	defer resetGlobals()

	buf := &bytes.Buffer{}
	outputWriter = buf
	logger = slog.New(slog.NewTextHandler(io.Discard, nil))

	ts := setupOutputTestServer(t, &outputTestCase{
		taskID:      300,
		taskTitle:   "Table Task",
		description: "",
		projectID:   1,
		bucketIDs:   []int64{},
	})
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
	assert.Contains(t, output, "Table Task")
	assert.Contains(t, output, "300")
}
