package lsif

import (
	"encoding/json"
	"time"
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
	Argumentss           *json.RawMessage `json:"arguments"`
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
