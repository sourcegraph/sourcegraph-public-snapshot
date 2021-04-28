package types

import "time"

// BulkJob represents a virtual entity of a bulk job, as represented in the database.
type BulkJob struct {
	ID         string
	Type       ChangesetJobType
	State      ReconcilerState
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
