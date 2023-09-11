package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"sort"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/batches/search"
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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
	"unsupported",
	"ignored",
	"skipped",
	"cached_result_found",
	"step_cache_results",

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
	"batch_spec_workspaces.unsupported",
	"batch_spec_workspaces.ignored",
	"batch_spec_workspaces.skipped",
	"batch_spec_workspaces.cached_result_found",
	"batch_spec_workspaces.step_cache_results",

	"batch_spec_workspaces.created_at",
	"batch_spec_workspaces.updated_at",
}

// CreateBatchSpecWorkspace creates the given batch spec workspace jobs.
func (s *Store) CreateBatchSpecWorkspace(ctx context.Context, ws ...*btypes.BatchSpecWorkspace) (err error) {
	ctx, _, endObservation := s.operations.createBatchSpecWorkspace.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("count", len(ws)),
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

			marshaledStepCacheResults, err := json.Marshal(wj.StepCacheResults)
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
				wj.Unsupported,
				wj.Ignored,
				wj.Skipped,
				wj.CachedResultFound,
				marshaledStepCacheResults,
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
		s.Handle(),
		"batch_spec_workspaces",
		batch.MaxNumPostgresParameters,
		batchSpecWorkspaceInsertColumns,
		"",
		BatchSpecWorkspaceColums,
		func(rows dbutil.Scanner) error {
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
	ctx, _, endObservation := s.operations.getBatchSpecWorkspace.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("ID", int(opts.ID)),
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
	IDs         []int64

	State                            btypes.BatchSpecWorkspaceExecutionJobState
	OnlyWithoutExecutionAndNotCached bool
	OnlyCachedOrCompleted            bool
	Cancel                           *bool
	Skipped                          *bool
	TextSearch                       []search.TextSearchTerm
}

func (opts ListBatchSpecWorkspacesOpts) SQLConds(ctx context.Context, db database.DB, forCount bool) (where *sqlf.Query, joinStatements *sqlf.Query, err error) {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.deleted_at IS NULL"),
	}
	joins := []*sqlf.Query{}

	if len(opts.IDs) != 0 {
		preds = append(preds, sqlf.Sprintf("batch_spec_workspaces.id = ANY(%s)", pq.Array(opts.IDs)))
	}

	if opts.BatchSpecID != 0 {
		preds = append(preds, sqlf.Sprintf("batch_spec_workspaces.batch_spec_id = %d", opts.BatchSpecID))
	}

	if !forCount && opts.Cursor > 0 {
		preds = append(preds, sqlf.Sprintf("batch_spec_workspaces.id >= %s", opts.Cursor))
	}

	joinedExecution := false
	ensureJoinExecution := func() {
		if joinedExecution {
			return
		}
		joins = append(joins, sqlf.Sprintf("LEFT JOIN batch_spec_workspace_execution_jobs ON batch_spec_workspace_execution_jobs.batch_spec_workspace_id = batch_spec_workspaces.id"))
		joinedExecution = true
	}

	if opts.State != "" {
		ensureJoinExecution()
		preds = append(preds, sqlf.Sprintf("batch_spec_workspace_execution_jobs.state = %s", opts.State))
	}

	if opts.OnlyWithoutExecutionAndNotCached {
		ensureJoinExecution()
		preds = append(preds, sqlf.Sprintf("batch_spec_workspace_execution_jobs.id IS NULL AND NOT batch_spec_workspaces.cached_result_found"))
	}

	if opts.OnlyCachedOrCompleted {
		ensureJoinExecution()
		preds = append(preds, sqlf.Sprintf("(batch_spec_workspaces.cached_result_found OR batch_spec_workspace_execution_jobs.state = %s)", btypes.BatchSpecWorkspaceExecutionJobStateCompleted))
	}

	if opts.Cancel != nil {
		ensureJoinExecution()
		preds = append(preds, sqlf.Sprintf("batch_spec_workspace_execution_jobs.cancel = %s", *opts.Cancel))
	}

	if opts.Skipped != nil {
		preds = append(preds, sqlf.Sprintf("batch_spec_workspaces.skipped = %s", *opts.Skipped))
	}

	if len(opts.TextSearch) != 0 {
		for _, term := range opts.TextSearch {
			preds = append(preds, textSearchTermToClause(
				term,
				// TODO: Add more terms here later.
				sqlf.Sprintf("repo.name"),
			))

			// If we do text-search, we need to only consider workspaces in repos that are visible to the user.
			// Otherwise we would leak the existance of repos.

			repoAuthzConds, err := database.AuthzQueryConds(ctx, db)
			if err != nil {
				return nil, nil, errors.Wrap(err, "ListBatchSpecWorkspacesOpts.SQLConds generating authz query conds")
			}

			preds = append(preds, repoAuthzConds)
		}
	}

	return sqlf.Join(preds, "\n AND "), sqlf.Join(joins, "\n"), nil
}

// ListBatchSpecWorkspaces lists batch spec workspaces with the given filters.
func (s *Store) ListBatchSpecWorkspaces(ctx context.Context, opts ListBatchSpecWorkspacesOpts) (cs []*btypes.BatchSpecWorkspace, next int64, err error) {
	ctx, _, endObservation := s.operations.listBatchSpecWorkspaces.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q, err := listBatchSpecWorkspacesQuery(ctx, s.DatabaseDB(), opts)
	if err != nil {
		return nil, 0, err
	}

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
SELECT %s FROM batch_spec_workspaces
INNER JOIN repo ON repo.id = batch_spec_workspaces.repo_id
%s
WHERE %s
ORDER BY id ASC
`

func listBatchSpecWorkspacesQuery(ctx context.Context, db database.DB, opts ListBatchSpecWorkspacesOpts) (*sqlf.Query, error) {
	where, joins, err := opts.SQLConds(ctx, db, false)
	if err != nil {
		return nil, err
	}
	return sqlf.Sprintf(
		listBatchSpecWorkspacesQueryFmtstr+opts.LimitOpts.ToDB(),
		sqlf.Join(BatchSpecWorkspaceColums.ToSqlf(), ", "),
		joins,
		where,
	), nil
}

// CountBatchSpecWorkspaces counts batch spec workspaces with the given filters.
func (s *Store) CountBatchSpecWorkspaces(ctx context.Context, opts ListBatchSpecWorkspacesOpts) (count int64, err error) {
	ctx, _, endObservation := s.operations.countBatchSpecWorkspaces.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q, err := countBatchSpecWorkspacesQuery(ctx, s.DatabaseDB(), opts)
	if err != nil {
		return 0, err
	}

	count, _, err = basestore.ScanFirstInt64(s.Query(ctx, q))
	return count, err
}

var countBatchSpecWorkspacesQueryFmtstr = `
SELECT
	COUNT(1)
FROM
	batch_spec_workspaces
INNER JOIN repo ON repo.id = batch_spec_workspaces.repo_id
%s
WHERE %s
`

func countBatchSpecWorkspacesQuery(ctx context.Context, db database.DB, opts ListBatchSpecWorkspacesOpts) (*sqlf.Query, error) {
	where, joins, err := opts.SQLConds(ctx, db, true)
	if err != nil {
		return nil, err
	}

	return sqlf.Sprintf(
		countBatchSpecWorkspacesQueryFmtstr+opts.LimitOpts.ToDB(),
		joins,
		where,
	), nil
}

const markSkippedBatchSpecWorkspacesQueryFmtstr = `
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
	ctx, _, endObservation := s.operations.markSkippedBatchSpecWorkspaces.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("batchSpecID", int(batchSpecID)),
	}})
	defer endObservation(1, observation.Args{})

	q := sqlf.Sprintf(
		markSkippedBatchSpecWorkspacesQueryFmtstr,
		batchSpecID,
		sqlf.Sprintf(executableWorkspaceJobsConditionFmtstr),
	)
	return s.Exec(ctx, q)
}

// ListRetryBatchSpecWorkspacesOpts options to determine which btypes.BatchSpecWorkspace to retrieve for retrying.
type ListRetryBatchSpecWorkspacesOpts struct {
	BatchSpecID      int64
	IncludeCompleted bool
}

// ListRetryBatchSpecWorkspaces lists all btypes.BatchSpecWorkspace to retry.
func (s *Store) ListRetryBatchSpecWorkspaces(ctx context.Context, opts ListRetryBatchSpecWorkspacesOpts) (cs []*btypes.BatchSpecWorkspace, err error) {
	ctx, _, endObservation := s.operations.listRetryBatchSpecWorkspaces.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q := getListRetryBatchSpecWorkspacesQuery(&opts)
	cs = make([]*btypes.BatchSpecWorkspace, 0)
	err = s.query(ctx, q, func(sc dbutil.Scanner) error {
		var c btypes.BatchSpecWorkspace
		if err := sc.Scan(
			&c.ID,
			&jsonIDsSet{Assocs: &c.ChangesetSpecIDs},
		); err != nil {
			return err
		}
		cs = append(cs, &c)
		return nil
	})

	return cs, err
}

func getListRetryBatchSpecWorkspacesQuery(opts *ListRetryBatchSpecWorkspacesOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.deleted_at IS NULL"),
		sqlf.Sprintf("batch_spec_workspaces.batch_spec_id = %s", opts.BatchSpecID),
	}

	if !opts.IncludeCompleted {
		preds = append(preds, sqlf.Sprintf("batch_spec_workspace_execution_jobs.state != %s", btypes.BatchSpecWorkspaceExecutionJobStateCompleted))
	}

	return sqlf.Sprintf(
		listRetryBatchSpecWorkspacesFmtstr,
		sqlf.Join(preds, "\n AND "),
	)
}

const listRetryBatchSpecWorkspacesFmtstr = `
SELECT batch_spec_workspaces.id, batch_spec_workspaces.changeset_spec_ids
FROM batch_spec_workspaces
		 INNER JOIN repo ON repo.id = batch_spec_workspaces.repo_id
		 INNER JOIN batch_spec_workspace_execution_jobs
					ON batch_spec_workspaces.id = batch_spec_workspace_execution_jobs.batch_spec_workspace_id
WHERE %s
`

const disableBatchSpecWorkspaceExecutionCacheQueryFmtstr = `
WITH batch_spec AS (
	SELECT
		id
	FROM
		batch_specs
	WHERE
		id = %s
		AND
		no_cache
),
candidate_batch_spec_workspaces AS (
	SELECT
		id, changeset_spec_ids
	FROM
		batch_spec_workspaces
	WHERE
		batch_spec_workspaces.batch_spec_id = (SELECT id FROM batch_spec)
	ORDER BY id
),
removable_changeset_specs AS (
	SELECT
		id
	FROM
		changeset_specs
	WHERE
		id IN (SELECT jsonb_object_keys(changeset_spec_ids)::bigint FROM candidate_batch_spec_workspaces)
	ORDER BY id
),
removed_changeset_specs AS (
	DELETE FROM changeset_specs
	WHERE
		id IN (SELECT id FROM removable_changeset_specs)
)
UPDATE
	batch_spec_workspaces
SET
	cached_result_found = FALSE,
	changeset_spec_ids = '{}',
	step_cache_results = '{}'
WHERE
	id IN (SELECT id FROM candidate_batch_spec_workspaces)
`

// DisableBatchSpecWorkspaceExecutionCache removes caching information from workspaces prior to execution.
func (s *Store) DisableBatchSpecWorkspaceExecutionCache(ctx context.Context, batchSpecID int64) (err error) {
	ctx, _, endObservation := s.operations.disableBatchSpecWorkspaceExecutionCache.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("batchSpecID", int(batchSpecID)),
	}})
	defer endObservation(1, observation.Args{})

	q := sqlf.Sprintf(disableBatchSpecWorkspaceExecutionCacheQueryFmtstr, batchSpecID)
	return s.Exec(ctx, q)
}

func scanBatchSpecWorkspace(wj *btypes.BatchSpecWorkspace, s dbutil.Scanner) error {
	var stepCacheResults json.RawMessage

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
		&wj.Unsupported,
		&wj.Ignored,
		&wj.Skipped,
		&wj.CachedResultFound,
		&stepCacheResults,
		&wj.CreatedAt,
		&wj.UpdatedAt,
	); err != nil {
		return err
	}

	if err := json.Unmarshal(stepCacheResults, &wj.StepCacheResults); err != nil {
		return errors.Wrap(err, "scanBatchSpecWorkspace: failed to unmarshal StepCacheResults")
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
func (n *jsonIDsSet) Scan(value any) error {
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
