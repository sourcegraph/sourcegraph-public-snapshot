package stores

import (
	"context"
	"database/sql"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type CodeIntelDB interface {
	dbutil.DB
	basestore.ShareableStore

	Transact(context.Context) (CodeIntelDB, error)
	Done(error) error
}

func NewCodeIntelDB(logger log.Logger, inner *sql.DB) CodeIntelDB {
	return &codeIntelDB{basestore.NewWithHandle(basestore.NewHandleWithDB(logger, inner, sql.TxOptions{}))}
}

func NewCodeIntelDBWith(other basestore.ShareableStore) CodeIntelDB {
	return &codeIntelDB{basestore.NewWithHandle(other.Handle())}
}

type codeIntelDB struct {
	*basestore.Store
}

func (d *codeIntelDB) Transact(ctx context.Context) (CodeIntelDB, error) {
	tx, err := d.Store.Transact(ctx)
	return &codeIntelDB{tx}, err
}

func (db *codeIntelDB) Done(err error) error {
	return db.Store.Done(err)
}

func (db *codeIntelDB) QueryContext(ctx context.Context, q string, args ...any) (*sql.Rows, error) {
	return db.Handle().QueryContext(ctx, q, args...)
}

func (db *codeIntelDB) ExecContext(ctx context.Context, q string, args ...any) (sql.Result, error) {
	return db.Handle().ExecContext(ctx, q, args...)
}

func (db *codeIntelDB) QueryRowContext(ctx context.Context, q string, args ...any) *sql.Row {
	return db.Handle().QueryRowContext(ctx, q, args...)
}
