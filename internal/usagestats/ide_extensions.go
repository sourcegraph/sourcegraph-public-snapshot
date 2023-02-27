package usagestats

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func GetIDEExtensionsUsageStatistics(ctx context.Context, db database.DB) (*types.IDEExtensionsUsage, error) {
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
			&ideExtensionUsage.Month.SearchesPerformed.UniquesCount,
			&ideExtensionUsage.Month.SearchesPerformed.TotalCount,
			&ideExtensionUsage.Week.StartTime,
			&ideExtensionUsage.Week.SearchesPerformed.UniquesCount,
			&ideExtensionUsage.Week.SearchesPerformed.TotalCount,
			&ideExtensionUsage.Day.StartTime,
			&ideExtensionUsage.Day.SearchesPerformed.UniquesCount,
			&ideExtensionUsage.Day.SearchesPerformed.TotalCount,
			&ideExtensionUsage.Day.UserState.Installs,
			&ideExtensionUsage.Day.UserState.Uninstalls,
			&ideExtensionUsage.Day.RedirectsCount,
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
		WHERE
			timestamp >= DATE_TRUNC('month', TIMEZONE('UTC', $1::timestamp))
			AND
			(
				source = 'IDEEXTENSION'
				OR
				(
					source = 'BACKEND'
					AND
					(
						name LIKE 'IDE%'
						OR
						name = 'VSCESearchSubmitted'
					)
				)
			)
	)
	SELECT
		ide_kind,
		current_month,
		COUNT(DISTINCT user_id) FILTER (WHERE (name = 'IDESearchSubmitted' OR name = 'VSCESearchSubmitted') AND month = current_month) AS monthly_uniques_searches,
		COUNT(*) FILTER (WHERE (name = 'IDESearchSubmitted' OR name = 'VSCESearchSubmitted') AND month = current_month) AS monthly_total_searches,
		current_week,
		COUNT(DISTINCT user_id) FILTER (WHERE (name = 'IDESearchSubmitted' OR name = 'VSCESearchSubmitted') AND timestamp > current_week) AS weekly_uniques_searches,
		COUNT(*) FILTER (WHERE (name = 'IDESearchSubmitted' OR name = 'VSCESearchSubmitted') AND week = current_week) AS weekly_total_searches,
		current_day,
		COUNT(DISTINCT user_id) FILTER (WHERE (name = 'IDESearchSubmitted' OR name = 'VSCESearchSubmitted') AND day = current_day) AS daily_uniques_searches,
		COUNT(*) FILTER (WHERE (name = 'IDESearchSubmitted' OR name = 'VSCESearchSubmitted') AND day = current_day) AS daily_total_searches,
		COUNT(DISTINCT user_id) FILTER (WHERE name = 'IDEInstalled' AND day = current_day) AS daily_installs,
		COUNT(DISTINCT user_id) FILTER (WHERE name = 'IDEUninstalled' AND day = current_day) AS daily_uninstalls,
		COUNT(*) FILTER (WHERE name = 'IDERedirected' AND day = current_day) AS daily_redirects
	FROM events
	GROUP BY ide_kind, current_month, current_week, current_day;
`
