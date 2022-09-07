package usagestats

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func GetMigratedExtensionsUsageStatistics(ctx context.Context, db database.DB) (*types.MigratedExtensionsUsageStatistics, error) {
	stats := types.MigratedExtensionsUsageStatistics{}

	if err := db.QueryRowContext(ctx, MigratedExtensionsUsageQuery).Scan(
		&stats.GitBlameEnabled,
		&stats.GitBlameEnabledUniqueUsers,
		&stats.GitBlameDisabled,
		&stats.GitBlameDisabledUniqueUsers,
		&stats.GitBlamePopupViewed,
		&stats.GitBlamePopupViewedUniqueUsers,
		&stats.GitBlamePopupClicked,
		&stats.GitBlamePopupClickedUniqueUsers,

		&stats.SearchExportPerformed,
		&stats.SearchExportPerformedUniqueUsers,
		&stats.SearchExportFailed,
		&stats.SearchExportFailedUniqueUsers,

		&stats.GoImportsSearchQueryTransformed,
		&stats.GoImportsSearchQueryTransformedUniqueUsers,
	); err != nil {
		return nil, err
	}

	return &stats, nil
}

var MigratedExtensionsUsageQuery = `
	WITH event_log_stats AS (
		SELECT
			NULLIF(COUNT(*) FILTER (WHERE name = 'GitBlameEnabled'), 0) :: INT AS git_blame_enabled,
			NULLIF(COUNT(DISTINCT event_logs.user_id) FILTER (WHERE name = 'GitBlameEnabled'), 0) :: INT AS git_blame_enabled_unique_users,
			NULLIF(COUNT(*) FILTER (WHERE name = 'GitBlameDisabled'), 0) :: INT AS git_blame_disabled,
			NULLIF(COUNT(DISTINCT event_logs.user_id) FILTER (WHERE name = 'GitBlameDisabled'), 0) :: INT AS git_blame_disabled_unique_users,
			NULLIF(COUNT(*) FILTER (WHERE name = 'GitBlamePopupViewed'), 0) :: INT AS git_blame_popup_viewed,
			NULLIF(COUNT(DISTINCT event_logs.user_id) FILTER (WHERE name = 'GitBlamePopupViewed'), 0) :: INT AS git_blame_popup_viewed_unique_users,
			NULLIF(COUNT(*) FILTER (WHERE name = 'GitBlamePopupClicked'), 0) :: INT AS git_blame_popup_clicked,
			NULLIF(COUNT(DISTINCT event_logs.user_id) FILTER (WHERE name = 'GitBlamePopupClicked'), 0) :: INT AS git_blame_popup_clicked_unique_users,

			NULLIF(COUNT(*) FILTER (WHERE name = 'SearchExportPerformed'), 0) :: INT AS search_export_performed,
			NULLIF(COUNT(DISTINCT event_logs.user_id) FILTER (WHERE name = 'SearchExportPerformed'), 0) :: INT AS search_export_performed_unique_users,
			NULLIF(COUNT(*) FILTER (WHERE name = 'SearchExportFailed'), 0) :: INT AS search_export_failed,
			NULLIF(COUNT(DISTINCT event_logs.user_id) FILTER (WHERE name = 'SearchExportFailed'), 0) :: INT AS search_export_failed_unique_users,

			NULLIF(COUNT(*) FILTER (WHERE name = 'GoImportsSearchQueryTransformed'), 0) :: INT AS go_imports_search_query_transformed,
			NULLIF(COUNT(DISTINCT event_logs.user_id) FILTER (WHERE name = 'GoImportsSearchQueryTransformed'), 0) :: INT AS go_imports_search_query_transformed_unique_users
		FROM event_logs
		WHERE
			name IN (
				'GitBlameEnabled',
				'GitBlameDisabled',
				'GitBlamePopupViewed',
				'GitBlamePopupClicked',

				'SearchExportPerformed',
				'SearchExportFailed',

				'GoImportsSearchQueryTransformed'
			)
	)
	SELECT
		event_log_stats.git_blame_enabled,
		event_log_stats.git_blame_enabled_unique_users,
		event_log_stats.git_blame_disabled,
		event_log_stats.git_blame_disabled_unique_users,
		event_log_stats.git_blame_popup_viewed,
		event_log_stats.git_blame_popup_viewed_unique_users,
		event_log_stats.git_blame_popup_clicked,
		event_log_stats.git_blame_popup_clicked_unique_users,


		event_log_stats.search_export_performed,
		event_log_stats.search_export_performed_unique_users,
		event_log_stats.search_export_failed,
		event_log_stats.search_export_failed_unique_users,

		event_log_stats.go_imports_search_query_transformed,
		event_log_stats.go_imports_search_query_transformed_unique_users
	FROM
		event_log_stats;
`
