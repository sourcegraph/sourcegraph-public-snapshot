package lsif

import (
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

type LSIFUploadStats struct {
	ProcessingCount int32 `json:"processingCount"`
	ErroredCount    int32 `json:"erroredCount"`
	CompletedCount  int32 `json:"completedCount"`
	QueuedCount     int32 `json:"queuedCount"`
}

type LSIFUpload struct {
	ID                string     `json:"id"`
	Repository        string     `json:"repository"`
	Commit            string     `json:"commit"`
	Root              string     `json:"root"`
	Filename          string     `json:"filename"`
	State             string     `json:"state"`
	FailureSummary    *string    `json:"failureSummary"`
	FailureStacktrace *string    `json:"failureStacktrace"`
	UploadedAt        time.Time  `json:"uploadedAt"`
	StartedAt         *time.Time `json:"startedAt"`
	FinishedAt        *time.Time `json:"finishedAt"`
}
