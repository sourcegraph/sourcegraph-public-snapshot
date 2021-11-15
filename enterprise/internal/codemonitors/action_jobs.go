package codemonitors

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

type ActionJob struct {
	Id           int
	Email        *int
	Webhook      *int
	SlackWebhook *int
	TriggerEvent int

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
	return a.Id
}

type ActionJobMetadata struct {
	Description string
	MonitorID   int64
	NumResults  *int

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
	TriggerEventID *int

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
WITH due AS (
	SELECT e.id
	FROM cm_emails e
	INNER JOIN cm_queries q ON e.monitor = q.monitor
	WHERE q.id = %s AND e.enabled = true
),
busy AS (
	SELECT DISTINCT email as id FROM cm_action_jobs
	WHERE state = 'queued'
	OR state = 'processing'
)
INSERT INTO cm_action_jobs (email, trigger_event)
SELECT id, %s::integer from due EXCEPT SELECT id, %s::integer from busy ORDER BY id
`

// TODO(camdencheek): could we enqueue based on monitor ID rather than query ID? Would avoid joins above.
func (s *codeMonitorStore) EnqueueActionEmailsForQueryIDInt64(ctx context.Context, queryID int64, triggerEventID int) (err error) {
	return s.Store.Exec(ctx, sqlf.Sprintf(enqueueActionEmailFmtStr, queryID, triggerEventID, triggerEventID))
}

const getActionJobMetadataFmtStr = `
SELECT
	cm.description,
	ctj.query_string,
	cm.id AS monitorID,
	ctj.num_results
FROM cm_action_jobs caj
INNER JOIN cm_trigger_jobs ctj on caj.trigger_event = ctj.id
INNER JOIN cm_queries cq on cq.id = ctj.query
INNER JOIN cm_monitors cm on cm.id = cq.monitor
WHERE caj.id = %s
`

func (s *codeMonitorStore) GetActionJobMetadata(ctx context.Context, recordID int) (*ActionJobMetadata, error) {
	row := s.Store.QueryRow(ctx, sqlf.Sprintf(getActionJobMetadataFmtStr, recordID))
	m := &ActionJobMetadata{}
	return m, row.Scan(&m.Description, &m.Query, &m.MonitorID, &m.NumResults)
}

const actionJobForIDFmtStr = `
SELECT %s -- ActionJobsColumns
FROM cm_action_jobs
WHERE id = %s
`

func (s *codeMonitorStore) ActionJobForIDInt(ctx context.Context, recordID int) (*ActionJob, error) {
	q := sqlf.Sprintf(actionJobForIDFmtStr, sqlf.Join(ActionJobColumns, ", "), recordID)
	row := s.QueryRow(ctx, q)
	return scanActionJob(row)
}

// ScanActionJobRecord implements the worker RecordScanFn
func ScanActionJobRecord(rows *sql.Rows, err error) (workerutil.Record, bool, error) {
	if err != nil {
		return &TriggerJobs{}, false, err
	}
	defer rows.Close()

	records, err := scanActionJobs(rows)
	if err != nil || len(records) == 0 {
		return &TriggerJobs{}, false, err
	}
	return records[0], true, nil
}

func scanActionJobs(rows *sql.Rows) ([]*ActionJob, error) {
	var ajs []*ActionJob
	for rows.Next() {
		aj, err := scanActionJob(rows)
		if err != nil {
			return nil, err
		}
		ajs = append(ajs, aj)
	}
	return ajs, rows.Err()
}

func scanActionJob(row dbutil.Scanner) (*ActionJob, error) {
	aj := &ActionJob{}
	return aj, row.Scan(
		&aj.Id,
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
