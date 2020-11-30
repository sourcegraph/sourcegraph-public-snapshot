package codemonitors

import (
	"context"
	"database/sql"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/db/basestore"
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

const logSearchFmtStr = `
UPDATE cm_trigger_jobs
SET query_string = %s,
	results = %s
WHERE id = %s
`

func (s *Store) LogSearch(ctx context.Context, queryString string, results bool, recordID int) error {
	return s.Store.Exec(ctx, sqlf.Sprintf(logSearchFmtStr, queryString, results, recordID))
}

const deleteObsoleteJobLogsFmtStr = `
DELETE FROM cm_trigger_jobs
WHERE results IS NOT TRUE
AND state = 'completed'
`

// DeleteObsoleteJobLogs deletes all runs which are marked as completed and did
// not return results.
func (s *Store) DeleteObsoleteJobLogs(ctx context.Context) error {
	return s.Store.Exec(ctx, sqlf.Sprintf(deleteObsoleteJobLogsFmtStr))
}

const getEventsForQueryIDInt64FmtStr = `
SELECT id, query, query_string, results, state, failure_message, started_at, finished_at, process_after, num_resets, num_failures, log_contents
FROM cm_trigger_jobs
WHERE ((state = 'completed' AND results IS TRUE) OR (state != 'completed'))
AND query = %s
AND id > %s
ORDER BY id ASC
LIMIT %s;
`

func (s *Store) GetEventsForQueryIDInt64(ctx context.Context, queryID int64, args *graphqlbackend.ListEventsArgs) ([]*TriggerJobs, error) {
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
	var rows *sql.Rows
	rows, err = s.Store.Query(ctx, q)
	return scanTriggerJobs(rows, err)
}

type TriggerJobs struct {
	Id    int32
	Query int64

	// The query we ran including after: filter.
	QueryString *string

	// Whether we got any results.
	Results *bool

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
	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()
	var ms []*TriggerJobs
	for rows.Next() {
		m := &TriggerJobs{}
		if err := rows.Scan(
			&m.Id,
			&m.Query,
			&m.QueryString,
			&m.Results,
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
	sqlf.Sprintf("cm_trigger_jobs.query_string"),
	sqlf.Sprintf("cm_trigger_jobs.results"),
	sqlf.Sprintf("cm_trigger_jobs.state"),
	sqlf.Sprintf("cm_trigger_jobs.failure_message"),
	sqlf.Sprintf("cm_trigger_jobs.started_at"),
	sqlf.Sprintf("cm_trigger_jobs.finished_at"),
	sqlf.Sprintf("cm_trigger_jobs.process_after"),
	sqlf.Sprintf("cm_trigger_jobs.num_resets"),
	sqlf.Sprintf("cm_trigger_jobs.num_failures"),
	sqlf.Sprintf("cm_trigger_jobs.log_contents"),
}

func unmarshalAfter(after *string) (int64, error) {
	var a int64
	if after == nil {
		a = 0
	} else {
		err := relay.UnmarshalSpec(graphql.ID(*after), &a)
		if err != nil {
			return -1, err
		}
	}
	return a, nil
}
