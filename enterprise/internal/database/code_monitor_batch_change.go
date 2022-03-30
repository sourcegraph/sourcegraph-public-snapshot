package database

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type BatchChangeAction struct {
	ID          int64
	Monitor     int64
	BatchChange int64
	Enabled     bool

	// CreatedBy int32
	// CreatedAt time.Time
	// ChangedBy int32
	// ChangedAt time.Time
}

// batchChangeActionColumns is the set of columns in the cm_slack_webhooks table
// This must be kept in sync with scanBatchChangeAction.
var batchChangeActionColumns = []*sqlf.Query{
	sqlf.Sprintf("code_monitors_batch_changes.id"),
	sqlf.Sprintf("code_monitors_batch_changes.code_monitor"),
	sqlf.Sprintf("code_monitors_batch_changes.batch_change"),
	sqlf.Sprintf("code_monitors_batch_changes.enabled"),
	// sqlf.Sprintf("code_monitors_batch_changes.created_by"),
	// sqlf.Sprintf("code_monitors_batch_changes.created_at"),
	// sqlf.Sprintf("code_monitors_batch_changes.changed_by"),
	// sqlf.Sprintf("code_monitors_batch_changes.changed_at"),
}

const createBatchChangeActionQuery = `
INSERT INTO code_monitors_batch_changes
(code_monitor_id, batch_change_id, enabled) -- created_by, created_at, changed_by, changed_at)
VALUES (%s,%s,%s,%s) -- ,%s,%s,%s,%s)
RETURNING %s;
`

func (s *codeMonitorStore) CreateBatchChangeAction(ctx context.Context, monitorID int64, enabled bool, batchChangeID int64) (*BatchChangeAction, error) {
	// now := s.Now()
	// a := actor.FromContext(ctx)
	q := sqlf.Sprintf(
		createBatchChangeActionQuery,
		monitorID,
		batchChangeID,
		enabled,
		// a.UID,
		// now,
		// a.UID,
		// now,
		sqlf.Join(batchChangeActionColumns, ","),
	)

	row := s.QueryRow(ctx, q)
	return scanBatchChangeAction(row)
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

// scanBatchChangeAction scans a BatchChangeAction from a *sql.Row or *sql.Rows.
// It must be kept in sync with BatchChangeActionColumns.
func scanBatchChangeAction(scanner dbutil.Scanner) (*BatchChangeAction, error) {
	var w BatchChangeAction
	err := scanner.Scan(
		&w.ID,
		&w.Monitor,
		&w.BatchChange,
		&w.Enabled,
		// &w.CreatedBy,
		// &w.CreatedAt,
		// &w.ChangedBy,
		// &w.ChangedAt,
	)
	return &w, err
}
