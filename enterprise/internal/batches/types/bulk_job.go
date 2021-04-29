package types

import (
	"time"
)

// BulkJobState defines the possible states of a bulk job.
type BulkJobState string

// BulkJobState constants.
const (
	BulkJobStateProcessing BulkJobState = "PROCESSING"
	BulkJobStateFailed     BulkJobState = "FAILED"
	BulkJobStateCompleted  BulkJobState = "COMPLETED"
)

// Valid returns true if the given BulkJobState is valid.
func (s BulkJobState) Valid() bool {
	switch s {
	case BulkJobStateProcessing,
		BulkJobStateFailed,
		BulkJobStateCompleted:
		return true
	default:
		return false
	}
}

// BulkJob represents a virtual entity of a bulk job, as represented in the database.
type BulkJob struct {
	ID string
	// DBID is only used internally for pagination. Don't make any assumptions
	// about this field.
	DBID       int64
	Type       ChangesetJobType
	State      BulkJobState
	Progress   float64
	CreatedAt  time.Time
	FinishedAt time.Time
}

// BulkJobError represents an error on a changeset that occurred within a bulk
// job while executing.
type BulkJobError struct {
	ChangesetID int64
	Error       string
}
