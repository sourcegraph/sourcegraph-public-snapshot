package telemetrygatewayexporter

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/telemetrygateway"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type exporterJob struct {
	logger           log.Logger
	exportQueueStore database.TelemetryEventsExportQueueStore
	globalStateStore database.GlobalStateStore

	maxBatchSize int

	// batchSizeHistogram records real batch sizes of each export.
	batchSizeHistogram prometheus.Histogram
	// exportedEventsCounter records successfully exported events.
	exportedEventsCounter prometheus.Counter
}

func newExporterJob(
	obctx *observation.Context,
	db database.DB,
	cfg config,
) goroutine.BackgroundRoutine {
	job := &exporterJob{
		logger: obctx.Logger,

		exportQueueStore: db.TelemetryEventsExportQueue(),
		globalStateStore: db.GlobalState(),

		maxBatchSize: cfg.MaxExportBatchSize,

		batchSizeHistogram: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: "src",
			Subsystem: "telemetrygatewayexporter",
			Name:      "batch_size",
			Help:      "Size of event batches exported from the queue.",
			Buckets:   prometheus.ExponentialBucketsRange(1, float64(cfg.MaxExportBatchSize), 10),
		}),
		exportedEventsCounter: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "src",
			Subsystem: "telemetrygatewayexporter",
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
			Name:    "TelemetryGatewayExporter.Export",
			Metrics: metrics.NewREDMetrics(prometheus.DefaultRegisterer, "telemetrygatewayexporter_exporter"),
		})),
	)
}

func (j *exporterJob) Handle(ctx context.Context) error {
	logger := trace.Logger(ctx, j.logger).
		With(log.Int("maxBatchSize", j.maxBatchSize))

	// Check the current licensing mode.
	if licensing.GetTelemetryEventsExportMode(conf.DefaultClient()) ==
		licensing.TelemetryEventsExportDisabled {
		logger.Info("export is currently disabled entirely via licensing")
		return nil
	}

	// Create a new connection in each handle execution. This helps us recover
	// from any potential Telemetry Gateway downtime, in case a connection fails,
	// as each worker run will create a new one.
	exporter, err := telemetrygateway.NewExporter(
		ctx,
		j.logger.Scoped("exporter"),
		conf.DefaultClient(),
		j.globalStateStore,
		ConfigInst.ExportAddress,
	)
	if err != nil {
		return errors.Wrap(err, "connect to Telemetry Gateway")
	}
	defer exporter.Close()

	// Get events from the queue
	batch, err := j.exportQueueStore.ListForExport(ctx, j.maxBatchSize)
	if err != nil {
		return errors.Wrap(err, "ListForExport")
	}
	j.batchSizeHistogram.Observe(float64(len(batch)))
	if len(batch) == 0 {
		logger.Debug("no events to export")
		return nil
	}

	logger.Debug("exporting events", log.Int("count", len(batch)))

	// Send out events
	succeeded, exportErr := exporter.ExportEvents(ctx, batch)

	// Mark succeeded events
	j.exportedEventsCounter.Add(float64(len(succeeded)))
	if err := j.exportQueueStore.MarkAsExported(ctx, succeeded); err != nil {
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
