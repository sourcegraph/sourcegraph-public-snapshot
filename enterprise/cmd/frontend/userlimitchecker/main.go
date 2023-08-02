package userlimitchecker

import (
	"context"
	"time"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func init() {
	logger := log.Scoped(LC_LOGGER_SCOPE, LC_LOGGER_DESC)
	ctx := context.Background()

	observationCtx := observation.NewContext(logger)
	conn, err := connections.EnsureNewFrontendDB(observationCtx, "", APP_NAME)
	if err != nil {
		errors.Wrap(err, CONN_FAILED)
	}

	db := database.NewDB(logger, conn)
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		for range ticker.C {
			if err := SendApproachingUserLimitAlert(ctx, db); err != nil {
				errors.Wrap(err, LIMIT_CHECK_FAILED)
			}
		}
	}()
}
