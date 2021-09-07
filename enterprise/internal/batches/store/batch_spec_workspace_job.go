package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"sort"

	"github.com/cockroachdb/errors"
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

// batchSpecWorkspaceJobInsertColumns is the list of changeset_jobs columns that are
// modified in CreateChangesetJob.
var batchSpecWorkspaceJobInsertColumns = []string{
	"batch_spec_id",
	"changeset_spec_ids",

	"repo_id",
	"branch",
	"commit",
	"path",
	"file_matches",
	"only_fetch_workspace",
	"steps",

	"state",

	"created_at",
	"updated_at",
}

// ChangesetJobColumns are used by the changeset job related Store methods to query
// and create changeset jobs.
var BatchSpecWorkspaceJobColumns = SQLColumns{
	"batch_spec_workspace_jobs.id",

	"batch_spec_workspace_jobs.batch_spec_id",
	"batch_spec_workspace_jobs.changeset_spec_ids",

	"batch_spec_workspace_jobs.repo_id",
	"batch_spec_workspace_jobs.branch",
	"batch_spec_workspace_jobs.commit",
	"batch_spec_workspace_jobs.path",
	"batch_spec_workspace_jobs.file_matches",
	"batch_spec_workspace_jobs.only_fetch_workspace",
	"batch_spec_workspace_jobs.steps",

	"batch_spec_workspace_jobs.state",
	"batch_spec_workspace_jobs.failure_message",
	"batch_spec_workspace_jobs.started_at",
	"batch_spec_workspace_jobs.finished_at",
	"batch_spec_workspace_jobs.process_after",
	"batch_spec_workspace_jobs.num_resets",
	"batch_spec_workspace_jobs.num_failures",
	"batch_spec_workspace_jobs.execution_logs",
	"batch_spec_workspace_jobs.worker_hostname",

	"batch_spec_workspace_jobs.created_at",
	"batch_spec_workspace_jobs.updated_at",
}

// CreateBatchSpecWorkspaceJob creates the given batch spec workspace jobs.
func (s *Store) CreateBatchSpecWorkspaceJob(ctx context.Context, ws ...*btypes.BatchSpecWorkspaceJob) (err error) {
	ctx, endObservation := s.operations.createBatchSpecWorkspaceJob.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("count", len(ws)),
	}})
	defer endObservation(1, observation.Args{})

	inserter := func(inserter *batch.Inserter) error {
		for _, wj := range ws {
			if wj.CreatedAt.IsZero() {
				wj.CreatedAt = s.now()
			}

			if wj.UpdatedAt.IsZero() {
				wj.UpdatedAt = wj.CreatedAt
			}

			changesetSpecIDs := make(map[int64]struct{}, len(wj.ChangesetSpecIDs))
			for _, id := range wj.ChangesetSpecIDs {
				changesetSpecIDs[id] = struct{}{}
			}

			marshaledIDs, err := json.Marshal(changesetSpecIDs)
			if err != nil {
				return err
			}

			if wj.FileMatches == nil {
				wj.FileMatches = []string{}
			}

			marshaledSteps, err := json.Marshal(wj.Steps)
			if err != nil {
				return err
			}

			if err := inserter.Insert(
				ctx,
				wj.BatchSpecID,
				marshaledIDs,
				wj.RepoID,
				wj.Branch,
				wj.Commit,
				wj.Path,
				pq.Array(wj.FileMatches),
				wj.OnlyFetchWorkspace,
				marshaledSteps,
				wj.State.ToDB(),
				wj.CreatedAt,
				wj.UpdatedAt,
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
		"batch_spec_workspace_jobs",
		batchSpecWorkspaceJobInsertColumns,
		BatchSpecWorkspaceJobColumns,
		func(rows *sql.Rows) error {
			i++
			return scanBatchSpecWorkspaceJob(ws[i], rows)
		},
		inserter,
	)
}

// GetBatchSpecWorkspaceJobOpts captures the query options needed for getting a BatchSpecWorkspaceJob
type GetBatchSpecWorkspaceJobOpts struct {
	ID int64
}

// GetBatchSpecWorkspaceJob gets a BatchSpecWorkspaceJob matching the given options.
func (s *Store) GetBatchSpecWorkspaceJob(ctx context.Context, opts GetBatchSpecWorkspaceJobOpts) (job *btypes.BatchSpecWorkspaceJob, err error) {
	ctx, endObservation := s.operations.getBatchSpecWorkspaceJob.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("ID", int(opts.ID)),
	}})
	defer endObservation(1, observation.Args{})

	q := getBatchSpecWorkspaceJobQuery(&opts)
	var c btypes.BatchSpecWorkspaceJob
	err = s.query(ctx, q, func(sc scanner) (err error) {
		return scanBatchSpecWorkspaceJob(&c, sc)
	})
	if err != nil {
		return nil, err
	}

	if c.ID == 0 {
		return nil, ErrNoResults
	}

	return &c, nil
}

var getBatchSpecWorkspaceJobsQueryFmtstr = `
-- source: enterprise/internal/batches/store/batch_spec_workspace_jobs.go:GetBatchSpecWorkspaceJob
SELECT %s FROM batch_spec_workspace_jobs
INNER JOIN repo ON repo.id = batch_spec_workspace_jobs.repo_id
WHERE %s
LIMIT 1
`

func getBatchSpecWorkspaceJobQuery(opts *GetBatchSpecWorkspaceJobOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.deleted_at IS NULL"),
		sqlf.Sprintf("batch_spec_workspace_jobs.id = %s", opts.ID),
	}

	return sqlf.Sprintf(
		getBatchSpecWorkspaceJobsQueryFmtstr,
		sqlf.Join(BatchSpecWorkspaceJobColumns.ToSqlf(), ", "),
		sqlf.Join(preds, "\n AND "),
	)
}

// ListBatchSpecWorkspaceJobsOpts captures the query options needed for
// listing batch spec workspace jobs.
type ListBatchSpecWorkspaceJobsOpts struct {
	State          btypes.BatchSpecWorkspaceJobState
	WorkerHostname string
}

// ListBatchSpecWorkspaceJobs lists batch changes with the given filters.
func (s *Store) ListBatchSpecWorkspaceJobs(ctx context.Context, opts ListBatchSpecWorkspaceJobsOpts) (cs []*btypes.BatchSpecWorkspaceJob, err error) {
	ctx, endObservation := s.operations.listBatchSpecWorkspaceJobs.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q := listBatchSpecWorkspaceJobsQuery(opts)

	cs = make([]*btypes.BatchSpecWorkspaceJob, 0)
	err = s.query(ctx, q, func(sc scanner) error {
		var c btypes.BatchSpecWorkspaceJob
		if err := scanBatchSpecWorkspaceJob(&c, sc); err != nil {
			return err
		}
		cs = append(cs, &c)
		return nil
	})

	return cs, err
}

var listBatchSpecWorkspaceJobsQueryFmtstr = `
-- source: enterprise/internal/batches/store/batch_spec_workspace_job.go:ListBatchSpecWorkspaceJobs
SELECT %s FROM batch_spec_workspace_jobs
INNER JOIN repo ON repo.id = batch_spec_workspace_jobs.repo_id
WHERE %s
ORDER BY id ASC
`

func listBatchSpecWorkspaceJobsQuery(opts ListBatchSpecWorkspaceJobsOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.deleted_at IS NULL"),
	}

	if opts.State != "" {
		preds = append(preds, sqlf.Sprintf("batch_spec_workspace_jobs.state = %s", opts.State))
	}

	if opts.WorkerHostname != "" {
		preds = append(preds, sqlf.Sprintf("batch_spec_workspace_jobs.worker_hostname = %s", opts.WorkerHostname))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(
		listBatchSpecWorkspaceJobsQueryFmtstr,
		sqlf.Join(BatchSpecWorkspaceJobColumns.ToSqlf(), ", "),
		sqlf.Join(preds, "\n AND "),
	)
}

func scanBatchSpecWorkspaceJob(wj *btypes.BatchSpecWorkspaceJob, s scanner) error {
	var executionLogs []dbworkerstore.ExecutionLogEntry
	var failureMessage string
	var steps json.RawMessage

	if err := s.Scan(
		&wj.ID,
		&wj.BatchSpecID,
		&jsonIDsSet{Assocs: &wj.ChangesetSpecIDs},
		&wj.RepoID,
		&wj.Branch,
		&wj.Commit,
		&wj.Path,
		pq.Array(&wj.FileMatches),
		&wj.OnlyFetchWorkspace,
		&steps,
		&wj.State,
		&dbutil.NullString{S: &failureMessage},
		&dbutil.NullTime{Time: &wj.StartedAt},
		&dbutil.NullTime{Time: &wj.FinishedAt},
		&dbutil.NullTime{Time: &wj.ProcessAfter},
		&wj.NumResets,
		&wj.NumFailures,
		pq.Array(&executionLogs),
		&wj.WorkerHostname,
		&wj.CreatedAt,
		&wj.UpdatedAt,
	); err != nil {
		return err
	}

	if err := json.Unmarshal(steps, &wj.Steps); err != nil {
		return errors.Wrap(err, "scanBatchSpecWorkspaceJob: failed to unmarshal Steps")
	}

	if failureMessage != "" {
		wj.FailureMessage = &failureMessage
	}

	for _, entry := range executionLogs {
		wj.ExecutionLogs = append(wj.ExecutionLogs, workerutil.ExecutionLogEntry(entry))
	}

	return nil
}

func ScanFirstBatchSpecWorkspaceJob(rows *sql.Rows, err error) (*btypes.BatchSpecWorkspaceJob, bool, error) {
	jobs, err := scanBatchSpecWorkspaceJobs(rows, err)
	if err != nil || len(jobs) == 0 {
		return nil, false, err
	}
	return jobs[0], true, nil
}

func scanBatchSpecWorkspaceJobs(rows *sql.Rows, queryErr error) ([]*btypes.BatchSpecWorkspaceJob, error) {
	if queryErr != nil {
		return nil, queryErr
	}

	var jobs []*btypes.BatchSpecWorkspaceJob

	return jobs, scanAll(rows, func(sc scanner) (err error) {
		var j btypes.BatchSpecWorkspaceJob
		if err = scanBatchSpecWorkspaceJob(&j, sc); err != nil {
			return err
		}
		jobs = append(jobs, &j)
		return nil
	})
}

// jsonIDsSet represents a "join table" set as a JSONB object where the keys
// are the ids and the values are json objects. It implements the sql.Scanner
// interface so it can be used as a scan destination, similar to
// sql.NullString.
type jsonIDsSet struct {
	Assocs *[]int64
}

// Scan implements the Scanner interface.
func (n *jsonIDsSet) Scan(value interface{}) error {
	m := make(map[int64]struct{})

	switch value := value.(type) {
	case nil:
	case []byte:
		if err := json.Unmarshal(value, &m); err != nil {
			return err
		}
	default:
		return errors.Errorf("value is not []byte: %T", value)
	}

	if *n.Assocs == nil {
		*n.Assocs = make([]int64, 0, len(m))
	} else {
		*n.Assocs = (*n.Assocs)[:0]
	}

	for id := range m {
		*n.Assocs = append(*n.Assocs, id)
	}

	sort.Slice(*n.Assocs, func(i, j int) bool {
		return (*n.Assocs)[i] < (*n.Assocs)[j]
	})

	return nil
}
