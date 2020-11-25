package codemonitors

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

const enqueueTriggerQueryFmtStr = `
WITH due AS (
    SELECT cm_queries.id as id
    FROM cm_queries INNER JOIN cm_monitors ON cm_queries.monitor = cm_monitors.id
    WHERE (cm_queries.next_run <= clock_timestamp() OR cm_queries.next_run IS NULL)
	AND cm_monitors.enabled = true
),
busy AS (
    SELECT DISTINCT query as id FROM cm_trigger_jobs
    WHERE state = 'queued'
    OR state = 'processing'
)
INSERT INTO cm_trigger_jobs (query)
SELECT id from due EXCEPT SELECT id from busy
`

func (s *Store) EnqueueTriggerQueries(ctx context.Context) (err error) {
	return s.Store.Exec(ctx, sqlf.Sprintf(enqueueTriggerQueryFmtStr))
}

type TriggerJobs struct {
	Id             int32
	Query          int64
	State          string
	FailureMessage *string
	StartedAt      *time.Time
	FinishedAt     *time.Time
	ProcessAfter   *time.Time
	NumResets      int32
	NumFailures    int32
	LogContents    *string
}

func (r *TriggerJobs) RecordID() int {
	return int(r.Id)
}

func ScanTriggerJobs(rows *sql.Rows, err error) (workerutil.Record, bool, error) {
	records, err := scanTriggerJobs(rows, err)
	if err != nil {
		return &TriggerJobs{}, false, err
	}
	return records[0], true, nil
}

func scanTriggerJobs(rows *sql.Rows, err error) ([]*TriggerJobs, error) {
	var ms []*TriggerJobs
	for rows.Next() {
		m := &TriggerJobs{}
		if err := rows.Scan(
			&m.Id,
			&m.Query,
			&m.State,
			&m.FailureMessage,
			&m.StartedAt,
			&m.FinishedAt,
			&m.ProcessAfter,
			&m.NumResets,
			&m.NumFailures,
			&m.LogContents,
		); err != nil {
			return nil, err
		}
		ms = append(ms, m)
	}
	err = rows.Close()
	if err != nil {
		return nil, err
	}
	// Rows.Err will report the last error encountered by Rows.Scan.
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ms, nil
}

var TriggerJobsColumns = []*sqlf.Query{
	sqlf.Sprintf("cm_trigger_jobs.id"),
	sqlf.Sprintf("cm_trigger_jobs.query"),
	sqlf.Sprintf("cm_trigger_jobs.state"),
	sqlf.Sprintf("cm_trigger_jobs.failure_message"),
	sqlf.Sprintf("cm_trigger_jobs.started_at"),
	sqlf.Sprintf("cm_trigger_jobs.finished_at"),
	sqlf.Sprintf("cm_trigger_jobs.process_after"),
	sqlf.Sprintf("cm_trigger_jobs.num_resets"),
	sqlf.Sprintf("cm_trigger_jobs.num_failures"),
	sqlf.Sprintf("cm_trigger_jobs.log_contents"),
}
