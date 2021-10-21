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
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
)

// batchSpecWorkspaceInsertColumns is the list of batch_spec_workspaces columns
// that are modified in CreateBatchSpecWorkspace
var batchSpecWorkspaceInsertColumns = []string{
	"batch_spec_id",
	"changeset_spec_ids",

	"repo_id",
	"branch",
	"commit",
	"path",
	"file_matches",
	"only_fetch_workspace",
	"steps",
	"unsupported",
	"ignored",
	"skipped",

	"created_at",
	"updated_at",
}

// BatchSpecWorkspaceColums are used by the changeset job related Store methods to query
// and create changeset jobs.
var BatchSpecWorkspaceColums = SQLColumns{
	"batch_spec_workspaces.id",

	"batch_spec_workspaces.batch_spec_id",
	"batch_spec_workspaces.changeset_spec_ids",

	"batch_spec_workspaces.repo_id",
	"batch_spec_workspaces.branch",
	"batch_spec_workspaces.commit",
	"batch_spec_workspaces.path",
	"batch_spec_workspaces.file_matches",
	"batch_spec_workspaces.only_fetch_workspace",
	"batch_spec_workspaces.steps",
	"batch_spec_workspaces.unsupported",
	"batch_spec_workspaces.ignored",
	"batch_spec_workspaces.skipped",

	"batch_spec_workspaces.created_at",
	"batch_spec_workspaces.updated_at",
}

// CreateBatchSpecWorkspace creates the given batch spec workspace jobs.
func (s *Store) CreateBatchSpecWorkspace(ctx context.Context, ws ...*btypes.BatchSpecWorkspace) (err error) {
	ctx, endObservation := s.operations.createBatchSpecWorkspace.With(ctx, &err, observation.Args{LogFields: []log.Field{
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

			if wj.Steps == nil {
				wj.Steps = []batcheslib.Step{}
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
				wj.Unsupported,
				wj.Ignored,
				wj.Skipped,
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
		"batch_spec_workspaces",
		batchSpecWorkspaceInsertColumns,
		"",
		BatchSpecWorkspaceColums,
		func(rows *sql.Rows) error {
			i++
			return scanBatchSpecWorkspace(ws[i], rows)
		},
		inserter,
	)
}

// GetBatchSpecWorkspaceOpts captures the query options needed for getting a BatchSpecWorkspace
type GetBatchSpecWorkspaceOpts struct {
	ID int64
}

// GetBatchSpecWorkspace gets a BatchSpecWorkspace matching the given options.
func (s *Store) GetBatchSpecWorkspace(ctx context.Context, opts GetBatchSpecWorkspaceOpts) (job *btypes.BatchSpecWorkspace, err error) {
	ctx, endObservation := s.operations.getBatchSpecWorkspace.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("ID", int(opts.ID)),
	}})
	defer endObservation(1, observation.Args{})

	q := getBatchSpecWorkspaceQuery(&opts)
	var c btypes.BatchSpecWorkspace
	err = s.query(ctx, q, func(sc dbutil.Scanner) (err error) {
		return scanBatchSpecWorkspace(&c, sc)
	})
	if err != nil {
		return nil, err
	}

	if c.ID == 0 {
		return nil, ErrNoResults
	}

	return &c, nil
}

var getBatchSpecWorkspacesQueryFmtstr = `
-- source: enterprise/internal/batches/store/batch_spec_workspaces.go:GetBatchSpecWorkspace
SELECT %s FROM batch_spec_workspaces
INNER JOIN repo ON repo.id = batch_spec_workspaces.repo_id
WHERE %s
LIMIT 1
`

func getBatchSpecWorkspaceQuery(opts *GetBatchSpecWorkspaceOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.deleted_at IS NULL"),
		sqlf.Sprintf("batch_spec_workspaces.id = %s", opts.ID),
	}

	return sqlf.Sprintf(
		getBatchSpecWorkspacesQueryFmtstr,
		sqlf.Join(BatchSpecWorkspaceColums.ToSqlf(), ", "),
		sqlf.Join(preds, "\n AND "),
	)
}

// ListBatchSpecWorkspacesOpts captures the query options needed for
// listing batch spec workspace jobs.
type ListBatchSpecWorkspacesOpts struct {
	LimitOpts
	Cursor      int64
	BatchSpecID int64
}

// ListBatchSpecWorkspaces lists batch changes with the given filters.
func (s *Store) ListBatchSpecWorkspaces(ctx context.Context, opts ListBatchSpecWorkspacesOpts) (cs []*btypes.BatchSpecWorkspace, next int64, err error) {
	ctx, endObservation := s.operations.listBatchSpecWorkspaces.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q := listBatchSpecWorkspacesQuery(opts)

	cs = make([]*btypes.BatchSpecWorkspace, 0)
	err = s.query(ctx, q, func(sc dbutil.Scanner) error {
		var c btypes.BatchSpecWorkspace
		if err := scanBatchSpecWorkspace(&c, sc); err != nil {
			return err
		}
		cs = append(cs, &c)
		return nil
	})

	if opts.Limit != 0 && len(cs) == opts.DBLimit() {
		next = cs[len(cs)-1].ID
		cs = cs[:len(cs)-1]
	}

	return cs, next, err
}

var listBatchSpecWorkspacesQueryFmtstr = `
-- source: enterprise/internal/batches/store/batch_spec_workspace_job.go:ListBatchSpecWorkspaces
SELECT %s FROM batch_spec_workspaces
INNER JOIN repo ON repo.id = batch_spec_workspaces.repo_id
WHERE %s
ORDER BY id ASC
`

func listBatchSpecWorkspacesQuery(opts ListBatchSpecWorkspacesOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.deleted_at IS NULL"),
	}

	if opts.BatchSpecID != 0 {
		preds = append(preds, sqlf.Sprintf("batch_spec_workspaces.batch_spec_id = %d", opts.BatchSpecID))
	}

	if opts.Cursor > 0 {
		preds = append(preds, sqlf.Sprintf("batch_spec_workspaces.id >= %s", opts.Cursor))
	}

	return sqlf.Sprintf(
		listBatchSpecWorkspacesQueryFmtstr+opts.LimitOpts.ToDB(),
		sqlf.Join(BatchSpecWorkspaceColums.ToSqlf(), ", "),
		sqlf.Join(preds, "\n AND "),
	)
}

const markSkippedBatchSpecWorkspacesQueryFmtstr = `
-- source: enterprise/internal/batches/store/batch_spec_workspaces.go:MarkSkippedBatchSpecWorkspaces
UPDATE
	batch_spec_workspaces
SET skipped = TRUE
FROM batch_specs
WHERE
	batch_spec_workspaces.batch_spec_id = %s
AND
    batch_specs.id = batch_spec_workspaces.batch_spec_id
AND NOT %s
`

// MarkSkippedBatchSpecWorkspaces marks the workspace that were skipped in
// CreateBatchSpecWorkspaceExecutionJobs as skipped.
func (s *Store) MarkSkippedBatchSpecWorkspaces(ctx context.Context, batchSpecID int64) (err error) {
	ctx, endObservation := s.operations.markSkippedBatchSpecWorkspaces.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("batchSpecID", int(batchSpecID)),
	}})
	defer endObservation(1, observation.Args{})

	q := sqlf.Sprintf(
		markSkippedBatchSpecWorkspacesQueryFmtstr,
		batchSpecID,
		sqlf.Sprintf(executableWorkspaceJobsConditionFmtstr),
	)
	return s.Exec(ctx, q)
}

func scanBatchSpecWorkspace(wj *btypes.BatchSpecWorkspace, s dbutil.Scanner) error {
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
		&wj.Unsupported,
		&wj.Ignored,
		&wj.Skipped,
		&wj.CreatedAt,
		&wj.UpdatedAt,
	); err != nil {
		return err
	}

	if err := json.Unmarshal(steps, &wj.Steps); err != nil {
		return errors.Wrap(err, "scanBatchSpecWorkspace: failed to unmarshal Steps")
	}

	return nil
}

func ScanFirstBatchSpecWorkspace(rows *sql.Rows, err error) (*btypes.BatchSpecWorkspace, bool, error) {
	jobs, err := scanBatchSpecWorkspaces(rows, err)
	if err != nil || len(jobs) == 0 {
		return nil, false, err
	}
	return jobs[0], true, nil
}

func scanBatchSpecWorkspaces(rows *sql.Rows, queryErr error) ([]*btypes.BatchSpecWorkspace, error) {
	if queryErr != nil {
		return nil, queryErr
	}

	var jobs []*btypes.BatchSpecWorkspace

	return jobs, scanAll(rows, func(sc dbutil.Scanner) (err error) {
		var j btypes.BatchSpecWorkspace
		if err = scanBatchSpecWorkspace(&j, sc); err != nil {
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
