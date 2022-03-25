package background

import (
	"context"
	"database/sql"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/workerdb"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// localCloneJob implements the job.Job interface. It is used by the worker service
// to spawn a new background worker.

type localCloneJob struct{}

// NewLocalCloneJob creates a new job for cloning a repository from source to the destination.
// This job must be run on all environments (i.e. OSS, Enterprise, Cloud, etc.).
func NewLocalCloneJob() job.Job {
	return &localCloneJob{}
}

// TODO(asdine): Load environment variables from here.
func (j *localCloneJob) Config() []env.Config {
	return []env.Config{}
}

// Routines is called by the worker service to start the worker.
// It returns a list of goroutines that the worker service should start and manage.
func (j *localCloneJob) Routines(ctx context.Context) ([]goroutine.BackgroundRoutine, error) {
	wdb, err := workerdb.Init()
	if err != nil {
		return nil, err
	}
	db := database.NewDB(wdb)

	localCloneStore := db.GitserverLocalClone()
	localCloneMetrics := newMetricsForLocalCloneQueries()

	return []goroutine.BackgroundRoutine{
		newLocalCloneWorker(db, localCloneMetrics),
		newLocalCloneResetter(localCloneStore, localCloneMetrics),
	}, nil
}

// localCloneHandler handles the execution of a single gitserver_localclone_jobs record.
type localCloneHandler struct {
	client gitserver.Client
}

// Handle takes the given job and clones the repository from source to the destination.
func (h *localCloneHandler) Handle(ctx context.Context, record workerutil.Record) (err error) {
	defer func() {
		if err != nil {
			log15.Error("localCloneHandler.Handle", "error", err)
		}
	}()

	job, ok := record.(*types.GitserverLocalCloneJob)
	if !ok {
		return errors.Errorf("unexpected record type %T", record)
	}

	resp, err := h.client.RequestRepoMigrate(ctx, api.RepoName(job.RepoName), job.SourceHostname, job.DestHostname)
	if err != nil {
		return errors.Wrap(err, "error requesting repo migrate")
	} else if resp != nil && resp.Error != "" {
		return errors.New("error migrating repo: " + resp.Error)
	}

	if job.DeleteSource {
		if err := h.client.RemoveFrom(ctx, api.RepoName(job.RepoName), job.SourceHostname); err != nil {
			// It is fine to return an error here because the job will be retried and, instead of being cloned again, the repo will
			// get updated, which is much cheaper.
			return errors.Wrap(err, "error deleting repo from source")
		}
	}

	return nil
}

// newLocalCloneWorker creates a worker that reads the gitserver_localclone_jobs table and
// executes the jobs.
// TODO(asdine): Fine tune the retry strategy and make some parameters configurable.
func newLocalCloneWorker(db database.DB, metrics localCloneMetrics) *workerutil.Worker {
	options := workerutil.WorkerOptions{
		Name:              "gitserver_localclone_jobs_worker",
		NumHandlers:       3,
		Interval:          1 * time.Second,
		HeartbeatInterval: 15 * time.Second,
		Metrics:           metrics.workerMetrics,
	}

	return dbworker.NewWorker(context.Background(), createLocalCloneStore(db), &localCloneHandler{client: gitserver.NewClient(db)}, options)
}

// newLocalCloneResetter implements resetter for the gitserver_localclone_jobs table.
// See resetter documentation for more details. https://docs.sourcegraph.com/dev/background-information/workers#dequeueing-and-resetting-jobs
func newLocalCloneResetter(s database.GitserverLocalCloneStore, metrics localCloneMetrics) *dbworker.Resetter {
	workerStore := createLocalCloneStore(s)

	options := dbworker.ResetterOptions{
		Name:     "gitserver_localclone_jobs_worker_resetter",
		Interval: 1 * time.Minute,
		Metrics: dbworker.ResetterMetrics{
			Errors:              metrics.errors,
			RecordResetFailures: metrics.resetFailures,
			RecordResets:        metrics.resets,
		},
	}
	return dbworker.NewResetter(workerStore, options)
}

// createLocalCloneStore creates a store that reads and writes to the gitserver_localclone_jobs table.
// It is used by the worker and resetter.
// TODO(asdine): Fine tune the retry strategy and make some parameters configurable.
func createLocalCloneStore(s basestore.ShareableStore) dbworkerstore.Store {
	return dbworkerstore.New(s.Handle(), dbworkerstore.Options{
		Name:      "gitserver_localclone_jobs_store",
		TableName: "gitserver_localclone_jobs j",
		ViewName:  "gitserver_localclone_jobs_with_repo_name j",
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
			sqlf.Sprintf("j.repo_id"),
			sqlf.Sprintf("j.repo_name"),
			sqlf.Sprintf("j.source_hostname"),
			sqlf.Sprintf("j.dest_hostname"),
			sqlf.Sprintf("j.delete_source"),
		},
		Scan:              scanFirstGitserverLocalCloneJob,
		StalledMaxAge:     60 * time.Second,
		RetryAfter:        10 * time.Second,
		MaxNumRetries:     5,
		OrderByExpression: sqlf.Sprintf("id"),
	})
}

// scanFirstLocalCloneJob scans a single job from the return value of `*Store.query`.
func scanFirstGitserverLocalCloneJob(rows *sql.Rows, queryErr error) (_ workerutil.Record, exists bool, err error) {
	if queryErr != nil {
		return nil, false, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	if rows.Next() {
		var job types.GitserverLocalCloneJob
		var executionLogs []dbworkerstore.ExecutionLogEntry

		if err := rows.Scan(
			&job.ID,
			&job.State,
			&job.FailureMessage,
			&job.QueuedAt,
			&job.StartedAt,
			&job.FinishedAt,
			&job.ProcessAfter,
			&job.NumResets,
			&job.NumFailures,
			&job.LastHeartbeatAt,
			pq.Array(&executionLogs),
			&job.WorkerHostname,
			&job.RepoID,
			&job.RepoName,
			&job.SourceHostname,
			&job.DestHostname,
			&job.DeleteSource,
		); err != nil {
			return nil, false, err
		}

		for _, entry := range executionLogs {
			job.ExecutionLogs = append(job.ExecutionLogs, workerutil.ExecutionLogEntry(entry))
		}

		return &job, true, nil
	}

	return nil, false, nil
}

// These are the metrics that are used by the worker and resetter.
// They are required by the workerutil package for automatic metrics collection.
type localCloneMetrics struct {
	workerMetrics workerutil.WorkerMetrics
	resets        prometheus.Counter
	resetFailures prometheus.Counter
	errors        prometheus.Counter
}

func newMetricsForLocalCloneQueries() localCloneMetrics {
	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	resetFailures := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_gitserver_localclone_query_reset_failures_total",
		Help: "The number of reset failures.",
	})
	observationContext.Registerer.MustRegister(resetFailures)

	resets := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_gitserver_localclone_query_resets_total",
		Help: "The number of records reset.",
	})
	observationContext.Registerer.MustRegister(resets)

	errors := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_gitserver_localclone_query_errors_total",
		Help: "The number of errors that occur during job.",
	})
	observationContext.Registerer.MustRegister(errors)

	return localCloneMetrics{
		workerMetrics: workerutil.NewMetrics(observationContext, "gitserver_localclone_queries"),
		resets:        resets,
		resetFailures: resetFailures,
		errors:        errors,
	}
}
