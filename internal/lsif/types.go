package lsif

import (
	"encoding/json"
	"time"

	"github.com/sourcegraph/go-lsp"
)

type LSIFDump struct {
	ID           int64     `json:"id"`
	Repository   string    `json:"repository"`
	Commit       string    `json:"commit"`
	Root         string    `json:"root"`
	VisibleAtTip bool      `json:"visibleAtTip"`
	UploadedAt   time.Time `json:"uploadedAt"`
	ProcessedAt  time.Time `json:"processedAt"`
}

type LSIFLocation struct {
	Repository string    `json:"repository"`
	Commit     string    `json:"commit"`
	Path       string    `json:"path"`
	Range      lsp.Range `json:"range"`
}

type LSIFJobStats struct {
	ProcessingCount int32 `json:"processingCount"`
	ErroredCount    int32 `json:"erroredCount"`
	CompletedCount  int32 `json:"completedCount"`
	QueuedCount     int32 `json:"queuedCount"`
	ScheduledCount  int32 `json:"scheduledCount"`
}

type LSIFJob struct {
	ID                   string           `json:"id"`
	Type                 string           `json:"type"`
	Arguments            *json.RawMessage `json:"arguments"`
	State                string           `json:"state"`
	Failure              *LSIFJobFailure  `json:"failure"`
	QueuedAt             time.Time        `json:"queuedAt"`
	StartedAt            *time.Time       `json:"startedAt"`
	CompletedOrErroredAt *time.Time       `json:"completedOrErroredAt"`
}

type LSIFJobFailure struct {
	Summary     string   `json:"summary"`
	Stacktraces []string `json:"stacktraces"`
}
