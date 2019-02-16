package repos

import (
	"context"
	"database/sql"
	"io"
	"strings"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
)

// A Store exposes methods to read and write persistent repositories.
type Store interface {
	ListRepos(ctx context.Context) ([]*Repo, error)
	UpsertRepos(ctx context.Context, repos ...*Repo) error
}

// A Transactor can initialise and return a TxStore which operates
// within the context of a transaction.
type Transactor interface {
	Transact(context.Context) (TxStore, error)
}

// A TxStore is a Store that operates within the context of a transaction.
// Done should be called to terminate the underlying transaction. Once a TxStore
// is done, it can't be used further. Initiate a new one from its original
// Transactor.
type TxStore interface {
	Store
	Done(...*error)
}

// DBStore implements the Store interface for reading and writing repos directly
// from the Postgres database.
type DBStore struct {
	db     DB
	kinds  []string // which source kinds to list
	txOpts sql.TxOptions
}

// NewDBStore instantiates and returns a new DBStore with prepared statements.
func NewDBStore(ctx context.Context, db DB, kinds []string, txOpts sql.TxOptions) *DBStore {
	return &DBStore{db: db, kinds: kinds, txOpts: txOpts}
}

// Transact returns a TxStore whose methods operate within the context of a transaction.
// This method will return an error if the underlying DB cannot be interface upgraded
// to a TxBeginner.
func (s *DBStore) Transact(ctx context.Context) (TxStore, error) {
	if _, ok := s.db.(Tx); ok { // Already in a Tx.
		return s, nil
	}

	tb, ok := s.db.(TxBeginner)
	if !ok { // Not a Tx nor a TxBeginner, error.
		return nil, errors.New("dbstore: not transactable")
	}

	tx, err := tb.BeginTx(ctx, &s.txOpts)
	if err != nil {
		return nil, errors.Wrap(err, "dbstore: BeginTx")
	}

	return &DBStore{
		db:     tx,
		txOpts: s.txOpts,
	}, nil
}

// Done terminates the underlying Tx in a DBStore either by committing or rolling
// back based on the value pointed to by the first given error pointer.
// It's a no-op if the `DBStore` is not operating within a transaction,
// which can only be done via `BeginTxStore`.
//
// When the error value pointed to by the first given `err` is nil, or when no error
// pointer is given, the transaction is commited. Otherwise, it's rolled-back.
func (s *DBStore) Done(errs ...*error) {
	switch tx, ok := s.db.(Tx); {
	case !ok:
		return
	case len(errs) == 0:
		_ = tx.Commit()
	case errs[0] != nil && *errs[0] != nil:
		_ = tx.Rollback()
	default:
		_ = tx.Commit()
	}
}

// ListRepos lists all stored repos having any of the configured source kinds (via the constructor)
// as captured by the external_service_type column AND repos belonging to some source.
func (s DBStore) ListRepos(ctx context.Context) (repos []*Repo, err error) {
	return repos, s.paginate(ctx, &repos, listReposQuery(s.kinds))
}

const listReposQueryFmtstr = `
SELECT id, name, description, language, created_at, updated_at, deleted_at,
  external_id, external_service_type, external_service_id, enabled, archived, fork
FROM repo WHERE id > %s AND %s AND deleted_at IS NULL ORDER BY id ASC LIMIT %s`

// a paginatedQuery returns a query with the given pagination
// parameters
type paginatedQuery func(cursor, limit int64) *sqlf.Query

func (s DBStore) paginate(ctx context.Context, repos *[]*Repo, q paginatedQuery) (err error) {
	var cursor, next int64 = -1, 0
	for cursor != next && err == nil {
		cursor = next
		if err = s.page(ctx, q(cursor, 500), repos); len(*repos) > 0 {
			next = int64((*repos)[len(*repos)-1]._ID)
		}
	}
	return err

}

func (s DBStore) page(ctx context.Context, q *sqlf.Query, repos *[]*Repo) (err error) {
	rows, err := s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}
	return scanAll(rows, func(sc scanner) error {
		var r Repo
		if err = scanRepo(&r, sc); err != nil {
			return err
		}
		*repos = append(*repos, &r)
		return nil
	})
}

func listReposQuery(kinds []string) paginatedQuery {
	qs := make([]*sqlf.Query, 0, len(kinds))
	for _, kind := range kinds {
		qs = append(qs, sqlf.Sprintf("%s", strings.ToUpper(kind)))
	}
	q := sqlf.Join(qs, ",")

	return func(cursor, limit int64) *sqlf.Query {
		return sqlf.Sprintf(
			listReposQueryFmtstr,
			cursor,
			q,
			limit,
		)
	}
}

const updateReposQueryFmtstr = `
WITH updated (
	UPDATE repo r
	SET
	  name        = updated.name,
	  description = updated.description,
	  language    = updated.language,
	  updated_at  = updated.updated_at,
	  deleted_at  = updated.deleted_at,
	  archived    = updated.archived,
	  fork        = updated.fork,
	  sources     = updated.sources,
	  metadata    = updated.metadata,
	FROM
)
FROM updated
WHERE id = %s
OR name = %s
OR (external_service_id = %s AND external_id = %s)
`

// UpsertRepos updates or inserts the given repos in the Sourcegraph repository store.
// The _ID field of each given Repo is set on inserts.
func (s *DBStore) UpsertRepos(ctx context.Context, repos ...*Repo) (err error) {
	if len(repos) == 0 {
		return nil
	}

	// TODO: Figure out how to do batch update and / or insert in a single query.
	// UPSERT doesn't work with multiple constraints (e.g. unique name AND unique external_id)
	// - https://tapoueh.org/blog/2013/03/batch-update/
	// - https://tapoueh.org/blog/2018/07/batch-updates-and-concurrency/
	panic("not implemented")
}

var upsertRepoColumnNames = []string{
	"name",
	"description",
	"language",
	"uri",
	"created_at",
	"updated_at",
	"deleted_at",
	"external_id",
	"external_service_type",
	"external_service_id",
	"enabled",
	"archived",
	"fork",
}

func upsertRepoColumns(r *Repo) []interface{} {
	return []interface{}{
		r.Name,
		r.Description,
		r.Language,
		"", // URI
		timeColumn(r.CreatedAt),
		timeColumn(r.UpdatedAt),
		timeColumn(r.DeletedAt),
		r.ExternalRepo.ID,
		r.ExternalRepo.ServiceType,
		r.ExternalRepo.ServiceID,
		r.Enabled,
		r.Archived,
		r.Fork,
	}
}

func timeColumn(t time.Time) interface{} {
	if t.IsZero() {
		return nil
	}
	return t.UTC()
}

// scanner captures the Scan method of sql.Rows and sql.Row
type scanner interface {
	Scan(dst ...interface{}) error
}

func scanAll(rows *sql.Rows, scan func(scanner) error) (err error) {
	defer closeErr(rows, &err)

	for rows.Next() {
		if err := scan(rows); err != nil {
			return err
		}
	}

	return rows.Err()
}

func closeErr(c io.Closer, err *error) {
	if e := c.Close(); err != nil && *err == nil {
		*err = e
	}
}

func scanRepo(r *Repo, s scanner) error {
	return s.Scan(
		&r._ID,
		&r.Name,
		&r.Description,
		&r.Language,
		&r.CreatedAt,
		&nullTime{&r.UpdatedAt},
		&nullTime{&r.DeletedAt},
		&nullString{&r.ExternalRepo.ID},
		&nullString{&r.ExternalRepo.ServiceType},
		&nullString{&r.ExternalRepo.ServiceID},
		&r.Enabled,
		&r.Archived,
		&r.Fork,
	)
}
