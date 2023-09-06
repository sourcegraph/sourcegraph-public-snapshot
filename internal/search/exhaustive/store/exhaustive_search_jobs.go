package store

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
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

func (s *Store) CreateExhaustiveSearchJob(ctx context.Context, job types.ExhaustiveSearchJob) (_ int64, err error) {
	ctx, _, endObservation := s.operations.createExhaustiveSearchJob.With(ctx, &err, opAttrs(
		attribute.String("query", job.Query),
		attribute.Int("initiator_id", int(job.InitiatorID)),
	))
	defer endObservation(1, observation.Args{})

	if job.Query == "" {
		return 0, MissingQueryErr
	}
	if job.InitiatorID <= 0 {
		return 0, MissingInitiatorIDErr
	}

	// 🚨 SECURITY: InitiatorID has to match the actor or can be overridden by SiteAdmin.
	if err := auth.CheckSiteAdminOrSameUser(ctx, s.db, job.InitiatorID); err != nil {
		return 0, err
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

func (s *Store) GetExhaustiveSearchJob(ctx context.Context, id int64) (_ *types.ExhaustiveSearchJob, err error) {
	ctx, _, endObservation := s.operations.getExhaustiveSearchJob.With(ctx, &err, opAttrs(
		attribute.Int64("ID", id),
	))
	defer endObservation(1, observation.Args{})

	where := sqlf.Sprintf("id = %d", id)
	q := sqlf.Sprintf(
		getExhaustiveSearchJobQueryFmtStr,
		sqlf.Join(exhaustiveSearchJobColumns, ", "),
		where,
	)

	job, err := scanExhaustiveSearchJob(s.Store.QueryRow(ctx, q))
	if err != nil {
		return nil, err
	}
	if job.ID == 0 {
		return nil, ErrNoResults
	}

	// 🚨 SECURITY: only the initiator, internal or site admins may view a job
	if err := auth.CheckSiteAdminOrSameUser(ctx, s.db, job.InitiatorID); err != nil {
		// job id is just an incrementing integer that on any new job is
		// returned. So this information is not private so we can just return
		// err to indicate the reason for not returning the job.
		return nil, err
	}

	return job, nil
}

const getExhaustiveSearchJobQueryFmtStr = `
SELECT %s FROM exhaustive_search_jobs
WHERE (%s)
LIMIT 1
`

func (s *Store) ListExhaustiveSearchJobs(ctx context.Context) (jobs []*types.ExhaustiveSearchJob, err error) {
	ctx, _, endObservation := s.operations.listExhaustiveSearchJobs.With(ctx, &err, observation.Args{})
	defer func() {
		endObservation(1, opAttrs(attribute.Int("length", len(jobs))))
	}()

	actor := actor.FromContext(ctx)
	if !actor.IsAuthenticated() {
		return nil, errors.New("can only list jobs for an authenticated user")
	}

	q := sqlf.Sprintf(
		listExhaustiveSearchJobsQueryFmtStr,
		sqlf.Join(exhaustiveSearchJobColumns, ", "),
		actor.UID,
	)

	rows, err := s.Store.Query(ctx, q)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		job, err := scanExhaustiveSearchJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}

	return jobs, rows.Err()
}

const listExhaustiveSearchJobsQueryFmtStr = `
SELECT %s FROM exhaustive_search_jobs
WHERE initiator_id = %d
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
