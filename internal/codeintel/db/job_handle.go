package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
)

// JobHandle wraps a transaction used by the upload converter. This transaction marks the upload as
// ineligible for a dequeue to other worker processes. All updates to the database while this record
// is being processes should happen through the JobHandle's transaction, which must be explicitly
// closed (via CloseTx) at the end of processing by the caller.
//
// Before closing the transaction, the caller MUST be sure to have invoked either MarkComplete or
// MarkErrored. Failure to do so will result in an upload record that will be indefinitely selected
// for processing.
type JobHandle interface {
	// DB retrieves the underlying Database object, which wraps the transaction in which the
	// target upload is locked.
	DB() DB

	// Done closes the underlying transaction. If neither MarkComplete or MarkErrored were invoked,
	// this method returns an error.
	Done(err error) error

	// Savepoint creates a named position in the transaction from which all additional work can
	// be discarded.
	Savepoint(ctx context.Context) error

	// RollbackToLastSavepoint throws away all the work on the underlying transaction since the
	// last call to Savepoint. This method returns an error if there is no live savepoint in this
	// transaction.
	RollbackToLastSavepoint(ctx context.Context) error

	// MarkComplete updates the state of the upload to complete.
	MarkComplete(ctx context.Context) error

	// MarkErrored updates the state of the upload to errored and updates the failure summary data.
	MarkErrored(ctx context.Context, failureSummary, failureStacktrace string) error
}

// execer is the interface that the DB reference in jobHandleImpl must conform to. This allows the
// job handler to call any method in the working tranaction, and also call dbImpl's unexported
// exec method to run execute SQL.
type execer interface {
	DB
	exec(ctx context.Context, query *sqlf.Query) error
}

type jobHandleImpl struct {
	db              execer
	id              int
	savepoints      []string
	marked          bool
	markedSavepoint string
}

var _ JobHandle = &jobHandleImpl{}

// DB retrieves the underlying Database object, which wraps the transaction in which the
// target upload is locked.
func (h *jobHandleImpl) DB() DB {
	return h.db
}

// ErrJobNotFinalized occurs when the job handler's transaction is closed without finalizing the job.
var ErrJobNotFinalized = errors.New("job not finalized")

// Done closes the underlying transaction. If neither MarkComplete or MarkErrored were invoked,
// this method returns an error.
func (h *jobHandleImpl) Done(err error) error {
	if err == nil && !h.marked {
		err = multierror.Append(err, ErrJobNotFinalized)
	}

	return h.db.Done(err)
}

// Savepoint creates a named position in the transaction from which all additional work can
// be discarded.
func (h *jobHandleImpl) Savepoint(ctx context.Context) error {
	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	savepointID := fmt.Sprintf("sp_%s", strings.ReplaceAll(id.String(), "-", "_"))
	h.savepoints = append(h.savepoints, savepointID)
	// Unfortunately, it's a syntax error to supply this as a param
	return h.db.exec(ctx, sqlf.Sprintf("SAVEPOINT "+savepointID))
}

// ErrNoSavepoint occurs when there is no savepont to rollback to.
var ErrNoSavepoint = errors.New("no savepoint defined")

// RollbackToLastSavepoint throws away all the work on the underlying transaction since the
// last call to Savepoint. This method returns an error if there is no live savepoint in this
// transaction.
func (h *jobHandleImpl) RollbackToLastSavepoint(ctx context.Context) error {
	n := len(h.savepoints)
	if n == 0 {
		return ErrNoSavepoint
	}

	// Pop savepoint id to rollback
	savepointID := h.savepoints[n-1]
	h.savepoints = h.savepoints[:n-1]

	// Clear marked flag if we're rolling back the mark
	if savepointID == h.markedSavepoint {
		h.marked = false
		h.markedSavepoint = ""
	}

	// Perform rollback
	return h.db.exec(ctx, sqlf.Sprintf("ROLLBACK TO SAVEPOINT "+savepointID))
}

// MarkComplete updates the state of the upload to complete.
func (h *jobHandleImpl) MarkComplete(ctx context.Context) (err error) {
	h.mark()

	return h.db.exec(ctx, sqlf.Sprintf(`
		UPDATE lsif_uploads
		SET state = 'completed', finished_at = now()
		WHERE id = %s
	`, h.id))
}

// MarkErrored updates the state of the upload to errored and updates the failure summary data.
func (h *jobHandleImpl) MarkErrored(ctx context.Context, failureSummary, failureStacktrace string) (err error) {
	h.mark()

	return h.db.exec(ctx, sqlf.Sprintf(`
		UPDATE lsif_uploads
		SET state = 'errored', finished_at = now(), failure_summary = %s, failure_stacktrace = %s
		WHERE id = %s
	`, failureSummary, failureStacktrace, h.id))
}

func (h *jobHandleImpl) mark() {
	h.marked = true

	if len(h.savepoints) == 0 {
		h.markedSavepoint = ""
	} else {
		// Mark the current savepoint we're inside so we can unset
		// the marked flag if we later perform a rollback on error.
		h.markedSavepoint = h.savepoints[len(h.savepoints)-1]
	}
}
