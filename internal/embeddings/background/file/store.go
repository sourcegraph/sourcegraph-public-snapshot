package file

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	bgrepo "github.com/sourcegraph/sourcegraph/internal/embeddings/background/repo"
	"github.com/sourcegraph/sourcegraph/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type FileEmbeddingJobNotFoundErr struct {
	embeddingPluginID int32
}

func (r *FileEmbeddingJobNotFoundErr) Error() string {
	return fmt.Sprintf("file embedding job not found: embeddingPluginID=%d", r.embeddingPluginID)
}

func (r *FileEmbeddingJobNotFoundErr) NotFound() bool {
	return true
}

var fileEmbeddingJobsColumns = []*sqlf.Query{
	sqlf.Sprintf("file_embedding_jobs.id"),
	sqlf.Sprintf("file_embedding_jobs.state"),
	sqlf.Sprintf("file_embedding_jobs.failure_message"),
	sqlf.Sprintf("file_embedding_jobs.queued_at"),
	sqlf.Sprintf("file_embedding_jobs.started_at"),
	sqlf.Sprintf("file_embedding_jobs.finished_at"),
	sqlf.Sprintf("file_embedding_jobs.process_after"),
	sqlf.Sprintf("file_embedding_jobs.num_resets"),
	sqlf.Sprintf("file_embedding_jobs.num_failures"),
	sqlf.Sprintf("file_embedding_jobs.last_heartbeat_at"),
	sqlf.Sprintf("file_embedding_jobs.execution_logs"),
	sqlf.Sprintf("file_embedding_jobs.worker_hostname"),
	sqlf.Sprintf("file_embedding_jobs.cancel"),

	sqlf.Sprintf("file_embedding_jobs.embedding_plugin_id"),
	sqlf.Sprintf("file_embedding_jobs.file_type"),
}

func scanFileEmbeddingJob(s dbutil.Scanner) (*FileEmbeddingJob, error) {
	var job FileEmbeddingJob
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
		&job.EmbeddingPluginID,
		&job.FileType,
	); err != nil {
		return nil, err
	}
	job.ExecutionLogs = append(job.ExecutionLogs, executionLogs...)
	return &job, nil
}

func NewFileEmbeddingJobWorkerStore(observationCtx *observation.Context, dbHandle basestore.TransactableHandle) dbworkerstore.Store[*FileEmbeddingJob] {
	return dbworkerstore.New(observationCtx, dbHandle, dbworkerstore.Options[*FileEmbeddingJob]{
		Name:              "file_embedding_job_worker",
		TableName:         "file_embedding_jobs",
		ColumnExpressions: fileEmbeddingJobsColumns,
		Scan:              dbworkerstore.BuildWorkerScan(scanFileEmbeddingJob),
		OrderByExpression: sqlf.Sprintf("file_embedding_jobs.queued_at, file_embedding_jobs.id"),
		StalledMaxAge:     time.Second * 60,
		MaxNumResets:      5,
		MaxNumRetries:     1,
	})
}

type FileEmbeddingJobsStore interface {
	basestore.ShareableStore

	Transact(ctx context.Context) (FileEmbeddingJobsStore, error)
	Exec(ctx context.Context, query *sqlf.Query) error
	Done(err error) error

	CreateFileEmbeddingJob(ctx context.Context, embeddingPluginID int32, fileType string) (int, error)
	GetLastCompletedFileEmbeddingJob(ctx context.Context, repoID int32) (*FileEmbeddingJob, error)
	ListFileEmbeddingJobs(ctx context.Context, args ListOpts) ([]*FileEmbeddingJob, error)
	CountFileEmbeddingJobs(ctx context.Context, args ListOpts) (int, error)
	GetEmbeddableFiles(ctx context.Context, opts EmbeddableFileOpts) ([]EmbeddableFile, error)
	CancelFileEmbeddingJob(ctx context.Context, job int) error

	UpdateFileEmbeddingJobStats(ctx context.Context, jobID int, stats *bgrepo.EmbedRepoStats) error
	GetFileEmbeddingJobStats(ctx context.Context, jobID int) (EmbedFileStats, error)
}

var _ basestore.ShareableStore = &fileEmbeddingJobsStore{}

type fileEmbeddingJobsStore struct {
	*basestore.Store
}

type EmbeddableFile struct {
	ID          int32
	lastChanged time.Time
}

var scanEmbeddableFiles = basestore.NewSliceScanner(func(scanner dbutil.Scanner) (r EmbeddableFile, _ error) {
	err := scanner.Scan(&r.ID, &r.lastChanged)
	return r, err
})

const getEmbeddableFilesFmtStr = `
WITH
last_queued_jobs AS (
	SELECT DISTINCT ON (embedding_plugin_id) embedding_plugin_id, queued_at
	FROM file_embedding_jobs
	ORDER BY embedding_plugin_id, queued_at DESC
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

type EmbeddableFileOpts struct {
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
	Query           *string
	State           *string
	EmbeddingPlugin *int32
}

func GetEmbeddableFileOpts() EmbeddableFileOpts {
	embeddingsConf := conf.GetEmbeddingsConfig(conf.Get().SiteConfig())
	// Embeddings are disabled, nothing we can do.
	if embeddingsConf == nil {
		return EmbeddableFileOpts{}
	}

	return EmbeddableFileOpts{
		MinimumInterval:            embeddingsConf.MinimumInterval,
		PolicyRepositoryMatchLimit: embeddingsConf.PolicyRepositoryMatchLimit,
	}
}

func (s *fileEmbeddingJobsStore) GetEmbeddableFiles(ctx context.Context, opts EmbeddableFileOpts) ([]EmbeddableFile, error) {
	var limitClause *sqlf.Query
	if opts.PolicyRepositoryMatchLimit != nil && *opts.PolicyRepositoryMatchLimit >= 0 {
		limitClause = sqlf.Sprintf("LIMIT %d", *opts.PolicyRepositoryMatchLimit)
	} else {
		limitClause = sqlf.Sprintf("")
	}

	q := sqlf.Sprintf(
		getEmbeddableFilesFmtStr,
		limitClause,
		opts.MinimumInterval.Seconds(),
	)

	return scanEmbeddableFiles(s.Query(ctx, q))
}

func NewFileEmbeddingJobsStore(other basestore.ShareableStore) FileEmbeddingJobsStore {
	return &fileEmbeddingJobsStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (s *fileEmbeddingJobsStore) Transact(ctx context.Context) (FileEmbeddingJobsStore, error) {
	tx, err := s.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	return &fileEmbeddingJobsStore{Store: tx}, nil
}

const createFileEmbeddingJobFmtStr = `INSERT INTO file_embedding_jobs (embedding_plugin_id, file_type) VALUES (%s, %s) RETURNING id`

func (s *fileEmbeddingJobsStore) CreateFileEmbeddingJob(ctx context.Context, embeddingPluginID int32, fileType string) (int, error) {
	q := sqlf.Sprintf(createFileEmbeddingJobFmtStr, embeddingPluginID, fileType)
	id, _, err := basestore.ScanFirstInt(s.Query(ctx, q))
	return id, err
}

var fileEmbeddingJobStatsColumns = []*sqlf.Query{
	sqlf.Sprintf("file_embedding_job_stats.job_id"),
	sqlf.Sprintf("file_embedding_job_stats.is_incremental"),
	sqlf.Sprintf("file_embedding_job_stats.files_total"),
	sqlf.Sprintf("file_embedding_job_stats.files_embedded"),
	sqlf.Sprintf("file_embedding_job_stats.chunks_embedded"),
	sqlf.Sprintf("file_embedding_job_stats.files_skipped"),
	sqlf.Sprintf("file_embedding_job_stats.bytes_embedded"),
}

func scanFileEmbeddingStats(s dbutil.Scanner) (EmbedFileStats, error) {
	var stats EmbedFileStats
	var jobID int
	err := s.Scan(
		&jobID,
		&stats.IsIncremental,
		&stats.CodeIndexStats.FilesScheduled,
		&stats.CodeIndexStats.FilesEmbedded,
		&stats.CodeIndexStats.ChunksEmbedded,
		dbutil.JSONMessage(&stats.CodeIndexStats.FilesSkipped),
		&stats.CodeIndexStats.BytesEmbedded,
	)
	return stats, err
}

func (s *fileEmbeddingJobsStore) GetFileEmbeddingJobStats(ctx context.Context, jobID int) (EmbedFileStats, error) {
	const getFileEmbeddingJobStats = `SELECT %s FROM file_embedding_job_stats WHERE job_id = %s`
	q := sqlf.Sprintf(
		getFileEmbeddingJobStats,
		sqlf.Join(fileEmbeddingJobStatsColumns, ","),
		jobID,
	)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return EmbedFileStats{}, err
	}
	defer rows.Close()

	if !rows.Next() {
		return EmbedFileStats{}, nil // not an error condition, just no progress
	}

	return scanFileEmbeddingStats(rows)
}

func (s *fileEmbeddingJobsStore) UpdateFileEmbeddingJobStats(ctx context.Context, jobID int, stats *bgrepo.EmbedRepoStats) error {
	const updateFileEmbeddingJobStats = `
	INSERT INTO file_embedding_job_stats (
		job_id,
		is_incremental,
		files_total,
		files_embedded,
		chunks_embedded,
		files_skipped,
		bytes_embedded
	) VALUES (
		%s, %s, %s, %s,
		%s, %s, %s
	)
	ON CONFLICT (job_id) DO UPDATE
	SET
		is_incremental = %s,
		files_total = %s,
		files_embedded = %s,
		chunks_embedded = %s,
		files_skipped = %s,
		bytes_embedded = %s
	`

	q := sqlf.Sprintf(
		updateFileEmbeddingJobStats,

		jobID,
		stats.IsIncremental,
		stats.CodeIndexStats.FilesScheduled,
		stats.CodeIndexStats.FilesEmbedded,
		stats.CodeIndexStats.ChunksEmbedded,
		dbutil.JSONMessage(&stats.CodeIndexStats.FilesSkipped),
		stats.CodeIndexStats.BytesEmbedded,

		stats.IsIncremental,
		stats.CodeIndexStats.FilesScheduled,
		stats.CodeIndexStats.FilesEmbedded,
		stats.CodeIndexStats.ChunksEmbedded,
		dbutil.JSONMessage(&stats.CodeIndexStats.FilesSkipped),
		stats.CodeIndexStats.BytesEmbedded,
	)

	return s.Exec(ctx, q)
}

const getLastFinishedFileEmbeddingJob = `
SELECT %s
FROM file_embedding_jobs
WHERE state = 'completed' AND embedding_plugin_id = %d
ORDER BY finished_at DESC
LIMIT 1
`

func (s *fileEmbeddingJobsStore) GetLastCompletedFileEmbeddingJob(ctx context.Context, embeddingPluginID int32) (*FileEmbeddingJob, error) {
	q := sqlf.Sprintf(getLastFinishedFileEmbeddingJob, sqlf.Join(fileEmbeddingJobsColumns, ", "), embeddingPluginID)
	job, err := scanFileEmbeddingJob(s.QueryRow(ctx, q))
	if err == sql.ErrNoRows {
		return nil, &FileEmbeddingJobNotFoundErr{embeddingPluginID: embeddingPluginID}
	}
	return job, nil
}

const countFileEmbeddingJobsQuery = `
SELECT COUNT(*)
FROM file_embedding_jobs
%s -- joinClause
%s -- whereClause
`

// CountRepoEmbeddingJobs returns the number of repo embedding jobs that match
// the query. If there is no query, all repo embedding jobs are counted.
func (s *fileEmbeddingJobsStore) CountFileEmbeddingJobs(ctx context.Context, opts ListOpts) (int, error) {
	var conds []*sqlf.Query

	var joinClause *sqlf.Query
	if opts.Query != nil && *opts.Query != "" {
		conds = append(conds, sqlf.Sprintf("embedding_plugins.name LIKE %s", "%"+*opts.Query+"%"))
		joinClause = sqlf.Sprintf("JOIN embedding_plugins ON embedding_plugins.id = file_embedding_jobs.embedding_plugin_id")
	} else {
		joinClause = sqlf.Sprintf("")
	}

	if opts.State != nil && *opts.State != "" {
		conds = append(conds, sqlf.Sprintf("file_embedding_jobs.state = %s", strings.ToLower(*opts.State)))
	}

	if opts.EmbeddingPlugin != nil {
		conds = append(conds, sqlf.Sprintf("file_embedding_jobs.embedding_plugin_id = %d", *opts.EmbeddingPlugin))
	}

	var whereClause *sqlf.Query
	if len(conds) != 0 {
		whereClause = sqlf.Sprintf("WHERE %s", sqlf.Join(conds, "\n AND "))
	} else {
		whereClause = sqlf.Sprintf("")
	}

	q := sqlf.Sprintf(countFileEmbeddingJobsQuery, joinClause, whereClause)
	var count int
	if err := s.QueryRow(ctx, q).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

const listFileEmbeddingJobsQueryFmtstr = `
SELECT %s
FROM file_embedding_jobs
%s -- joinClause
%s -- whereClause
`

func (s *fileEmbeddingJobsStore) ListFileEmbeddingJobs(ctx context.Context, opts ListOpts) ([]*FileEmbeddingJob, error) {
	pagination := opts.PaginationArgs.SQL()

	var conds []*sqlf.Query
	if pagination.Where != nil {
		conds = append(conds, pagination.Where)
	}

	var joinClause *sqlf.Query
	if opts.Query != nil && *opts.Query != "" {
		conds = append(conds, sqlf.Sprintf("embedding_plugins.name LIKE %s", "%"+*opts.Query+"%"))
		joinClause = sqlf.Sprintf("JOIN embedding_plugins ON embedding_plugins.id = file_embedding_jobs.embedding_plugin_id")
	} else {
		joinClause = sqlf.Sprintf("")
	}

	if opts.State != nil && *opts.State != "" {
		conds = append(conds, sqlf.Sprintf("file_embedding_jobs.state = %s", strings.ToLower(*opts.State)))
	}

	if opts.EmbeddingPlugin != nil {
		conds = append(conds, sqlf.Sprintf("file_embedding_jobs.embedding_plugin_id = %d", *opts.EmbeddingPlugin))
	}

	var whereClause *sqlf.Query
	if len(conds) != 0 {
		whereClause = sqlf.Sprintf("WHERE %s", sqlf.Join(conds, "\n AND "))
	} else {
		whereClause = sqlf.Sprintf("")
	}

	q := sqlf.Sprintf(listFileEmbeddingJobsQueryFmtstr, sqlf.Join(fileEmbeddingJobsColumns, ", "), joinClause, whereClause)
	q = pagination.AppendOrderToQuery(q)
	q = pagination.AppendLimitToQuery(q)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var jobs []*FileEmbeddingJob
	for rows.Next() {
		job, err := scanFileEmbeddingJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func (s *fileEmbeddingJobsStore) CancelFileEmbeddingJob(ctx context.Context, jobID int) error {
	now := time.Now()
	q := sqlf.Sprintf(cancelFileEmbeddingJobQueryFmtstr, now, jobID)

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

const cancelFileEmbeddingJobQueryFmtstr = `
UPDATE
	file_embedding_jobs
SET
    cancel = TRUE,
    -- If the embeddings job is still queued, we directly abort, otherwise we keep the
    -- state, so the worker can do teardown and later mark it failed.
    state = CASE WHEN file_embedding_jobs.state = 'processing' THEN file_embedding_jobs.state ELSE 'canceled' END,
    finished_at = CASE WHEN file_embedding_jobs.state = 'processing' THEN file_embedding_jobs.finished_at ELSE %s END
WHERE
	id = %d
	AND
	state IN ('queued', 'processing')
`
