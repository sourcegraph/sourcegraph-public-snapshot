package embeddings

import (
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

var embeddingJobColumns = []*sqlf.Query{
	sqlf.Sprintf("embedding_jobs.id"),
	sqlf.Sprintf("embedding_jobs.state"),
	sqlf.Sprintf("embedding_jobs.failure_message"),
	sqlf.Sprintf("embedding_jobs.queued_at"),
	sqlf.Sprintf("embedding_jobs.started_at"),
	sqlf.Sprintf("embedding_jobs.finished_at"),
	sqlf.Sprintf("embedding_jobs.process_after"),
	sqlf.Sprintf("embedding_jobs.num_resets"),
	sqlf.Sprintf("embedding_jobs.num_failures"),
	sqlf.Sprintf("embedding_jobs.last_heartbeat_at"),
	sqlf.Sprintf("embedding_jobs.execution_logs"),
	sqlf.Sprintf("embedding_jobs.worker_hostname"),
	sqlf.Sprintf("embedding_jobs.cancel"),
	sqlf.Sprintf("embedding_jobs.repo_id"),
	sqlf.Sprintf("embedding_jobs.revision"),
}

func scanEmbeddingJob(s dbutil.Scanner) (*EmbeddingJob, error) {
	var job EmbeddingJob
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
		&job.RepoID,
		&job.Revision,
	); err != nil {
		return nil, err
	}
	job.ExecutionLogs = append(job.ExecutionLogs, executionLogs...)
	return &job, nil
}

func newEmbeddingJobWorkerStore(observationCtx *observation.Context, dbHandle basestore.TransactableHandle) dbworkerstore.Store[*EmbeddingJob] {
	return dbworkerstore.New(observationCtx, dbHandle, dbworkerstore.Options[*EmbeddingJob]{
		Name:              "embedding_job_worker",
		TableName:         "embedding_jobs",
		ColumnExpressions: embeddingJobColumns,
		Scan:              dbworkerstore.BuildWorkerScan(scanEmbeddingJob),
		OrderByExpression: sqlf.Sprintf("embedding_jobs.queued_at, embedding_jobs.id"),
		StalledMaxAge:     time.Second * 5,
		MaxNumResets:      5,
	})
}
