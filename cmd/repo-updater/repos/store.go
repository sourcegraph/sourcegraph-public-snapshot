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
	UpsertSources(ctx context.Context, inserts, updates, deletes map[api.RepoID][]SourceInfo) error
	SetClonedRepos(ctx context.Context, repoNames ...string) error
	CountNotClonedRepos(ctx context.Context) (uint64, error)

	// EnqueueSyncJobs enqueues sync jobs per external service where their next_sync_at is due.
	// If ignoreSiteAdmin is true then we only sync user added external services.
	EnqueueSyncJobs(ctx context.Context, ignoreSiteAdmin bool) error

	// TODO: These two methods should not be used in production, move them to
	// an extension interface that's explicitly for testing.
	InsertRepos(context.Context, ...*Repo) error
	DeleteRepos(ctx context.Context, ids ...api.RepoID) error
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
	// ExternalServiceID, if non zero, will only return repos added by the given external service.
	// The id is that of the external_services table NOT the external_service_id in the repo table
	ExternalServiceID int64
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

const DefaultListExternalServicesPerPage = 500

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

	// NamespaceUserID limits to only fetch external service owned by the given user. When zero-valued
	// this is omitted from the predicate set.
	// The special value -1 should be passed to return only service owned by NO user. ie, owned by site admin.
	NamespaceUserID int32

	// Limit is the total number of items to list. The zero value means no limit.
	Limit int64
	// Cursor will limit the query to external services that have an id greater than Cursor.
	Cursor int64

	// PerPage defines how many external services to fetch per page when listing all services. If zero-valued
	// DefaultListExternalServicePageSize will be used.
	PerPage int64
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

// repoRecord is the json representation of a repository as used in this package
// Postgres CTEs.
type repoRecord struct {
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
	Metadata            json.RawMessage `json:"metadata"`
	Sources             json.RawMessage `json:"sources,omitempty"`
}

func newRepoRecord(r *Repo) (*repoRecord, error) {
	metadata, err := metadataColumn(r.Metadata)
	if err != nil {
		return nil, errors.Wrapf(err, "newRecord: metadata marshalling failed")
	}

	sources, err := sourcesColumn(r.ID, r.Sources)
	if err != nil {
		return nil, errors.Wrapf(err, "newRecord: sources marshalling failed")
	}

	return &repoRecord{
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
		Metadata:            metadata,
		Sources:             sources,
	}, nil
}

type sourceRecord struct {
	ExternalServiceID int64  `json:"external_service_id"`
	RepoID            int64  `json:"repo_id"`
	CloneURL          string `json:"clone_url"`
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

// WithStore is a store that can take a db handle and return
// a new Store implementation that uses it.
type WithStore interface {
	With(dbutil.DB) Store
}

// With returns a new store using the given db handle.
// It implements the WithStore interface.
func (s *DBStore) With(db dbutil.DB) Store {
	return &DBStore{
		db: db,
	}
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
	if args.PerPage <= 0 {
		args.PerPage = DefaultListExternalServicesPerPage
	}
	return svcs, s.paginate(ctx, args.Limit, args.PerPage, args.Cursor, listExternalServicesQuery(args),
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
  deleted_at,
  last_sync_at,
  next_sync_at,
  namespace_user_id
FROM external_services
WHERE id > %s
AND %s
ORDER BY id ASC LIMIT %s
`

const listRepoExternalServiceIDsSubquery = `
SELECT DISTINCT(external_service_id) repo_external_service_ids
FROM external_service_repos
WHERE repo_id IN (%s)
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

	switch {
	case args.NamespaceUserID > 0:
		preds = append(preds, sqlf.Sprintf("namespace_user_id = %d", args.NamespaceUserID))
	case args.NamespaceUserID == -1:
		preds = append(preds, sqlf.Sprintf("namespace_user_id IS NULL"))
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
		return svcs[i].ID, 1, err
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
			nullTimeColumn(s.LastSyncAt.UTC()),
			nullTimeColumn(s.NextSyncAt.UTC()),
			nullInt32Column(s.NamespaceUserID),
		))
	}

	return sqlf.Sprintf(
		upsertExternalServicesQueryFmtstr,
		sqlf.Join(vals, ",\n"),
	)
}

const upsertExternalServicesQueryValueFmtstr = `
  (COALESCE(NULLIF(%s, 0), (SELECT nextval('external_services_id_seq'))), UPPER(%s), %s, %s, %s, %s, %s, %s, %s, %s)
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
  deleted_at,
  last_sync_at,
  next_sync_at,
  namespace_user_id
)
VALUES %s
ON CONFLICT(id) DO UPDATE
SET
  kind         = UPPER(excluded.kind),
  display_name = excluded.display_name,
  config       = excluded.config,
  created_at   = excluded.created_at,
  updated_at   = excluded.updated_at,
  deleted_at   = excluded.deleted_at,
  last_sync_at = excluded.last_sync_at,
  next_sync_at = excluded.next_sync_at,
  namespace_user_id = excluded.namespace_user_id
RETURNING *
`

// InsertRepos inserts the given repos and their associated sources.
// It sets the ID field of each given repo to the value of its corresponding row.
func (s DBStore) InsertRepos(ctx context.Context, repos ...*Repo) error {
	records := make([]*repoRecord, 0, len(repos))

	for _, r := range repos {
		repoRec, err := newRepoRecord(r)
		if err != nil {
			return err
		}

		records = append(records, repoRec)
	}

	encodedRepos, err := json.Marshal(records)
	if err != nil {
		return err
	}

	q := sqlf.Sprintf(insertReposQuery, string(encodedRepos))

	rows, err := s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return errors.Wrap(err, "insert")
	}
	defer rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	for i := 0; rows.Next(); i++ {
		if err := rows.Scan(&repos[i].ID); err != nil {
			return err
		}
	}

	return nil
}

var insertReposQuery = `
WITH repos_list AS (
  SELECT * FROM ROWS FROM (
	json_to_recordset(%s)
	AS (
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
		metadata              jsonb,
		sources               jsonb
	  )
	)
	WITH ORDINALITY
),
inserted_repos AS (
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
	metadata
  FROM repos_list
  RETURNING id
),
inserted_repos_rows AS (
  SELECT id, ROW_NUMBER() OVER () AS rn FROM inserted_repos
),
repos_list_rows AS (
  SELECT *, ROW_NUMBER() OVER () AS rn FROM repos_list
),
inserted_repos_with_ids AS (
  SELECT
	inserted_repos_rows.id,
	repos_list_rows.*
  FROM repos_list_rows
  JOIN inserted_repos_rows USING (rn)
),
sources_list AS (
  SELECT
    inserted_repos_with_ids.id AS repo_id,
	sources.external_service_id AS external_service_id,
	sources.clone_url AS clone_url
  FROM
    inserted_repos_with_ids,
	jsonb_to_recordset(inserted_repos_with_ids.sources)
	  AS sources(
		external_service_id bigint,
		repo_id             integer,
		clone_url           text
	  )
),
insert_sources AS (
  INSERT INTO external_service_repos (
    external_service_id,
    repo_id,
    clone_url
  )
  SELECT
    external_service_id,
    repo_id,
    clone_url
  FROM sources_list
)
SELECT id FROM inserted_repos_with_ids;
`

// DeleteRepos deletes repos associated with the given ids and their associated sources.
func (s DBStore) DeleteRepos(ctx context.Context, ids ...api.RepoID) error {
	if len(ids) == 0 {
		return nil
	}

	// the number of deleted repos can potentially be higher
	// than the maximum number of arguments we can pass to postgres.
	// We pass them as a json array instead to overcome this limitation.
	encodedIds, err := json.Marshal(ids)
	if err != nil {
		return err
	}

	q := sqlf.Sprintf(deleteReposQuery, string(encodedIds))

	_, err = s.db.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return errors.Wrap(err, "delete")
	}

	return nil
}

const deleteReposQuery = `
WITH repo_ids AS (
  SELECT jsonb_array_elements_text(%s) AS id
)
UPDATE repo
SET
  name = soft_deleted_repository_name(name),
  deleted_at = transaction_timestamp()
FROM repo_ids
WHERE deleted_at IS NULL
AND repo.id = repo_ids.id::int
`

// ListRepos lists all stored repos that match the given arguments.
func (s DBStore) ListRepos(ctx context.Context, args StoreListReposArgs) (repos []*Repo, _ error) {
	return repos, s.paginate(ctx, args.Limit, args.PerPage, 0, listReposQuery(args),
		func(sc scanner) (last, count int64, err error) {
			var r Repo
			if err := scanRepo(&r, sc); err != nil {
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
  repo.external_service_id,
  external_id,
  archived,
  cloned,
  fork,
  private,
  (
	SELECT
	  json_agg(
	    json_build_object(
          'CloneURL', esr.clone_url,
          'ID', esr.external_service_id,
          'Kind', LOWER(svcs.kind)
	    )
	  )
	FROM external_service_repos AS esr
	JOIN external_services AS svcs ON esr.external_service_id = svcs.id
	WHERE
	    esr.repo_id = repo.id
	  AND
	    svcs.deleted_at IS NULL
  ),
  metadata
-- repo or repo joined with external_service_repos
FROM %s
WHERE id > %s
-- preds
AND %s
AND deleted_at IS NULL
-- join filters
AND %s
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

	fromClause := sqlf.Sprintf("repo")
	joinFilter := sqlf.Sprintf("TRUE")
	if args.ExternalServiceID != 0 {
		fromClause = sqlf.Sprintf("repo JOIN external_service_repos e ON repo.id = e.repo_id")
		joinFilter = sqlf.Sprintf("e.external_service_id = %d", args.ExternalServiceID)
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
			fromClause,
			cursor,
			sqlf.Sprintf("(%s)", predQ),
			joinFilter,
			limit,
		)
	}
}

func (s DBStore) UpsertSources(ctx context.Context, inserts, updates, deletes map[api.RepoID][]SourceInfo) error {
	type source struct {
		ExternalServiceID int64  `json:"external_service_id"`
		RepoID            int64  `json:"repo_id"`
		CloneURL          string `json:"clone_url"`
	}

	marshalSourceList := func(sources map[api.RepoID][]SourceInfo) ([]byte, error) {
		srcs := make([]source, 0, len(sources))
		for rid, infoList := range sources {
			for _, info := range infoList {
				srcs = append(srcs, source{
					ExternalServiceID: info.ExternalServiceID(),
					RepoID:            int64(rid),
					CloneURL:          info.CloneURL,
				})
			}
		}
		return json.Marshal(srcs)
	}

	insertedSources, err := marshalSourceList(inserts)
	if err != nil {
		return err
	}

	updatedSources, err := marshalSourceList(updates)
	if err != nil {
		return err
	}

	deletedSources, err := marshalSourceList(deletes)
	if err != nil {
		return err
	}

	q := sqlf.Sprintf(upsertSourcesQueryFmtstr,
		string(deletedSources),
		string(updatedSources),
		string(insertedSources),
	)

	_, err = s.db.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	return err
}

const upsertSourcesQueryFmtstr = `
-- source: cmd/repo-updater/repos/store.go:DBStore.UpsertSources
WITH deleted_sources_list AS (
  SELECT * FROM ROWS FROM (
	json_to_recordset(%s)
	AS (
		external_service_id bigint,
		repo_id             integer,
		clone_url           text
	)
  )
  WITH ORDINALITY
),
updated_sources_list AS (
  SELECT * FROM ROWS FROM (
    json_to_recordset(%s)
    AS (
      external_service_id bigint,
      repo_id             integer,
      clone_url           text
    )
  )
  WITH ORDINALITY
),
inserted_sources_list AS (
  SELECT * FROM ROWS FROM (
    json_to_recordset(%s)
    AS (
        external_service_id bigint,
        repo_id             integer,
        clone_url           text
    )
  )
  WITH ORDINALITY
),
delete_sources AS (
  DELETE FROM external_service_repos AS e
  USING deleted_sources_list AS d
  WHERE
	  e.external_service_id = d.external_service_id
	AND
      e.repo_id = d.repo_id
),
update_sources AS (
  UPDATE external_service_repos AS e
  SET
	clone_url = u.clone_url
  FROM updated_sources_list AS u
  WHERE
      e.repo_id = u.repo_id
	AND
	  e.external_service_id = u.external_service_id
)
INSERT INTO external_service_repos (
  external_service_id,
  repo_id,
  clone_url
) SELECT
  external_service_id,
  repo_id,
  clone_url
FROM inserted_sources_list
`

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

func (s DBStore) paginate(ctx context.Context, limit, perPage int64, cursor int64, q paginatedQuery, scan scanFunc) (err error) {
	const defaultPerPageLimit = 10000

	if perPage <= 0 {
		perPage = defaultPerPageLimit
	}

	if limit > 0 && perPage > limit {
		perPage = limit
	}

	var (
		remaining   = limit
		next, count int64
	)

	// We need this so that we enter the loop below for the first iteration
	// since cursor will be < next
	next = cursor
	cursor--

	for cursor < next && err == nil && (limit <= 0 || remaining > 0) {
		cursor = next
		next, count, err = s.list(ctx, q(cursor, perPage), scan)
		if limit > 0 {
			if remaining -= count; perPage > remaining {
				perPage = remaining
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
// This method does NOT update sources in the external_services_repo table.
// Use UpsertSources for that purpose.
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
		{"delete", batchDeleteReposQuery, deletes},
		{"update", batchUpdateReposQuery, updates},
		{"insert", batchInsertReposQuery, inserts},
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

func (s *DBStore) EnqueueSyncJobs(ctx context.Context, ignoreSiteAdmin bool) error {
	filter := "TRUE"
	if ignoreSiteAdmin {
		filter = "namespace_user_id IS NULL"
	}
	q := sqlf.Sprintf(enqueueSyncJobsQueryFmtstr, filter)
	_, err := s.db.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	return err
}

const enqueueSyncJobsQueryFmtstr = `
WITH due AS (
    SELECT id
    FROM external_services
    WHERE (next_sync_at <= clock_timestamp() OR next_sync_at IS NULL)
    AND %s
),
busy AS (
    SELECT DISTINCT external_service_id id FROM external_service_sync_jobs
    WHERE state = 'queued'
    OR state = 'processing'
)
INSERT INTO external_service_sync_jobs (external_service_id)
SELECT id from due EXCEPT SELECT id from busy
`

// ListSyncJobs returns all sync jobs.
func (s *DBStore) ListSyncJobs(ctx context.Context) ([]SyncJob, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT * FROM external_service_sync_jobs_with_next_sync_at")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanJobs(rows)
}

func scanJobs(rows *sql.Rows) ([]SyncJob, error) {
	var jobs []SyncJob

	for rows.Next() {
		var job SyncJob
		if err := rows.Scan(
			&job.ID,
			&job.State,
			&job.FailureMessage,
			&job.StartedAt,
			&job.FinishedAt,
			&job.ProcessAfter,
			&job.NumResets,
			&job.ExternalServiceID,
			&job.NextSyncAt,
		); err != nil {
			return nil, err
		}

		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return jobs, nil
}

func batchReposQuery(fmtstr string, repos []*Repo) (_ *sqlf.Query, err error) {
	records := make([]*repoRecord, 0, len(repos))
	for _, r := range repos {
		rec, err := newRepoRecord(r)
		if err != nil {
			return nil, err
		}

		records = append(records, rec)
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
      metadata              jsonb
    )
  )
  WITH ORDINALITY
)`

var batchUpdateReposQuery = batchReposQueryFmtstr + `
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
var batchDeleteReposQuery = batchReposQueryFmtstr + `
UPDATE repo
SET
  name = soft_deleted_repository_name(batch.name),
  deleted_at = batch.deleted_at
FROM batch
WHERE batch.deleted_at IS NOT NULL
AND repo.id = batch.id
`

var batchInsertReposQuery = batchReposQueryFmtstr + `
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

func nullInt32Column(i int32) *int32 {
	if i == 0 {
		return nil
	}
	return &i
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

func sourcesColumn(repoID api.RepoID, sources map[string]*SourceInfo) (json.RawMessage, error) {
	var records []sourceRecord
	for _, src := range sources {
		records = append(records, sourceRecord{
			ExternalServiceID: src.ExternalServiceID(),
			RepoID:            int64(repoID),
			CloneURL:          src.CloneURL,
		})
	}

	return json.MarshalIndent(records, "        ", "    ")
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
		&dbutil.NullTime{Time: &svc.LastSyncAt},
		&dbutil.NullTime{Time: &svc.NextSyncAt},
		&dbutil.NullInt32{N: &svc.NamespaceUserID},
	)
}

func scanRepo(r *Repo, s scanner) error {
	var sources dbutil.NullJSONRawMessage
	var metadata json.RawMessage
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

	type sourceInfo struct {
		ID       int64
		CloneURL string
		Kind     string
	}

	r.Sources = make(map[string]*SourceInfo)

	if sources.Raw != nil {
		var srcs []sourceInfo
		if err = json.Unmarshal(sources.Raw, &srcs); err != nil {
			return errors.Wrap(err, "scanRepo: failed to unmarshal sources")
		}
		for _, src := range srcs {
			urn := extsvc.URN(src.Kind, src.ID)
			r.Sources[urn] = &SourceInfo{
				ID:       urn,
				CloneURL: src.CloneURL,
			}
		}
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
