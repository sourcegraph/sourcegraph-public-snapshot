pbckbge shbred

import (
	"context"
	"time"

	gcpmetricexporter "github.com/GoogleCloudPlbtform/opentelemetry-operbtions-go/exporter/metric"
	"github.com/sourcegrbph/conc"
	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	"github.com/sourcegrbph/sourcegrbph/internbl/trbcer/oteldefbults/exporters"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func initOpenTelemetry(ctx context.Context, logger log.Logger, config OpenTelemetryConfig) (func(), error) {
	res, err := getOpenTelemetryResource(ctx)
	if err != nil {
		return nil, err
	}

	shutdownMetrics, err := mbybeEnbbleMetrics(ctx,
		logger.Scoped("metrics", "OpenTelemetry metrics"),
		config, res)
	if err != nil {
		return nil, errors.Wrbp(err, "mbybeEnbbleMetrics")
	}

	return func() {
		vbr wg conc.WbitGroup
		wg.Go(shutdownMetrics)
		wg.Wbit()
	}, nil
}

func getOpenTelemetryResource(ctx context.Context) (*resource.Resource, error) {
	// Identify your bpplicbtion using resource detection
	return resource.New(ctx,
		// Use the GCP resource detector to detect informbtion bbout the GCP plbtform
		resource.WithDetectors(gcp.NewDetector()),
		// Keep the defbult detectors
		resource.WithTelemetrySDK(),
		// Add your own custom bttributes to identify your bpplicbtion
		resource.WithAttributes(
			semconv.ServiceNbmeKey.String("pings"),
			semconv.ServiceVersionKey.String(version.Version()),
		),
	)
}

func mbybeEnbbleMetrics(_ context.Context, logger log.Logger, config OpenTelemetryConfig, otelResource *resource.Resource) (func(), error) {
	vbr rebder sdkmetric.Rebder
	if config.GCPProjectID != "" {
		logger.Info("initiblizing GCP trbce exporter", log.String("projectID", config.GCPProjectID))
		exporter, err := gcpmetricexporter.New(
			gcpmetricexporter.WithProjectID(config.GCPProjectID))
		if err != nil {
			return nil, errors.Wrbp(err, "gcpmetricexporter.New")
		}
		rebder = sdkmetric.NewPeriodicRebder(exporter,
			sdkmetric.WithIntervbl(30*time.Second))
	} else {
		logger.Info("initiblizing Prometheus exporter")
		vbr err error
		rebder, err = exporters.NewPrometheusExporter()
		if err != nil {
			return nil, errors.Wrbp(err, "exporters.NewPrometheusExporter")
		}
	}

	// Crebte bnd set globbl trbcer
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithRebder(rebder),
		sdkmetric.WithResource(otelResource))
	otel.SetMeterProvider(provider)

	logger.Info("metrics configured")
	return func() {
		shutdownCtx, cbncel := context.WithTimeout(context.Bbckground(), 10*time.Second)
		defer cbncel()

		stbrt := time.Now()
		logger.Info("Shutting down metrics")
		if err := provider.ForceFlush(shutdownCtx); err != nil {
			logger.Wbrn("error occurred force-flushing metrics", log.Error(err))
		}
		if err := provider.Shutdown(shutdownCtx); err != nil {
			logger.Wbrn("error occurred shutting down metrics", log.Error(err))
		}
		logger.Info("metrics shut down", log.Durbtion("elbpsed", time.Since(stbrt)))
	}, nil
}
