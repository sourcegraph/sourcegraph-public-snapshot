package shared

import (
	"context"
	"time"

	gcpmetricexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"

	"github.com/sourcegraph/sourcegraph/internal/tracer/oteldefaults/exporters"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func maybeEnableMetrics(_ context.Context, logger log.Logger, config OpenTelemetryConfig, otelResource *resource.Resource) (func(), error) {
	var reader sdkmetric.Reader
	if config.GCPProjectID != "" {
		logger.Info("initializing GCP trace exporter", log.String("projectID", config.GCPProjectID))
		exporter, err := gcpmetricexporter.New(
			gcpmetricexporter.WithProjectID(config.GCPProjectID))
		if err != nil {
			return nil, errors.Wrap(err, "gcpmetricexporter.New")
		}
		reader = sdkmetric.NewPeriodicReader(exporter,
			sdkmetric.WithInterval(30*time.Second))
	} else {
		logger.Info("initializing Prometheus exporter")
		var err error
		reader, err = exporters.NewPrometheusExporter()
		if err != nil {
			return nil, errors.Wrap(err, "exporters.NewPrometheusExporter")
		}
	}

	// Create and set global tracer
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithResource(otelResource))
	otel.SetMeterProvider(provider)

	logger.Info("metrics configured")
	return func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		start := time.Now()
		logger.Info("Shutting down metrics")
		if err := provider.ForceFlush(shutdownCtx); err != nil {
			logger.Warn("error occurred force-flushing metrics", log.Error(err))
		}
		if err := provider.Shutdown(shutdownCtx); err != nil {
			logger.Warn("error occurred shutting down metrics", log.Error(err))
		}
		logger.Info("metrics shut down", log.Duration("elapsed", time.Since(start)))
	}, nil
}
