package store

import (
	"context"
	"database/sql"
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

const maxNumResets = 5
const maxNumRetries = 3

var exhaustiveSearchJobWorkerOpts = dbworkerstore.Options[*types.ExhaustiveSearchJob]{
	Name:              "exhaustive_search_worker_store",
	TableName:         "exhaustive_search_jobs",
	ColumnExpressions: exhaustiveSearchJobColumns,

	Scan: dbworkerstore.BuildWorkerScan(scanExhaustiveSearchJob),

	OrderByExpression: sqlf.Sprintf("exhaustive_search_jobs.state = 'errored', exhaustive_search_jobs.updated_at DESC"),

	StalledMaxAge: 60 * time.Second,
	MaxNumResets:  maxNumResets,

	RetryAfter:    5 * time.Second,
	MaxNumRetries: maxNumRetries,
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
	err = s.UserHasAccess(ctx, id)
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

func listSearchJobQuery(where *sqlf.Query) *sqlf.Query {
	return sqlf.Sprintf(
		listExhaustiveSearchJobsQueryFmtStr,
		sqlf.Join(exhaustiveSearchJobColumns, ", "),
		sqlf.Sprintf(
			aggStateSubQuery,
			sqlf.Sprintf(
				getAggregateStateTable,
				sqlf.Sprintf("exhaustive_search_jobs.id"),
				sqlf.Sprintf("exhaustive_search_jobs.id"),
				sqlf.Sprintf("exhaustive_search_jobs.id"),
			),
		),
		where,
	)
}

func (s *Store) GetExhaustiveSearchJob(ctx context.Context, id int64) (_ *types.ExhaustiveSearchJob, err error) {
	ctx, _, endObservation := s.operations.getExhaustiveSearchJob.With(ctx, &err, opAttrs(
		attribute.Int64("ID", id),
	))
	defer endObservation(1, observation.Args{})

	where := sqlf.Sprintf("WHERE id = %d", id)
	q := listSearchJobQuery(where)

	job, err := scanExhaustiveSearchJobList(s.Store.QueryRow(ctx, q))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Wrapf(ErrNoResults, "failed to scan job with id %d: %s", id, err.Error())
		}
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

// UserHasAccess is a helper function to check if the user has access to the
// job. It returns an error if the job cannot be found or the user is not
// authorized, and nil otherwise. It avoids expensive joins and aggregations and
// is therefore much cheaper to call than GetExhaustiveSearchJob. If you want to
// get the job, call GetExhaustiveSearchJob instead.
func (s *Store) UserHasAccess(ctx context.Context, id int64) (err error) {
	ctx, _, endObservation := s.operations.userHasAccess.With(ctx, &err, opAttrs(
		attribute.Int64("ID", id),
	))
	defer endObservation(1, observation.Args{})

	q := sqlf.Sprintf("SELECT initiator_id FROM exhaustive_search_jobs WHERE id = %s", id)

	var initiatorID int32
	err = s.Store.QueryRow(ctx, q).Scan(&initiatorID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.Wrapf(ErrNoResults, "failed to scan job with id %d: %s", id, err.Error())
		}
		return err
	}
	if initiatorID == 0 {
		return ErrNoResults
	}

	// 🚨 SECURITY: only the initiator, internal or site admins may view a job.
	//
	// job id is just an incrementing integer that on any new job is returned. So
	// this information is not private so we can just return err to indicate the
	// reason for not returning the job.
	return auth.CheckSiteAdminOrSameUser(ctx, s.db, initiatorID)
}

// aggStateSubQuery takes the results from getAggregateStateTable and computes a
// single aggregate state that reflects the state of the entire search job
// cascade better than the state of the top-level worker.
//
// The processing chain is as follows:
//
// Execute getAggregateStateTable -> transpose table -> compute aggregate state
//
// # The result looks like this:
//
// | agg_state  |
// |------------|
// | processing |
//
// We want the aggregate state to be returned by the db, so we can use db
// filtering and pagination.
const aggStateSubQuery = `
		SELECT
		    -- Compute aggregate state
			CASE
				WHEN canceled > 0 THEN 'canceled'
				WHEN processing > 0 THEN 'processing'
				WHEN queued > 0 THEN 'queued'
				WHEN errored > 0 THEN 'processing'
				WHEN failed > 0 THEN 'failed'
				WHEN completed > 0 THEN 'completed'
			    -- This should never happen
				ELSE 'queued'
			END
		FROM (
-- | processing | queued | failed | completed |
-- |------------|--------|--------|-----------|
-- | 2          | 3      | 1      | 8         |
			SELECT
			    -- transpose the table
				max( CASE WHEN state = 'failed' THEN count END) AS failed,
				max( CASE WHEN state = 'processing' THEN count END) AS processing,
				max( CASE WHEN state = 'completed' THEN count END) AS completed,
				max( CASE WHEN state = 'queued' THEN count END) AS queued,
				max( CASE WHEN state = 'canceled' THEN count END) AS canceled,
				max( CASE WHEN state = 'errored' THEN count END) AS errored
			FROM (
				-- getAggregateStateTable
				%s) AS state_histogram) AS transposed_state_histogram
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
		conds = append(conds, sqlf.Sprintf("agg_state in (%s)", sqlf.Join(states, ",")))
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

	q := listSearchJobQuery(whereClause)
	if pagination != nil {
		q = pagination.AppendOrderToQuery(q)
		q = pagination.AppendLimitToQuery(q)
	}

	return scanExhaustiveSearchJobsList(s.Store.Query(ctx, q))
}

const listExhaustiveSearchJobsQueryFmtStr = `
SELECT * FROM (SELECT %s, (%s) as agg_state FROM exhaustive_search_jobs) as outer_query
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
	err = s.UserHasAccess(ctx, id)
	if err != nil {
		return err
	}

	return s.Exec(ctx, sqlf.Sprintf(deleteExhaustiveSearchJobQueryFmtStr, id))
}

// | state      | count |
// |------------|-------|
// | processing | 2     |
// | queued     | 3     |
// | failed     | 1     |
// | completed  | 8     |
const getAggregateStateTable = `
SELECT state, COUNT(*) as count
FROM
  (
		(SELECT state
		 -- we need the alias to avoid conflicts with embedding queries.
		 FROM exhaustive_search_jobs sj
		 WHERE sj.id = %s)
    UNION ALL
		(SELECT state
		 FROM exhaustive_search_repo_jobs rj
		 WHERE rj.search_job_id = %s)
    UNION ALL
		(SELECT rrj.state
		 FROM exhaustive_search_repo_revision_jobs rrj
		JOIN exhaustive_search_repo_jobs rj ON rrj.search_repo_job_id = rj.id
		WHERE rj.search_job_id = %s)
  ) AS sub
GROUP BY state
`

func (s *Store) GetAggregateRepoRevState(ctx context.Context, id int64) (_ map[string]int, err error) {
	ctx, _, endObservation := s.operations.getAggregateRepoRevState.With(ctx, &err, opAttrs(
		attribute.Int64("ID", id),
	))
	defer endObservation(1, observation.Args{})

	// 🚨 SECURITY: only someone with access to the job may cancel the job
	err = s.UserHasAccess(ctx, id)
	if err != nil {
		return nil, err
	}

	q := sqlf.Sprintf(getAggregateStateTable, id, id, id)

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
r.name,
rjj.revision,
rjj.state,
rjj.failure_message,
rjj.started_at,
rjj.finished_at
FROM exhaustive_search_repo_revision_jobs rjj
JOIN exhaustive_search_repo_jobs rj ON rjj.search_repo_job_id = rj.id
JOIN repo r ON r.id = rj.repo_id
%s
`

type GetJobLogsOpts struct {
	From  int64
	Limit int
}

func (s *Store) GetJobLogs(ctx context.Context, id int64, opts *GetJobLogsOpts) ([]types.SearchJobLog, error) {
	// 🚨 SECURITY: only someone with access to the job may access the logs
	err := s.UserHasAccess(ctx, id)
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
			&job.RepoName,
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

func defaultScanTargets(job *types.ExhaustiveSearchJob) []any {
	// required field for the sync worker, but
	// the value is thrown out here
	var executionLogs *[]any

	return []any{
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
	}
}

func scanExhaustiveSearchJob(sc dbutil.Scanner) (*types.ExhaustiveSearchJob, error) {
	var job types.ExhaustiveSearchJob

	return &job, sc.Scan(
		defaultScanTargets(&job)...,
	)
}

func scanExhaustiveSearchJobList(sc dbutil.Scanner) (*types.ExhaustiveSearchJob, error) {
	var job types.ExhaustiveSearchJob

	return &job, sc.Scan(
		append(
			defaultScanTargets(&job),
			&job.AggState,
		)...,
	)
}

var scanExhaustiveSearchJobsList = basestore.NewSliceScanner(scanExhaustiveSearchJobList)
