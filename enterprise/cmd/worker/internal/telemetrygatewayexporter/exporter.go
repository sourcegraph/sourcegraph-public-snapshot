package telemetrygatewayexporter

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/telemetrygateway"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type exporterJob struct {
	logger       log.Logger
	store        database.TelemetryEventsExportQueueStore
	exporter     telemetrygateway.Exporter
	maxBatchSize int

	// batchSizeHistogram records real batch sizes of each export.
	batchSizeHistogram prometheus.Histogram
	// exportedEventsCounter records successfully exported events.
	exportedEventsCounter prometheus.Counter
}

func newExporterJob(
	obctx *observation.Context,
	store database.TelemetryEventsExportQueueStore,
	exporter telemetrygateway.Exporter,
	cfg config,
) goroutine.BackgroundRoutine {
	job := &exporterJob{
		logger:       obctx.Logger,
		store:        store,
		maxBatchSize: cfg.MaxExportBatchSize,
		exporter:     exporter,

		batchSizeHistogram: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: "src",
			Subsystem: "telemetrygatewayexport",
			Name:      "batch_size",
			Help:      "Size of event batches exported from the queue.",
		}),
		exportedEventsCounter: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "src",
			Subsystem: "telemetrygatewayexport",
			Name:      "exported_events",
			Help:      "Number of events exported from the queue.",
		}),
	}
	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		job,
		goroutine.WithName("telemetrygatewayexporter.exporter"),
		goroutine.WithDescription("telemetrygatewayexporter events export job"),
		goroutine.WithInterval(cfg.ExportInterval),
		goroutine.WithOperation(obctx.Operation(observation.Op{
			Name:    "TelemetryGateway.Export",
			Metrics: metrics.NewREDMetrics(prometheus.DefaultRegisterer, "telemetrygatewayexporter_exporter"),
		})),
	)
}

var _ goroutine.Finalizer = (*exporterJob)(nil)

func (j *exporterJob) OnShutdown() { _ = j.exporter.Close() }

func (j *exporterJob) Handle(ctx context.Context) error {
	logger := trace.Logger(ctx, j.logger).
		With(log.Int("maxBatchSize", j.maxBatchSize))

	if conf.Get().LicenseKey == "" {
		logger.Debug("license key not set, skipping export")
		return nil
	}

	// Get events from the queue
	batch, err := j.store.ListForExport(ctx, j.maxBatchSize)
	if err != nil {
		return errors.Wrap(err, "ListForExport")
	}
	j.batchSizeHistogram.Observe(float64(len(batch)))
	if len(batch) == 0 {
		logger.Debug("no events to export")
		return nil
	}

	logger.Info("exporting events", log.Int("count", len(batch)))

	// Send out events
	succeeded, exportErr := j.exporter.ExportEvents(ctx, batch)

	// Mark succeeded events
	j.exportedEventsCounter.Add(float64(len(succeeded)))
	if err := j.store.MarkAsExported(ctx, succeeded); err != nil {
		logger.Error("failed to mark exported events as exported",
			log.Strings("succeeded", succeeded),
			log.Error(err))
	}

	// Report export status
	if exportErr != nil {
		return exportErr
	}

	logger.Info("events exported", log.Int("succeeded", len(succeeded)))
	return nil
}
