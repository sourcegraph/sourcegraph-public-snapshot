package dbworker

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// resetter periodically moves all unlocked records that have been in the processing state
// for a while back to queued.
//
// An unlocked record signifies that it is not actively being processed and records in this
// state for more than a few seconds are very likely to be stuck after the worker processing
// them has crashed.
type resetter struct {
	store   store.Store
	options ResetterOptions
}

var _ goroutine.Handler = &resetter{}

type ResetterOptions struct {
	Name     string
	Interval time.Duration
	Metrics  ResetterMetrics
}

type ResetterMetrics struct {
	RecordResets        prometheus.Counter
	RecordResetFailures prometheus.Counter
	Errors              prometheus.Counter
}

func NewResetter(store store.Store, options ResetterOptions) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), options.Interval, &resetter{
		store:   store,
		options: options,
	})
}

func (r *resetter) Handle(ctx context.Context) error {
	resetIDs, erroredIDs, err := r.store.ResetStalled(ctx)
	if err != nil {
		return err
	}

	for _, id := range resetIDs {
		log15.Debug("Reset stalled record", "name", r.options.Name, "id", id)
	}
	for _, id := range erroredIDs {
		log15.Debug("Reset stalled record", "name", r.options.Name, "id", id)
	}

	r.options.Metrics.RecordResets.Add(float64(len(resetIDs)))
	r.options.Metrics.RecordResetFailures.Add(float64(len(erroredIDs)))
	return nil
}

func (r *resetter) HandleError(err error) {
	r.options.Metrics.Errors.Inc()
	log15.Error("Failed to reset stalled records", "name", r.options.Name, "error", err)
}
