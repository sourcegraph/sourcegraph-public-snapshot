package scheduler

import (
	"context"
	"fmt"
	"time"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"

	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"

	"github.com/sourcegraph/sourcegraph/internal/observation"

	"github.com/lib/pq"
	log "github.com/sourcegraph/log"

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
}

func (b *BaseJob) RecordID() int {
	return b.ID
}

var baseJobColumns = []*sqlf.Query{
	sqlf.Sprintf("insights_background_jobs.id"),
	sqlf.Sprintf("insights_background_jobs.state"),
	sqlf.Sprintf("insights_background_jobs.failure_message"),
	sqlf.Sprintf("insights_background_jobs.queued_at"),
	sqlf.Sprintf("insights_background_jobs.started_at"),
	sqlf.Sprintf("insights_background_jobs.finished_at"),
	sqlf.Sprintf("insights_background_jobs.process_after"),
	sqlf.Sprintf("insights_background_jobs.num_resets"),
	sqlf.Sprintf("insights_background_jobs.num_failures"),
	sqlf.Sprintf("insights_background_jobs.last_heartbeat_at"),
	sqlf.Sprintf("insights_background_jobs.execution_logs"),
	sqlf.Sprintf("insights_background_jobs.worker_hostname"),
	sqlf.Sprintf("insights_background_jobs.cancel"),
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
	); err != nil {
		return nil, err
	}

	for _, entry := range executionLogs {
		job.ExecutionLogs = append(job.ExecutionLogs, workerutil.ExecutionLogEntry(entry))
	}

	return &job, nil
}

type scheduler struct {
	workerStore dbworkerstore.Store
	worker      *workerutil.Worker
	resetter    *dbworker.Resetter
}

func NewScheduler(ctx context.Context, db edb.InsightsDB, obsContext *observation.Context) *scheduler {
	workerStore := makeStore(db.Handle(), obsContext)
	worker := makeWorker(ctx, workerStore, obsContext)
	resetter := makeResetter(workerStore, obsContext)

	return &scheduler{
		workerStore: workerStore,
		worker:      worker,
		resetter:    resetter,
	}
}

func (s *scheduler) Routines() []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		s.worker,
		s.resetter,
	}
}

func makeStore(db basestore.TransactableHandle, obsContext *observation.Context) dbworkerstore.Store {
	return dbworkerstore.NewWithMetrics(db, dbworkerstore.Options{
		Name:              "insights_background_job_worker_store",
		TableName:         "insights_background_jobs",
		ColumnExpressions: baseJobColumns,
		Scan:              dbworkerstore.BuildWorkerScan(scanBaseJob),
		OrderByExpression: sqlf.Sprintf("insights_background_jobs.id"),
		MaxNumResets:      100,
		StalledMaxAge:     time.Second * 30,
	}, obsContext)
}

const jobName = "insights_background_job_scheduler"

func makeWorker(ctx context.Context, store dbworkerstore.Store, obsContext *observation.Context) *workerutil.Worker {
	task := &handler{}
	name := fmt.Sprintf("%s_worker", jobName)
	return dbworker.NewWorker(ctx, store, task, workerutil.WorkerOptions{
		Name:        name,
		NumHandlers: 1,
		Interval:    5 * time.Second,
		Metrics:     workerutil.NewMetrics(obsContext, name),
	})
}

func makeResetter(store dbworkerstore.Store, obsContext *observation.Context) *dbworker.Resetter {
	name := fmt.Sprintf("%s_resetter", jobName)
	return dbworker.NewResetter(log.Scoped("", ""), store, dbworker.ResetterOptions{
		Name:     name,
		Interval: time.Second * 20,
		Metrics:  *dbworker.NewMetrics(obsContext, name),
	})
}

type handler struct {
	store dbworkerstore.Store
}

var _ workerutil.Handler = &handler{}

func (h *handler) Handle(ctx context.Context, logger log.Logger, record workerutil.Record) error {
	logger.Info("handler called", log.String("job", jobName), log.Int("recordId", record.RecordID()))
	return nil
}
