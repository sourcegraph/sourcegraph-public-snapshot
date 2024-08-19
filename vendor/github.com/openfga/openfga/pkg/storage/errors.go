package storage

import (
	"errors"
	"fmt"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"

	"github.com/openfga/openfga/pkg/tuple"
)

var (
	// ErrCollision is returned when an item already exists within the store.
	ErrCollision = errors.New("item already exists")

	// ErrInvalidContinuationToken is returned when the continuation token is invalid.
	ErrInvalidContinuationToken = errors.New("invalid continuation token")

	// ErrMismatchObjectType is returned when there is a type discrepancy between the requested
	// object in the ReadChanges API and the type indicated by the continuation token.
	ErrMismatchObjectType = errors.New("mismatched types in request and continuation token")

	// ErrInvalidWriteInput is returned when the tuple to be written
	// already existed or the tuple to be deleted did not exist.
	ErrInvalidWriteInput = errors.New("invalid write input")

	// ErrTransactionalWriteFailed is returned when two writes attempt to write the same tuple at the same time.
	ErrTransactionalWriteFailed = errors.New("transactional write failed due to conflict")

	// ErrExceededWriteBatchLimit is returned when MaxTuplesPerWrite is exceeded.
	ErrExceededWriteBatchLimit = errors.New("number of operations exceeded write batch limit")

	// ErrCancelled is returned when the request has been cancelled.
	ErrCancelled = errors.New("request has been cancelled")

	// ErrDeadlineExceeded is returned when the request's deadline is exceeded.
	ErrDeadlineExceeded = errors.New("request deadline exceeded")

	// ErrNotFound is returned when the object does not exist.
	ErrNotFound = errors.New("not found")
)

// ExceededMaxTypeDefinitionsLimitError constructs an error indicating that
// the maximum allowed limit for type definitions has been exceeded.
func ExceededMaxTypeDefinitionsLimitError(limit int) error {
	return fmt.Errorf("exceeded number of allowed type definitions: %d", limit)
}

// InvalidWriteInputError generates an error for invalid operations in a tuple store.
// This function is invoked when an attempt is made to write or delete a tuple with invalid conditions.
// Specifically, it addresses two scenarios:
// 1. Attempting to delete a non-existent tuple.
// 2. Attempting to write a tuple that already exists.
func InvalidWriteInputError(tk tuple.TupleWithoutCondition, operation openfgav1.TupleOperation) error {
	switch operation {
	case openfgav1.TupleOperation_TUPLE_OPERATION_DELETE:
		return fmt.Errorf(
			"cannot delete a tuple which does not exist: user: '%s', relation: '%s', object: '%s': %w",
			tk.GetUser(),
			tk.GetRelation(),
			tk.GetObject(),
			ErrInvalidWriteInput,
		)
	case openfgav1.TupleOperation_TUPLE_OPERATION_WRITE:
		return fmt.Errorf(
			"cannot write a tuple which already exists: user: '%s', relation: '%s', object: '%s': %w",
			tk.GetUser(),
			tk.GetRelation(),
			tk.GetObject(),
			ErrInvalidWriteInput,
		)
	default:
		return nil
	}
}
