package stores

import (
	"context"
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type CodeIntelDB interface {
	dbutil.DB
	basestore.ShareableStore

	Transact(context.Context) (CodeIntelDB, error)
	Done(error) error
}

func NewCodeIntelDB(inner *sql.DB) CodeIntelDB {
	return &codeIntelDB{basestore.NewWithHandle(basestore.NewHandleWithDB(inner, sql.TxOptions{}))}
}

func NewCodeIntelDBWith(other basestore.ShareableStore) CodeIntelDB {
	return &codeIntelDB{basestore.NewWithHandle(other.Handle())}
}

type codeIntelDB struct {
	*basestore.Store
}

func (d *codeIntelDB) Transact(ctx context.Context) (CodeIntelDB, error) {
	tx, err := d.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	return &codeIntelDB{tx}, nil
}

func (d *codeIntelDB) Done(err error) error {
	return d.Store.Done(err)
}

func (d *codeIntelDB) Unwrap() dbutil.DB {
	// Recursively unwrap in case we ever call `NewInsightsDB()` with an `InsightsDB`
	if unwrapper, ok := d.Handle().(dbutil.Unwrapper); ok {
		return unwrapper.Unwrap()
	}
	return d.Handle()
}

func (d *codeIntelDB) QueryContext(ctx context.Context, q string, args ...any) (*sql.Rows, error) {
	return d.Handle().QueryContext(ctx, q, args...)
}

func (d *codeIntelDB) ExecContext(ctx context.Context, q string, args ...any) (sql.Result, error) {
	return d.Handle().ExecContext(ctx, q, args...)

}

func (d *codeIntelDB) QueryRowContext(ctx context.Context, q string, args ...any) *sql.Row {
	return d.Handle().QueryRowContext(ctx, q, args...)
}
