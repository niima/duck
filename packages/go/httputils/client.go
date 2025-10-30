package httputils

import (
	"duck/common"
	"fmt"
	"net/http"
	"time"
)

// Client wraps http.Client with logging
type Client struct {
	client *http.Client
	logger *common.Logger
}

// NewClient creates a new HTTP client
func NewClient() *Client {
	return &Client{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: common.NewLogger("httputils"),
	}
}

// Get performs a GET request
func (c *Client) Get(url string) (*http.Response, error) {
	c.logger.Info(fmt.Sprintf("GET request to %s", url))
	return c.client.Get(url)
}

// Post performs a POST request
func (c *Client) Post(url, contentType string, body interface{}) (*http.Response, error) {
	c.logger.Info(fmt.Sprintf("POST request to %s", url))
	// Simplified for demo purposes
	return nil, fmt.Errorf("not implemented")
}
