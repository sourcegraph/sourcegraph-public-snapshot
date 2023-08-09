package database

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type RepoCloneJobStore interface {
	basestore.ShareableStore
	With(other basestore.ShareableStore) PermissionSyncJobStore
	Transact(ctx context.Context) (PermissionSyncJobStore, error)
	Done(err error) error
	Create(ctx context.Context, opts RepoCloneJobOpts) error
	List(ctx context.Context) ([]*types.RepoCloneJob, error)
}

type repoCloneJobStore struct {
	*basestore.Store
}

func RepoCloneJobStoreWith(other basestore.ShareableStore) RepoCloneJobStore {
	return &repoCloneJobStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (s *repoCloneJobStore) With(other basestore.ShareableStore) PermissionSyncJobStore {
	// TODO implement me
	panic("implement me")
}

func (s *repoCloneJobStore) Transact(ctx context.Context) (PermissionSyncJobStore, error) {
	// TODO implement me
	panic("implement me")
}

func (s *repoCloneJobStore) List(ctx context.Context) ([]*types.RepoCloneJob, error) {
	// TODO implement me
	panic("implement me")
}

type RepoCloneJobOpts struct {
	GitserverAddress string
	UpdateAfter      int
	RepoName         api.RepoName
	Clone            bool
}

const createRepoCloneJobQueryFmtstr = `
INSERT INTO repo_clone_jobs(gitserver_address, repo_name, update_after, clone)
VALUES (%s, %s, %s, %s)
ON CONFLICT DO NOTHING
RETURNING %s
`

func (s *repoCloneJobStore) Create(ctx context.Context, opts RepoCloneJobOpts) error {
	job := &types.RepoCloneJob{
		GitserverAddress: opts.GitserverAddress,
		UpdateAfter:      opts.UpdateAfter,
		RepoName:         string(opts.RepoName),
		Clone:            opts.Clone,
	}
	query := createRepoCloneJobQuery(job)
	row := s.QueryRow(ctx, query)
	return scanRepoCloneJob(row, job)
}

func createRepoCloneJobQuery(job *types.RepoCloneJob) *sqlf.Query {
	return sqlf.Sprintf(
		createRepoCloneJobQueryFmtstr,
		job.GitserverAddress,
		job.RepoName,
		job.UpdateAfter,
		job.Clone,
		sqlf.Join(RepoCloneJobColumns, ", "),
	)
}

var RepoCloneJobColumns = []*sqlf.Query{
	sqlf.Sprintf("repo_clone_jobs.id"),
	sqlf.Sprintf("repo_clone_jobs.state"),
	sqlf.Sprintf("repo_clone_jobs.failure_message"),
	sqlf.Sprintf("repo_clone_jobs.queued_at"),
	sqlf.Sprintf("repo_clone_jobs.started_at"),
	sqlf.Sprintf("repo_clone_jobs.finished_at"),
	sqlf.Sprintf("repo_clone_jobs.process_after"),
	sqlf.Sprintf("repo_clone_jobs.num_resets"),
	sqlf.Sprintf("repo_clone_jobs.num_failures"),
	sqlf.Sprintf("repo_clone_jobs.last_heartbeat_at"),
	sqlf.Sprintf("repo_clone_jobs.execution_logs"),
	sqlf.Sprintf("repo_clone_jobs.worker_hostname"),
	sqlf.Sprintf("repo_clone_jobs.cancel"),

	sqlf.Sprintf("repo_clone_jobs.gitserver_address"),
	sqlf.Sprintf("repo_clone_jobs.update_after"),
	sqlf.Sprintf("repo_clone_jobs.repo_name"),
	sqlf.Sprintf("repo_clone_jobs.clone"),
}

func ScanRepoCloneJob(s dbutil.Scanner) (*types.RepoCloneJob, error) {
	var job types.RepoCloneJob
	if err := scanRepoCloneJob(s, &job); err != nil {
		return nil, err
	}
	return &job, nil
}

func scanRepoCloneJob(s dbutil.Scanner, job *types.RepoCloneJob) error {
	var executionLogs []executor.ExecutionLogEntry
	return s.Scan(
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
		&job.GitserverAddress,
		&job.UpdateAfter,
		&job.RepoName,
		&job.Clone,
	)
}
