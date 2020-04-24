package db

import (
	"context"
	"database/sql"

	"github.com/hashicorp/go-multierror"
	"github.com/keegancsmith/sqlf"
)

// TxCloser is a convenience wrapper for closing SQL transactions.
type TxCloser interface {
	// CloseTx commits the transaction on a nil error value and performs a rollback
	// otherwise. If an error occurs during commit or rollback of the transaction,
	// the error is added to the resulting error value.
	CloseTx(err error) error
}

type txCloser struct {
	tx *sql.Tx
}

func (txc *txCloser) CloseTx(err error) error {
	return closeTx(txc.tx, err)
}

func closeTx(tx *sql.Tx, err error) error {
	if err != nil {
		if rollErr := tx.Rollback(); rollErr != nil {
			err = multierror.Append(err, rollErr)
		}
		return err
	}

	return tx.Commit()
}

type transactionWrapper struct {
	tx *sql.Tx
}

// query performs QueryContext on the underlying transaction.
func (tw *transactionWrapper) query(ctx context.Context, query *sqlf.Query) (*sql.Rows, error) {
	return tw.tx.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
}

// queryRow performs QueryRow on the underlying transaction.
func (tw *transactionWrapper) queryRow(ctx context.Context, query *sqlf.Query) *sql.Row {
	return tw.tx.QueryRowContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
}

// exec performs Exec on the underlying transaction.
func (tw *transactionWrapper) exec(ctx context.Context, query *sqlf.Query) (sql.Result, error) {
	return tw.tx.ExecContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
}
