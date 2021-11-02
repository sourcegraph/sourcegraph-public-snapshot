package webhooks

import (
	"context"
	"database/sql"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

// PurgeHandler is responsible for purging webhook logs that are beyond the
// retention period configured in the site configuration.
type PurgeHandler struct {
	retention time.Duration
	store     database.WebhookLogStore
}

var _ goroutine.Handler = &PurgeHandler{}

func NewPurgeHandler(db *sql.DB, key encryption.Key) *PurgeHandler {
	ph := &PurgeHandler{
		// 72h matches the default value defined in the site configuration
		// schema.
		retention: 72 * time.Hour,
		store:     database.WebhookLogs(db, key),
	}

	conf.Watch(func() {
		logCfg := conf.Get().WebhookLogging
		if logCfg == nil {
			// Nothing to do; the default (or previous value) will be fine here.
			return
		}
		raw := logCfg.Retention

		retention, err := time.ParseDuration(raw)
		if err != nil {
			log15.Warn("invalid webhook log retention period; ignoring", "raw", raw, "err", err, "previous", ph.retention)
			retention = 72 * time.Hour
		}

		ph.retention = retention
	})

	return ph
}

func (ph *PurgeHandler) Handle(ctx context.Context) error {
	log15.Debug("purging webhook logs", "retention", ph.retention)

	if err := ph.store.DeleteStale(ctx, ph.retention); err != nil {
		log15.Error("error deleting stale webhook logs", "err", err)
		return err
	}

	return nil
}
