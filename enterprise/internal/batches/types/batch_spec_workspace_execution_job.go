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
	BatchSpecWorkspaceExecutionJobStateErrored    BatchSpecWorkspaceExecutionJobState = "errored"
	BatchSpecWorkspaceExecutionJobStateFailed     BatchSpecWorkspaceExecutionJobState = "failed"
	BatchSpecWorkspaceExecutionJobStateCompleted  BatchSpecWorkspaceExecutionJobState = "completed"
)

// Valid returns true if the given BatchSpecWorkspaceExecutionJobState is valid.
func (s BatchSpecWorkspaceExecutionJobState) Valid() bool {
	switch s {
	case BatchSpecWorkspaceExecutionJobStateQueued,
		BatchSpecWorkspaceExecutionJobStateProcessing,
		BatchSpecWorkspaceExecutionJobStateErrored,
		BatchSpecWorkspaceExecutionJobStateFailed,
		BatchSpecWorkspaceExecutionJobStateCompleted:
		return true
	default:
		return false
	}
}

// ToGraphQL returns the GraphQL representation of the worker state.
func (s BatchSpecWorkspaceExecutionJobState) ToGraphQL() string { return strings.ToUpper(string(s)) }

type BatchSpecWorkspaceExecutionJob struct {
	ID int64

	BatchSpecWorkspaceID int64
	AccessTokenID        int64

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

	PlaceInQueue int64

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (j *BatchSpecWorkspaceExecutionJob) RecordID() int { return int(j.ID) }
