package store

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/types"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var repoSearchJobWorkerOpts = dbworkerstore.Options[*types.ExhaustiveSearchRepoJob]{
	Name:              "exhaustive_search_repo_worker_store",
	TableName:         "exhaustive_search_repo_jobs",
	ColumnExpressions: repoSearchJobColumns,

	Scan: dbworkerstore.BuildWorkerScan(scanRepoSearchJob),

	OrderByExpression: sqlf.Sprintf("exhaustive_search_repo_jobs.state = 'errored', exhaustive_search_repo_jobs.updated_at DESC"),

	StalledMaxAge: 60 * time.Second,
	MaxNumResets:  maxNumResets,

	RetryAfter:    5 * time.Second,
	MaxNumRetries: maxNumRetries,
}

// NewRepoSearchJobWorkerStore returns a dbworkerstore.Store that wraps the "exhaustive_search_repo_jobs" table.
func NewRepoSearchJobWorkerStore(observationCtx *observation.Context, handle basestore.TransactableHandle) dbworkerstore.Store[*types.ExhaustiveSearchRepoJob] {
	return dbworkerstore.New(observationCtx, handle, repoSearchJobWorkerOpts)
}

var repoSearchJobColumns = []*sqlf.Query{
	sqlf.Sprintf("id"),
	sqlf.Sprintf("state"),
	sqlf.Sprintf("repo_id"),
	sqlf.Sprintf("ref_spec"),
	sqlf.Sprintf("search_job_id"),
	sqlf.Sprintf("failure_message"),
	sqlf.Sprintf("started_at"),
	sqlf.Sprintf("finished_at"),
	sqlf.Sprintf("process_after"),
	sqlf.Sprintf("num_resets"),
	sqlf.Sprintf("num_failures"),
	sqlf.Sprintf("execution_logs"),
	sqlf.Sprintf("worker_hostname"),
	sqlf.Sprintf("cancel"),
	sqlf.Sprintf("created_at"),
	sqlf.Sprintf("updated_at"),
}

func (s *Store) CreateExhaustiveSearchRepoJob(ctx context.Context, job types.ExhaustiveSearchRepoJob) (int64, error) {
	var err error
	ctx, _, endObservation := s.operations.createExhaustiveSearchRepoJob.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	if job.SearchJobID <= 0 {
		return 0, MissingSearchJobIDErr
	}
	if job.RepoID <= 0 {
		return 0, MissingRepoIDErr
	}
	if job.RefSpec == "" {
		return 0, MissingRefSpecErr
	}

	row := s.Store.QueryRow(
		ctx,
		sqlf.Sprintf(createExhaustiveSearchRepoJobQueryFmtr, job.RepoID, job.SearchJobID, job.RefSpec),
	)

	var id int64
	if err = row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// MissingSearchJobIDErr is returned when a search job ID is missing.
var MissingSearchJobIDErr = errors.New("missing search job ID")

// MissingRepoIDErr is returned when a repo ID is missing.
var MissingRepoIDErr = errors.New("missing repo ID")

// MissingRefSpecErr is returned when a ref spec is missing.
var MissingRefSpecErr = errors.New("missing ref spec")

const createExhaustiveSearchRepoJobQueryFmtr = `
INSERT INTO exhaustive_search_repo_jobs (repo_id, search_job_id, ref_spec)
VALUES (%s, %s, %s)
RETURNING id
`

func scanRepoSearchJob(sc dbutil.Scanner) (*types.ExhaustiveSearchRepoJob, error) {
	var job types.ExhaustiveSearchRepoJob
	// required field for the sync worker, but
	// the value is thrown out here
	var executionLogs *[]any

	return &job, sc.Scan(
		&job.ID,
		&job.State,
		&job.RepoID,
		&job.RefSpec,
		&job.SearchJobID,
		&dbutil.NullString{S: &job.FailureMessage},
		&dbutil.NullTime{Time: &job.StartedAt},
		&dbutil.NullTime{Time: &job.FinishedAt},
		&dbutil.NullTime{Time: &job.ProcessAfter},
		&job.NumResets,
		&job.NumFailures,
		&executionLogs,
		&job.WorkerHostname,
		&job.Cancel,
		&job.CreatedAt,
		&job.UpdatedAt,
	)
}
