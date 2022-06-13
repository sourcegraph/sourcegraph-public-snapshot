package basestore

import (
	"context"
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type oldTransactableHandle struct {
	dbutil.DB
	savepoints []*savepoint
	txOptions  sql.TxOptions
}

// NewHandleWithUntypedDB returns a new transactable database handle using the given database connection.
func NewHandleWithUntypedDB(db dbutil.DB, txOptions sql.TxOptions) TransactableHandle {
	return &oldTransactableHandle{DB: db, txOptions: txOptions}
}

// NewHandleWithDB returns a new transactable database handle using the given database connection.
func NewHandleWithDB(db *sql.DB, txOptions sql.TxOptions) TransactableHandle {
	return &dbHandle{DB: db, txOptions: txOptions}
}

// InTransaction returns true if the underlying database handle is in a transaction.
func (h *oldTransactableHandle) InTransaction() bool {
	db := tryUnwrap(h.DB)
	_, ok := db.(dbutil.Tx)
	return ok
}

// Transact returns a new transactional database handle whose methods operate within the context of
// a new transaction or a new savepoint. This method will return an error if the underlying connection
// cannot be interface upgraded to a TxBeginner.
//
// Note that it is not valid to call Transact or Done on the same database handle from distinct goroutines.
// Because we support properly nested transactions via savepoints, calling Transact from two different
// goroutines on the same handle will not be deterministic: either transaction could nest the other one,
// and calling Done in one goroutine may not finalize the expected unit of work.
func (h *oldTransactableHandle) Transact(ctx context.Context) (TransactableHandle, error) {
	db := tryUnwrap(h.DB)

	if h.InTransaction() {
		savepoint, err := newSavepoint(ctx, db)
		if err != nil {
			return nil, err
		}

		h.savepoints = append(h.savepoints, savepoint)
		return h, nil
	}

	tb, ok := db.(dbutil.TxBeginner)
	if !ok {
		return nil, ErrNotTransactable
	}

	tx, err := tb.BeginTx(ctx, &h.txOptions)
	if err != nil {
		return nil, err
	}

	return &oldTransactableHandle{DB: tx, txOptions: h.txOptions}, nil
}

// Done performs a commit or rollback of the underlying transaction/savepoint depending
// on the value of the error parameter. The resulting error value is a multierror containing
// the error parameter along with any error that occurs during commit or rollback of the
// transaction/savepoint. If the store does not wrap a transaction the original error value
// is returned unchanged.
func (h *oldTransactableHandle) Done(err error) error {
	db := tryUnwrap(h.DB)

	if n := len(h.savepoints); n > 0 {
		var savepoint *savepoint
		savepoint, h.savepoints = h.savepoints[n-1], h.savepoints[:n-1]

		if err == nil {
			return savepoint.Commit()
		}
		return combineErrors(err, savepoint.Rollback())
	}

	tx, ok := db.(dbutil.Tx)
	if !ok {
		return err
	}

	if err == nil {
		return tx.Commit()
	}
	return combineErrors(err, tx.Rollback())
}

// tryUnwrap attempts to unwrap a dbutil.DB into a child dbutil.DB.
// This is necessary because for transactions, we do interface assertions
// on the concrete type, but these interface assertions will fail if dbutil.DB
// is not of the concrete type *sql.DB or *sql.Tx. With types like database.db,
// which implement dbutil.DB by embedding the interface, this is problematic.
// Eventually, this should go away once dbutil.DB is subsumed by database.DB.
func tryUnwrap(db dbutil.DB) dbutil.DB {
	if unwrapper, ok := db.(dbutil.Unwrapper); ok {
		return unwrapper.Unwrap()
	}
	return db
}

// combineErrors returns a multierror containing all fo the non-nil error parameter values.
// This method should be used over multierror when it is not guaranteed that the original
// error was non-nil (errors.Append creates a non-nil error even if it is empty).
func combineErrors(errs ...error) (err error) {
	for _, e := range errs {
		if e != nil {
			if err == nil {
				err = e
			} else {
				err = errors.Append(err, e)
			}
		}
	}

	return err
}
