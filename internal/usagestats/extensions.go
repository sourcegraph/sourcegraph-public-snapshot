package usagestats

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func GetExtensionsUsageStatistics(ctx context.Context, db dbutil.DB) (*types.ExtensionsUsageStatistics, error) {
	stats := types.ExtensionsUsageStatistics{}

	// Query for evaluating success of individual extensions
	extensionsQuery := `
	SELECT
		argument ->> 'extension_id'::text          AS extension_id,
		COUNT(DISTINCT user_id)                    AS user_count,
		COUNT(*)::decimal/COUNT(DISTINCT user_id)  AS average_activations
	FROM event_logs
	WHERE
		event_logs.name = 'ExtensionActivation'
			AND timestamp > DATE_TRUNC('week', $1::timestamp)
	GROUP BY extension_id;
	`

	usageStatisticsByExtension := []*types.ExtensionUsageStatistics{}
	rows, err := db.QueryContext(ctx, extensionsQuery, timeNow())

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		extensionUsageStatistics := types.ExtensionUsageStatistics{}

		if err := rows.Scan(
			&extensionUsageStatistics.ExtensionID,
			&extensionUsageStatistics.UserCount,
			&extensionUsageStatistics.AverageActivations,
		); err != nil {
			return nil, err
		}

		usageStatisticsByExtension = append(usageStatisticsByExtension, &extensionUsageStatistics)
	}
	stats.UsageStatisticsByExtension = usageStatisticsByExtension

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Query for evaluating the success of the extensions platform
	platformQuery := `
	WITH
		non_default_extensions_by_user AS (
			SELECT
					user_id,
					COUNT(DISTINCT argument ->> 'extension_id') AS non_default_extensions
			FROM event_logs
			WHERE name = 'ExtensionActivation'
					AND timestamp > DATE_TRUNC('week', $1::timestamp)
			GROUP BY user_id
		)

	SELECT
		DATE_TRUNC('week', $1::timestamp) AS week_start,
		AVG(non_default_extensions) AS average_non_default_extensions,
		COUNT(user_id)              AS non_default_extension_users
	FROM non_default_extensions_by_user;
	`

	if err := db.QueryRowContext(ctx, platformQuery, timeNow()).Scan(
		&stats.WeekStart,
		&stats.AverageNonDefaultExtensions,
		&stats.NonDefaultExtensionUsers,
	); err != nil {
		return nil, err
	}

	return &stats, nil
}
