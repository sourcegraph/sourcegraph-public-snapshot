package bg

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func DeleteOldEventLogsInPostgres(ctx context.Context, logger log.Logger, db database.DB) error {
	logger = logger.Scoped("deleteOldEventLogs")

	// We choose 93 days as the interval to ensure that we have at least the last three months
	// of logs at all times.
	_, err := db.ExecContext(
		ctx,
		`DELETE FROM event_logs WHERE "timestamp" < now() - interval '93' day`,
	)
	return err
}

func DeleteOldSecurityEventLogsInPostgres(ctx context.Context, logger log.Logger, db database.DB) error {
	logger = logger.Scoped("deleteOldSecurityEventLogs")

	// Only clean up if security event logs are being stored in the database.
	c := conf.Get()
	if c.Log == nil || c.Log.SecurityEventLog == nil {
		return nil
	}
	if c.Log.SecurityEventLog.Location != "database" && c.Log.SecurityEventLog.Location != "all" {
		return nil
	}

	// We choose 30 days as the interval to ensure that we have at least the last month's worth of
	// logs at all times.
	_, err := db.ExecContext(
		ctx,
		`DELETE FROM security_event_logs WHERE "timestamp" < now() - interval '30' day`,
	)
	return err
}
