package permissions

import (
	"context"
	"database/sql"
	"sort"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/errcode"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// bitbucketProjectPermissionsJob implements the job.Job interface. It is used by the worker service
// to spawn a new background worker.
type bitbucketProjectPermissionsJob struct{}

// NewBitbucketProjectPermissionsJob creates a new job for applying explicit permissions
// to all the repositories of a Bitbucket Project.
func NewBitbucketProjectPermissionsJob() job.Job {
	return &bitbucketProjectPermissionsJob{}
}

func (j *bitbucketProjectPermissionsJob) Description() string {
	return "Applies explicit permissions to all repositories of a Bitbucket Project."
}

// TODO(asdine): Load environment variables from here if needed.
func (j *bitbucketProjectPermissionsJob) Config() []env.Config {
	return []env.Config{}
}

// Routines is called by the worker service to start the worker.
// It returns a list of goroutines that the worker service should start and manage.
func (j *bitbucketProjectPermissionsJob) Routines(ctx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	wdb, err := workerdb.Init()
	if err != nil {
		return nil, err
	}
	db := edb.NewEnterpriseDB(database.NewDB(wdb))

	bbProjectMetrics := newMetricsForBitbucketProjectPermissionsQueries(logger)

	return []goroutine.BackgroundRoutine{
		newBitbucketProjectPermissionsWorker(db, bbProjectMetrics),
		newBitbucketProjectPermissionsResetter(db, bbProjectMetrics),
	}, nil
}

// bitbucketProjectPermissionsHandler handles the execution of a single explicit_permissions_bitbucket_projects_jobs record.
type bitbucketProjectPermissionsHandler struct {
	db     edb.EnterpriseDB
	client *bitbucketserver.Client
}

// Handle implements the workerutil.Handler interface.
func (h *bitbucketProjectPermissionsHandler) Handle(ctx context.Context, logger log.Logger, record workerutil.Record) (err error) {
	defer func() {
		if err != nil {
			logger.Error("bitbucketProjectPermissionsHandler.Handle", log.Error(err))
		}
	}()

	job := record.(*types.BitbucketProjectPermissionJob)

	// get the external service
	svc, err := h.db.ExternalServices().GetByID(ctx, job.ExternalServiceID)
	if err != nil {
		return errors.Wrapf(err, "failed to get external service %d", job.ExternalServiceID)
	}

	if svc.Kind != extsvc.KindBitbucketServer {
		return errcode.MakeNonRetryable(errors.Newf("expected Bitbucket Server external service, got: %s", svc.Kind))
	}

	// get repos from the Bitbucket project
	client, err := h.getBitbucketClient(ctx, svc)
	if err != nil {
		return errors.Wrapf(err, "failed to build Bitbucket client for external service %d", svc.ID)
	}
	repos, err := client.ProjectRepos(ctx, job.ProjectKey)
	if err != nil {
		return errors.Wrapf(err, "failed to list repositories of Bitbucket Project %q", job.ProjectKey)
	}

	if job.Unrestricted {
		return h.setReposUnrestricted(ctx, repos, job.ProjectKey)
	}

	repoIDs := make([]api.RepoID, len(repos))
	for i, repo := range repos {
		repoIDs[i] = api.RepoID(repo.ID)
	}

	err = h.setPermissionsForUsers(ctx, job.Permissions, repoIDs)
	if err != nil {
		return errors.Wrapf(err, "failed to set permissions for Bitbucket Project %q", job.ProjectKey)
	}

	return nil
}

// getBitbucketClient creates a Bitbucket client for the given external service.
func (h *bitbucketProjectPermissionsHandler) getBitbucketClient(ctx context.Context, svc *types.ExternalService) (*bitbucketserver.Client, error) {
	// for testing purpose
	if h.client != nil {
		return h.client, nil
	}

	var c schema.BitbucketServerConnection
	if err := jsonc.Unmarshal(svc.Config, &c); err != nil {
		return nil, errors.Errorf("external service id=%d config error: %s", svc.ID, err)
	}

	var opts []httpcli.Opt
	if c.Certificate != "" {
		opts = append(opts, httpcli.NewCertPoolOpt(c.Certificate))
	}

	cli, err := httpcli.ExternalClientFactory.Doer(opts...)
	if err != nil {
		return nil, err
	}

	return bitbucketserver.NewClient(svc.URN(), &c, cli)
}

func (h *bitbucketProjectPermissionsHandler) setReposUnrestricted(ctx context.Context, repos []*bitbucketserver.Repo, projectKey string) error {
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].ID < repos[j].ID
	})
	repoIDs := make([]int32, len(repos))
	for i, repo := range repos {
		repoIDs[i] = int32(repo.ID)
	}

	err := h.db.Perms().SetRepoPermissionsUnrestricted(ctx, repoIDs, true)
	if err != nil {
		return errors.Wrapf(err, "failed to set permissions to unrestricted for Bitbucket Project %q", projectKey)
	}

	return nil
}

// setPermissionsForUsers applies user permissions to a list of repos.
// It updates the repo_permissions, user_permissions, repo_pending_permissions and user_pending_permissions table.
// Each repo is processed atomically. In case of error, the task fails but doesn't rollback the committed changes
// done on previous repos. This is fine because when the task is retried, previous repos won't incur any
// additional writes.
func (h *bitbucketProjectPermissionsHandler) setPermissionsForUsers(ctx context.Context, perms []types.UserPermission, repoIDs []api.RepoID) error {
	sort.Slice(perms, func(i, j int) bool {
		return perms[i].BindID < perms[j].BindID
	})
	sort.Slice(repoIDs, func(i, j int) bool {
		return repoIDs[i] < repoIDs[j]
	})

	bindIDs := make([]string, 0, len(perms))
	for _, up := range perms {
		bindIDs = append(bindIDs, up.BindID)
	}

	// bind the bindIDs to actual user IDs
	mapping, err := h.db.Perms().MapUsers(ctx, bindIDs, globals.PermissionsUserMapping())
	if err != nil {
		return errors.Wrap(err, "failed to map bind IDs to user IDs")
	}

	if len(mapping) == 0 {
		return errors.Errorf("no users found for bind IDs: %v", bindIDs)
	}

	userIDs := make(map[int32]struct{}, len(mapping))
	for _, id := range mapping {
		userIDs[id] = struct{}{}
	}

	// determine which users don't exist yet
	pendingBindIDs := make([]string, 0, len(bindIDs))
	for _, bindID := range bindIDs {
		if _, ok := mapping[bindID]; !ok {
			pendingBindIDs = append(pendingBindIDs, bindID)
		}
	}

	// apply the permissions for each repo
	for _, repoID := range repoIDs {
		err = h.setRepoPermissions(ctx, repoID, perms, userIDs, pendingBindIDs)
		if err != nil {
			return errors.Wrapf(err, "failed to set permissions for repo %d", repoID)
		}
	}

	return nil
}

func (h *bitbucketProjectPermissionsHandler) setRepoPermissions(ctx context.Context, repoID api.RepoID, _ []types.UserPermission, userIDs map[int32]struct{}, pendingBindIDs []string) (err error) {
	// Make sure the repo ID is valid.
	if _, err := h.db.Repos().Get(ctx, repoID); err != nil {
		return errors.Wrapf(err, "failed to query repo %d", repoID)
	}

	p := authz.RepoPermissions{
		RepoID:  int32(repoID),
		Perm:    authz.Read, // Note: We currently only support read for repository permissions.
		UserIDs: userIDs,
	}

	txs, err := h.db.Perms().Transact(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to start transaction")
	}
	defer func() { err = txs.Done(err) }()

	accounts := &extsvc.Accounts{
		ServiceType: authz.SourcegraphServiceType,
		ServiceID:   authz.SourcegraphServiceID,
		AccountIDs:  pendingBindIDs,
	}

	// make sure the repo is not unrestricted
	err = txs.SetRepoPermissionsUnrestricted(ctx, []int32{int32(repoID)}, false)
	if err != nil {
		return errors.Wrapf(err, "failed to set repo %d to restricted", repoID)
	}

	// set repo permissions (and user permissions)
	err = txs.SetRepoPermissions(ctx, &p)
	if err != nil {
		return errors.Wrapf(err, "failed to set repo permissions for repo %d", repoID)
	}

	// set pending permissions
	err = txs.SetRepoPendingPermissions(ctx, accounts, &p)
	if err != nil {
		return errors.Wrapf(err, "failed to set pending permissions for repo %d", repoID)
	}

	return nil
}

// newBitbucketProjectPermissionsWorker creates a worker that reads the explicit_permissions_bitbucket_projects_jobs table and
// executes the jobs.
// TODO(asdine): Fine tune the retry strategy and make some parameters configurable.
func newBitbucketProjectPermissionsWorker(db edb.EnterpriseDB, metrics bitbucketProjectPermissionsMetrics) *workerutil.Worker {
	options := workerutil.WorkerOptions{
		Name:              "explicit_permissions_bitbucket_projects_jobs_worker",
		NumHandlers:       3,
		Interval:          1 * time.Second,
		HeartbeatInterval: 15 * time.Second,
		Metrics:           metrics.workerMetrics,
	}

	return dbworker.NewWorker(context.Background(), createBitbucketProjectPermissionsStore(db), &bitbucketProjectPermissionsHandler{db: db}, options)
}

// newBitbucketProjectPermissionsResetter implements resetter for the explicit_permissions_bitbucket_projects_jobs table.
// See resetter documentation for more details. https://docs.sourcegraph.com/dev/background-information/workers#dequeueing-and-resetting-jobs
func newBitbucketProjectPermissionsResetter(db edb.EnterpriseDB, metrics bitbucketProjectPermissionsMetrics) *dbworker.Resetter {
	workerStore := createBitbucketProjectPermissionsStore(db)

	options := dbworker.ResetterOptions{
		Name:     "explicit_permissions_bitbucket_projects_jobs_worker_resetter",
		Interval: 1 * time.Minute,
		Metrics: dbworker.ResetterMetrics{
			Errors:              metrics.errors,
			RecordResetFailures: metrics.resetFailures,
			RecordResets:        metrics.resets,
		},
	}
	return dbworker.NewResetter(workerStore, options)
}

// createBitbucketProjectPermissionsStore creates a store that reads and writes to the explicit_permissions_bitbucket_projects_jobs table.
// It is used by the worker and resetter.
// TODO(asdine): Fine tune the retry strategy and make some parameters configurable.
func createBitbucketProjectPermissionsStore(s basestore.ShareableStore) dbworkerstore.Store {
	return dbworkerstore.New(s.Handle(), dbworkerstore.Options{
		Name:      "explicit_permissions_bitbucket_projects_jobs_store",
		TableName: "explicit_permissions_bitbucket_projects_jobs",
		ColumnExpressions: []*sqlf.Query{
			sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.id"),
			sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.state"),
			sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.failure_message"),
			sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.queued_at"),
			sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.started_at"),
			sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.finished_at"),
			sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.process_after"),
			sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.num_resets"),
			sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.num_failures"),
			sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.last_heartbeat_at"),
			sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.execution_logs"),
			sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.worker_hostname"),
			sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.project_key"),
			sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.external_service_id"),
			sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.permissions"),
			sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.unrestricted"),
		},
		Scan: func(rows *sql.Rows, err error) (workerutil.Record, bool, error) {
			j, ok, err := database.ScanFirstBitbucketProjectPermissionsJob(rows, err)
			return j, ok, err
		},
		StalledMaxAge:     60 * time.Second,
		RetryAfter:        10 * time.Second,
		MaxNumRetries:     5,
		OrderByExpression: sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.id"),
	})
}

// These are the metrics that are used by the worker and resetter.
// They are required by the workerutil package for automatic metrics collection.
type bitbucketProjectPermissionsMetrics struct {
	workerMetrics workerutil.WorkerMetrics
	resets        prometheus.Counter
	resetFailures prometheus.Counter
	errors        prometheus.Counter
}

func newMetricsForBitbucketProjectPermissionsQueries(logger log.Logger) bitbucketProjectPermissionsMetrics {
	observationContext := &observation.Context{
		Logger:     logger.Scoped("routines", "bitbucket projects explicit permissions job routines"),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	resetFailures := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_explicit_permissions_bitbucket_project_query_reset_failures_total",
		Help: "The number of reset failures.",
	})
	observationContext.Registerer.MustRegister(resetFailures)

	resets := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_explicit_permissions_bitbucket_project_query_resets_total",
		Help: "The number of records reset.",
	})
	observationContext.Registerer.MustRegister(resets)

	errors := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_explicit_permissions_bitbucket_project_query_errors_total",
		Help: "The number of errors that occur during job.",
	})
	observationContext.Registerer.MustRegister(errors)

	return bitbucketProjectPermissionsMetrics{
		workerMetrics: workerutil.NewMetrics(observationContext, "explicit_permissions_bitbucket_project_queries"),
		resets:        resets,
		resetFailures: resetFailures,
		errors:        errors,
	}
}
