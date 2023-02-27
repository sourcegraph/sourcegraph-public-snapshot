package insights

import (
	"context"
	"fmt"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// migrateDashboards runs migrateDashboard over each of the given values. The number of successful migrations
// are returned, along with a list of errors that occurred on failing migrations. Each migration is ran in a
// fresh transaction so that failures do not influence one another.
func (m *insightsMigrator) migrateDashboards(ctx context.Context, job insightsMigrationJob, dashboards []settingDashboard, uniqueIDSuffix string) (count int, err error) {
	for _, dashboard := range dashboards {
		if migrationErr := m.migrateDashboard(ctx, job, dashboard, uniqueIDSuffix); migrationErr != nil {
			err = errors.Append(err, migrationErr)
		} else {
			count++
		}
	}

	return count, err
}

func (m *insightsMigrator) migrateDashboard(ctx context.Context, job insightsMigrationJob, dashboard settingDashboard, uniqueIDSuffix string) (err error) {
	if dashboard.ID == "" {
		// Soft-fail this record
		m.logger.Warn("missing dashboard identifier", log.String("owner", getOwnerName(dashboard.UserID, dashboard.OrgID)))
		return nil
	}

	tx, err := m.insightsStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	if count, _, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(
		insightsMigratorMigrateDashboardQuery,
		dashboard.Title,
		dashboardGrantCondition(dashboard),
	))); err != nil {
		return errors.Wrap(err, "failed to count dashboards")
	} else if count != 0 {
		// Already migrated
		return nil
	}

	return m.createDashboard(ctx, tx, job, dashboard.Title, dashboard.InsightIDs, uniqueIDSuffix)
}

const insightsMigratorMigrateDashboardQuery = `
SELECT COUNT(*) from dashboard
JOIN dashboard_grants dg ON dashboard.id = dg.dashboard_id
WHERE dashboard.title = %s AND %s
`

func (m *insightsMigrator) createDashboard(ctx context.Context, tx *basestore.Store, job insightsMigrationJob, title string, insightIDs []string, uniqueIDSuffix string) (err error) {
	// Collect unique IDs matching the given insight + user/org identifiers
	uniqueIDs := make([]string, 0, len(insightIDs))
	for _, insightID := range insightIDs {
		uniqueID, _, err := basestore.ScanFirstString(tx.Query(ctx, sqlf.Sprintf(
			insightsMigratorCreateDashboardSelectQuery,
			insightID,
			fmt.Sprintf("%s-%s", insightID, uniqueIDSuffix),
		)))
		if err != nil {
			return errors.Wrap(err, "failed to retrieve unique id of insight view")
		}

		uniqueIDs = append(uniqueIDs, uniqueID)
	}

	// Create dashboard record
	dashboardID, _, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(insightsMigratorCreateDashboardInsertQuery, title)))
	if err != nil {
		return errors.Wrap(err, "failed to insert dashboard")
	}

	if len(uniqueIDs) > 0 {
		uniqueIDPairs := make([]*sqlf.Query, 0, len(uniqueIDs))
		for i, uniqueID := range uniqueIDs {
			uniqueIDPairs = append(uniqueIDPairs, sqlf.Sprintf("(%s, %s)", uniqueID, fmt.Sprintf("%d", i)))
		}
		values := sqlf.Join(uniqueIDPairs, ", ")

		// Create insight views
		if err := tx.Exec(ctx, sqlf.Sprintf(insightsMigratorCreateDashboardInsertInsightViewQuery, dashboardID, values, pq.Array(uniqueIDs))); err != nil {
			return errors.Wrap(err, "failed to insert dashboard insight view")
		}
	}

	// Create dashboard grants
	grantArgs := append([]any{dashboardID}, grantTiple(job.userID, job.orgID)...)
	if err := tx.Exec(ctx, sqlf.Sprintf(insightsMigratorCreateDashboardInsertGrantQuery, grantArgs...)); err != nil {
		return errors.Wrap(err, "failed to insert dashboard grants")
	}

	return nil
}

const insightsMigratorCreateDashboardSelectQuery = `
SELECT unique_id
FROM insight_view
WHERE unique_id = %s OR unique_id SIMILAR TO %s
LIMIT 1
`

const insightsMigratorCreateDashboardInsertQuery = `
INSERT INTO dashboard (title, save, type)
VALUES (%s, true, 'standard')
RETURNING id
`

const insightsMigratorCreateDashboardInsertInsightViewQuery = `
INSERT INTO dashboard_insight_view (dashboard_id, insight_view_id)
SELECT %s AS dashboard_id, insight_view.id AS insight_view_id
FROM insight_view
JOIN (VALUES %s) AS ids (id, ordering) ON ids.id = insight_view.unique_id
WHERE unique_id = ANY(%s)
ORDER BY ids.ordering
ON CONFLICT DO NOTHING
`

const insightsMigratorCreateDashboardInsertGrantQuery = `
INSERT INTO dashboard_grants (dashboard_id, user_id, org_id, global) VALUES (%s, %s, %s, %s)
`

func dashboardGrantCondition(dashboard settingDashboard) *sqlf.Query {
	if dashboard.UserID != nil {
		return sqlf.Sprintf("dg.user_id = %s", *dashboard.UserID)
	} else if dashboard.OrgID != nil {
		return sqlf.Sprintf("dg.org_id = %s", *dashboard.OrgID)
	} else {
		return sqlf.Sprintf("dg.global IS TRUE")
	}
}
