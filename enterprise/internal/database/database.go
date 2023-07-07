package database

import (
	"context"
	"database/sql"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type InsightsDB interface {
	dbutil.DB
	basestore.ShareableStore

	Transact(context.Context) (InsightsDB, error)
	Done(error) error
}

func NewInsightsDB(inner *sql.DB, logger log.Logger) InsightsDB {
	return &insightsDB{basestore.NewWithHandle(basestore.NewHandleWithDB(logger, inner, sql.TxOptions{}))}
}

func NewInsightsDBWith(other basestore.ShareableStore) InsightsDB {
	return &insightsDB{basestore.NewWithHandle(other.Handle())}
}

type insightsDB struct {
	*basestore.Store
}

func (d *insightsDB) Transact(ctx context.Context) (InsightsDB, error) {
	tx, err := d.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	return &insightsDB{tx}, nil
}

func (d *insightsDB) Done(err error) error {
	return d.Store.Done(err)
}

func (d *insightsDB) QueryContext(ctx context.Context, q string, args ...any) (*sql.Rows, error) {
	return d.Handle().QueryContext(dbconn.SkipFrameForQuerySource(ctx), q, args...)
}

func (d *insightsDB) ExecContext(ctx context.Context, q string, args ...any) (sql.Result, error) {
	return d.Handle().ExecContext(dbconn.SkipFrameForQuerySource(ctx), q, args...)
}

func (d *insightsDB) QueryRowContext(ctx context.Context, q string, args ...any) *sql.Row {
	return d.Handle().QueryRowContext(dbconn.SkipFrameForQuerySource(ctx), q, args...)
}
