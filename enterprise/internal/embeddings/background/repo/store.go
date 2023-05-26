package repo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type RepoEmbeddingJobNotFoundErr struct {
	repoID api.RepoID
}

func (r *RepoEmbeddingJobNotFoundErr) Error() string {
	return fmt.Sprintf("repo embedding job not found: repoID=%d", r.repoID)
}

func (r *RepoEmbeddingJobNotFoundErr) NotFound() bool {
	return true
}

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
		StalledMaxAge:     time.Second * 60,
		MaxNumResets:      5,
		MaxNumRetries:     1,
	})
}

type RepoEmbeddingJobsStore interface {
	basestore.ShareableStore

	Transact(ctx context.Context) (RepoEmbeddingJobsStore, error)
	Exec(ctx context.Context, query *sqlf.Query) error
	Done(err error) error

	CreateRepoEmbeddingJob(ctx context.Context, repoID api.RepoID, revision api.CommitID) (int, error)
	GetLastCompletedRepoEmbeddingJob(ctx context.Context, repoID api.RepoID) (*RepoEmbeddingJob, error)
	GetLastRepoEmbeddingJobForRevision(ctx context.Context, repoID api.RepoID, revision api.CommitID) (*RepoEmbeddingJob, error)
	ListRepoEmbeddingJobs(ctx context.Context, args *database.PaginationArgs) ([]*RepoEmbeddingJob, error)
	CountRepoEmbeddingJobs(ctx context.Context) (int, error)
	GetEmbeddableRepos(ctx context.Context) ([]EmbeddableRepo, error)
	CancelRepoEmbeddingJob(ctx context.Context, job int) error
}

var _ basestore.ShareableStore = &repoEmbeddingJobsStore{}

type repoEmbeddingJobsStore struct {
	*basestore.Store
}

type EmbeddableRepo struct {
	ID          api.RepoID
	lastChanged time.Time
}

var scanEmbeddableRepos = basestore.NewSliceScanner(func(scanner dbutil.Scanner) (r EmbeddableRepo, _ error) {
	err := scanner.Scan(&r.ID, &r.lastChanged)
	return r, err
})

const getEmbeddableReposFmtStr = `
WITH
global_policy_descriptor AS MATERIALIZED (
	SELECT 1
	FROM codeintel_configuration_policies p
	WHERE
		p.embeddings_enabled AND
		p.repository_id IS NULL AND
		p.repository_patterns IS NULL
	LIMIT 1
),
repositories_matching_policy AS (
    (
        SELECT r.id, gr.last_changed
        FROM repo r
        JOIN gitserver_repos gr ON gr.repo_id = r.id
        JOIN global_policy_descriptor gpd ON TRUE
        WHERE
            r.deleted_at IS NULL AND
            r.blocked IS NULL AND
            gr.clone_status = 'cloned'
        ORDER BY stars DESC NULLS LAST, id
        LIMIT 5000 -- Some repository match limit to stop you from returning all of dotcom
    ) UNION ALL (
        SELECT r.id, gr.last_changed
        FROM repo r
        JOIN gitserver_repos gr ON gr.repo_id = r.id
        JOIN codeintel_configuration_policies p ON p.repository_id = r.id
        WHERE
            r.deleted_at IS NULL AND
            r.blocked IS NULL AND
            p.embeddings_enabled AND
            gr.clone_status = 'cloned'
    ) UNION ALL (
        SELECT r.id, gr.last_changed
        FROM repo r
        JOIN gitserver_repos gr ON gr.repo_id = r.id
        JOIN codeintel_configuration_policies_repository_pattern_lookup rpl ON rpl.repo_id = r.id
        JOIN codeintel_configuration_policies p ON p.id = rpl.policy_id
        WHERE
            r.deleted_at IS NULL AND
            r.blocked IS NULL AND
            p.embeddings_enabled AND
            gr.clone_status = 'cloned'
    )
)

--
SELECT DISTINCT ON (id) * FROM repositories_matching_policy;
`

func (s *repoEmbeddingJobsStore) GetEmbeddableRepos(ctx context.Context) ([]EmbeddableRepo, error) {
	q := sqlf.Sprintf(getEmbeddableReposFmtStr)
	return scanEmbeddableRepos(s.Query(ctx, q))
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
	job, err := scanRepoEmbeddingJob(s.QueryRow(ctx, q))
	if err == sql.ErrNoRows {
		return nil, &RepoEmbeddingJobNotFoundErr{repoID: repoID}
	}
	return job, nil
}

const getLastRepoEmbeddingJobForRevision = `
SELECT %s
FROM repo_embedding_jobs
WHERE repo_id = %d AND revision = %s
ORDER BY queued_at DESC
LIMIT 1
`

func (s *repoEmbeddingJobsStore) GetLastRepoEmbeddingJobForRevision(ctx context.Context, repoID api.RepoID, revision api.CommitID) (*RepoEmbeddingJob, error) {
	q := sqlf.Sprintf(getLastRepoEmbeddingJobForRevision, sqlf.Join(repoEmbeddingJobsColumns, ", "), repoID, revision)
	job, err := scanRepoEmbeddingJob(s.QueryRow(ctx, q))
	if err == sql.ErrNoRows {
		return nil, &RepoEmbeddingJobNotFoundErr{repoID: repoID}
	}
	return job, nil
}

const countRepoEmbeddingJobsQuery = `
SELECT COUNT(*)
FROM repo_embedding_jobs
`

func (s *repoEmbeddingJobsStore) CountRepoEmbeddingJobs(ctx context.Context) (int, error) {
	q := sqlf.Sprintf(countRepoEmbeddingJobsQuery)
	var count int
	if err := s.QueryRow(ctx, q).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

const listRepoEmbeddingJobsQueryFmtstr = `
SELECT %s
FROM repo_embedding_jobs
%s -- whereClause
`

func (s *repoEmbeddingJobsStore) ListRepoEmbeddingJobs(ctx context.Context, paginationArgs *database.PaginationArgs) ([]*RepoEmbeddingJob, error) {
	pagination := paginationArgs.SQL()

	var conds []*sqlf.Query
	if pagination.Where != nil {
		conds = append(conds, pagination.Where)
	}

	var whereClause *sqlf.Query
	if len(conds) != 0 {
		whereClause = sqlf.Sprintf("WHERE %s", sqlf.Join(conds, "\n AND "))
	} else {
		whereClause = sqlf.Sprintf("")
	}

	q := sqlf.Sprintf(listRepoEmbeddingJobsQueryFmtstr, sqlf.Join(repoEmbeddingJobsColumns, ", "), whereClause)
	q = pagination.AppendOrderToQuery(q)
	q = pagination.AppendLimitToQuery(q)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var jobs []*RepoEmbeddingJob
	for rows.Next() {
		job, err := scanRepoEmbeddingJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func (s *repoEmbeddingJobsStore) CancelRepoEmbeddingJob(ctx context.Context, jobID int) error {
	now := time.Now()
	q := sqlf.Sprintf(cancelRepoEmbeddingJobQueryFmtstr, now, jobID)

	res, err := s.ExecResult(ctx, q)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return errors.Newf("could not find cancellable embedding job: jobID=%d", jobID)
	}
	return nil
}

const cancelRepoEmbeddingJobQueryFmtstr = `
UPDATE
	repo_embedding_jobs
SET
    cancel = TRUE,
    -- If the embeddings job is still queued, we directly abort, otherwise we keep the
    -- state, so the worker can do teardown and later mark it failed.
    state = CASE WHEN repo_embedding_jobs.state = 'processing' THEN repo_embedding_jobs.state ELSE 'canceled' END,
    finished_at = CASE WHEN repo_embedding_jobs.state = 'processing' THEN repo_embedding_jobs.finished_at ELSE %s END
WHERE
	id = %d
	AND
	state IN ('queued', 'processing')
`
