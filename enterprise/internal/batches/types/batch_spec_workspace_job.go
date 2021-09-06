package types

import (
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

// BatchSpecWorkspaceJobState defines the possible states of a changeset job.
type BatchSpecWorkspaceJobState string

// BatchSpecWorkspaceJobState constants.
const (
	BatchSpecWorkspaceJobStatePending BatchSpecWorkspaceJobState = "PENDING"

	BatchSpecWorkspaceJobStateQueued     BatchSpecWorkspaceJobState = "QUEUED"
	BatchSpecWorkspaceJobStateProcessing BatchSpecWorkspaceJobState = "PROCESSING"
	BatchSpecWorkspaceJobStateErrored    BatchSpecWorkspaceJobState = "ERRORED"
	BatchSpecWorkspaceJobStateFailed     BatchSpecWorkspaceJobState = "FAILED"
	BatchSpecWorkspaceJobStateCompleted  BatchSpecWorkspaceJobState = "COMPLETED"
)

// Valid returns true if the given BatchSpecWorkspaceJobState is valid.
func (s BatchSpecWorkspaceJobState) Valid() bool {
	switch s {
	case BatchSpecWorkspaceJobStateQueued,
		BatchSpecWorkspaceJobStateProcessing,
		BatchSpecWorkspaceJobStateErrored,
		BatchSpecWorkspaceJobStateFailed,
		BatchSpecWorkspaceJobStateCompleted:
		return true
	default:
		return false
	}
}

// ToDB returns the database representation of the worker state. That's
// needed because we want to use UPPERCASE in the application and GraphQL layer,
// but need to use lowercase in the database to make it work with workerutil.Worker.
func (s BatchSpecWorkspaceJobState) ToDB() string { return strings.ToLower(string(s)) }

type BatchSpecWorkspaceJob struct {
	ID int64

	BatchSpecID      int64
	ChangesetSpecIDs []int64

	RepoID api.RepoID
	Branch string
	Commit string
	Path   string

	// workerutil fields
	State           BatchSpecWorkspaceJobState
	FailureMessage  *string
	StartedAt       time.Time
	FinishedAt      time.Time
	ProcessAfter    time.Time
	NumResets       int64
	NumFailures     int64
	LastHeartbeatAt time.Time

	ExecutionLogs  []workerutil.ExecutionLogEntry
	WorkerHostname string

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (j *BatchSpecWorkspaceJob) RecordID() int {
	return int(j.ID)
}
