package workerutil

import (
	"context"
)

// Record is a generic interface for record conforming to the requirements of the store.
type Record interface {
	// RecordID returns the integer primary key of the record.
	RecordID() int
	// RecordUID returns a UID of the record, of which the format is defined by the concrete type.
	RecordUID() string
}

// Store is the persistence layer for the workerutil package that handles worker-side operations.
type Store[T Record] interface {
	// QueuedCount returns the number of records in the queued state.
	QueuedCount(ctx context.Context) (int, error)

	// Dequeue selects a record for processing. Any extra arguments supplied will be used in accordance with the
	// concrete persistence layer (e.g. additional SQL conditions for a database layer). This method returns a boolean
	// flag indicating the existence of a processable record.
	Dequeue(ctx context.Context, workerHostname string, extraArguments any) (T, bool, error)

	// Heartbeat updates last_heartbeat_at of all the given jobs, when they're processing. All IDs of records that were
	// touched are returned. Additionally, jobs in the working set that are flagged as to be canceled are returned.
	Heartbeat(ctx context.Context, jobIDs []string) (knownIDs, cancelIDs []string, err error)

	// MarkComplete attempts to update the state of the record to complete. This method returns a boolean flag indicating
	// if the record was updated.
	MarkComplete(ctx context.Context, rec T) (bool, error)

	// MarkErrored attempts to update the state of the record to errored. This method returns a boolean flag indicating
	// if the record was updated.
	MarkErrored(ctx context.Context, rec T, failureMessage string) (bool, error)

	// MarkFailed attempts to update the state of the record to failed. This method returns a boolean flag indicating
	// if the record was updated.
	MarkFailed(ctx context.Context, rec T, failureMessage string) (bool, error)
}
