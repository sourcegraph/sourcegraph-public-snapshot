package contextdetection

import (
	"time"

	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/observation"

	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

var repoEmbeddingJobsColumns = []*sqlf.Query{
	sqlf.Sprintf("context_detection_embedding_jobs.id"),
	sqlf.Sprintf("context_detection_embedding_jobs.state"),
	sqlf.Sprintf("context_detection_embedding_jobs.failure_message"),
	sqlf.Sprintf("context_detection_embedding_jobs.queued_at"),
	sqlf.Sprintf("context_detection_embedding_jobs.started_at"),
	sqlf.Sprintf("context_detection_embedding_jobs.finished_at"),
	sqlf.Sprintf("context_detection_embedding_jobs.process_after"),
	sqlf.Sprintf("context_detection_embedding_jobs.num_resets"),
	sqlf.Sprintf("context_detection_embedding_jobs.num_failures"),
	sqlf.Sprintf("context_detection_embedding_jobs.last_heartbeat_at"),
	sqlf.Sprintf("context_detection_embedding_jobs.execution_logs"),
	sqlf.Sprintf("context_detection_embedding_jobs.worker_hostname"),
	sqlf.Sprintf("context_detection_embedding_jobs.cancel"),
}

func scanContextDetectionEmbeddingJob(s dbutil.Scanner) (*ContextDetectionEmbeddingJob, error) {
	var job ContextDetectionEmbeddingJob
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
	); err != nil {
		return nil, err
	}
	job.ExecutionLogs = append(job.ExecutionLogs, executionLogs...)
	return &job, nil
}

func NewContextDetectionEmbeddingJobWorkerStore(observationCtx *observation.Context, dbHandle basestore.TransactableHandle) dbworkerstore.Store[*ContextDetectionEmbeddingJob] {
	return dbworkerstore.New(observationCtx, dbHandle, dbworkerstore.Options[*ContextDetectionEmbeddingJob]{
		Name:              "context_detection_embedding_job_worker",
		TableName:         "context_detection_embedding_jobs",
		ColumnExpressions: repoEmbeddingJobsColumns,
		Scan:              dbworkerstore.BuildWorkerScan(scanContextDetectionEmbeddingJob),
		OrderByExpression: sqlf.Sprintf("context_detection_embedding_jobs.queued_at, context_detection_embedding_jobs.id"),
		StalledMaxAge:     time.Second * 60,
		MaxNumResets:      5,
		MaxNumRetries:     1,
	})
}

type ContextDetectionEmbeddingJobsStore interface {
	basestore.ShareableStore

	CreateContextDetectionEmbeddingJob(ctx context.Context) (int, error)
}

type contextDetectionEmbeddingJobsStore struct {
	*basestore.Store
}

func NewContextDetectionEmbeddingJobsStore(other basestore.ShareableStore) ContextDetectionEmbeddingJobsStore {
	return &contextDetectionEmbeddingJobsStore{Store: basestore.NewWithHandle(other.Handle())}
}

var _ basestore.ShareableStore = &contextDetectionEmbeddingJobsStore{}

var createContextDetectionEmbeddingJobFmtStr = `INSERT INTO context_detection_embedding_jobs DEFAULT VALUES RETURNING id`

func (s *contextDetectionEmbeddingJobsStore) CreateContextDetectionEmbeddingJob(ctx context.Context) (int, error) {
	q := sqlf.Sprintf(createContextDetectionEmbeddingJobFmtStr)
	id, _, err := basestore.ScanFirstInt(s.Query(ctx, q))
	return id, err
}
