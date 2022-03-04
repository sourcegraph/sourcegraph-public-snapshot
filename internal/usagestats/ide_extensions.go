package usagestats

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func GetIdeExtensionsUsageStatistics(ctx context.Context, db database.DB) (*types.IDEExtensionsUsage, error) {
	stats := types.IDEExtensionsUsage{}

	usageStatisticsByIdext := []*types.IDEExtensionsUsageStatistics{}

	rows, err := db.QueryContext(ctx, ideExtensionsPeriodUsageQuery, timeNow())
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		ideExtensionUsage := types.IDEExtensionsUsageStatistics{}

		if err := rows.Scan(
			&ideExtensionUsage.IdeKind,
			&ideExtensionUsage.Month.StartTime,
			&ideExtensionUsage.Month.SearchPerformed.UniqueCount,
			&ideExtensionUsage.Month.SearchPerformed.TotalCount,
			&ideExtensionUsage.Month.RedirectCount,
			&ideExtensionUsage.Month.UserState.Installs,
			&ideExtensionUsage.Month.UserState.Uninstalls,
			&ideExtensionUsage.Week.StartTime,
			&ideExtensionUsage.Week.SearchPerformed.UniqueCount,
			&ideExtensionUsage.Week.SearchPerformed.TotalCount,
			&ideExtensionUsage.Day.StartTime,
			&ideExtensionUsage.Day.SearchPerformed.UniqueCount,
			&ideExtensionUsage.Day.SearchPerformed.TotalCount,
		); err != nil {
			return nil, err
		}

		usageStatisticsByIdext = append(usageStatisticsByIdext, &ideExtensionUsage)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	stats.IDEs = usageStatisticsByIdext

	return &stats, nil

}

var ideExtensionsPeriodUsageQuery = `
	WITH events AS (
		SELECT
			public_argument ->> 'editor'::text AS ide_kind,
			name,
			user_id,
			public_argument,
			source,
			timestamp,
			DATE_TRUNC('month', TIMEZONE('UTC', timestamp)) as month,
			DATE_TRUNC('week', TIMEZONE('UTC', timestamp)) as week,
			DATE_TRUNC('day', TIMEZONE('UTC', timestamp)) as day,
			DATE_TRUNC('month', TIMEZONE('UTC', $1::timestamp)) as current_month,
			DATE_TRUNC('week', TIMEZONE('UTC', $1::timestamp)) as current_week,
			DATE_TRUNC('day', TIMEZONE('UTC', $1::timestamp)) as current_day
		FROM event_logs
		WHERE timestamp >= DATE_TRUNC('month', TIMEZONE('UTC', $1::timestamp)) AND source = 'IDEEXTENSION' OR (source = 'BACKEND' AND name LIKE 'IDE%')
	)
	SELECT
		ide_kind,
		current_month,
		COUNT(DISTINCT user_id) FILTER (WHERE name = 'IDESearchSubmitted' AND timestamp > current_month),
		COUNT(*) FILTER (WHERE name = 'IDESearchSubmitted' AND timestamp > current_month),
		COUNT(*) FILTER (WHERE name = 'IDERedirects' AND timestamp > current_month),
		COUNT(DISTINCT user_id) FILTER (WHERE name = 'IDEInstalled' AND (SELECT MIN(timestamp) FROM events) > current_month),
		COUNT(DISTINCT user_id) FILTER (WHERE name = 'IDEUninstalled' AND timestamp > current_month),
		current_week,
		COUNT(DISTINCT user_id) FILTER (WHERE name = 'IDESearchSubmitted' AND timestamp > current_week),
		COUNT(*) FILTER (WHERE name = 'IDESearchSubmitted' AND timestamp > current_week),
		current_day,
		COUNT(DISTINCT user_id) FILTER (WHERE name = 'IDESearchSubmitted' AND timestamp > current_day),
		COUNT(*) FILTER (WHERE name = 'IDESearchSubmitted' AND timestamp > current_day)
	FROM events
	GROUP BY ide_kind, current_month, current_week, current_day;
`
