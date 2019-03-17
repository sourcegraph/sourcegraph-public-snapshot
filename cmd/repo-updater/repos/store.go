package repos

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/github"
)

// A Store exposes methods to read and write repos and external services.
type Store interface {
	ListExternalServices(ctx context.Context, kinds ...string) ([]*ExternalService, error)
	UpsertExternalServices(ctx context.Context, svcs ...*ExternalService) error

	GetRepoByName(ctx context.Context, name string) (*Repo, error)
	ListRepos(ctx context.Context, kinds ...string) ([]*Repo, error)
	UpsertRepos(ctx context.Context, repos ...*Repo) error
}

// ErrNoResults is returned by Store method invocations that yield no result set.
var ErrNoResults = errors.New("store: no results")

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
	txOpts sql.TxOptions
}

// NewDBStore instantiates and returns a new DBStore with prepared statements.
func NewDBStore(ctx context.Context, db DB, txOpts sql.TxOptions) *DBStore {
	return &DBStore{db: db, txOpts: txOpts}
}

// Transact returns a TxStore whose methods operate within the context of a transaction.
// This method will return an error if the underlying DB cannot be interface upgraded
// to a TxBeginner.
func (s *DBStore) Transact(ctx context.Context) (TxStore, error) {
	if _, ok := s.db.(Tx); ok { // Already in a Tx.
		return nil, errors.New("dbstore: already in a transaction")
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

// ListExternalServices lists all stored external services that are not deleted and have one of the
// specified kinds.
func (s DBStore) ListExternalServices(ctx context.Context, kinds ...string) (svcs []*ExternalService, _ error) {
	return svcs, s.paginate(ctx, listExternalServicesQuery(kinds), func(sc scanner) (int64, error) {
		var svc ExternalService
		err := scanExternalService(&svc, sc)
		if err != nil {
			return 0, err
		}
		svcs = append(svcs, &svc)
		return svc.ID, nil
	})
}

const listExternalServicesQueryFmtstr = `
-- source: cmd/repo-updater/repos/store.go:DBStore.ListExternalServices
SELECT
  id,
  LOWER(kind),
  display_name,
  config,
  created_at,
  updated_at,
  deleted_at
FROM external_services
WHERE id > %s
AND %s
ORDER BY id ASC LIMIT %s
`

func listExternalServicesQuery(kinds []string) paginatedQuery {
	kq := sqlf.Sprintf("TRUE")
	if len(kinds) > 0 {
		ks := make([]*sqlf.Query, 0, len(kinds))
		for _, kind := range kinds {
			ks = append(ks, sqlf.Sprintf("%s", strings.ToLower(kind)))
		}
		kq = sqlf.Sprintf("LOWER(kind) IN (%s)", sqlf.Join(ks, ","))
	}

	return func(cursor, limit int64) *sqlf.Query {
		return sqlf.Sprintf(
			listExternalServicesQueryFmtstr,
			cursor,
			kq,
			limit,
		)
	}
}

// UpsertExternalServices updates or inserts the given ExternalServices.
func (s DBStore) UpsertExternalServices(ctx context.Context, svcs ...*ExternalService) error {
	if len(svcs) == 0 {
		return nil
	}

	q := upsertExternalServicesQuery(svcs)
	rows, err := s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}

	i := -1
	_, err = scanAll(rows, func(sc scanner) (int64, error) {
		i++
		err := scanExternalService(svcs[i], sc)
		return int64(svcs[i].ID), err
	})

	return err
}

func upsertExternalServicesQuery(svcs []*ExternalService) *sqlf.Query {
	vals := make([]*sqlf.Query, 0, len(svcs))
	for _, s := range svcs {
		vals = append(vals, sqlf.Sprintf(
			upsertExternalServicesQueryValueFmtstr,
			s.ID,
			s.Kind,
			s.DisplayName,
			s.Config,
			s.CreatedAt.UTC(),
			s.UpdatedAt.UTC(),
			nullTimeColumn(s.DeletedAt.UTC()),
		))
	}

	return sqlf.Sprintf(
		upsertExternalServicesQueryFmtstr,
		sqlf.Join(vals, ",\n"),
	)
}

const upsertExternalServicesQueryValueFmtstr = `
  (COALESCE(NULLIF(%s, 0), (SELECT nextval('external_services_id_seq'))), %s, %s, %s, %s, %s, %s)
`

const upsertExternalServicesQueryFmtstr = `
-- source: cmd/repo-updater/repos/store.go:DBStore.UpsertExternalServices
INSERT INTO external_services (
  id,
  kind,
  display_name,
  config,
  created_at,
  updated_at,
  deleted_at
)
VALUES %s
ON CONFLICT(id) DO UPDATE
SET
  kind         = excluded.kind,
  display_name = excluded.display_name,
  config       = excluded.config,
  created_at   = excluded.created_at,
  updated_at   = excluded.updated_at,
  deleted_at   = excluded.deleted_at
RETURNING *
`

// GetRepoByName looks up the repo with the given name.
func (s DBStore) GetRepoByName(ctx context.Context, name string) (*Repo, error) {
	var r Repo
	id, err := s.list(ctx, getRepoByNameQuery(name), func(s scanner) (int64, error) {
		err := scanRepo(&r, s)
		return int64(r.ID), err
	})

	switch {
	case err != nil:
		return nil, err
	case id == -1:
		return nil, ErrNoResults
	default:
		return &r, nil
	}
}

const getRepoByNameQueryFmtstr = `
-- source: cmd/repo-updater/repos/store.go:DBStore.GetRepoByName
SELECT
  id,
  name,
  description,
  language,
  created_at,
  updated_at,
  deleted_at,
  external_service_type,
  external_service_id,
  external_id,
  enabled,
  archived,
  fork,
  sources,
  metadata
FROM repo
WHERE name = %s
AND deleted_at IS NULL
`

func getRepoByNameQuery(name string) *sqlf.Query {
	return sqlf.Sprintf(getRepoByNameQueryFmtstr, name)
}

// ListRepos lists all stored repos that are not deleted, have one of the
// specified external service kind.
func (s DBStore) ListRepos(ctx context.Context, kinds ...string) (repos []*Repo, _ error) {
	return repos, s.paginate(ctx, listReposQuery(kinds), func(sc scanner) (int64, error) {
		var r Repo
		if err := scanRepo(&r, sc); err != nil {
			return 0, err
		}
		repos = append(repos, &r)
		return int64(r.ID), nil
	})
}

const listReposQueryFmtstr = `
-- source: cmd/repo-updater/repos/store.go:DBStore.ListRepos
SELECT
  id,
  name,
  description,
  language,
  created_at,
  updated_at,
  deleted_at,
  external_service_type,
  external_service_id,
  external_id,
  enabled,
  archived,
  fork,
  sources,
  metadata
FROM repo
WHERE id > %s
AND %s
ORDER BY id ASC LIMIT %s
`

func listReposQuery(kinds []string) paginatedQuery {
	kq := sqlf.Sprintf("TRUE")
	if len(kinds) > 0 {
		ks := make([]*sqlf.Query, 0, len(kinds))
		for _, kind := range kinds {
			ks = append(ks, sqlf.Sprintf("%s", strings.ToLower(kind)))
		}
		kq = sqlf.Sprintf("LOWER(external_service_type) IN (%s)", sqlf.Join(ks, ","))
	}

	return func(cursor, limit int64) *sqlf.Query {
		return sqlf.Sprintf(
			listReposQueryFmtstr,
			cursor,
			kq,
			limit,
		)
	}
}

// a paginatedQuery returns a query with the given pagination
// parameters
type paginatedQuery func(cursor, limit int64) *sqlf.Query

func (s DBStore) paginate(ctx context.Context, q paginatedQuery, scan scanFunc) (err error) {
	var cursor, next int64 = -1, 0
	for cursor < next && err == nil {
		cursor = next
		next, err = s.list(ctx, q(cursor, 500), scan)
	}
	return err
}

func (s DBStore) list(ctx context.Context, q *sqlf.Query, scan scanFunc) (last int64, err error) {
	rows, err := s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return 0, err
	}
	return scanAll(rows, scan)
}

// UpsertRepos updates or inserts the given repos in the Sourcegraph repository store.
// The ID field is used to distinguish between Repos that need to be updated and Repos
// that need to be inserted. On inserts, the _ID field of each given Repo is set on inserts.
func (s *DBStore) UpsertRepos(ctx context.Context, repos ...*Repo) (err error) {
	if len(repos) == 0 {
		return nil
	}

	q, err := upsertReposQuery(repos)
	if err != nil {
		return err
	}

	rows, err := s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}

	i := -1

	_, err = scanAll(rows, func(sc scanner) (int64, error) {
		i++
		err := scanRepo(repos[i], sc)
		return int64(repos[i].ID), err
	})

	return err
}

func upsertReposQuery(repos []*Repo) (_ *sqlf.Query, err error) {
	type record struct {
		ID                  uint32          `json:"id"`
		Name                string          `json:"name"`
		Description         string          `json:"description"`
		Language            string          `json:"language"`
		CreatedAt           time.Time       `json:"created_at"`
		UpdatedAt           *time.Time      `json:"updated_at,omitempty"`
		DeletedAt           *time.Time      `json:"deleted_at,omitempty"`
		ExternalServiceType *string         `json:"external_service_type,omitempty"`
		ExternalServiceID   *string         `json:"external_service_id,omitempty"`
		ExternalID          *string         `json:"external_id,omitempty"`
		Enabled             bool            `json:"enabled"`
		Archived            bool            `json:"archived"`
		Fork                bool            `json:"fork"`
		Sources             json.RawMessage `json:"sources"`
		Metadata            json.RawMessage `json:"metadata"`
	}

	records := make([]record, 0, len(repos))
	for _, r := range repos {
		sources, err := json.Marshal(r.Sources)
		if err != nil {
			return nil, errors.Wrapf(err, "upsertReposQuery: sources marshalling failed")
		}

		metadata, err := metadataColumn(r.Metadata)
		if err != nil {
			return nil, errors.Wrapf(err, "upsertReposQuery: metadata marshalling failed")
		}

		records = append(records, record{
			ID:                  r.ID,
			Name:                r.Name,
			Description:         r.Description,
			Language:            r.Language,
			CreatedAt:           r.CreatedAt.UTC(),
			UpdatedAt:           nullTimeColumn(r.UpdatedAt.UTC()),
			DeletedAt:           nullTimeColumn(r.DeletedAt.UTC()),
			ExternalServiceType: nullStringColumn(r.ExternalRepo.ServiceType),
			ExternalServiceID:   nullStringColumn(r.ExternalRepo.ServiceID),
			ExternalID:          nullStringColumn(r.ExternalRepo.ID),
			Enabled:             r.Enabled,
			Archived:            r.Archived,
			Fork:                r.Fork,
			Sources:             sources,
			Metadata:            metadata,
		})
	}

	batch, err := json.MarshalIndent(records, "    ", "    ")
	if err != nil {
		return nil, err
	}

	return sqlf.Sprintf(upsertReposQueryFmtstr, string(batch)), nil
}

var upsertReposQueryFmtstr = `
-- source: cmd/repo-updater/repos/store.go:DBStore.UpsertRepos

--
-- The "batch" Common Table Expression (CTE) produces the records to be upserted,
-- leveraging the "json_to_recordset" Postgres function to parse the JSON
-- serialised repos array being passed from the application code.
--
-- This is done for two reasons:
--
--     1. To circumvent Postgres' limit of 32767 bind parameters per statement.
--     2. To use the WITH ORDINALITY table function feature which gives us
--        an auto-generated ordering column we rely on to maintain the order of
--        the final result set produced (see the last SELECT statement).
--
WITH batch AS (
  SELECT * FROM ROWS FROM (
  json_to_recordset(%s)
  AS (
      id                    integer,
      name                  citext,
      description           text,
      language              text,
      created_at            timestamptz,
      updated_at            timestamptz,
      deleted_at            timestamptz,
      external_service_type text,
      external_service_id   text,
      external_id           text,
      enabled               boolean,
      archived              boolean,
      fork                  boolean,
      sources               jsonb,
      metadata              jsonb
    )
  )
  WITH ORDINALITY
),

--
-- The "updated" Common Table Expression (CTE) updates all records from our "batch"
-- that already exist in the repos table as informed by their unique contraints.
-- It returns the set of updated records.
--
updated AS (
  UPDATE repo
  SET
    name                  = batch.name,
    description           = batch.description,
    language              = batch.language,
    created_at            = COALESCE(batch.created_at, repo.created_at),
    updated_at            = COALESCE(batch.updated_at, repo.updated_at),
    deleted_at            = batch.deleted_at,
    external_service_type = COALESCE(NULLIF(BTRIM(batch.external_service_type), ''), repo.external_service_type),
    external_service_id   = COALESCE(NULLIF(BTRIM(batch.external_service_id), ''), repo.external_service_id),
    external_id           = COALESCE(NULLIF(BTRIM(batch.external_id), ''), repo.external_service_id),
    enabled               = batch.enabled,
    archived              = batch.archived,
    fork                  = batch.fork,
    sources               = batch.sources,
    metadata              = batch.metadata
  FROM batch
  WHERE repo.name = batch.name OR (
    repo.external_id IS NOT NULL
    AND repo.external_service_id IS NOT NULL
    AND repo.external_service_type IS NOT NULL
    AND batch.external_id IS NOT NULL
    AND batch.external_service_id IS NOT NULL
    AND batch.external_service_type IS NOT NULL
    AND repo.external_service_id = batch.external_service_id
    AND repo.external_id = batch.external_id
    AND repo.external_service_type = batch.external_service_type
  )
  RETURNING repo.*
),

--
-- The "inserted" Common Table Expression (CTE) inserts all records from our "batch"
-- that don't already exist in the repos table as informed by their unique constraints.
-- It returns the set of new records with their "id" column set as well as other columns
-- with default values that had no value set in the batch.
--
inserted AS (
  INSERT INTO repo (
    name,
    description,
    language,
    created_at,
    updated_at,
    deleted_at,
    external_service_type,
    external_service_id,
    external_id,
    enabled,
    archived,
    fork,
    sources,
    metadata
  )
  SELECT
    name,
    description,
    language,
    created_at,
    updated_at,
    deleted_at,
    NULLIF(BTRIM(external_service_type), ''),
    NULLIF(BTRIM(external_service_id), ''),
    NULLIF(BTRIM(external_id), ''),
    enabled,
    archived,
    fork,
    sources,
    metadata
  FROM batch
  WHERE NOT EXISTS (SELECT 1 FROM repo WHERE repo.name = batch.name)
  AND NOT EXISTS (
    SELECT 1
    FROM repo
    WHERE repo.external_id IS NOT NULL
      AND repo.external_service_type IS NOT NULL
      AND repo.external_service_id IS NOT NULL
      AND batch.external_id IS NOT NULL
      AND batch.external_service_type IS NOT NULL
      AND batch.external_service_id IS NOT NULL
      AND repo.external_service_id = batch.external_service_id
      AND repo.external_id = batch.external_id
      AND repo.external_service_type = batch.external_service_type
    )
  RETURNING repo.*
)

--
-- This select statement produces rows of the "batch" CTE hydrated with the
-- column values returned by the "updated" and "inserted" CTEs, in the same
-- order as the original batch.
--
SELECT
  GREATEST(updated.id, inserted.id, batch.id) AS id,
  COALESCE(updated.name, inserted.name, batch.name) AS name,
  COALESCE(updated.description, inserted.description, batch.description) AS description,
  COALESCE(updated.language, inserted.language, batch.language) AS language,
  COALESCE(updated.created_at, inserted.created_at, batch.created_at) AS created_at,
  COALESCE(updated.updated_at, inserted.updated_at, batch.updated_at) AS updated_at,
  COALESCE(updated.deleted_at, inserted.deleted_at, batch.deleted_at) AS deleted_at,
  COALESCE(updated.external_service_type, inserted.external_service_type, batch.external_service_type) AS external_service_type,
  COALESCE(updated.external_service_id, inserted.external_service_id, batch.external_service_id) AS external_service_id,
  COALESCE(updated.external_id, inserted.external_id, batch.external_id) AS external_id,
  COALESCE(updated.enabled, inserted.enabled, batch.enabled) AS enabled,
  COALESCE(updated.archived, inserted.archived, batch.archived) AS archived,
  COALESCE(updated.fork, inserted.fork, batch.fork) AS fork,
  COALESCE(updated.sources, inserted.sources, batch.sources) AS sources,
  COALESCE(updated.metadata, inserted.metadata, batch.metadata) AS metadata
FROM batch
LEFT JOIN updated  ON batch.name = updated.name
LEFT JOIN inserted ON batch.name = inserted.name
ORDER BY batch.ordinality
`

func nullTimeColumn(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func nullStringColumn(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func metadataColumn(metadata interface{}) (msg json.RawMessage, err error) {
	switch m := metadata.(type) {
	case nil:
		msg = json.RawMessage("{}")
	case string:
		msg = json.RawMessage(m)
	case []byte:
		msg = m
	case json.RawMessage:
		msg = m
	default:
		msg, err = json.MarshalIndent(m, "        ", "    ")
	}
	return
}

// scanner captures the Scan method of sql.Rows and sql.Row
type scanner interface {
	Scan(dst ...interface{}) error
}

// a scanFunc scans one or more rows from a scanner, returning
// the last id column scanned.
type scanFunc func(scanner) (last int64, err error)

func scanAll(rows *sql.Rows, scan scanFunc) (last int64, err error) {
	defer closeErr(rows, &err)

	last = -1
	for rows.Next() {
		if last, err = scan(rows); err != nil {
			return last, err
		}
	}

	return last, rows.Err()
}

func closeErr(c io.Closer, err *error) {
	if e := c.Close(); err != nil && *err == nil {
		*err = e
	}
}

func scanExternalService(svc *ExternalService, s scanner) error {
	return s.Scan(
		&svc.ID,
		&svc.Kind,
		&svc.DisplayName,
		&svc.Config,
		&svc.CreatedAt,
		&nullTime{&svc.UpdatedAt},
		&nullTime{&svc.DeletedAt},
	)
}

func scanRepo(r *Repo, s scanner) error {
	var sources, metadata json.RawMessage
	err := s.Scan(
		&r.ID,
		&r.Name,
		&r.Description,
		&r.Language,
		&r.CreatedAt,
		&nullTime{&r.UpdatedAt},
		&nullTime{&r.DeletedAt},
		&nullString{&r.ExternalRepo.ServiceType},
		&nullString{&r.ExternalRepo.ServiceID},
		&nullString{&r.ExternalRepo.ID},
		&r.Enabled,
		&r.Archived,
		&r.Fork,
		&sources,
		&metadata,
	)

	if err != nil {
		return err
	}

	if err = json.Unmarshal(sources, &r.Sources); err != nil {
		return errors.Wrap(err, "scanRepo: failed to unmarshal sources")
	}

	typ := strings.ToLower(r.ExternalRepo.ServiceType)
	switch typ {
	case "github":
		r.Metadata = new(github.Repository)
	default:
		return nil
	}

	if err = json.Unmarshal(metadata, &r.Metadata); err != nil {
		return errors.Wrapf(err, "scanRepo: failed to unmarshal %q metadata", typ)
	}

	return nil
}
