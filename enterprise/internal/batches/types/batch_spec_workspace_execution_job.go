package types

import (
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

// BatchSpecWorkspaceExecutionJobState defines the possible states of a changeset job.
type BatchSpecWorkspaceExecutionJobState string

// BatchSpecWorkspaceExecutionJobState constants.
const (
	BatchSpecWorkspaceExecutionJobStateQueued     BatchSpecWorkspaceExecutionJobState = "queued"
	BatchSpecWorkspaceExecutionJobStateProcessing BatchSpecWorkspaceExecutionJobState = "processing"
	BatchSpecWorkspaceExecutionJobStateFailed     BatchSpecWorkspaceExecutionJobState = "failed"
	BatchSpecWorkspaceExecutionJobStateCompleted  BatchSpecWorkspaceExecutionJobState = "completed"

	// There is no Errored state because automatic-retry of
	// BatchSpecWorkspaceExecutionJobs is disabled. If a job fails, it's
	// "failed" and needs to be retried manually.
)

// Valid returns true if the given BatchSpecWorkspaceExecutionJobState is valid.
func (s BatchSpecWorkspaceExecutionJobState) Valid() bool {
	switch s {
	case BatchSpecWorkspaceExecutionJobStateQueued,
		BatchSpecWorkspaceExecutionJobStateProcessing,
		BatchSpecWorkspaceExecutionJobStateFailed,
		BatchSpecWorkspaceExecutionJobStateCompleted:
		return true
	default:
		return false
	}
}

// ToGraphQL returns the GraphQL representation of the worker state.
func (s BatchSpecWorkspaceExecutionJobState) ToGraphQL() string { return strings.ToUpper(string(s)) }

// Retryable returns whether the state is retryable.
func (s BatchSpecWorkspaceExecutionJobState) Retryable() bool {
	return s == BatchSpecWorkspaceExecutionJobStateFailed ||
		s == BatchSpecWorkspaceExecutionJobStateCompleted
}

type BatchSpecWorkspaceExecutionJob struct {
	ID int64

	BatchSpecWorkspaceID int64
	UserID               int32

	State           BatchSpecWorkspaceExecutionJobState
	FailureMessage  *string
	StartedAt       time.Time
	FinishedAt      time.Time
	ProcessAfter    time.Time
	NumResets       int64
	NumFailures     int64
	LastHeartbeatAt time.Time
	ExecutionLogs   []workerutil.ExecutionLogEntry
	WorkerHostname  string
	Cancel          bool

	PlaceInUserQueue   int64
	PlaceInGlobalQueue int64

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (j *BatchSpecWorkspaceExecutionJob) RecordID() int { return int(j.ID) }
