package insights

import (
	"context"
	"fmt"
	"strings"

	"github.com/inconshreveable/log15"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (m *insightsMigrator) migrateDashboards(ctx context.Context, dashboards []settingDashboard, userID int, orgIDs []int) (count int, err error) {
	for _, dashboard := range dashboards {
		if migrationErr := m.migrateDashboard(ctx, dashboard, userID, orgIDs); migrationErr != nil {
			err = errors.Append(err, migrationErr)
		} else {
			count++
		}
	}

	return count, err
}

func (m *insightsMigrator) migrateDashboard(ctx context.Context, dashboard settingDashboard, userID int, orgIDs []int) (err error) {
	if dashboard.ID == "" {
		// we need a unique ID, and if for some reason this insight doesn't have one, it can't be migrated.
		// since it can never be migrated, we count it towards the total
		log15.Error(schemaErrorPrefix, "owner", getOwnerNameFromDashboard(dashboard), "error msg", "dashboard failed to migrate due to missing id")
		return nil
	}

	tx, err := m.insightsStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	var grantsQuery *sqlf.Query
	if dashboard.UserID != nil {
		grantsQuery = sqlf.Sprintf("dg.user_id = %s", *dashboard.UserID)
	} else if dashboard.OrgID != nil {
		grantsQuery = sqlf.Sprintf("dg.org_id = %s", *dashboard.OrgID)
	} else {
		grantsQuery = sqlf.Sprintf("dg.global IS TRUE")
	}
	count, _, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(insightsMigratorMigrateDashboardQuery, dashboard.Title, grantsQuery)))
	if err != nil || count != 0 {
		return err
	}

	return m.createDashboard(ctx, tx, dashboard.Title, dashboard.InsightIds, userID, orgIDs)
}

const insightsMigratorMigrateDashboardQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/dashboards.go:migrateDashboard
SELECT COUNT(*) from dashboard
JOIN dashboard_grants dg ON dashboard.id = dg.dashboard_id
WHERE dashboard.title = %s AND %s
`

func (m *insightsMigrator) createDashboard(ctx context.Context, tx *basestore.Store, title string, insightReferences []string, userID int, orgIDs []int) (err error) {
	targetsUniqueIDs := make([]string, 0, len(orgIDs)+1)
	if userID != 0 {
		targetsUniqueIDs = append(targetsUniqueIDs, fmt.Sprintf("user-%d", userID))
	}
	for _, orgID := range orgIDs {
		targetsUniqueIDs = append(targetsUniqueIDs, fmt.Sprintf("org-%d", orgID))
	}

	uniqueIDs := make([]string, 0, len(insightReferences))
	for _, reference := range insightReferences {
		id, _, err := basestore.ScanFirstString(m.insightsStore.Query(ctx, sqlf.Sprintf(
			insightsMigratorCreateDashboardSelectQuery,
			reference,
			fmt.Sprintf("%s-%%(%s)%%", reference, strings.Join(targetsUniqueIDs, "|")),
		)))
		if err != nil {
			return err
		}
		uniqueIDs = append(uniqueIDs, id)
	}

	dashboardID, _, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(
		insightsMigratorCreateDashboardInsertQuery,
		title,
	)))
	if err != nil {
		return err
	}

	indexedViewIDs := make([]*sqlf.Query, 0, len(uniqueIDs))
	for i, viewID := range uniqueIDs {
		indexedViewIDs = append(indexedViewIDs, sqlf.Sprintf(
			"(%s, %s)",
			viewID,
			fmt.Sprintf("%d", i),
		))
	}
	if len(indexedViewIDs) > 0 {
		if err := tx.Exec(ctx, sqlf.Sprintf(
			insightsMigratorCreateDashboardInsertInsightViewQuery,
			dashboardID,
			sqlf.Join(indexedViewIDs, ", "),
			pq.Array(uniqueIDs),
		)); err != nil {
			return errors.Wrap(err, "AddViewsToDashboard")
		}
	}

	grantValues := func() []any {
		if userID != 0 {
			return []any{dashboardID, userID, nil, nil}
		}
		if len(orgIDs) != 0 {
			return []any{dashboardID, nil, orgIDs[0], nil}
		}
		return []any{dashboardID, nil, nil, true}
	}()
	if err := tx.Exec(ctx, sqlf.Sprintf(insightsMigratorCreateDashboardInsertGrantQuery, grantValues...)); err != nil {
		return errors.Wrap(err, "AddDashboardGrants")
	}

	return nil
}

const insightsMigratorCreateDashboardSelectQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/dashboards.go:createDashboard
SELECT
	unique_id
FROM insight_view
WHERE
	unique_id = %s OR
	unique_id SIMILAR TO %s
LIMIT 1
`

const insightsMigratorCreateDashboardInsertQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/dashboards.go:createDashboard
INSERT INTO dashboard (title, save, type)
VALUES (%s, true, 'standard')
RETURNING id
`

const insightsMigratorCreateDashboardInsertInsightViewQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/dashboards.go:createDashboard
INSERT INTO dashboard_insight_view (dashboard_id, insight_view_id)
SELECT
	%s AS dashboard_id,
	insight_view.id AS insight_view_id
FROM insight_view
JOIN (VALUES %s) AS ids (id, ordering) ON ids.id = insight_view.unique_id
WHERE unique_id = ANY(%s)
ORDER BY ids.ordering
ON CONFLICT DO NOTHING
`

const insightsMigratorCreateDashboardInsertGrantQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/dashboards.go:createDashboard
INSERT INTO dashboard_grants (dashboard_id, user_id, org_id, global) VALUES (%s, %s, %s, %s)
`

func (m *insightsMigrator) createSpecialCaseDashboard(ctx context.Context, subjectName string, insightReferences []string, userID int, orgIDs []int) error {
	tx, err := m.insightsStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	if subjectName != "Global" {
		subjectName += "'s"
	}

	if err := m.createDashboard(ctx, tx, fmt.Sprintf("%s Insights", subjectName), insightReferences, userID, orgIDs); err != nil {
		return errors.Wrap(err, "CreateSpecialCaseDashboard")
	}
	return nil
}
