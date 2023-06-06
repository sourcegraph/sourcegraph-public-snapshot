package webhooks

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// janitor is a worker responsible for expunging stale webhook logs from the
// database.
type janitor struct{}

var _ job.Job = &janitor{}

func NewJanitor() job.Job {
	return &janitor{}
}

func (j *janitor) Description() string {
	return ""
}

func (j *janitor) Config() []env.Config {
	return nil
}

func (j *janitor) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	return []goroutine.BackgroundRoutine{
		// The site configuration schema notes that retention values under an
		// hour aren't supported, and this is why: there's no point running this
		// operation more frequently than that, given it's purely a debugging
		// tool.
		goroutine.NewPeriodicGoroutine(
			context.Background(),
			&handler{
				store: db.WebhookLogs(keyring.Default().WebhookLogKey),
			},
			goroutine.WithName("batchchanges.webhook-log-janitor"),
			goroutine.WithDescription("cleans up stale webhook logs"),
			goroutine.WithInterval(1*time.Hour),
		),
	}, nil
}
