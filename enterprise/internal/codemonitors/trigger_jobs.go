package codemonitors

import (
	"context"
	"database/sql"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

func (r *TriggerJob) RecordID() int {
	return r.Id
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
`

func (s *codeMonitorStore) EnqueueQueryTriggerJobs(ctx context.Context) error {
	return s.Store.Exec(ctx, sqlf.Sprintf(enqueueTriggerQueryFmtStr))
}

const logSearchFmtStr = `
UPDATE cm_trigger_jobs
SET query_string = %s,
    results = %s,
    num_results = %s
WHERE id = %s
`

func (s *codeMonitorStore) LogSearch(ctx context.Context, queryString string, numResults int, recordID int) error {
	return s.Store.Exec(ctx, sqlf.Sprintf(logSearchFmtStr, queryString, numResults > 0, numResults, recordID))
}

const deleteObsoleteJobLogsFmtStr = `
DELETE FROM cm_trigger_jobs
WHERE results IS NOT TRUE
AND state = 'completed'
`

// DeleteObsoleteJobLogs deletes all runs which are marked as completed and did
// not return results.
func (s *codeMonitorStore) DeleteObsoleteJobLogs(ctx context.Context) error {
	return s.Store.Exec(ctx, sqlf.Sprintf(deleteObsoleteJobLogsFmtStr))
}

const deleteOldJobLogsFmtStr = `
DELETE FROM cm_trigger_jobs
WHERE finished_at < (NOW() - (%s * '1 day'::interval));
`

// DeleteOldJobLogs deletes trigger jobs which have finished and are older than
// 'retention' days. Due to cascading, action jobs will be deleted as well.
func (s *codeMonitorStore) DeleteOldJobLogs(ctx context.Context, retentionInDays int) error {
	return s.Store.Exec(ctx, sqlf.Sprintf(deleteOldJobLogsFmtStr, retentionInDays))
}

const getEventsForQueryIDInt64FmtStr = `
SELECT id, query, query_string, results, num_results, state, failure_message, started_at, finished_at, process_after, num_resets, num_failures, log_contents
FROM cm_trigger_jobs
WHERE ((state = 'completed' AND results IS TRUE) OR (state != 'completed'))
AND query = %s
AND id > %s
ORDER BY id ASC
LIMIT %s;
`

func (s *codeMonitorStore) ListQueryTriggerJobs(ctx context.Context, queryID int64, args *graphqlbackend.ListEventsArgs) ([]*TriggerJob, error) {
	after, err := unmarshalAfter(args.After)
	if err != nil {
		return nil, err
	}
	q := sqlf.Sprintf(
		getEventsForQueryIDInt64FmtStr,
		queryID,
		after,
		args.First,
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
WHERE ((state = 'completed' AND results IS TRUE) OR (state != 'completed'))
AND query = %s
`

func (s *codeMonitorStore) TotalCountEventsForQueryIDInt64(ctx context.Context, queryID int64) (int32, error) {
	q := sqlf.Sprintf(
		totalCountEventsForQueryIDInt64FmtStr,
		queryID,
	)
	var count int32
	err := s.Store.QueryRow(ctx, q).Scan(&count)
	return count, err
}

type TriggerJob struct {
	Id    int
	Query int64

	// The query we ran including after: filter.
	QueryString *string

	// Whether we got any results.
	Results    *bool
	NumResults *int

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
	m := &TriggerJob{}
	err := scanner.Scan(
		&m.Id,
		&m.Query,
		&m.QueryString,
		&m.Results,
		&m.NumResults,
		&m.State,
		&m.FailureMessage,
		&m.StartedAt,
		&m.FinishedAt,
		&m.ProcessAfter,
		&m.NumResets,
		&m.NumFailures,
		&m.LogContents,
	)
	return m, err
}

var TriggerJobsColumns = []*sqlf.Query{
	sqlf.Sprintf("cm_trigger_jobs.id"),
	sqlf.Sprintf("cm_trigger_jobs.query"),
	sqlf.Sprintf("cm_trigger_jobs.query_string"),
	sqlf.Sprintf("cm_trigger_jobs.results"),
	sqlf.Sprintf("cm_trigger_jobs.num_results"),
	sqlf.Sprintf("cm_trigger_jobs.state"),
	sqlf.Sprintf("cm_trigger_jobs.failure_message"),
	sqlf.Sprintf("cm_trigger_jobs.started_at"),
	sqlf.Sprintf("cm_trigger_jobs.finished_at"),
	sqlf.Sprintf("cm_trigger_jobs.process_after"),
	sqlf.Sprintf("cm_trigger_jobs.num_resets"),
	sqlf.Sprintf("cm_trigger_jobs.num_failures"),
	sqlf.Sprintf("cm_trigger_jobs.log_contents"),
}

func unmarshalAfter(after *string) (int, error) {
	if after == nil {
		return 0, nil
	}

	var a int
	err := relay.UnmarshalSpec(graphql.ID(*after), &a)
	return a, err
}
