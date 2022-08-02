package webhookworker

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	workerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

func NewWorker(ctx context.Context, handler workerutil.Handler, workerStore workerstore.Store, metrics workerutil.WorkerMetrics) *workerutil.Worker {
	options := workerutil.WorkerOptions{
		Name:              "webhook_build_worker",
		NumHandlers:       3,
		Interval:          5 * time.Second,
		HeartbeatInterval: 15 * time.Second,
		Metrics:           metrics,
	}

	return dbworker.NewWorker(ctx, workerStore, handler, options)
}

func NewResetter(ctx context.Context, workerStore workerstore.Store, metrics dbworker.ResetterMetrics) *dbworker.Resetter {
	options := dbworker.ResetterOptions{
		Name:     "webhook_build_resetter",
		Interval: 1 * time.Minute,
		Metrics:  metrics,
	}

	return dbworker.NewResetter(workerStore, options)
}

func CreateWorkerStore(dbHandle basestore.TransactableHandle) workerstore.Store {
	return workerstore.New(dbHandle, workerstore.Options{
		Name:              "webhook_build_worker_store",
		TableName:         "webhook_build_jobs",
		Scan:              scanWebhookBuildJobs,
		OrderByExpression: sqlf.Sprintf("webhook_build_jobs.queued_at"),
		ColumnExpressions: jobColumns,
		StalledMaxAge:     30 * time.Second,
		MaxNumResets:      5,
		MaxNumRetries:     5,
	})
}

func EnqueueJob(ctx context.Context, workerBaseStore *basestore.Store, job *Job) (id int, err error) {
	tx, err := workerBaseStore.Transact(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { err = tx.Done(err) }()

	id, _, err = basestore.ScanFirstInt(tx.Query(
		ctx,
		sqlf.Sprintf(
			enqueueJobFmtStr,
			job.RepoID,
			job.RepoName,
			job.ExtSvcKind,
		),
	))
	if err != nil {
		return 0, err
	}
	job.ID = id
	return id, nil
}

const enqueueJobFmtStr = `
-- source: internal/repos/worker/worker.go:EnqueueJob
INSERT INTO webhook_build_jobs (
	repo_id,
	repo_name,
	extsvc_kind
) VALUES (%s, %s, %s)
RETURNING id
`

func scanWebhookBuildJobs(rows *sql.Rows, err error) (workerutil.Record, bool, error) {
	records, err := doScanWebhookBuildJobs(rows, err)
	if err != nil || len(records) == 0 {
		return &Job{}, false, err
	}

	return records[0], true, nil
}

func doScanWebhookBuildJobs(rows *sql.Rows, err error) ([]*Job, error) {
	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()
	var jobs []*Job

	for rows.Next() {
		var job Job
		if err := rows.Scan(
			// Webhook builder fields
			&job.RepoID,
			&job.RepoName,
			&job.ExtSvcKind,
			&job.QueuedAt,

			// Standard dbworker fields
			&job.ID,
			&job.State,
			&job.FailureMessage,
			&job.StartedAt,
			&job.FinishedAt,
			&job.ProcessAfter,
			&job.NumResets,
			&job.NumFailures,
			pq.Array(&job.ExecutionLogs),
		); err != nil {
			return nil, err
		}
		jobs = append(jobs, &job)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return jobs, nil
}
