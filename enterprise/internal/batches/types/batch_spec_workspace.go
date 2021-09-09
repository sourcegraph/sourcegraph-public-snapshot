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
	BatchSpecWorkspaceStatePending BatchSpecWorkspaceState = "pending"

	BatchSpecWorkspaceStateQueued     BatchSpecWorkspaceState = "queued"
	BatchSpecWorkspaceStateProcessing BatchSpecWorkspaceState = "processing"
	BatchSpecWorkspaceStateErrored    BatchSpecWorkspaceState = "errored"
	BatchSpecWorkspaceStateFailed     BatchSpecWorkspaceState = "failed"
	BatchSpecWorkspaceStateCompleted  BatchSpecWorkspaceState = "completed"
)

// Valid returns true if the given BatchSpecWorkspaceState is valid.
func (s BatchSpecWorkspaceState) Valid() bool {
	switch s {
	case BatchSpecWorkspaceStatePending,
		BatchSpecWorkspaceStateQueued,
		BatchSpecWorkspaceStateProcessing,
		BatchSpecWorkspaceStateErrored,
		BatchSpecWorkspaceStateFailed,
		BatchSpecWorkspaceStateCompleted:
		return true
	default:
		return false
	}
}

// ToGraphQL returns the GraphQL representation of the worker state.
func (s BatchSpecWorkspaceState) ToGraphQL() string { return strings.ToUpper(string(s)) }

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
