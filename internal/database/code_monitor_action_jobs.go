package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

type ActionJob struct {
	ID           int32
	Email        *int64
	Webhook      *int64
	SlackWebhook *int64
	TriggerEvent int32

	// Fields demanded by any dbworker.
	State          string
	FailureMessage *string
	StartedAt      *time.Time
	FinishedAt     *time.Time
	ProcessAfter   *time.Time
	NumResets      int32
	NumFailures    int32
	LogContents    *string
}

func (a *ActionJob) RecordID() int {
	return int(a.ID)
}

func (a *ActionJob) RecordUID() string {
	return strconv.FormatInt(int64(a.ID), 10)
}

type ActionJobMetadata struct {
	Description string
	MonitorID   int64
	Results     []*result.CommitMatch
	OwnerName   string

	// The query with after: filter.
	Query string
}

// ActionJobColumns is the list of db columns used to populate an ActionJob struct.
// This must stay in sync with the scanned columns in scanActionJob.
var ActionJobColumns = []*sqlf.Query{
	sqlf.Sprintf("cm_action_jobs.id"),
	sqlf.Sprintf("cm_action_jobs.email"),
	sqlf.Sprintf("cm_action_jobs.webhook"),
	sqlf.Sprintf("cm_action_jobs.slack_webhook"),
	sqlf.Sprintf("cm_action_jobs.trigger_event"),
	sqlf.Sprintf("cm_action_jobs.state"),
	sqlf.Sprintf("cm_action_jobs.failure_message"),
	sqlf.Sprintf("cm_action_jobs.started_at"),
	sqlf.Sprintf("cm_action_jobs.finished_at"),
	sqlf.Sprintf("cm_action_jobs.process_after"),
	sqlf.Sprintf("cm_action_jobs.num_resets"),
	sqlf.Sprintf("cm_action_jobs.num_failures"),
	sqlf.Sprintf("cm_action_jobs.log_contents"),
}

// ListActionJobsOpts is a struct that contains options for listing and
// counting action events.
type ListActionJobsOpts struct {
	// TriggerEventID, if set, will filter to only action jobs that were
	// created in response to the provided trigger event.  Refers to
	// cm_trigger_jobs(id)
	TriggerEventID *int32

	// EmailID, if set, will filter to only actions jobs that are executing the
	// given email action. Refers to cm_emails(id)
	EmailID *int

	// WebhookID, if set, will filter to only actions jobs that are executing
	// the given webhook action. Refers to cm_webhooks(id)
	WebhookID *int

	// WebhookID, if set, will filter to only actions jobs that are executing
	// the given slack webhook action. Refers to cm_slack_webhooks(id)
	SlackWebhookID *int

	// First, if defined, limits the operation to only the first n results
	First *int

	// After, if defined, starts after the provided id. Refers to
	// cm_action_jobs(id)
	After *int
}

// Conds generates a set of conditions for a SQL WHERE clause
func (o ListActionJobsOpts) Conds() *sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if o.TriggerEventID != nil {
		conds = append(conds, sqlf.Sprintf("trigger_event = %s", *o.TriggerEventID))
	}
	if o.EmailID != nil {
		conds = append(conds, sqlf.Sprintf("email = %s", *o.EmailID))
	}
	if o.WebhookID != nil {
		conds = append(conds, sqlf.Sprintf("webhook = %s", *o.WebhookID))
	}
	if o.SlackWebhookID != nil {
		conds = append(conds, sqlf.Sprintf("slack_webhook = %s", *o.SlackWebhookID))
	}
	if o.After != nil {
		conds = append(conds, sqlf.Sprintf("id > %s", *o.After))
	}
	return sqlf.Join(conds, "AND")
}

// Limit generates an argument for a SQL LIMIT clause
func (o ListActionJobsOpts) Limit() *sqlf.Query {
	if o.First == nil {
		return sqlf.Sprintf("ALL")
	}
	return sqlf.Sprintf("%s", *o.First)
}

const listActionsFmtStr = `
SELECT %s -- ActionJobsColumns
FROM cm_action_jobs
WHERE %s
ORDER BY id ASC
LIMIT %s;
`

// ListActionJobs lists events from cm_action_jobs using the provided options
func (s *codeMonitorStore) ListActionJobs(ctx context.Context, opts ListActionJobsOpts) ([]*ActionJob, error) {
	q := sqlf.Sprintf(
		listActionsFmtStr,
		sqlf.Join(ActionJobColumns, ","),
		opts.Conds(),
		opts.Limit(),
	)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanActionJobs(rows)
}

const countActionsFmtStr = `
SELECT COUNT(*)
FROM cm_action_jobs
WHERE %s
LIMIT %s
`

// CountActionJobs returns a count of the number of action jobs matching the provided list options
func (s *codeMonitorStore) CountActionJobs(ctx context.Context, opts ListActionJobsOpts) (int, error) {
	q := sqlf.Sprintf(
		countActionsFmtStr,
		opts.Conds(),
		opts.Limit(),
	)

	var count int
	err := s.QueryRow(ctx, q).Scan(&count)
	return count, err
}

const enqueueActionEmailFmtStr = `
WITH due_emails AS (
	SELECT id
	FROM cm_emails
	WHERE monitor = %s
		AND enabled = true
	EXCEPT
	SELECT DISTINCT email as id FROM cm_action_jobs
	WHERE state = 'queued'
		OR state = 'processing'
), due_webhooks AS (
	SELECT id
	FROM cm_webhooks
	WHERE monitor = %s
		AND enabled = true
	EXCEPT
	SELECT DISTINCT webhook as id FROM cm_action_jobs
	WHERE state = 'queued'
		OR state = 'processing'
), due_slack_webhooks AS (
	SELECT id
	FROM cm_slack_webhooks
	WHERE monitor = %s
		AND enabled = true
	EXCEPT
	SELECT DISTINCT slack_webhook as id FROM cm_action_jobs
	WHERE state = 'queued'
		OR state = 'processing'
)
INSERT INTO cm_action_jobs (email, webhook, slack_webhook, trigger_event)
SELECT id, CAST(NULL AS BIGINT), CAST(NULL AS BIGINT), %s::integer from due_emails
UNION
SELECT CAST(NULL AS BIGINT), id, CAST(NULL AS BIGINT), %s::integer from due_webhooks
UNION
SELECT CAST(NULL AS BIGINT), CAST(NULL AS BIGINT), id, %s::integer from due_slack_webhooks
ORDER BY 1, 2, 3
RETURNING %s
`

func (s *codeMonitorStore) EnqueueActionJobsForMonitor(ctx context.Context, monitorID int64, triggerJobID int32) ([]*ActionJob, error) {
	q := sqlf.Sprintf(
		enqueueActionEmailFmtStr,
		monitorID,
		monitorID,
		monitorID,
		triggerJobID,
		triggerJobID,
		triggerJobID,
		sqlf.Join(ActionJobColumns, ","),
	)
	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanActionJobs(rows)
}

const getActionJobMetadataFmtStr = `
SELECT
	cm.description,
	ctj.query_string,
	cm.id AS monitorID,
	ctj.search_results,
	CASE WHEN LENGTH(users.display_name) > 0 THEN users.display_name ELSE users.username END
FROM cm_action_jobs caj
INNER JOIN cm_trigger_jobs ctj on caj.trigger_event = ctj.id
INNER JOIN cm_queries cq on cq.id = ctj.query
INNER JOIN cm_monitors cm on cm.id = cq.monitor
INNER JOIN users on cm.namespace_user_id = users.id
WHERE caj.id = %s
`

// GetActionJobMetada returns the set of fields needed to execute all action jobs
func (s *codeMonitorStore) GetActionJobMetadata(ctx context.Context, jobID int32) (*ActionJobMetadata, error) {
	row := s.Store.QueryRow(ctx, sqlf.Sprintf(getActionJobMetadataFmtStr, jobID))
	var resultsJSON []byte
	m := &ActionJobMetadata{}
	err := row.Scan(&m.Description, &m.Query, &m.MonitorID, &resultsJSON, &m.OwnerName)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(resultsJSON, &m.Results); err != nil {
		return nil, err
	}
	return m, nil
}

const actionJobForIDFmtStr = `
SELECT %s -- ActionJobColumns
FROM cm_action_jobs
WHERE id = %s
`

func (s *codeMonitorStore) GetActionJob(ctx context.Context, jobID int32) (*ActionJob, error) {
	q := sqlf.Sprintf(actionJobForIDFmtStr, sqlf.Join(ActionJobColumns, ", "), jobID)
	row := s.QueryRow(ctx, q)
	return ScanActionJob(row)
}

func scanActionJobs(rows *sql.Rows) ([]*ActionJob, error) {
	var ajs []*ActionJob
	for rows.Next() {
		aj, err := ScanActionJob(rows)
		if err != nil {
			return nil, err
		}
		ajs = append(ajs, aj)
	}
	return ajs, rows.Err()
}

func ScanActionJob(row dbutil.Scanner) (*ActionJob, error) {
	aj := &ActionJob{}
	return aj, row.Scan(
		&aj.ID,
		&aj.Email,
		&aj.Webhook,
		&aj.SlackWebhook,
		&aj.TriggerEvent,
		&aj.State,
		&aj.FailureMessage,
		&aj.StartedAt,
		&aj.FinishedAt,
		&aj.ProcessAfter,
		&aj.NumResets,
		&aj.NumFailures,
		&aj.LogContents,
	)
}
