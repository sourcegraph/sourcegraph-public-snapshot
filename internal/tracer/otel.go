package tracer

import (
	"context"
	"io"
	"regexp"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/sourcegraph/log"
	jaegerpropagator "go.opentelemetry.io/contrib/propagators/jaeger"
	otpropagator "go.opentelemetry.io/contrib/propagators/ot"
	"go.opentelemetry.io/otel"
	otelbridge "go.opentelemetry.io/otel/bridge/opentracing"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	w3cpropagator "go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	oteltracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/sourcegraph/sourcegraph/internal/otlpenv"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// newOTelBridgeTracer creates an opentracing.Tracer that exports all OpenTracing traces
// as OpenTelemetry traces to an OpenTelemetry collector (effectively "bridging" the two
// APIs). This enables us to continue leveraging the OpenTracing API (which is a predecessor
// to OpenTelemetry tracing) without making changes to existing tracing code.
//
// All configuration is sourced directly from the environment using the specification
// laid out in https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/exporter.md
func newOTelBridgeTracer(logger log.Logger, opts *options) (opentracing.Tracer, oteltrace.TracerProvider, io.Closer, error) {
	// We don't support OTEL_EXPORTER_OTLP_ENDPOINT yet, see newOTelCollectorExporter
	// docstring.
	endpoint := otlpenv.GetEndpoint()
	logger = logger.Scoped("otel", "OpenTelemetry tracer").With(log.String("endpoint", endpoint))

	// Ensure propagation between services continues to work. This is also done by another
	// project that uses the OpenTracing bridge:
	// https://sourcegraph.com/github.com/thanos-io/thanos/-/blob/pkg/tracing/migration/bridge.go?L62
	compositePropagator := w3cpropagator.NewCompositeTextMapPropagator(
		jaegerpropagator.Jaeger{},
		otpropagator.OT{},
		w3cpropagator.TraceContext{},
		w3cpropagator.Baggage{},
	)
	otel.SetTextMapPropagator(compositePropagator)

	// Initialize OpenTelemetry processor and tracer provider
	processor, err := newOTelCollectorExporter(context.Background(), logger, endpoint, opts.debug)
	if err != nil {
		return nil, nil, nil, err
	}
	provider := oteltracesdk.NewTracerProvider(
		oteltracesdk.WithResource(newResource(opts.resource)),
		oteltracesdk.WithSampler(oteltracesdk.AlwaysSample()),
		oteltracesdk.WithSpanProcessor(processor),
	)

	// Set up bridge for converting opentracing API calls to OpenTelemetry.
	bridge, otelTracerProvider := otelbridge.NewTracerPair(provider.Tracer("tracer.global"))
	bridge.SetTextMapPropagator(compositePropagator)

	// Set up logging
	otelLogger := logger.AddCallerSkip(2) // no additional scope needed, this is already otel scope
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) { otelLogger.Warn("error encountered", log.Error(err)) }))
	bridgeLogger := logger.AddCallerSkip(2).Scoped("bridge", "OpenTracing to OpenTelemetry compatibility layer")
	bridge.SetWarningHandler(func(msg string) { bridgeLogger.Debug(msg) })

	// Done
	return bridge, otelTracerProvider, &otelBridgeCloser{provider}, nil
}

// newOTelCollectorExporter creates a processor that exports spans to an OpenTelemetry
// collector.
func newOTelCollectorExporter(ctx context.Context, logger log.Logger, endpoint string, debug bool) (oteltracesdk.SpanProcessor, error) {
	// Set up client to otel-collector - we replicate some of the logic used internally in
	// https://github.com/open-telemetry/opentelemetry-go/blob/21c1641831ca19e3acf341cc11459c87b9791f2f/exporters/otlp/internal/otlpconfig/envconfig.go
	// based on our own inferred endpoint.
	var (
		client          otlptrace.Client
		protocol        = otlpenv.GetProtocol()
		trimmedEndpoint = trimSchema(endpoint)
		insecure        = otlpenv.IsInsecure(endpoint)
	)

	// Work with different protocols
	switch protocol {
	case otlpenv.ProtocolGRPC:
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(trimmedEndpoint),
		}
		if insecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}
		client = otlptracegrpc.NewClient(opts...)

	case otlpenv.ProtocolHTTPJSON:
		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(trimmedEndpoint),
		}
		if insecure {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		client = otlptracehttp.NewClient(opts...)
	}

	// Initialize the exporter
	traceExporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create trace exporter")
	}

	// If in debug mode, we use a synchronous span processor to force spans to get pushed
	// immediately.
	if debug {
		logger.Warn("using synchronous span processor - disable 'observability.debug' to use something more suitable for production")
		return oteltracesdk.NewSimpleSpanProcessor(traceExporter), nil
	}
	return oteltracesdk.NewBatchSpanProcessor(traceExporter), nil
}

// otelBridgeCloser shuts down the wrapped TracerProvider, and unsets the global OTel
// trace provider.
type otelBridgeCloser struct{ *oteltracesdk.TracerProvider }

var _ io.Closer = &otelBridgeCloser{}

func (p otelBridgeCloser) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return p.Shutdown(ctx)
}

// newResource adapts sourcegraph/log.Resource into the OpenTelemetry package's Resource
// type.
func newResource(r log.Resource) *resource.Resource {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(r.Name),
		semconv.ServiceNamespaceKey.String(r.Namespace),
		semconv.ServiceInstanceIDKey.String(r.InstanceID),
		semconv.ServiceVersionKey.String(r.Version))
}

var httpSchemeRegexp = regexp.MustCompile(`(?i)^http://|https://`)

func trimSchema(endpoint string) string {
	return httpSchemeRegexp.ReplaceAllString(endpoint, "")
}
