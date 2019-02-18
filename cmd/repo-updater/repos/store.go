package repos

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

// A Store exposes methods to read and write persistent repositories.
type Store interface {
	ListRepos(ctx context.Context, names ...string) ([]*Repo, error)
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

// ListRepos lists all stored repos that are not deleted, have one of the configured
// external service kind and have any sources defined OR their name is one of the given
// names.
func (s DBStore) ListRepos(ctx context.Context, names ...string) (repos []*Repo, err error) {
	return repos, s.paginate(ctx, &repos, listReposQuery(s.kinds, names))
}

const listReposQueryFmtstr = `
-- cmd/repo-updater/repos/store.go:DBStore.ListRepos
SELECT id, name, description, language, created_at, updated_at, deleted_at,
  external_id, external_service_type, external_service_id, enabled, archived, fork
  ARRAY(SELECT jsonb_object_keys(sources)) as sources, metadata
FROM repo
WHERE id > %s
AND deleted_at IS NULL
AND %s
AND external_service_type IN (%s)
AND (sources != '{}' OR %s)
ORDER BY id ASC LIMIT %s`

func listReposQuery(kinds, names []string) paginatedQuery {
	kq := sqlf.Sprintf("TRUE")
	if len(kinds) > 0 {
		ks := make([]*sqlf.Query, 0, len(kinds))
		for _, kind := range kinds {
			ks = append(ks, sqlf.Sprintf("%s", strings.ToUpper(kind)))
		}
		kq = sqlf.Sprintf("external_service_type IN (%s)", sqlf.Join(ks, ","))
	}

	nq := sqlf.Sprintf("FALSE")
	if len(names) > 0 {
		ns := make([]*sqlf.Query, 0, len(names))
		for _, name := range names {
			ns = append(ns, sqlf.Sprintf("%s", name))
		}
		nq = sqlf.Sprintf("name IN (%s)", sqlf.Join(ns, ",\n"))
	}

	return func(cursor, limit int64) *sqlf.Query {
		return sqlf.Sprintf(
			listReposQueryFmtstr,
			cursor,
			kq,
			nq,
			limit,
		)
	}
}

// a paginatedQuery returns a query with the given pagination
// parameters
type paginatedQuery func(cursor, limit int64) *sqlf.Query

func (s DBStore) paginate(ctx context.Context, repos *[]*Repo, q paginatedQuery) (err error) {
	var cursor, next int64 = -1, 0
	for cursor != next && err == nil {
		cursor = next
		if err = s.page(ctx, q(cursor, 500), repos); len(*repos) > 0 {
			next = int64((*repos)[len(*repos)-1].ID)
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

// UpsertRepos updates or inserts the given repos in the Sourcegraph repository store.
// The _ID field of each given Repo is set on inserts.
func (s *DBStore) UpsertRepos(ctx context.Context, repos ...*Repo) (err error) {
	if len(repos) == 0 {
		return nil
	}

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

var upsertReposQueryFmtstr = fmt.Sprintf(`
-- cmd/repo-updater/repos/store.go:DBStore.UpsertRepos
WITH batch AS (
  SELECT %s
  FROM (VALUES %%s) AS batch(%s)
),

updated AS (
  UPDATE repo
  SET (%s) = (%s)
  FROM batch
  WHERE %s
  AND (%s) <> (%s)
  RETURNING repo.id
),

inserted AS (
  INSERT INTO repo
  SELECT %s FROM batch
  WHERE NOT EXISTS (SELECT 1 FROM repo WHERE %s)
  RETURNING repo.id
)

SELECT %s FROM batch INNER JOIN repo ON repo.name = batch.name`,

	upsertColumnNamesQuery,
	upsertColumnNamesQuery,
	upsertColumnNamesQuery,
	upsertBatchColumnNamesQuery,
	repoIDConditional,
	upsertBatchColumnNamesQuery,
	upsertRepoColumnNamesQuery,
	upsertColumnNamesQuery,
	repoIDConditional,
	upsertColumnNamesQuery,
)

const repoIDConditional = `
repo.id == batch.id OR repo.name = batch.name OR (
  repo.external_id IS NOT NULL
  AND repo.external_service_id IS NOT NULL
  AND batch.external_id IS NOT NULL
  AND batch.external_service_id IS NOT NULL
  AND repo.external_service_id = batch.external_service_id
  AND repo.external_id = batch.external_id
)
`

func upsertReposQuery(repos []*Repo) *sqlf.Query {
	values := make([]*sqlf.Query, 0, len(repos))
	for _, repo := range repos {
		values = append(values, upsertRepoColumnValues(repo))
	}
	return sqlf.Sprintf(
		upsertReposQueryFmtstr,
		sqlf.Join(values, ",\n"),
	)
}

var upsertRepoColumnNames = []string{
	"name",
	"description",
	"language",
	"uri",
	"created_at",
	"updated_at",
	"deleted_at",
	"external_service_type",
	"external_service_id",
	"external_id",
	"enabled",
	"archived",
	"fork",
	"sources",
	"metadata",
}

var (
	upsertColumnNamesQuery      = columnNames("", upsertRepoColumnNames)
	upsertBatchColumnNamesQuery = columnNames("batch", upsertRepoColumnNames)
	upsertRepoColumnNamesQuery  = columnNames("repo", upsertRepoColumnNames)
)

func upsertRepoColumnValues(r *Repo) *sqlf.Query {
	return sqlf.Sprintf(
		"("+strings.Repeat("%s ", len(upsertRepoColumnNames))+")",
		r.Name,
		r.Description,
		r.Language,
		"", // URI
		timeColumn(r.CreatedAt),
		timeColumn(r.UpdatedAt),
		timeColumn(r.DeletedAt),
		nullStringColumn(r.ExternalRepo.ServiceType),
		nullStringColumn(r.ExternalRepo.ServiceID),
		nullStringColumn(r.ExternalRepo.ID),
		r.Enabled,
		r.Archived,
		r.Fork,
		sourcesColumn(r.Sources),
		metadataColumn(r.Metadata),
	)
}

func columnNames(table string, names []string) string {
	columns := make([]string, 0, len(names))
	for _, name := range names {
		column := pq.QuoteIdentifier(name)
		if table != "" {
			column = pq.QuoteIdentifier(table) + "." + column
		}
		columns = append(columns, column)
	}
	return strings.Join(columns, ", ")
}

func timeColumn(t time.Time) interface{} {
	if t.IsZero() {
		return nil
	}
	return t.UTC()
}

func nullStringColumn(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func sourcesColumn(sources []string) *sqlf.Query {
	args := make([]*sqlf.Query, 0, len(sources)*2)

	for _, s := range sources {
		args = append(args,
			sqlf.Sprintf("%s", s),
			sqlf.Sprintf("NULL"),
		)
	}

	return sqlf.Sprintf(
		"jsonb_build_object(%s)",
		sqlf.Join(args, ","),
	)
}

func metadataColumn(metadata interface{}) *sqlf.Query {
	if metadata == nil {
		return sqlf.Sprintf("'{}'::jsonb")
	}

	// TODO What to do in the rare case this doesn't serialize?
	b, err := json.Marshal(metadata)
	if err != nil {
		return sqlf.Sprintf("'{}'::jsonb")
	}

	return sqlf.Sprintf("jsonb_object(%s)", string(b))
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
		&r.ID,
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
