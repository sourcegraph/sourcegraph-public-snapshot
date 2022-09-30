package repos

import (
	"context"
	"database/sql"
	"encoding/json"
	"sort"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	otlog "github.com/opentracing/opentracing-go/log"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Store interface {
	// RepoStore returns a database.RepoStore using the same database handle.
	RepoStore() database.RepoStore
	// GitserverReposStore returns a database.GitserverReposStore using the same
	// database handle.
	GitserverReposStore() database.GitserverRepoStore
	// ExternalServiceStore returns a database.ExternalServiceStore using the same
	// database handle.
	ExternalServiceStore() database.ExternalServiceStore

	// SetMetrics updates metrics for the store in place.
	SetMetrics(m StoreMetrics)
	// SetTracer updates tracer for the store in place.
	SetTracer(t trace.Tracer)

	basestore.ShareableStore
	With(other basestore.ShareableStore) Store
	// Transact begins a new transaction and make a new Store over it.
	Transact(ctx context.Context) (Store, error)
	Done(err error) error

	// CountNamespacedRepos counts the total number of repos that have been added by
	// user or organization owned external services. If userID is specified, only
	// repos owned by that user are counted. If orgID is specified, only repos owned
	// by that organization are counted.
	CountNamespacedRepos(ctx context.Context, userID, orgID int32) (count uint64, err error)
	// DeleteExternalServiceReposNotIn calls DeleteExternalServiceRepo for every repo
	// not in the given ids that is owned by the given external service. We run one
	// query per repo rather than one batch query in order to reduce the chances of
	// this whole operation blocking on locks other queries acquire when referencing
	// external_service_repos or repo. Since the syncer runs periodically, it's
	// better to fail to delete some repos and try to delete them again in the next
	// run, than to have one failure prevent all deletes from happening.
	DeleteExternalServiceReposNotIn(ctx context.Context, svc *types.ExternalService, ids map[api.RepoID]struct{}) (deleted []api.RepoID, err error)
	// DeleteExternalServiceRepo deletes a repo's association to an external service
	// and the repo itself if there are no more associations to that repo by any
	// other external service.
	DeleteExternalServiceRepo(ctx context.Context, svc *types.ExternalService, id api.RepoID) (err error)
	// ListExternalServiceUserIDsByRepoID returns the user IDs associated with a
	// given repository. These users have proven that they have read access to the
	// repository given records are present in the "external_service_repos" table.
	ListExternalServiceUserIDsByRepoID(ctx context.Context, repoID api.RepoID) (userIDs []int32, err error)
	// ListExternalServicePrivateRepoIDsByUserID returns the private repo IDs
	// associated with a given user. As with ListExternalServiceUserIDsByRepoID, the
	// user has already proven that they have read access to the repositories since
	// records are present in the "external_service_repos" table.
	ListExternalServicePrivateRepoIDsByUserID(ctx context.Context, userID int32) (repoIDs []api.RepoID, err error)
	// CreateExternalServiceRepo inserts a single repo and its association to an
	// external service, respectively in the repo and "external_service_repos" table.
	// The associated external service must already exist.
	CreateExternalServiceRepo(ctx context.Context, svc *types.ExternalService, r *types.Repo) (err error)
	// UpdateExternalServiceRepo updates a single repo and its association to an
	// external service, respectively in the repo and external_service_repos table.
	// The associated external service must already exist.
	UpdateExternalServiceRepo(ctx context.Context, svc *types.ExternalService, r *types.Repo) (err error)
	// UpdateRepo updates a single repo without updating its association to an
	// external service. This must only be used when updating metadata on a repo
	// that cannot affect its associations.
	UpdateRepo(ctx context.Context, r *types.Repo) (saved *types.Repo, err error)
	// EnqueueSingleSyncJob enqueues a single sync job for the given external service
	// if it is not already queued or processing. Additionally, it also skips
	// queueing up a sync job for cloud_default external services. This is done to
	// avoid the sync job for the cloud_default triggering a deletion of repos
	// because:
	//  1. cloud_default does not define any repos in its config
	//  2. repos under the cloud_default are lazily synced the first time a user accesses them
	//
	// This is a limitation of our current repo syncing architecture. The
	// cloud_default flag is only set on sourcegraph.com and manages public GitHub
	// and GitLab repositories that have been lazily synced.
	EnqueueSingleSyncJob(ctx context.Context, extSvcID int64) (err error)
	// EnqueueSyncJobs enqueues sync jobs for all external services that are due.
	EnqueueSyncJobs(ctx context.Context, isCloud bool) (err error)
	// ListSyncJobs returns all sync jobs.
	ListSyncJobs(ctx context.Context) ([]SyncJob, error)
}

// A Store exposes methods to read and write repos and external services.
type store struct {
	*basestore.Store

	// Logger used by the store. Does not have a default - it must be provided.
	Logger log.Logger
	// Metrics are sent to Prometheus by default.
	Metrics StoreMetrics
	// Used for tracing calls to store methods. Uses opentracing.GlobalTracer() by default.
	Tracer trace.Tracer

	txtrace *trace.Trace
	txctx   context.Context
}

// NewStore instantiates and returns a new Store with given database handle.
func NewStore(logger log.Logger, db database.DB) Store {
	s := basestore.NewWithHandle(db.Handle())
	return &store{
		Store:  s,
		Logger: logger,
		Tracer: trace.Tracer{TracerProvider: otel.GetTracerProvider()},
	}
}

func (s *store) RepoStore() database.RepoStore {
	return database.ReposWith(s.Logger, s)
}

func (s *store) GitserverReposStore() database.GitserverRepoStore {
	return database.GitserverReposWith(s)
}

func (s *store) ExternalServiceStore() database.ExternalServiceStore {
	return database.ExternalServicesWith(s.Logger, s)
}

func (s *store) SetMetrics(m StoreMetrics) { s.Metrics = m }
func (s *store) SetTracer(t trace.Tracer)  { s.Tracer = t }

func (s *store) With(other basestore.ShareableStore) Store {
	return &store{
		Store:   s.Store.With(other),
		Logger:  s.Logger,
		Metrics: s.Metrics,
		Tracer:  s.Tracer,
	}
}

func (s *store) Transact(ctx context.Context) (Store, error) {
	return s.transact(ctx)
}

func (s *store) transact(ctx context.Context) (stx *store, err error) {
	tr, ctx := s.trace(ctx, "Store.Transact")
	logger := trace.Logger(ctx, s.Logger)

	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		s.Metrics.Transact.Observe(secs, 1, &err)

		if err != nil {
			logger.Error("store.transact", log.Error(err))

			tr.SetError(err)
			// Finish is called in Done in the non-error case
			tr.Finish()
		}
	}(time.Now())

	txBase, err := s.Store.Transact(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "starting transaction")
	}
	return &store{
		Store:   txBase,
		Logger:  s.Logger,
		Metrics: s.Metrics,
		Tracer:  s.Tracer,
		txtrace: tr,
		txctx:   ctx,
	}, nil
}

// Done calls into the inner Store Done method.
func (s *store) Done(err error) error {
	tr := s.txtrace
	tr.LogFields(otlog.String("event", "Store.Done"))
	logger := trace.Logger(s.txctx, s.Logger)

	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		done := false

		if err != nil {
			done = true
			tr.SetError(err)
			s.Metrics.Done.Observe(secs, 1, &err)

			logger.Error("store.done", log.Error(err))
		}

		if !done {
			s.Metrics.Done.Observe(secs, 1, nil)
		}

		tr.Finish()
	}(time.Now())

	return s.Store.Done(err)
}

func (s *store) trace(ctx context.Context, family string) (*trace.Trace, context.Context) {
	txctx := s.txctx
	if txctx == nil {
		txctx = ctx
	}
	tr, txctx := s.Tracer.New(txctx, family, "")
	ctx = trace.CopyContext(ctx, txctx)
	return tr, ctx
}

func (s *store) CountNamespacedRepos(ctx context.Context, userID, orgID int32) (count uint64, err error) {
	tr, ctx := s.trace(ctx, "Store.CountNamespacedRepos")
	logger := trace.Logger(ctx, s.Logger).With(
		log.Int("userID", int(userID)),
		log.Int("orgID", int(orgID)),
	)

	defer func(began time.Time) {
		secs := time.Since(began).Seconds()

		tr.LogFields(otlog.Int32("user-id", userID), otlog.Int32("org-id", orgID))
		s.Metrics.CountNamespacedRepos.Observe(secs, float64(count), &err)

		if err != nil {
			logger.Error("store.count-namespaced-repos", log.Uint64("count", count), log.Error(err))
		}

		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	var q *sqlf.Query
	if userID > 0 {
		q = sqlf.Sprintf(countTotalNamespacedReposQueryFmtstr+"\nAND user_id = %d", userID)
	} else if orgID > 0 {
		q = sqlf.Sprintf(countTotalNamespacedReposQueryFmtstr+"\nAND org_id = %d", orgID)
	} else {
		q = sqlf.Sprintf(countTotalNamespacedReposQueryFmtstr)
	}

	err = s.QueryRow(ctx, q).Scan(&count)
	return count, err
}

const countTotalNamespacedReposQueryFmtstr = `
SELECT COUNT(DISTINCT(repo_id))
FROM external_service_repos
WHERE (user_id IS NOT NULL OR org_id IS NOT NULL)`

func (s *store) DeleteExternalServiceReposNotIn(ctx context.Context, svc *types.ExternalService, ids map[api.RepoID]struct{}) (deleted []api.RepoID, err error) {
	tr, ctx := s.trace(ctx, "Store.DeleteExternalServiceReposNotIn")
	tr.LogFields(
		otlog.Int("len(ids)", len(ids)),
		otlog.Int64("external_service_id", svc.ID),
	)
	logger := trace.Logger(ctx, s.Logger).With(log.Int64("externalServiceID", svc.ID), log.Int("len(ids)", len(ids)))

	defer func(began time.Time) {
		secs := time.Since(began).Seconds()

		s.Metrics.DeleteExternalServiceReposNotIn.Observe(secs, 1, &err)

		if err != nil {
			logger.Error("store.delete-external-service-repos-not-in", log.Error(err))
		}

		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	set := make(pq.Int64Array, 0, len(ids))
	for id := range ids {
		set = append(set, int64(id))
	}

	sort.Slice(set, func(a, b int) bool { return set[a] < set[b] })

	var toDelete pq.Int64Array
	if err = s.QueryRow(ctx, sqlf.Sprintf(listExternalServiceReposNotInQuery, svc.ID, set)).Scan(&toDelete); err != nil {
		return nil, errors.Wrap(err, "failed to list external service repo ids")
	}

	var errs error
	for _, id := range toDelete {
		if err = s.DeleteExternalServiceRepo(ctx, svc, api.RepoID(id)); err != nil {
			errs = errors.Append(errs, errors.Wrapf(err, "failed to delete external service repo (%d, %d)", svc.ID, id))
		} else {
			deleted = append(deleted, api.RepoID(id))
		}
	}

	return deleted, errs
}

const listExternalServiceReposNotInQuery = `
SELECT array_agg(repo_id)
FROM external_service_repos
WHERE external_service_id = %s AND repo_id != ALL(%s)
`

func (s *store) DeleteExternalServiceRepo(ctx context.Context, svc *types.ExternalService, id api.RepoID) (err error) {
	tr, ctx := s.trace(ctx, "Store.DeleteExternalServiceRepo")
	tr.LogFields(
		otlog.Int32("id", int32(id)),
		otlog.Int64("external_service_id", svc.ID),
	)
	logger := trace.Logger(ctx, s.Logger).With(log.Int64("externalServiceID", svc.ID), log.Int("repoID", int(id)))

	defer func(began time.Time) {
		secs := time.Since(began).Seconds()

		s.Metrics.DeleteExternalServiceRepo.Observe(secs, 1, &err)

		if err != nil {
			logger.Error("store.delete-external-service-repo", log.Error(err))
		}

		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	if !s.InTransaction() {
		s, err = s.transact(ctx)
		if err != nil {
			return errors.Wrap(err, "DeleteExternalServiceRepo")
		}
		defer func() { s.Done(err) }()
	}

	err = s.Exec(ctx, sqlf.Sprintf(deleteExternalServiceRepoQuery, svc.ID, id))
	if err != nil {
		return errors.Wrap(err, "failed to delete external service repo")
	}

	err = s.Exec(ctx, sqlf.Sprintf(deleteRepoIfOrphanQuery, id, id))
	if err != nil {
		return errors.Wrap(err, "failed to delete orphaned repo")
	}

	return nil
}

const deleteExternalServiceRepoQuery = `
DELETE FROM external_service_repos
WHERE external_service_id = %s AND repo_id = %s
`

const deleteRepoIfOrphanQuery = `
UPDATE repo
SET name = soft_deleted_repository_name(name), deleted_at = now()
WHERE id = %s AND NOT EXISTS (
	SELECT FROM external_service_repos
	WHERE repo_id = %s LIMIT 1
)
`

const listExternalServiceUserIDsByRepoIDQuery = `
SELECT user_id FROM external_service_repos
WHERE repo_id = %s AND user_id IS NOT NULL
`

func (s *store) ListExternalServiceUserIDsByRepoID(ctx context.Context, repoID api.RepoID) (userIDs []int32, err error) {
	tr, ctx := s.trace(ctx, "Store.ListExternalServiceUserIDsByRepoID")
	tr.LogFields(
		otlog.Int32("repo_id", int32(repoID)),
	)
	logger := trace.Logger(ctx, s.Logger).With(log.Int("repoID", int(repoID)))

	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		s.Metrics.ListExternalServiceUserIDsByRepoID.Observe(secs, 1, &err)

		if err != nil {
			logger.Error("store.list-external-service-user-ids-by-repo-id", log.Error(err))
		}

		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	q := sqlf.Sprintf(listExternalServiceUserIDsByRepoIDQuery, repoID)
	return basestore.ScanInt32s(s.Query(ctx, q))
}

const listExternalServiceRepoIDsByUserIDQuery = `
SELECT repo_id
FROM external_service_repos esr
JOIN repo ON repo.id = esr.repo_id
WHERE
	user_id = %s
AND repo.private
`

func (s *store) ListExternalServicePrivateRepoIDsByUserID(ctx context.Context, userID int32) (repoIDs []api.RepoID, err error) {
	tr, ctx := s.trace(ctx, "Store.ListExternalServicePrivateRepoIDsByUserID")
	tr.LogFields(
		otlog.Int32("user_id", userID),
	)
	logger := trace.Logger(ctx, s.Logger).With(log.Int("userID", int(userID)))

	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		s.Metrics.ListExternalServiceRepoIDsByUserID.Observe(secs, 1, &err)

		if err != nil {
			logger.Error("store.list-external-service-private-repo-id-by-user-id", log.Error(err))
		}

		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	q := sqlf.Sprintf(listExternalServiceRepoIDsByUserIDQuery, userID)
	ids, err := basestore.ScanInt32s(s.Query(ctx, q))
	if err != nil {
		return nil, err
	}

	repoIDs = make([]api.RepoID, len(ids))
	for i := range ids {
		repoIDs[i] = api.RepoID(ids[i])
	}
	return repoIDs, nil
}

func (s *store) CreateExternalServiceRepo(ctx context.Context, svc *types.ExternalService, r *types.Repo) (err error) {
	tr, ctx := s.trace(ctx, "Store.CreateExternalServiceRepo")
	tr.LogFields(
		otlog.String("name", string(r.Name)),
		otlog.Int64("external_service_id", svc.ID),
		otlog.String("external_repo_spec", r.ExternalRepo.String()),
	)
	logger := trace.Logger(ctx, s.Logger).With(
		log.Int("externalServiceID", int(svc.ID)),
		log.String("Name", string(r.Name)),
		log.Object("ExternalRepo",
			log.String("ID", r.ExternalRepo.ID),
			log.String("ServiceID", r.ExternalRepo.ServiceID),
			log.String("ServiceType", r.ExternalRepo.ServiceType),
		),
	)

	defer func(began time.Time) {
		secs := time.Since(began).Seconds()

		s.Metrics.CreateExternalServiceRepo.Observe(secs, 1, &err)

		if err != nil {
			logger.Error("store.create-external-service-repo", log.Error(err))
		}

		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	metadata, err := json.Marshal(r.Metadata)
	if err != nil {
		return err
	}

	q := sqlf.Sprintf(createRepoQuery,
		r.Name,
		r.URI,
		r.Description,
		r.ExternalRepo.ServiceType,
		r.ExternalRepo.ServiceID,
		r.ExternalRepo.ID,
		r.Archived,
		r.Fork,
		r.Stars,
		r.Private,
		metadata,
	)

	src := r.Sources[svc.URN()]
	if src == nil || src.CloneURL == "" {
		return errors.New("CreateExternalServiceRepo: repo missing source info for external service")
	}

	if !s.InTransaction() {
		s, err = s.transact(ctx)
		if err != nil {
			return errors.Wrap(err, "CreateExternalServiceRepo")
		}
		defer func() { s.Done(err) }()
	}

	if err = s.QueryRow(ctx, q).Scan(&r.ID, &r.CreatedAt); err != nil {
		return err
	}

	return s.Exec(ctx, sqlf.Sprintf(upsertExternalServiceRepoQuery,
		svc.ID,
		r.ID,
		svc.NamespaceUserID,
		svc.NamespaceOrgID,
		src.CloneURL,
	))
}

const createRepoQuery = `
INSERT INTO repo (
	name,
	uri,
	description,
	external_service_type,
	external_service_id,
	external_id,
	archived,
	fork,
	stars,
	private,
	metadata,
	created_at
)
VALUES (%s, NULLIF(%s, ''), %s, %s, %s, %s, %s, %s, %s, %s, %s, now())
RETURNING id, created_at
`

const upsertExternalServiceRepoQuery = `
INSERT INTO external_service_repos (
	external_service_id,
	repo_id,
	user_id,
	org_id,
	clone_url
)
VALUES (%s, %s, NULLIF(%s, 0), NULLIF(%s, 0), %s)
ON CONFLICT (external_service_id, repo_id)
DO UPDATE SET
	clone_url = excluded.clone_url,
	user_id   = excluded.user_id,
	org_id    =  excluded.org_id
WHERE
	external_service_repos.clone_url != excluded.clone_url OR
	external_service_repos.user_id   != excluded.user_id OR
	external_service_repos.org_id    != excluded.org_id
`

func (s *store) UpdateRepo(ctx context.Context, r *types.Repo) (saved *types.Repo, err error) {
	tr, ctx := s.trace(ctx, "Store.UpdateRepo")
	tr.LogFields(
		otlog.String("name", string(r.Name)),
		otlog.Int32("id", int32(r.ID)),
	)
	logger := trace.Logger(ctx, s.Logger).With(
		log.Int32("id", int32(r.ID)),
		log.String("name", string(r.Name)),
	)

	defer func(began time.Time) {
		secs := time.Since(began).Seconds()

		s.Metrics.UpdateRepo.Observe(secs, 1, &err)
		if err != nil {
			logger.Error("store.update-repo", log.Error(err))
		}

		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	if r.ID == 0 {
		return nil, errors.New("empty repo id in update")
	}

	metadata, err := metadataColumn(r.Metadata)
	if err != nil {
		return nil, errors.Wrap(err, "metadata marshalling failed")
	}

	q := sqlf.Sprintf(updateRepoQuery,
		r.Name,
		r.URI,
		r.Description,
		r.ExternalRepo.ServiceType,
		r.ExternalRepo.ServiceID,
		r.ExternalRepo.ID,
		r.Archived,
		r.Fork,
		r.Stars,
		r.Private,
		metadata,
		r.ID,
	)

	if err = s.QueryRow(ctx, q).Scan(&r.UpdatedAt); err != nil {
		return nil, err
	}
	return r, nil
}

func (s *store) UpdateExternalServiceRepo(ctx context.Context, svc *types.ExternalService, r *types.Repo) (err error) {
	tr, ctx := s.trace(ctx, "Store.UpdateExternalServiceRepo")
	tr.LogFields(
		otlog.String("name", string(r.Name)),
		otlog.Int64("external_service_id", svc.ID),
		otlog.String("external_repo_spec", r.ExternalRepo.String()),
	)
	logger := trace.Logger(ctx, s.Logger).With(
		log.Int("externalServiceID", int(svc.ID)),
		log.String("Name", string(r.Name)),
		log.Object("ExternalRepo",
			log.String("ID", r.ExternalRepo.ID),
			log.String("ServiceID", r.ExternalRepo.ServiceID),
			log.String("ServiceType", r.ExternalRepo.ServiceType),
		),
	)

	defer func(began time.Time) {
		secs := time.Since(began).Seconds()

		s.Metrics.UpdateExternalServiceRepo.Observe(secs, 1, &err)
		if err != nil {
			logger.Error("store.update-external-service-repo", log.Error(err))
		}

		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	if r.ID == 0 {
		return errors.New("empty repo id in update")
	}

	metadata, err := metadataColumn(r.Metadata)
	if err != nil {
		return errors.Wrapf(err, "metadata marshalling failed")
	}

	q := sqlf.Sprintf(updateRepoQuery,
		r.Name,
		r.URI,
		r.Description,
		r.ExternalRepo.ServiceType,
		r.ExternalRepo.ServiceID,
		r.ExternalRepo.ID,
		r.Archived,
		r.Fork,
		r.Stars,
		r.Private,
		metadata,
		r.ID,
	)

	src := r.Sources[svc.URN()]
	if src == nil || src.CloneURL == "" {
		return errors.New("UpdateExternalServiceRepo: repo missing source info for external service")
	}

	if !s.InTransaction() {
		s, err = s.transact(ctx)
		if err != nil {
			return errors.Wrap(err, "UpdateExternalServiceRepo")
		}
		defer func() { _ = s.Done(err) }()
	}

	if err = s.QueryRow(ctx, q).Scan(&r.UpdatedAt); err != nil {
		return err
	}

	return s.Exec(ctx, sqlf.Sprintf(upsertExternalServiceRepoQuery,
		svc.ID,
		r.ID,
		svc.NamespaceUserID,
		svc.NamespaceOrgID,
		src.CloneURL,
	))
}

const updateRepoQuery = `
UPDATE repo
SET
	name                  = %s,
	uri                   = NULLIF(%s, ''),
	description           = %s,
	external_service_type = %s,
	external_service_id   = %s,
	external_id           = %s,
	archived              = %s,
	fork                  = %s,
	stars                 = %s,
	private               = %s,
	metadata              = %s,
	updated_at            = now(),
	deleted_at            = NULL
WHERE id = %s
RETURNING updated_at
`

func (s *store) EnqueueSingleSyncJob(ctx context.Context, extSvcID int64) (err error) {
	q := sqlf.Sprintf(`
INSERT INTO external_service_sync_jobs (external_service_id)
SELECT %s
WHERE NOT EXISTS (
	SELECT
	FROM external_services es
	LEFT JOIN external_service_sync_jobs j ON es.id = j.external_service_id
	WHERE es.id = %s
	AND (
		j.state IN ('queued', 'processing')
		OR es.cloud_default
	)
)
`, extSvcID, extSvcID)
	return s.Exec(ctx, q)
}

func (s *store) EnqueueSyncJobs(ctx context.Context, isCloud bool) (err error) {
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

func (s *store) ListSyncJobs(ctx context.Context) ([]SyncJob, error) {
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
		job, err := scanJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, *job)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return jobs, nil
}

func scanJob(sc dbutil.Scanner) (*SyncJob, error) {
	// required field for the sync worker, but
	// the value is thrown out here
	var executionLogs *[]any

	var job SyncJob
	return &job, sc.Scan(
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
	)
}

func metadataColumn(metadata any) (msg json.RawMessage, err error) {
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
