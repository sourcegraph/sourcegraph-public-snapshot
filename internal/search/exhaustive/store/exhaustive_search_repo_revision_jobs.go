package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/types"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var revSearchJobWorkerOpts = dbworkerstore.Options[*types.ExhaustiveSearchRepoRevisionJob]{
	Name:              "exhaustive_search_repo_revision_worker_store",
	TableName:         "exhaustive_search_repo_revision_jobs",
	ColumnExpressions: revSearchJobColumns,

	Scan: dbworkerstore.BuildWorkerScan(scanRevSearchJob),

	OrderByExpression: sqlf.Sprintf("exhaustive_search_repo_revision_jobs.state = 'errored', exhaustive_search_repo_revision_jobs.updated_at DESC"),

	StalledMaxAge: 60 * time.Second,
	MaxNumResets:  0,

	RetryAfter:    5 * time.Second,
	MaxNumRetries: 0,
}

// NewRevSearchJobWorkerStore returns a dbworkerstore.Store that wraps the "exhaustive_search_repo_revision_jobs" table.
func NewRevSearchJobWorkerStore(observationCtx *observation.Context, handle basestore.TransactableHandle) dbworkerstore.Store[*types.ExhaustiveSearchRepoRevisionJob] {
	return dbworkerstore.New(observationCtx, handle, revSearchJobWorkerOpts)
}

var revSearchJobColumns = []*sqlf.Query{
	sqlf.Sprintf("id"),
	sqlf.Sprintf("state"),
	sqlf.Sprintf("search_repo_job_id"),
	sqlf.Sprintf("revision"),
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

// ExhaustiveSearchRepoRevisionJobStore is the interface for interacting with "exhaustive_search_repo_revision_jobs".
type ExhaustiveSearchRepoRevisionJobStore interface {
	// CreateExhaustiveSearchRepoRevisionJob creates a new types.ExhaustiveSearchRepoRevisionJob.
	CreateExhaustiveSearchRepoRevisionJob(ctx context.Context, job types.ExhaustiveSearchRepoRevisionJob) (int64, error)
	// ListExhaustiveSearchRepoRevisionJobs lists types.ExhaustiveSearchRepoRevisionJob matching the given options.
	ListExhaustiveSearchRepoRevisionJobs(ctx context.Context, opts ListExhaustiveSearchRepoRevisionJobsOpts) ([]*types.ExhaustiveSearchRepoRevisionJob, error)
}

// ListExhaustiveSearchRepoRevisionJobsOpts captures the query options needed for listing exhaustive search repo revision jobs.
type ListExhaustiveSearchRepoRevisionJobsOpts struct {
	// First, if set, limits the number of actions returned
	// to the first n.
	First *int
	// After, if set, begins listing actions after the given id.
	After *int
	// SearchRepoJobID will constrain the listed actions to only.
	SearchRepoJobID int64
}

// Conds returns the SQL conditions for the list options.
func (o ListExhaustiveSearchRepoRevisionJobsOpts) Conds() *sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("search_repo_job_id = %s", o.SearchRepoJobID)}
	if o.After != nil {
		conds = append(conds, sqlf.Sprintf("id > %s", *o.After))
	}
	return sqlf.Join(conds, "AND")
}

// Limit returns the SQL limit for the list options.
func (o ListExhaustiveSearchRepoRevisionJobsOpts) Limit() *sqlf.Query {
	if o.First == nil {
		return sqlf.Sprintf("ALL")
	}
	return sqlf.Sprintf("%s", *o.First)
}

var _ ExhaustiveSearchRepoJobStore = &Store{}

func (s *Store) CreateExhaustiveSearchRepoRevisionJob(ctx context.Context, job types.ExhaustiveSearchRepoRevisionJob) (int64, error) {
	var err error
	ctx, _, endObservation := s.operations.createExhaustiveSearchRepoJob.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	if job.SearchRepoJobID <= 0 {
		return 0, MissingSearchRepoJobIDErr
	}
	if job.Revision == "" {
		return 0, MissingRevisionErr
	}

	row := s.Store.QueryRow(
		ctx,
		sqlf.Sprintf(createExhaustiveSearchRepoRevisionJobQueryFmtr, job.Revision, job.SearchRepoJobID),
	)

	var id int64
	if err = row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// MissingSearchRepoJobIDErr is returned when a search repo job ID is missing.
var MissingSearchRepoJobIDErr = errors.New("missing search repo job ID")

// MissingRevisionErr is returned when a revision is missing.
var MissingRevisionErr = errors.New("missing revision")

const createExhaustiveSearchRepoRevisionJobQueryFmtr = `
INSERT INTO exhaustive_search_repo_revision_jobs (revision, search_repo_job_id)
VALUES (%s, %s)
RETURNING id
`

func (s *Store) ListExhaustiveSearchRepoRevisionJobs(ctx context.Context, opts ListExhaustiveSearchRepoRevisionJobsOpts) ([]*types.ExhaustiveSearchRepoRevisionJob, error) {
	var jobs []*types.ExhaustiveSearchRepoRevisionJob
	var err error
	ctx, _, endObservation := s.operations.listExhaustiveSearchRepoRevisionJob.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("ID", int(opts.SearchRepoJobID)),
	}})
	defer endObservation(1, observation.Args{})

	if opts.SearchRepoJobID <= 0 {
		return nil, MissingSearchRepoJobIDErr
	}

	q := sqlf.Sprintf(
		listExhaustiveSearchRepoRevisionJobsQueryFmtr,
		sqlf.Join(revSearchJobColumns, ","),
		opts.Conds(),
		opts.Limit(),
	)
	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	jobs, err = scanRevSearchJobs(rows)
	return jobs, err
}

const listExhaustiveSearchRepoRevisionJobsQueryFmtr = `
SELECT %s FROM exhaustive_search_repo_revision_jobs
WHERE %s
ORDER BY id ASC
LIMIT %s
`

func scanRevSearchJobs(rows *sql.Rows) ([]*types.ExhaustiveSearchRepoRevisionJob, error) {
	var jobs []*types.ExhaustiveSearchRepoRevisionJob
	for rows.Next() {
		job, err := scanRevSearchJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, rows.Err()
}

func scanRevSearchJob(sc dbutil.Scanner) (*types.ExhaustiveSearchRepoRevisionJob, error) {
	var job types.ExhaustiveSearchRepoRevisionJob
	// required field for the sync worker, but
	// the value is thrown out here
	var executionLogs *[]any

	return &job, sc.Scan(
		&job.ID,
		&job.State,
		&job.SearchRepoJobID,
		&job.Revision,
		&dbutil.NullString{S: job.FailureMessage},
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
