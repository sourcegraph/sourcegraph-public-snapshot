package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type EnterpriseDB interface {
	database.DB
	CodeMonitors() CodeMonitorStore
	Perms() PermsStore
}

func NewEnterpriseDB(db database.DB) EnterpriseDB {
	// If the underlying type already implements EnterpriseDB,
	// return that rather than wrapping it. This enables us to
	// pass a mock EnterpriseDB through as a database.DB, and
	// avoid overwriting its mocked methods by wrapping it.
	if edb, ok := db.(EnterpriseDB); ok {
		return edb
	}
	return &enterpriseDB{db}
}

type enterpriseDB struct {
	database.DB
}

func (edb *enterpriseDB) CodeMonitors() CodeMonitorStore {
	return &codeMonitorStore{Store: basestore.NewWithHandle(edb.Handle()), now: time.Now}
}

func (edb *enterpriseDB) Perms() PermsStore {
	return &permsStore{Store: basestore.NewWithHandle(edb.Handle()), clock: time.Now}
}

type InsightsDB interface {
	dbutil.DB
	basestore.ShareableStore

	Transact(context.Context) (InsightsDB, error)
	Done(error) error
}

func NewInsightsDB(inner *sql.DB) InsightsDB {
	return &insightsDB{basestore.NewWithHandle(basestore.NewHandleWithDB(inner, sql.TxOptions{}))}
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

func (d *insightsDB) Unwrap() dbutil.DB {
	// Recursively unwrap in case we ever call `NewInsightsDB()` with an `InsightsDB`
	if unwrapper, ok := d.Handle().(dbutil.Unwrapper); ok {
		return unwrapper.Unwrap()
	}
	return d.Handle()
}

func (d *insightsDB) QueryContext(ctx context.Context, q string, args ...any) (*sql.Rows, error) {
	return d.Handle().QueryContext(ctx, q, args...)
}

func (d *insightsDB) ExecContext(ctx context.Context, q string, args ...any) (sql.Result, error) {
	return d.Handle().ExecContext(ctx, q, args...)

}

func (d *insightsDB) QueryRowContext(ctx context.Context, q string, args ...any) *sql.Row {
	return d.Handle().QueryRowContext(ctx, q, args...)
}
