package telemetrygatewayexporter

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type queueCleanupJob struct {
	store database.TelemetryEventsExportQueueStore

	retentionWindow time.Duration

	prunedHistogram prometheus.Histogram
}

func newQueueCleanupJob(store database.TelemetryEventsExportQueueStore, cfg config) goroutine.BackgroundRoutine {
	job := &queueCleanupJob{
		store: store,
		prunedHistogram: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: "src",
			Subsystem: "telemetrygatewayexport",
			Name:      "pruned",
			Help:      "Size of exported events pruned from the queue table.",
		}),
	}
	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		job,
		goroutine.WithName("telemetrygatewayexporter.queue_cleanup"),
		goroutine.WithDescription("telemetrygatewayexporter queue cleanup"),
		goroutine.WithInterval(cfg.QueueCleanupInterval),
	)
}

func (j *queueCleanupJob) Handle(ctx context.Context) error {
	count, err := j.store.DeletedExported(ctx, time.Now().Add(-j.retentionWindow))
	if err != nil {
		return errors.Wrap(err, "store.DeletedExported")
	}
	j.prunedHistogram.Observe(float64(count))

	return nil
}
