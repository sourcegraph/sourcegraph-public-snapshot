package tracer

import (
	"context"
	"strconv"

	"go.opentelemetry.io/otel/sdk/resource"
	oteltracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
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

	// Wrap the exporter with instrumentation
	exporter = instrumentedExporter{exporter}

	// Always use batch span processor - to get more immediate exports in e.g.
	// local dev, toggle the OTEL_BSP_* configurations instead:
	// https://sourcegraph.com/github.com/open-telemetry/opentelemetry-go@1d1ecbc5f936208a91521ede9d0b2f557170425e/-/blob/sdk/internal/env/env.go?L26-37
	return oteltracesdk.NewBatchSpanProcessor(exporter), nil
}

var metricExportedSpans = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: "otelsdk",
	Subsystem: "trace_exporter",
	Name:      "exported_spans",
}, []string{"succeeded"})

type instrumentedExporter struct{ oteltracesdk.SpanExporter }

func (i instrumentedExporter) ExportSpans(ctx context.Context, spans []oteltracesdk.ReadOnlySpan) error {
	err := i.SpanExporter.ExportSpans(ctx, spans)

	// Wrap the export with instrumentation, as the SDK does not provide any out
	// of the box.
	metricExportedSpans.
		With(prometheus.Labels{
			"succeeded": strconv.FormatBool(err == nil),
		}).
		Add(float64(len(spans)))

	return err
}
