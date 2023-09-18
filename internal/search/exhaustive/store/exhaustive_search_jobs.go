package store

import (
	"context"
	"strings"
	"time"

	"github.com/keegancsmith/sqlf"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
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

	return basestore.ScanAny[int64](s.Store.QueryRow(
		ctx,
		sqlf.Sprintf(createExhaustiveSearchJobQueryFmtr, job.Query, job.InitiatorID),
	))
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

func (s *Store) CancelSearchJob(ctx context.Context, id int64) (totalCanceled int, err error) {
	ctx, _, endObservation := s.operations.cancelSearchJob.With(ctx, &err, opAttrs(
		attribute.Int64("ID", id),
	))
	defer endObservation(1, observation.Args{})

	// 🚨 SECURITY: only someone with access to the job may cancel the job
	_, err = s.GetExhaustiveSearchJob(ctx, id)
	if err != nil {
		return -1, err
	}

	now := time.Now()
	q := sqlf.Sprintf(cancelJobFmtStr, now, id, now, now)

	row := s.QueryRow(ctx, q)

	err = row.Scan(&totalCanceled)
	if err != nil {
		return -1, err
	}

	return totalCanceled, nil
}

const cancelJobFmtStr = `
WITH updated_jobs AS (
    -- Update the state of the main job
    UPDATE exhaustive_search_jobs
    SET CANCEL = TRUE,
    -- If the embeddings job is still queued, we directly abort, otherwise we keep the
    -- state, so the worker can do teardown and later mark it failed.
    state = CASE WHEN exhaustive_search_jobs.state = 'processing' THEN exhaustive_search_jobs.state ELSE 'canceled' END,
    finished_at = CASE WHEN exhaustive_search_jobs.state = 'processing' THEN exhaustive_search_jobs.finished_at ELSE %s END
    WHERE id = %s
    RETURNING id
),
updated_repo_jobs AS (
    -- Update the state of the dependent repo_jobs
    UPDATE exhaustive_search_repo_jobs
    SET CANCEL = TRUE,
    -- If the embeddings job is still queued, we directly abort, otherwise we keep the
    -- state, so the worker can do teardown and later mark it failed.
    state = CASE WHEN exhaustive_search_repo_jobs.state = 'processing' THEN exhaustive_search_repo_jobs.state ELSE 'canceled' END,
    finished_at = CASE WHEN exhaustive_search_repo_jobs.state = 'processing' THEN exhaustive_search_repo_jobs.finished_at ELSE %s END
    WHERE search_job_id IN (SELECT id FROM updated_jobs)
    RETURNING id
),
updated_repo_revision_jobs AS (
    -- Update the state of the dependent repo_revision_jobs
    UPDATE exhaustive_search_repo_revision_jobs
    SET CANCEL = TRUE,
	-- If the embeddings job is still queued, we directly abort, otherwise we keep the
	-- state, so the worker can do teardown and later mark it failed.
    state = CASE WHEN exhaustive_search_repo_revision_jobs.state = 'processing' THEN exhaustive_search_repo_revision_jobs.state ELSE 'canceled' END,
    finished_at = CASE WHEN exhaustive_search_repo_revision_jobs.state = 'processing' THEN exhaustive_search_repo_revision_jobs.finished_at ELSE %s END
    WHERE search_repo_job_id IN (SELECT id FROM updated_repo_jobs)
    RETURNING id
)
SELECT (SELECT count(*) FROM updated_jobs) + (SELECT count(*) FROM updated_repo_jobs) + (SELECT count(*) FROM updated_repo_revision_jobs) as total_canceled
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

type ListArgs struct {
	*database.PaginationArgs
	Query   string
	States  []string
	UserIDs []int32
}

func (s *Store) ListExhaustiveSearchJobs(ctx context.Context, args ListArgs) (jobs []*types.ExhaustiveSearchJob, err error) {
	ctx, _, endObservation := s.operations.listExhaustiveSearchJobs.With(ctx, &err, observation.Args{})
	defer func() {
		endObservation(1, opAttrs(attribute.Int("length", len(jobs))))
	}()

	a := actor.FromContext(ctx)

	// 🚨 SECURITY: Only authenticated users can list search jobs.
	if !a.IsAuthenticated() {
		return nil, errors.New("can only list jobs for an authenticated user")
	}

	var conds []*sqlf.Query

	// Filter by query.
	if args.Query != "" {
		conds = append(conds, sqlf.Sprintf("query LIKE %s", "%"+args.Query+"%"))
	}

	// Filter by state.
	if len(args.States) > 0 {
		states := make([]*sqlf.Query, len(args.States))
		for i, state := range args.States {
			states[i] = sqlf.Sprintf("%s", strings.ToLower(state))
		}
		conds = append(conds, sqlf.Sprintf("state in (%s)", sqlf.Join(states, ",")))
	}

	// 🚨 SECURITY: Site admins see any job and may filter based on args.UserIDs.
	// Other users only see their own jobs.
	isSiteAdmin := auth.CheckUserIsSiteAdmin(ctx, s.db, a.UID) == nil
	if isSiteAdmin {
		if len(args.UserIDs) > 0 {
			ids := make([]*sqlf.Query, len(args.UserIDs))
			for i, id := range args.UserIDs {
				ids[i] = sqlf.Sprintf("%d", id)
			}
			conds = append(conds, sqlf.Sprintf("initiator_id in (%s)", sqlf.Join(ids, ",")))
		}
	} else {
		if len(args.UserIDs) > 0 {
			return nil, errors.New("cannot filter by user id if not a site admin")
		}
		conds = append(conds, sqlf.Sprintf("initiator_id = %d", a.UID))
	}

	var pagination *database.QueryArgs
	if args.PaginationArgs != nil {
		pagination = args.PaginationArgs.SQL()
		if pagination.Where != nil {
			conds = append(conds, pagination.Where)
		}
	}

	var whereClause *sqlf.Query
	if len(conds) != 0 {
		whereClause = sqlf.Sprintf("WHERE %s", sqlf.Join(conds, "\n AND "))
	} else {
		whereClause = sqlf.Sprintf("")
	}

	q := sqlf.Sprintf(
		listExhaustiveSearchJobsQueryFmtStr,
		sqlf.Join(exhaustiveSearchJobColumns, ", "),
		whereClause,
	)

	if pagination != nil {
		q = pagination.AppendOrderToQuery(q)
		q = pagination.AppendLimitToQuery(q)
	}

	return scanExhaustiveSearchJobs(s.Store.Query(ctx, q))
}

const listExhaustiveSearchJobsQueryFmtStr = `
SELECT %s FROM exhaustive_search_jobs
%s -- whereClause
`

const deleteExhaustiveSearchJobQueryFmtStr = `
DELETE FROM exhaustive_search_jobs
WHERE id = %d
`

func (s *Store) DeleteExhaustiveSearchJob(ctx context.Context, id int64) (err error) {
	ctx, _, endObservation := s.operations.deleteExhaustiveSearchJob.With(ctx, &err, opAttrs(
		attribute.Int64("ID", id),
	))
	defer endObservation(1, observation.Args{})

	// 🚨 SECURITY: only someone with access to the job may delete the job
	_, err = s.GetExhaustiveSearchJob(ctx, id)
	if err != nil {
		return err
	}

	return s.Exec(ctx, sqlf.Sprintf(deleteExhaustiveSearchJobQueryFmtStr, id))
}

const getAggregateRepoRevState = `
SELECT rrj.state, count(rrj.id) as count FROM exhaustive_search_repo_revision_jobs rrj
JOIN exhaustive_search_repo_jobs rj ON rrj.search_repo_job_id = rj.id
WHERE rj.search_job_id = %s
GROUP BY rrj.state
`

func (s *Store) GetAggregateRepoRevState(ctx context.Context, id int64) (_ map[string]int, err error) {
	ctx, _, endObservation := s.operations.getAggregateRepoRevState.With(ctx, &err, opAttrs(
		attribute.Int64("ID", id),
	))
	defer endObservation(1, observation.Args{})

	// 🚨 SECURITY: only someone with access to the job may cancel the job
	_, err = s.GetExhaustiveSearchJob(ctx, id)
	if err != nil {
		return nil, err
	}

	q := sqlf.Sprintf(getAggregateRepoRevState, id)

	rows, err := s.Store.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := make(map[string]int)
	for rows.Next() {
		var state string
		var count int
		if err := rows.Scan(&state, &count); err != nil {
			return nil, err
		}

		m[state] = count
	}

	return m, nil
}

const getJobLogsFmtStr = `
SELECT
rjj.id,
rj.repo_id,
rjj.revision,
rjj.state,
rjj.failure_message,
rjj.started_at,
rjj.finished_at
FROM exhaustive_search_repo_revision_jobs rjj
JOIN exhaustive_search_repo_jobs rj ON rjj.search_repo_job_id = rj.id
%s
`

type GetJobLogsOpts struct {
	From  int64
	Limit int
}

func (s *Store) GetJobLogs(ctx context.Context, id int64, opts *GetJobLogsOpts) ([]types.SearchJobLog, error) {
	// 🚨 SECURITY: only someone with access to the job may access the logs
	_, err := s.GetExhaustiveSearchJob(ctx, id)
	if err != nil {
		return nil, err
	}

	conds := []*sqlf.Query{sqlf.Sprintf("rj.search_job_id = %s", id)}
	var limit *sqlf.Query
	if opts != nil {
		if opts.From != 0 {
			conds = append(conds, sqlf.Sprintf("rjj.id >= %s", opts.From))
		}

		if opts.Limit != 0 {
			limit = sqlf.Sprintf("LIMIT %s", opts.Limit)
		}
	}

	q := sqlf.Sprintf(
		getJobLogsFmtStr,
		sqlf.Sprintf("WHERE %s ORDER BY id ASC", sqlf.Join(conds, "AND")),
	)
	if limit != nil {
		q = sqlf.Sprintf("%v %v", q, limit)
	}

	rows, err := s.Store.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []types.SearchJobLog
	for rows.Next() {
		job := types.SearchJobLog{}
		if err := rows.Scan(
			&job.ID,
			&job.RepoID,
			&job.Revision,
			&job.State,
			&dbutil.NullString{S: &job.FailureMessage},
			&dbutil.NullTime{Time: &job.StartedAt},
			&dbutil.NullTime{Time: &job.FinishedAt},
		); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

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

var scanExhaustiveSearchJobs = basestore.NewSliceScanner(scanExhaustiveSearchJob)
