package store

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go/log"

	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var BatchSpecWorkspaceExecutionJobColumns = SQLColumns{
	"batch_spec_workspace_execution_jobs.id",

	"batch_spec_workspace_execution_jobs.batch_spec_workspace_id",
	"batch_spec_workspace_execution_jobs.user_id",

	"batch_spec_workspace_execution_jobs.state",
	"batch_spec_workspace_execution_jobs.failure_message",
	"batch_spec_workspace_execution_jobs.started_at",
	"batch_spec_workspace_execution_jobs.finished_at",
	"batch_spec_workspace_execution_jobs.process_after",
	"batch_spec_workspace_execution_jobs.num_resets",
	"batch_spec_workspace_execution_jobs.num_failures",
	"batch_spec_workspace_execution_jobs.execution_logs",
	"batch_spec_workspace_execution_jobs.worker_hostname",
	"batch_spec_workspace_execution_jobs.cancel",

	"exec.place_in_user_queue",
	"exec.place_in_global_queue",

	"batch_spec_workspace_execution_jobs.created_at",
	"batch_spec_workspace_execution_jobs.updated_at",
}

var batchSpecWorkspaceExecutionJobColumnsWithNullQueue = SQLColumns{
	"batch_spec_workspace_execution_jobs.id",

	"batch_spec_workspace_execution_jobs.batch_spec_workspace_id",
	"batch_spec_workspace_execution_jobs.user_id",

	"batch_spec_workspace_execution_jobs.state",
	"batch_spec_workspace_execution_jobs.failure_message",
	"batch_spec_workspace_execution_jobs.started_at",
	"batch_spec_workspace_execution_jobs.finished_at",
	"batch_spec_workspace_execution_jobs.process_after",
	"batch_spec_workspace_execution_jobs.num_resets",
	"batch_spec_workspace_execution_jobs.num_failures",
	"batch_spec_workspace_execution_jobs.execution_logs",
	"batch_spec_workspace_execution_jobs.worker_hostname",
	"batch_spec_workspace_execution_jobs.cancel",

	"NULL AS place_in_user_queue",
	"NULL AS place_in_global_queue",

	"batch_spec_workspace_execution_jobs.created_at",
	"batch_spec_workspace_execution_jobs.updated_at",
}

const executionPlaceInQueueFragment = `
SELECT
	id, place_in_user_queue, place_in_global_queue
FROM batch_spec_workspace_execution_queue
`

const createBatchSpecWorkspaceExecutionJobsQueryFmtstr = `
-- source: enterprise/internal/batches/store/batch_spec_workspace_execution_jobs.go:CreateBatchSpecWorkspaceExecutionJobs
INSERT INTO
	batch_spec_workspace_execution_jobs (batch_spec_workspace_id, user_id)
SELECT
	batch_spec_workspaces.id,
	batch_specs.user_id
FROM
	batch_spec_workspaces
JOIN batch_specs ON batch_specs.id = batch_spec_workspaces.batch_spec_id
WHERE
	batch_spec_workspaces.batch_spec_id = %s
AND
	%s
`

const executableWorkspaceJobsConditionFmtstr = `
(
	(batch_specs.allow_ignored OR NOT batch_spec_workspaces.ignored)
	AND
	(batch_specs.allow_unsupported OR NOT batch_spec_workspaces.unsupported)
	AND
	-- TODO: Reimplement this. It was broken already, so no regression from the current state.
	-- NOT batch_spec_workspaces.skipped
	-- AND
	batch_spec_workspaces.cached_result_found IS FALSE
)`

// CreateBatchSpecWorkspaceExecutionJobs creates the given batch spec workspace jobs.
func (s *Store) CreateBatchSpecWorkspaceExecutionJobs(ctx context.Context, batchSpecID int64) (err error) {
	ctx, _, endObservation := s.operations.createBatchSpecWorkspaceExecutionJobs.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("batchSpecID", int(batchSpecID)),
	}})
	defer endObservation(1, observation.Args{})

	cond := sqlf.Sprintf(executableWorkspaceJobsConditionFmtstr)
	q := sqlf.Sprintf(createBatchSpecWorkspaceExecutionJobsQueryFmtstr, batchSpecID, cond)
	return s.Exec(ctx, q)
}

const createBatchSpecWorkspaceExecutionJobsForWorkspacesQueryFmtstr = `
-- source: enterprise/internal/batches/store/batch_spec_workspace_execution_jobs.go:CreateBatchSpecWorkspaceExecutionJobsForWorkspaces
INSERT INTO
	batch_spec_workspace_execution_jobs (batch_spec_workspace_id, user_id)
SELECT
	batch_spec_workspaces.id,
	batch_specs.user_id
FROM
	batch_spec_workspaces
JOIN
	batch_specs ON batch_specs.id = batch_spec_workspaces.batch_spec_id
WHERE
	batch_spec_workspaces.id = ANY (%s)
`

// CreateBatchSpecWorkspaceExecutionJobsForWorkspaces creates the batch spec workspace jobs for the given workspaces.
func (s *Store) CreateBatchSpecWorkspaceExecutionJobsForWorkspaces(ctx context.Context, workspaceIDs []int64) (err error) {
	ctx, _, endObservation := s.operations.createBatchSpecWorkspaceExecutionJobsForWorkspaces.With(ctx, &err, observation.Args{LogFields: []log.Field{}})
	defer endObservation(1, observation.Args{})

	q := sqlf.Sprintf(createBatchSpecWorkspaceExecutionJobsForWorkspacesQueryFmtstr, pq.Array(workspaceIDs))
	return s.Exec(ctx, q)
}

const deleteBatchSpecWorkspaceExecutionJobsQueryFmtstr = `
-- source: enterprise/internal/batches/store/batch_spec_workspace_execution_jobs.go:DeleteBatchSpecWorkspaceExecutionJobs
DELETE FROM
	batch_spec_workspace_execution_jobs
WHERE
	id = ANY (%s)
RETURNING id
`

// DeleteBatchSpecWorkspaceExecutionJobs
func (s *Store) DeleteBatchSpecWorkspaceExecutionJobs(ctx context.Context, ids []int64) (err error) {
	ctx, _, endObservation := s.operations.deleteBatchSpecWorkspaceExecutionJobs.With(ctx, &err, observation.Args{LogFields: []log.Field{}})
	defer endObservation(1, observation.Args{})

	q := sqlf.Sprintf(deleteBatchSpecWorkspaceExecutionJobsQueryFmtstr, pq.Array(ids))
	deleted, err := basestore.ScanInts(s.Query(ctx, q))
	if err != nil {
		return err
	}
	if len(deleted) != len(ids) {
		return errors.Newf("wrong number of jobs deleted: %d instead of %d", len(deleted), len(ids))
	}
	return nil
}

// GetBatchSpecWorkspaceExecutionJobOpts captures the query options needed for getting a BatchSpecWorkspaceExecutionJob
type GetBatchSpecWorkspaceExecutionJobOpts struct {
	ID                   int64
	BatchSpecWorkspaceID int64
}

// GetBatchSpecWorkspaceExecutionJob gets a BatchSpecWorkspaceExecutionJob matching the given options.
func (s *Store) GetBatchSpecWorkspaceExecutionJob(ctx context.Context, opts GetBatchSpecWorkspaceExecutionJobOpts) (job *btypes.BatchSpecWorkspaceExecutionJob, err error) {
	ctx, _, endObservation := s.operations.getBatchSpecWorkspaceExecutionJob.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("ID", int(opts.ID)),
	}})
	defer endObservation(1, observation.Args{})

	q := getBatchSpecWorkspaceExecutionJobQuery(&opts)
	var c btypes.BatchSpecWorkspaceExecutionJob
	err = s.query(ctx, q, func(sc dbutil.Scanner) (err error) {
		return ScanBatchSpecWorkspaceExecutionJob(&c, sc)
	})
	if err != nil {
		return nil, err
	}

	if c.ID == 0 {
		return nil, ErrNoResults
	}

	return &c, nil
}

var getBatchSpecWorkspaceExecutionJobsQueryFmtstr = `
-- source: enterprise/internal/batches/store/batch_spec_workspace_execution_jobs.go:GetBatchSpecWorkspaceExecutionJob
SELECT
	%s
FROM
	batch_spec_workspace_execution_jobs
LEFT JOIN (` + executionPlaceInQueueFragment + `) AS exec ON batch_spec_workspace_execution_jobs.id = exec.id
WHERE
	%s
LIMIT 1
`

func getBatchSpecWorkspaceExecutionJobQuery(opts *GetBatchSpecWorkspaceExecutionJobOpts) *sqlf.Query {
	var preds []*sqlf.Query
	if opts.ID != 0 {
		preds = append(preds, sqlf.Sprintf("batch_spec_workspace_execution_jobs.id = %s", opts.ID))
	}

	if opts.BatchSpecWorkspaceID != 0 {
		preds = append(preds, sqlf.Sprintf("batch_spec_workspace_execution_jobs.batch_spec_workspace_id = %s", opts.BatchSpecWorkspaceID))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(
		getBatchSpecWorkspaceExecutionJobsQueryFmtstr,
		sqlf.Join(BatchSpecWorkspaceExecutionJobColumns.ToSqlf(), ", "),
		sqlf.Join(preds, "\n AND "),
	)
}

// ListBatchSpecWorkspaceExecutionJobsOpts captures the query options needed for
// listing batch spec workspace execution jobs.
type ListBatchSpecWorkspaceExecutionJobsOpts struct {
	Cancel                 *bool
	State                  btypes.BatchSpecWorkspaceExecutionJobState
	WorkerHostname         string
	BatchSpecWorkspaceIDs  []int64
	IDs                    []int64
	OnlyWithFailureMessage bool
	BatchSpecID            int64
}

// ListBatchSpecWorkspaceExecutionJobs lists batch changes with the given filters.
func (s *Store) ListBatchSpecWorkspaceExecutionJobs(ctx context.Context, opts ListBatchSpecWorkspaceExecutionJobsOpts) (cs []*btypes.BatchSpecWorkspaceExecutionJob, err error) {
	ctx, _, endObservation := s.operations.listBatchSpecWorkspaceExecutionJobs.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q := listBatchSpecWorkspaceExecutionJobsQuery(opts)

	cs = make([]*btypes.BatchSpecWorkspaceExecutionJob, 0)
	err = s.query(ctx, q, func(sc dbutil.Scanner) error {
		var c btypes.BatchSpecWorkspaceExecutionJob
		if err := ScanBatchSpecWorkspaceExecutionJob(&c, sc); err != nil {
			return err
		}
		cs = append(cs, &c)
		return nil
	})

	return cs, err
}

var listBatchSpecWorkspaceExecutionJobsQueryFmtstr = `
SELECT
	%s
FROM
	batch_spec_workspace_execution_jobs
LEFT JOIN (` + executionPlaceInQueueFragment + `) as exec ON batch_spec_workspace_execution_jobs.id = exec.id
%s       -- joins
WHERE
	%s   -- preds
ORDER BY batch_spec_workspace_execution_jobs.id ASC
`

func listBatchSpecWorkspaceExecutionJobsQuery(opts ListBatchSpecWorkspaceExecutionJobsOpts) *sqlf.Query {
	var preds []*sqlf.Query
	var joins []*sqlf.Query

	if opts.State != "" {
		preds = append(preds, sqlf.Sprintf("batch_spec_workspace_execution_jobs.state = %s", opts.State))
	}

	if opts.WorkerHostname != "" {
		preds = append(preds, sqlf.Sprintf("batch_spec_workspace_execution_jobs.worker_hostname = %s", opts.WorkerHostname))
	}

	if opts.Cancel != nil {
		preds = append(preds, sqlf.Sprintf("batch_spec_workspace_execution_jobs.cancel = %s", *opts.Cancel))
	}

	if len(opts.BatchSpecWorkspaceIDs) != 0 {
		preds = append(preds, sqlf.Sprintf("batch_spec_workspace_execution_jobs.batch_spec_workspace_id = ANY (%s)", pq.Array(opts.BatchSpecWorkspaceIDs)))
	}

	if len(opts.IDs) != 0 {
		preds = append(preds, sqlf.Sprintf("batch_spec_workspace_execution_jobs.id = ANY (%s)", pq.Array(opts.IDs)))
	}

	if opts.OnlyWithFailureMessage {
		preds = append(preds, sqlf.Sprintf("batch_spec_workspace_execution_jobs.state IN ('errored', 'failed')"))
		preds = append(preds, sqlf.Sprintf("batch_spec_workspace_execution_jobs.failure_message IS NOT NULL"))
	}

	if opts.BatchSpecID != 0 {
		joins = append(joins, sqlf.Sprintf("JOIN batch_spec_workspaces ON batch_spec_workspace_execution_jobs.batch_spec_workspace_id = batch_spec_workspaces.id"))
		preds = append(preds, sqlf.Sprintf("batch_spec_workspaces.batch_spec_id = %d", opts.BatchSpecID))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(
		listBatchSpecWorkspaceExecutionJobsQueryFmtstr,
		sqlf.Join(BatchSpecWorkspaceExecutionJobColumns.ToSqlf(), ", "),
		sqlf.Join(joins, "\n"),
		sqlf.Join(preds, "\n AND "),
	)
}

// CancelBatchSpecWorkspaceExecutionJobsOpts captures the query options needed for
// canceling batch spec workspace execution jobs.
type CancelBatchSpecWorkspaceExecutionJobsOpts struct {
	BatchSpecID int64
	IDs         []int64
}

// CancelBatchSpecWorkspaceExecutionJobs cancels the matching
// BatchSpecWorkspaceExecutionJobs.
//
// The returned list of records may not match the list of the given IDs, if
// some of the records were already canceled, completed, failed, errored, etc.
func (s *Store) CancelBatchSpecWorkspaceExecutionJobs(ctx context.Context, opts CancelBatchSpecWorkspaceExecutionJobsOpts) (jobs []*btypes.BatchSpecWorkspaceExecutionJob, err error) {
	ctx, _, endObservation := s.operations.cancelBatchSpecWorkspaceExecutionJobs.With(ctx, &err, observation.Args{LogFields: []log.Field{}})
	defer endObservation(1, observation.Args{})

	if opts.BatchSpecID == 0 && len(opts.IDs) == 0 {
		return nil, errors.New("invalid options: would cancel all jobs")
	}

	q := s.cancelBatchSpecWorkspaceExecutionJobQuery(opts)

	jobs = make([]*btypes.BatchSpecWorkspaceExecutionJob, 0)
	err = s.query(ctx, q, func(sc dbutil.Scanner) (err error) {
		var j btypes.BatchSpecWorkspaceExecutionJob
		if err := ScanBatchSpecWorkspaceExecutionJob(&j, sc); err != nil {
			return err
		}
		jobs = append(jobs, &j)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return jobs, nil
}

var cancelBatchSpecWorkspaceExecutionJobsQueryFmtstr = `
-- source: enterprise/internal/batches/store/batch_spec_workspace_execution_jobs.go:CancelBatchSpecWorkspaceExecutionJobs
WITH candidates AS (
	SELECT
		batch_spec_workspace_execution_jobs.id
	FROM
		batch_spec_workspace_execution_jobs
	%s  -- joins
	WHERE
		%s -- preds
		AND
		-- It must be queued or processing, we cannot cancel jobs that have already completed.
		batch_spec_workspace_execution_jobs.state IN (%s, %s)
	ORDER BY id
	FOR UPDATE
),
updated_candidates AS (
	UPDATE
		batch_spec_workspace_execution_jobs
	SET
		cancel = TRUE,
		-- If the execution is still queued, we directly abort, otherwise we keep the
		-- state, so the worker can to teardown and, at some point, mark it failed itself.
		state = CASE WHEN batch_spec_workspace_execution_jobs.state = %s THEN batch_spec_workspace_execution_jobs.state ELSE %s END,
		finished_at = CASE WHEN batch_spec_workspace_execution_jobs.state = %s THEN batch_spec_workspace_execution_jobs.finished_at ELSE %s END,
		updated_at = %s
	WHERE
		id IN (SELECT id FROM candidates)
	RETURNING *
)
SELECT
	%s
FROM updated_candidates batch_spec_workspace_execution_jobs
LEFT JOIN (` + executionPlaceInQueueFragment + `) as exec ON batch_spec_workspace_execution_jobs.id = exec.id
`

func (s *Store) cancelBatchSpecWorkspaceExecutionJobQuery(opts CancelBatchSpecWorkspaceExecutionJobsOpts) *sqlf.Query {
	var preds []*sqlf.Query
	var joins []*sqlf.Query

	if len(opts.IDs) != 0 {
		preds = append(preds, sqlf.Sprintf("batch_spec_workspace_execution_jobs.id = ANY (%s)", pq.Array(opts.IDs)))
	}

	if opts.BatchSpecID != 0 {
		joins = append(joins, sqlf.Sprintf("JOIN batch_spec_workspaces ON batch_spec_workspaces.id = batch_spec_workspace_execution_jobs.batch_spec_workspace_id"))
		preds = append(preds, sqlf.Sprintf("batch_spec_workspaces.batch_spec_id = %s", opts.BatchSpecID))
	}

	return sqlf.Sprintf(
		cancelBatchSpecWorkspaceExecutionJobsQueryFmtstr,
		sqlf.Join(joins, "\n"),
		sqlf.Join(preds, "\n AND "),
		btypes.BatchSpecWorkspaceExecutionJobStateQueued,
		btypes.BatchSpecWorkspaceExecutionJobStateProcessing,
		btypes.BatchSpecWorkspaceExecutionJobStateProcessing,
		btypes.BatchSpecWorkspaceExecutionJobStateFailed,
		btypes.BatchSpecWorkspaceExecutionJobStateProcessing,
		s.now(),
		s.now(),
		sqlf.Join(BatchSpecWorkspaceExecutionJobColumns.ToSqlf(), ", "),
	)
}

func ScanBatchSpecWorkspaceExecutionJob(wj *btypes.BatchSpecWorkspaceExecutionJob, s dbutil.Scanner) error {
	var executionLogs []dbworkerstore.ExecutionLogEntry
	var failureMessage string

	if err := s.Scan(
		&wj.ID,
		&wj.BatchSpecWorkspaceID,
		&wj.UserID,
		&wj.State,
		&dbutil.NullString{S: &failureMessage},
		&dbutil.NullTime{Time: &wj.StartedAt},
		&dbutil.NullTime{Time: &wj.FinishedAt},
		&dbutil.NullTime{Time: &wj.ProcessAfter},
		&wj.NumResets,
		&wj.NumFailures,
		pq.Array(&executionLogs),
		&wj.WorkerHostname,
		&wj.Cancel,
		&dbutil.NullInt64{N: &wj.PlaceInUserQueue},
		&dbutil.NullInt64{N: &wj.PlaceInGlobalQueue},
		&wj.CreatedAt,
		&wj.UpdatedAt,
	); err != nil {
		return err
	}

	if failureMessage != "" {
		wj.FailureMessage = &failureMessage
	}

	for _, entry := range executionLogs {
		wj.ExecutionLogs = append(wj.ExecutionLogs, workerutil.ExecutionLogEntry(entry))
	}

	return nil
}

func scanFirstBatchSpecWorkspaceExecutionJob(rows *sql.Rows, err error) (*btypes.BatchSpecWorkspaceExecutionJob, bool, error) {
	jobs, err := scanBatchSpecWorkspaceExecutionJobs(rows, err)
	if err != nil || len(jobs) == 0 {
		return nil, false, err
	}
	return jobs[0], true, nil
}

func scanBatchSpecWorkspaceExecutionJobs(rows *sql.Rows, queryErr error) ([]*btypes.BatchSpecWorkspaceExecutionJob, error) {
	if queryErr != nil {
		return nil, queryErr
	}

	var jobs []*btypes.BatchSpecWorkspaceExecutionJob

	return jobs, scanAll(rows, func(sc dbutil.Scanner) (err error) {
		var j btypes.BatchSpecWorkspaceExecutionJob
		if err = ScanBatchSpecWorkspaceExecutionJob(&j, sc); err != nil {
			return err
		}
		jobs = append(jobs, &j)
		return nil
	})
}
