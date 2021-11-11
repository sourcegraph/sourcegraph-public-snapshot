package codemonitors

import (
	"context"
	"database/sql"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

type ActionJob struct {
	Id                int
	Email             *int64
	Webhook           *int64
	SlackNotification *int64
	TriggerEvent      int

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

var ActionJobsColumns = []*sqlf.Query{
	sqlf.Sprintf("cm_action_jobs.id"),
	sqlf.Sprintf("cm_action_jobs.email"),
	sqlf.Sprintf("cm_action_jobs.webhook"),
	sqlf.Sprintf("cm_action_jobs.slack_notification"),
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

const readActionEmailEventsFmtStr = `
SELECT %s -- ActionJobsColumns
FROM cm_action_jobs
WHERE %s
AND id > %s
ORDER BY id ASC
LIMIT %s;
`

func (s *Store) ReadActionEmailEvents(ctx context.Context, emailID int64, triggerEventID *int, args *graphqlbackend.ListEventsArgs) (js []*ActionJob, err error) {
	var where *sqlf.Query
	if triggerEventID == nil {
		where = sqlf.Sprintf("email = %s", emailID)
	} else {
		where = sqlf.Sprintf("email = %s AND trigger_event = %s", emailID, *triggerEventID)
	}
	var rows *sql.Rows
	after, err := unmarshalAfter(args.After)
	if err != nil {
		return nil, err
	}
	rows, err = s.Query(ctx, sqlf.Sprintf(readActionEmailEventsFmtStr, sqlf.Join(ActionJobsColumns, ", "), where, after, args.First))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanActionJobs(rows, err)
}

const totalActionEmailEventsFmtStr = `
SELECT COUNT(*)
FROM cm_action_jobs
WHERE %s
`

func (s *Store) TotalActionEmailEvents(ctx context.Context, emailID int64, triggerEventID *int) (totalCount int32, err error) {
	var where *sqlf.Query
	if triggerEventID == nil {
		where = sqlf.Sprintf("email = %s", emailID)
	} else {
		where = sqlf.Sprintf("email = %s AND trigger_event = %s", emailID, *triggerEventID)
	}
	err = s.QueryRow(ctx, sqlf.Sprintf(totalActionEmailEventsFmtStr, where)).Scan(&totalCount)
	if err != nil {
		return -1, err
	}
	return totalCount, nil
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

func (s *Store) EnqueueActionEmailsForQueryIDInt64(ctx context.Context, queryID int64, triggerEventID int) (err error) {
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

func (s *Store) GetActionJobMetadata(ctx context.Context, recordID int) (m *ActionJobMetadata, err error) {
	row := s.Store.QueryRow(ctx, sqlf.Sprintf(getActionJobMetadataFmtStr, recordID))
	m = &ActionJobMetadata{}
	err = row.Scan(&m.Description, &m.Query, &m.MonitorID, &m.NumResults)
	if err != nil {
		return nil, err
	}
	return m, nil
}

const actionJobForIDFmtStr = `
SELECT %s -- ActionJobsColumns
FROM cm_action_jobs
WHERE id = %s
`

func (s *Store) ActionJobForIDInt(ctx context.Context, recordID int) (*ActionJob, error) {
	return s.runActionJobQuery(ctx, sqlf.Sprintf(actionJobForIDFmtStr, sqlf.Join(ActionJobsColumns, ", "), recordID))
}

func (s *Store) runActionJobQuery(ctx context.Context, q *sqlf.Query) (ajs *ActionJob, err error) {
	var rows *sql.Rows
	rows, err = s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var es []*ActionJob
	es, err = scanActionJobs(rows, err)
	if err != nil {
		return nil, err
	}
	if len(es) == 0 {
		return nil, errors.Errorf("operation failed. Query should have returned at least 1 row")
	}
	return es[0], nil
}

func ScanActionJobs(rows *sql.Rows, err error) (workerutil.Record, bool, error) {
	records, err := scanActionJobs(rows, err)
	if err != nil || len(records) == 0 {
		return &TriggerJobs{}, false, err
	}
	return records[0], true, nil
}

func scanActionJobs(rows *sql.Rows, err error) ([]*ActionJob, error) {
	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()
	var ajs []*ActionJob
	for rows.Next() {
		aj := &ActionJob{}
		// Columns should be kept in sync with ActionJobsColumns
		if err := rows.Scan(
			&aj.Id,
			&aj.Email,
			&aj.Webhook,
			&aj.SlackNotification,
			&aj.TriggerEvent,
			&aj.State,
			&aj.FailureMessage,
			&aj.StartedAt,
			&aj.FinishedAt,
			&aj.ProcessAfter,
			&aj.NumResets,
			&aj.NumFailures,
			&aj.LogContents,
		); err != nil {
			return nil, err
		}
		ajs = append(ajs, aj)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return ajs, nil
}
