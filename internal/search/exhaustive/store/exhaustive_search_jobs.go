package store

import (
	"context"
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

var exhaustiveSearchJobWorkerOpts = dbworkerstore.Options[*types.ExhaustiveSearchJob]{
	Name:              "exhaustive_search_worker_store",
	TableName:         "exhaustive_search_jobs",
	ColumnExpressions: exhaustiveSearchJobColumns,

	Scan: dbworkerstore.BuildWorkerScan(scanExhaustiveSearchJob),

	OrderByExpression: sqlf.Sprintf("exhaustive_search_jobs.state = 'errored', exhaustive_search_jobs.updated_at DESC"),

	StalledMaxAge: 60 * time.Second,
	MaxNumResets:  0,

	RetryAfter:    5 * time.Second,
	MaxNumRetries: 0,
}

// NewExhaustiveSearchJobWorkerStore returns a dbworkerstore.Store that wraps the "exhaustive_search_jobs" table.
func NewExhaustiveSearchJobWorkerStore(observationCtx *observation.Context, handle basestore.TransactableHandle) dbworkerstore.Store[*types.ExhaustiveSearchJob] {
	return dbworkerstore.New(observationCtx, handle, exhaustiveSearchJobWorkerOpts)
}

var exhaustiveSearchJobColumns = []*sqlf.Query{
	sqlf.Sprintf("id"),
	sqlf.Sprintf("initiator_id"),
	sqlf.Sprintf("state"),
	sqlf.Sprintf("query"),
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

// ExhaustiveSearchJobStore is the interface for interacting with "exhaustive_search_jobs".
type ExhaustiveSearchJobStore interface {
	// CreateExhaustiveSearchJob creates a new types.ExhaustiveSearchJob.
	CreateExhaustiveSearchJob(ctx context.Context, job types.ExhaustiveSearchJob) (int64, error)
	// GetExhaustiveSearchJobByID returns the types.ExhaustiveSearchJob with the given ID.
	GetExhaustiveSearchJobByID(ctx context.Context, id int64) (*types.ExhaustiveSearchJob, error)
}

var _ ExhaustiveSearchJobStore = &Store{}

func (s *Store) CreateExhaustiveSearchJob(ctx context.Context, job types.ExhaustiveSearchJob) (int64, error) {
	var err error
	ctx, _, endObservation := s.operations.createExhaustiveSearchJob.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	if job.Query == "" {
		return 0, MissingQueryErr
	}
	if job.InitiatorID <= 0 {
		return 0, MissingInitiatorIDErr
	}

	row := s.Store.QueryRow(
		ctx,
		sqlf.Sprintf(createExhaustiveSearchJobQueryFmtr, job.Query, job.InitiatorID),
	)

	var id int64
	if err = row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// MissingQueryErr is returned when a query is missing from a types.ExhaustiveSearchJob.
var MissingQueryErr = errors.New("missing query")

// MissingInitiatorIDErr is returned when an initiator ID is missing from a types.ExhaustiveSearchJob.
var MissingInitiatorIDErr = errors.New("missing initiator ID")

const createExhaustiveSearchJobQueryFmtr = `
INSERT INTO exhaustive_search_jobs (query, initiator_id)
VALUES (%s, %s)
RETURNING id
`

func (s *Store) GetExhaustiveSearchJobByID(ctx context.Context, id int64) (*types.ExhaustiveSearchJob, error) {
	var job *types.ExhaustiveSearchJob
	var err error
	ctx, _, endObservation := s.operations.getExhaustiveSearchJobByID.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("ID", int(id)),
	}})
	defer endObservation(1, observation.Args{})

	if id <= 0 {
		return job, InvalidIDErr
	}

	row := s.Store.QueryRow(
		ctx,
		sqlf.Sprintf(
			getExhaustiveSearchJobByIDQueryFmtr,
			sqlf.Join(exhaustiveSearchJobColumns, ", "),
			id,
		),
	)
	job, err = scanExhaustiveSearchJob(row)
	if err != nil {
		return job, err
	}
	return job, nil
}

// InvalidIDErr is returned when an invalid ID is passed.
var InvalidIDErr = errors.New("invalid id")

const getExhaustiveSearchJobByIDQueryFmtr = `
SELECT %s FROM exhaustive_search_jobs
WHERE id = %s
`

func scanExhaustiveSearchJob(sc dbutil.Scanner) (*types.ExhaustiveSearchJob, error) {
	var job types.ExhaustiveSearchJob
	// required field for the sync worker, but
	// the value is thrown out here
	var executionLogs *[]any

	return &job, sc.Scan(
		&job.ID,
		&job.InitiatorID,
		&job.State,
		&job.Query,
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
