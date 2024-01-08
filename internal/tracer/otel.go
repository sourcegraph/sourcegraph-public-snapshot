package tracer

import (
	"context"

	"go.opentelemetry.io/otel/sdk/resource"
	oteltracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/tracer/oteldefaults/exporters"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// newOtelTracerProvider creates a baseline OpenTelemetry TracerProvider that doesn't do
// anything with incoming spans.
func newOtelTracerProvider(r log.Resource) *oteltracesdk.TracerProvider {
	return oteltracesdk.NewTracerProvider(
		// Adapt log.Resource to OpenTelemetry's internal resource type
		oteltracesdk.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(r.Name),
				semconv.ServiceNamespaceKey.String(r.Namespace),
				semconv.ServiceInstanceIDKey.String(r.InstanceID),
				semconv.ServiceVersionKey.String(r.Version),
			),
		),
		// We do not have OpenTracing bridging enabled, so we can use a sampler
		// that configures traces for export based on trace policy in context
		// while leaving valid traces in context.
		oteltracesdk.WithSampler(tracePolicySampler{}),
	)
}

// newOtelSpanProcessor is the default builder for OpenTelemetry span processors to
// register on the underlying OpenTelemetry TracerProvider.
func newOtelSpanProcessor(logger log.Logger, opts options, debug bool) (oteltracesdk.SpanProcessor, error) {
	var exporter oteltracesdk.SpanExporter
	var err error
	switch opts.TracerType {
	case OpenTelemetry:
		exporter, err = exporters.NewOTLPTraceExporter(context.Background(), logger)

	case Jaeger:
		exporter, err = exporters.NewJaegerExporter()

	case None:
		exporter = tracetest.NewNoopExporter()

	default:
		err = errors.Newf("unknown tracer type %q", opts.TracerType)
	}
	if err != nil {
		return nil, err
	}

	// If in debug mode, we use a synchronous span processor to force spans to get pushed
	// immediately, otherwise we batch
	if debug {
		logger.Warn("using synchronous span processor - disable 'observability.debug' to use something more suitable for production")
		return oteltracesdk.NewSimpleSpanProcessor(exporter), nil
	}
	return oteltracesdk.NewBatchSpanProcessor(exporter), nil
}
