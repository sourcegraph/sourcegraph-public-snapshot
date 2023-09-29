package telemetrygatewayexporter

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/telemetrygateway"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type config struct {
	env.BaseConfig

	ExportAddress string

	ExportInterval     time.Duration
	MaxExportBatchSize int

	ExportedEventsRetentionWindow time.Duration

	QueueCleanupInterval time.Duration
}

var ConfigInst = &config{}

func (c *config) Load() {
	// exportAddress currently has no default value, as the feature is not enabled
	// by default. In a future release, the default will be something like
	// 'https://telemetry-gateway.sourcegraph.com', and eventually, won't be configurable.
	c.ExportAddress = env.Get("TELEMETRY_GATEWAY_EXPORTER_EXPORT_ADDR", "", "Target Telemetry Gateway address")

	c.ExportInterval = env.MustGetDuration("TELEMETRY_GATEWAY_EXPORTER_EXPORT_INTERVAL", 10*time.Minute,
		"Interval at which to export telemetry")
	if c.ExportInterval > 1*time.Hour {
		c.AddError(errors.New("TELEMETRY_GATEWAY_EXPORTER_EXPORT_INTERVAL cannot be more than 1 hour"))
	}

	c.MaxExportBatchSize = env.MustGetInt("TELEMETRY_GATEWAY_EXPORTER_EXPORT_BATCH_SIZE", 5000,
		"Maximum number of events to export in each batch")
	if c.MaxExportBatchSize < 100 {
		c.AddError(errors.New("TELEMETRY_GATEWAY_EXPORTER_EXPORT_BATCH_SIZE must be no less than 100"))
	}

	c.ExportedEventsRetentionWindow = env.MustGetDuration("TELEMETRY_GATEWAY_EXPORTER_EXPORTED_EVENTS_RETENTION",
		2*24*time.Hour, "Duration to retain already-exported telemetry events before deleting")

	c.QueueCleanupInterval = env.MustGetDuration("TELEMETRY_GATEWAY_EXPORTER_QUEUE_CLEANUP_INTERVAL",
		30*time.Minute, "Interval at which to clean up telemetry export queue")
}

type telemetryGatewayExporter struct{}

func NewJob() *telemetryGatewayExporter {
	return &telemetryGatewayExporter{}
}

func (t *telemetryGatewayExporter) Description() string {
	return "A background routine that exports telemetry events to Sourcegraph's Telemetry Gateway"
}

func (t *telemetryGatewayExporter) Config() []env.Config {
	return []env.Config{ConfigInst}
}

func (t *telemetryGatewayExporter) Routines(initCtx context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	if ConfigInst.ExportAddress == "" {
		return nil, nil
	}

	observationCtx.Logger.Info("Telemetry Gateway export enabled - initializing background routines")

	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	exporter, err := telemetrygateway.NewExporter(
		initCtx,
		observationCtx.Logger.Scoped("exporter", "exporter client"),
		conf.DefaultClient(),
		db.GlobalState(),
		ConfigInst.ExportAddress,
	)
	if err != nil {
		return nil, errors.Wrap(err, "initializing export client")
	}

	observationCtx.Logger.Info("connected to Telemetry Gateway",
		log.String("address", ConfigInst.ExportAddress))

	return []goroutine.BackgroundRoutine{
		newExporterJob(
			observationCtx,
			db.TelemetryEventsExportQueue(),
			exporter,
			*ConfigInst,
		),
		newQueueCleanupJob(observationCtx, db.TelemetryEventsExportQueue(), *ConfigInst),
		newQueueMetricsJob(observationCtx, db.TelemetryEventsExportQueue()),
	}, nil
}
