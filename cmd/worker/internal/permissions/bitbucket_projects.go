package permissions

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
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
	db := database.NewDB(wdb)

	bbProjectMetrics := newMetricsForBitbucketProjectPermissionsQueries(logger)

	return []goroutine.BackgroundRoutine{
		newBitbucketProjectPermissionsWorker(db, bbProjectMetrics),
		newBitbucketProjectPermissionsResetter(db, bbProjectMetrics),
	}, nil
}

// bitbucketProjectPermissionsHandler handles the execution of a single explicit_permissions_bitbucket_projects_jobs record.
type bitbucketProjectPermissionsHandler struct{}

// Handle implements the workerutil.Handler interface.
func (h *bitbucketProjectPermissionsHandler) Handle(ctx context.Context, logger log.Logger, record workerutil.Record) (err error) {
	defer func() {
		if err != nil {
			logger.Error("bitbucketProjectPermissionsHandler.Handle", log.Error(err))
		}
	}()

	// TODO: handle the job
	return nil
}

// newBitbucketProjectPermissionsWorker creates a worker that reads the explicit_permissions_bitbucket_projects_jobs table and
// executes the jobs.
// TODO(asdine): Fine tune the retry strategy and make some parameters configurable.
func newBitbucketProjectPermissionsWorker(db database.DB, metrics bitbucketProjectPermissionsMetrics) *workerutil.Worker {
	options := workerutil.WorkerOptions{
		Name:              "explicit_permissions_bitbucket_projects_jobs_worker",
		NumHandlers:       3,
		Interval:          1 * time.Second,
		HeartbeatInterval: 15 * time.Second,
		Metrics:           metrics.workerMetrics,
	}

	return dbworker.NewWorker(context.Background(), createBitbucketProjectPermissionsStore(db), &bitbucketProjectPermissionsHandler{}, options)
}

// newBitbucketProjectPermissionsResetter implements resetter for the explicit_permissions_bitbucket_projects_jobs table.
// See resetter documentation for more details. https://docs.sourcegraph.com/dev/background-information/workers#dequeueing-and-resetting-jobs
func newBitbucketProjectPermissionsResetter(db database.DB, metrics bitbucketProjectPermissionsMetrics) *dbworker.Resetter {
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
		TableName: "explicit_permissions_bitbucket_projects_jobs j",
		ColumnExpressions: []*sqlf.Query{
			sqlf.Sprintf("j.id"),
			sqlf.Sprintf("j.state"),
			sqlf.Sprintf("j.failure_message"),
			sqlf.Sprintf("j.queued_at"),
			sqlf.Sprintf("j.started_at"),
			sqlf.Sprintf("j.finished_at"),
			sqlf.Sprintf("j.process_after"),
			sqlf.Sprintf("j.num_resets"),
			sqlf.Sprintf("j.num_failures"),
			sqlf.Sprintf("j.last_heartbeat_at"),
			sqlf.Sprintf("j.execution_logs"),
			sqlf.Sprintf("j.worker_hostname"),
			sqlf.Sprintf("j.project_key"),
			sqlf.Sprintf("j.external_service_id"),
			sqlf.Sprintf("j.permissions"),
			sqlf.Sprintf("j.unrestricted"),
		},
		Scan: func(rows *sql.Rows, err error) (workerutil.Record, bool, error) {
			j, ok, err := database.ScanFirstBitbucketProjectPermissionsJob(rows, err)
			return j, ok, err
		},
		StalledMaxAge:     60 * time.Second,
		RetryAfter:        10 * time.Second,
		MaxNumRetries:     5,
		OrderByExpression: sqlf.Sprintf("j.id"),
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
