package bg

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

func DeleteOldEventLogsInPostgres(ctx context.Context, db dbutil.DB) {
	for {
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

func DeleteOldSecurityEventLogsInPostgres(ctx context.Context, db dbutil.DB) {
	for {
		_, err := db.ExecContext(
			ctx,
			`DELETE FROM security_event_logs WHERE "timestamp" < now() - interval '186' day`,
		)
		if err != nil {
			log15.Error("deleting expired rows from security_event_logs table", "error", err)
		}
		time.Sleep(time.Hour)
	}
}
