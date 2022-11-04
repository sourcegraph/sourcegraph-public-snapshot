package users

import (
	"context"
	"math/rand"
	"strings"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/eventlogger"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
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
	if _, err := db.ExecContext(ctx, updateAggregatedUsersStatisticsQuery); err != nil {
		return err
	}
	return nil
}

var started bool

func StartUpdateAggregatedUsersStatisticsTable(ctx context.Context, db database.DB) {
	logger := log.Scoped("aggregated_user_statistics:cache-refresh", "aggregated_user_statistics cache refresh")

	if started {
		panic("already started")
	}

	started = true

	// Wait until table creation migration finishes
	time.Sleep(5 * time.Minute)

	ctx = featureflag.WithFlags(ctx, db.FeatureFlags())

	const delay = 12 * time.Hour
	for {
		if !featureflag.FromContext(ctx).GetBoolOr("user_management_cache_disabled", false) {
			if err := updateAggregatedUsersStatisticsTable(ctx, db); err != nil {
				logger.Error("Error refreshing aggregated_user_statistics cache", log.Error(err))
			}
		}

		// Randomize sleep to prevent thundering herds.
		randomDelay := time.Duration(rand.Intn(600)) * time.Second
		time.Sleep(delay + randomDelay)
	}
}
