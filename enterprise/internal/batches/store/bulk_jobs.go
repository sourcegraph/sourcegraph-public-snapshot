package store

import (
	"context"

	"github.com/keegancsmith/sqlf"

	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

var bulkJobColumns = []*sqlf.Query{
	sqlf.Sprintf("changeset_jobs.bulk_group AS id"),
	sqlf.Sprintf("MIN(changeset_jobs.id) AS db_id"),
	sqlf.Sprintf("changeset_jobs.job_type AS type"),
	sqlf.Sprintf(
		`CASE
	WHEN COUNT(*) FILTER (WHERE changeset_jobs.state IN (%s, %s, %s)) > 0 THEN 'PROCESSING'
	WHEN COUNT(*) FILTER (WHERE changeset_jobs.state = %s) > 0 THEN 'FAILED'
	ELSE 'COMPLETED'
END AS state`,
		btypes.ReconcilerStateProcessing.ToDB(),
		btypes.ReconcilerStateQueued.ToDB(),
		btypes.ReconcilerStateErrored.ToDB(),
		btypes.ReconcilerStateFailed.ToDB(),
	),
	sqlf.Sprintf(
		"CAST(COUNT(*) FILTER (WHERE changeset_jobs.state IN (%s, %s)) AS float) / CAST(COUNT(*) AS float) AS progress",
		btypes.ReconcilerStateCompleted.ToDB(),
		btypes.ReconcilerStateFailed.ToDB(),
	),
	sqlf.Sprintf("MIN(changeset_jobs.created_at) AS created_at"),
	sqlf.Sprintf(
		"CASE WHEN (COUNT(*) FILTER (WHERE changeset_jobs.state IN (%s, %s)) / COUNT(*)) = 1.0 THEN MAX(changeset_jobs.finished_at) ELSE null END AS finished_at",
		btypes.ReconcilerStateCompleted.ToDB(),
		btypes.ReconcilerStateFailed.ToDB(),
	),
}

// GetBulkJobOpts captures the query options needed for getting a BulkJob.
type GetBulkJobOpts struct {
	ID string
}

// GetBulkJob gets a BulkJob matching the given options.
func (s *Store) GetBulkJob(ctx context.Context, opts GetBulkJobOpts) (*btypes.BulkJob, error) {
	q := getBulkJobQuery(&opts)

	var c btypes.BulkJob
	err := s.query(ctx, q, func(sc scanner) (err error) {
		return scanBulkJob(&c, sc)
	})
	if err != nil {
		return nil, err
	}

	if c.ID == "" {
		return nil, ErrNoResults
	}

	return &c, nil
}

var getBulkJobsQueryFmtstr = `
-- source: enterprise/internal/batches/store/bulk_jobs.go:GetBulkJob
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

func getBulkJobQuery(opts *GetBulkJobOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.deleted_at IS NULL"),
		sqlf.Sprintf("changeset_jobs.bulk_group = %s", opts.ID),
	}

	return sqlf.Sprintf(
		getBulkJobsQueryFmtstr,
		sqlf.Join(bulkJobColumns, ","),
		sqlf.Join(preds, "\n AND "),
	)
}

// ListBulkJobsOpts captures the query options needed for getting a list of bulk jobs.
type ListBulkJobsOpts struct {
	LimitOpts
	Cursor int64

	BatchChangeID int64
}

// ListBulkJobs gets a list of BulkJobs matching the given options.
func (s *Store) ListBulkJobs(ctx context.Context, opts ListBulkJobsOpts) (bs []*btypes.BulkJob, next int64, err error) {
	q := listBulkJobsQuery(&opts)

	bs = make([]*btypes.BulkJob, 0, opts.DBLimit())
	err = s.query(ctx, q, func(sc scanner) error {
		var c btypes.BulkJob
		if err := scanBulkJob(&c, sc); err != nil {
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

var listBulkJobsQueryFmtstr = `
-- source: enterprise/internal/batches/store/bulk_jobs.go:ListBulkJobs
SELECT
    %s
FROM changeset_jobs
INNER JOIN changesets ON changesets.id = changeset_jobs.changeset_id
INNER JOIN repo ON repo.id = changesets.repo_id
WHERE
    %s
GROUP BY
    changeset_jobs.bulk_group, changeset_jobs.job_type
ORDER BY MIN(changeset_jobs.id) ASC
`

func listBulkJobsQuery(opts *ListBulkJobsOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.deleted_at IS NULL"),
		sqlf.Sprintf("changeset_jobs.batch_change_id = %s", opts.BatchChangeID),
	}

	if opts.Cursor > 0 {
		preds = append(preds, sqlf.Sprintf("changeset_jobs.id >= %s", opts.Cursor))
	}

	return sqlf.Sprintf(
		listBulkJobsQueryFmtstr+opts.LimitOpts.ToDB(),
		sqlf.Join(bulkJobColumns, ","),
		sqlf.Join(preds, "\n AND "),
	)
}

// CountBulkJobsOpts captures the query options needed counting BulkJobs.
type CountBulkJobsOpts struct {
	BatchChangeID int64
}

// CountBulkJobs gets the count of BulkJobs in the given batch change.
func (s *Store) CountBulkJobs(ctx context.Context, opts CountBulkJobsOpts) (int, error) {
	return s.queryCount(ctx, countBulkJobsQuery(&opts))
}

var countBulkJobsQueryFmtstr = `
-- source: enterprise/internal/batches/store/bulk_jobs.go:CountBulkJobs
SELECT
    COUNT(DISTINCT(changeset_jobs.bulk_group))
FROM changeset_jobs
INNER JOIN changesets ON changesets.id = changeset_jobs.changeset_id
INNER JOIN repo ON repo.id = changesets.repo_id
WHERE
    %s
`

func countBulkJobsQuery(opts *CountBulkJobsOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.deleted_at IS NULL"),
		sqlf.Sprintf("changeset_jobs.batch_change_id = %s", opts.BatchChangeID),
	}

	return sqlf.Sprintf(
		countBulkJobsQueryFmtstr,
		sqlf.Join(preds, "\n AND "),
	)
}

// ListBulkJobErrorsOpts captures the query options needed for getting a list of
// BulkJobErrors.
type ListBulkJobErrorsOpts struct {
	BulkJobID string
}

// ListBulkJobErrors gets a list of BulkJobErrors in a given BulkJob.
func (s *Store) ListBulkJobErrors(ctx context.Context, opts ListBulkJobErrorsOpts) (es []*btypes.BulkJobError, err error) {
	q := listBulkJobErrorsQuery(&opts)

	es = make([]*btypes.BulkJobError, 0)
	err = s.query(ctx, q, func(sc scanner) error {
		var c btypes.BulkJobError
		if err := scanBulkJobError(&c, sc); err != nil {
			return err
		}
		es = append(es, &c)
		return nil
	})

	return es, err
}

var listBulkJobErrorsQueryFmtstr = `
-- source: enterprise/internal/batches/store/bulk_jobs.go:ListBulkJobErrors
SELECT
    changeset_jobs.changeset_id AS changeset_id,
    changeset_jobs.failure_message AS error
FROM changeset_jobs
INNER JOIN changesets ON changesets.id = changeset_jobs.changeset_id
INNER JOIN repo ON repo.id = changesets.repo_id
WHERE
    %s
`

func listBulkJobErrorsQuery(opts *ListBulkJobErrorsOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.deleted_at IS NULL"),
		sqlf.Sprintf("changeset_jobs.failure_message IS NOT NULL"),
		sqlf.Sprintf("changeset_jobs.bulk_group = %s", opts.BulkJobID),
	}

	return sqlf.Sprintf(
		listBulkJobErrorsQueryFmtstr,
		sqlf.Join(preds, "\n AND "),
	)
}

func scanBulkJob(b *btypes.BulkJob, s scanner) error {
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

func scanBulkJobError(b *btypes.BulkJobError, s scanner) error {
	return s.Scan(
		&b.ChangesetID,
		&b.Error,
	)
}
