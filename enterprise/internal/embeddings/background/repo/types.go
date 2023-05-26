package repo

import (
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/executor"
)

type RepoEmbeddingJob struct {
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

	RepoID   api.RepoID
	Revision api.CommitID
}

func (j *RepoEmbeddingJob) RecordID() int {
	return j.ID
}

func (j *RepoEmbeddingJob) RecordUID() string {
	return strconv.Itoa(j.ID)
}

func (j *RepoEmbeddingJob) IsRepoEmbeddingJobScheduledOrCompleted() bool {
	return j != nil && (j.State == "completed" || j.State == "processing" || j.State == "queued" || j.Cancel)
}
