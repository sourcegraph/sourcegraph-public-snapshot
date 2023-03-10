package contextdetection

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/executor"
)

type ContextDetectionEmbeddingJob struct {
	ID              int
	State           string
	FailureMessage  *string
	QueuedAt        time.Time
	StartedAt       *time.Time
	FinishedAt      *time.Time
	ProcessAfter    *time.Time
	NumResets       int
	NumFailures     int
	LastHeartbeatAt *time.Time
	ExecutionLogs   []executor.ExecutionLogEntry
	WorkerHostname  string
	Cancel          bool
}

func (j *ContextDetectionEmbeddingJob) RecordID() int {
	return j.ID
}
