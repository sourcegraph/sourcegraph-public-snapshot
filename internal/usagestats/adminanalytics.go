package usagestats

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func GetAdminAnalyticsUsageStatistics(ctx context.Context, db database.DB) (*types.AdminAnalyticsUsageStatistics, error) {
	const query = `
	SELECT
		COUNT(*) FILTER (WHERE name = 'AdminAnalyticsSearchViewed') AS view_search_page_count,
		COUNT(*) FILTER (WHERE name = 'AdminAnalyticsCodeIntelViewed') AS view_code_intel_page_count,
		COUNT(*) FILTER (WHERE name = 'AdminAnalyticsUsersViewed') AS view_users_page_count,
		COUNT(*) FILTER (WHERE name = 'AdminAnalyticsBatchChangesViewed') AS view_batch_changes_page_count,
		COUNT(*) FILTER (WHERE name = 'AdminAnalyticsNotebooksViewed') AS view_notebooks_page_count,
		COUNT(*) FILTER (WHERE name = 'AdminAnalyticsSearchPercentageInputEdited') AS search_page_percentage_change_count,
		COUNT(*) FILTER (WHERE name = 'AdminAnalyticsSearchMinutesInputEdited') AS search_page_minutes_change_count,
		COUNT(*) FILTER (WHERE name = 'AdminAnalyticsCodeIntelPercentageInputEdited') AS code_intel_page_percentage_change_count,
		COUNT(*) FILTER (WHERE name = 'AdminAnalyticsCodeIntelMinutesInputEdited') AS code_intel_page_minutes_change_count,
		COUNT(*) FILTER (WHERE name = 'AdminAnalyticsBatchChangesPercentageInputEdited') AS batch_changes_page_percentage_change_count,
		COUNT(*) FILTER (WHERE name = 'AdminAnalyticsBatchChangesMinutesInputEdited') AS batch_changes_page_minutes_change_count,
		COUNT(*) FILTER (WHERE name = 'AdminAnalyticsNotebooksPercentageInputEdited') AS notebooks_page_percentage_change_count,
		COUNT(*) FILTER (WHERE name = 'AdminAnalyticsNotebooksMinutesInputEdited') AS notebooks_page_minutes_change_count,
		COUNT(*) FILTER (WHERE name = 'AdminAnalyticsSearchDateRangeLAST_WEEKSelected') AS search_date_range_last_week_count,
		COUNT(*) FILTER (WHERE name = 'AdminAnalyticsSearchDateRangeLAST_MONTHSelected') AS search_date_range_last_month_count,
		COUNT(*) FILTER (WHERE name = 'AdminAnalyticsSearchDateRangeLAST_THREE_MONTHSSelected') AS search_date_range_last_three_months_count,
		COUNT(*) FILTER (WHERE name = 'AdminAnalyticsCodeIntelDateRangeLAST_WEEKSelected') AS codeIntel_date_range_last_week_count,
		COUNT(*) FILTER (WHERE name = 'AdminAnalyticsCodeIntelDateRangeLAST_MONTHSelected') AS codeIntel_date_range_last_month_count,
		COUNT(*) FILTER (WHERE name = 'AdminAnalyticsCodeIntelDateRangeLAST_THREE_MONTHSSelected') AS codeIntel_date_range_last_three_months_count,
		COUNT(*) FILTER (WHERE name = 'AdminAnalyticsUsersDateRangeLAST_WEEKSelected') AS users_date_range_last_week_count,
		COUNT(*) FILTER (WHERE name = 'AdminAnalyticsUsersDateRangeLAST_MONTHSelected') AS users_date_range_last_month_count,
		COUNT(*) FILTER (WHERE name = 'AdminAnalyticsUsersDateRangeLAST_THREE_MONTHSSelected') AS users_date_range_last_three_months_count,
		COUNT(*) FILTER (WHERE name = 'AdminAnalyticsBatchChangesDateRangeLAST_WEEKSelected') AS batchChanges_date_range_last_week_count,
		COUNT(*) FILTER (WHERE name = 'AdminAnalyticsBatchChangesDateRangeLAST_MONTHSelected') AS batchChanges_date_range_last_month_count,
		COUNT(*) FILTER (WHERE name = 'AdminAnalyticsBatchChangesDateRangeLAST_THREE_MONTHSSelected') AS batchChanges_date_range_last_three_months_count,
		COUNT(*) FILTER (WHERE name = 'AdminAnalyticsNotebooksDateRangeLAST_WEEKSelected') AS notebooks_date_range_last_week_count,
		COUNT(*) FILTER (WHERE name = 'AdminAnalyticsNotebooksDateRangeLAST_MONTHSelected') AS notebooks_date_range_last_month_count,
		COUNT(*) FILTER (WHERE name = 'AdminAnalyticsNotebooksDateRangeLAST_THREE_MONTHSSelected') AS notebooks_date_range_last_three_months_count,
		COUNT(*) FILTER (WHERE name = 'AdminAnalyticsSearchAggTotalsClicked') AS search_agg_totals_clicks_count,
		COUNT(*) FILTER (WHERE name = 'AdminAnalyticsSearchAggUniquesClicked') AS search_agg_uniques_clicks_count,
		COUNT(*) FILTER (WHERE name = 'AdminAnalyticsCodeIntelAggTotalsClicked') AS codeIntel_agg_totals_clicks_count,
		COUNT(*) FILTER (WHERE name = 'AdminAnalyticsCodeIntelAggUniquesClicked') AS codeIntel_agg_uniques_clicks_count,
		COUNT(*) FILTER (WHERE name = 'AdminAnalyticsUsersAggTotalsClicked') AS users_agg_totals_clicks_count,
		COUNT(*) FILTER (WHERE name = 'AdminAnalyticsUsersAggUniquesClicked') AS users_agg_uniques_clicks_count,
		COUNT(*) FILTER (WHERE name = 'AdminAnalyticsNotebooksAggTotalsClicked') AS notebooks_agg_totals_clicks_count,
		COUNT(*) FILTER (WHERE name = 'AdminAnalyticsNotebooksAggUniquesClicked') AS notebooks_agg_uniques_clicks_count
	FROM event_logs
	WHERE name IN (
		'AdminAnalyticsSearchViewed',
		'AdminAnalyticsCodeIntelViewed',
		'AdminAnalyticsUsersViewed',
		'AdminAnalyticsBatchChangesViewed',
		'AdminAnalyticsNotebooksViewed',
		'AdminAnalyticsSearchPercentageInputEdited',
		'AdminAnalyticsSearchMinutesInputEdited',
		'AdminAnalyticsCodeIntelPercentageInputEdited',
		'AdminAnalyticsCodeIntelMinutesInputEdited',
		'AdminAnalyticsBatchChangesPercentageInputEdited',
		'AdminAnalyticsBatchChangesMinutesInputEdited',
		'AdminAnalyticsNotebooksPercentageInputEdited',
		'AdminAnalyticsNotebooksMinutesInputEdited',
		'AdminAnalyticsSearchDateRangeLAST_WEEKSelected',
		'AdminAnalyticsSearchDateRangeLAST_MONTHSelected',
		'AdminAnalyticsSearchDateRangeLAST_THREE_MONTHSSelected',
		'AdminAnalyticsCodeIntelDateRangeLAST_WEEKSelected',
		'AdminAnalyticsCodeIntelDateRangeLAST_MONTHSelected',
		'AdminAnalyticsCodeIntelDateRangeLAST_THREE_MONTHSSelected',
		'AdminAnalyticsUsersDateRangeLAST_WEEKSelected',
		'AdminAnalyticsUsersDateRangeLAST_MONTHSelected',
		'AdminAnalyticsUsersDateRangeLAST_THREE_MONTHSSelected',
		'AdminAnalyticsBatchChangesDateRangeLAST_WEEKSelected',
		'AdminAnalyticsBatchChangesDateRangeLAST_MONTHSelected',
		'AdminAnalyticsBatchChangesDateRangeLAST_THREE_MONTHSSelected',
		'AdminAnalyticsNotebooksDateRangeLAST_WEEKSelected',
		'AdminAnalyticsNotebooksDateRangeLAST_MONTHSelected',
		'AdminAnalyticsNotebooksDateRangeLAST_THREE_MONTHSSelected',
		'AdminAnalyticsSearchAggTotalsClicked',
		'AdminAnalyticsSearchAggUniquesClicked',
		'AdminAnalyticsCodeIntelAggTotalsClicked',
		'AdminAnalyticsCodeIntelAggUniquesClicked',
		'AdminAnalyticsUsersAggTotalsClicked',
		'AdminAnalyticsUsersAggUniquesClicked',
		'AdminAnalyticsNotebooksAggTotalsClicked',
		'AdminAnalyticsNotebooksAggUniquesClicked'
	)
`

	stats := &types.AdminAnalyticsUsageStatistics{}
	if err := db.QueryRowContext(ctx, query).Scan(
		&stats.AdminAnalyticsSearchPageViews,
		&stats.AdminAnalyticsCodeIntelPageViews,
		&stats.AdminAnalyticsUsersPageViews,
		&stats.AdminAnalyticsBatchChangesPageViews,
		&stats.AdminAnalyticsNotebooksPageViews,
		&stats.AdminAnalyticsSearchPercentageInputEdited,
		&stats.AdminAnalyticsSearchMinutesInputEdited,
		&stats.AdminAnalyticsCodeIntelPercentageInputEdited,
		&stats.AdminAnalyticsCodeIntelMinutesInputEdited,
		&stats.AdminAnalyticsBatchChangesPercentageInputEdited,
		&stats.AdminAnalyticsBatchChangesMinutesInputEdited,
		&stats.AdminAnalyticsNotebooksPercentageInputEdited,
		&stats.AdminAnalyticsNotebooksMinutesInputEdited,
		&stats.AdminAnalyticsSearchDateRangeLAST_WEEKSelected,
		&stats.AdminAnalyticsSearchDateRangeLAST_MONTHSelected,
		&stats.AdminAnalyticsSearchDateRangeLAST_THREE_MONTHSSelected,
		&stats.AdminAnalyticsCodeIntelDateRangeLAST_WEEKSelected,
		&stats.AdminAnalyticsCodeIntelDateRangeLAST_MONTHSelected,
		&stats.AdminAnalyticsCodeIntelDateRangeLAST_THREE_MONTHSSelected,
		&stats.AdminAnalyticsUsersDateRangeLAST_WEEKSelected,
		&stats.AdminAnalyticsUsersDateRangeLAST_MONTHSelected,
		&stats.AdminAnalyticsUsersDateRangeLAST_THREE_MONTHSSelected,
		&stats.AdminAnalyticsBatchChangesDateRangeLAST_WEEKSelected,
		&stats.AdminAnalyticsBatchChangesDateRangeLAST_MONTHSelected,
		&stats.AdminAnalyticsBatchChangesDateRangeLAST_THREE_MONTHSSelected,
		&stats.AdminAnalyticsNotebooksDateRangeLAST_WEEKSelected,
		&stats.AdminAnalyticsNotebooksDateRangeLAST_MONTHSelected,
		&stats.AdminAnalyticsNotebooksDateRangeLAST_THREE_MONTHSSelected,
		&stats.AdminAnalyticsSearchAggTotalsClicked,
		&stats.AdminAnalyticsSearchAggUniquesClicked,
		&stats.AdminAnalyticsCodeIntelAggTotalsClicked,
		&stats.AdminAnalyticsCodeIntelAggUniquesClicked,
		&stats.AdminAnalyticsUsersAggTotalsClicked,
		&stats.AdminAnalyticsUsersAggUniquesClicked,
		&stats.AdminAnalyticsNotebooksAggTotalsClicked,
		&stats.AdminAnalyticsNotebooksAggUniquesClicked,
	); err != nil {
		return nil, err
	}

	return stats, nil
}
