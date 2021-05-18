package types

import (
	"time"
)

// BulkOperationState defines the possible states of a bulk operation.
type BulkOperationState string

// BulkOperationState constants.
const (
	BulkOperationStateProcessing BulkOperationState = "PROCESSING"
	BulkOperationStateFailed     BulkOperationState = "FAILED"
	BulkOperationStateCompleted  BulkOperationState = "COMPLETED"
)

// Valid returns true if the given BulkOperationState is valid.
func (s BulkOperationState) Valid() bool {
	switch s {
	case BulkOperationStateProcessing,
		BulkOperationStateFailed,
		BulkOperationStateCompleted:
		return true
	default:
		return false
	}
}

// BulkOperation represents a virtual entity of a bulk operation, as represented in the database.
type BulkOperation struct {
	ID string
	// DBID is only used internally for pagination. Don't make any assumptions
	// about this field.
	DBID           int64
	Type           ChangesetJobType
	State          BulkOperationState
	Progress       float64
	UserID         int32
	ChangesetCount int32
	CreatedAt      time.Time
	FinishedAt     time.Time
}

// BulkOperationError represents an error on a changeset that occurred within a bulk
// job while executing.
type BulkOperationError struct {
	ChangesetID int64
	Error       string
}
