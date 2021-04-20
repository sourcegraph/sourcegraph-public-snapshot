package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/keegancsmith/sqlf"

	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

// changesetJobInsertColumns is the list of changeset_jobs columns that are
// modified in CreateChangesetJob and UpdateChangesetJob.
var changesetJobInsertColumns = []*sqlf.Query{
	sqlf.Sprintf("bulk_group"),
	sqlf.Sprintf("user_id"),
	sqlf.Sprintf("batch_change_id"),
	sqlf.Sprintf("changeset_id"),
	sqlf.Sprintf("job_type"),
	sqlf.Sprintf("payload"),
	sqlf.Sprintf("state"),
	sqlf.Sprintf("failure_message"),
	sqlf.Sprintf("started_at"),
	sqlf.Sprintf("finished_at"),
	sqlf.Sprintf("process_after"),
	sqlf.Sprintf("num_resets"),
	sqlf.Sprintf("num_failures"),
	sqlf.Sprintf("created_at"),
	sqlf.Sprintf("updated_at"),
}

// ChangesetJobColumns are used by the batch change related Store methods to insert,
// update and query changeset jobs.
var ChangesetJobColumns = []*sqlf.Query{
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

// CreateChangesetJob creates the given changeset job.
func (s *Store) CreateChangesetJob(ctx context.Context, c *btypes.ChangesetJob) error {
	if c.CreatedAt.IsZero() {
		c.CreatedAt = s.now()
	}

	if c.UpdatedAt.IsZero() {
		c.UpdatedAt = c.CreatedAt
	}

	q, err := createChangesetJobQuery(c)
	if err != nil {
		return err
	}

	return s.query(ctx, q, func(sc scanner) (err error) {
		return scanChangesetJob(c, sc)
	})
}

var createChangesetJobQueryFmtstr = `
-- source: enterprise/internal/batches/store/changeset_jobs.go:CreateChangesetJob
INSERT INTO changeset_jobs (%s)
VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
RETURNING %s
`

func createChangesetJobQuery(c *btypes.ChangesetJob) (*sqlf.Query, error) {
	payload, err := jsonbColumn(c.Payload)
	if err != nil {
		return nil, err
	}
	return sqlf.Sprintf(
		createChangesetJobQueryFmtstr,
		sqlf.Join(changesetJobInsertColumns, ", "),
		c.BulkGroup,
		c.UserID,
		c.BatchChangeID,
		c.ChangesetID,
		c.JobType,
		payload,
		c.State,
		c.FailureMessage,
		&dbutil.NullTime{Time: &c.StartedAt},
		&dbutil.NullTime{Time: &c.FinishedAt},
		&dbutil.NullTime{Time: &c.ProcessAfter},
		c.NumResets,
		c.NumFailures,
		c.CreatedAt,
		c.UpdatedAt,
		sqlf.Join(ChangesetJobColumns, ", "),
	), nil
}

// GetChangesetJobOpts captures the query options needed for getting a ChangesetJob
type GetChangesetJobOpts struct {
	ID int64
}

// GetChangesetJob gets a ChangesetJob matching the given options.
func (s *Store) GetChangesetJob(ctx context.Context, opts GetChangesetJobOpts) (*btypes.ChangesetJob, error) {
	q := getChangesetJobQuery(&opts)

	var c btypes.ChangesetJob
	err := s.query(ctx, q, func(sc scanner) (err error) {
		return scanChangesetJob(&c, sc)
	})
	if err != nil {
		return nil, err
	}

	if c.ID == 0 {
		return nil, ErrNoResults
	}

	return &c, nil
}

var getChangesetJobsQueryFmtstr = `
-- source: enterprise/internal/batches/store/changeset_jobs.go:GetChangesetJob
SELECT %s FROM changeset_jobs
INNER JOIN changesets ON changesets.id = changeset_jobs.changeset_id
INNER JOIN repo ON repo.id = changesets.repo_id
WHERE %s
LIMIT 1
`

func getChangesetJobQuery(opts *GetChangesetJobOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.deleted_at IS NULL"),
		sqlf.Sprintf("changeset_jobs.id = %s", opts.ID),
	}

	return sqlf.Sprintf(
		getChangesetJobsQueryFmtstr,
		sqlf.Join(ChangesetJobColumns, ", "),
		sqlf.Join(preds, "\n AND "),
	)
}

func scanChangesetJob(c *btypes.ChangesetJob, s scanner) error {
	var raw json.RawMessage
	if err := s.Scan(
		&c.ID,
		&c.BulkGroup,
		&c.UserID,
		&c.BatchChangeID,
		&c.ChangesetID,
		&c.JobType,
		&raw,
		&c.State,
		&dbutil.NullString{S: c.FailureMessage},
		&dbutil.NullTime{Time: &c.StartedAt},
		&dbutil.NullTime{Time: &c.FinishedAt},
		&dbutil.NullTime{Time: &c.ProcessAfter},
		&c.NumResets,
		&c.NumFailures,
		&c.CreatedAt,
		&c.UpdatedAt,
	); err != nil {
		return err
	}
	switch c.JobType {
	case btypes.ChangesetJobTypeComment:
		c.Payload = new(btypes.ChangesetJobCommentPayload)
	default:
		return fmt.Errorf("unknown job type %q", c.JobType)
	}
	return json.Unmarshal(raw, &c.Payload)
}

func ScanFirstChangesetJob(rows *sql.Rows, err error) (*btypes.ChangesetJob, bool, error) {
	jobs, err := scanChangesetJobs(rows, err)
	if err != nil || len(jobs) == 0 {
		return nil, false, err
	}
	return jobs[0], true, nil
}

func scanChangesetJobs(rows *sql.Rows, queryErr error) ([]*btypes.ChangesetJob, error) {
	if queryErr != nil {
		return nil, queryErr
	}

	var jobs []*btypes.ChangesetJob

	return jobs, scanAll(rows, func(sc scanner) (err error) {
		var j btypes.ChangesetJob
		if err = scanChangesetJob(&j, sc); err != nil {
			return err
		}
		jobs = append(jobs, &j)
		return nil
	})
}
