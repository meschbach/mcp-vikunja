package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusConstants(t *testing.T) {
	assert.Equal(t, Status("healthy"), StatusHealthy)
	assert.Equal(t, Status("unhealthy"), StatusUnhealthy)
	assert.Equal(t, Status("unknown"), StatusUnknown)
}

func TestCheckTypeConstants(t *testing.T) {
	assert.Equal(t, CheckType("liveness"), CheckTypeLiveness)
	assert.Equal(t, CheckType("readiness"), CheckTypeReadiness)
}

func TestNewHealthChecker(t *testing.T) {
	hc := NewHealthChecker()
	assert.NotNil(t, hc)
	assert.NotNil(t, hc.checks)
	assert.Empty(t, hc.checks)
}

func TestHealthChecker_Register(t *testing.T) {
	hc := NewHealthChecker()
	check := &ServerCheck{}

	hc.Register(check)
	assert.Len(t, hc.checks, 1)
}

func TestHealthChecker_CheckAll(t *testing.T) {
	t.Run("with server check", func(t *testing.T) {
		hc := NewHealthChecker()
		hc.Register(&ServerCheck{})

		response := hc.CheckAll(context.Background())

		assert.Equal(t, string(StatusHealthy), response.Status)
		assert.NotNil(t, response.Timestamp)
		assert.Len(t, response.Checks, 1)
		assert.Contains(t, response.Checks, "server")
	})

	t.Run("empty checks", func(t *testing.T) {
		hc := NewHealthChecker()

		response := hc.CheckAll(context.Background())

		assert.Equal(t, string(StatusHealthy), response.Status)
		assert.Empty(t, response.Checks)
	})
}

func TestHealthChecker_CheckLiveness(t *testing.T) {
	hc := NewHealthChecker()
	response := hc.CheckLiveness(context.Background())

	assert.Equal(t, string(StatusHealthy), response.Status)
	assert.NotNil(t, response.Timestamp)
	assert.Len(t, response.Checks, 1)
	assert.Contains(t, response.Checks, "server")
	assert.Equal(t, StatusHealthy, response.Checks["server"].Status)
}

func TestHealthChecker_CheckReadiness(t *testing.T) {
	hc := NewHealthChecker()
	hc.Register(&ServerCheck{})

	response := hc.CheckReadiness(context.Background())

	assert.Equal(t, string(StatusHealthy), response.Status)
	assert.Len(t, response.Checks, 1)
}

func TestHealthChecker_HTTPHandler(t *testing.T) {
	hc := NewHealthChecker()
	hc.Register(&ServerCheck{})

	t.Run("health endpoint", func(t *testing.T) {
		handler := hc.HTTPHandler("")
		req := httptest.NewRequest("GET", "/health", nil)
		rec := httptest.NewRecorder()

		handler(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response Response
		err := json.NewDecoder(rec.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, string(StatusHealthy), response.Status)
	})

	t.Run("liveness endpoint", func(t *testing.T) {
		handler := hc.HTTPHandler(CheckTypeLiveness)
		req := httptest.NewRequest("GET", "/health/live", nil)
		rec := httptest.NewRecorder()

		handler(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response Response
		err := json.NewDecoder(rec.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, string(StatusHealthy), response.Status)
	})

	t.Run("readiness endpoint", func(t *testing.T) {
		handler := hc.HTTPHandler(CheckTypeReadiness)
		req := httptest.NewRequest("GET", "/health/ready", nil)
		rec := httptest.NewRecorder()

		handler(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response Response
		err := json.NewDecoder(rec.Body).Decode(&response)
		require.NoError(t, err)
		assert.NotNil(t, response.Checks)
	})

	t.Run("unhealthy response", func(t *testing.T) {
		// Create a checker that returns unhealthy
		unhealthyChecker := &mockChecker{
			name:   "test",
			status: StatusUnhealthy,
		}

		hc := NewHealthChecker()
		hc.Register(unhealthyChecker)

		handler := hc.HTTPHandler("")
		req := httptest.NewRequest("GET", "/health", nil)
		rec := httptest.NewRecorder()

		handler(rec, req)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

		var response Response
		err := json.NewDecoder(rec.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, string(StatusUnhealthy), response.Status)
	})
}

func TestServerCheck(t *testing.T) {
	check := &ServerCheck{}

	t.Run("name", func(t *testing.T) {
		assert.Equal(t, "server", check.Name())
	})

	t.Run("check", func(t *testing.T) {
		result := check.Check(context.Background())

		assert.Equal(t, "server", result.Name)
		assert.Equal(t, StatusHealthy, result.Status)
		assert.Equal(t, "Server is running", result.Message)
	})
}

func TestVikunjaCheck(t *testing.T) {
	t.Run("healthy vikunja", func(t *testing.T) {
		mockClient := &mockVikunjaClient{
			projects: []interface{}{map[string]interface{}{"id": 1}},
			err:      nil,
		}

		check := NewVikunjaCheck(mockClient)
		result := check.Check(context.Background())

		assert.Equal(t, "vikunja", result.Name)
		assert.Equal(t, StatusHealthy, result.Status)
		assert.Contains(t, result.Message, "Successfully connected")
		assert.GreaterOrEqual(t, result.ResponseTime, time.Duration(0))
	})

	t.Run("unhealthy vikunja", func(t *testing.T) {
		mockClient := &mockVikunjaClient{
			projects: nil,
			err:      errors.New("connection failed"),
		}

		check := NewVikunjaCheck(mockClient)
		result := check.Check(context.Background())

		assert.Equal(t, "vikunja", result.Name)
		assert.Equal(t, StatusUnhealthy, result.Status)
		assert.Contains(t, result.Message, "Failed to connect")
	})

	t.Run("name", func(t *testing.T) {
		check := NewVikunjaCheck(&mockVikunjaClient{})
		assert.Equal(t, "vikunja", check.Name())
	})
}

func TestCheckResult_Metadata(t *testing.T) {
	result := CheckResult{
		Name:    "test",
		Status:  StatusHealthy,
		Message: "Test message",
		Metadata: map[string]interface{}{
			"version": "1.0.0",
			"count":   42,
		},
	}

	assert.Equal(t, "test", result.Name)
	assert.Equal(t, StatusHealthy, result.Status)
	assert.Equal(t, "1.0.0", result.Metadata["version"])
}

func TestResponse_JSON(t *testing.T) {
	response := Response{
		Status:    string(StatusHealthy),
		Timestamp: time.Now(),
		Checks: map[string]CheckResult{
			"server": {
				Name:    "server",
				Status:  StatusHealthy,
				Message: "OK",
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(response)
	require.NoError(t, err)

	// Unmarshal and verify
	var decoded Response
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, response.Status, decoded.Status)
	assert.Equal(t, response.Checks["server"].Status, decoded.Checks["server"].Status)
}

// mock implementations for testing

type mockChecker struct {
	name   string
	status Status
}

func (m *mockChecker) Name() string {
	return m.name
}

func (m *mockChecker) Check(ctx context.Context) CheckResult {
	return CheckResult{
		Name:   m.name,
		Status: m.status,
	}
}

type mockVikunjaClient struct {
	projects []interface{}
	err      error
}

func (m *mockVikunjaClient) GetProjects(ctx context.Context) ([]interface{}, error) {
	return m.projects, m.err
}
