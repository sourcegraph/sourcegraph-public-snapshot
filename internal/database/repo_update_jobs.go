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
)

type RepoUpdateJobStore interface {
	WithTransaction(ctx context.Context, f func(s RepoUpdateJobStore) error) error
	Handle() *basestore.Store
	Create(ctx context.Context, opts RepoUpdateJobOpts) (types.RepoUpdateJob, bool, error)
	SaveUpdateJobResults(ctx context.Context, jobID int, opts UpdateRepoUpdateJobOpts) error
	List(ctx context.Context) ([]*types.RepoUpdateJob, error)
}

type repoUpdateJobStore struct {
	db *basestore.Store
}

func RepoUpdateJobStoreWith(other basestore.ShareableStore) RepoUpdateJobStore {
	return &repoUpdateJobStore{db: basestore.NewWithHandle(other.Handle())}
}

func (s *repoUpdateJobStore) WithTransaction(ctx context.Context, f func(s RepoUpdateJobStore) error) error {
	return s.withTransaction(ctx, func(s *repoUpdateJobStore) error { return f(s) })
}

func (s *repoUpdateJobStore) withTransaction(ctx context.Context, f func(s *repoUpdateJobStore) error) error {
	return basestore.InTransaction[*repoUpdateJobStore](ctx, s, f)
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

func (s *repoUpdateJobStore) List(ctx context.Context) ([]*types.RepoUpdateJob, error) {
	// TODO implement me
	panic("implement me")
}

type RepoUpdateJobOpts struct {
	RepoID       api.RepoID
	Priority     int // TODO(sasha): make it "enum"
	ProcessAfter time.Time
}

type UpdateRepoUpdateJobOpts struct {
	LastFetched           time.Time
	LastChanged           time.Time
	UpdateIntervalSeconds int
}

const createRepoCloneJobQueryFmtstr = `
INSERT INTO repo_update_jobs(repo_id, priority, process_after)
VALUES (%s, %s, %s)
ON CONFLICT DO NOTHING
RETURNING %s
`

func (s *repoUpdateJobStore) Create(ctx context.Context, opts RepoUpdateJobOpts) (types.RepoUpdateJob, bool, error) {
	return scanFirstRepoUpdateJob(s.db.Query(ctx, createRepoCloneJobQuery(opts)))
}

const saveUpdateJobResultsFmtstr = `
UPDATE repo_update_jobs
SET last_fetched = %s, last_changed = %s, update_interval_seconds = %s
WHERE id = %s
`

func (s *repoUpdateJobStore) SaveUpdateJobResults(ctx context.Context, jobID int, opts UpdateRepoUpdateJobOpts) error {
	return s.db.Exec(ctx, sqlf.Sprintf(saveUpdateJobResultsFmtstr,
		dbutil.NullTimeColumn(opts.LastFetched),
		dbutil.NullTimeColumn(opts.LastFetched),
		opts.UpdateIntervalSeconds,
		jobID,
	))
}

func createRepoCloneJobQuery(opts RepoUpdateJobOpts) *sqlf.Query {
	return sqlf.Sprintf(
		createRepoCloneJobQueryFmtstr,
		opts.RepoID,
		opts.Priority,
		dbutil.NullTimeColumn(opts.ProcessAfter),
		sqlf.Join(RepoCloneJobColumns, ", "))
}

// FullRepoCloneJobColumns contains columns in the
// `repo_update_jobs_with_repo_name` view.
var FullRepoCloneJobColumns = []*sqlf.Query{
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
	// These 5 columns are in both `repo_update_jobs` table and
	// `repo_update_jobs_with_repo_name` view.
	sqlf.Sprintf("repo_update_jobs.repo_id"),
	sqlf.Sprintf("repo_update_jobs.priority"),
	sqlf.Sprintf("repo_update_jobs.last_fetched"),
	sqlf.Sprintf("repo_update_jobs.last_changed"),
	sqlf.Sprintf("repo_update_jobs.update_interval_seconds"),
	// These 2 columns are only in the `repo_update_jobs_with_repo_name` view.
	sqlf.Sprintf("repo_update_jobs.repository_name"),
	sqlf.Sprintf("repo_update_jobs.pool_repo_id"),
}

// RepoCloneJobColumns is a subset of columns which are in the
// `repo_update_jobs` table.
var RepoCloneJobColumns = FullRepoCloneJobColumns[:18]

var scanFirstRepoUpdateJob = basestore.NewFirstScanner(ScanRepoUpdateJob)

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
	); err != nil {
		return job, err
	}
	job.ExecutionLogs = append(job.ExecutionLogs, executionLogs...)
	return job, nil
}

var scanFirstFullRepoUpdateJob = basestore.NewFirstScanner(ScanFullRepoUpdateJob)

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
		&dbutil.NullTime{Time: &job.LastFetched},
		&dbutil.NullTime{Time: &job.LastChanged},
		&dbutil.NullInt{N: &job.UpdateIntervalSeconds},
		&job.RepositoryName,
		&dbutil.NullInt32{N: job.PoolRepoID},
	); err != nil {
		return job, err
	}
	job.ExecutionLogs = append(job.ExecutionLogs, executionLogs...)
	return job, nil
}
