package internal

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/api"
	internaldb "github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/db/basestore"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// A Store exposes methods to read and write repos and external services.
type Store struct {
	*basestore.Store

	// Logger used by the store. Defaults to log15.Root().
	Log logging.ErrorLogger
	// Metrics are sent to Prometheus by default.
	Metrics StoreMetrics
	// Used for tracing calls to store methods. Uses opentracing.GlobalTracer() by default.
	Tracer trace.Tracer

	txtrace *trace.Trace
	txctx   context.Context
}

// NewStore instantiates and returns a new Store with prepared statements.
// Store wraps a basestore with error logging, Prometheus metrics and tracing.
func NewStore(db dbutil.DB, txOpts sql.TxOptions) *Store {
	m := NewStoreMetrics()
	m.MustRegister(prometheus.DefaultRegisterer)

	return &Store{
		Store:   basestore.NewWithDB(db, txOpts),
		Log:     log15.Root(),
		Metrics: m,
		Tracer:  trace.Tracer{Tracer: opentracing.GlobalTracer()},
	}
}

func (s *Store) trace(ctx context.Context, family string) (*trace.Trace, context.Context) {
	txctx := s.txctx
	if txctx == nil {
		txctx = ctx
	}
	tr, txctx := s.Tracer.New(txctx, family, "")
	ctx = trace.CopyContext(ctx, txctx)
	return tr, ctx
}

func (s *Store) Transact(ctx context.Context) (tx *Store, err error) {
	tr, ctx := s.trace(ctx, "Store.Transact")

	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		s.Metrics.Transact.Observe(secs, 1, &err)
		logging.Log(s.Log, "store.transact", &err)
		if err != nil {
			tr.SetError(err)
			// Finish is called in Done in the non-error case
			tr.Finish()
		}
	}(time.Now())

	txBase, err := s.Store.Transact(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "transact error")
	}

	return &Store{
		Store:   txBase,
		Log:     s.Log,
		Metrics: s.Metrics,
		Tracer:  s.Tracer,
		txtrace: tr,
		txctx:   ctx,
	}, nil
}

// Done calls into the inner Store Done method.
func (s *Store) Done(err error) error {
	tr := s.txtrace
	tr.LogFields(otlog.String("event", "Store.Done"))

	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		done := false

		if err != nil {
			done = true
			tr.SetError(err)
			s.Metrics.Done.Observe(secs, 1, &err)
			logging.Log(s.Log, "store.done", &err)
		}

		if !done {
			s.Metrics.Done.Observe(secs, 1, nil)
		}

		tr.Finish()
	}(time.Now())

	return s.Store.Done(err)
}

func (s *Store) With(other basestore.ShareableStore) *Store {
	return &Store{Store: s.Store.With(other)}
}

// RepoStore returns a db.ReposStore using the same database handle.
func (s *Store) RepoStore() *internaldb.RepoStore {
	return internaldb.Repos.With(s.Store)
}

// ExternalServiceStore returns a db.ExternalServiceStore using the same database handle.
func (s *Store) ExternalServiceStore() *internaldb.ExternalServiceStore {
	return internaldb.ExternalServices.With(s.Store)
}

// a paginatedQuery returns a query with the given pagination
// parameters
type paginatedQuery func(cursor, limit int64) *sqlf.Query

func (s *Store) paginate(ctx context.Context, limit, perPage int64, cursor int64, q paginatedQuery, scan dbutil.ScanFunc) (err error) {
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

func (s *Store) list(ctx context.Context, q *sqlf.Query, scan dbutil.ScanFunc) (last, count int64, err error) {
	rows, err := s.Query(ctx, q)
	if err != nil {
		return 0, 0, err
	}
	return dbutil.ScanAll(rows, scan)
}

func (s *Store) ListExternalRepoSpecs(ctx context.Context) (ids map[api.ExternalRepoSpec]struct{}, err error) {
	tr, ctx := s.trace(ctx, "Store.ListExternalRepoSpecs")

	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		count := float64(len(ids))

		s.Metrics.ListExternalRepoSpecs.Observe(secs, count, &err)
		logging.Log(s.Log, "store.list-external-repo-specs", &err,
			"count", len(ids),
		)

		tr.LogFields(otlog.Int("count", len(ids)))
		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	const ListExternalRepoSpecsQueryFmtstr = `
-- source: cmd/repo-updater/repos/store.go:DBStore.ListExternalRepoSpecs
SELECT
	id,
	external_id,
	external_service_type,
	external_service_id
FROM repo
WHERE
	deleted_at IS NULL
AND	external_id IS NOT NULL
AND	external_service_type IS NOT NULL
AND	external_service_id IS NOT NULL
AND	id > %s
ORDER BY id ASC LIMIT %s
`
	paginatedQuery := func(cursor, limit int64) *sqlf.Query {
		return sqlf.Sprintf(
			ListExternalRepoSpecsQueryFmtstr,
			cursor,
			limit,
		)
	}
	ids = make(map[api.ExternalRepoSpec]struct{})
	return ids, s.paginate(ctx, 0, 0, 0, paginatedQuery,
		func(sc dbutil.Scanner) (last, count int64, err error) {
			var id int64
			var spec api.ExternalRepoSpec
			if err := sc.Scan(&id, &spec.ID, &spec.ServiceType, &spec.ServiceID); err != nil {
				return 0, 0, err
			}

			ids[spec] = struct{}{}
			return id, 1, nil
		},
	)
}

// ListSyncJobs returns all sync jobs.
func (s *Store) ListSyncJobs(ctx context.Context) ([]SyncJob, error) {
	q := sqlf.Sprintf(`
		SELECT
			id,
			state,
			failure_message,
			started_at,
			finished_at,
			process_after,
			num_resets,
			num_failures,
			external_service_id,
			next_sync_at
		 FROM external_service_sync_jobs_with_next_sync_at
	`)
	rows, err := s.Query(ctx, q)
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
			&job.NumFailures,
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

// UpsertRepos updates or inserts the given repos in the Sourcegraph repository store.
// The ID field is used to distinguish between Repos that need to be updated and Repos
// that need to be inserted. On inserts, the _ID field of each given Repo is set on inserts.
// The cloned column is not updated by this function.
// This method does NOT update sources in the external_services_repo table.
// Use UpsertSources for that purpose.
func (s *Store) UpsertRepos(ctx context.Context, repos ...*types.Repo) (err error) {
	tr, ctx := s.trace(ctx, "Store.UpsertRepos")
	tr.LogFields(otlog.Int("count", len(repos)))

	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		count := float64(len(repos))

		s.Metrics.UpsertRepos.Observe(secs, count, &err)
		logging.Log(s.Log, "store.upsert-repos", &err, "count", len(repos))

		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	if len(repos) == 0 {
		return nil
	}

	var deletes, updates, inserts []*types.Repo
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
		repos []*types.Repo
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

		rows, err := s.Query(ctx, q)
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

		_, _, err = dbutil.ScanAll(rows, func(sc dbutil.Scanner) (last, count int64, err error) {
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

func batchReposQuery(fmtstr string, repos []*types.Repo) (_ *sqlf.Query, err error) {
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

func (s *Store) UpsertSources(ctx context.Context, inserts, updates, deletes map[api.RepoID][]types.SourceInfo) (err error) {
	tr, ctx := s.trace(ctx, "Store.UpsertSources")
	tr.LogFields(otlog.Int("count", len(inserts)+len(deletes)))

	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		count := float64(len(inserts) + len(updates) + len(deletes))

		s.Metrics.UpsertSources.Observe(secs, count, &err)
		logging.Log(s.Log, "store.upsert-sources", &err, "count", count)

		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	if len(inserts)+len(updates)+len(deletes) == 0 {
		return nil
	}

	marshalSourceList := func(sources map[api.RepoID][]types.SourceInfo) ([]byte, error) {
		srcs := make([]externalServiceRepo, 0, len(sources))
		for rid, infoList := range sources {
			for _, info := range infoList {
				srcs = append(srcs, externalServiceRepo{
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

	err = s.Exec(ctx, q)
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
ON CONFLICT ON CONSTRAINT external_service_repos_repo_id_external_service_id_unique
DO
  UPDATE SET clone_url = EXCLUDED.clone_url
  WHERE external_service_repos.clone_url != EXCLUDED.clone_url
`

func (s *Store) SetClonedRepos(ctx context.Context, repoNames ...string) (err error) {
	tr, ctx := s.trace(ctx, "Store.SetClonedRepos")
	tr.LogFields(otlog.Int("count", len(repoNames)))

	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		count := float64(len(repoNames))

		s.Metrics.SetClonedRepos.Observe(secs, count, &err)
		logging.Log(s.Log, "store.set-cloned-repos", &err, "count", len(repoNames))

		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	if len(repoNames) == 0 {
		return nil
	}

	q := sqlf.Sprintf(setClonedReposQueryFmtstr, pq.StringArray(repoNames))

	return s.Exec(ctx, q)
}

const setClonedReposQueryFmtstr = `
-- source: cmd/repo-updater/repos/store.go:DBStore.SetClonedRepos
WITH repo_names AS (
  SELECT unnest(%s::citext[])::citext AS name
),
cloned_repos AS (
  SELECT repo.id AS id FROM repo_names JOIN repo ON repo.name = repo_names.name
),
not_cloned AS (
  UPDATE repo SET cloned = false
  WHERE NOT EXISTS (SELECT FROM cloned_repos WHERE repo.id = id) AND cloned
)
UPDATE repo
SET cloned = true
WHERE repo.id IN (SELECT id FROM cloned_repos) AND NOT cloned
`

func (s *Store) CountNotClonedRepos(ctx context.Context) (count uint64, err error) {
	tr, ctx := s.trace(ctx, "Store.CountNotClonedRepos")

	defer func(began time.Time) {
		secs := time.Since(began).Seconds()

		s.Metrics.CountNotClonedRepos.Observe(secs, float64(count), &err)
		logging.Log(s.Log, "store.count-not-cloned-repos", &err, "count", count)

		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	q := sqlf.Sprintf(CountNotClonedReposQueryFmtstr)
	total, ok, err := basestore.ScanFirstInt(s.Query(ctx, q))
	if err != nil || !ok {
		return 0, err
	}
	return uint64(total), nil
}

const CountNotClonedReposQueryFmtstr = `
-- source: cmd/repo-updater/repos/store.go:DBStore.CountNotClonedRepos
SELECT COUNT(*) FROM repo WHERE deleted_at IS NULL AND NOT cloned
`

func (s *Store) CountUserAddedRepos(ctx context.Context) (count uint64, err error) {
	tr, ctx := s.trace(ctx, "Store.CountUserAddedRepos")

	defer func(began time.Time) {
		secs := time.Since(began).Seconds()

		s.Metrics.CountUserAddedRepos.Observe(secs, float64(count), &err)
		logging.Log(s.Log, "store.count-user-added-repos", &err, "count", count)

		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	q := sqlf.Sprintf(CountTotalUserAddedReposQueryFmtstr)
	total, ok, err := basestore.ScanFirstInt(s.Query(ctx, q))
	if err != nil || !ok {
		return 0, err
	}
	return uint64(total), nil
}

const CountTotalUserAddedReposQueryFmtstr = `
-- source: cmd/repo-updater/repos/store.go:DBStore.CountUserAddedRepos
SELECT COUNT(*)
FROM
    repo r
WHERE
    EXISTS (
        SELECT
        FROM
            external_service_repos sr
            INNER JOIN external_services s ON s.id = sr.external_service_id
        WHERE
            s.namespace_user_id IS NOT NULL
            AND s.deleted_at IS NULL
            AND r.id = sr.repo_id
            AND r.deleted_at IS NULL)
`

// EnqueueSyncJobs adds all external services whose next_sync_at column is in the past to
// the external_service_sync_jobs table.
// If ignoreSiteAdmin is true, it will not select external services owned by the global admin user.
// This option is set to true on Cloud.
func (s *Store) EnqueueSyncJobs(ctx context.Context, ignoreSiteAdmin bool) (err error) {
	tr, ctx := s.trace(ctx, "Store.EnqueueSyncJobs")

	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		s.Metrics.EnqueueSyncJobs.Observe(secs, 0, &err)
		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	filter := "TRUE"
	if ignoreSiteAdmin {
		filter = "namespace_user_id IS NOT NULL"
	}
	q := sqlf.Sprintf(enqueueSyncJobsQueryFmtstr, sqlf.Sprintf(filter))
	return s.Exec(ctx, q)
}

// We ignore Phabricator repos here as they are currently synced using
// RunPhabricatorRepositorySyncWorker
const enqueueSyncJobsQueryFmtstr = `
WITH due AS (
    SELECT id
    FROM external_services
    WHERE (next_sync_at <= clock_timestamp() OR next_sync_at IS NULL)
    AND deleted_at IS NULL
    AND LOWER(kind) != 'phabricator'
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

// repoRecord is the json representation of a repository as used in this package
// Postgres CTEs.
type repoRecord struct {
	ID                  api.RepoID      `json:"id"`
	Name                api.RepoName    `json:"name"`
	URI                 *string         `json:"uri,omitempty"`
	Description         string          `json:"description"`
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

func newRepoRecord(r *types.Repo) (*repoRecord, error) {
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
		URI:                 dbutil.NullStringColumn(r.URI),
		Description:         r.Description,
		CreatedAt:           r.CreatedAt.UTC(),
		UpdatedAt:           dbutil.NullTimeColumn(r.UpdatedAt.UTC()),
		DeletedAt:           dbutil.NullTimeColumn(r.DeletedAt.UTC()),
		ExternalServiceType: dbutil.NullStringColumn(r.ExternalRepo.ServiceType),
		ExternalServiceID:   dbutil.NullStringColumn(r.ExternalRepo.ServiceID),
		ExternalID:          dbutil.NullStringColumn(r.ExternalRepo.ID),
		Archived:            r.Archived,
		Fork:                r.Fork,
		Private:             r.Private,
		Metadata:            metadata,
		Sources:             sources,
	}, nil
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

type externalServiceRepo struct {
	ExternalServiceID int64  `json:"external_service_id"`
	RepoID            int64  `json:"repo_id"`
	CloneURL          string `json:"clone_url"`
}

func sourcesColumn(repoID api.RepoID, sources map[string]*types.SourceInfo) (json.RawMessage, error) {
	var records []externalServiceRepo
	for _, src := range sources {
		records = append(records, externalServiceRepo{
			ExternalServiceID: src.ExternalServiceID(),
			RepoID:            int64(repoID),
			CloneURL:          src.CloneURL,
		})
	}

	return json.MarshalIndent(records, "        ", "    ")
}
