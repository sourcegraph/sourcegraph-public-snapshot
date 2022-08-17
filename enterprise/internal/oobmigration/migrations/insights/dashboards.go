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

	grantsQuery := func() *sqlf.Query {
		if dashboard.UserID != nil {
			return sqlf.Sprintf("dg.user_id = %s", *dashboard.UserID)
		}
		if dashboard.OrgID != nil {
			return sqlf.Sprintf("dg.org_id = %s", *dashboard.OrgID)
		}
		return sqlf.Sprintf("dg.global IS TRUE")
	}()
	count, _, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(insightsMigratorMigrateDashboardQuery, dashboard.Title, grantsQuery)))
	if err != nil || count != 0 {
		return errors.Wrap(err, "failed to count dashboards")
	}

	return m.createDashboard(ctx, tx, dashboard.Title, dashboard.InsightIDs, userID, orgIDs)
}

const insightsMigratorMigrateDashboardQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/dashboards.go:migrateDashboard
SELECT COUNT(*) from dashboard
JOIN dashboard_grants dg ON dashboard.id = dg.dashboard_id
WHERE dashboard.title = %s AND %s
`

func (m *insightsMigrator) createSpecialCaseDashboard(ctx context.Context, subjectName string, insightIDs []string, userID int, orgIDs []int) (err error) {
	tx, err := m.insightsStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	return m.createDashboard(ctx, tx, specialCaseDashboardName(subjectName), insightIDs, userID, orgIDs)
}

func (m *insightsMigrator) createDashboard(ctx context.Context, tx *basestore.Store, title string, insightIDs []string, userID int, orgIDs []int) (err error) {
	targetsUniqueIDs := make([]string, 0, len(orgIDs)+1)
	if userID != 0 {
		targetsUniqueIDs = append(targetsUniqueIDs, fmt.Sprintf("user-%d", userID))
	}
	for _, orgID := range orgIDs {
		targetsUniqueIDs = append(targetsUniqueIDs, fmt.Sprintf("org-%d", orgID))
	}

	uniqueIDs := make([]string, 0, len(insightIDs))
	for _, insightID := range insightIDs {
		searchTerm := fmt.Sprintf("%s-%%(%s)%%", insightID, strings.Join(targetsUniqueIDs, "|"))
		uniqueID, _, err := basestore.ScanFirstString(tx.Query(ctx, sqlf.Sprintf(insightsMigratorCreateDashboardSelectQuery, insightID, searchTerm)))
		if err != nil {
			return errors.Wrap(err, "failed to retrieve unique id of insight view")
		}

		uniqueIDs = append(uniqueIDs, uniqueID)
	}

	dashboardID, _, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(insightsMigratorCreateDashboardInsertQuery, title)))
	if err != nil {
		return errors.Wrap(err, "failed to insert dashboard")
	}

	if len(uniqueIDs) > 0 {
		indexedUniqueIDs := make([]*sqlf.Query, 0, len(uniqueIDs))
		for i, uniqueID := range uniqueIDs {
			indexedUniqueIDs = append(indexedUniqueIDs, sqlf.Sprintf("(%s, %s)", uniqueID, fmt.Sprintf("%d", i)))
		}
		values := sqlf.Join(indexedUniqueIDs, ", ")

		if err := tx.Exec(ctx, sqlf.Sprintf(insightsMigratorCreateDashboardInsertInsightViewQuery, dashboardID, values, pq.Array(uniqueIDs))); err != nil {
			return errors.Wrap(err, "failed to insert dashboard insight view")
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
		return errors.Wrap(err, "failed to insert dashboard grants")
	}

	return nil
}

const insightsMigratorCreateDashboardSelectQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/dashboards.go:createDashboard
SELECT unique_id
FROM insight_view
WHERE unique_id = %s OR unique_id SIMILAR TO %s
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
SELECT %s AS dashboard_id, insight_view.id AS insight_view_id
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

func specialCaseDashboardName(subjectName string) string {
	if subjectName != "Global" {
		subjectName += "'s"
	}

	return fmt.Sprintf("%s Insights", subjectName)
}
