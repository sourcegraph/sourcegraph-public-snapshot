package store

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/sqliteutil"
)

// ExecableDB extends the dbutil.DB interface with ExecContext. This interface is
// valid as SQLite (and Postgres) allow transactional DDLs, thus both query and
// exec methods can be performed on raw connections as well as transactional ones.
// Part of the SQLite migration requires updating the current database's schema,
// which can only be performed with Exec methods.
type ExecableDB interface {
	dbutil.DB
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

// Store wraps a SQLite connection.
type Store struct {
	db ExecableDB
}

var _ sqliteutil.Execable = &Store{}

// Open creates a new SQLite connection and wraps it in a store.
func Open(filename string) (*Store, func() error, error) {
	db, err := sqlx.Open("sqlite3_with_pcre", filename)
	if err != nil {
		return nil, nil, err
	}

	return &Store{db: db}, db.Close, nil
}

// Query calls into the underlying connection.
func (s *Store) Query(ctx context.Context, query *sqlf.Query) (*sql.Rows, error) {
	return s.db.QueryContext(ctx, query.Query(sqlf.SimpleBindVar), query.Args()...)
}

// Exec calls into the underlying connection.
func (s *Store) Exec(ctx context.Context, query *sqlf.Query) error {
	_, err := s.db.ExecContext(ctx, query.Query(sqlf.SimpleBindVar), query.Args()...)
	return err
}

// TODO(efritz) - rework ExecContext interface

// ExecContext calls into the underlying connection. This method is here so that the
// Store can be used as an argument to a SQLite batch inserter.
func (s *Store) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return s.db.ExecContext(ctx, query, args...)
}
