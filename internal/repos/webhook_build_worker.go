package repos

import (
	"context"
	"database/sql"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	trace "github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	workerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

type WebhookBuildOptions struct {
	NumHandlers            int
	WorkerInterval         time.Duration
	PrometheusRegisterer   prometheus.Registerer
	CleanupOldJobs         bool
	CleanupOldJobsInterval time.Duration
}

func NewWebhookBuildWorker(
	ctx context.Context,
	dbHandle basestore.TransactableHandle,
	handler workerutil.Handler,
	opts WebhookBuildOptions,
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

	webhookBuildJobColumns := []*sqlf.Query{
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
		Scan:              scanWebhookBuildJob,
		OrderByExpression: sqlf.Sprintf("webhook_build_jobs.queued_at"),
		ColumnExpressions: webhookBuildJobColumns,
		StalledMaxAge:     30 * time.Second,
		MaxNumResets:      5,
		MaxNumRetries:     0,
	})

	worker := dbworker.NewWorker(ctx, store, handler, workerutil.WorkerOptions{
		Name:              "webhook_build_worker",
		NumHandlers:       opts.NumHandlers,
		Interval:          opts.WorkerInterval,
		HeartbeatInterval: 15 * time.Second,
		Metrics:           newWebhookBuildWorkerMetrics(opts.PrometheusRegisterer),
	})

	resetter := dbworker.NewResetter(store, dbworker.ResetterOptions{
		Name:     "webhook_build_resetter",
		Interval: 5 * time.Minute,
		Metrics:  newWebhookBuildResetterMetrics(opts.PrometheusRegisterer),
	})

	if opts.CleanupOldJobs {
		go runJobCleaner(ctx, dbHandle, opts.CleanupOldJobsInterval)
	}

	return worker, resetter
}

func newWebhookBuildWorkerMetrics(r prometheus.Registerer) workerutil.WorkerMetrics {
	var observationContext *observation.Context

	if r == nil {
		observationContext = &observation.TestContext
	} else {
		observationContext = &observation.Context{
			Logger:     log.Scoped("webhook_build_worker", ""),
			Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
			Registerer: r,
		}
	}

	return workerutil.NewMetrics(observationContext, "repo_updater_webhook_build_worker")
}

func newWebhookBuildResetterMetrics(r prometheus.Registerer) dbworker.ResetterMetrics {
	return dbworker.ResetterMetrics{
		RecordResets: promauto.With(r).NewCounter(prometheus.CounterOpts{
			Name: "src_webhook_build_queue_resets_total",
			Help: "Total number of webhook build jobs put back into queued state",
		}),
		RecordResetFailures: promauto.With(r).NewCounter(prometheus.CounterOpts{
			Name: "src_webhook_build_queue_max_resets_total",
			Help: "Total number of webhook build jobs that exceed the max number of resets",
		}),
		Errors: promauto.With(r).NewCounter(prometheus.CounterOpts{
			Name: "src_webhook_build_queue_reset_errors_total",
			Help: "Total number of errors when running the webhook build resetter",
		}),
	}
}

func runWebhookBuildCleaner(ctx context.Context, handle basestore.TransactableHandle, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()

	for {
		_, err := handle.ExecContext(ctx, `
-- source: internal/repos/webhook_builder.go:runWebhookBuildCleaner
DELETE FROM webhook_build_jobs
WHERE
  finished_at < now() - INTERVAL '1 day'
  AND state IN ('completed', 'errored')
`)
		if err != nil && err != context.Canceled {
			log15.Error("error while running job cleaner", "err", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

func scanWebhookBuildJob(rows *sql.Rows, err error) (workerutil.Record, bool, error) {
	if err != nil {
		return nil, false, err
	}

	jobs, err := scanWebhookBuildJobs(rows)
	if err != nil || len(jobs) == 0 {
		return nil, false, err
	}

	return &jobs[0], true, nil
}

type WebhookBuildJob struct {
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
	ExtSvcKind     string
	QueuedAt       sql.NullTime
}

func (cw *WebhookBuildJob) RecordID() int {
	return cw.ID
}
