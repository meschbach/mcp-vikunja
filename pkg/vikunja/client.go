package vikunja

import (
	"fmt"
	"net/http"
	"time"
)

// Client wraps the Vikunja API client
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient creates a new Vikunja API client
func NewClient(host, token string, insecure bool) (*Client, error) {
	scheme := "https"
	if insecure {
		scheme = "http"
	}

	baseURL := fmt.Sprintf("%s://%s/api/v1", scheme, host)

	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}
