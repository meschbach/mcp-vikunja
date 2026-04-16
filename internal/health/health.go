// Package health provides health check functionality for the MCP Vikunja server.
package health

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// Status represents the health status of a component
type Status string

const (
	// StatusHealthy indicates the component is healthy
	StatusHealthy Status = "healthy"
	// StatusUnhealthy indicates the component is unhealthy
	StatusUnhealthy Status = "unhealthy"
	// StatusUnknown indicates the component status is unknown
	StatusUnknown Status = "unknown"
)

// CheckType represents the type of health check
type CheckType string

const (
	// CheckTypeLiveness represents a liveness check
	CheckTypeLiveness CheckType = "liveness"
	// CheckTypeReadiness represents a readiness check
	CheckTypeReadiness CheckType = "readiness"
)

// CheckResult represents the result of a health check
type CheckResult struct {
	Name         string                 `json:"name"`
	Status       Status                 `json:"status"`
	Message      string                 `json:"message,omitempty"`
	ResponseTime time.Duration          `json:"response_time_ms,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// Response represents the overall health check response
type Response struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Checks    map[string]CheckResult `json:"checks"`
}

// Checker defines the interface for health checks
type Checker interface {
	Name() string
	Check(ctx context.Context) CheckResult
}

// Manager manages multiple health checks
type Manager struct {
	checks []Checker
}

// New creates a new health check manager
func New() *Manager {
	return &Manager{
		checks: make([]Checker, 0),
	}
}

// Register adds a health check to the manager
func (hc *Manager) Register(checker Checker) {
	hc.checks = append(hc.checks, checker)
}

// CheckAll runs all registered health checks
func (hc *Manager) CheckAll(ctx context.Context) Response {
	response := Response{
		Status:    string(StatusHealthy),
		Timestamp: time.Now(),
		Checks:    make(map[string]CheckResult),
	}

	for _, checker := range hc.checks {
		result := checker.Check(ctx)
		response.Checks[checker.Name()] = result

		// If any check is unhealthy, mark overall as unhealthy
		if result.Status == StatusUnhealthy {
			response.Status = string(StatusUnhealthy)
		}
	}

	return response
}

// CheckLiveness returns liveness status
func (hc *Manager) CheckLiveness(_ context.Context) Response {
	return Response{
		Status:    string(StatusHealthy),
		Timestamp: time.Now(),
		Checks: map[string]CheckResult{
			"server": {
				Name:    "server",
				Status:  StatusHealthy,
				Message: "Server is running",
			},
		},
	}
}

// CheckReadiness returns readiness status
func (hc *Manager) CheckReadiness(ctx context.Context) Response {
	return hc.CheckAll(ctx)
}

// HTTPHandler returns an HTTP handler for health checks
func (hc *Manager) HTTPHandler(checkType CheckType) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var response Response

		switch checkType {
		case CheckTypeLiveness:
			response = hc.CheckLiveness(ctx)
		case CheckTypeReadiness:
			response = hc.CheckReadiness(ctx)
		default:
			response = hc.CheckAll(ctx)
		}

		// Set status code based on health status
		if response.Status == string(StatusUnhealthy) {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			slog.Error("failed to encode health response", "error", err)
		}
	}
}

// ServerCheck is a basic check for server status
type ServerCheck struct{}

// Name returns the name of the check
func (sc *ServerCheck) Name() string {
	return "server"
}

// Check performs the health check
func (sc *ServerCheck) Check(_ context.Context) CheckResult {
	return CheckResult{
		Name:    "server",
		Status:  StatusHealthy,
		Message: "Server is running",
	}
}

// VikunjaCheck checks Vikunja API connectivity
type VikunjaCheck struct {
	client interface {
		GetProjects(ctx context.Context) ([]interface{}, error)
	}
}

// NewVikunjaCheck creates a new Vikunja health check
func NewVikunjaCheck(client interface {
	GetProjects(ctx context.Context) ([]interface{}, error)
}) *VikunjaCheck {
	return &VikunjaCheck{client: client}
}

// Name returns the name of the check
func (vc *VikunjaCheck) Name() string {
	return "vikunja"
}

// Check performs the health check
func (vc *VikunjaCheck) Check(ctx context.Context) CheckResult {
	start := time.Now()

	// Try to fetch projects as a connectivity check
	_, err := vc.client.GetProjects(ctx)
	duration := time.Since(start)

	if err != nil {
		return CheckResult{
			Name:         "vikunja",
			Status:       StatusUnhealthy,
			Message:      fmt.Sprintf("Failed to connect to Vikunja: %v", err),
			ResponseTime: duration,
		}
	}

	return CheckResult{
		Name:         "vikunja",
		Status:       StatusHealthy,
		Message:      "Successfully connected to Vikunja",
		ResponseTime: duration,
	}
}
