package stores

import (
	"context"
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

type CodeIntelDB interface {
	basestore.ShareableStore
	basestore.TransactableHandle
}

type codeIntelDB struct {
	*basestore.Store
}

func NewCodeIntelDB(inner *sql.DB) CodeIntelDB {
	return &codeIntelDB{basestore.NewWithHandle(basestore.NewHandleWithDB(inner, sql.TxOptions{}))}
}

func (db *codeIntelDB) Transact(ctx context.Context) (basestore.TransactableHandle, error) {
	return db.Handle().Transact(ctx)
}

func (db *codeIntelDB) Done(err error) error {
	return db.Handle().Done(err)
}

func (db *codeIntelDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.Handle().QueryContext(ctx, query, args...)
}

func (db *codeIntelDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return db.Handle().ExecContext(ctx, query, args...)
}

func (db *codeIntelDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return db.Handle().QueryRowContext(ctx, query, args...)
}
