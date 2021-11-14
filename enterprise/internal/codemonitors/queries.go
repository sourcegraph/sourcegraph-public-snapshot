package codemonitors

import (
	"context"
	"time"

	"github.com/graph-gophers/graphql-go/relay"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type QueryTrigger struct {
	Id           int64
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

func (s *codeMonitorStore) CreateQueryTrigger(ctx context.Context, monitorID int64, args *graphqlbackend.CreateTriggerArgs) error {
	q, err := s.createTriggerQueryQuery(ctx, monitorID, args)
	if err != nil {
		return err
	}
	return s.Exec(ctx, q)
}

func (s *codeMonitorStore) UpdateTriggerQuery(ctx context.Context, args *graphqlbackend.UpdateCodeMonitorArgs) error {
	q, err := s.updateTriggerQueryQuery(ctx, args)
	if err != nil {
		return err
	}
	return s.Exec(ctx, q)
}

const triggerQueryByMonitorFmtStr = `
SELECT %s -- queryColumns
FROM cm_queries
WHERE monitor = %s;
`

func (s *codeMonitorStore) TriggerQueryByMonitorIDInt64(ctx context.Context, monitorID int64) (*QueryTrigger, error) {
	q := sqlf.Sprintf(
		triggerQueryByMonitorFmtStr,
		sqlf.Join(queryColumns, ","),
		monitorID,
	)
	row := s.QueryRow(ctx, q)
	return scanTriggerQuery(row)
}

const triggerQueryByIDFmtStr = `
SELECT %s -- queryColumns
FROM cm_queries
WHERE id = %s;
`

func (s *codeMonitorStore) triggerQueryByIDInt64(ctx context.Context, queryID int64) (*QueryTrigger, error) {
	q := sqlf.Sprintf(
		triggerQueryByIDFmtStr,
		sqlf.Join(queryColumns, ","),
		queryID,
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

func (s *codeMonitorStore) ResetTriggerQueryTimestamps(ctx context.Context, queryID int64) error {
	return s.Exec(ctx, sqlf.Sprintf(resetTriggerQueryTimestamps, s.Now(), queryID))
}

const createTriggerQueryFmtStr = `
INSERT INTO cm_queries
(monitor, query, created_by, created_at, changed_by, changed_at, next_run, latest_result)
VALUES (%s,%s,%s,%s,%s,%s,%s,%s)
RETURNING %s;
`

func (s *codeMonitorStore) createTriggerQueryQuery(ctx context.Context, monitorID int64, args *graphqlbackend.CreateTriggerArgs) (*sqlf.Query, error) {
	now := s.Now()
	a := actor.FromContext(ctx)
	return sqlf.Sprintf(
		createTriggerQueryFmtStr,
		monitorID,
		args.Query,
		a.UID,
		now,
		a.UID,
		now,
		now,
		now,
		sqlf.Join(queryColumns, ", "),
	), nil
}

const updateTriggerQueryFmtStr = `
UPDATE cm_queries
SET query = %s,
	changed_by = %s,
	changed_at = %s,
	latest_result = %s
WHERE id = %s
AND monitor = %s
RETURNING %s;
`

func (s *codeMonitorStore) updateTriggerQueryQuery(ctx context.Context, args *graphqlbackend.UpdateCodeMonitorArgs) (*sqlf.Query, error) {
	now := s.Now()
	a := actor.FromContext(ctx)

	var triggerID int64
	err := relay.UnmarshalSpec(args.Trigger.Id, &triggerID)
	if err != nil {
		return nil, err
	}

	var monitorID int64
	err = relay.UnmarshalSpec(args.Monitor.Id, &monitorID)
	if err != nil {
		return nil, err
	}

	return sqlf.Sprintf(
		updateTriggerQueryFmtStr,
		args.Trigger.Update.Query,
		a.UID,
		now,
		now,
		triggerID,
		monitorID,
		sqlf.Join(queryColumns, ", "),
	), nil
}

const getQueryByRecordIDFmtStr = `
SELECT %s -- queryColumns
FROM cm_queries
INNER JOIN cm_trigger_jobs j ON cm_queries.id = j.query
WHERE j.id = %s
`

func (s *codeMonitorStore) GetQueryByRecordID(ctx context.Context, recordID int) (*QueryTrigger, error) {
	q := sqlf.Sprintf(
		getQueryByRecordIDFmtStr,
		sqlf.Join(queryColumns, ","),
		recordID,
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

func (s *codeMonitorStore) SetTriggerQueryNextRun(ctx context.Context, triggerQueryID int64, next time.Time, latestResults time.Time) error {
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
		&m.Id,
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
