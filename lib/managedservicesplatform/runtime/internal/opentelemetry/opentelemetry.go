package opentelemetry

import (
	"context"

	"github.com/sourcegraph/conc"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	gcpdetector "go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

type Config struct {
	GCPProjectID string
}

// Init initializes OpenTelemetry integrations. If config.GCPProjectID is set,
// all OpenTelemetry integrations will point to a GCP exporter - otherwise, a
// local dev default is chosen:
//
//   - traces: OTLP exporter
//   - metrics: Prometheus exporter
func Init(ctx context.Context, logger log.Logger, config Config, r log.Resource) (func(), error) {
	res, err := getOpenTelemetryResource(ctx, r)
	if err != nil {
		return nil, errors.Wrap(err, "init resource")
	}

	shutdownTracing, err := maybeEnableTracing(ctx, logger, config, res)
	if err != nil {
		return nil, errors.Wrap(err, "enable tracing")
	}

	shutdownMetrics, err := maybeEnableMetrics(ctx, logger, config, res)
	if err != nil {
		return nil, errors.Wrap(err, "enable metrics")
	}

	return func() {
		logger.Debug("shutting down OpenTelemetry")
		var wg conc.WaitGroup
		wg.Go(shutdownTracing)
		wg.Go(shutdownMetrics)
		wg.Wait()
	}, nil
}

func getOpenTelemetryResource(ctx context.Context, r log.Resource) (*resource.Resource, error) {
	// Identify your application using resource detection
	return resource.New(ctx,
		// Use the GCP resource detector to detect information about the GCP platform
		resource.WithDetectors(gcpdetector.NewDetector()),
		// Use the default detectors
		resource.WithTelemetrySDK(),
		// Add our own attributes
		resource.WithAttributes(
			semconv.ServiceNameKey.String(r.Name),
			semconv.ServiceVersionKey.String(r.Version),
			semconv.ServiceInstanceIDKey.String(r.InstanceID),
			semconv.ServiceNamespaceKey.String(r.Namespace),
		),
	)
}
