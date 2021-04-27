package store

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

// BulkJobColumns are used by the batch change related Store methods to insert,
// update and query changeset jobs.
var BulkJobColumns = []*sqlf.Query{
	sqlf.Sprintf("changeset_jobs.id"),
	sqlf.Sprintf("changeset_jobs.bulk_group"),
	sqlf.Sprintf("changeset_jobs.user_id"),
	sqlf.Sprintf("changeset_jobs.batch_change_id"),
	sqlf.Sprintf("changeset_jobs.changeset_id"),
	sqlf.Sprintf("changeset_jobs.job_type"),
	sqlf.Sprintf("changeset_jobs.payload"),
	sqlf.Sprintf("changeset_jobs.state"),
	sqlf.Sprintf("changeset_jobs.failure_message"),
	sqlf.Sprintf("changeset_jobs.started_at"),
	sqlf.Sprintf("changeset_jobs.finished_at"),
	sqlf.Sprintf("changeset_jobs.process_after"),
	sqlf.Sprintf("changeset_jobs.num_resets"),
	sqlf.Sprintf("changeset_jobs.num_failures"),
	sqlf.Sprintf("changeset_jobs.created_at"),
	sqlf.Sprintf("changeset_jobs.updated_at"),
}

// GetBulkJobOpts captures the query options needed for getting a ChangesetJob
type GetBulkJobOpts struct {
	ID string
}

// GetBulkJob gets a ChangesetJob matching the given options.
func (s *Store) GetBulkJob(ctx context.Context, opts GetBulkJobOpts) (*BulkJob, error) {
	q := getBulkJobQuery(&opts)

	var c BulkJob
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
-- source: enterprise/internal/batches/store/changeset_jobs.go:GetBulkJob
SELECT
    changeset_jobs.bulk_group AS id,
    changeset_jobs.job_type AS type,
    CASE
    	WHEN COUNT(*) FILTER (WHERE changeset_jobs.state IN (%s, %s, %s)) > 0 THEN 'PROCESSING'
		WHEN COUNT(*) FILTER (WHERE changeset_jobs.state = %s) > 0 THEN 'FAILED'
        ELSE 'COMPLETED'
    END AS state,
	COUNT(*) FILTER (WHERE changeset_jobs.state IN (%s, %s)) / COUNT(*) AS progress,
	MIN(changeset_jobs.created_at) AS created_at,
	CASE WHEN (COUNT(*) FILTER (WHERE changeset_jobs.state IN (%s, %s)) / COUNT(*)) = 1.0 THEN MAX(changeset_jobs.finished_at) ELSE null END AS finished_at
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
		btypes.ReconcilerStateProcessing.ToDB(),
		btypes.ReconcilerStateQueued.ToDB(),
		btypes.ReconcilerStateErrored.ToDB(),
		btypes.ReconcilerStateFailed.ToDB(),
		btypes.ReconcilerStateCompleted.ToDB(),
		btypes.ReconcilerStateFailed.ToDB(),
		btypes.ReconcilerStateCompleted.ToDB(),
		btypes.ReconcilerStateFailed.ToDB(),
		sqlf.Join(preds, "\n AND "),
	)
}

// ListBulkJobsOpts captures the query options needed for getting a ChangesetJob
type ListBulkJobsOpts struct {
	LimitOpts
	Cursor *string

	BatchChangeID int64
}

// ListBulkJobs gets a ChangesetJob matching the given options.
func (s *Store) ListBulkJobs(ctx context.Context, opts ListBulkJobsOpts) (cs []*BulkJob, next string, err error) {
	q := listBulkJobsQuery(&opts)

	cs = make([]*BulkJob, 0, opts.DBLimit())
	err = s.query(ctx, q, func(sc scanner) error {
		var c BulkJob
		if err := scanBulkJob(&c, sc); err != nil {
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

type BulkJob struct {
	ID         string
	Type       btypes.ChangesetJobType
	State      btypes.ReconcilerState
	Progress   float64
	CreatedAt  time.Time
	FinishedAt time.Time
}

var listBulkJobsQueryFmtstr = `
-- source: enterprise/internal/batches/store/changeset_jobs.go:ListBulkJobs
SELECT
    changeset_jobs.bulk_group AS id,
    changeset_jobs.job_type AS type,
    CASE
    	WHEN COUNT(*) FILTER (WHERE changeset_jobs.state IN (%s, %s, %s)) > 0 THEN 'PROCESSING'
		WHEN COUNT(*) FILTER (WHERE changeset_jobs.state = %s) > 0 THEN 'FAILED'
        ELSE 'COMPLETED'
    END AS state,
	COUNT(*) FILTER (WHERE changeset_jobs.state IN (%s, %s)) / COUNT(*) AS progress,
	MIN(changeset_jobs.created_at) AS created_at,
	CASE WHEN (COUNT(*) FILTER (WHERE changeset_jobs.state IN (%s, %s)) / COUNT(*)) = 1.0 THEN MAX(changeset_jobs.finished_at) ELSE null END AS finished_at
FROM changeset_jobs
INNER JOIN changesets ON changesets.id = changeset_jobs.changeset_id
INNER JOIN repo ON repo.id = changesets.repo_id
WHERE
    %s
GROUP BY
    changeset_jobs.bulk_group, changeset_jobs.job_type
ORDER BY id ASC
`

func listBulkJobsQuery(opts *ListBulkJobsOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.deleted_at IS NULL"),
		sqlf.Sprintf("changeset_jobs.batch_change_id = %s", opts.BatchChangeID),
	}

	if opts.Cursor != nil {
		preds = append(preds, sqlf.Sprintf("changeset_jobs.bulk_group > %s", *opts.Cursor))
	}

	return sqlf.Sprintf(
		listBulkJobsQueryFmtstr+opts.LimitOpts.ToDB(),
		btypes.ReconcilerStateProcessing.ToDB(),
		btypes.ReconcilerStateQueued.ToDB(),
		btypes.ReconcilerStateErrored.ToDB(),
		btypes.ReconcilerStateFailed.ToDB(),
		btypes.ReconcilerStateCompleted.ToDB(),
		btypes.ReconcilerStateFailed.ToDB(),
		btypes.ReconcilerStateCompleted.ToDB(),
		btypes.ReconcilerStateFailed.ToDB(),
		sqlf.Join(preds, "\n AND "),
	)
}

// CountBulkJobsOpts captures the query options needed for getting a ChangesetJob
type CountBulkJobsOpts struct {
	BatchChangeID int64
}

// CountBulkJobs gets a ChangesetJob matching the given options.
func (s *Store) CountBulkJobs(ctx context.Context, opts CountBulkJobsOpts) (int, error) {
	return s.queryCount(ctx, countBulkJobsQuery(&opts))
}

var countBulkJobsQueryFmtstr = `
-- source: enterprise/internal/batches/store/changeset_jobs.go:CountBulkJobs
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

type BulkJobError struct {
	ChangesetID int64
	Error       string
}

// ListBulkJobErrorsOpts captures the query options needed for getting a ChangesetJob
type ListBulkJobErrorsOpts struct {
	BulkJobID string
}

// ListBulkJobErrors gets a ChangesetJob matching the given options.
func (s *Store) ListBulkJobErrors(ctx context.Context, opts ListBulkJobErrorsOpts) (cs []*BulkJobError, err error) {
	q := listBulkJobErrorsQuery(&opts)

	cs = make([]*BulkJobError, 0)
	err = s.query(ctx, q, func(sc scanner) error {
		var c BulkJobError
		if err := scanBulkJobError(&c, sc); err != nil {
			return err
		}
		cs = append(cs, &c)
		return nil
	})

	return cs, err
}

var listBulkJobErrorsQueryFmtstr = `
-- source: enterprise/internal/batches/store/changeset_jobs.go:ListBulkJobErrors
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

func scanBulkJob(b *BulkJob, s scanner) error {
	return s.Scan(
		&b.ID,
		&b.Type,
		&b.State,
		&b.Progress,
		&b.CreatedAt,
		&dbutil.NullTime{Time: &b.FinishedAt},
	)
}

func scanBulkJobError(b *BulkJobError, s scanner) error {
	return s.Scan(
		&b.ChangesetID,
		&b.Error,
	)
}
