package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type Monitor struct {
	ID          int64
	CreatedBy   int32
	CreatedAt   time.Time
	ChangedBy   int32
	ChangedAt   time.Time
	Description string
	Enabled     bool
	UserID      int32
}

// monitorColumns are the columns needed to fill out a Monitor.
// Its fields and order must be kept in sync with scanMonitor.
var monitorColumns = []*sqlf.Query{
	sqlf.Sprintf("cm_monitors.id"),
	sqlf.Sprintf("cm_monitors.created_by"),
	sqlf.Sprintf("cm_monitors.created_at"),
	sqlf.Sprintf("cm_monitors.changed_by"),
	sqlf.Sprintf("cm_monitors.changed_at"),
	sqlf.Sprintf("cm_monitors.description"),
	sqlf.Sprintf("cm_monitors.enabled"),
	sqlf.Sprintf("cm_monitors.namespace_user_id"),
}

type MonitorArgs struct {
	Description     string
	Enabled         bool
	NamespaceUserID *int32
	NamespaceOrgID  *int32
}

const insertCodeMonitorFmtStr = `
INSERT INTO cm_monitors
(created_at, created_by, changed_at, changed_by, description, enabled, namespace_user_id, namespace_org_id)
VALUES (%s,%s,%s,%s,%s,%s,%s,%s)
RETURNING %s -- monitorColumns
`

func (s *codeMonitorStore) CreateMonitor(ctx context.Context, args MonitorArgs) (*Monitor, error) {
	now := s.Now()
	a := actor.FromContext(ctx)
	q := sqlf.Sprintf(
		insertCodeMonitorFmtStr,
		now,
		a.UID,
		now,
		a.UID,
		args.Description,
		args.Enabled,
		args.NamespaceUserID,
		args.NamespaceOrgID,
		sqlf.Join(monitorColumns, ", "),
	)

	row := s.QueryRow(ctx, q)
	return scanMonitor(row)
}

const updateCodeMonitorFmtStr = `
UPDATE cm_monitors
SET description = %s,
	enabled = %s,
	namespace_user_id = %s,
	namespace_org_id = %s,
	changed_by = %s,
	changed_at = %s
WHERE
	id = %s
	AND namespace_user_id = %s
RETURNING %s; -- monitorColumns
`

func (s *codeMonitorStore) UpdateMonitor(ctx context.Context, id int64, args MonitorArgs) (*Monitor, error) {
	a := actor.FromContext(ctx)

	q := sqlf.Sprintf(
		updateCodeMonitorFmtStr,
		args.Description,
		args.Enabled,
		args.NamespaceUserID,
		args.NamespaceOrgID,
		a.UID,
		s.Now(),
		id,
		args.NamespaceUserID,
		sqlf.Join(monitorColumns, ", "),
	)

	row := s.QueryRow(ctx, q)
	return scanMonitor(row)
}

const toggleCodeMonitorFmtStr = `
UPDATE cm_monitors
SET enabled = %s,
	changed_by = %s,
	changed_at = %s
WHERE id = %s
RETURNING %s -- monitorColumns
`

func (s *codeMonitorStore) UpdateMonitorEnabled(ctx context.Context, id int64, enabled bool) (*Monitor, error) {
	actorUID := actor.FromContext(ctx).UID
	q := sqlf.Sprintf(
		toggleCodeMonitorFmtStr,
		enabled,
		actorUID,
		s.Now(),
		id,
		sqlf.Join(monitorColumns, ", "),
	)

	row := s.QueryRow(ctx, q)
	return scanMonitor(row)
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
	// If true, we will filter out monitors that are associated with a user that has
	// been soft-deleted.
	SkipOrphaned bool
}

func (o ListMonitorsOpts) Conds() *sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if o.UserID != nil {
		conds = append(conds, sqlf.Sprintf("cm_monitors.namespace_user_id = %s", *o.UserID))
	}
	if o.After != nil {
		conds = append(conds, sqlf.Sprintf("cm_monitors.id > %s", *o.After))
	}
	if o.SkipOrphaned {
		conds = append(conds, sqlf.Sprintf("users.deleted_at IS NULL"))
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
JOIN users
ON cm_monitors.created_by = users.id
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
	return scanMonitors(rows)
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
	return scanMonitor(row)
}

const totalCountMonitorsFmtStr = `
SELECT COUNT(*)
FROM cm_monitors
JOIN users
ON cm_monitors.created_by = users.id
WHERE %s
`

func (s *codeMonitorStore) CountMonitors(ctx context.Context, opts ListMonitorsOpts) (int32, error) {
	var count int32
	query := sqlf.Sprintf(totalCountMonitorsFmtStr, opts.Conds())
	err := s.QueryRow(ctx, query).Scan(&count)
	return count, err
}

func scanMonitors(rows *sql.Rows) ([]*Monitor, error) {
	var ms []*Monitor
	for rows.Next() {
		m, err := scanMonitor(rows)
		if err != nil {
			return nil, err
		}
		ms = append(ms, m)
	}
	return ms, rows.Err()
}

// scanMonitor scans a monitor from either a *sql.Row or *sql.Rows.
// It must be kept in sync with monitorColumns.
func scanMonitor(scanner dbutil.Scanner) (*Monitor, error) {
	m := &Monitor{}
	err := scanner.Scan(
		&m.ID,
		&m.CreatedBy,
		&m.CreatedAt,
		&m.ChangedBy,
		&m.ChangedAt,
		&m.Description,
		&m.Enabled,
		&m.UserID,
	)
	return m, err
}
