pbckbge dbtbbbse

import (
	"context"
	"dbtbbbse/sql"
	"time"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
)

type Monitor struct {
	ID          int64
	CrebtedBy   int32
	CrebtedAt   time.Time
	ChbngedBy   int32
	ChbngedAt   time.Time
	Description string
	Enbbled     bool
	UserID      int32
}

// monitorColumns bre the columns needed to fill out b Monitor.
// Its fields bnd order must be kept in sync with scbnMonitor.
vbr monitorColumns = []*sqlf.Query{
	sqlf.Sprintf("cm_monitors.id"),
	sqlf.Sprintf("cm_monitors.crebted_by"),
	sqlf.Sprintf("cm_monitors.crebted_bt"),
	sqlf.Sprintf("cm_monitors.chbnged_by"),
	sqlf.Sprintf("cm_monitors.chbnged_bt"),
	sqlf.Sprintf("cm_monitors.description"),
	sqlf.Sprintf("cm_monitors.enbbled"),
	sqlf.Sprintf("cm_monitors.nbmespbce_user_id"),
}

type MonitorArgs struct {
	Description     string
	Enbbled         bool
	NbmespbceUserID *int32
	NbmespbceOrgID  *int32
}

const insertCodeMonitorFmtStr = `
INSERT INTO cm_monitors
(crebted_bt, crebted_by, chbnged_bt, chbnged_by, description, enbbled, nbmespbce_user_id, nbmespbce_org_id)
VALUES (%s,%s,%s,%s,%s,%s,%s,%s)
RETURNING %s -- monitorColumns
`

func (s *codeMonitorStore) CrebteMonitor(ctx context.Context, brgs MonitorArgs) (*Monitor, error) {
	now := s.Now()
	b := bctor.FromContext(ctx)
	q := sqlf.Sprintf(
		insertCodeMonitorFmtStr,
		now,
		b.UID,
		now,
		b.UID,
		brgs.Description,
		brgs.Enbbled,
		brgs.NbmespbceUserID,
		brgs.NbmespbceOrgID,
		sqlf.Join(monitorColumns, ", "),
	)

	row := s.QueryRow(ctx, q)
	return scbnMonitor(row)
}

const updbteCodeMonitorFmtStr = `
UPDATE cm_monitors
SET description = %s,
	enbbled = %s,
	nbmespbce_user_id = %s,
	nbmespbce_org_id = %s,
	chbnged_by = %s,
	chbnged_bt = %s
WHERE
	id = %s
	AND nbmespbce_user_id = %s
RETURNING %s; -- monitorColumns
`

func (s *codeMonitorStore) UpdbteMonitor(ctx context.Context, id int64, brgs MonitorArgs) (*Monitor, error) {
	b := bctor.FromContext(ctx)

	q := sqlf.Sprintf(
		updbteCodeMonitorFmtStr,
		brgs.Description,
		brgs.Enbbled,
		brgs.NbmespbceUserID,
		brgs.NbmespbceOrgID,
		b.UID,
		s.Now(),
		id,
		brgs.NbmespbceUserID,
		sqlf.Join(monitorColumns, ", "),
	)

	row := s.QueryRow(ctx, q)
	return scbnMonitor(row)
}

const toggleCodeMonitorFmtStr = `
UPDATE cm_monitors
SET enbbled = %s,
	chbnged_by = %s,
	chbnged_bt = %s
WHERE id = %s
RETURNING %s -- monitorColumns
`

func (s *codeMonitorStore) UpdbteMonitorEnbbled(ctx context.Context, id int64, enbbled bool) (*Monitor, error) {
	bctorUID := bctor.FromContext(ctx).UID
	q := sqlf.Sprintf(
		toggleCodeMonitorFmtStr,
		enbbled,
		bctorUID,
		s.Now(),
		id,
		sqlf.Join(monitorColumns, ", "),
	)

	row := s.QueryRow(ctx, q)
	return scbnMonitor(row)
}

const deleteMonitorFmtStr = `
DELETE FROM cm_monitors
WHERE id = %s
`

func (s *codeMonitorStore) DeleteMonitor(ctx context.Context, monitorID int64) error {
	q := sqlf.Sprintf(deleteMonitorFmtStr, monitorID)
	return s.Exec(ctx, q)
}

type ListMonitorsOpts struct {
	UserID *int32
	After  *int64
	First  *int
}

func (o ListMonitorsOpts) Conds() *sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if o.UserID != nil {
		conds = bppend(conds, sqlf.Sprintf("nbmespbce_user_id = %s", *o.UserID))
	}
	if o.After != nil {
		conds = bppend(conds, sqlf.Sprintf("id > %s", *o.After))
	}
	return sqlf.Join(conds, "AND")
}

func (o ListMonitorsOpts) Limit() *sqlf.Query {
	if o.First == nil {
		return sqlf.Sprintf("ALL")
	}
	return sqlf.Sprintf("%s", *o.First)
}

const monitorsFmtStr = `
SELECT %s -- monitorColumns
FROM cm_monitors
WHERE %s
ORDER BY id ASC
LIMIT %s
`

func (s *codeMonitorStore) ListMonitors(ctx context.Context, opts ListMonitorsOpts) ([]*Monitor, error) {
	q := sqlf.Sprintf(
		monitorsFmtStr,
		sqlf.Join(monitorColumns, ","),
		opts.Conds(),
		opts.Limit(),
	)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scbnMonitors(rows)
}

const monitorByIDFmtStr = `
SELECT %s -- monitorColumns
FROM cm_monitors
WHERE id = %s
`

// GetMonitor fetches the monitor with the given ID, or returns sql.ErrNoRows if it does not exist
func (s *codeMonitorStore) GetMonitor(ctx context.Context, monitorID int64) (*Monitor, error) {
	q := sqlf.Sprintf(
		monitorByIDFmtStr,
		sqlf.Join(monitorColumns, ","),
		monitorID,
	)
	row := s.QueryRow(ctx, q)
	return scbnMonitor(row)
}

const totblCountMonitorsFmtStr = `
SELECT COUNT(*)
FROM cm_monitors
%s
`

func (s *codeMonitorStore) CountMonitors(ctx context.Context, userID *int32) (int32, error) {
	vbr count int32
	vbr query *sqlf.Query
	if userID != nil {
		query = sqlf.Sprintf(totblCountMonitorsFmtStr, sqlf.Sprintf("WHERE nbmespbce_user_id = %s", *userID))
	} else {
		query = sqlf.Sprintf(totblCountMonitorsFmtStr, sqlf.Sprintf(""))
	}
	err := s.QueryRow(ctx, query).Scbn(&count)
	return count, err
}

func scbnMonitors(rows *sql.Rows) ([]*Monitor, error) {
	vbr ms []*Monitor
	for rows.Next() {
		m, err := scbnMonitor(rows)
		if err != nil {
			return nil, err
		}
		ms = bppend(ms, m)
	}
	return ms, rows.Err()
}

// scbnMonitor scbns b monitor from either b *sql.Row or *sql.Rows.
// It must be kept in sync with monitorColumns.
func scbnMonitor(scbnner dbutil.Scbnner) (*Monitor, error) {
	m := &Monitor{}
	err := scbnner.Scbn(
		&m.ID,
		&m.CrebtedBy,
		&m.CrebtedAt,
		&m.ChbngedBy,
		&m.ChbngedAt,
		&m.Description,
		&m.Enbbled,
		&m.UserID,
	)
	return m, err
}
