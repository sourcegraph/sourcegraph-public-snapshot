package types

import (
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
)

// BatchSpecWorkspaceState defines the possible states of a changeset job.
type BatchSpecWorkspaceState string

// BatchSpecWorkspaceState constants.
const (
	BatchSpecWorkspaceStatePending BatchSpecWorkspaceState = "PENDING"

	BatchSpecWorkspaceStateQueued     BatchSpecWorkspaceState = "QUEUED"
	BatchSpecWorkspaceStateProcessing BatchSpecWorkspaceState = "PROCESSING"
	BatchSpecWorkspaceStateErrored    BatchSpecWorkspaceState = "ERRORED"
	BatchSpecWorkspaceStateFailed     BatchSpecWorkspaceState = "FAILED"
	BatchSpecWorkspaceStateCompleted  BatchSpecWorkspaceState = "COMPLETED"
)

// Valid returns true if the given BatchSpecWorkspaceState is valid.
func (s BatchSpecWorkspaceState) Valid() bool {
	switch s {
	case BatchSpecWorkspaceStateQueued,
		BatchSpecWorkspaceStateProcessing,
		BatchSpecWorkspaceStateErrored,
		BatchSpecWorkspaceStateFailed,
		BatchSpecWorkspaceStateCompleted:
		return true
	default:
		return false
	}
}

// ToDB returns the database representation of the worker state. That's
// needed because we want to use UPPERCASE in the application and GraphQL layer,
// but need to use lowercase in the database to make it work with workerutil.Worker.
func (s BatchSpecWorkspaceState) ToDB() string { return strings.ToLower(string(s)) }

type BatchSpecWorkspace struct {
	ID int64

	BatchSpecID      int64
	ChangesetSpecIDs []int64

	RepoID             api.RepoID
	Branch             string
	Commit             string
	Path               string
	Steps              []batcheslib.Step
	FileMatches        []string
	OnlyFetchWorkspace bool

	// workerutil fields
	State           BatchSpecWorkspaceState
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

func (j *BatchSpecWorkspace) RecordID() int {
	return int(j.ID)
}
