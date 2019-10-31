package types

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
}

type LSIFJobStats struct {
	ProcessingCount int32 `json:"processingCount"`
	ErroredCount    int32 `json:"erroredCount"`
	CompletedCount  int32 `json:"completedCount"`
	QueuedCount     int32 `json:"queuedCount"`
	ScheduledCount  int32 `json:"scheduledCount"`
}

type LSIFJob struct {
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	Args         *json.RawMessage `json:"args"`
	State        string           `json:"state"`
	Progress     float64          `json:"progress"`
	FailedReason *string          `json:"failedReason"`
	Stacktrace   *[]string        `json:"stacktrace"`
	Timestamp    time.Time        `json:"timestamp"`
	ProcessedOn  *time.Time       `json:"processedOn"`
	FinishedOn   *time.Time       `json:"finishedOn"`
}
