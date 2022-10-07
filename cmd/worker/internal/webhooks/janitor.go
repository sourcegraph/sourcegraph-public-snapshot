package webhooks

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// janitor is a worker responsible for expunging stale webhook logs from the
// database.
type janitor struct {
	observationContext *observation.Context
}

var _ job.Job = &janitor{}

func NewJanitor(observationContext *observation.Context) job.Job {
	return &janitor{observationContext: &observation.Context{
		Logger:       log.NoOp(),
		Tracer:       observationContext.Tracer,
		Registerer:   observationContext.Registerer,
		HoneyDataset: observationContext.HoneyDataset,
	}}
}

func (j *janitor) Description() string {
	return ""
}

func (j *janitor) Config() []env.Config {
	return nil
}

func (j *janitor) Routines(startupCtx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.InitDBWithLogger(logger, j.observationContext)
	if err != nil {
		return nil, err
	}

	return []goroutine.BackgroundRoutine{
		// The site configuration schema notes that retention values under an
		// hour aren't supported, and this is why: there's no point running this
		// operation more frequently than that, given it's purely a debugging
		// tool.
		goroutine.NewPeriodicGoroutine(context.Background(), 1*time.Hour, &handler{
			store: db.WebhookLogs(keyring.Default().WebhookLogKey),
		}),
	}, nil
}
