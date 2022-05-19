package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

type TriggerJob struct {
	ID    int32
	Query int64

	// The query we ran including after: filter.
	QueryString *string

	SearchResults []*result.CommitMatch

	// Fields demanded for any dbworker.
	State          string
	FailureMessage *string
	StartedAt      *time.Time
	FinishedAt     *time.Time
	ProcessAfter   *time.Time
	NumResets      int32
	NumFailures    int32
	LogContents    *string
}

func (r *TriggerJob) RecordID() int {
	return int(r.ID)
}

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
SELECT id from due EXCEPT SELECT id from busy ORDER BY id
RETURNING %s
`

func (s *codeMonitorStore) EnqueueQueryTriggerJobs(ctx context.Context) ([]*TriggerJob, error) {
	rows, err := s.Store.Query(ctx, sqlf.Sprintf(enqueueTriggerQueryFmtStr, sqlf.Join(TriggerJobsColumns, ",")))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTriggerJobs(rows)
}

const logSearchFmtStr = `
UPDATE cm_trigger_jobs
SET query_string = %s,
    search_results = %s
WHERE id = %s
`

func (s *codeMonitorStore) UpdateTriggerJobWithResults(ctx context.Context, triggerJobID int32, queryString string, results []*result.CommitMatch) error {
	if results == nil {
		// appease db non-null constraint
		results = []*result.CommitMatch{}
	}

	resultsJSON, err := json.Marshal(results)
	if err != nil {
		return err
	}
	return s.Store.Exec(ctx, sqlf.Sprintf(logSearchFmtStr, queryString, resultsJSON, triggerJobID))
}

const deleteObsoleteJobLogsFmtStr = `
DELETE FROM cm_trigger_jobs
WHERE jsonb_array_length(search_results) = 0
AND state = 'completed'
`

// DeleteObsoleteTriggerJobs deletes all runs which are marked as completed and did
// not return results.
func (s *codeMonitorStore) DeleteObsoleteTriggerJobs(ctx context.Context) error {
	return s.Store.Exec(ctx, sqlf.Sprintf(deleteObsoleteJobLogsFmtStr))
}

const deleteOldJobLogsFmtStr = `
DELETE FROM cm_trigger_jobs
WHERE finished_at < (NOW() - (%s * '1 day'::interval));
`

// DeleteOldTriggerJobs deletes trigger jobs which have finished and are older than
// 'retention' days. Due to cascading, action jobs will be deleted as well.
func (s *codeMonitorStore) DeleteOldTriggerJobs(ctx context.Context, retentionInDays int) error {
	return s.Store.Exec(ctx, sqlf.Sprintf(deleteOldJobLogsFmtStr, retentionInDays))
}

type ListTriggerJobsOpts struct {
	QueryID *int64
	First   *int
	After   *int64
}

func (o ListTriggerJobsOpts) Conds() *sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("true")}
	if o.QueryID != nil {
		conds = append(conds, sqlf.Sprintf("query = %s", *o.QueryID))
	}
	if o.After != nil {
		conds = append(conds, sqlf.Sprintf("id < %s", *o.After))
	}
	return sqlf.Join(conds, "AND")
}

func (o ListTriggerJobsOpts) Limit() *sqlf.Query {
	if o.First == nil {
		return sqlf.Sprintf("ALL")
	}
	return sqlf.Sprintf("%s", *o.First)
}

const getEventsForQueryIDInt64FmtStr = `
SELECT %s
FROM cm_trigger_jobs
WHERE %s
ORDER BY id DESC
LIMIT %s;
`

func (s *codeMonitorStore) ListQueryTriggerJobs(ctx context.Context, opts ListTriggerJobsOpts) ([]*TriggerJob, error) {
	q := sqlf.Sprintf(
		getEventsForQueryIDInt64FmtStr,
		sqlf.Join(TriggerJobsColumns, ","),
		opts.Conds(),
		opts.Limit(),
	)
	rows, err := s.Store.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTriggerJobs(rows)
}

const totalCountEventsForQueryIDInt64FmtStr = `
SELECT COUNT(*)
FROM cm_trigger_jobs
WHERE ((state = 'completed' AND jsonb_array_length(search_results) > 0) OR (state != 'completed'))
AND query = %s
`

func (s *codeMonitorStore) CountQueryTriggerJobs(ctx context.Context, queryID int64) (int32, error) {
	q := sqlf.Sprintf(
		totalCountEventsForQueryIDInt64FmtStr,
		queryID,
	)
	var count int32
	err := s.Store.QueryRow(ctx, q).Scan(&count)
	return count, err
}

func ScanTriggerJobsRecord(rows *sql.Rows, err error) (workerutil.Record, bool, error) {
	if err != nil {
		return nil, false, err
	}
	records, err := scanTriggerJobs(rows)
	if err != nil || len(records) == 0 {
		return &TriggerJob{}, false, err
	}
	return records[0], true, nil
}

func scanTriggerJobs(rows *sql.Rows) ([]*TriggerJob, error) {
	var js []*TriggerJob
	for rows.Next() {
		j, err := scanTriggerJob(rows)
		if err != nil {
			return nil, err
		}
		js = append(js, j)
	}
	return js, rows.Err()
}

func scanTriggerJob(scanner dbutil.Scanner) (*TriggerJob, error) {
	var resultsJSON []byte
	m := &TriggerJob{}
	err := scanner.Scan(
		&m.ID,
		&m.Query,
		&m.QueryString,
		&resultsJSON,
		&m.State,
		&m.FailureMessage,
		&m.StartedAt,
		&m.FinishedAt,
		&m.ProcessAfter,
		&m.NumResets,
		&m.NumFailures,
		&m.LogContents,
	)
	if err != nil {
		return nil, err
	}

	if len(resultsJSON) > 0 {
		if err := json.Unmarshal(resultsJSON, &m.SearchResults); err != nil {
			return nil, err
		}
	}

	return m, nil
}

var TriggerJobsColumns = []*sqlf.Query{
	sqlf.Sprintf("cm_trigger_jobs.id"),
	sqlf.Sprintf("cm_trigger_jobs.query"),
	sqlf.Sprintf("cm_trigger_jobs.query_string"),
	sqlf.Sprintf("cm_trigger_jobs.search_results"),
	sqlf.Sprintf("cm_trigger_jobs.state"),
	sqlf.Sprintf("cm_trigger_jobs.failure_message"),
	sqlf.Sprintf("cm_trigger_jobs.started_at"),
	sqlf.Sprintf("cm_trigger_jobs.finished_at"),
	sqlf.Sprintf("cm_trigger_jobs.process_after"),
	sqlf.Sprintf("cm_trigger_jobs.num_resets"),
	sqlf.Sprintf("cm_trigger_jobs.num_failures"),
	sqlf.Sprintf("cm_trigger_jobs.log_contents"),
}
