package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
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
	ListRepoEmbeddingJobs(ctx context.Context, args ListOpts) ([]*RepoEmbeddingJob, error)
	CountRepoEmbeddingJobs(ctx context.Context, args ListOpts) (int, error)
	GetEmbeddableRepos(ctx context.Context, opts EmbeddableRepoOpts) ([]EmbeddableRepo, error)
	CancelRepoEmbeddingJob(ctx context.Context, job int) error

	UpdateRepoEmbeddingJobStats(ctx context.Context, jobID int, stats *EmbedRepoStats) error
	GetRepoEmbeddingJobStats(ctx context.Context, jobID int) (EmbedRepoStats, error)

	CountRepoEmbeddings(ctx context.Context) (int, error)
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
last_queued_jobs AS (
	SELECT DISTINCT ON (repo_id) repo_id, queued_at
	FROM repo_embedding_jobs
	ORDER BY repo_id, queued_at DESC
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
        %s -- limit clause
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
SELECT DISTINCT ON (rmp.id) rmp.id, rmp.last_changed
FROM repositories_matching_policy rmp
LEFT JOIN last_queued_jobs lqj ON lqj.repo_id = rmp.id
WHERE lqj.queued_at IS NULL OR lqj.queued_at < current_timestamp - (%s * '1 second'::interval);
`

type EmbeddableRepoOpts struct {
	// MinimumInterval is the minimum amount of time that must have passed since the last
	// successful embedding job.
	MinimumInterval time.Duration

	// PolicyRepositoryMatchLimit limits the maximum number of repositories that can
	// be matched by a global policy. If set to nil or a negative value, the policy
	// is unlimited.
	PolicyRepositoryMatchLimit *int
}

type ListOpts struct {
	*database.PaginationArgs
	Query *string
	State *string
	Repo  *api.RepoID
}

func GetEmbeddableRepoOpts() EmbeddableRepoOpts {
	embeddingsConf := conf.GetEmbeddingsConfig(conf.Get().SiteConfig())
	// Embeddings are disabled, nothing we can do.
	if embeddingsConf == nil {
		return EmbeddableRepoOpts{}
	}

	return EmbeddableRepoOpts{
		MinimumInterval:            embeddingsConf.MinimumInterval,
		PolicyRepositoryMatchLimit: embeddingsConf.PolicyRepositoryMatchLimit,
	}
}

func (s *repoEmbeddingJobsStore) GetEmbeddableRepos(ctx context.Context, opts EmbeddableRepoOpts) ([]EmbeddableRepo, error) {
	var limitClause *sqlf.Query
	if opts.PolicyRepositoryMatchLimit != nil && *opts.PolicyRepositoryMatchLimit >= 0 {
		limitClause = sqlf.Sprintf("LIMIT %d", *opts.PolicyRepositoryMatchLimit)
	} else {
		limitClause = sqlf.Sprintf("")
	}

	q := sqlf.Sprintf(
		getEmbeddableReposFmtStr,
		limitClause,
		opts.MinimumInterval.Seconds(),
	)

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

var repoEmbeddingJobStatsColumns = []*sqlf.Query{
	sqlf.Sprintf("repo_embedding_job_stats.job_id"),
	sqlf.Sprintf("repo_embedding_job_stats.is_incremental"),
	sqlf.Sprintf("repo_embedding_job_stats.code_files_total"),
	sqlf.Sprintf("repo_embedding_job_stats.code_files_embedded"),
	sqlf.Sprintf("repo_embedding_job_stats.code_chunks_embedded"),
	sqlf.Sprintf("repo_embedding_job_stats.code_chunks_excluded"),
	sqlf.Sprintf("repo_embedding_job_stats.code_files_skipped"),
	sqlf.Sprintf("repo_embedding_job_stats.code_bytes_embedded"),
	sqlf.Sprintf("repo_embedding_job_stats.text_files_total"),
	sqlf.Sprintf("repo_embedding_job_stats.text_files_embedded"),
	sqlf.Sprintf("repo_embedding_job_stats.text_chunks_embedded"),
	sqlf.Sprintf("repo_embedding_job_stats.text_chunks_excluded"),
	sqlf.Sprintf("repo_embedding_job_stats.text_files_skipped"),
	sqlf.Sprintf("repo_embedding_job_stats.text_bytes_embedded"),
}

func scanRepoEmbeddingStats(s dbutil.Scanner) (EmbedRepoStats, error) {
	var stats EmbedRepoStats
	var jobID int
	err := s.Scan(
		&jobID,
		&stats.IsIncremental,
		&stats.CodeIndexStats.FilesScheduled,
		&stats.CodeIndexStats.FilesEmbedded,
		&stats.CodeIndexStats.ChunksEmbedded,
		&stats.CodeIndexStats.ChunksExcluded,
		dbutil.JSONMessage(&stats.CodeIndexStats.FilesSkipped),
		&stats.CodeIndexStats.BytesEmbedded,
		&stats.TextIndexStats.FilesScheduled,
		&stats.TextIndexStats.FilesEmbedded,
		&stats.TextIndexStats.ChunksEmbedded,
		&stats.TextIndexStats.ChunksExcluded,
		dbutil.JSONMessage(&stats.TextIndexStats.FilesSkipped),
		&stats.TextIndexStats.BytesEmbedded,
	)
	return stats, err
}

func (s *repoEmbeddingJobsStore) GetRepoEmbeddingJobStats(ctx context.Context, jobID int) (EmbedRepoStats, error) {
	const getRepoEmbeddingJobStats = `SELECT %s FROM repo_embedding_job_stats WHERE job_id = %s`
	q := sqlf.Sprintf(
		getRepoEmbeddingJobStats,
		sqlf.Join(repoEmbeddingJobStatsColumns, ","),
		jobID,
	)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return EmbedRepoStats{}, err
	}
	defer rows.Close()

	if !rows.Next() {
		return EmbedRepoStats{}, nil // not an error condition, just no progress
	}

	return scanRepoEmbeddingStats(rows)
}

func (s *repoEmbeddingJobsStore) UpdateRepoEmbeddingJobStats(ctx context.Context, jobID int, stats *EmbedRepoStats) error {
	const updateRepoEmbeddingJobStats = `
	INSERT INTO repo_embedding_job_stats (
		job_id,
		is_incremental,
		code_files_total,
		code_files_embedded,
		code_chunks_embedded,
		code_chunks_excluded,
		code_files_skipped,
		code_bytes_embedded,
		text_files_total,
		text_files_embedded,
		text_chunks_embedded,
		text_chunks_excluded,
		text_files_skipped,
		text_bytes_embedded
	) VALUES (
		%s, %s, %s, %s,
		%s, %s, %s, %s,
		%s, %s, %s, %s,
	    %s, %s
	)
	ON CONFLICT (job_id) DO UPDATE
	SET
		is_incremental = %s,
		code_files_total = %s,
		code_files_embedded = %s,
		code_chunks_embedded = %s,
		code_chunks_excluded = %s,
		code_files_skipped = %s,
		code_bytes_embedded = %s,
		text_files_total = %s,
		text_files_embedded = %s,
		text_chunks_embedded = %s,
		text_chunks_excluded = %s,
		text_files_skipped = %s,
		text_bytes_embedded = %s
	`

	q := sqlf.Sprintf(
		updateRepoEmbeddingJobStats,

		jobID,
		stats.IsIncremental,
		stats.CodeIndexStats.FilesScheduled,
		stats.CodeIndexStats.FilesEmbedded,
		stats.CodeIndexStats.ChunksEmbedded,
		stats.CodeIndexStats.ChunksExcluded,
		dbutil.JSONMessage(&stats.CodeIndexStats.FilesSkipped),
		stats.CodeIndexStats.BytesEmbedded,
		stats.TextIndexStats.FilesScheduled,
		stats.TextIndexStats.FilesEmbedded,
		stats.TextIndexStats.ChunksEmbedded,
		stats.TextIndexStats.ChunksExcluded,
		dbutil.JSONMessage(&stats.TextIndexStats.FilesSkipped),
		stats.TextIndexStats.BytesEmbedded,

		stats.IsIncremental,
		stats.CodeIndexStats.FilesScheduled,
		stats.CodeIndexStats.FilesEmbedded,
		stats.CodeIndexStats.ChunksEmbedded,
		stats.CodeIndexStats.ChunksExcluded,
		dbutil.JSONMessage(&stats.CodeIndexStats.FilesSkipped),
		stats.CodeIndexStats.BytesEmbedded,
		stats.TextIndexStats.FilesScheduled,
		stats.TextIndexStats.FilesEmbedded,
		stats.TextIndexStats.ChunksEmbedded,
		stats.TextIndexStats.ChunksExcluded,
		dbutil.JSONMessage(&stats.TextIndexStats.FilesSkipped),
		stats.TextIndexStats.BytesEmbedded,
	)

	return s.Exec(ctx, q)
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
%s -- joinClause
%s -- whereClause
`

// CountRepoEmbeddingJobs returns the number of repo embedding jobs that match
// the query. If there is no query, all repo embedding jobs are counted.
func (s *repoEmbeddingJobsStore) CountRepoEmbeddingJobs(ctx context.Context, opts ListOpts) (int, error) {
	var conds []*sqlf.Query

	var joinClause *sqlf.Query
	if opts.Query != nil && *opts.Query != "" {
		conds = append(conds, sqlf.Sprintf("repo.name LIKE %s", "%"+*opts.Query+"%"))
		joinClause = sqlf.Sprintf("JOIN repo ON repo.id = repo_embedding_jobs.repo_id")
	} else {
		joinClause = sqlf.Sprintf("")
	}

	if opts.State != nil && *opts.State != "" {
		conds = append(conds, sqlf.Sprintf("repo_embedding_jobs.state = %s", strings.ToLower(*opts.State)))
	}

	if opts.Repo != nil {
		conds = append(conds, sqlf.Sprintf("repo_embedding_jobs.repo_id = %d", *opts.Repo))
	}

	var whereClause *sqlf.Query
	if len(conds) != 0 {
		whereClause = sqlf.Sprintf("WHERE %s", sqlf.Join(conds, "\n AND "))
	} else {
		whereClause = sqlf.Sprintf("")
	}

	q := sqlf.Sprintf(countRepoEmbeddingJobsQuery, joinClause, whereClause)
	var count int
	if err := s.QueryRow(ctx, q).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

const listRepoEmbeddingJobsQueryFmtstr = `
SELECT %s
FROM repo_embedding_jobs
%s -- joinClause
%s -- whereClause
`

func (s *repoEmbeddingJobsStore) ListRepoEmbeddingJobs(ctx context.Context, opts ListOpts) ([]*RepoEmbeddingJob, error) {
	pagination := opts.PaginationArgs.SQL()

	var conds []*sqlf.Query
	if pagination.Where != nil {
		conds = append(conds, pagination.Where)
	}

	var joinClause *sqlf.Query
	if opts.Query != nil && *opts.Query != "" {
		conds = append(conds, sqlf.Sprintf("repo.name LIKE %s", "%"+*opts.Query+"%"))
		joinClause = sqlf.Sprintf("JOIN repo ON repo.id = repo_embedding_jobs.repo_id")
	} else {
		joinClause = sqlf.Sprintf("")
	}

	if opts.State != nil && *opts.State != "" {
		conds = append(conds, sqlf.Sprintf("repo_embedding_jobs.state = %s", strings.ToLower(*opts.State)))
	}

	if opts.Repo != nil {
		conds = append(conds, sqlf.Sprintf("repo_embedding_jobs.repo_id = %d", *opts.Repo))
	}

	var whereClause *sqlf.Query
	if len(conds) != 0 {
		whereClause = sqlf.Sprintf("WHERE %s", sqlf.Join(conds, "\n AND "))
	} else {
		whereClause = sqlf.Sprintf("")
	}

	q := sqlf.Sprintf(listRepoEmbeddingJobsQueryFmtstr, sqlf.Join(repoEmbeddingJobsColumns, ", "), joinClause, whereClause)
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

const countRepoEmbeddingsQuery = `
SELECT COUNT(DISTINCT repo_id) AS count
FROM repo_embedding_jobs
WHERE state = 'completed';
`

func (s *repoEmbeddingJobsStore) CountRepoEmbeddings(ctx context.Context) (int, error) {
	return basestore.ScanInt(s.QueryRow(ctx, sqlf.Sprintf(countRepoEmbeddingsQuery)))
}
