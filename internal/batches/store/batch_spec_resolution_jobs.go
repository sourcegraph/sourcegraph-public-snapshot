package store

import (
	"context"
	"fmt"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/attribute"

	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// batchSpecResolutionJobInsertColumns is the list of changeset_jobs columns that are
// modified in CreateBatchSpecResolutionJob.
var batchSpecResolutionJobInsertColumns = SQLColumns{
	"batch_spec_id",
	"initiator_id",

	"state",

	"created_at",
	"updated_at",
}

const batchSpecResolutionJobInsertColsFmt = `(%s, %s, %s, %s, %s)`

// ChangesetJobColumns are used by the changeset job related Store methods to query
// and create changeset jobs.
var batchSpecResolutionJobColums = SQLColumns{
	"batch_spec_resolution_jobs.id",

	"batch_spec_resolution_jobs.batch_spec_id",
	"batch_spec_resolution_jobs.initiator_id",

	"batch_spec_resolution_jobs.state",
	"batch_spec_resolution_jobs.failure_message",
	"batch_spec_resolution_jobs.started_at",
	"batch_spec_resolution_jobs.finished_at",
	"batch_spec_resolution_jobs.process_after",
	"batch_spec_resolution_jobs.num_resets",
	"batch_spec_resolution_jobs.num_failures",
	"batch_spec_resolution_jobs.execution_logs",
	"batch_spec_resolution_jobs.worker_hostname",

	"batch_spec_resolution_jobs.created_at",
	"batch_spec_resolution_jobs.updated_at",
}

// ErrResolutionJobAlreadyExists can be returned by
// CreateBatchSpecResolutionJob if a BatchSpecResolutionJob pointing at the
// same BatchSpec already exists.
type ErrResolutionJobAlreadyExists struct {
	BatchSpecID int64
}

func (e ErrResolutionJobAlreadyExists) Error() string {
	return fmt.Sprintf("a resolution job for batch spec %d already exists", e.BatchSpecID)
}

// CreateBatchSpecResolutionJob creates the given batch spec resolutionjob jobs.
func (s *Store) CreateBatchSpecResolutionJob(ctx context.Context, wj *btypes.BatchSpecResolutionJob) (err error) {
	ctx, _, endObservation := s.operations.createBatchSpecResolutionJob.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q := s.createBatchSpecResolutionJobQuery(wj)

	err = s.query(ctx, q, func(sc dbutil.Scanner) (err error) {
		return scanBatchSpecResolutionJob(wj, sc)
	})
	if err != nil && isUniqueConstraintViolation(err, "batch_spec_resolution_jobs_batch_spec_id_unique") {
		return ErrResolutionJobAlreadyExists{BatchSpecID: wj.BatchSpecID}
	}
	return err
}

var createBatchSpecResolutionJobQueryFmtstr = `
INSERT INTO batch_spec_resolution_jobs (%s)
VALUES ` + batchSpecResolutionJobInsertColsFmt + `
RETURNING %s
`

func (s *Store) createBatchSpecResolutionJobQuery(wj *btypes.BatchSpecResolutionJob) *sqlf.Query {
	if wj.CreatedAt.IsZero() {
		wj.CreatedAt = s.now()
	}

	if wj.UpdatedAt.IsZero() {
		wj.UpdatedAt = wj.CreatedAt
	}

	state := string(wj.State)
	if state == "" {
		state = string(btypes.BatchSpecResolutionJobStateQueued)
	}

	return sqlf.Sprintf(
		createBatchSpecResolutionJobQueryFmtstr,
		sqlf.Join(batchSpecResolutionJobInsertColumns.ToSqlf(), ","),
		wj.BatchSpecID,
		wj.InitiatorID,
		state,
		wj.CreatedAt,
		wj.UpdatedAt,
		sqlf.Join(batchSpecResolutionJobColums.ToSqlf(), ", "),
	)
}

// GetBatchSpecResolutionJobOpts captures the query options needed for getting a BatchSpecResolutionJob
type GetBatchSpecResolutionJobOpts struct {
	ID          int64
	BatchSpecID int64
}

// GetBatchSpecResolutionJob gets a BatchSpecResolutionJob matching the given options.
func (s *Store) GetBatchSpecResolutionJob(ctx context.Context, opts GetBatchSpecResolutionJobOpts) (job *btypes.BatchSpecResolutionJob, err error) {
	ctx, _, endObservation := s.operations.getBatchSpecResolutionJob.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("ID", int(opts.ID)),
		attribute.Int("BatchSpecID", int(opts.BatchSpecID)),
	}})
	defer endObservation(1, observation.Args{})

	q := getBatchSpecResolutionJobQuery(&opts)
	var c btypes.BatchSpecResolutionJob
	err = s.query(ctx, q, func(sc dbutil.Scanner) (err error) {
		return scanBatchSpecResolutionJob(&c, sc)
	})
	if err != nil {
		return nil, err
	}

	if c.ID == 0 {
		return nil, ErrNoResults
	}

	return &c, nil
}

var getBatchSpecResolutionJobsQueryFmtstr = `
SELECT %s FROM batch_spec_resolution_jobs
WHERE %s
LIMIT 1
`

func getBatchSpecResolutionJobQuery(opts *GetBatchSpecResolutionJobOpts) *sqlf.Query {
	var preds []*sqlf.Query

	if opts.ID != 0 {
		preds = append(preds, sqlf.Sprintf("batch_spec_resolution_jobs.id = %s", opts.ID))
	}

	if opts.BatchSpecID != 0 {
		preds = append(preds, sqlf.Sprintf("batch_spec_resolution_jobs.batch_spec_id = %s", opts.BatchSpecID))
	}

	return sqlf.Sprintf(
		getBatchSpecResolutionJobsQueryFmtstr,
		sqlf.Join(batchSpecResolutionJobColums.ToSqlf(), ", "),
		sqlf.Join(preds, "\n AND "),
	)
}

// ListBatchSpecResolutionJobsOpts captures the query options needed for
// listing batch spec resolutionjob jobs.
type ListBatchSpecResolutionJobsOpts struct {
	State          btypes.BatchSpecResolutionJobState
	WorkerHostname string
}

// ListBatchSpecResolutionJobs lists batch changes with the given filters.
func (s *Store) ListBatchSpecResolutionJobs(ctx context.Context, opts ListBatchSpecResolutionJobsOpts) (cs []*btypes.BatchSpecResolutionJob, err error) {
	ctx, _, endObservation := s.operations.listBatchSpecResolutionJobs.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q := listBatchSpecResolutionJobsQuery(opts)

	cs = make([]*btypes.BatchSpecResolutionJob, 0)
	err = s.query(ctx, q, func(sc dbutil.Scanner) error {
		var c btypes.BatchSpecResolutionJob
		if err := scanBatchSpecResolutionJob(&c, sc); err != nil {
			return err
		}
		cs = append(cs, &c)
		return nil
	})

	return cs, err
}

var listBatchSpecResolutionJobsQueryFmtstr = `
SELECT %s FROM batch_spec_resolution_jobs
WHERE %s
ORDER BY id ASC
`

func listBatchSpecResolutionJobsQuery(opts ListBatchSpecResolutionJobsOpts) *sqlf.Query {
	var preds []*sqlf.Query

	if opts.State != "" {
		preds = append(preds, sqlf.Sprintf("batch_spec_resolution_jobs.state = %s", opts.State))
	}

	if opts.WorkerHostname != "" {
		preds = append(preds, sqlf.Sprintf("batch_spec_resolution_jobs.worker_hostname = %s", opts.WorkerHostname))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(
		listBatchSpecResolutionJobsQueryFmtstr,
		sqlf.Join(batchSpecResolutionJobColums.ToSqlf(), ", "),
		sqlf.Join(preds, "\n AND "),
	)
}

func scanBatchSpecResolutionJob(rj *btypes.BatchSpecResolutionJob, s dbutil.Scanner) error {
	var executionLogs []executor.ExecutionLogEntry
	var failureMessage string

	if err := s.Scan(
		&rj.ID,
		&rj.BatchSpecID,
		&rj.InitiatorID,
		&rj.State,
		&dbutil.NullString{S: &failureMessage},
		&dbutil.NullTime{Time: &rj.StartedAt},
		&dbutil.NullTime{Time: &rj.FinishedAt},
		&dbutil.NullTime{Time: &rj.ProcessAfter},
		&rj.NumResets,
		&rj.NumFailures,
		pq.Array(&executionLogs),
		&rj.WorkerHostname,
		&rj.CreatedAt,
		&rj.UpdatedAt,
	); err != nil {
		return err
	}

	if failureMessage != "" {
		rj.FailureMessage = &failureMessage
	}

	rj.ExecutionLogs = append(rj.ExecutionLogs, executionLogs...)

	return nil
}
