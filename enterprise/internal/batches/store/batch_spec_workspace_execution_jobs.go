package store

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go/log"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

var batchSpecWorkspaceExecutionJobInsertColumns = []string{
	"batch_spec_workspace_id",

	"created_at",
	"updated_at",
}

var BatchSpecWorkspaceExecutionJobColums = SQLColumns{
	"batch_spec_workspace_execution_jobs.id",

	"batch_spec_workspace_execution_jobs.batch_spec_workspace_id",

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

	"batch_spec_workspace_execution_jobs.created_at",
	"batch_spec_workspace_execution_jobs.updated_at",
}

// CreateBatchSpecWorkspaceExecutionJob creates the given batch spec workspace jobs.
func (s *Store) CreateBatchSpecWorkspaceExecutionJob(ctx context.Context, jobs ...*btypes.BatchSpecWorkspaceExecutionJob) (err error) {
	ctx, endObservation := s.operations.createBatchSpecWorkspaceExecutionJob.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("count", len(jobs)),
	}})
	defer endObservation(1, observation.Args{})

	inserter := func(inserter *batch.Inserter) error {
		for _, job := range jobs {
			if job.CreatedAt.IsZero() {
				job.CreatedAt = s.now()
			}

			if job.UpdatedAt.IsZero() {
				job.UpdatedAt = job.CreatedAt
			}

			if err := inserter.Insert(
				ctx,
				job.BatchSpecWorkspaceID,
				job.CreatedAt,
				job.UpdatedAt,
			); err != nil {
				return err
			}
		}

		return nil
	}
	i := -1
	return batch.WithInserterWithReturn(
		ctx,
		s.Handle().DB(),
		"batch_spec_workspace_execution_jobs",
		batchSpecWorkspaceExecutionJobInsertColumns,
		BatchSpecWorkspaceExecutionJobColums,
		func(rows *sql.Rows) error {
			i++
			return scanBatchSpecWorkspaceExecutionJob(jobs[i], rows)
		},
		inserter,
	)
}

// GetBatchSpecWorkspaceExecutionJobOpts captures the query options needed for getting a BatchSpecWorkspaceExecutionJob
type GetBatchSpecWorkspaceExecutionJobOpts struct {
	ID                   int64
	BatchSpecWorkspaceID int64
}

// GetBatchSpecWorkspaceExecutionJob gets a BatchSpecWorkspaceExecutionJob matching the given options.
func (s *Store) GetBatchSpecWorkspaceExecutionJob(ctx context.Context, opts GetBatchSpecWorkspaceExecutionJobOpts) (job *btypes.BatchSpecWorkspaceExecutionJob, err error) {
	ctx, endObservation := s.operations.getBatchSpecWorkspaceExecutionJob.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("ID", int(opts.ID)),
	}})
	defer endObservation(1, observation.Args{})

	q := getBatchSpecWorkspaceExecutionJobQuery(&opts)
	var c btypes.BatchSpecWorkspaceExecutionJob
	err = s.query(ctx, q, func(sc scanner) (err error) {
		return scanBatchSpecWorkspaceExecutionJob(&c, sc)
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
SELECT %s FROM batch_spec_workspace_execution_jobs WHERE %s LIMIT 1
`

func getBatchSpecWorkspaceExecutionJobQuery(opts *GetBatchSpecWorkspaceExecutionJobOpts) *sqlf.Query {
	var preds []*sqlf.Query
	if opts.ID != 0 {
		preds = append(preds, sqlf.Sprintf("id = %s", opts.ID))
	}

	if opts.BatchSpecWorkspaceID != 0 {
		preds = append(preds, sqlf.Sprintf("batch_spec_workspace_id = %s", opts.BatchSpecWorkspaceID))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(
		getBatchSpecWorkspaceExecutionJobsQueryFmtstr,
		sqlf.Join(BatchSpecWorkspaceExecutionJobColums.ToSqlf(), ", "),
		sqlf.Join(preds, "\n AND "),
	)
}

// ListBatchSpecWorkspaceExecutionJobsOpts captures the query options needed for
// listing batch spec workspace execution jobs.
type ListBatchSpecWorkspaceExecutionJobsOpts struct {
	Cancel         *bool
	State          btypes.BatchSpecWorkspaceExecutionJobState
	WorkerHostname string
}

// ListBatchSpecWorkspaceExecutionJobs lists batch changes with the given filters.
func (s *Store) ListBatchSpecWorkspaceExecutionJobs(ctx context.Context, opts ListBatchSpecWorkspaceExecutionJobsOpts) (cs []*btypes.BatchSpecWorkspaceExecutionJob, err error) {
	ctx, endObservation := s.operations.listBatchSpecWorkspaceExecutionJobs.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q := listBatchSpecWorkspaceExecutionJobsQuery(opts)

	cs = make([]*btypes.BatchSpecWorkspaceExecutionJob, 0)
	err = s.query(ctx, q, func(sc scanner) error {
		var c btypes.BatchSpecWorkspaceExecutionJob
		if err := scanBatchSpecWorkspaceExecutionJob(&c, sc); err != nil {
			return err
		}
		cs = append(cs, &c)
		return nil
	})

	return cs, err
}

var listBatchSpecWorkspaceExecutionJobsQueryFmtstr = `
-- source: enterprise/internal/batches/store/batch_spec_workspace_execution_jobs.go:ListBatchSpecWorkspaceExecutionJobs
SELECT %s FROM batch_spec_workspace_execution_jobs
WHERE %s
ORDER BY id ASC
`

func listBatchSpecWorkspaceExecutionJobsQuery(opts ListBatchSpecWorkspaceExecutionJobsOpts) *sqlf.Query {
	var preds []*sqlf.Query

	if opts.State != "" {
		preds = append(preds, sqlf.Sprintf("batch_spec_workspace_execution_jobs.state = %s", opts.State))
	}

	if opts.WorkerHostname != "" {
		preds = append(preds, sqlf.Sprintf("batch_spec_workspace_execution_jobs.worker_hostname = %s", opts.WorkerHostname))
	}

	if opts.Cancel != nil {
		preds = append(preds, sqlf.Sprintf("batch_spec_workspace_execution_jobs.cancel = %s", *opts.Cancel))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(
		listBatchSpecWorkspaceExecutionJobsQueryFmtstr,
		sqlf.Join(BatchSpecWorkspaceExecutionJobColums.ToSqlf(), ", "),
		sqlf.Join(preds, "\n AND "),
	)
}

// CancelBatchSpecWorkspaceExecutionJob lists batch changes with the given filters.
func (s *Store) CancelBatchSpecWorkspaceExecutionJob(ctx context.Context, id int64) (job *btypes.BatchSpecWorkspaceExecutionJob, err error) {
	ctx, endObservation := s.operations.cancelBatchSpecWorkspaceExecutionJob.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("ID", int(id)),
	}})
	defer endObservation(1, observation.Args{})

	q := s.cancelBatchSpecWorkspaceExecutionJobQuery(id)
	var c btypes.BatchSpecWorkspaceExecutionJob
	err = s.query(ctx, q, func(sc scanner) (err error) {
		return scanBatchSpecWorkspaceExecutionJob(&c, sc)
	})
	if err != nil {
		return nil, err
	}

	if c.ID == 0 {
		return nil, ErrNoResults
	}

	return &c, nil
}

var cancelBatchSpecWorkspaceExecutionJobQueryFmtstr = `
-- source: enterprise/internal/batches/store/batch_spec_workspace_execution_jobs.go:CancelBatchSpecWorkspaceExecutionJob
WITH candidate AS (
	SELECT
		id
	FROM
		batch_spec_workspace_execution_jobs
	WHERE
		id = %s
		AND
		-- It must be queued or processing, we cannot cancel jobs that have already completed.
		state IN (%s, %s)
	FOR UPDATE
)
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
	id IN (SELECT id FROM candidate)
RETURNING %s
`

func (s *Store) cancelBatchSpecWorkspaceExecutionJobQuery(id int64) *sqlf.Query {
	return sqlf.Sprintf(
		cancelBatchSpecWorkspaceExecutionJobQueryFmtstr,
		id,
		btypes.BatchSpecWorkspaceExecutionJobStateQueued,
		btypes.BatchSpecWorkspaceExecutionJobStateProcessing,
		btypes.BatchSpecWorkspaceExecutionJobStateProcessing,
		btypes.BatchSpecWorkspaceExecutionJobStateFailed,
		btypes.BatchSpecWorkspaceExecutionJobStateProcessing,
		s.now(),
		s.now(),
		sqlf.Join(BatchSpecWorkspaceExecutionJobColums.ToSqlf(), ", "),
	)
}

func scanBatchSpecWorkspaceExecutionJob(wj *btypes.BatchSpecWorkspaceExecutionJob, s scanner) error {
	var executionLogs []dbworkerstore.ExecutionLogEntry
	var failureMessage string

	if err := s.Scan(
		&wj.ID,
		&wj.BatchSpecWorkspaceID,
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

func ScanFirstBatchSpecWorkspaceExecutionJob(rows *sql.Rows, err error) (*btypes.BatchSpecWorkspaceExecutionJob, bool, error) {
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

	return jobs, scanAll(rows, func(sc scanner) (err error) {
		var j btypes.BatchSpecWorkspaceExecutionJob
		if err = scanBatchSpecWorkspaceExecutionJob(&j, sc); err != nil {
			return err
		}
		jobs = append(jobs, &j)
		return nil
	})
}
