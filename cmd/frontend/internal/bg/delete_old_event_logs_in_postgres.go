package bg

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

func DeleteOldEventLogsInPostgres(ctx context.Context, db database.DB) {
	for {
		// We choose 93 days as the interval to ensure that we have at least the last three months
		// of logs at all times.
		_, err := db.ExecContext(
			ctx,
			`DELETE FROM event_logs WHERE "timestamp" < now() - interval '93' day`,
		)
		if err != nil {
			log15.Error("deleting expired rows from event_logs table", "error", err)
		}
		time.Sleep(time.Hour)
	}
}

func DeleteOldSecurityEventLogsInPostgres(ctx context.Context, db database.DB) {
	for {
		// We choose 30 days as the interval to ensure that we have at least the last month's worth of
		// logs at all times.
		_, err := db.ExecContext(
			ctx,
			`DELETE FROM security_event_logs WHERE "timestamp" < now() - interval '30' day`,
		)
		if err != nil {
			log15.Error("deleting expired rows from security_event_logs table", "error", err)
		}
		time.Sleep(time.Hour)
	}
}
