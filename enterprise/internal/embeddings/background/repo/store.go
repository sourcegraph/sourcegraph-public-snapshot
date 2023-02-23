package repo

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

var repoEmbeddingJobsColumns = []*sqlf.Query{
	sqlf.Sprintf("repo_embedding_jobs.id"),
	sqlf.Sprintf("repo_embedding_jobs.state"),
	sqlf.Sprintf("repo_embedding_jobs.failure_message"),
	sqlf.Sprintf("repo_embedding_jobs.queued_at"),
	sqlf.Sprintf("repo_embedding_jobs.started_at"),
	sqlf.Sprintf("repo_embedding_jobs.finished_at"),
	sqlf.Sprintf("repo_embedding_jobs.process_after"),
	sqlf.Sprintf("repo_embedding_jobs.num_resets"),
	sqlf.Sprintf("repo_embedding_jobs.num_failures"),
	sqlf.Sprintf("repo_embedding_jobs.last_heartbeat_at"),
	sqlf.Sprintf("repo_embedding_jobs.execution_logs"),
	sqlf.Sprintf("repo_embedding_jobs.worker_hostname"),
	sqlf.Sprintf("repo_embedding_jobs.cancel"),
	sqlf.Sprintf("repo_embedding_jobs.repo_id"),
	sqlf.Sprintf("repo_embedding_jobs.revision"),
}

func scanRepoEmbeddingJob(s dbutil.Scanner) (*RepoEmbeddingJob, error) {
	var job RepoEmbeddingJob
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

func NewRepoEmbeddingJobWorkerStore(observationCtx *observation.Context, dbHandle basestore.TransactableHandle) dbworkerstore.Store[*RepoEmbeddingJob] {
	return dbworkerstore.New(observationCtx, dbHandle, dbworkerstore.Options[*RepoEmbeddingJob]{
		Name:              "repo_embedding_job_worker",
		TableName:         "repo_embedding_jobs",
		ColumnExpressions: repoEmbeddingJobsColumns,
		Scan:              dbworkerstore.BuildWorkerScan(scanRepoEmbeddingJob),
		OrderByExpression: sqlf.Sprintf("repo_embedding_jobs.queued_at, repo_embedding_jobs.id"),
		StalledMaxAge:     time.Second * 5,
		MaxNumResets:      5,
	})
}

type RepoEmbeddingJobsStore interface {
	basestore.ShareableStore

	Transact(ctx context.Context) (RepoEmbeddingJobsStore, error)
	Done(err error) error

	CreateRepoEmbeddingJob(ctx context.Context, repoID api.RepoID, revision api.CommitID) (int, error)
	GetLastCompletedRepoEmbeddingJob(ctx context.Context, repoID api.RepoID) (*RepoEmbeddingJob, error)
}

var _ basestore.ShareableStore = &repoEmbeddingJobsStore{}

type repoEmbeddingJobsStore struct {
	*basestore.Store
}

func NewRepoEmbeddingJobsStore(other basestore.ShareableStore) RepoEmbeddingJobsStore {
	return &repoEmbeddingJobsStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (s *repoEmbeddingJobsStore) Transact(ctx context.Context) (RepoEmbeddingJobsStore, error) {
	tx, err := s.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	return &repoEmbeddingJobsStore{Store: tx}, nil
}

const createRepoEmbeddingJobFmtStr = `INSERT INTO repo_embedding_jobs (repo_id, revision) VALUES (%s, %s) RETURNING id`

func (s *repoEmbeddingJobsStore) CreateRepoEmbeddingJob(ctx context.Context, repoID api.RepoID, revision api.CommitID) (int, error) {
	q := sqlf.Sprintf(createRepoEmbeddingJobFmtStr, repoID, revision)
	id, _, err := basestore.ScanFirstInt(s.Query(ctx, q))
	return id, err
}

const getLastFinishedRepoEmbeddingJob = `
SELECT %s
FROM repo_embedding_jobs
WHERE state = 'completed' AND repo_id = %d
ORDER BY finished_at DESC
LIMIT 1
`

func (s *repoEmbeddingJobsStore) GetLastCompletedRepoEmbeddingJob(ctx context.Context, repoID api.RepoID) (*RepoEmbeddingJob, error) {
	q := sqlf.Sprintf(getLastFinishedRepoEmbeddingJob, sqlf.Join(repoEmbeddingJobsColumns, ", "), repoID)
	return scanRepoEmbeddingJob(s.QueryRow(ctx, q))
}
