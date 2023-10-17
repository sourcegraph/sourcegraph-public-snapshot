package database

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type QueryTrigger struct {
	ID           int64
	Monitor      int64
	QueryString  string
	NextRun      time.Time
	LatestResult *time.Time
	CreatedBy    int32
	CreatedAt    time.Time
	ChangedBy    int32
	ChangedAt    time.Time
}

// queryColumns is the set of columns in cm_queries
// It must be kept in sync with scanTriggerQuery
var queryColumns = []*sqlf.Query{
	sqlf.Sprintf("cm_queries.id"),
	sqlf.Sprintf("cm_queries.monitor"),
	sqlf.Sprintf("cm_queries.query"),
	sqlf.Sprintf("cm_queries.next_run"),
	sqlf.Sprintf("cm_queries.latest_result"),
	sqlf.Sprintf("cm_queries.created_by"),
	sqlf.Sprintf("cm_queries.created_at"),
	sqlf.Sprintf("cm_queries.changed_by"),
	sqlf.Sprintf("cm_queries.changed_at"),
}

const createTriggerQueryFmtStr = `
INSERT INTO cm_queries
(monitor, query, created_by, created_at, changed_by, changed_at, next_run, latest_result)
VALUES (%s,%s,%s,%s,%s,%s,%s,%s)
RETURNING %s;
`

func (s *codeMonitorStore) CreateQueryTrigger(ctx context.Context, monitorID int64, query string) (*QueryTrigger, error) {
	now := s.Now()
	a := actor.FromContext(ctx)
	q := sqlf.Sprintf(
		createTriggerQueryFmtStr,
		monitorID,
		query,
		a.UID,
		now,
		a.UID,
		now,
		now,
		now,
		sqlf.Join(queryColumns, ", "),
	)
	row := s.QueryRow(ctx, q)
	return scanTriggerQuery(row)
}

const updateTriggerQueryFmtStr = `
UPDATE cm_queries
SET query = %s,
	changed_by = %s,
	changed_at = %s,
	latest_result = %s
WHERE
	id = %s
	AND EXISTS (
		SELECT 1 FROM cm_monitors
		WHERE cm_monitors.id = cm_queries.monitor
			AND %s
	)
RETURNING %s;
`

func (s *codeMonitorStore) UpdateQueryTrigger(ctx context.Context, id int64, query string) error {
	now := s.Now()
	a := actor.FromContext(ctx)

	user, err := a.User(ctx, s.userStore)
	if err != nil {
		return err
	}

	q := sqlf.Sprintf(
		updateTriggerQueryFmtStr,
		query,
		a.UID,
		now,
		now,
		id,
		namespaceScopeQuery(user),
		sqlf.Join(queryColumns, ", "),
	)
	row := s.QueryRow(ctx, q)
	_, err = scanTriggerQuery(row)
	return err
}

const triggerQueryByMonitorFmtStr = `
SELECT %s -- queryColumns
FROM cm_queries
WHERE monitor = %s;
`

func (s *codeMonitorStore) GetQueryTriggerForMonitor(ctx context.Context, monitorID int64) (*QueryTrigger, error) {
	q := sqlf.Sprintf(
		triggerQueryByMonitorFmtStr,
		sqlf.Join(queryColumns, ","),
		monitorID,
	)
	row := s.QueryRow(ctx, q)
	return scanTriggerQuery(row)
}

const resetTriggerQueryTimestamps = `
UPDATE cm_queries
SET latest_result = null,
    next_run = %s
WHERE id = %s;
`

func (s *codeMonitorStore) ResetQueryTriggerTimestamps(ctx context.Context, queryID int64) error {
	return s.Exec(ctx, sqlf.Sprintf(resetTriggerQueryTimestamps, s.Now(), queryID))
}

const getQueryByRecordIDFmtStr = `
SELECT %s -- queryColumns
FROM cm_queries
INNER JOIN cm_trigger_jobs j ON cm_queries.id = j.query
WHERE j.id = %s
`

func (s *codeMonitorStore) GetQueryTriggerForJob(ctx context.Context, triggerJob int32) (*QueryTrigger, error) {
	q := sqlf.Sprintf(
		getQueryByRecordIDFmtStr,
		sqlf.Join(queryColumns, ","),
		triggerJob,
	)
	row := s.QueryRow(ctx, q)
	return scanTriggerQuery(row)
}

const setTriggerQueryNextRunFmtStr = `
UPDATE cm_queries
SET next_run = %s,
latest_result = %s
WHERE id = %s
`

func (s *codeMonitorStore) SetQueryTriggerNextRun(ctx context.Context, triggerQueryID int64, next time.Time, latestResults time.Time) error {
	q := sqlf.Sprintf(
		setTriggerQueryNextRunFmtStr,
		next,
		latestResults,
		triggerQueryID,
	)
	return s.Exec(ctx, q)
}

// scanQueryTrigger scans a *sql.Rows or *sql.Row into a MonitorQuery
// It must be kept in sync with queryColumns
func scanTriggerQuery(scanner dbutil.Scanner) (*QueryTrigger, error) {
	m := &QueryTrigger{}
	err := scanner.Scan(
		&m.ID,
		&m.Monitor,
		&m.QueryString,
		&m.NextRun,
		&m.LatestResult,
		&m.CreatedBy,
		&m.CreatedAt,
		&m.ChangedBy,
		&m.ChangedAt,
	)
	return m, err
}
