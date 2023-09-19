package telemetrygatewayexporter

import (
	"context"
	"time"

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
}

var ConfigInst = &config{}

func (c *config) Load() {
	// exportAddress currently has no default value, as the feature is not enabled
	// by default. In a future release, the default will be something like
	// 'https://telemetry-gateway.sourcegraph.com', and eventually, won't be configurable.
	c.ExportAddress = env.Get("TELEMETRY_GATEWAY_EXPORTER_EXPORT_ADDR", "", "Target Telemetry Gateway address")

	c.ExportInterval = env.MustGetDuration("TELEMETRY_GATEWAY_EXPORTER_EXPORT_INTERVAL", 10*time.Minute, "Interval at which to export telemetry")
	c.MaxExportBatchSize = env.MustGetInt("TELEMETRY_GATEWAY_EXPORTER_EXPORT_BATCH_SIZE", 10000, "Maximum number of events to export in each batch")

	c.ExportedEventsRetentionWindow = env.MustGetDuration("TELEMETRY_GATEWAY_EXPORTER_EXPORTED_EVENTS_RETENTION",
		2*24*time.Hour, "Duration to retain already-exported telemetry events before deleting")
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
		ConfigInst.ExportAddress,
	)
	if err != nil {
		return nil, errors.Wrap(err, "initializing export client")
	}

	return []goroutine.BackgroundRoutine{
		newExporterJob(
			observationCtx,
			db.TelemetryEventsExportQueue(),
			exporter,
			*ConfigInst,
		),
		newQueueCleanupJob(db.TelemetryEventsExportQueue(), *ConfigInst),
		newBacklogMetricsJob(db.TelemetryEventsExportQueue()),
	}, nil
}
