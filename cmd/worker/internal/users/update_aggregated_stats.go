package users

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/eventlogger"
)

var (
	updateAggregatedUsersStatisticsQuery = `
	INSERT INTO aggregated_user_statistics (user_id, user_last_active_at, user_events_count)
	SELECT
		user_id,
		last_active_at,
		events_count
	FROM
		(
			SELECT
				user_id,
				MAX(timestamp) AS last_active_at,
				-- count billable only events for each user
				COUNT(*) FILTER (WHERE name NOT IN ('` + strings.Join(eventlogger.NonActiveUserEvents, "','") + `')) AS events_count
			FROM
				event_logs
			GROUP BY
				user_id
		) AS events
		INNER JOIN users ON users.id = events.user_id
	ON CONFLICT (user_id) DO UPDATE
		SET
			user_last_active_at = EXCLUDED.user_last_active_at,
			user_events_count = EXCLUDED.user_events_count,
			updated_at = NOW();
	`
)

func updateAggregatedUsersStatisticsTable(ctx context.Context, db database.DB) error {
	_, err := db.ExecContext(ctx, updateAggregatedUsersStatisticsQuery)
	return err
}
