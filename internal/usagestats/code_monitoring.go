package usagestats

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func GetCodeMonitoringUsageStatistics(ctx context.Context, db dbutil.DB) (*types.CodeMonitoringUsageStatistics, error) {
	const getCodeMonitoringUsageStatisticsQuery = `
SELECT
    codeMonitoringPageViews,
    createCodeMonitorPageViews,
    createCodeMonitorPageViewsWithTriggerQuery,
    createCodeMonitorPageViewsWithoutTriggerQuery,
    manageCodeMonitorPageViews,
    codeMonitorEmailLinkClicks
FROM (
    SELECT
        NULLIF(COUNT(*) FILTER (WHERE name = 'ViewCodeMonitoringPage'), 0) :: INT AS codeMonitoringPageViews,
        NULLIF(COUNT(*) FILTER (WHERE name = 'ViewCreateCodeMonitorPage'), 0) :: INT AS createCodeMonitorPageViews,
        NULLIF(COUNT(*) FILTER (WHERE name = 'ViewCreateCodeMonitorPage' AND (argument->>'hasTriggerQuery')::bool), 0) :: INT AS createCodeMonitorPageViewsWithTriggerQuery,
        NULLIF(COUNT(*) FILTER (WHERE name = 'ViewCreateCodeMonitorPage' AND NOT (argument->>'hasTriggerQuery')::bool), 0) :: INT AS createCodeMonitorPageViewsWithoutTriggerQuery,
        NULLIF(COUNT(*) FILTER (WHERE name = 'ViewManageCodeMonitorPage'), 0) :: INT AS manageCodeMonitorPageViews,
        NULLIF(COUNT(*) FILTER (WHERE name = 'CodeMonitorEmailLinkClicked'), 0) :: INT AS codeMonitorEmailLinkClicks
    FROM event_logs
    WHERE
        name IN (
            'ViewCodeMonitoringPage',
            'ViewCreateCodeMonitorPage',
            'ViewManageCodeMonitorPage',
            'CodeMonitorEmailLinkClicked'
        )
) sub`

	codeMonitoringUsageStats := &types.CodeMonitoringUsageStatistics{}
	if err := db.QueryRowContext(ctx, getCodeMonitoringUsageStatisticsQuery).Scan(
		&codeMonitoringUsageStats.CodeMonitoringPageViews,
		&codeMonitoringUsageStats.CreateCodeMonitorPageViews,
		&codeMonitoringUsageStats.CreateCodeMonitorPageViewsWithTriggerQuery,
		&codeMonitoringUsageStats.CreateCodeMonitorPageViewsWithoutTriggerQuery,
		&codeMonitoringUsageStats.ManageCodeMonitorPageViews,
		&codeMonitoringUsageStats.CodeMonitorEmailLinkClicks,
	); err != nil {
		return nil, err
	}

	return codeMonitoringUsageStats, nil
}
