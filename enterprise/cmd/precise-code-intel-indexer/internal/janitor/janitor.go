package janitor

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
)

type Janitor struct {
	store           store.Store
	janitorInterval time.Duration
	metrics         JanitorMetrics
	ctx             context.Context
	cancel          func()
	finished        chan (struct{})
}

func New(
	store store.Store,
	janitorInterval time.Duration,
	metrics JanitorMetrics,
) *Janitor {
	ctx, cancel := context.WithCancel(context.Background())

	return &Janitor{
		store:           store,
		janitorInterval: janitorInterval,
		metrics:         metrics,
		ctx:             ctx,
		cancel:          cancel,
		finished:        make(chan struct{}),
	}
}

// Run periodically performs a best-effort cleanup process.
func (j *Janitor) Run() {
	defer close(j.finished)

	for {
		if err := j.run(); err != nil {
			j.metrics.Errors.Inc()
			log15.Error("Failed to run janitor process", "err", err)
		}

		select {
		case <-time.After(j.janitorInterval):
		case <-j.ctx.Done():
			return
		}
	}
}

func (j *Janitor) Stop() {
	j.cancel()
	<-j.finished
}

func (j *Janitor) run() error {
	// TODO(efritz) - use cancellable context for API calls

	if err := j.removeRecordsForDeletedRepositories(); err != nil {
		return errors.Wrap(err, "janitor.removeRecordsForDeletedRepositories")
	}

	return nil
}
