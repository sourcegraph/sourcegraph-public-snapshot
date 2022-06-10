package basestore

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type ITransactableHandle interface {
	DB() dbutil.DB
	InTransaction() bool
	Transact(context.Context) (ITransactableHandle, error)
	Done(error) error
}

var (
	_ ITransactableHandle = (*dbHandle)(nil)
	_ ITransactableHandle = (*txHandle)(nil)
	_ ITransactableHandle = (*savepointHandle)(nil)
)

type dbHandle struct {
	db        *sql.DB
	txOptions sql.TxOptions
}

func (h *dbHandle) DB() dbutil.DB {
	return h.db
}

func (h *dbHandle) InTransaction() bool {
	return false
}

func (h *dbHandle) Transact(ctx context.Context) (ITransactableHandle, error) {
	tx, err := h.db.BeginTx(ctx, &h.txOptions)
	if err != nil {
		return nil, err
	}
	return &txHandle{tx: tx, txOptions: h.txOptions}, nil
}

func (h *dbHandle) Done(err error) error {
	return errors.Append(err, ErrNotInTransaction)
}

type txHandle struct {
	tx        *sql.Tx
	txOptions sql.TxOptions
}

func (h *txHandle) DB() dbutil.DB {
	return h.tx
}

func (h *txHandle) InTransaction() bool {
	return true
}

func (h *txHandle) Transact(ctx context.Context) (ITransactableHandle, error) {
	savepointID, err := newTxSavepoint(ctx, h.tx)
	if err != nil {
		return nil, err
	}

	return &savepointHandle{tx: h.tx, savepointID: savepointID}, nil
}

func (h *txHandle) Done(err error) error {
	if err == nil {
		return h.tx.Commit()
	}
	return errors.Append(err, h.tx.Rollback())
}

type savepointHandle struct {
	tx          *sql.Tx
	savepointID string
}

func (h *savepointHandle) DB() dbutil.DB {
	return h.tx
}

func (h *savepointHandle) InTransaction() bool {
	return true
}

func (h *savepointHandle) Transact(ctx context.Context) (ITransactableHandle, error) {
	savepointID, err := newTxSavepoint(ctx, h.tx)
	if err != nil {
		return nil, err
	}

	return &savepointHandle{tx: h.tx, savepointID: savepointID}, nil
}

func (h *savepointHandle) Done(err error) error {
	if err == nil {
		_, execErr := h.tx.Exec(fmt.Sprintf(commitSavepointQuery, h.savepointID))
		return execErr
	}

	_, execErr := h.tx.Exec(fmt.Sprintf(rollbackSavepointQuery, h.savepointID))
	return errors.Append(err, execErr)
}

func newTxSavepoint(ctx context.Context, tx *sql.Tx) (string, error) {
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
