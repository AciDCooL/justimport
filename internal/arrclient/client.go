package arrclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Client is an HTTP client for a Radarr or Sonarr instance.
type Client struct {
	baseURL    string
	apiKey     string
	name       string
	httpClient *http.Client
}

// NewClient creates a new Client for the given *arr instance.
func NewClient(baseURL, apiKey, name string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		name:    name,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name returns the instance name (e.g. "radarr" or "sonarr").
func (c *Client) Name() string {
	return c.name
}

func (c *Client) doGet(ctx context.Context, path string, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("X-Api-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d for %s", resp.StatusCode, path)
	}

	err = json.NewDecoder(resp.Body).Decode(out)
	if err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

// CheckConnectivity verifies the connection to the instance by fetching system status.
// Returns the app name and version on success.
func (c *Client) CheckConnectivity(ctx context.Context) (string, string, error) {
	var status struct {
		AppName string `json:"appName"`
		Version string `json:"version"`
	}

	if err := c.doGet(ctx, "/api/v3/system/status", &status); err != nil {
		return "", "", err
	}

	return status.AppName, status.Version, nil
}

// GetQueue fetches all queue records from the instance.
func (c *Client) GetQueue(ctx context.Context) ([]QueueRecord, error) {
	var resp QueueResponse

	path := "/api/v3/queue?pageSize=1000&includeUnknownMovieItems=true&includeUnknownSeriesItems=true"
	if err := c.doGet(ctx, path, &resp); err != nil {
		return nil, err
	}

	return resp.Records, nil
}

// GetManualImport fetches the list of files available for manual import for the given download ID.
func (c *Client) GetManualImport(ctx context.Context, downloadID string) ([]ManualImportItem, error) {
	var items []ManualImportItem

	path := "/api/v3/manualimport?downloadId=" + downloadID
	if err := c.doGet(ctx, path, &items); err != nil {
		return nil, err
	}

	return items, nil
}

// PostManualImport sends a manual import approval request for the given item.
func (c *Client) PostManualImport(ctx context.Context, item ManualImportItem, importMode string) error {
	item.ImportApproved = true
	item.ImportMode = importMode

	data, err := json.Marshal([]ManualImportItem{item})
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v3/manualimport", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("X-Api-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status %d for manual import POST", resp.StatusCode)
	}

	return nil
}
