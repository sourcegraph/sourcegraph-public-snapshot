package repos

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"

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
	db        DB
	listRepos *sql.Stmt
	txOpts    sql.TxOptions
}

// NewDBStore instantiates and returns a new DBStore with prepared statements.
func NewDBStore(ctx context.Context, db DB, txOpts sql.TxOptions) (*DBStore, error) {
	store := DBStore{db: db, txOpts: txOpts}
	return &store, store.prepare(ctx)
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
		db:        tx,
		listRepos: tx.StmtContext(ctx, s.listRepos),
		txOpts:    s.txOpts,
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

// ListRepos lists all configured repositories in Sourcegraph.
func (s DBStore) ListRepos(ctx context.Context) (repos []*Repo, err error) {
	var cursor, next int64 = -1, 0
	for cursor != next && err == nil {
		cursor = next
		if err = s.listReposPage(ctx, cursor, 500, &repos); len(repos) > 0 {
			next = int64(repos[len(repos)-1]._ID)
		}
	}
	return repos, err
}

func (s DBStore) listReposPage(ctx context.Context, cursor, limit int64, repos *[]*Repo) (err error) {
	rows, err := s.listRepos.QueryContext(ctx, cursor, limit)
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

const listReposSQL = `
SELECT id, name, description, language, created_at, updated_at, deleted_at,
  external_id, external_service_type, external_service_id, enabled, archived, fork
FROM repo WHERE id > $1 ORDER BY id ASC LIMIT $2
`

// UpsertRepos updates or inserts the given repos in the Sourcegraph repository store.
// The _ID field of each given Repo is set on inserts.
func (s *DBStore) UpsertRepos(ctx context.Context, repos ...*Repo) error {
	q := upsertReposQuery(repos)
	rows, err := s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}

	i := -1
	return scanAll(rows, func(sc scanner) error {
		i++
		return scanRepo(repos[i], sc)
	})
}

func (s *DBStore) prepare(ctx context.Context) error {
	for _, st := range []struct {
		stmt  **sql.Stmt
		query string
	}{
		{&s.listRepos, listReposSQL},
	} {
		stmt, err := s.db.PrepareContext(ctx, st.query)
		if err != nil {
			return errors.Wrapf(err, "failed to prepare: %s", st.query)
		}
		*st.stmt = stmt
	}

	return nil
}

func upsertReposQuery(repos []*Repo) *sqlf.Query {
	values := make([]*sqlf.Query, 0, len(repos))
	for _, r := range repos {
		values = append(values, sqlf.Sprintf(
			upsertRepoValuesFmtstr,
			upsertRepoColumns(r)...,
		))
	}
	return sqlf.Sprintf(upsertReposQueryFmtstr, sqlf.Join(values, ",\n"))
}

var upsertReposQueryFmtstr = strings.TrimSpace(fmt.Sprintf(`
INSERT INTO repo
(%s)
VALUES
%%s
ON CONFLICT ON CONSTRAINT repo_external_service_unique
DO UPDATE SET
  name        = excluded.name,
  description = excluded.description,
  language    = excluded.language,
  updated_at  = excluded.updated_at,
  deleted_at  = excluded.deleted_at,
  archived    = excluded.archived,
  fork        = excluded.fork
RETURNING id, name, description, language, created_at, updated_at,
  deleted_at, external_id, external_service_type, external_service_id,
  enabled, archived, fork
`, strings.Join(upsertRepoColumnNames, ", ")))

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

var upsertRepoValuesFmtstr = "(" + strings.TrimSuffix(strings.Repeat("%s, ", len(upsertRepoColumnNames)), ", ") + ")"

func upsertRepoColumns(r *Repo) []interface{} {
	return []interface{}{
		r.Name,
		r.Description,
		r.Language,
		"", // URI
		r.CreatedAt.UTC(),
		r.UpdatedAt.UTC(),
		r.DeletedAt.UTC(),
		r.ExternalRepo.ID,
		r.ExternalRepo.ServiceType,
		r.ExternalRepo.ServiceID,
		r.Enabled,
		r.Archived,
		r.Fork,
	}
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
		&r.ExternalRepo.ID,
		&r.ExternalRepo.ServiceType,
		&r.ExternalRepo.ServiceID,
		&r.Enabled,
		&r.Archived,
		&r.Fork,
	)
}
