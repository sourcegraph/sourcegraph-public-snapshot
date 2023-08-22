package database

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type RepoUpdateJobStore interface {
	Handle() *basestore.Store
	Create(ctx context.Context, opts CreateRepoUpdateJobOpts) (types.RepoUpdateJob, bool, error)
	List(ctx context.Context, opts ListRepoUpdateJobOpts) ([]types.RepoUpdateJob, error)
	SaveUpdateJobResults(ctx context.Context, jobID int, opts SaveUpdateJobResultsOpts) error
	GetCloningProgress(ctx context.Context, repoName api.RepoName) (string, error)
	// SetCloningProgress updates a piece of text description from how cloning proceeds.
	SetCloningProgress(context.Context, int, string) error
}

type repoUpdateJobStore struct {
	db *basestore.Store
}

func RepoUpdateJobStoreWith(other basestore.ShareableStore) RepoUpdateJobStore {
	return &repoUpdateJobStore{db: basestore.NewWithHandle(other.Handle())}
}

func (s *repoUpdateJobStore) Transact(ctx context.Context) (*repoUpdateJobStore, error) {
	tx, err := s.db.Transact(ctx)
	if err != nil {
		return nil, err
	}

	return &repoUpdateJobStore{
		db: tx,
	}, nil
}

func (s *repoUpdateJobStore) Done(err error) error {
	return s.db.Done(err)
}

func (s *repoUpdateJobStore) Handle() *basestore.Store {
	return s.db
}

type CreateRepoUpdateJobOpts struct {
	RepoName       api.RepoName
	Priority       types.RepoUpdateJobPriority
	ProcessAfter   time.Time
	Clone          bool
	OverwriteClone bool
	FetchRevision  string
}

const createRepoUpdateJobQueryFmtstr = `
INSERT INTO repo_update_jobs(repo_id, priority, process_after, clone, overwrite_clone, fetch_revision)
SELECT r.id, %s, %s, %s, %s, %s
FROM repo r
WHERE r.name = %s
ON CONFLICT DO NOTHING
RETURNING %s
`

func (s *repoUpdateJobStore) Create(ctx context.Context, opts CreateRepoUpdateJobOpts) (types.RepoUpdateJob, bool, error) {
	return scanFirstRepoUpdateJob(s.db.Query(ctx, createRepoUpdateJobQuery(opts)))
}

func createRepoUpdateJobQuery(opts CreateRepoUpdateJobOpts) *sqlf.Query {
	return sqlf.Sprintf(
		createRepoUpdateJobQueryFmtstr,
		opts.Priority,
		dbutil.NullTimeColumn(opts.ProcessAfter),
		opts.Clone,
		opts.OverwriteClone,
		opts.FetchRevision,
		opts.RepoName,
		sqlf.Join(RepoUpdateJobColumns, ", "))
}

// ListRepoUpdateJobOpts contains variables which are used as predicates, for
// ordering and limiting while listing the repo update jobs.
//
// Note: if both repoID and repoName are provided, repoID takes precedence.
type ListRepoUpdateJobOpts struct {
	ID             int
	RepoID         api.RepoID
	RepoName       api.RepoName
	States         []string
	PaginationArgs *PaginationArgs
}

const listRepoUpdateJobsFmtstr = `
SELECT %s
FROM repo_update_jobs
%s -- optional join with repo table
WHERE %s
`

func (s *repoUpdateJobStore) List(ctx context.Context, opts ListRepoUpdateJobOpts) ([]types.RepoUpdateJob, error) {
	return scanRepoUpdateJobs(s.db.Query(ctx, createListRepoUpdateJobsQuery(opts)))
}

func createListRepoUpdateJobsQuery(opts ListRepoUpdateJobOpts) *sqlf.Query {
	preds := []*sqlf.Query{}
	if opts.ID != 0 {
		preds = append(preds, sqlf.Sprintf("id = %s", opts.ID))
	}
	if opts.RepoID != 0 {
		preds = append(preds, sqlf.Sprintf("repo_id = %s", opts.RepoID))
	}
	if len(opts.States) != 0 {
		states := []*sqlf.Query{}
		for _, state := range opts.States {
			states = append(states, sqlf.Sprintf("%s", state))
		}
		preds = append(preds, sqlf.Sprintf("state IN (%s)", sqlf.Join(states, ", ")))
	}

	joinClause := sqlf.Sprintf("")
	if opts.RepoID == 0 && opts.RepoName != "" {
		joinClause = sqlf.Sprintf("JOIN repo ON repo_update_jobs.repo_id = repo.id")
		preds = append(preds, sqlf.Sprintf("repo.name = %s", opts.RepoName))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	query := sqlf.Sprintf(listRepoUpdateJobsFmtstr,
		sqlf.Join(RepoUpdateJobColumns, ", "),
		joinClause,
		sqlf.Join(preds, "AND "))

	if opts.PaginationArgs != nil {
		args := opts.PaginationArgs.SQL()
		query = args.AppendOrderToQuery(query)
		query = args.AppendLimitToQuery(query)
	}
	return query
}

type SaveUpdateJobResultsOpts struct {
	LastFetched           time.Time
	LastChanged           time.Time
	UpdateIntervalSeconds int
}

const saveUpdateJobResultsFmtstr = `
UPDATE repo_update_jobs
SET last_fetched = %s, last_changed = %s, update_interval_seconds = %s
WHERE id = %s
`

func (s *repoUpdateJobStore) SaveUpdateJobResults(ctx context.Context, jobID int, opts SaveUpdateJobResultsOpts) error {
	return s.db.Exec(ctx, sqlf.Sprintf(saveUpdateJobResultsFmtstr,
		dbutil.NullTimeColumn(opts.LastFetched),
		dbutil.NullTimeColumn(opts.LastChanged),
		opts.UpdateIntervalSeconds,
		jobID,
	))
}

const getCloningProgressQueryFmtstr = `
SELECT %s
FROM repo_update_jobs
JOIN repo ON repo_update_jobs.repo_id = repo.id
WHERE repo.name = %s AND repo_update_jobs.state = 'processing'
ORDER BY repo_update_jobs.queued_at DESC
LIMIT 1
`

func (s *repoUpdateJobStore) GetCloningProgress(ctx context.Context, repoName api.RepoName) (progress string, err error) {
	query := sqlf.Sprintf(getCloningProgressQueryFmtstr, sqlf.Join(RepoUpdateJobColumns, ", "), repoName)
	job, ok, err := scanFirstRepoUpdateJob(s.db.Query(ctx, query))
	if err != nil || !ok {
		return
	}
	return job.CloningProgress, err
}

const setCloningProgressQueryFmtstr = `
UPDATE repo_update_jobs
SET cloning_progress = %s
WHERE id = %s
`

func (s *repoUpdateJobStore) SetCloningProgress(ctx context.Context, jobID int, progressLine string) error {
	res, err := s.db.ExecResult(ctx, sqlf.Sprintf(setCloningProgressQueryFmtstr, progressLine, jobID))
	if err != nil {
		return err
	}
	if nrows, err := res.RowsAffected(); err != nil {
		return err
	} else if nrows != 1 {
		return errors.Newf("failed to set repo update job cloning progress, job with ID=%d not found", jobID)
	}
	return nil
}

// RepoUpdateJobColumns is a set of columns which are in the `repo_update_jobs`
// table.
var RepoUpdateJobColumns = []*sqlf.Query{
	// Regular worker columns.
	sqlf.Sprintf("repo_update_jobs.id"),
	sqlf.Sprintf("repo_update_jobs.state"),
	sqlf.Sprintf("repo_update_jobs.failure_message"),
	sqlf.Sprintf("repo_update_jobs.queued_at"),
	sqlf.Sprintf("repo_update_jobs.started_at"),
	sqlf.Sprintf("repo_update_jobs.finished_at"),
	sqlf.Sprintf("repo_update_jobs.process_after"),
	sqlf.Sprintf("repo_update_jobs.num_resets"),
	sqlf.Sprintf("repo_update_jobs.num_failures"),
	sqlf.Sprintf("repo_update_jobs.last_heartbeat_at"),
	sqlf.Sprintf("repo_update_jobs.execution_logs"),
	sqlf.Sprintf("repo_update_jobs.worker_hostname"),
	sqlf.Sprintf("repo_update_jobs.cancel"),
	// These 6 columns are in both `repo_update_jobs` table and
	// `repo_update_jobs_with_repo_name` view.
	sqlf.Sprintf("repo_update_jobs.repo_id"),
	sqlf.Sprintf("repo_update_jobs.priority"),
	sqlf.Sprintf("repo_update_jobs.clone"),
	sqlf.Sprintf("repo_update_jobs.overwrite_clone"),
	sqlf.Sprintf("repo_update_jobs.fetch_revision"),
	sqlf.Sprintf("repo_update_jobs.last_fetched"),
	sqlf.Sprintf("repo_update_jobs.last_changed"),
	sqlf.Sprintf("repo_update_jobs.update_interval_seconds"),
	sqlf.Sprintf("repo_update_jobs.cloning_progress"),
}

// FullRepoUpdateJobColumns is a set of columns of `repo_update_jobs_with_repo_name` view.
var FullRepoUpdateJobColumns = append(RepoUpdateJobColumns,
	sqlf.Sprintf("repo_update_jobs.repository_name"),
	sqlf.Sprintf("repo_update_jobs.pool_repo_id"),
)

var scanFirstRepoUpdateJob = basestore.NewFirstScanner(ScanRepoUpdateJob)
var scanRepoUpdateJobs = basestore.NewSliceScanner(ScanRepoUpdateJob)

func ScanRepoUpdateJob(s dbutil.Scanner) (job types.RepoUpdateJob, _ error) {
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
		&dbutil.NullTime{Time: &job.LastHeartbeatAt},
		pq.Array(&executionLogs),
		&job.WorkerHostname,
		&job.Cancel,
		&job.RepoID,
		&job.Priority,
		&job.Clone,
		&job.OverwriteClone,
		&job.FetchRevision,
		&dbutil.NullTime{Time: &job.LastFetched},
		&dbutil.NullTime{Time: &job.LastChanged},
		&dbutil.NullInt{N: &job.UpdateIntervalSeconds},
		&dbutil.NullString{S: &job.CloningProgress},
	); err != nil {
		return job, err
	}
	job.ExecutionLogs = append(job.ExecutionLogs, executionLogs...)
	return job, nil
}

func ScanFullRepoUpdateJob(s dbutil.Scanner) (job types.RepoUpdateJob, _ error) {
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
		&dbutil.NullTime{Time: &job.LastHeartbeatAt},
		pq.Array(&executionLogs),
		&job.WorkerHostname,
		&job.Cancel,
		&job.RepoID,
		&job.Priority,
		&job.Clone,
		&job.OverwriteClone,
		&job.FetchRevision,
		&dbutil.NullTime{Time: &job.LastFetched},
		&dbutil.NullTime{Time: &job.LastChanged},
		&dbutil.NullInt{N: &job.UpdateIntervalSeconds},
		&dbutil.NullString{S: &job.CloningProgress},
		&job.RepositoryName,
		&job.PoolRepoID,
	); err != nil {
		return job, err
	}
	job.ExecutionLogs = append(job.ExecutionLogs, executionLogs...)
	return job, nil
}
