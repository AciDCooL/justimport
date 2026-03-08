package arrclient

import "encoding/json"

// QueueResponse is the response from GET /api/v3/queue.
type QueueResponse struct {
	Records      []QueueRecord `json:"records"`
	TotalRecords int           `json:"totalRecords"`
}

// QueueRecord represents a single item in the Radarr/Sonarr queue.
type QueueRecord struct {
	ID             int             `json:"id"`
	Title          string          `json:"title"`
	DownloadID     string          `json:"downloadId"`
	StatusMessages []StatusMessage `json:"statusMessages"`
}

// StatusMessage is a status message attached to a queue record.
type StatusMessage struct {
	Title    string   `json:"title"`
	Messages []string `json:"messages"`
}

// ManualImportItem represents a file available for manual import.
// Used for both the GET response and the POST request body.
type ManualImportItem struct {
	ID             int             `json:"id"`
	Path           string          `json:"path"`
	Name           string          `json:"name,omitempty"`
	Size           int64           `json:"size,omitempty"`
	Movie          *MediaTitle     `json:"movie,omitempty"`
	Series         *MediaTitle     `json:"series,omitempty"`
	Rejections     []Rejection     `json:"rejections"`
	DownloadID     string          `json:"downloadId,omitempty"`
	Quality        json.RawMessage `json:"quality,omitempty"`
	Languages      json.RawMessage `json:"languages,omitempty"`
	ReleaseGroup   string          `json:"releaseGroup,omitempty"`
	ImportMode     string          `json:"importMode,omitempty"`
	ImportApproved bool            `json:"importApproved,omitempty"`
}

// MediaTitle holds the title of a matched movie or series.
type MediaTitle struct {
	Title string `json:"title"`
}

// Rejection represents a reason a manual import item cannot be imported.
type Rejection struct {
	Reason string `json:"reason"`
	Type   string `json:"type"`
}
