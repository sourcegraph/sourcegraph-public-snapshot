package codemonitors

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"
)

type MonitorQuery struct {
	Id           int64
	Monitor      int64
	QueryString  string
	NextRun      time.Time
	LatestResult *time.Time
	CreatedBy    int64
	CreatedAt    time.Time
	ChangedBy    int64
	ChangedAt    time.Time
}

const getQueryByRecordIDFmtStr = `
SELECT q.id, q.monitor, q.query, q.next_run, q.latest_result, q.created_by, q.created_at, q.changed_by, q.changed_at
FROM cm_queries q INNER JOIN cm_trigger_jobs j ON q.id = j.query
WHERE j.id = %s
`

func (s *Store) GetQueryByRecordID(ctx context.Context, recordID int) (query *MonitorQuery, err error) {
	q := sqlf.Sprintf(
		getQueryByRecordIDFmtStr,
		recordID,
	)
	var ms []*MonitorQuery
	var rows *sql.Rows
	rows, err = s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	ms, err = ScanTriggerQueries(rows)
	if err != nil {
		return nil, err
	}
	if len(ms) != 1 {
		return nil, fmt.Errorf("query should have returned 1 row")
	}
	return ms[0], nil
}

const setTriggerQueryNextRunFmtStr = `
UPDATE cm_queries
SET next_run = %s,
latest_result = %s
WHERE id = %s
`

func (s *Store) SetTriggerQueryNextRun(ctx context.Context, triggerQueryID int64, next time.Time, latestResults time.Time) error {
	q := sqlf.Sprintf(
		setTriggerQueryNextRunFmtStr,
		next,
		latestResults,
		triggerQueryID,
	)
	return s.Exec(ctx, q)
}

func ScanTriggerQueries(rows *sql.Rows) (ms []*MonitorQuery, err error) {
	for rows.Next() {
		m := &MonitorQuery{}
		if err := rows.Scan(
			&m.Id,
			&m.Monitor,
			&m.QueryString,
			&m.NextRun,
			&m.LatestResult,
			&m.CreatedBy,
			&m.CreatedAt,
			&m.ChangedBy,
			&m.ChangedAt,
		); err != nil {
			return nil, err
		}
		ms = append(ms, m)
	}
	err = rows.Close()
	if err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ms, nil
}
