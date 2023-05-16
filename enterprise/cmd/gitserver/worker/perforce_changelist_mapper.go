package perforce

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegraph/log"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type ExecutionLogEntry struct {
	RepoID  int
	Message string
}

type PerforceChangelistMappingJob struct {
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

	ExecutionLogs  []ExecutionLogEntry
	WorkerHostname string
	Cancel         bool

	RepoID int
}

func (j *PerforceChangelistMappingJob) RecordID() int {
	return j.ID
}

var perforceChangelistMappingJobColumns = []*sqlf.Query{
	sqlf.Sprintf("perforce_changelist_mapping_jobs.id"),
	sqlf.Sprintf("perforce_changelist_mapping_jobs.state"),
	sqlf.Sprintf("perforce_changelist_mapping_jobs.failure_message"),
	sqlf.Sprintf("perforce_changelist_mapping_jobs.queued_at"),
	sqlf.Sprintf("perforce_changelist_mapping_jobs.started_at"),
	sqlf.Sprintf("perforce_changelist_mapping_jobs.finished_at"),
	sqlf.Sprintf("perforce_changelist_mapping_jobs.process_after"),
	sqlf.Sprintf("perforce_changelist_mapping_jobs.num_resets"),
	sqlf.Sprintf("perforce_changelist_mapping_jobs.num_failures"),
	sqlf.Sprintf("perforce_changelist_mapping_jobs.last_heartbeat_at"),
	sqlf.Sprintf("perforce_changelist_mapping_jobs.execution_logs"),
	sqlf.Sprintf("perforce_changelist_mapping_jobs.worker_hostname"),
	sqlf.Sprintf("perforce_changelist_mapping_jobs.cancel"),
	sqlf.Sprintf("perforce_changelist_mapping_jobs.repo_id"),
}

func scanPerforceChangelistMappingJob(s dbutil.Scanner) (*PerforceChangelistMappingJob, error) {
	var job PerforceChangelistMappingJob
	var executionLogs []ExecutionLogEntry

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
		&job.RepoID,
	); err != nil {
		return nil, err
	}

	for _, entry := range executionLogs {
		job.ExecutionLogs = append(job.ExecutionLogs, ExecutionLogEntry(entry))
	}

	return &job, nil
}

func makeStore(observationCtx *observation.Context, dbHandle basestore.TransactableHandle) dbworkerstore.Store[*PerforceChangelistMappingJob] {
	return dbworkerstore.New(observationCtx, dbHandle, dbworkerstore.Options[*PerforceChangelistMappingJob]{
		Name:              "perforce_changelist_mapping_job_worker_store",
		TableName:         "perforce_changelist_mapping_jobs",
		ColumnExpressions: perforceChangelistMappingJobColumns,
		Scan:              dbworkerstore.BuildWorkerScan(scanPerforceChangelistMappingJob),
		OrderByExpression: sqlf.Sprintf("perforce_changelist_mapping_jobs.repository_id, perforce_changelist_mapping_jobs.id"),
		MaxNumResets:      5,
		StalledMaxAge:     time.Second * 5,
	})
}

type PerforceChangelistMappingJobStore interface {
	basestore.ShareableStore
	DataForRepo(int) string
}

type handler struct {
	perforceChangelistMappingJobStore PerforceChangelistMappingJobStore
}

var _ workerutil.Handler[*PerforceChangelistMappingJob] = &handler{}

func (h *handler) Handle(ctx context.Context, logger log.Logger, record *PerforceChangelistMappingJob) error {
	data := h.perforceChangelistMappingJobStore.DataForRepo(record.RepoID)

	return h.process(data)
}

func (h *handler) process(data string) error {
	// Do the actual processing

	return nil
}

func makeWorker(
	ctx context.Context,
	observationCtx *observation.Context,
	workerStore dbworkerstore.Store[*PerforceChangelistMappingJob],
	jobStore PerforceChangelistMappingJobStore,
) *workerutil.Worker[*PerforceChangelistMappingJob] {
	handler := &handler{
		perforceChangelistMappingJobStore: jobStore,
	}

	return dbworker.NewWorker[*PerforceChangelistMappingJob](ctx, workerStore, handler, workerutil.WorkerOptions{
		Name:              "perforce_changelist_mapping_job_worker",
		Interval:          time.Second, // Poll for a job once per second
		NumHandlers:       1,           // Process only one job at a time (per instance)
		HeartbeatInterval: 10 * time.Second,
		Metrics:           workerutil.NewMetrics(observationCtx, "perforce_changelist_mapping_job_worker"),
	})
}

func makeResetter(observationCtx *observation.Context, workerStore dbworkerstore.Store[*PerforceChangelistMappingJob]) *dbworker.Resetter[*PerforceChangelistMappingJob] {
	return dbworker.NewResetter(observationCtx.Logger, workerStore, dbworker.ResetterOptions{
		Name:     "perforce_changelist_mapping_job_worker_resetter",
		Interval: time.Second * 30, // Check for orphaned jobs every 30 seconds
		Metrics:  dbworker.NewResetterMetrics(observationCtx, "perforce_changelist_mapping_job_worker"),
	})
}

type perforceChangelistMappingJobScheduler struct{}

func (perforceChangelistMappingJobScheduler) Description() string {
	return "Schedule changelist to commit mapping job for a perforce depot imported as a git repo"
}

func (perforceChangelistMappingJobScheduler) Config() []env.Config {
	return nil
}

func (perforceChangelistMappingJobScheduler) Routines(startupCtx context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, errors.Wrap(err, "perforceChangelistMappingJobScheduler.Routines: InitDB failed")
	}

	m := metrics.NewREDMetrics(
		observationCtx.Registerer,
		"permission_sync_job_scheduler",
		metrics.WithCountHelp("Total number of permissions syncer scheduler executions."),
	)
	operation := observationCtx.Operation(observation.Op{
		Name:    "PermissionsSyncer.Scheduler.Run",
		Metrics: m,
	})

}
