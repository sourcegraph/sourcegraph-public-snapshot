package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type WebhookAction struct {
	ID             int64
	Monitor        int64
	Enabled        bool
	URL            string
	IncludeResults bool

	CreatedBy int32
	CreatedAt time.Time
	ChangedBy int32
	ChangedAt time.Time
}

const updateWebhookActionQuery = `
UPDATE cm_webhooks
SET enabled = %s,
    include_results = %s,
	url = %s,
	changed_by = %s,
	changed_at = %s
WHERE
	id = %s
	AND EXISTS (
		SELECT 1 FROM cm_monitors
		WHERE cm_monitors.id = cm_webhooks.monitor
			AND %s
	)
RETURNING %s;
`

func (s *codeMonitorStore) UpdateWebhookAction(ctx context.Context, id int64, enabled, includeResults bool, url string) (*WebhookAction, error) {
	a := actor.FromContext(ctx)

	user, err := a.User(ctx, s.userStore)
	if err != nil {
		return nil, err
	}

	q := sqlf.Sprintf(
		updateWebhookActionQuery,
		enabled,
		includeResults,
		url,
		a.UID,
		s.Now(),
		id,
		namespaceScopeQuery(user),
		sqlf.Join(webhookActionColumns, ","),
	)

	row := s.QueryRow(ctx, q)
	return scanWebhookAction(row)
}

const createWebhookActionQuery = `
INSERT INTO cm_webhooks
(monitor, enabled, include_results, url, created_by, created_at, changed_by, changed_at)
VALUES (%s,%s,%s,%s,%s,%s,%s,%s)
RETURNING %s;
`

func (s *codeMonitorStore) CreateWebhookAction(ctx context.Context, monitorID int64, enabled, includeResults bool, url string) (*WebhookAction, error) {
	now := s.Now()
	a := actor.FromContext(ctx)
	q := sqlf.Sprintf(
		createWebhookActionQuery,
		monitorID,
		enabled,
		includeResults,
		url,
		a.UID,
		now,
		a.UID,
		now,
		sqlf.Join(webhookActionColumns, ","),
	)

	row := s.QueryRow(ctx, q)
	return scanWebhookAction(row)
}

const deleteWebhookActionQuery = `
DELETE FROM cm_webhooks
WHERE id in (%s)
	AND MONITOR = %s
`

func (s *codeMonitorStore) DeleteWebhookActions(ctx context.Context, monitorID int64, webhookIDs ...int64) error {
	if len(webhookIDs) == 0 {
		return nil
	}

	deleteIDs := make([]*sqlf.Query, 0, len(webhookIDs))
	for _, ids := range webhookIDs {
		deleteIDs = append(deleteIDs, sqlf.Sprintf("%d", ids))
	}
	q := sqlf.Sprintf(
		deleteWebhookActionQuery,
		sqlf.Join(deleteIDs, ","),
		monitorID,
	)

	return s.Exec(ctx, q)
}

const countWebhookActionsQuery = `
SELECT COUNT(*)
FROM cm_webhooks
WHERE monitor = %s;
`

func (s *codeMonitorStore) CountWebhookActions(ctx context.Context, monitorID int64) (int, error) {
	var count int
	err := s.QueryRow(ctx, sqlf.Sprintf(countWebhookActionsQuery, monitorID)).Scan(&count)
	return count, err
}

const getWebhookActionQuery = `
SELECT %s -- WebhookActionColumns
FROM cm_webhooks
WHERE id = %s
`

func (s *codeMonitorStore) GetWebhookAction(ctx context.Context, webhookID int64) (*WebhookAction, error) {
	q := sqlf.Sprintf(
		getWebhookActionQuery,
		sqlf.Join(webhookActionColumns, ","),
		webhookID,
	)
	row := s.QueryRow(ctx, q)
	return scanWebhookAction(row)
}

const listWebhookActionsQuery = `
SELECT %s -- WebhookActionColumns
FROM cm_webhooks
WHERE %s
ORDER BY id ASC
LIMIT %s;
`

func (s *codeMonitorStore) ListWebhookActions(ctx context.Context, opts ListActionsOpts) ([]*WebhookAction, error) {
	q := sqlf.Sprintf(
		listWebhookActionsQuery,
		sqlf.Join(webhookActionColumns, ","),
		opts.Conds(),
		opts.Limit(),
	)
	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanWebhookActions(rows)
}

// webhookActionColumns is the set of columns in the cm_webhooks table
// This must be kept in sync with scanWebhook
var webhookActionColumns = []*sqlf.Query{
	sqlf.Sprintf("cm_webhooks.id"),
	sqlf.Sprintf("cm_webhooks.monitor"),
	sqlf.Sprintf("cm_webhooks.enabled"),
	sqlf.Sprintf("cm_webhooks.url"),
	sqlf.Sprintf("cm_webhooks.include_results"),
	sqlf.Sprintf("cm_webhooks.created_by"),
	sqlf.Sprintf("cm_webhooks.created_at"),
	sqlf.Sprintf("cm_webhooks.changed_by"),
	sqlf.Sprintf("cm_webhooks.changed_at"),
}

func scanWebhookActions(rows *sql.Rows) ([]*WebhookAction, error) {
	var ws []*WebhookAction
	for rows.Next() {
		w, err := scanWebhookAction(rows)
		if err != nil {
			return nil, err
		}
		ws = append(ws, w)
	}
	return ws, rows.Err()
}

// scanWebhookAction scans a WebhookAction from a *sql.Row or *sql.Rows.
// It must be kept in sync with webhookActionColumns.
func scanWebhookAction(scanner dbutil.Scanner) (*WebhookAction, error) {
	var w WebhookAction
	err := scanner.Scan(
		&w.ID,
		&w.Monitor,
		&w.Enabled,
		&w.URL,
		&w.IncludeResults,
		&w.CreatedBy,
		&w.CreatedAt,
		&w.ChangedBy,
		&w.ChangedAt,
	)
	return &w, err
}
