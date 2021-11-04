package webhooks

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type handler struct {
	store database.WebhookLogStore
}

var _ goroutine.Handler = &handler{}

func (h *handler) Handle(ctx context.Context) error {
	retention := calculateRetention(conf.Get())
	log15.Debug("purging webhook logs", "retention", retention)

	if err := h.store.DeleteStale(ctx, retention); err != nil {
		log15.Error("error deleting stale webhook logs", "err", err)
		return err
	}

	return nil
}

// This matches the documented value in the site configuration schema.
const defaultRetention = 72 * time.Hour

func calculateRetention(c *conf.Unified) time.Duration {
	if cfg := c.WebhookLogging; cfg != nil {
		retention, err := time.ParseDuration(cfg.Retention)
		if err != nil {
			log15.Warn("invalid webhook log retention period; ignoring", "raw", cfg.Retention, "err", err)
		} else {
			return retention
		}
	}

	return defaultRetention
}
