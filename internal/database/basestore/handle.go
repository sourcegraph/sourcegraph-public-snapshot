package basestore

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// TransactableHandle is a wrapper around a database connection that provides
// nested transactions through registration and finalization of savepoints. A
// transactable database handler can be shared by multiple stores.
type TransactableHandle interface {
	dbutil.DB

	// InTransaction returns whether the handle represents a handle to a transaction.
	InTransaction() bool

	// Transact returns a new transactional database handle whose methods operate within the context of
	// a new transaction or a new savepoint.
	//
	// Note that it is not safe to use transactions from multiple goroutines.
	Transact(context.Context) (TransactableHandle, error)

	// Done performs a commit or rollback of the underlying transaction/savepoint depending
	// on the value of the error parameter. The resulting error value is a multierror containing
	// the error parameter along with any error that occurs during commit or rollback of the
	// transaction/savepoint. If the store does not wrap a transaction the original error value
	// is returned unchanged.
	Done(error) error
}

var (
	_ TransactableHandle = (*dbHandle)(nil)
	_ TransactableHandle = (*txHandle)(nil)
	_ TransactableHandle = (*savepointHandle)(nil)
)

// NewHandleWithDB returns a new transactable database handle using the given database connection.
func NewHandleWithDB(db *sql.DB, txOptions sql.TxOptions) TransactableHandle {
	return &dbHandle{DB: db, txOptions: txOptions}
}

// NewHandleWithTx returns a new transactable database handle using the given transaction.
func NewHandleWithTx(tx *sql.Tx, txOptions sql.TxOptions) TransactableHandle {
	return &txHandle{lockingTx: &lockingTx{tx: tx}, txOptions: txOptions}
}

type dbHandle struct {
	*sql.DB
	txOptions sql.TxOptions
}

func (h *dbHandle) InTransaction() bool {
	return false
}

func (h *dbHandle) Transact(ctx context.Context) (TransactableHandle, error) {
	tx, err := h.DB.BeginTx(ctx, &h.txOptions)
	if err != nil {
		return nil, err
	}
	return &txHandle{lockingTx: &lockingTx{tx: tx}, txOptions: h.txOptions}, nil
}

func (h *dbHandle) Done(err error) error {
	return errors.Append(err, ErrNotInTransaction)
}

type txHandle struct {
	*lockingTx
	txOptions sql.TxOptions
}

func (h *txHandle) InTransaction() bool {
	return true
}

func (h *txHandle) Transact(ctx context.Context) (TransactableHandle, error) {
	savepointID, err := newTxSavepoint(ctx, h.lockingTx)
	if err != nil {
		return nil, err
	}

	return &savepointHandle{lockingTx: h.lockingTx, savepointID: savepointID}, nil
}

func (h *txHandle) Done(err error) error {
	if err == nil {
		return h.Commit()
	}
	return errors.Append(err, h.Rollback())
}

type savepointHandle struct {
	*lockingTx
	savepointID string
}

func (h *savepointHandle) InTransaction() bool {
	return true
}

func (h *savepointHandle) Transact(ctx context.Context) (TransactableHandle, error) {
	savepointID, err := newTxSavepoint(ctx, h.lockingTx)
	if err != nil {
		return nil, err
	}

	return &savepointHandle{lockingTx: h.lockingTx, savepointID: savepointID}, nil
}

func (h *savepointHandle) Done(err error) error {
	if err == nil {
		_, execErr := h.ExecContext(context.Background(), fmt.Sprintf(commitSavepointQuery, h.savepointID))
		return execErr
	}

	_, execErr := h.ExecContext(context.Background(), fmt.Sprintf(rollbackSavepointQuery, h.savepointID))
	return errors.Append(err, execErr)
}

const (
	savepointQuery         = "SAVEPOINT %s"
	commitSavepointQuery   = "RELEASE %s"
	rollbackSavepointQuery = "ROLLBACK TO %s"
)

func newTxSavepoint(ctx context.Context, tx *lockingTx) (string, error) {
	savepointID, err := makeSavepointID()
	if err != nil {
		return "", err
	}

	_, err = tx.ExecContext(ctx, fmt.Sprintf(savepointQuery, savepointID))
	if err != nil {
		return "", err
	}

	return savepointID, nil
}

func makeSavepointID() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("sp_%s", strings.ReplaceAll(id.String(), "-", "_")), nil
}

var ErrConcurrentTransactionAccess = errors.New("transaction used concurrently")

// lockingTx wraps a *sql.Tx with a mutex, and reports when a caller tries to
// use the transaction concurrently. Since using a transaction concurrently is
// unsafe, we want to catch these issues. If lockingTx detects that a
// transaction is being used concurrently, it will return
// ErrConcurrentTransactionAccess and log an error.
type lockingTx struct {
	tx *sql.Tx
	mu sync.Mutex
}

func (t *lockingTx) lock() error {
	if !t.mu.TryLock() {
		err := errors.WithStack(ErrConcurrentTransactionAccess)
		log.Scoped("internal", "database").Error("transaction used concurrently", log.Error(err))
		return err
	}
	return nil
}

func (t *lockingTx) unlock() {
	t.mu.Unlock()
}

func (t *lockingTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if err := t.lock(); err != nil {
		return nil, err
	}
	defer t.unlock()

	return t.tx.ExecContext(ctx, query, args...)
}

func (t *lockingTx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if err := t.lock(); err != nil {
		return nil, err
	}
	defer t.unlock()

	return t.tx.QueryContext(ctx, query, args...)
}

func (t *lockingTx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	if err := t.lock(); err != nil {
		// XXX(camdencheek): There is no way to construct a row that has an
		// error, so in order for this to return an error, we'll need to have
		// it return an interface. Instead, we attempt to acquire the lock
		// to make this at least be safer if we can't report the error other
		// than through logs.
		t.mu.Lock()
	}
	defer t.unlock()

	return t.tx.QueryRowContext(ctx, query, args...)
}

func (t *lockingTx) Commit() error {
	if err := t.lock(); err != nil {
		return err
	}
	defer t.unlock()

	return t.tx.Commit()
}

func (t *lockingTx) Rollback() error {
	if err := t.lock(); err != nil {
		return err
	}
	defer t.unlock()

	return t.tx.Rollback()
}
