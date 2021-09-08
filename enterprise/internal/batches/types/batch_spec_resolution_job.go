package types

import (
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

// BatchSpecResolutionJobState defines the possible states of a changeset job.
type BatchSpecResolutionJobState string

// BatchSpecResolutionJobState constants.
const (
	BatchSpecResolutionJobStateQueued     BatchSpecResolutionJobState = "QUEUED"
	BatchSpecResolutionJobStateProcessing BatchSpecResolutionJobState = "PROCESSING"
	BatchSpecResolutionJobStateErrored    BatchSpecResolutionJobState = "ERRORED"
	BatchSpecResolutionJobStateFailed     BatchSpecResolutionJobState = "FAILED"
	BatchSpecResolutionJobStateCompleted  BatchSpecResolutionJobState = "COMPLETED"
)

// Valid returns true if the given BatchSpecResolutionJobState is valid.
func (s BatchSpecResolutionJobState) Valid() bool {
	switch s {
	case BatchSpecResolutionJobStateQueued,
		BatchSpecResolutionJobStateProcessing,
		BatchSpecResolutionJobStateErrored,
		BatchSpecResolutionJobStateFailed,
		BatchSpecResolutionJobStateCompleted:
		return true
	default:
		return false
	}
}

// ToDB returns the database representation of the worker state. That's
// needed because we want to use UPPERCASE in the application and GraphQL layer,
// but need to use lowercase in the database to make it work with workerutil.Worker.
func (s BatchSpecResolutionJobState) ToDB() string { return strings.ToLower(string(s)) }

type BatchSpecResolutionJob struct {
	ID int64

	BatchSpecID      int64
	AllowUnsupported bool
	AllowIgnored     bool

	// workerutil fields
	State           BatchSpecResolutionJobState
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

func (j *BatchSpecResolutionJob) RecordID() int {
	return int(j.ID)
}
