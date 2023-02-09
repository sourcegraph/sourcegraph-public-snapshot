package expertnetwork

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

type ExpertNetworkJob struct {
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
	ExecutionLogs   []executor.ExecutionLogEntry
	WorkerHostname  string
	Cancel          bool

	RepositoryID int
}

var exampleJobColumns = []*sqlf.Query{
	sqlf.Sprintf("expert_network_jobs.id"),
	sqlf.Sprintf("expert_network_jobs.state"),
	sqlf.Sprintf("expert_network_jobs.failure_message"),
	sqlf.Sprintf("expert_network_jobs.queued_at"),
	sqlf.Sprintf("expert_network_jobs.started_at"),
	sqlf.Sprintf("expert_network_jobs.finished_at"),
	sqlf.Sprintf("expert_network_jobs.process_after"),
	sqlf.Sprintf("expert_network_jobs.num_resets"),
	sqlf.Sprintf("expert_network_jobs.num_failures"),
	sqlf.Sprintf("expert_network_jobs.last_heartbeat_at"),
	sqlf.Sprintf("expert_network_jobs.execution_logs"),
	sqlf.Sprintf("expert_network_jobs.worker_hostname"),
	sqlf.Sprintf("expert_network_jobs.cancel"),
	sqlf.Sprintf("expert_network_jobs.repository_id"),
}

func (j *ExpertNetworkJob) RecordID() int {
	return j.ID
}

func scanExampleJob(s dbutil.Scanner) (*ExpertNetworkJob, error) {
	var job ExpertNetworkJob
	var executionLogs []executor.ExecutionLogEntry

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
	); err != nil {
		return nil, err
	}

	for _, entry := range executionLogs {
		job.ExecutionLogs = append(job.ExecutionLogs, executor.ExecutionLogEntry(entry))
	}

	return &job, nil
}

func makeStore(observationCtx *observation.Context, dbHandle basestore.TransactableHandle) dbworkerstore.Store[*ExpertNetworkJob] {
	return dbworkerstore.New(observationCtx, dbHandle, dbworkerstore.Options[*ExpertNetworkJob]{
		Name:              "example_job_worker_store",
		TableName:         "expert_network_jobs",
		ColumnExpressions: exampleJobColumns,
		Scan:              dbworkerstore.BuildWorkerScan(scanExampleJob),
		OrderByExpression: sqlf.Sprintf("expert_network_jobs.repository_id, expert_network_jobs.id"),
		MaxNumResets:      5,
		StalledMaxAge:     time.Hour * 2,
	})
}

func makeWorker(
	ctx context.Context,
	observationCtx *observation.Context,
	workerStore dbworkerstore.Store[*ExpertNetworkJob],
	db database.DB,
	gitserverClient gitserver.Client,
) *workerutil.Worker[*ExpertNetworkJob] {
	handler := &handler{
		db:              db,
		gitserverClient: gitserverClient,
	}

	return dbworker.NewWorker[*ExpertNetworkJob](ctx, workerStore, handler, workerutil.WorkerOptions{
		Name:              "example_job_worker",
		Interval:          time.Second, // Poll for a job once per second
		NumHandlers:       1,           // Process only one job at a time (per instance)
		HeartbeatInterval: 10 * time.Second,
		Metrics:           workerutil.NewMetrics(observationCtx, "example_job_worker"),
	})
}

func makeResetter(observationCtx *observation.Context, workerStore dbworkerstore.Store[*ExpertNetworkJob]) *dbworker.Resetter[*ExpertNetworkJob] {
	return dbworker.NewResetter[*ExpertNetworkJob](observationCtx.Logger, workerStore, dbworker.ResetterOptions{
		Name:     "example_job_worker_resetter",
		Interval: time.Second * 30, // Check for orphaned jobs every 30 seconds
		Metrics:  dbworker.NewResetterMetrics(observationCtx, "example_job_worker"),
	})
}

type expertNetworkIndexerJob struct{}

func NewExpertNetworkIndexerJob() job.Job {
	return &expertNetworkIndexerJob{}
}

func (j *expertNetworkIndexerJob) Description() string {
	return ""
}

func (j *expertNetworkIndexerJob) Config() []env.Config {
	return []env.Config{}
}

func (j *expertNetworkIndexerJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}
	workCtx := actor.WithInternalActor(context.Background())
	gitserverClient := gitserver.NewClient()

	store := makeStore(observationCtx, db.Handle())
	return []goroutine.BackgroundRoutine{makeWorker(workCtx, observationCtx, store, db, gitserverClient)}, nil
}
