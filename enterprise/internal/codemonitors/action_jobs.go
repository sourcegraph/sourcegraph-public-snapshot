package codemonitors

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

type ActionJob struct {
	Id           int
	Email        int64
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

var ActionJobsColumns = []*sqlf.Query{
	sqlf.Sprintf("cm_action_jobs.id"),
	sqlf.Sprintf("cm_action_jobs.email"),
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

func (s *Store) ReadActionEmailEvents(ctx context.Context, emailID int64, triggerEventID *int, args *graphqlbackend.ListEventsArgs) ([]*ActionJob, error) {
	conditions := []*sqlf.Query{sqlf.Sprintf("email = %s", emailID)}
	if triggerEventID != nil {
		conditions = append(conditions, sqlf.Sprintf("trigger_event = %s", *triggerEventID))
	}
	after, err := unmarshalAfter(args.After)
	if err != nil {
		return nil, err
	}
	q := sqlf.Sprintf(
		readActionEmailEventsFmtStr,
		sqlf.Join(ActionJobsColumns, ","),
		sqlf.Join(conditions, "AND"),
		after,
		args.First,
	)
	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanActionJobs(rows)
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

// TODO(camdencheek): could we enqueue based on monitor ID rather than query ID? Would avoid joins above.
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

func (s *Store) GetActionJobMetadata(ctx context.Context, recordID int) (*ActionJobMetadata, error) {
	row := s.Store.QueryRow(ctx, sqlf.Sprintf(getActionJobMetadataFmtStr, recordID))
	m := &ActionJobMetadata{}
	return m, row.Scan(&m.Description, &m.Query, &m.MonitorID, &m.NumResults)
}

const actionJobForIDFmtStr = `
SELECT %s -- ActionJobsColumns
FROM cm_action_jobs
WHERE id = %s
`

func (s *Store) ActionJobForIDInt(ctx context.Context, recordID int) (*ActionJob, error) {
	q := sqlf.Sprintf(actionJobForIDFmtStr, sqlf.Join(ActionJobsColumns, ", "), recordID)
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
