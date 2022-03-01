package usagestats

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func GetIdeExtensionsUsageStatistics(ctx context.Context, db database.DB) (*types.IdeExtensionsUsage, error) {
	stats := types.IdeExtensionsUsage{}

	usageStatisticsByIdext := []*types.IdeExtensionsUsageStatistics{}

	rows, err := db.QueryContext(ctx, ideExtensionsPeriodUsageQuery, timeNow())
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		ideaExtensionUsage := types.IdeExtensionsUsageStatistics{}

		if err := rows.Scan(
			&ideaExtensionUsage.IdeKind,
			&ideaExtensionUsage.Month.StartTime,
			&ideaExtensionUsage.Month.UserCount,
			&ideaExtensionUsage.Month.SearchPerformed.UniqueCount,
			&ideaExtensionUsage.Month.SearchPerformed.TotalCount,
			&ideaExtensionUsage.Month.RedirectCount,
			&ideaExtensionUsage.Month.MonthlyUserState.Installs,
			&ideaExtensionUsage.Month.MonthlyUserState.Uninstalls,
			&ideaExtensionUsage.Week.StartTime,
			&ideaExtensionUsage.Week.UserCount,
			&ideaExtensionUsage.Week.SearchPerformed.UniqueCount,
			&ideaExtensionUsage.Week.SearchPerformed.TotalCount,
			&ideaExtensionUsage.Week.RedirectCount,
			&ideaExtensionUsage.Day.StartTime,
			&ideaExtensionUsage.Day.UserCount,
			&ideaExtensionUsage.Day.SearchPerformed.UniqueCount,
			&ideaExtensionUsage.Day.SearchPerformed.TotalCount,
			&ideaExtensionUsage.Day.RedirectCount,
		); err != nil {
			return nil, err
		}

		usageStatisticsByIdext = append(usageStatisticsByIdext, &ideaExtensionUsage)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Get VSCE data from older instances that do not support EventSource.IDEEXTENSION
	oldVsceUsageStats := types.IdeExtensionsUsageStatistics{}

	oldVsceUsageStats.IdeKind = "vscode_pre_ide"

	if err := db.QueryRowContext(ctx, ideExtensionsOldVSCEPeriodUsageQuery, timeNow()).Scan(
		&oldVsceUsageStats.Month.StartTime,
		&oldVsceUsageStats.Month.UserCount,
		&oldVsceUsageStats.Month.SearchPerformed.UniqueCount,
		&oldVsceUsageStats.Month.SearchPerformed.TotalCount,
		&oldVsceUsageStats.Month.RedirectCount,
		&oldVsceUsageStats.Month.MonthlyUserState.Installs,
		&oldVsceUsageStats.Month.MonthlyUserState.Uninstalls,
		&oldVsceUsageStats.Week.StartTime,
		&oldVsceUsageStats.Week.UserCount,
		&oldVsceUsageStats.Week.SearchPerformed.UniqueCount,
		&oldVsceUsageStats.Week.SearchPerformed.TotalCount,
		&oldVsceUsageStats.Week.RedirectCount,
		&oldVsceUsageStats.Day.StartTime,
		&oldVsceUsageStats.Day.UserCount,
		&oldVsceUsageStats.Day.SearchPerformed.UniqueCount,
		&oldVsceUsageStats.Day.SearchPerformed.TotalCount,
		&oldVsceUsageStats.Day.RedirectCount,
	); err != nil {
		return nil, err
	}

	usageStatisticsByIdext = append(usageStatisticsByIdext, &oldVsceUsageStats)

	stats.IDEs = usageStatisticsByIdext

	return &stats, nil

}

var ideExtensionsPeriodUsageQuery = `
	WITH events AS (
		SELECT
			argument ->> 'platform'::text AS ide_kind,
			name,
			user_id,
			argument,
			source,
			timestamp,
			DATE_TRUNC('month', TIMEZONE('UTC', timestamp)) as month,
			DATE_TRUNC('week', TIMEZONE('UTC', timestamp)) as week,
			DATE_TRUNC('day', TIMEZONE('UTC', timestamp)) as day,
			DATE_TRUNC('month', TIMEZONE('UTC', $1::timestamp)) as current_month,
			DATE_TRUNC('week', TIMEZONE('UTC', $1::timestamp)) as current_week,
			DATE_TRUNC('day', TIMEZONE('UTC', $1::timestamp)) as current_day
		FROM event_logs
		WHERE timestamp >= DATE_TRUNC('month', TIMEZONE('UTC', $1::timestamp)) AND source = 'IDEEXTENSION'
	)
	SELECT
		ide_kind,
		current_month,
		COUNT(DISTINCT user_id) FILTER (WHERE timestamp > current_month),
		COUNT(DISTINCT user_id) FILTER (WHERE name = 'IDESearchSubmitted' AND timestamp > current_month),
		COUNT(*) FILTER (WHERE name = 'IDESearchSubmitted' AND timestamp > current_month),
		COUNT(*) FILTER (WHERE name = 'IDERedirects' AND timestamp > current_month),
		COUNT(DISTINCT user_id) FILTER (WHERE (SELECT MIN(timestamp) FROM events) > current_month),
		COUNT(DISTINCT user_id) FILTER (WHERE timestamp > current_month AND name = 'IDEUninstalled'),
		current_week,
		COUNT(DISTINCT user_id) FILTER (WHERE timestamp > current_week),
		COUNT(DISTINCT user_id) FILTER (WHERE name = 'IDESearchSubmitted' AND timestamp > current_week),
		COUNT(*) FILTER (WHERE name = 'IDESearchSubmitted' AND timestamp > current_week),
		COUNT(*) FILTER (WHERE name = 'IDERedirects' AND timestamp > current_week),
		current_day,
		COUNT(DISTINCT user_id) FILTER (WHERE timestamp > current_day),
		COUNT(DISTINCT user_id) FILTER (WHERE name = 'IDESearchSubmitted' AND timestamp > current_day),
		COUNT(*) FILTER (WHERE name = 'IDESearchSubmitted' AND timestamp > current_day),
		COUNT(*) FILTER (WHERE name = 'IDERedirects' AND timestamp > current_day)
	FROM events
	GROUP BY ide_kind, current_month, current_week, current_day;
`

var ideExtensionsOldVSCEPeriodUsageQuery = `
	WITH events AS (
		SELECT
			name,
			user_id,
			url,
			timestamp,
			DATE_TRUNC('month', TIMEZONE('UTC', timestamp)) as month,
			DATE_TRUNC('week', TIMEZONE('UTC', timestamp)) as week,
			DATE_TRUNC('day', TIMEZONE('UTC', timestamp)) as day,
			DATE_TRUNC('month', TIMEZONE('UTC', $1::timestamp)) as current_month,
			DATE_TRUNC('week', TIMEZONE('UTC', $1::timestamp)) as current_week,
			DATE_TRUNC('day', TIMEZONE('UTC', $1::timestamp)) as current_day
		FROM event_logs
		WHERE timestamp >= DATE_TRUNC('month', TIMEZONE('UTC', $1::timestamp)) AND source <> 'IDEEXTENSION' AND (name LIKE 'VSCE%' OR name LIKE 'IDE%' OR (url LIKE '%&utm_source=VSCode-%' AND name = 'ViewBlob'))
	)
	SELECT
		current_month,
		COUNT(DISTINCT user_id) FILTER (WHERE timestamp > current_month),
		COUNT(DISTINCT user_id) FILTER (WHERE timestamp > current_month AND name = 'VSCESearchSubmitted'),
		COUNT(*) FILTER (WHERE timestamp > current_month AND name = 'VSCESearchSubmitted'),
		COUNT(*) FILTER (WHERE timestamp > current_month AND name = 'ViewBlob' AND url LIKE '%&utm_source=VSCode-%'),
		COUNT(DISTINCT user_id) FILTER (WHERE (SELECT MIN(timestamp) FROM events) > current_month),
		COUNT(DISTINCT user_id) FILTER (WHERE timestamp > current_month AND name = 'IDEUninstalled'),
		current_week,
		COUNT(DISTINCT user_id) FILTER (WHERE timestamp > current_week),
		COUNT(DISTINCT user_id) FILTER (WHERE timestamp > current_week AND name = 'VSCESearchSubmitted'),
		COUNT(*) FILTER (WHERE timestamp > current_week AND name = 'VSCESearchSubmitted'),
		COUNT(*) FILTER (WHERE timestamp > current_week AND name = 'ViewBlob' AND url LIKE '%&utm_source=VSCode-%'),
		current_day,
		COUNT(DISTINCT user_id) FILTER (WHERE timestamp > current_day),
		COUNT(DISTINCT user_id) FILTER (WHERE timestamp > current_day AND name = 'VSCESearchSubmitted'),
		COUNT(*) FILTER (WHERE timestamp > current_day AND name = 'VSCESearchSubmitted'),
		COUNT(*) FILTER (WHERE timestamp > current_day AND name = 'ViewBlob' AND url LIKE '%&utm_source=VSCode-%')
	FROM events
	GROUP BY current_month, current_week, current_day;
`
