package telemetrygatewayexporter

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type queueCleanupJob struct {
	store database.TelemetryEventsExportQueueStore

	retentionWindow time.Duration

	prunedHistogram prometheus.Histogram
}

func newQueueCleanupJob(obctx *observation.Context, store database.TelemetryEventsExportQueueStore, cfg config) goroutine.BackgroundRoutine {
	job := &queueCleanupJob{
		store: store,
		prunedHistogram: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: "src",
			Subsystem: "telemetrygatewayexporter",
			Name:      "events_pruned",
			Help:      "Size of exported events pruned from the queue table.",
			Buckets: prometheus.ExponentialBucketsRange(
				1,
				// Max bucket at range if each export in a cleanup interval
				// exports the maximum number of events.
				(cfg.ExportInterval.Seconds()/cfg.QueueCleanupInterval.Seconds())*float64(cfg.MaxExportBatchSize),
				10),
		}),
	}
	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		job,
		goroutine.WithName("telemetrygatewayexporter.queue_cleanup"),
		goroutine.WithDescription("telemetrygatewayexporter queue cleanup"),
		goroutine.WithInterval(cfg.QueueCleanupInterval),
		goroutine.WithOperation(obctx.Operation(observation.Op{
			Name:    "TelemetryGatewayExporter.QueueCleanup",
			Metrics: metrics.NewREDMetrics(prometheus.DefaultRegisterer, "telemetrygatewayexporter_queue_cleanup"),
		})),
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
