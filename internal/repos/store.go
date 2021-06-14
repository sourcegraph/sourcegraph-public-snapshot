package repos

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
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
	// RepoStore is a database.RepoStore using the same database handle.
	RepoStore *database.RepoStore
	// ExternalServiceStore is a database.ExternalServiceStore using the same database handle.
	ExternalServiceStore *database.ExternalServiceStore
	// Used to mock calls to certain methods.
	Mocks MockStore

	txtrace *trace.Trace
	txctx   context.Context
}

// NewStore instantiates and returns a new DBStore with prepared statements.
func NewStore(db dbutil.DB, txOpts sql.TxOptions) *Store {
	s := basestore.NewWithDB(db, txOpts)
	return &Store{
		Store:                s,
		RepoStore:            database.ReposWith(s),
		ExternalServiceStore: database.ExternalServicesWith(s),
		Log:                  log15.Root(),
		Tracer:               trace.Tracer{Tracer: opentracing.GlobalTracer()},
	}
}

func (s *Store) With(other basestore.ShareableStore) *Store {
	return &Store{
		Store:                s.Store.With(other),
		RepoStore:            s.RepoStore.With(other),
		ExternalServiceStore: s.ExternalServiceStore.With(other),
		Log:                  s.Log,
		Metrics:              s.Metrics,
		Tracer:               s.Tracer,
		Mocks:                s.Mocks,
	}
}

// Transact returns a TxStore whose methods operate within the context of a transaction.
func (s *Store) Transact(ctx context.Context) (stx *Store, err error) {
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
		return nil, errors.Wrap(err, "starting transaction")
	}
	return &Store{
		Store:                txBase,
		RepoStore:            s.RepoStore.With(txBase),
		ExternalServiceStore: s.ExternalServiceStore.With(txBase),
		Log:                  s.Log,
		Metrics:              s.Metrics,
		Tracer:               s.Tracer,
		Mocks:                s.Mocks,
		txtrace:              tr,
		txctx:                ctx,
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

func (s *Store) trace(ctx context.Context, family string) (*trace.Trace, context.Context) {
	txctx := s.txctx
	if txctx == nil {
		txctx = ctx
	}
	tr, txctx := s.Tracer.New(txctx, family, "")
	ctx = trace.CopyContext(ctx, txctx)
	return tr, ctx
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
-- source: internal/repos/store.go:DBStore.ListExternalRepoSpecs
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
		func(sc scanner) (last, count int64, err error) {
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

type externalServiceRepo struct {
	ExternalServiceID int64  `json:"external_service_id"`
	RepoID            int64  `json:"repo_id"`
	CloneURL          string `json:"clone_url"`
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

	type sourceSlices struct {
		externalServiceIDs []int64
		repoIDs            []int64
		cloneURLs          []string
	}

	makeSourceSlices := func(sources map[api.RepoID][]types.SourceInfo) sourceSlices {
		srcs := sourceSlices{
			externalServiceIDs: make([]int64, 0, len(sources)),
			repoIDs:            make([]int64, 0, len(sources)),
			cloneURLs:          make([]string, 0, len(sources)),
		}
		for rid, infoList := range sources {
			for _, info := range infoList {
				srcs.externalServiceIDs = append(srcs.externalServiceIDs, info.ExternalServiceID())
				srcs.repoIDs = append(srcs.repoIDs, int64(rid))
				srcs.cloneURLs = append(srcs.cloneURLs, info.CloneURL)
			}
		}
		return srcs
	}

	insertedSources := makeSourceSlices(inserts)
	updatedSources := makeSourceSlices(updates)

	var q *sqlf.Query

	if len(deletes) > 0 {
		deletedSources := makeSourceSlices(deletes)
		q = sqlf.Sprintf(upsertSourcesWithDeletesQueryFmtstr,
			// Updated
			pq.Int64Array(updatedSources.externalServiceIDs),
			pq.Int64Array(updatedSources.repoIDs),
			pq.StringArray(updatedSources.cloneURLs),
			// Inserted
			pq.Int64Array(insertedSources.externalServiceIDs),
			pq.Int64Array(insertedSources.repoIDs),
			pq.StringArray(insertedSources.cloneURLs),
			// Deleted
			pq.Int64Array(deletedSources.externalServiceIDs),
			pq.Int64Array(deletedSources.repoIDs),
		)
	} else {
		q = sqlf.Sprintf(upsertSourcesQueryFmtstr,
			// Updated
			pq.Int64Array(updatedSources.externalServiceIDs),
			pq.Int64Array(updatedSources.repoIDs),
			pq.StringArray(updatedSources.cloneURLs),
			// Inserted
			pq.Int64Array(insertedSources.externalServiceIDs),
			pq.Int64Array(insertedSources.repoIDs),
			pq.StringArray(insertedSources.cloneURLs),
		)
	}

	err = s.Exec(ctx, q)
	if err != nil {
		return err
	}

	if len(deletes) > 0 {
		// if we deleted some sources we must manually run the soft_delete_orphan_repo_by_external_service_repos function
		// to cleanup orphaned repos
		return s.Exec(ctx, sqlf.Sprintf(`SELECT soft_delete_orphan_repo_by_external_service_repos()`))
	}

	return nil
}

var upsertSourcesQueryFmtstr = upsertSourcesFmtstrPrefix + upsertSourcesFmtstrSuffix
var upsertSourcesWithDeletesQueryFmtstr = upsertSourcesFmtstrPrefix + upsertSourcesFmtstrDeletes + upsertSourcesFmtstrSuffix

const upsertSourcesFmtstrPrefix = `
-- source: internal/repos/store.go:DBStore.UpsertSources
WITH updated_sources_list AS (
  SELECT * FROM
  unnest(%s::bigint[], %s::integer[], %s::text[]) AS
  x ( external_service_id, repo_id, clone_url )
),
inserted_sources_list AS (
  SELECT * FROM
  unnest(%s::bigint[], %s::integer[], %s::text[]) AS
  x ( external_service_id, repo_id, clone_url )
),
`

const upsertSourcesFmtstrSuffix = `
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
  user_id,
  clone_url
) SELECT
  external_service_id,
  repo_id,
  es.namespace_user_id,
  clone_url
FROM inserted_sources_list
JOIN external_services es ON (id = external_service_id)
ON CONFLICT ON CONSTRAINT external_service_repos_repo_id_external_service_id_unique
DO
  UPDATE SET clone_url = EXCLUDED.clone_url
  WHERE external_service_repos.clone_url != EXCLUDED.clone_url
`

const upsertSourcesFmtstrDeletes = `
deleted_sources_list AS (
  SELECT * FROM
  unnest(%s::bigint[], %s::integer[]) AS
  x ( external_service_id, repo_id )
),
delete_sources AS (
  DELETE FROM external_service_repos AS e
  USING deleted_sources_list AS d
  WHERE
	  e.external_service_id = d.external_service_id
	AND
      e.repo_id = d.repo_id
),
`

// SetClonedRepos updates cloned status for all repositories.
// All repositories whose name is in repoNames will have their cloned column set to true
// and every other repository will have it set to false.
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
-- source: internal/repos/store.go:DBStore.SetClonedRepos
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

// CountUserAddedRepos counts the total number of repos that have been added
// by user owned external services.
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
	c, ok, err := basestore.ScanFirstInt(s.Query(ctx, q))
	if err != nil || !ok {
		return 0, err
	}
	return uint64(c), nil
}

const CountTotalUserAddedReposQueryFmtstr = `
-- source: internal/repos/store.go:DBStore.CountUserAddedRepos
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
            sr.user_id IS NOT NULL
            AND s.deleted_at IS NULL
            AND r.id = sr.repo_id
            AND r.deleted_at IS NULL)
`

// a paginatedQuery returns a query with the given pagination
// parameters
type paginatedQuery func(cursor, limit int64) *sqlf.Query

func (s *Store) paginate(ctx context.Context, limit, perPage int64, cursor int64, q paginatedQuery, scan scanFunc) (err error) {
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

func (s *Store) list(ctx context.Context, q *sqlf.Query, scan scanFunc) (last, count int64, err error) {
	rows, err := s.Query(ctx, q)
	if err != nil {
		return 0, 0, err
	}
	return scanAll(rows, scan)
}

// UpsertRepos updates or inserts the given repos in the Sourcegraph repository
// store. The ID field is used to distinguish between Repos that need to be
// updated and types.Repos that need to be inserted. On inserts, the _ID field of
// each given Repo is set on inserts. The cloned column is not updated by this
// function. This method does NOT update sources in the external_services_repo
// table. Use UpsertSources for that purpose.
func (s *Store) UpsertRepos(ctx context.Context, repos ...*types.Repo) (err error) {
	if s.Mocks.UpsertRepos != nil {
		return s.Mocks.UpsertRepos(ctx, repos...)
	}
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

// EnqueueSingleSyncJob enqueues a single sync job for the given external
// service if it is not already queued or processing.
func (s *Store) EnqueueSingleSyncJob(ctx context.Context, id int64) (err error) {
	q := sqlf.Sprintf(`
INSERT INTO external_service_sync_jobs (external_service_id)
SELECT %s
WHERE NOT EXISTS(
        SELECT 1
        FROM external_service_sync_jobs
        WHERE external_service_id = %s
          AND state IN ('queued', 'processing'))
`, id, id)
	return s.Exec(ctx, q)
}

// EnqueueSyncJobs enqueues sync jobs for all external services that are due.
func (s *Store) EnqueueSyncJobs(ctx context.Context, isCloud bool) (err error) {
	tr, ctx := s.trace(ctx, "Store.EnqueueSyncJobs")

	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		s.Metrics.EnqueueSyncJobs.Observe(secs, 0, &err)
		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	filter := "TRUE"
	// On Cloud we don't sync our default sources in the background, they are synced
	// on demand instead.
	if isCloud {
		filter = "cloud_default = false"
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
			execution_logs,
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
		// required field for the sync worker, but
		// the value is thrown out here
		var executionLogs *[]interface{}

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
			&executionLogs,
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
      stars                 integer,
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
  stars                 = batch.stars,
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
  stars,
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
  stars,
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

// repoRecord is the json representation of a repository as used in this package
// Postgres CTEs.
type repoRecord struct {
	ID                  api.RepoID      `json:"id"`
	Name                string          `json:"name"`
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
	Stars               int             `json:"stars"`
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
		Name:                string(r.Name),
		URI:                 nullStringColumn(r.URI),
		Description:         r.Description,
		CreatedAt:           r.CreatedAt.UTC(),
		UpdatedAt:           nullTimeColumn(r.UpdatedAt.UTC()),
		DeletedAt:           nullTimeColumn(r.DeletedAt.UTC()),
		ExternalServiceType: nullStringColumn(r.ExternalRepo.ServiceType),
		ExternalServiceID:   nullStringColumn(r.ExternalRepo.ServiceID),
		ExternalID:          nullStringColumn(r.ExternalRepo.ID),
		Archived:            r.Archived,
		Fork:                r.Fork,
		Stars:               r.Stars,
		Private:             r.Private,
		Metadata:            metadata,
		Sources:             sources,
	}, nil
}

// MockStore is used to mock calls to certain DBStore methods.
type MockStore struct {
	UpsertRepos func(ctx context.Context, repos ...*types.Repo) (err error)
}
