package repos

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	workerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

type WhBuildOptions struct {
	NumHandlers            int
	WorkerInterval         time.Duration
	PrometheusRegisterer   prometheus.Registerer
	CleanupOldJobs         bool
	CleanupOldJobsInterval time.Duration
}

func NewWhBuildWorker(
	ctx context.Context,
	dbHandle basestore.TransactableHandle,
	handler workerutil.Handler,
	opts WhBuildOptions,
) (*workerutil.Worker, *dbworker.Resetter) {
	if opts.NumHandlers == 0 {
		opts.NumHandlers = 3
	}
	if opts.WorkerInterval == 0 {
		opts.WorkerInterval = 10 * time.Second
	}
	if opts.CleanupOldJobsInterval == 0 {
		opts.CleanupOldJobsInterval = time.Hour
	}

	whBuildJobColumns := []*sqlf.Query{
		sqlf.Sprintf("id"),
		sqlf.Sprintf("state"),
		sqlf.Sprintf("failure_message"),
		sqlf.Sprintf("started_at"),
		sqlf.Sprintf("finished_at"),
		sqlf.Sprintf("process_after"),
		sqlf.Sprintf("num_resets"),
		sqlf.Sprintf("num_failures"),
		sqlf.Sprintf("execution_logs"),
		sqlf.Sprintf("repo_id"),
		sqlf.Sprintf("repo_name"),
		sqlf.Sprintf("extsvc_kind"),
		sqlf.Sprintf("queued_at"),
	}

	store := workerstore.New(dbHandle, workerstore.Options{
		Name:      "webhook_build_worker_store",
		TableName: "webhook_build_jobs",
		// ViewName:          "webhook_build_jobs_with_next_in_queue",
		// Scan:              scanWhBuildJob, to be implemented in next PR
		OrderByExpression: sqlf.Sprintf("webhook_build_jobs.queued_at"),
		ColumnExpressions: whBuildJobColumns,
		StalledMaxAge:     30 * time.Second,
		MaxNumResets:      5,
		MaxNumRetries:     0,
	})

	worker := dbworker.NewWorker(ctx, store, handler, workerutil.WorkerOptions{
		Name:              "webhook_build_worker",
		NumHandlers:       opts.NumHandlers,
		Interval:          opts.WorkerInterval,
		HeartbeatInterval: 15 * time.Second,
		Metrics:           newWorkerMetrics(opts.PrometheusRegisterer), // move to central pacckage
	})

	resetter := dbworker.NewResetter(store, dbworker.ResetterOptions{
		Name:     "webhook_build_resetter",
		Interval: 5 * time.Minute,
		Metrics:  newResetterMetrics(opts.PrometheusRegisterer), // move to central package
	})

	if opts.CleanupOldJobs {
		go runJobCleaner(ctx, dbHandle, opts.CleanupOldJobsInterval) // move to central package
	}

	return worker, resetter
}

type WhBuildJob struct {
	ID             int
	State          string
	FailureMessage sql.NullString
	StartedAt      sql.NullTime
	FinishedAt     sql.NullTime
	ProcessAfter   sql.NullTime
	NumResets      int
	NumFailures    int
	RepoID         int64
	RepoName       string
	ExtsvcKind     string
	QueuedAt       sql.NullTime
}

func (cw *WhBuildJob) RecordID() int {
	return cw.ID
}
