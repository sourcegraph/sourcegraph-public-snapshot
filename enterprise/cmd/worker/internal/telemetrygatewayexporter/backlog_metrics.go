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

type backlogMetricsJob struct {
	store database.TelemetryEventsExportQueueStore

	sizeGauge prometheus.Gauge
}

func newBacklogMetricsJob(store database.TelemetryEventsExportQueueStore) goroutine.BackgroundRoutine {
	job := &backlogMetricsJob{
		store: store,
		sizeGauge: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: "src",
			Subsystem: "telemetrygatewayexport",
			Name:      "backlog_size",
			Help:      "Current number of events waiting to be exported.",
		}),
	}
	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		job,
		goroutine.WithName("telemetrygatewayexporter.backlog_metrics"),
		goroutine.WithDescription("telemetrygatewayexporter backlog metrics"),
		goroutine.WithInterval(time.Minute*5),
	)
}

func (j *backlogMetricsJob) Handle(ctx context.Context) error {
	count, err := j.store.CountUnexported(ctx)
	if err != nil {
		return errors.Wrap(err, "store.CountUnexported")
	}
	j.sizeGauge.Set(float64(count))

	return nil
}
