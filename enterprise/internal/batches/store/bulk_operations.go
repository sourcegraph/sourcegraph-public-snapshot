package store

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

var bulkOperationColumns = []*sqlf.Query{
	sqlf.Sprintf("changeset_jobs.bulk_group AS id"),
	sqlf.Sprintf("MIN(changeset_jobs.id) AS db_id"),
	sqlf.Sprintf("changeset_jobs.job_type AS type"),
	sqlf.Sprintf(
		`CASE
	WHEN COUNT(*) FILTER (WHERE changeset_jobs.state IN (%s, %s, %s)) > 0 THEN %s
	WHEN COUNT(*) FILTER (WHERE changeset_jobs.state = %s) > 0 THEN %s
	ELSE %s
END AS state`,
		btypes.ChangesetJobStateProcessing.ToDB(),
		btypes.ChangesetJobStateQueued.ToDB(),
		btypes.ChangesetJobStateErrored.ToDB(),
		btypes.BulkOperationStateProcessing,
		btypes.ChangesetJobStateFailed.ToDB(),
		btypes.BulkOperationStateFailed,
		btypes.BulkOperationStateCompleted,
	),
	sqlf.Sprintf(
		"CAST(COUNT(*) FILTER (WHERE changeset_jobs.state IN (%s, %s)) AS float) / CAST(COUNT(*) AS float) AS progress",
		btypes.ChangesetJobStateCompleted.ToDB(),
		btypes.ChangesetJobStateFailed.ToDB(),
	),
	sqlf.Sprintf("MIN(changeset_jobs.created_at) AS created_at"),
	sqlf.Sprintf(
		"CASE WHEN (COUNT(*) FILTER (WHERE changeset_jobs.state IN (%s, %s)) / COUNT(*)) = 1.0 THEN MAX(changeset_jobs.finished_at) ELSE null END AS finished_at",
		btypes.ChangesetJobStateCompleted.ToDB(),
		btypes.ChangesetJobStateFailed.ToDB(),
	),
}

// GetBulkOperationOpts captures the query options needed for getting a BulkOperation.
type GetBulkOperationOpts struct {
	ID string
}

// GetBulkOperation gets a BulkOperation matching the given options.
func (s *Store) GetBulkOperation(ctx context.Context, opts GetBulkOperationOpts) (*btypes.BulkOperation, error) {
	q := getBulkOperationQuery(&opts)

	var c btypes.BulkOperation
	err := s.query(ctx, q, func(sc scanner) (err error) {
		return scanBulkOperation(&c, sc)
	})
	if err != nil {
		return nil, err
	}

	if c.ID == "" {
		return nil, ErrNoResults
	}

	return &c, nil
}

var getBulkOperationQueryFmtstr = `
-- source: enterprise/internal/batches/store/bulk_operations.go:GetBulkOperation
SELECT
    %s
FROM changeset_jobs
INNER JOIN changesets ON changesets.id = changeset_jobs.changeset_id
INNER JOIN repo ON repo.id = changesets.repo_id
WHERE
    %s
GROUP BY
    changeset_jobs.bulk_group, changeset_jobs.job_type
LIMIT 1
`

func getBulkOperationQuery(opts *GetBulkOperationOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.deleted_at IS NULL"),
		sqlf.Sprintf("changeset_jobs.bulk_group = %s", opts.ID),
	}

	return sqlf.Sprintf(
		getBulkOperationQueryFmtstr,
		sqlf.Join(bulkOperationColumns, ","),
		sqlf.Join(preds, "\n AND "),
	)
}

// ListBulkOperationsOpts captures the query options needed for getting a list of bulk operations.
type ListBulkOperationsOpts struct {
	LimitOpts
	Cursor       int64
	States       []btypes.BulkOperationState
	CreatedAfter time.Time

	BatchChangeID int64
}

// ListBulkOperations gets a list of BulkOperations matching the given options.
func (s *Store) ListBulkOperations(ctx context.Context, opts ListBulkOperationsOpts) (bs []*btypes.BulkOperation, next int64, err error) {
	q := listBulkOperationsQuery(&opts)

	bs = make([]*btypes.BulkOperation, 0, opts.DBLimit())
	err = s.query(ctx, q, func(sc scanner) error {
		var c btypes.BulkOperation
		if err := scanBulkOperation(&c, sc); err != nil {
			return err
		}
		bs = append(bs, &c)
		return nil
	})

	if opts.Limit != 0 && len(bs) == opts.DBLimit() {
		next = bs[len(bs)-1].DBID
		bs = bs[:len(bs)-1]
	}

	return bs, next, err
}

var listBulkOperationsQueryFmtstr = `
-- source: enterprise/internal/batches/store/bulk_operations.go:ListBulkOperations
SELECT
    %s
FROM changeset_jobs
INNER JOIN changesets ON changesets.id = changeset_jobs.changeset_id
INNER JOIN repo ON repo.id = changesets.repo_id
WHERE
    %s
GROUP BY
    changeset_jobs.bulk_group, changeset_jobs.job_type
%s
ORDER BY MIN(changeset_jobs.id) DESC
`

func listBulkOperationsQuery(opts *ListBulkOperationsOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.deleted_at IS NULL"),
		sqlf.Sprintf("changeset_jobs.batch_change_id = %s", opts.BatchChangeID),
	}
	having := sqlf.Sprintf("")

	if opts.Cursor > 0 {
		preds = append(preds, sqlf.Sprintf("changeset_jobs.id >= %s", opts.Cursor))
	}

	// TODO: Broken logic. Needs a view.
	if len(opts.States) > 0 {
		states := make([]*sqlf.Query, 0)
		for _, state := range opts.States {
			if state == btypes.BulkOperationStateProcessing {
				states = append(states, sqlf.Sprintf("%s", btypes.ChangesetJobStateProcessing.ToDB()))
				states = append(states, sqlf.Sprintf("%s", btypes.ChangesetJobStateQueued.ToDB()))
				states = append(states, sqlf.Sprintf("%s", btypes.ChangesetJobStateErrored.ToDB()))
			} else if state == btypes.BulkOperationStateFailed {
				states = append(states, sqlf.Sprintf("%s", btypes.ChangesetJobStateFailed.ToDB()))
			}
		}
		preds = append(preds, sqlf.Sprintf("changeset_jobs.state IN (%s)", sqlf.Join(states, ",")))
	}

	if !opts.CreatedAfter.IsZero() {
		having = sqlf.Sprintf("HAVING MIN(changeset_jobs.created_at) >= %s", opts.CreatedAfter)
	}

	return sqlf.Sprintf(
		listBulkOperationsQueryFmtstr+opts.LimitOpts.ToDB(),
		sqlf.Join(bulkOperationColumns, ","),
		sqlf.Join(preds, "\n AND "),
		having,
	)
}

// CountBulkOperationsOpts captures the query options needed when counting BulkOperations.
type CountBulkOperationsOpts struct {
	CreatedAfter  time.Time
	BatchChangeID int64
}

// CountBulkOperations gets the count of BulkOperations in the given batch change.
func (s *Store) CountBulkOperations(ctx context.Context, opts CountBulkOperationsOpts) (int, error) {
	return s.queryCount(ctx, countBulkOperationsQuery(&opts))
}

var countBulkOperationsQueryFmtstr = `
-- source: enterprise/internal/batches/store/bulk_operations.go:CountBulkOperations
SELECT
    COUNT(DISTINCT(changeset_jobs.bulk_group))
FROM changeset_jobs
INNER JOIN changesets ON changesets.id = changeset_jobs.changeset_id
INNER JOIN repo ON repo.id = changesets.repo_id
WHERE
    %s
GROUP BY changeset_jobs.bulk_group
%s
`

func countBulkOperationsQuery(opts *CountBulkOperationsOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.deleted_at IS NULL"),
		sqlf.Sprintf("changeset_jobs.batch_change_id = %s", opts.BatchChangeID),
	}
	having := &sqlf.Query{}

	if !opts.CreatedAfter.IsZero() {
		having = sqlf.Sprintf("HAVING MIN(changeset_jobs.created_at) >= %s", opts.CreatedAfter)
	}

	return sqlf.Sprintf(
		countBulkOperationsQueryFmtstr,
		sqlf.Join(preds, "\n AND "),
		having,
	)
}

// ListBulkOperationErrorsOpts captures the query options needed for getting a list of
// BulkOperationErrors.
type ListBulkOperationErrorsOpts struct {
	BulkOperationID string
}

// ListBulkOperationErrors gets a list of BulkOperationErrors in a given BulkOperation.
func (s *Store) ListBulkOperationErrors(ctx context.Context, opts ListBulkOperationErrorsOpts) (es []*btypes.BulkOperationError, err error) {
	q := listBulkOperationErrorsQuery(&opts)

	es = make([]*btypes.BulkOperationError, 0)
	err = s.query(ctx, q, func(sc scanner) error {
		var c btypes.BulkOperationError
		if err := scanBulkOperationError(&c, sc); err != nil {
			return err
		}
		es = append(es, &c)
		return nil
	})

	return es, err
}

var listBulkOperationErrorsQueryFmtstr = `
-- source: enterprise/internal/batches/store/bulk_operations.go:ListBulkOperationErrors
SELECT
    changeset_jobs.changeset_id AS changeset_id,
    changeset_jobs.failure_message AS error
FROM changeset_jobs
INNER JOIN changesets ON changesets.id = changeset_jobs.changeset_id
INNER JOIN repo ON repo.id = changesets.repo_id
WHERE
    %s
`

func listBulkOperationErrorsQuery(opts *ListBulkOperationErrorsOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.deleted_at IS NULL"),
		sqlf.Sprintf("changeset_jobs.failure_message IS NOT NULL"),
		sqlf.Sprintf("changeset_jobs.bulk_group = %s", opts.BulkOperationID),
	}

	return sqlf.Sprintf(
		listBulkOperationErrorsQueryFmtstr,
		sqlf.Join(preds, "\n AND "),
	)
}

func scanBulkOperation(b *btypes.BulkOperation, s scanner) error {
	return s.Scan(
		&b.ID,
		&b.DBID,
		&b.Type,
		&b.State,
		&b.Progress,
		&b.CreatedAt,
		&dbutil.NullTime{Time: &b.FinishedAt},
	)
}

func scanBulkOperationError(b *btypes.BulkOperationError, s scanner) error {
	return s.Scan(
		&b.ChangesetID,
		&b.Error,
	)
}
