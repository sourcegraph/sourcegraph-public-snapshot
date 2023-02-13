package embeddings

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/executor"
)

type EmbeddingJob struct {
	ID              int
	State           string
	FailureMessage  *string
	QueuedAt        time.Time
	StartedAt       *time.Time
	FinishedAt      *time.Time
	ProcessAfter    *time.Time
	NumResets       int
	NumFailures     int
	LastHeartbeatAt time.Time
	ExecutionLogs   []executor.ExecutionLogEntry
	WorkerHostname  string
	Cancel          bool

	RepoID   api.RepoID
	Revision string
}

func (j *EmbeddingJob) RecordID() int {
	return j.ID
}
