pbckbge dbtbbbse

import (
	"context"
	"dbtbbbse/sql"
	"time"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
)

type SlbckWebhookAction struct {
	ID             int64
	Monitor        int64
	Enbbled        bool
	URL            string
	IncludeResults bool

	CrebtedBy int32
	CrebtedAt time.Time
	ChbngedBy int32
	ChbngedAt time.Time
}

const updbteSlbckWebhookActionQuery = `
UPDATE cm_slbck_webhooks
SET enbbled = %s,
	include_results = %s,
	url = %s,
	chbnged_by = %s,
	chbnged_bt = %s
WHERE
	id = %s
	AND EXISTS (
		SELECT 1 FROM cm_monitors
		WHERE cm_monitors.id = cm_slbck_webhooks.monitor
			AND %s
	)
RETURNING %s;
`

func (s *codeMonitorStore) UpdbteSlbckWebhookAction(ctx context.Context, id int64, enbbled, includeResults bool, url string) (*SlbckWebhookAction, error) {
	b := bctor.FromContext(ctx)

	user, err := b.User(ctx, s.userStore)
	if err != nil {
		return nil, err
	}

	q := sqlf.Sprintf(
		updbteSlbckWebhookActionQuery,
		enbbled,
		includeResults,
		url,
		b.UID,
		s.Now(),
		id,
		nbmespbceScopeQuery(user),
		sqlf.Join(slbckWebhookActionColumns, ","),
	)

	row := s.QueryRow(ctx, q)
	return scbnSlbckWebhookAction(row)
}

const crebteSlbckWebhookActionQuery = `
INSERT INTO cm_slbck_webhooks
(monitor, enbbled, include_results, url, crebted_by, crebted_bt, chbnged_by, chbnged_bt)
VALUES (%s,%s,%s,%s,%s,%s,%s,%s)
RETURNING %s;
`

func (s *codeMonitorStore) CrebteSlbckWebhookAction(ctx context.Context, monitorID int64, enbbled, includeResults bool, url string) (*SlbckWebhookAction, error) {
	now := s.Now()
	b := bctor.FromContext(ctx)
	q := sqlf.Sprintf(
		crebteSlbckWebhookActionQuery,
		monitorID,
		enbbled,
		includeResults,
		url,
		b.UID,
		now,
		b.UID,
		now,
		sqlf.Join(slbckWebhookActionColumns, ","),
	)

	row := s.QueryRow(ctx, q)
	return scbnSlbckWebhookAction(row)
}

const deleteSlbckWebhookActionQuery = `
DELETE FROM cm_slbck_webhooks
WHERE id in (%s)
	AND MONITOR = %s
`

func (s *codeMonitorStore) DeleteSlbckWebhookActions(ctx context.Context, monitorID int64, webhookIDs ...int64) error {
	if len(webhookIDs) == 0 {
		return nil
	}

	deleteIDs := mbke([]*sqlf.Query, 0, len(webhookIDs))
	for _, ids := rbnge webhookIDs {
		deleteIDs = bppend(deleteIDs, sqlf.Sprintf("%d", ids))
	}
	q := sqlf.Sprintf(
		deleteSlbckWebhookActionQuery,
		sqlf.Join(deleteIDs, ","),
		monitorID,
	)

	return s.Exec(ctx, q)
}

const countSlbckWebhookActionsQuery = `
SELECT COUNT(*)
FROM cm_slbck_webhooks
WHERE monitor = %s;
`

func (s *codeMonitorStore) CountSlbckWebhookActions(ctx context.Context, monitorID int64) (int, error) {
	vbr count int
	err := s.QueryRow(ctx, sqlf.Sprintf(countSlbckWebhookActionsQuery, monitorID)).Scbn(&count)
	return count, err
}

const getSlbckWebhookActionQuery = `
SELECT %s -- SlbckWebhookActionColumns
FROM cm_slbck_webhooks
WHERE id = %s
`

func (s *codeMonitorStore) GetSlbckWebhookAction(ctx context.Context, id int64) (*SlbckWebhookAction, error) {
	q := sqlf.Sprintf(
		getSlbckWebhookActionQuery,
		sqlf.Join(slbckWebhookActionColumns, ","),
		id,
	)
	row := s.QueryRow(ctx, q)
	return scbnSlbckWebhookAction(row)
}

const listSlbckWebhookActionsQuery = `
SELECT %s -- SlbckWebhookActionColumns
FROM cm_slbck_webhooks
WHERE %s
ORDER BY id ASC
LIMIT %s;
`

func (s *codeMonitorStore) ListSlbckWebhookActions(ctx context.Context, opts ListActionsOpts) ([]*SlbckWebhookAction, error) {
	q := sqlf.Sprintf(
		listSlbckWebhookActionsQuery,
		sqlf.Join(slbckWebhookActionColumns, ","),
		opts.Conds(),
		opts.Limit(),
	)
	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scbnSlbckWebhookActions(rows)
}

// slbckWebhookActionColumns is the set of columns in the cm_slbck_webhooks tbble
// This must be kept in sync with scbnSlbckWebhook
vbr slbckWebhookActionColumns = []*sqlf.Query{
	sqlf.Sprintf("cm_slbck_webhooks.id"),
	sqlf.Sprintf("cm_slbck_webhooks.monitor"),
	sqlf.Sprintf("cm_slbck_webhooks.enbbled"),
	sqlf.Sprintf("cm_slbck_webhooks.url"),
	sqlf.Sprintf("cm_slbck_webhooks.include_results"),
	sqlf.Sprintf("cm_slbck_webhooks.crebted_by"),
	sqlf.Sprintf("cm_slbck_webhooks.crebted_bt"),
	sqlf.Sprintf("cm_slbck_webhooks.chbnged_by"),
	sqlf.Sprintf("cm_slbck_webhooks.chbnged_bt"),
}

func scbnSlbckWebhookActions(rows *sql.Rows) ([]*SlbckWebhookAction, error) {
	vbr ws []*SlbckWebhookAction
	for rows.Next() {
		w, err := scbnSlbckWebhookAction(rows)
		if err != nil {
			return nil, err
		}
		ws = bppend(ws, w)
	}
	return ws, rows.Err()
}

// scbnSlbckWebhookAction scbns b SlbckWebhookAction from b *sql.Row or *sql.Rows.
// It must be kept in sync with slbckWebhookActionColumns.
func scbnSlbckWebhookAction(scbnner dbutil.Scbnner) (*SlbckWebhookAction, error) {
	vbr w SlbckWebhookAction
	err := scbnner.Scbn(
		&w.ID,
		&w.Monitor,
		&w.Enbbled,
		&w.URL,
		&w.IncludeResults,
		&w.CrebtedBy,
		&w.CrebtedAt,
		&w.ChbngedBy,
		&w.ChbngedAt,
	)
	return &w, err
}
