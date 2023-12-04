package bg

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func DeleteOldEventLogsInPostgres(ctx context.Context, logger log.Logger, db database.DB) {
	logger = logger.Scoped("deleteOldEventLogs")

	for {
		// We choose 93 days as the interval to ensure that we have at least the last three months
		// of logs at all times.
		_, err := db.ExecContext(
			ctx,
			`DELETE FROM event_logs WHERE "timestamp" < now() - interval '93' day`,
		)
		if err != nil {
			logger.Error("deleting expired rows from event_logs table", log.Error(err))
		}
		time.Sleep(time.Hour)
	}
}

func DeleteOldSecurityEventLogsInPostgres(ctx context.Context, logger log.Logger, db database.DB) {
	logger = logger.Scoped("deleteOldSecurityEventLogs")

	for {
		time.Sleep(time.Hour)

		// Only clean up if security event logs are being stored in the database.
		c := conf.Get()
		if c.Log == nil || c.Log.SecurityEventLog == nil {
			continue
		}
		if c.Log.SecurityEventLog.Location != "database" && c.Log.SecurityEventLog.Location != "all" {
			continue
		}

		// We choose 30 days as the interval to ensure that we have at least the last month's worth of
		// logs at all times.
		_, err := db.ExecContext(
			ctx,
			`DELETE FROM security_event_logs WHERE "timestamp" < now() - interval '30' day`,
		)
		if err != nil {
			logger.Error("deleting expired rows from security_event_logs table", log.Error(err))
		}
	}
}
