package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/keegancsmith/sqlf"
)

var ErrNoSavepoint = errors.New("no savepoint defined")
var ErrJobNotFinalized = errors.New("job not finalized")

// JobHandle wraps a transaction used by the upload converter. This transaction marks the upload as
// ineligible for a dequeue to other worker processes. All updates to the database while this record
// is being processes should happen through the JobHandle's transaction, which must be explicitly
// closed (via CloseTx) at the end of processing by the caller.
//
// Before closing the transaction, the caller MUST be sure to have invoked either MarkComplete or
// MarkErrored. Failure to do so will result in an upload record that will be indefinitely selected
// for processing.
type JobHandle interface {
	TxCloser

	// Tx retrieves the underlying transaction object. This should be passed to all method of DB
	// to ensure that if the job processing fails there are no externally visible changes to the
	// database.
	Tx() *sql.Tx

	// Savepoint creates a named position in the transaction from which all additional work can
	// be discarded.
	Savepoint() error

	// RollbackToLastSavepoint throws away all the work on the underlying transaction since the
	// last call to Savepoint. This method returns an error if there is no live savepoint in this
	// transaction.
	RollbackToLastSavepoint() error

	// MarkComplete updates the state of the upload to complete.
	MarkComplete() error

	// MarkErrored updates the state of the upload to errored and updates the failure summary data.
	MarkErrored(failureSummary, failureStacktrace string) error
}

type jobHandleImpl struct {
	ctx             context.Context
	id              int
	tw              *transactionWrapper
	txCloser        TxCloser
	savepoints      []string
	marked          bool
	markedSavepoint string
}

var _ JobHandle = &jobHandleImpl{}

// CloseTx commits the transaction on a nil error value and performs a rollback
// otherwise. If an error occurs during commit or rollback of the transaction,
// the error is added to the resulting error value. If neither MarkComplete or
// MarkErrored were invoked, this method returns an error.
func (h *jobHandleImpl) CloseTx(err error) error {
	if err == nil && !h.marked {
		err = ErrJobNotFinalized
	}

	return h.txCloser.CloseTx(err)
}

// Tx retrieves the underlying transaction object. This should be passed to all method of DB
// to ensure that if the job processing fails there are no externally visible changes to the
// database.
func (h *jobHandleImpl) Tx() *sql.Tx {
	return h.tw.tx
}

// Savepoint creates a named position in the transaction from which all additional work can
// be discarded.
func (h *jobHandleImpl) Savepoint() error {
	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	savepointID := fmt.Sprintf("sp_%s", strings.ReplaceAll(id.String(), "-", "_"))
	h.savepoints = append(h.savepoints, savepointID)
	// Unfortunately, it's a syntax error to supply this as a param
	_, err = h.tw.exec(h.ctx, sqlf.Sprintf("SAVEPOINT "+savepointID))
	return err
}

// RollbackToLastSavepoint throws away all the work on the underlying transaction since the
// last call to Savepoint. This method returns an error if there is no live savepoint in this
// transaction.
func (h *jobHandleImpl) RollbackToLastSavepoint() error {
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
	_, err := h.tw.exec(h.ctx, sqlf.Sprintf("ROLLBACK TO SAVEPOINT "+savepointID))
	return err
}

// MarkComplete updates the state of the upload to complete.
func (h *jobHandleImpl) MarkComplete() (err error) {
	query := `
		UPDATE lsif_uploads
		SET state = 'completed', finished_at = now()
		WHERE id = %s
	`

	h.mark()
	_, err = h.tw.exec(h.ctx, sqlf.Sprintf(query, h.id))
	return err
}

// MarkErrored updates the state of the upload to errored and updates the failure summary data.
func (h *jobHandleImpl) MarkErrored(failureSummary, failureStacktrace string) (err error) {
	query := `
		UPDATE lsif_uploads
		SET state = 'errored', finished_at = now(), failure_summary = %s, failure_stacktrace = %s
		WHERE id = %s
	`

	h.mark()
	_, err = h.tw.exec(h.ctx, sqlf.Sprintf(query, failureSummary, failureStacktrace, h.id))
	return err
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
