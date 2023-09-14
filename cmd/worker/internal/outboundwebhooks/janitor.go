package outboundwebhooks

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

const janitorFrequency = 1 * time.Hour

// makeJanitor creates a background goroutine to expunge old outbound webhook
// jobs and logs from the database.
func makeJanitor(observationCtx *observation.Context, store database.OutboundWebhookJobStore) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		goroutine.HandlerFunc(func(ctx context.Context) error {
			err := store.DeleteBefore(ctx, time.Now().Add(-1*calculateRetention(observationCtx.Logger, conf.Get())))
			if err != nil {
				observationCtx.Logger.Error("outbound webhook janitor error", log.Error(err))
			}
			return err
		}),
		goroutine.WithName("outbound-webhooks.janitor"),
		goroutine.WithDescription("cleans up stale outbound webhook jobs"),
		goroutine.WithInterval(janitorFrequency),
	)
}

// This matches the documented value in the site configuration schema.
const defaultRetention = 72 * time.Hour

func calculateRetention(logger log.Logger, c *conf.Unified) time.Duration {
	if cfg := c.WebhookLogging; cfg != nil {
		retention, err := time.ParseDuration(cfg.Retention)
		if err != nil {
			logger.Warn("invalid webhook log retention period; ignoring", log.String("raw", cfg.Retention), log.Error(err))
		} else {
			return retention
		}
	}

	return defaultRetention
}
