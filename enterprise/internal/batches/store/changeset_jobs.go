package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type ChangesetJobType string

var (
	ChangesetJobTypeComment ChangesetJobType = "commentatore"
)

type ChangesetJobCommentPayload struct {
	Message string `json:"message"`
}

type ChangesetJob struct {
	ID            int64
	BulkGroup     string
	BatchChangeID int64
	UserID        int32
	ChangesetID   int64
	JobType       ChangesetJobType
	Payload       interface{}

	State          string
	FailureMessage *string
	StartedAt      time.Time
	FinishedAt     time.Time
	ProcessAfter   time.Time
	NumResets      int64
	NumFailures    int64

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (j *ChangesetJob) RecordID() int {
	return int(j.ID)
}

// changesetJobInsertColumns is the list of batch changes columns that are
// modified in CreateBatchChange and UpdateBatchChange.
// update and query batches.
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
// update and query batches.
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

// CreateBatchChange creates the given batch change.
func (s *Store) CreateChangesetJobs(ctx context.Context, c *ChangesetJob) error {
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

func createChangesetJobQuery(c *ChangesetJob) (*sqlf.Query, error) {
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

func scanChangesetJob(c *ChangesetJob, s scanner) error {
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
	case ChangesetJobTypeComment:
		c.Payload = new(ChangesetJobCommentPayload)
	default:
		return fmt.Errorf("unknown job type %q", c.JobType)
	}
	return json.Unmarshal(raw, &c.Payload)
}

func ScanFirstChangesetJob(rows *sql.Rows, err error) (*ChangesetJob, bool, error) {
	jobs, err := scanChangesetJobs(rows, err)
	if err != nil || len(jobs) == 0 {
		return &ChangesetJob{}, false, err
	}
	return jobs[0], true, nil
}

func scanChangesetJobs(rows *sql.Rows, queryErr error) ([]*ChangesetJob, error) {
	if queryErr != nil {
		return nil, queryErr
	}

	var jobs []*ChangesetJob

	return jobs, scanAll(rows, func(sc scanner) (err error) {
		var j ChangesetJob
		if err = scanChangesetJob(&j, sc); err != nil {
			return err
		}
		jobs = append(jobs, &j)
		return nil
	})
}
