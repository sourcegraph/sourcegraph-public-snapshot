package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type BatchChangeAction struct {
	ID          int64
	Monitor     int64
	BatchChange int64
	Enabled     bool

	CreatedBy int32
	CreatedAt time.Time
	ChangedBy int32
	ChangedAt time.Time
}

const updateBatchChangeActionQuery = `
UPDATE code_monitors_batch_changes
SET enabled = %s,
	batch_change_id = %d,
	changed_by = %s,
	changed_at = %s
WHERE id = %s
RETURNING %s;
`

func (s *codeMonitorStore) UpdateBatchChangeAction(ctx context.Context, id int64, enabled bool, batchChangeID int64) (*BatchChangeAction, error) {
	a := actor.FromContext(ctx)
	q := sqlf.Sprintf(
		updateBatchChangeActionQuery,
		enabled,
		batchChangeID,
		a.UID,
		s.Now(),
		id,
		sqlf.Join(batchChangeActionColumns, ","),
	)

	row := s.QueryRow(ctx, q)
	return scanBatchChangeAction(row)
}

const createBatchChangeActionQuery = `
INSERT INTO code_monitors_batch_changes
(code_monitor_id, batch_change_id, enabled, created_by, created_at, changed_by, changed_at)
VALUES (%s,%s,%s,%s,%s,%s,%s,%s)
RETURNING %s;
`

func (s *codeMonitorStore) CreateBatchChangeAction(ctx context.Context, monitorID int64, enabled bool, batchChangeID int64) (*BatchChangeAction, error) {
	now := s.Now()
	a := actor.FromContext(ctx)
	q := sqlf.Sprintf(
		createBatchChangeActionQuery,
		monitorID,
		batchChangeID,
		enabled,
		a.UID,
		now,
		a.UID,
		now,
		sqlf.Join(batchChangeActionColumns, ","),
	)

	row := s.QueryRow(ctx, q)
	return scanBatchChangeAction(row)
}

const deleteBatchChangeActionQuery = `
DELETE FROM code_monitors_batch_changes
WHERE id in (%s)
	AND code_monitor_id = %s
`

func (s *codeMonitorStore) DeleteBatchChangeActions(ctx context.Context, monitorID int64, actionIDs ...int64) error {
	if len(actionIDs) == 0 {
		return nil
	}

	deleteIDs := make([]*sqlf.Query, 0, len(actionIDs))
	for _, ids := range actionIDs {
		deleteIDs = append(deleteIDs, sqlf.Sprintf("%d", ids))
	}
	q := sqlf.Sprintf(
		deleteBatchChangeActionQuery,
		sqlf.Join(deleteIDs, ","),
		monitorID,
	)

	return s.Exec(ctx, q)
}

const countBatchChangeActionsQuery = `
SELECT COUNT(*)
FROM code_monitors_batch_changes
WHERE code_monitor_id = %s;
`

func (s *codeMonitorStore) CountBatchChangeActions(ctx context.Context, monitorID int64) (int, error) {
	var count int
	err := s.QueryRow(ctx, sqlf.Sprintf(countBatchChangeActionsQuery, monitorID)).Scan(&count)
	return count, err
}

const getBatchChangeActionQuery = `
SELECT %s -- BatchChangeActionColumns
FROM code_monitors_batch_changes
WHERE id = %s
`

func (s *codeMonitorStore) GetBatchChangeAction(ctx context.Context, id int64) (*BatchChangeAction, error) {
	q := sqlf.Sprintf(
		getBatchChangeActionQuery,
		sqlf.Join(batchChangeActionColumns, ","),
		id,
	)
	row := s.QueryRow(ctx, q)
	return scanBatchChangeAction(row)
}

const listBatchChangeActionsQuery = `
SELECT %s -- BatchChangeActionColumns
FROM code_monitors_batch_changes
WHERE %s
ORDER BY id ASC
LIMIT %s;
`

func (s *codeMonitorStore) ListBatchChangeActions(ctx context.Context, opts ListActionsOpts) ([]*BatchChangeAction, error) {
	q := sqlf.Sprintf(
		listBatchChangeActionsQuery,
		sqlf.Join(batchChangeActionColumns, ","),
		opts.Conds(),
		opts.Limit(),
	)
	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanBatchChangeActions(rows)
}

// batchChangeActionColumns is the set of columns in the cm_slack_webhooks table
// This must be kept in sync with scanBatchChangeAction.
var batchChangeActionColumns = []*sqlf.Query{
	sqlf.Sprintf("code_monitors_batch_changes.id"),
	sqlf.Sprintf("code_monitors_batch_changes.code_monitor"),
	sqlf.Sprintf("code_monitors_batch_changes.batch_change"),
	sqlf.Sprintf("code_monitors_batch_changes.enabled"),
	sqlf.Sprintf("code_monitors_batch_changes.created_by"),
	sqlf.Sprintf("code_monitors_batch_changes.created_at"),
	sqlf.Sprintf("code_monitors_batch_changes.changed_by"),
	sqlf.Sprintf("code_monitors_batch_changes.changed_at"),
}

func scanBatchChangeActions(rows *sql.Rows) ([]*BatchChangeAction, error) {
	var ws []*BatchChangeAction
	for rows.Next() {
		w, err := scanBatchChangeAction(rows)
		if err != nil {
			return nil, err
		}
		ws = append(ws, w)
	}
	return ws, rows.Err()
}

// scanBatchChangeAction scans a BatchChangeAction from a *sql.Row or *sql.Rows.
// It must be kept in sync with BatchChangeActionColumns.
func scanBatchChangeAction(scanner dbutil.Scanner) (*BatchChangeAction, error) {
	var w BatchChangeAction
	err := scanner.Scan(
		&w.ID,
		&w.Monitor,
		&w.BatchChange,
		&w.Enabled,
		&w.CreatedBy,
		&w.CreatedAt,
		&w.ChangedBy,
		&w.ChangedAt,
	)
	return &w, err
}
