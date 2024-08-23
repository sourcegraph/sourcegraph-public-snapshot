package opentelemetry

import (
	"context"
	"time"

	gcptraceexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	jaegerpropagator "go.opentelemetry.io/contrib/propagators/jaeger"
	otpropagator "go.opentelemetry.io/contrib/propagators/ot"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/embedded"
	apioption "google.golang.org/api/option"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// configureTracing configures OpenTelemetry tracing if the GOOGLE_CLOUD_PROJECT is set.
// It differs from Sourcegraph's default tracing because we need to export directly to GCP,
// and the use case is more niche as a standalone service.
//
// Based on https://cloud.google.com/trace/docs/setup/go-ot
func configureTracing(ctx context.Context, logger log.Logger, config Config, res *resource.Resource) (func(), error) {
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
			gcptraceexporter.WithTraceClientOptions([]apioption.ClientOption{
				apioption.WithTelemetryDisabled(),
			}),
		)
		if err != nil {
			return nil, errors.Wrap(err, "gcptraceexporter.New")
		}
	} else {
		logger.Info("initializing OTLP exporter")
		var err error
		exporter, err = otlptrace.New(ctx, otlptracegrpc.NewClient()) // no opts, use OTLP convention
		if err != nil {
			return nil, errors.Wrap(err, "otlptrace.New")
		}
	}

	// Initialize instrumentation
	exportedSpans, err := meter.Int64Counter("otel.trace.exported_span.count",
		metric.WithDescription("Total number of spans exported, and whether the export succeeded or failed."))
	if err != nil {
		return nil, errors.Wrap(err, "initialize exportedSpans metric")
	}
	createdSpans, err := meter.Int64Counter("otel.trace.created_span.count",
		metric.WithDescription("Total number of spans created, and some metadata about them."))
	if err != nil {
		return nil, errors.Wrap(err, "initialize createdSpans metric")
	}

	// Create and set global tracer, wrapping exporter and tracer providers with
	// instrumentation.
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(&instrumentedSpanExporter{
			SpanExporter:  exporter,
			exportedSpans: exportedSpans,
		}),
		sdktrace.WithResource(res))
	otel.SetTracerProvider(&instrumentedTracerProvider{
		wrapped:      provider,
		createdSpans: createdSpans,
	})

	// Return a cleanup callback for everything
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
		logger.Info("tracing shut down",
			log.Duration("elapsed", time.Since(start)))
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

// instrumentedTracerProvider provides trace.Tracers wrapped in instrumentation
// (instrumentedTracer) because the SDK does not provide metrics out of the box.
type instrumentedTracerProvider struct {
	wrapped trace.TracerProvider
	embedded.TracerProvider

	createdSpans metric.Int64Counter
}

var _ trace.TracerProvider = &instrumentedTracerProvider{}

func (s *instrumentedTracerProvider) Tracer(instrumentationName string, opts ...trace.TracerOption) trace.Tracer {
	return &instrumentedTracer{
		wrapped:      s.wrapped.Tracer(instrumentationName, opts...),
		createdSpans: s.createdSpans,
	}
}

// instrumentedTracer wraps trace.Tracer in instrumentation because the SDK does
// not provide metrics out of the box.
type instrumentedTracer struct {
	wrapped trace.Tracer
	embedded.Tracer

	createdSpans metric.Int64Counter
}

var _ trace.Tracer = &instrumentedTracer{}

func (t *instrumentedTracer) Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	ctx, span := t.wrapped.Start(ctx, spanName, opts...)

	t.createdSpans.Add(ctx, 1, metric.WithAttributeSet(attribute.NewSet(
		attribute.Bool("recording", span.IsRecording()),
		attribute.Bool("sampled", span.SpanContext().IsSampled()),
		attribute.Bool("valid", span.SpanContext().IsValid()),
	)))

	return ctx, span
}

// instrumentedTracer wraps sdktrace.SpanExporter in instrumentation because the
// SDK does not provide metrics out of the box.
type instrumentedSpanExporter struct {
	sdktrace.SpanExporter

	exportedSpans metric.Int64Counter
}

var _ sdktrace.SpanExporter = &instrumentedSpanExporter{}

func (i *instrumentedSpanExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	err := i.SpanExporter.ExportSpans(ctx, spans) // call underlying first

	i.exportedSpans.Add(ctx, int64(len(spans)), metric.WithAttributeSet(attribute.NewSet(
		attribute.Bool("succeeded", err == nil),
	)))

	return err
}
