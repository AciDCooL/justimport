package importer

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/erkexzcx/justimport/internal/arrclient"
)

// ArrClient is the interface for interacting with a *arr instance.
type ArrClient interface {
	Name() string
	GetQueue(ctx context.Context) ([]arrclient.QueueRecord, error)
	GetManualImport(ctx context.Context, downloadID string) ([]arrclient.ManualImportItem, error)
	PostManualImport(ctx context.Context, item *arrclient.ManualImportItem) error
}

// sanitizeLog replaces newline and carriage return characters in a string
// to prevent log injection attacks via server-controlled data.
func sanitizeLog(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return s
}

// Importer polls Radarr/Sonarr instances and auto-imports eligible queue items.
type Importer struct {
	clients []ArrClient
	dryRun  bool
	seen    map[string]struct{}
}

// New creates a new Importer.
func New(clients []ArrClient, dryRun bool) *Importer {
	return &Importer{
		clients: clients,
		dryRun:  dryRun,
		seen:    make(map[string]struct{}),
	}
}

// Run starts the poll loop, performing an immediate poll then ticking at the given interval.
// It blocks until ctx is cancelled.
func (imp *Importer) Run(ctx context.Context, interval time.Duration) {
	imp.pollAll(ctx)

	if ctx.Err() != nil {
		return
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			imp.pollAll(ctx)
		}
	}
}

func (imp *Importer) pollAll(ctx context.Context) {
	for _, client := range imp.clients {
		imp.pollInstance(ctx, client)
	}
}

func (imp *Importer) pollInstance(ctx context.Context, client ArrClient) {
	name := client.Name()

	records, err := client.GetQueue(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("[%s] Failed to fetch queue: %v", name, err))
		return
	}

	pending := make([]arrclient.QueueRecord, 0, len(records))
	for _, r := range records {
		if needsManualImport(r) {
			pending = append(pending, r)
		}
	}

	slog.Info(fmt.Sprintf("[%s] Checking queue... found %d items requiring manual import", name, len(pending)))

	for _, record := range pending {
		if _, ok := imp.seen[record.DownloadID]; ok {
			continue
		}
		if imp.processItem(ctx, client, record) {
			imp.seen[record.DownloadID] = struct{}{}
		}
	}
}

// processItem returns true if the item was handled (imported or deliberately skipped)
// and should be marked as seen, or false if a transient error occurred and the item
// should be retried on the next poll.
func (imp *Importer) processItem(ctx context.Context, client ArrClient, record arrclient.QueueRecord) bool {
	name := client.Name()

	items, err := client.GetManualImport(ctx, record.DownloadID)
	if err != nil {
		slog.Error(fmt.Sprintf("[%s] Failed to fetch manual import for %q: %v", name, sanitizeLog(record.Title), err))
		return false
	}

	filtered := filterItems(items)

	switch {
	case len(filtered) == 0:
		slog.Warn(fmt.Sprintf("[%s] %q → SKIPPED: 0 files found after filtering", name, sanitizeLog(record.Title)))
		return true

	case len(filtered) > 1:
		slog.Warn(fmt.Sprintf("[%s] %q → SKIPPED: %d files found after filtering (expected exactly 1)", name, sanitizeLog(record.Title), len(filtered)))
		return true

	default:
		return imp.importSingleFile(ctx, client, record, &filtered[0])
	}
}

// importSingleFile returns true if the item was handled (imported or deliberately skipped)
// and should be marked as seen, or false if a transient error occurred.
func (imp *Importer) importSingleFile(ctx context.Context, client ArrClient, record arrclient.QueueRecord, item *arrclient.ManualImportItem) bool {
	name := client.Name()

	if len(item.Rejections) > 0 {
		slog.Warn(fmt.Sprintf("[%s] %q → SKIPPED: file has %d rejection(s): %s", name, sanitizeLog(record.Title), len(item.Rejections), sanitizeLog(item.Rejections[0].Reason)))
		return true
	}

	matched := matchedTitle(item)
	if matched == "" {
		slog.Warn(fmt.Sprintf("[%s] %q → SKIPPED: file not matched to any movie or series", name, sanitizeLog(record.Title)))
		return true
	}

	if imp.dryRun {
		slog.Info(fmt.Sprintf("[%s] %q → WOULD IMPORT (1 file, matched to %q)", name, sanitizeLog(record.Title), sanitizeLog(matched)))
		return true
	}

	if err := client.PostManualImport(ctx, item); err != nil {
		slog.Error(fmt.Sprintf("[%s] %q → IMPORT FAILED: %v", name, sanitizeLog(record.Title), err))
		return false
	}

	slog.Info(fmt.Sprintf("[%s] %q → IMPORTED (1 file, matched to %q)", name, sanitizeLog(record.Title), sanitizeLog(matched)))
	return true
}

// needsManualImport returns true if the queue record requires manual import.
func needsManualImport(record arrclient.QueueRecord) bool {
	for _, sm := range record.StatusMessages {
		if containsImportIndicator(sm.Title) {
			return true
		}
		for _, msg := range sm.Messages {
			if containsImportIndicator(msg) {
				return true
			}
		}
	}
	return false
}

func containsImportIndicator(s string) bool {
	lower := strings.ToLower(s)
	return strings.Contains(lower, "manual import required") ||
		strings.Contains(lower, "matched to movie by id")
}

// filterItems removes sample files from the list.
func filterItems(items []arrclient.ManualImportItem) []arrclient.ManualImportItem {
	filtered := make([]arrclient.ManualImportItem, 0, len(items))
	for i := range items {
		if !isSampleFile(items[i].Path) {
			filtered = append(filtered, items[i])
		}
	}
	return filtered
}

// isSampleFile returns true if the file path indicates a sample file.
func isSampleFile(path string) bool {
	return strings.Contains(strings.ToLower(path), "sample")
}

// matchedTitle returns the title of the matched movie or series, or empty string if unmatched.
func matchedTitle(item *arrclient.ManualImportItem) string {
	if item.Movie != nil && item.Movie.Title != "" {
		return item.Movie.Title
	}
	if item.Series != nil && item.Series.Title != "" {
		return item.Series.Title
	}
	return ""
}
