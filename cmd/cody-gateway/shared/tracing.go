package shared

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	gcptraceexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
	"github.com/sourcegraph/sourcegraph/internal/tracer/oteldefaults"
	"github.com/sourcegraph/sourcegraph/internal/tracer/oteldefaults/exporters"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// maybeEnableTracing configures OpenTelemetry tracing if the GOOGLE_CLOUD_PROJECT is set.
// It differs from Sourcegraph's default tracing because we need to export directly to GCP,
// and the use case is more niche as a standalone service.
//
// Based on https://cloud.google.com/trace/docs/setup/go-ot
func maybeEnableTracing(ctx context.Context, logger log.Logger, config OpenTelemetryConfig, otelResource *resource.Resource) (func(), error) {
	// Set globals
	policy.SetTracePolicy(config.TracePolicy)
	otel.SetTextMapPropagator(oteldefaults.Propagator())
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		logger.Debug("OpenTelemetry error", log.Error(err))
	}))

	// Initialize exporter
	var exporter sdktrace.SpanExporter
	if config.GCPProjectID != "" {
		logger.Info("initializing GCP trace exporter", log.String("projectID", config.GCPProjectID))
		var err error
		exporter, err = gcptraceexporter.New(
			gcptraceexporter.WithProjectID(config.GCPProjectID),
			gcptraceexporter.WithErrorHandler(otel.ErrorHandlerFunc(func(err error) {
				logger.Warn("gcptraceexporter error", log.Error(err))
			})),
		)
		if err != nil {
			return nil, errors.Wrap(err, "gcptraceexporter.New")
		}
	} else {
		logger.Info("initializing OTLP exporter")
		var err error
		exporter, err = exporters.NewOTLPTraceExporter(ctx, logger)
		if err != nil {
			return nil, errors.Wrap(err, "exporters.NewOTLPExporter")
		}
	}

	// Create and set global tracer
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(otelResource))
	otel.SetTracerProvider(provider)

	logger.Info("tracing configured")
	return func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		start := time.Now()
		logger.Info("Shutting down tracing")
		if err := provider.ForceFlush(shutdownCtx); err != nil {
			logger.Warn("error occurred force-flushing traces", log.Error(err))
		}
		if err := provider.Shutdown(shutdownCtx); err != nil {
			logger.Warn("error occured shutting down tracing", log.Error(err))
		}
		logger.Info("Tracing shut down", log.Duration("elapsed", time.Since(start)))
	}, nil
}
