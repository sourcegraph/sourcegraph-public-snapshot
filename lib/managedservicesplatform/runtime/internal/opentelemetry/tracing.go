package opentelemetry

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	gcptraceexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	jaegerpropagator "go.opentelemetry.io/contrib/propagators/jaeger"
	otpropagator "go.opentelemetry.io/contrib/propagators/ot"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// maybeEnableTracing configures OpenTelemetry tracing if the GOOGLE_CLOUD_PROJECT is set.
// It differs from Sourcegraph's default tracing because we need to export directly to GCP,
// and the use case is more niche as a standalone service.
//
// Based on https://cloud.google.com/trace/docs/setup/go-ot
func maybeEnableTracing(ctx context.Context, logger log.Logger, config Config, res *resource.Resource) (func(), error) {
	// Set globals
	otel.SetTextMapPropagator(defaultPropagator())
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		logger.Debug("OpenTelemetry error", log.Error(err))
	}))

	// Initialize exporter
	var exporter sdktrace.SpanExporter
	if config.GCPProjectID != "" {
		logger.Debug("initializing GCP trace exporter", log.String("projectID", config.GCPProjectID))
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
		logger.Debug("initializing OTLP exporter")
		var err error
		exporter, err = otlptrace.New(ctx, otlptracegrpc.NewClient()) // no opts, use OTLP convention
		if err != nil {
			return nil, errors.Wrap(err, "otlptrace.New")
		}
	}

	// Create and set global tracer
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res))
	otel.SetTracerProvider(provider)

	logger.Info("tracing configured")
	return func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		start := time.Now()
		logger.Debug("shutting down tracing")
		if err := provider.ForceFlush(shutdownCtx); err != nil {
			logger.Warn("error occurred force-flushing traces", log.Error(err))
		}
		if err := provider.Shutdown(shutdownCtx); err != nil {
			logger.Warn("error occured shutting down tracing", log.Error(err))
		}
		logger.Info("tracing shut down", log.Duration("elapsed", time.Since(start)))
	}, nil
}

// defaultPropagator returns a propagator that supports a bunch of common formats like
// W3C Trace Context, W3C Baggage, OpenTracing, and Jaeger (the latter two being
// the more commonly used legacy formats at Sourcegraph). This helps ensure
// propagation between services continues to work.
func defaultPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		jaegerpropagator.Jaeger{},
		otpropagator.OT{},
		// W3C Trace Context format (https://www.w3.org/TR/trace-context/)
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}
