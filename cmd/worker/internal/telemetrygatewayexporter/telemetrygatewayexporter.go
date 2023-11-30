package telemetrygatewayexporter

import (
	"context"
	"net/url"
	"time"

	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	// A past assessment of events throughput is used as a rough ballpark for
	// this configuration: roughly 3500 daily events per 10 users.
	//
	// Based on this, for a 5000-user instance, we need to be able to process
	// 1.75M events per day. At a 5-minute interval, we can export
	// 288 * 10000 =~ 2.8M events per day by default, leaving us with plenty of
	// headroom on most instances, with the ability to configure higher batch
	// sizes as needed.
	//
	// Observed 5k events ~ 3MB, 10k in each batch should be safe. The exporter
	// will split a batch into several payloads within an export stream.
	defaultExportInterval  = 5 * time.Minute
	defaultExportBatchSize = 10000
)

type config struct {
	env.BaseConfig

	ExportAddress *url.URL

	ExportInterval     time.Duration
	MaxExportBatchSize int

	ExportedEventsRetentionWindow time.Duration

	QueueCleanupInterval time.Duration
}

var ConfigInst = &config{}

const defaultTelemetryGatewayExporterAddress = "https://telemetry-gateway.sourcegraph.com:443"

func (c *config) Load() {
	// ExportAddress currently remains disable-able for 5.2.1 purely as a
	// last resort escape hatch.
	// TODO(5.2.2): Make ExportAddress a hard requirement, and only allow it to
	// be configurable from the default in dev.
	if exportAddress := env.Get(
		"TELEMETRY_GATEWAY_EXPORTER_EXPORT_ADDR",
		defaultTelemetryGatewayExporterAddress,
		"Target Telemetry Gateway address",
	); exportAddress != "" {
		var err error
		c.ExportAddress, err = url.Parse(exportAddress)
		if err != nil {
			c.AddError(errors.Wrap(err, "TELEMETRY_GATEWAY_EXPORTER_EXPORT_ADDR value is invalid"))
		}
	}

	c.ExportInterval = env.MustGetDuration("TELEMETRY_GATEWAY_EXPORTER_EXPORT_INTERVAL", defaultExportInterval,
		"Interval at which to export telemetry")
	if c.ExportInterval > 1*time.Hour {
		c.AddError(errors.New("TELEMETRY_GATEWAY_EXPORTER_EXPORT_INTERVAL cannot be more than 1 hour"))
	}

	c.MaxExportBatchSize = env.MustGetInt("TELEMETRY_GATEWAY_EXPORTER_EXPORT_BATCH_SIZE", defaultExportBatchSize,
		"Maximum number of events to export in each batch")
	if c.MaxExportBatchSize < 100 {
		c.AddError(errors.New("TELEMETRY_GATEWAY_EXPORTER_EXPORT_BATCH_SIZE must be no less than 100"))
	}

	c.ExportedEventsRetentionWindow = env.MustGetDuration("TELEMETRY_GATEWAY_EXPORTER_EXPORTED_EVENTS_RETENTION",
		24*time.Hour, "Duration to retain already-exported telemetry events before deleting")

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
	if ConfigInst.ExportAddress == nil {
		observationCtx.Logger.Warn("Telemetry Gateway export has been manually disabled. This capability will be removed in Sourcegraph 5.2.2")
		return nil, nil
	}

	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	return []goroutine.BackgroundRoutine{
		newExporterJob(
			observationCtx,
			db,
			*ConfigInst,
		),
		newQueueCleanupJob(observationCtx, db.TelemetryEventsExportQueue(), *ConfigInst),
		newQueueMetricsJob(observationCtx, db.TelemetryEventsExportQueue()),
	}, nil
}
