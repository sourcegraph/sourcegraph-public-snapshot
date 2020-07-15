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
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/awscodecommit"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitolite"
)

// A Store exposes methods to read and write repos and external services.
type Store interface {
	ListExternalServices(context.Context, StoreListExternalServicesArgs) ([]*ExternalService, error)
	UpsertExternalServices(ctx context.Context, svcs ...*ExternalService) error

	ListRepos(context.Context, StoreListReposArgs) ([]*Repo, error)
	UpsertRepos(ctx context.Context, repos ...*Repo) error
	SetClonedRepos(ctx context.Context, repoNames ...string) error
	CountNotClonedRepos(ctx context.Context) (uint64, error)
}

// StoreListReposArgs is a query arguments type used by
// the ListRepos method of Store implementations.
type StoreListReposArgs struct {
	// Names of repos to list. When zero-valued, this is omitted from the predicate set.
	Names []string
	// IDs of repos to list. When zero-valued, this is omitted from the predicate set.
	IDs []api.RepoID
	// Kinds of repos to list. When zero-valued, this is omitted from the predicate set.
	Kinds []string
	// ExternalRepos of repos to list. When zero-valued, this is omitted from the predicate set.
	ExternalRepos []api.ExternalRepoSpec
	// Limit the total number of repos returned. Zero means no limit
	Limit int64
	// PerPage determines the number of repos returned on each page. Zero means it defaults to 10000.
	PerPage int64
	// Only include private repositories.
	PrivateOnly bool
	// Only include cloned repositories.
	ClonedOnly bool

	// UseOr decides between ANDing or ORing the predicates together.
	UseOr bool
}

// StoreListExternalServicesArgs is a query arguments type used by
// the ListExternalServices method of Store implementations.
//
// Each defined argument must map to a disjunct (i.e. AND) filter predicate.
type StoreListExternalServicesArgs struct {
	// IDs of external services to list. When zero-valued, this is omitted from the predicate set.
	IDs []int64
	// RepoIDs that the listed external services own.
	RepoIDs []api.RepoID
	// Kinds of external services to list. When zero-valued, this is omitted from the predicate set.
	Kinds []string
}

// ErrNoResults is returned by Store method invocations that yield no result set.
var ErrNoResults = errors.New("store: no results")

// A Transactor can initialize and return a TxStore which operates
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
	db     dbutil.DB
	txOpts sql.TxOptions
}

// NewDBStore instantiates and returns a new DBStore with prepared statements.
func NewDBStore(db dbutil.DB, txOpts sql.TxOptions) *DBStore {
	return &DBStore{db: db, txOpts: txOpts}
}

// Transact returns a TxStore whose methods operate within the context of a transaction.
// This method will return an error if the underlying DB cannot be interface upgraded
// to a TxBeginner.
func (s *DBStore) Transact(ctx context.Context) (TxStore, error) {
	if _, ok := s.db.(dbutil.Tx); ok { // Already in a Tx.
		return nil, errors.New("dbstore: already in a transaction")
	}

	tb, ok := s.db.(dbutil.TxBeginner)
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
// pointer is given, the transaction is committed. Otherwise, it's rolled-back.
func (s *DBStore) Done(errs ...*error) {
	switch tx, ok := s.db.(dbutil.Tx); {
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

// ListExternalServices lists all stored external services matching the given args.
func (s DBStore) ListExternalServices(ctx context.Context, args StoreListExternalServicesArgs) (svcs []*ExternalService, _ error) {
	return svcs, s.paginate(ctx, 0, 500, listExternalServicesQuery(args),
		func(sc scanner) (last, count int64, err error) {
			var svc ExternalService
			err = scanExternalService(&svc, sc)
			if err != nil {
				return 0, 0, err
			}
			svcs = append(svcs, &svc)
			return svc.ID, 1, nil
		},
	)
}

const listExternalServicesQueryFmtstr = `
-- source: cmd/repo-updater/repos/store.go:DBStore.ListExternalServices
SELECT
  id,
  kind,
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

const listRepoExternalServiceIDsSubquery = `
SELECT DISTINCT(split_part(jsonb_object_keys(sources), ':', 3)::bigint) repo_external_service_ids
FROM repo
WHERE id IN (%s)
`

func listExternalServicesQuery(args StoreListExternalServicesArgs) paginatedQuery {
	var preds []*sqlf.Query

	if len(args.IDs) > 0 {
		ids := make([]*sqlf.Query, 0, len(args.IDs))
		for _, id := range args.IDs {
			if id != 0 {
				ids = append(ids, sqlf.Sprintf("%d", id))
			}
		}
		preds = append(preds, sqlf.Sprintf("id IN (%s)", sqlf.Join(ids, ",")))
	} else if len(args.RepoIDs) > 0 {
		ids := make([]*sqlf.Query, 0, len(args.RepoIDs))
		for _, id := range args.RepoIDs {
			if id != 0 {
				ids = append(ids, sqlf.Sprintf("%d", id))
			}
		}
		preds = append(preds, sqlf.Sprintf(
			"id IN ("+listRepoExternalServiceIDsSubquery+")",
			sqlf.Join(ids, ","),
		))
	}

	if len(args.Kinds) > 0 {
		ks := make([]*sqlf.Query, 0, len(args.Kinds))
		for _, kind := range args.Kinds {
			ks = append(ks, sqlf.Sprintf("%s", strings.ToLower(kind)))
		}
		preds = append(preds,
			sqlf.Sprintf("LOWER(kind) IN (%s)", sqlf.Join(ks, ",")))
	} else {
		// HACK(tsenart): The syncer and all other places that load all external
		// services do not want phabricator instances. These are handled separately
		// by RunPhabricatorRepositorySyncWorker.
		preds = append(preds,
			sqlf.Sprintf("LOWER(kind) != 'phabricator'"))
	}

	preds = append(preds, sqlf.Sprintf("deleted_at IS NULL"))

	return func(cursor, limit int64) *sqlf.Query {
		return sqlf.Sprintf(
			listExternalServicesQueryFmtstr,
			cursor,
			sqlf.Join(preds, "\n AND "),
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
	_, _, err = scanAll(rows, func(sc scanner) (last, count int64, err error) {
		i++
		err = scanExternalService(svcs[i], sc)
		return int64(svcs[i].ID), 1, err
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
  (COALESCE(NULLIF(%s, 0), (SELECT nextval('external_services_id_seq'))), UPPER(%s), %s, %s, %s, %s, %s)
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
  kind         = UPPER(excluded.kind),
  display_name = excluded.display_name,
  config       = excluded.config,
  created_at   = excluded.created_at,
  updated_at   = excluded.updated_at,
  deleted_at   = excluded.deleted_at
RETURNING *
`

// ListRepos lists all stored repos that match the given arguments.
func (s DBStore) ListRepos(ctx context.Context, args StoreListReposArgs) (repos []*Repo, _ error) {
	return repos, s.paginate(ctx, args.Limit, args.PerPage, listReposQuery(args),
		func(sc scanner) (last, count int64, err error) {
			var r Repo
			if err = scanRepo(&r, sc); err != nil {
				return 0, 0, err
			}
			repos = append(repos, &r)
			return int64(r.ID), 1, nil
		},
	)
}

const listReposQueryFmtstr = `
-- source: cmd/repo-updater/repos/store.go:DBStore.ListRepos
SELECT
  id,
  name,
  uri,
  description,
  language,
  created_at,
  updated_at,
  deleted_at,
  external_service_type,
  external_service_id,
  external_id,
  archived,
  cloned,
  fork,
  private,
  sources,
  metadata
FROM repo
WHERE id > %s
AND %s
AND deleted_at IS NULL
ORDER BY id ASC LIMIT %s
`

func listReposQuery(args StoreListReposArgs) paginatedQuery {
	var preds []*sqlf.Query

	if len(args.Names) > 0 {
		ns := make([]*sqlf.Query, 0, len(args.Names))
		for _, name := range args.Names {
			ns = append(ns, sqlf.Sprintf("%s", name))
		}
		preds = append(preds, sqlf.Sprintf("name IN (%s)", sqlf.Join(ns, ",")))
	}

	if len(args.IDs) > 0 {
		ids := make([]*sqlf.Query, 0, len(args.IDs))
		for _, id := range args.IDs {
			if id != 0 {
				ids = append(ids, sqlf.Sprintf("%d", id))
			}
		}
		preds = append(preds, sqlf.Sprintf("id IN (%s)", sqlf.Join(ids, ",")))
	}

	if len(args.Kinds) > 0 {
		ks := make([]*sqlf.Query, 0, len(args.Kinds))
		for _, kind := range args.Kinds {
			ks = append(ks, sqlf.Sprintf("%s", strings.ToLower(kind)))
		}
		preds = append(preds,
			sqlf.Sprintf("LOWER(external_service_type) IN (%s)", sqlf.Join(ks, ",")))
	}

	if len(args.ExternalRepos) > 0 {
		er := make([]*sqlf.Query, 0, len(args.ExternalRepos))
		for _, spec := range args.ExternalRepos {
			er = append(er, sqlf.Sprintf("(external_id = %s AND external_service_type = %s AND external_service_id = %s)", spec.ID, spec.ServiceType, spec.ServiceID))
		}
		preds = append(preds, sqlf.Sprintf("(%s)", sqlf.Join(er, "\n OR ")))
	}

	if args.PrivateOnly {
		preds = append(preds, sqlf.Sprintf("private = TRUE"))
	}

	if args.ClonedOnly {
		preds = append(preds, sqlf.Sprintf("cloned = TRUE"))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	var predQ *sqlf.Query
	if args.UseOr {
		predQ = sqlf.Join(preds, "\n OR ")
	} else {
		predQ = sqlf.Join(preds, "\n AND ")
	}

	return func(cursor, limit int64) *sqlf.Query {
		return sqlf.Sprintf(
			listReposQueryFmtstr,
			cursor,
			sqlf.Sprintf("(%s)", predQ),
			limit,
		)
	}
}

// SetClonedRepos updates cloned status for all repositories.
// All repositories whose name is in repoNames will have their cloned column set to true
// and every other repository will have it set to false.
func (s DBStore) SetClonedRepos(ctx context.Context, repoNames ...string) error {
	if len(repoNames) == 0 {
		return nil
	}

	names, err := json.Marshal(repoNames)
	if err != nil {
		return nil
	}

	q := sqlf.Sprintf(setClonedReposQueryFmtstr, sqlf.Sprintf("%s", string(names)))

	_, err = s.db.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	return err
}

const setClonedReposQueryFmtstr = `
-- source: cmd/repo-updater/repos/store.go:DBStore.SetClonedRepos
--
-- This query generates a diff by selecting only
-- the repos that need to be updated.
-- Selected repos will have their cloned column reversed if
-- their cloned column is true but they are not in cloned_repos
-- or they are in cloned_repos but their cloned column is false.
--
WITH cloned_repos AS (
  SELECT jsonb_array_elements_text(%s) AS name
),
diff AS (
  SELECT id,
    cloned
  FROM repo
  WHERE
    NOT cloned
      AND name IN (SELECT name::citext FROM cloned_repos)
    OR cloned
      AND name NOT IN (SELECT name::citext FROM cloned_repos)
)
UPDATE repo
SET cloned = NOT diff.cloned
FROM diff
WHERE repo.id = diff.id;
`

// CountNotClonedRepos returns the number of repos whose cloned column is true.
func (s DBStore) CountNotClonedRepos(ctx context.Context) (uint64, error) {
	q := sqlf.Sprintf(CountNotClonedReposQueryFmtstr)

	var count uint64
	err := s.db.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&count)
	return count, err
}

const CountNotClonedReposQueryFmtstr = `
-- source: cmd/repo-updater/repos/store.go:DBStore.CountNotClonedRepos
SELECT COUNT(*) FROM repo WHERE deleted_at IS NULL AND NOT cloned
`

// a paginatedQuery returns a query with the given pagination
// parameters
type paginatedQuery func(cursor, limit int64) *sqlf.Query

func (s DBStore) paginate(ctx context.Context, limit, page int64, q paginatedQuery, scan scanFunc) (err error) {
	const defaultPerPageLimit = 10000

	if page <= 0 {
		page = defaultPerPageLimit
	}

	if limit > 0 && page > limit {
		page = limit
	}

	var (
		cursor      = int64(-1)
		remaining   = limit
		next, count int64
	)

	for cursor < next && err == nil && (limit <= 0 || remaining > 0) {
		cursor = next
		next, count, err = s.list(ctx, q(cursor, page), scan)
		if limit > 0 {
			if remaining -= count; page > remaining {
				page = remaining
			}
		}
	}

	return err
}

func (s DBStore) list(ctx context.Context, q *sqlf.Query, scan scanFunc) (last, count int64, err error) {
	rows, err := s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return 0, 0, err
	}
	return scanAll(rows, scan)
}

// UpsertRepos updates or inserts the given repos in the Sourcegraph repository store.
// The ID field is used to distinguish between Repos that need to be updated and Repos
// that need to be inserted. On inserts, the _ID field of each given Repo is set on inserts.
// The cloned column is not updated by this function.
func (s *DBStore) UpsertRepos(ctx context.Context, repos ...*Repo) (err error) {
	if len(repos) == 0 {
		return nil
	}

	var deletes, updates, inserts []*Repo
	for _, r := range repos {
		switch {
		case r.IsDeleted():
			deletes = append(deletes, r)
		case r.ID != 0:
			updates = append(updates, r)
		default:
			// We also update to un-delete soft-deleted repositories. The
			// insert statement has an on conflict do nothing.
			updates = append(updates, r)
			inserts = append(inserts, r)
		}
	}

	for _, op := range []struct {
		name  string
		query string
		repos []*Repo
	}{
		{"delete", deleteReposQuery, deletes},
		{"update", updateReposQuery, updates},
		{"insert", insertReposQuery, inserts},
		{"list", listRepoIDsQuery, inserts}, // list must run last to pick up inserted IDs
	} {
		if len(op.repos) == 0 {
			continue
		}

		q, err := batchReposQuery(op.query, op.repos)
		if err != nil {
			return errors.Wrap(err, op.name)
		}

		rows, err := s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
		if err != nil {
			return errors.Wrap(err, op.name)
		}

		if op.name != "list" {
			if err = rows.Close(); err != nil {
				return errors.Wrap(err, op.name)
			}
			// Nothing to scan
			continue
		}

		_, _, err = scanAll(rows, func(sc scanner) (last, count int64, err error) {
			var (
				i  int
				id api.RepoID
			)

			err = sc.Scan(&i, &id)
			if err != nil {
				return 0, 0, err
			}
			op.repos[i-1].ID = id
			return int64(id), 1, nil
		})

		if err != nil {
			return errors.Wrap(err, op.name)
		}
	}

	// Assert we have set ID for all repos.
	for _, r := range repos {
		if r.ID == 0 && !r.IsDeleted() {
			return errors.Errorf("DBStore.UpsertRepos did not set ID for %v", r)
		}
	}

	return nil
}

func batchReposQuery(fmtstr string, repos []*Repo) (_ *sqlf.Query, err error) {
	type record struct {
		ID                  api.RepoID      `json:"id"`
		Name                string          `json:"name"`
		URI                 *string         `json:"uri,omitempty"`
		Description         string          `json:"description"`
		Language            string          `json:"language"`
		CreatedAt           time.Time       `json:"created_at"`
		UpdatedAt           *time.Time      `json:"updated_at,omitempty"`
		DeletedAt           *time.Time      `json:"deleted_at,omitempty"`
		ExternalServiceType *string         `json:"external_service_type,omitempty"`
		ExternalServiceID   *string         `json:"external_service_id,omitempty"`
		ExternalID          *string         `json:"external_id,omitempty"`
		Archived            bool            `json:"archived"`
		Fork                bool            `json:"fork"`
		Private             bool            `json:"private"`
		Sources             json.RawMessage `json:"sources"`
		Metadata            json.RawMessage `json:"metadata"`
	}

	records := make([]record, 0, len(repos))
	for _, r := range repos {
		sources, err := json.Marshal(r.Sources)
		if err != nil {
			return nil, errors.Wrapf(err, "batchReposQuery: sources marshalling failed")
		}

		metadata, err := metadataColumn(r.Metadata)
		if err != nil {
			return nil, errors.Wrapf(err, "batchReposQuery: metadata marshalling failed")
		}

		records = append(records, record{
			ID:                  r.ID,
			Name:                r.Name,
			URI:                 nullStringColumn(r.URI),
			Description:         r.Description,
			Language:            r.Language,
			CreatedAt:           r.CreatedAt.UTC(),
			UpdatedAt:           nullTimeColumn(r.UpdatedAt.UTC()),
			DeletedAt:           nullTimeColumn(r.DeletedAt.UTC()),
			ExternalServiceType: nullStringColumn(r.ExternalRepo.ServiceType),
			ExternalServiceID:   nullStringColumn(r.ExternalRepo.ServiceID),
			ExternalID:          nullStringColumn(r.ExternalRepo.ID),
			Archived:            r.Archived,
			Fork:                r.Fork,
			Private:             r.Private,
			Sources:             sources,
			Metadata:            metadata,
		})
	}

	batch, err := json.MarshalIndent(records, "    ", "    ")
	if err != nil {
		return nil, err
	}

	return sqlf.Sprintf(fmtstr, string(batch)), nil
}

const batchReposQueryFmtstr = `
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
      uri                   citext,
      description           text,
      language              text,
      created_at            timestamptz,
      updated_at            timestamptz,
      deleted_at            timestamptz,
      external_service_type text,
      external_service_id   text,
      external_id           text,
      archived              boolean,
      fork                  boolean,
      private               boolean,
      sources               jsonb,
      metadata              jsonb
    )
  )
  WITH ORDINALITY
)`

var updateReposQuery = batchReposQueryFmtstr + `
UPDATE repo
SET
  name                  = batch.name,
  uri                   = batch.uri,
  description           = batch.description,
  language              = batch.language,
  created_at            = batch.created_at,
  updated_at            = batch.updated_at,
  deleted_at            = batch.deleted_at,
  external_service_type = batch.external_service_type,
  external_service_id   = batch.external_service_id,
  external_id           = batch.external_id,
  archived              = batch.archived,
  fork                  = batch.fork,
  private               = batch.private,
  sources               = batch.sources,
  metadata              = batch.metadata
FROM batch
WHERE repo.external_service_type = batch.external_service_type
AND repo.external_service_id = batch.external_service_id
AND repo.external_id = batch.external_id
`

// delete is a soft-delete. name is unique and the syncer ensures we respect
// that constraint. However, the syncer is unaware of soft-deleted
// repositories. So we update the name to something unique to prevent
// violating this constraint between active and soft-deleted names.
var deleteReposQuery = batchReposQueryFmtstr + `
UPDATE repo
SET
  name = 'DELETED-' || extract(epoch from transaction_timestamp()) || '-' || batch.name,
  deleted_at = batch.deleted_at
FROM batch
WHERE batch.deleted_at IS NOT NULL
AND repo.id = batch.id
`

var insertReposQuery = batchReposQueryFmtstr + `
INSERT INTO repo (
  name,
  uri,
  description,
  language,
  created_at,
  updated_at,
  deleted_at,
  external_service_type,
  external_service_id,
  external_id,
  archived,
  fork,
  private,
  sources,
  metadata
)
SELECT
  name,
  NULLIF(BTRIM(uri), ''),
  description,
  language,
  created_at,
  updated_at,
  deleted_at,
  external_service_type,
  external_service_id,
  external_id,
  archived,
  fork,
  private,
  sources,
  metadata
FROM batch
ON CONFLICT (external_service_type, external_service_id, external_id) DO NOTHING
`

var listRepoIDsQuery = batchReposQueryFmtstr + `
SELECT batch.ordinality, repo.id
FROM batch
JOIN repo USING (external_service_type, external_service_id, external_id)
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
// the last id column scanned and the count of scanned rows.
type scanFunc func(scanner) (last, count int64, err error)

func scanAll(rows *sql.Rows, scan scanFunc) (last, count int64, err error) {
	defer closeErr(rows, &err)

	last = -1
	for rows.Next() {
		var n int64
		if last, n, err = scan(rows); err != nil {
			return last, count, err
		}
		count += n
	}

	return last, count, rows.Err()
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
		&dbutil.NullTime{Time: &svc.UpdatedAt},
		&dbutil.NullTime{Time: &svc.DeletedAt},
	)
}

func scanRepo(r *Repo, s scanner) error {
	var sources, metadata json.RawMessage
	err := s.Scan(
		&r.ID,
		&r.Name,
		&dbutil.NullString{S: &r.URI},
		&r.Description,
		&r.Language,
		&r.CreatedAt,
		&dbutil.NullTime{Time: &r.UpdatedAt},
		&dbutil.NullTime{Time: &r.DeletedAt},
		&dbutil.NullString{S: &r.ExternalRepo.ServiceType},
		&dbutil.NullString{S: &r.ExternalRepo.ServiceID},
		&dbutil.NullString{S: &r.ExternalRepo.ID},
		&r.Archived,
		&r.Cloned,
		&r.Fork,
		&r.Private,
		&sources,
		&metadata,
	)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(sources, &r.Sources); err != nil {
		return errors.Wrap(err, "scanRepo: failed to unmarshal sources")
	}

	typ, ok := extsvc.ParseServiceType(r.ExternalRepo.ServiceType)
	if !ok {
		return nil
	}
	switch typ {
	case extsvc.TypeGitHub:
		r.Metadata = new(github.Repository)
	case extsvc.TypeGitLab:
		r.Metadata = new(gitlab.Project)
	case extsvc.TypeBitbucketServer:
		r.Metadata = new(bitbucketserver.Repo)
	case extsvc.TypeBitbucketCloud:
		r.Metadata = new(bitbucketcloud.Repo)
	case extsvc.TypeAWSCodeCommit:
		r.Metadata = new(awscodecommit.Repository)
	case extsvc.TypeGitolite:
		r.Metadata = new(gitolite.Repo)
	default:
		return nil
	}

	if err = json.Unmarshal(metadata, r.Metadata); err != nil {
		return errors.Wrapf(err, "scanRepo: failed to unmarshal %q metadata", typ)
	}

	return nil
}
