package usagestats

import (
	"context"
	_ "embed"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

//go:embed code_monitoring_usage_stats.sql
var getCodeMonitoringUsageStatisticsQuery string

func GetCodeMonitoringUsageStatistics(ctx context.Context, db database.DB) (*types.CodeMonitoringUsageStatistics, error) {
	var stats types.CodeMonitoringUsageStatistics
	if err := db.QueryRowContext(ctx, getCodeMonitoringUsageStatisticsQuery).Scan(
		&stats.CodeMonitoringPageViews,
		&stats.CreateCodeMonitorPageViews,
		&stats.CreateCodeMonitorPageViewsWithTriggerQuery,
		&stats.CreateCodeMonitorPageViewsWithoutTriggerQuery,
		&stats.ManageCodeMonitorPageViews,
		&stats.CodeMonitorEmailLinkClicked,
		&stats.ExampleMonitorClicked,
		&stats.GettingStartedPageViewed,
		&stats.CreateFormSubmitted,
		&stats.ManageFormSubmitted,
		&stats.ManageDeleteSubmitted,
		&stats.LogsPageViewed,
		&stats.EmailActionsEnabled,
		&stats.EmailActionsEnabledUniqueUsers,
		&stats.SlackActionsEnabled,
		&stats.SlackActionsEnabledUniqueUsers,
		&stats.WebhookActionsEnabled,
		&stats.WebhookActionsEnabledUniqueUsers,
		&stats.EmailActionsTriggered,
		&stats.EmailActionsTriggeredUniqueUsers,
		&stats.EmailActionsErrored,
		&stats.SlackActionsTriggered,
		&stats.SlackActionsTriggeredUniqueUsers,
		&stats.SlackActionsErrored,
		&stats.WebhookActionsTriggered,
		&stats.WebhookActionsTriggeredUniqueUsers,
		&stats.WebhookActionsErrored,
		&stats.TriggerRuns,
		&stats.TriggerRunsErrored,
		&stats.P50TriggerRunTimeSeconds,
		&stats.P90TriggerRunTimeSeconds,
		&stats.MonitorsEnabled,
		&stats.MonitorsEnabledUniqueUsers,
	); err != nil {
		return nil, err
	}

	return &stats, nil
}
