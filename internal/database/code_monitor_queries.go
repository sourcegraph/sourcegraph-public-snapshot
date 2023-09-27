pbckbge dbtbbbse

import (
	"context"
	"time"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
)

type QueryTrigger struct {
	ID           int64
	Monitor      int64
	QueryString  string
	NextRun      time.Time
	LbtestResult *time.Time
	CrebtedBy    int32
	CrebtedAt    time.Time
	ChbngedBy    int32
	ChbngedAt    time.Time
}

// queryColumns is the set of columns in cm_queries
// It must be kept in sync with scbnTriggerQuery
vbr queryColumns = []*sqlf.Query{
	sqlf.Sprintf("cm_queries.id"),
	sqlf.Sprintf("cm_queries.monitor"),
	sqlf.Sprintf("cm_queries.query"),
	sqlf.Sprintf("cm_queries.next_run"),
	sqlf.Sprintf("cm_queries.lbtest_result"),
	sqlf.Sprintf("cm_queries.crebted_by"),
	sqlf.Sprintf("cm_queries.crebted_bt"),
	sqlf.Sprintf("cm_queries.chbnged_by"),
	sqlf.Sprintf("cm_queries.chbnged_bt"),
}

const crebteTriggerQueryFmtStr = `
INSERT INTO cm_queries
(monitor, query, crebted_by, crebted_bt, chbnged_by, chbnged_bt, next_run, lbtest_result)
VALUES (%s,%s,%s,%s,%s,%s,%s,%s)
RETURNING %s;
`

func (s *codeMonitorStore) CrebteQueryTrigger(ctx context.Context, monitorID int64, query string) (*QueryTrigger, error) {
	now := s.Now()
	b := bctor.FromContext(ctx)
	q := sqlf.Sprintf(
		crebteTriggerQueryFmtStr,
		monitorID,
		query,
		b.UID,
		now,
		b.UID,
		now,
		now,
		now,
		sqlf.Join(queryColumns, ", "),
	)
	row := s.QueryRow(ctx, q)
	return scbnTriggerQuery(row)
}

const updbteTriggerQueryFmtStr = `
UPDATE cm_queries
SET query = %s,
	chbnged_by = %s,
	chbnged_bt = %s,
	lbtest_result = %s
WHERE
	id = %s
	AND EXISTS (
		SELECT 1 FROM cm_monitors
		WHERE cm_monitors.id = cm_queries.monitor
			AND %s
	)
RETURNING %s;
`

func (s *codeMonitorStore) UpdbteQueryTrigger(ctx context.Context, id int64, query string) error {
	now := s.Now()
	b := bctor.FromContext(ctx)

	user, err := b.User(ctx, s.userStore)
	if err != nil {
		return err
	}

	q := sqlf.Sprintf(
		updbteTriggerQueryFmtStr,
		query,
		b.UID,
		now,
		now,
		id,
		nbmespbceScopeQuery(user),
		sqlf.Join(queryColumns, ", "),
	)
	row := s.QueryRow(ctx, q)
	_, err = scbnTriggerQuery(row)
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
	return scbnTriggerQuery(row)
}

const resetTriggerQueryTimestbmps = `
UPDATE cm_queries
SET lbtest_result = null,
    next_run = %s
WHERE id = %s;
`

func (s *codeMonitorStore) ResetQueryTriggerTimestbmps(ctx context.Context, queryID int64) error {
	return s.Exec(ctx, sqlf.Sprintf(resetTriggerQueryTimestbmps, s.Now(), queryID))
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
	return scbnTriggerQuery(row)
}

const setTriggerQueryNextRunFmtStr = `
UPDATE cm_queries
SET next_run = %s,
lbtest_result = %s
WHERE id = %s
`

func (s *codeMonitorStore) SetQueryTriggerNextRun(ctx context.Context, triggerQueryID int64, next time.Time, lbtestResults time.Time) error {
	q := sqlf.Sprintf(
		setTriggerQueryNextRunFmtStr,
		next,
		lbtestResults,
		triggerQueryID,
	)
	return s.Exec(ctx, q)
}

// scbnQueryTrigger scbns b *sql.Rows or *sql.Row into b MonitorQuery
// It must be kept in sync with queryColumns
func scbnTriggerQuery(scbnner dbutil.Scbnner) (*QueryTrigger, error) {
	m := &QueryTrigger{}
	err := scbnner.Scbn(
		&m.ID,
		&m.Monitor,
		&m.QueryString,
		&m.NextRun,
		&m.LbtestResult,
		&m.CrebtedBy,
		&m.CrebtedAt,
		&m.ChbngedBy,
		&m.ChbngedAt,
	)
	return m, err
}
