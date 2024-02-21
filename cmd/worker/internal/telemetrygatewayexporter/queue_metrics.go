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

type queueMetricsJob struct {
	store database.TelemetryEventsExportQueueDiagnosticsStore

	sizeGauge prometheus.Gauge
	// Age of the oldest unexported event in seconds
	oldestUnexportedEventGauge prometheus.Gauge
}

func newQueueMetricsJob(obctx *observation.Context, store database.TelemetryEventsExportQueueStore) goroutine.BackgroundRoutine {
	job := &queueMetricsJob{
		store: store,
		sizeGauge: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: "src",
			Subsystem: "telemetrygatewayexporter",
			Name:      "queue_size",
			Help:      "Current number of events waiting to be exported.",
		}),
		oldestUnexportedEventGauge: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: "src",
			Subsystem: "telemetrygatewayexporter",
			Name:      "oldest_unexported_event",
			Help:      "Age of the oldest unexported event in seconds.",
		}),
	}
	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		job,
		goroutine.WithName("telemetrygatewayexporter.queue_metrics_reporter"),
		goroutine.WithDescription("telemetrygatewayexporter backlog metrics reporting"),
		goroutine.WithInterval(time.Minute*5),
		goroutine.WithOperation(obctx.Operation(observation.Op{
			Name:    "TelemetryGatewayExporter.ReportQueueMetrics",
			Metrics: metrics.NewREDMetrics(prometheus.DefaultRegisterer, "telemetrygatewayexporter_queue_metrics_reporter"),
		})),
	)
}

func (j *queueMetricsJob) Handle(ctx context.Context) error {
	count, oldest, err := j.store.CountUnexported(ctx)
	if err != nil {
		return errors.Wrap(err, "store.CountUnexported")
	}
	j.sizeGauge.Set(float64(count))

	if oldest.IsZero() {
		j.oldestUnexportedEventGauge.Set(0)
	} else {
		oldestAge := time.Since(oldest)
		j.oldestUnexportedEventGauge.Set(float64(oldestAge.Seconds()))
	}

	return nil
}
