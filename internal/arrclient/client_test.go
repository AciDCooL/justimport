package arrclient_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/erkexzcx/justimport/internal/arrclient"
)

func TestCheckConnectivity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/system/status" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Header.Get("X-Api-Key") != "test-key" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if err := json.NewEncoder(w).Encode(map[string]string{
			"appName": "Radarr",
			"version": "5.0.0",
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	client := arrclient.NewClient(server.URL, "test-key", "radarr")
	appName, version, err := client.CheckConnectivity(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if appName != "Radarr" {
		t.Errorf("expected appName Radarr, got %s", appName)
	}
	if version != "5.0.0" {
		t.Errorf("expected version 5.0.0, got %s", version)
	}
}

func TestCheckConnectivity_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := arrclient.NewClient(server.URL, "test-key", "radarr")
	_, _, err := client.CheckConnectivity(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetQueue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Api-Key") != "test-key" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		resp := arrclient.QueueResponse{
			TotalRecords: 2,
			Records: []arrclient.QueueRecord{
				{
					ID:         1,
					Title:      "Movie.2020.1080p.mkv",
					DownloadID: "abc123",
					StatusMessages: []arrclient.StatusMessage{
						{
							Title:    "Manual import required",
							Messages: []string{"Release was matched to movie by ID."},
						},
					},
				},
				{
					ID:         2,
					Title:      "Another.Movie.2021.mkv",
					DownloadID: "def456",
					StatusMessages: []arrclient.StatusMessage{
						{Title: "Downloaded", Messages: nil},
					},
				},
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	client := arrclient.NewClient(server.URL, "test-key", "radarr")
	records, err := client.GetQueue(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}
	if records[0].DownloadID != "abc123" {
		t.Errorf("expected downloadId abc123, got %s", records[0].DownloadID)
	}
	if records[0].StatusMessages[0].Title != "Manual import required" {
		t.Errorf("unexpected status message title: %s", records[0].StatusMessages[0].Title)
	}
}

func TestGetManualImport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("downloadId") != "abc123" {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		items := []arrclient.ManualImportItem{
			{
				ID:   1,
				Path: "/downloads/Movie.2020.1080p.mkv",
				Name: "Movie.2020.1080p.mkv",
				Size: 8000000000,
				Movie: &arrclient.MediaTitle{
					Title: "Movie 2020",
				},
				Rejections: []arrclient.Rejection{},
			},
		}
		if err := json.NewEncoder(w).Encode(items); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	client := arrclient.NewClient(server.URL, "test-key", "radarr")
	items, err := client.GetManualImport(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Path != "/downloads/Movie.2020.1080p.mkv" {
		t.Errorf("unexpected path: %s", items[0].Path)
	}
	if items[0].Movie == nil || items[0].Movie.Title != "Movie 2020" {
		t.Errorf("unexpected movie title")
	}
}

func TestPostManualImport(t *testing.T) {
	var receivedItems []arrclient.ManualImportItem

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&receivedItems); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := arrclient.NewClient(server.URL, "test-key", "radarr")
	item := arrclient.ManualImportItem{
		ID:   1,
		Path: "/downloads/Movie.2020.mkv",
		Movie: &arrclient.MediaTitle{
			Title: "Movie 2020",
		},
	}

	if err := client.PostManualImport(context.Background(), item); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(receivedItems) != 1 {
		t.Fatalf("expected 1 item in POST body, got %d", len(receivedItems))
	}
	if !receivedItems[0].ImportApproved {
		t.Error("expected ImportApproved to be true")
	}
	if receivedItems[0].ImportMode != "Move" {
		t.Errorf("expected ImportMode Move, got %s", receivedItems[0].ImportMode)
	}
}

func TestPostManualImport_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := arrclient.NewClient(server.URL, "test-key", "radarr")
	item := arrclient.ManualImportItem{ID: 1, Path: "/downloads/movie.mkv"}

	if err := client.PostManualImport(context.Background(), item); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClientName(t *testing.T) {
	client := arrclient.NewClient("http://localhost:7878", "key", "radarr")
	if client.Name() != "radarr" {
		t.Errorf("expected radarr, got %s", client.Name())
	}
}

func TestJSONUnmarshal_QueueRecord(t *testing.T) {
	raw := `{
		"id": 42,
		"title": "White.Noise.2.2007.1080p.mkv",
		"downloadId": "XYZ789",
		"statusMessages": [
			{
				"title": "Found matching movie via grab history, but release was matched to movie by ID. Manual Import required.",
				"messages": ["Release was matched to movie by ID"]
			}
		]
	}`

	var record arrclient.QueueRecord
	if err := json.Unmarshal([]byte(raw), &record); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if record.ID != 42 {
		t.Errorf("expected id 42, got %d", record.ID)
	}
	if record.DownloadID != "XYZ789" {
		t.Errorf("expected downloadId XYZ789, got %s", record.DownloadID)
	}
	if len(record.StatusMessages) != 1 {
		t.Fatalf("expected 1 status message, got %d", len(record.StatusMessages))
	}
}

func TestJSONUnmarshal_ManualImportItem(t *testing.T) {
	raw := `{
		"id": 1,
		"path": "/downloads/White.Noise.2.2007.1080p.mkv",
		"name": "White.Noise.2.2007.1080p.mkv",
		"size": 9000000000,
		"movie": {"id": 5, "title": "White Noise 2: The Light"},
		"rejections": [],
		"downloadId": "XYZ789",
		"quality": {"quality": {"id": 7, "name": "Bluray-1080p"}},
		"languages": [{"id": 1, "name": "English"}]
	}`

	var item arrclient.ManualImportItem
	if err := json.Unmarshal([]byte(raw), &item); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Path != "/downloads/White.Noise.2.2007.1080p.mkv" {
		t.Errorf("unexpected path: %s", item.Path)
	}
	if item.Movie == nil || item.Movie.Title != "White Noise 2: The Light" {
		t.Errorf("unexpected movie title")
	}
	if item.Size != 9000000000 {
		t.Errorf("unexpected size: %d", item.Size)
	}
	if len(item.Rejections) != 0 {
		t.Errorf("expected 0 rejections, got %d", len(item.Rejections))
	}
}
