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

// ManualImportItem represents a file available for manual import
// (GET /api/v3/manualimport response).
type ManualImportItem struct {
	ID           int             `json:"id"`
	Path         string          `json:"path"`
	FolderName   string          `json:"folderName,omitempty"`
	Name         string          `json:"name,omitempty"`
	Size         int64           `json:"size,omitempty"`
	Movie        *MediaRef       `json:"movie,omitempty"`
	Series       *MediaRef       `json:"series,omitempty"`
	SeasonNumber *int            `json:"seasonNumber,omitempty"`
	Episodes     []EpisodeRef    `json:"episodes,omitempty"`
	Rejections   []Rejection     `json:"rejections"`
	DownloadID   string          `json:"downloadId,omitempty"`
	Quality      json.RawMessage `json:"quality,omitempty"`
	Languages    json.RawMessage `json:"languages,omitempty"`
	ReleaseGroup string          `json:"releaseGroup,omitempty"`
	IndexerFlags int             `json:"indexerFlags,omitempty"`
	ReleaseType  json.RawMessage `json:"releaseType,omitempty"`
}

// MediaRef holds the ID and title of a matched movie or series.
type MediaRef struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

// EpisodeRef holds the ID of an episode reference.
type EpisodeRef struct {
	ID int `json:"id"`
}

// Rejection represents a reason a manual import item cannot be imported.
type Rejection struct {
	Reason string `json:"reason"`
	Type   string `json:"type"`
}

// ManualImportCommand is the request body for POST /api/v3/command
// to trigger a manual import.
type ManualImportCommand struct {
	Name       string             `json:"name"`
	Files      []ManualImportFile `json:"files"`
	ImportMode string             `json:"importMode"`
}

// ManualImportFile represents a single file within a ManualImportCommand.
type ManualImportFile struct {
	Path         string          `json:"path"`
	FolderName   string          `json:"folderName,omitempty"`
	MovieID      int             `json:"movieId,omitempty"`
	SeriesID     int             `json:"seriesId,omitempty"`
	EpisodeIDs   []int           `json:"episodeIds,omitempty"`
	Quality      json.RawMessage `json:"quality,omitempty"`
	Languages    json.RawMessage `json:"languages,omitempty"`
	ReleaseGroup string          `json:"releaseGroup,omitempty"`
	IndexerFlags int             `json:"indexerFlags,omitempty"`
	DownloadID   string          `json:"downloadId,omitempty"`
	ReleaseType  json.RawMessage `json:"releaseType,omitempty"`
}
