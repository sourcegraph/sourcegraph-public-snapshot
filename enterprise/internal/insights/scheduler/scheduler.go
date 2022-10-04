package scheduler

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"

	"github.com/sourcegraph/sourcegraph/internal/observation"

	"github.com/lib/pq"
	logger "github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/database"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

type BaseJob struct {
	ID              int
	State           string
	FailureMessage  *string
	QueuedAt        time.Time
	StartedAt       *time.Time
	FinishedAt      *time.Time
	ProcessAfter    *time.Time
	NumResets       int
	NumFailures     int
	LastHeartbeatAt time.Time
	ExecutionLogs   []workerutil.ExecutionLogEntry
	WorkerHostname  string
	Cancel          bool

	RepositoryID   int
	RepositoryName string
}

func (b *BaseJob) RecordID() int {
	return b.ID
}

var baseJobColumns = []*sqlf.Query{
	sqlf.Sprintf("insights_jobs.id"),
	sqlf.Sprintf("insights_jobs.state"),
	sqlf.Sprintf("insights_jobs.failure_message"),
	sqlf.Sprintf("insights_jobs.queued_at"),
	sqlf.Sprintf("insights_jobs.started_at"),
	sqlf.Sprintf("insights_jobs.finished_at"),
	sqlf.Sprintf("insights_jobs.process_after"),
	sqlf.Sprintf("insights_jobs.num_resets"),
	sqlf.Sprintf("insights_jobs.num_failures"),
	sqlf.Sprintf("insights_jobs.last_heartbeat_at"),
	sqlf.Sprintf("insights_jobs.execution_logs"),
	sqlf.Sprintf("insights_jobs.worker_hostname"),
	sqlf.Sprintf("insights_jobs.cancel"),
	sqlf.Sprintf("insights_jobs.repository_id"),
	sqlf.Sprintf("insights_jobs.repository_name"),
}

func scanBaseJob(s dbutil.Scanner) (*BaseJob, error) {
	var job BaseJob
	var executionLogs []dbworkerstore.ExecutionLogEntry

	if err := s.Scan(
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
		&job.Cancel,
		&job.RepositoryID,
		&job.RepositoryName,
	); err != nil {
		return nil, err
	}

	for _, entry := range executionLogs {
		job.ExecutionLogs = append(job.ExecutionLogs, workerutil.ExecutionLogEntry(entry))
	}

	return &job, nil
}

func NewScheduler(ctx context.Context, db database.DB, obsContext *observation.Context) {

	workerStore := makeStore(db, obsContext)
	worker := makeWorker(ctx, workerStore, obsContext)
	resetter := makeResetter(workerStore, obsContext)
}

func makeStore(db database.DB, obsContext *observation.Context) dbworkerstore.Store {
	return dbworkerstore.NewWithMetrics(db.Handle(), dbworkerstore.Options{
		Name:              "example_job_worker_store",
		TableName:         "example_jobs",
		ViewName:          "example_jobs_with_repository_name example_jobs",
		ColumnExpressions: baseJobColumns,
		Scan:              dbworkerstore.BuildWorkerScan(scanBaseJob),
		OrderByExpression: sqlf.Sprintf("example_jobs.repository_id, example_jobs.id"),
		MaxNumResets:      100,
		StalledMaxAge:     time.Second * 30,
	}, obsContext)
}

const jobName = "insights_job_scheduler"

func makeWorker(ctx context.Context, store dbworkerstore.Store, obsContext *observation.Context) *workerutil.Worker {
	task := &handler{}
	return dbworker.NewWorker(ctx, store, task, workerutil.WorkerOptions{
		Name:        "example_job_worker",
		NumHandlers: 1,           // Process only one job at a time (per instance)
		Interval:    time.Second, // Poll for a job once per second
		Metrics:     workerutil.NewMetrics(obsContext, jobName),
	})
}

func makeResetter(store dbworkerstore.Store, obsContext *observation.Context) *dbworker.Resetter {
	return dbworker.NewResetter(logger.Scoped("", ""), store, dbworker.ResetterOptions{
		Name:     "",
		Interval: 0,
		Metrics:  *dbworker.NewMetrics(obsContext, jobName),
	})
}

type handler struct {
	store dbworkerstore.Store
}

var _ workerutil.Handler = &handler{}

func (h *handler) Handle(ctx context.Context, logger logger.Logger, record workerutil.Record) error {
	// TODO implement me
	panic("implement me")
}
