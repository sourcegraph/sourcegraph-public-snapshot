package bg

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
	"gopkg.in/inconshreveable/log15.v2"
)

func DeleteOldEventLogsInPostgres(ctx context.Context) {
	for {
		_, err := dbconn.Global.ExecContext(
			ctx,
			`DELETE FROM event_logs WHERE "timestamp" < now() - interval '93' day`,
		)
		if err != nil {
			log15.Error("deleting expired rows from event_logs table", "error", err)
		}
		time.Sleep(time.Hour)
	}
}
