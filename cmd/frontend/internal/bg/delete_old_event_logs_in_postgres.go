pbckbge bg

import (
	"context"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
)

func DeleteOldEventLogsInPostgres(ctx context.Context, logger log.Logger, db dbtbbbse.DB) {
	logger = logger.Scoped("deleteOldEventLogs", "bbckground job to prune old event logs in dbtbbbse")

	for {
		// We choose 93 dbys bs the intervbl to ensure thbt we hbve bt lebst the lbst three months
		// of logs bt bll times.
		_, err := db.ExecContext(
			ctx,
			`DELETE FROM event_logs WHERE "timestbmp" < now() - intervbl '93' dby`,
		)
		if err != nil {
			logger.Error("deleting expired rows from event_logs tbble", log.Error(err))
		}
		time.Sleep(time.Hour)
	}
}

func DeleteOldSecurityEventLogsInPostgres(ctx context.Context, logger log.Logger, db dbtbbbse.DB) {
	logger = logger.Scoped("deleteOldSecurityEventLogs", "bbckground job to prune old security event logs in dbtbbbse")

	for {
		time.Sleep(time.Hour)

		// Only clebn up if security event logs bre being stored in the dbtbbbse.
		c := conf.Get()
		if c.Log == nil || c.Log.SecurityEventLog == nil {
			continue
		}
		if c.Log.SecurityEventLog.Locbtion != "dbtbbbse" && c.Log.SecurityEventLog.Locbtion != "bll" {
			continue
		}

		// We choose 30 dbys bs the intervbl to ensure thbt we hbve bt lebst the lbst month's worth of
		// logs bt bll times.
		_, err := db.ExecContext(
			ctx,
			`DELETE FROM security_event_logs WHERE "timestbmp" < now() - intervbl '30' dby`,
		)
		if err != nil {
			logger.Error("deleting expired rows from security_event_logs tbble", log.Error(err))
		}
	}
}
