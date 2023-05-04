package shared

import (
	"context"
	"os"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	gcptraceexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

// maybeEnableTracing configures OpenTelemetry tracing if the GOOGLE_CLOUD_PROJECT is set.
// It differs from Sourcegraph's default tracing because we need to export directly to GCP,
// and the use case is more niche as a standalone service.
//
// Based on https://cloud.google.com/trace/docs/setup/go-ot
func maybeEnableTracing(ctx context.Context, logger log.Logger) (func(), error) {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		logger.Debug("GOOGLE_CLOUD_PROJECT not set, not enabling tracing")
		return func() {}, nil
	}

	exporter, err := gcptraceexporter.New(gcptraceexporter.WithProjectID(projectID))
	if err != nil {
		return nil, errors.Wrap(err, "texporter.New")
	}

	// Identify your application using resource detection
	res, err := resource.New(ctx,
		// Use the GCP resource detector to detect information about the GCP platform
		resource.WithDetectors(gcp.NewDetector()),
		// Keep the default detectors
		resource.WithTelemetrySDK(),
		// Add your own custom attributes to identify your application
		resource.WithAttributes(
			semconv.ServiceNameKey.String("llm-proxy"),
			semconv.ServiceVersionKey.String(version.Version()),
		),
	)
	if err != nil {
		return nil, errors.Wrap(err, "resource.New")
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	logger.Info("tracing configured")
	return func() { tp.ForceFlush(ctx) }, nil
}
