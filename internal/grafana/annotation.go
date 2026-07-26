// Package grafana posts annotations to a Grafana instance via its HTTP API.
package grafana

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type annotationRequest struct {
	Time int64    `json:"time"` // epoch milliseconds
	Tags []string `json:"tags"`
	Text string   `json:"text"`
}

// CreateAnnotation posts a global annotation at the given time.
func (c *Client) CreateAnnotation(ctx context.Context, at time.Time, text string, tags []string) error {
	body, err := json.Marshal(annotationRequest{
		Time: at.UnixMilli(),
		Tags: tags,
		Text: text,
	})
	if err != nil {
		return fmt.Errorf("marshal annotation: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/annotations", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("grafana returned status %d", resp.StatusCode)
	}
	return nil
}
