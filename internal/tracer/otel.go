package tracer

import (
	"context"
	"io"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/sourcegraph/log"
	otpropagator "go.opentelemetry.io/contrib/propagators/ot"
	"go.opentelemetry.io/otel"
	otelbridge "go.opentelemetry.io/otel/bridge/opentracing"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	oteltracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// If the OpenTelemetry Collector is running on a local cluster (minikube or
// microk8s), it should be accessible through the NodePort service at the
// `localhost:30080` endpoint. Otherwise, replace `localhost` with the
// endpoint of your cluster. If you run the app inside k8s, then you can
// probably connect directly to the service through dns
//
// OTEL_EXPORTER_OTLP_ENDPOINT is the name chosen because it is used in other
// projects: https://sourcegraph.com/search?q=OTEL_EXPORTER_OTLP_ENDPOINT+-f:vendor
var otelCollectorEndpoint = env.Get("OTEL_EXPORTER_OTLP_ENDPOINT", "127.0.0.1:4317", "Address of OpenTelemetry collector")

// newOTelBridgeTracer creates an opentracing.Tracer that exports all OpenTracing traces
// as OpenTelemetry traces to an OpenTelemetry collector (effectively "bridging" the two
// APIs). This enables us to continue leveraging the OpenTracing API (which is a predecessor
// to OpenTelemetry tracing) without making changes to existing tracing code.
func newOTelBridgeTracer(logger log.Logger, opts *options) (opentracing.Tracer, oteltrace.TracerProvider, io.Closer, error) {
	logger = logger.Scoped("otel", "OpenTelemetry tracer").
		With(log.String("otel-collector.endpoint", otelCollectorEndpoint))

	// Ensure propagation between services continues to work. This is also done by another
	// project that uses the OpenTracing bridge:
	// https://sourcegraph.com/github.com/thanos-io/thanos/-/blob/pkg/tracing/migration/bridge.go?L62
	compositePropagator := propagation.NewCompositeTextMapPropagator(otpropagator.OT{}, propagation.TraceContext{}, propagation.Baggage{})
	otel.SetTextMapPropagator(compositePropagator)

	// Initialize OpenTelemetry processor and tracer provider
	processor, err := newOTelCollectorExporter(context.Background(), logger, opts.debug)
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
	bridge.SetTextMapPropagator(propagation.TraceContext{})

	// Set up logging
	otelLogger := logger.AddCallerSkip(1) // no additional scope needed, this is already otel scope
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) { otelLogger.Warn("error encountered", log.Error(err)) }))
	bridgeLogger := logger.AddCallerSkip(1).Scoped("bridge", "OpenTracing to OpenTelemetry compatibility layer")
	bridge.SetWarningHandler(func(msg string) { bridgeLogger.Debug(msg) })

	// Done
	return &otelBridgeTracer{bridge}, otelTracerProvider, &otelBridgeCloser{provider}, nil
}

// newOTelCollectorExporter creates a processor that exports spans to an OpenTelemetry
// collector.
func newOTelCollectorExporter(ctx context.Context, logger log.Logger, debug bool) (oteltracesdk.SpanProcessor, error) {
	conn, err := grpc.DialContext(ctx, otelCollectorEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create gRPC connection to collector")
	}

	// Set up a trace exporter
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
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

// otelBridgeTracer wraps bridge.BridgeTracer with extended Inject/Extract support for
// opentracing.TextMap which is used in the codebase and similar carriers. It is adapted
// from the 'thanos-io/thanos' project
// https://sourcegraph.com/github.com/thanos-io/thanos@4de555db87d38d69b78602c1e1d0fb8ed6e0371b/-/blob/pkg/tracing/migration/bridge.go?L76-88#tab=references
// but behaves differently, disregarding the provided format if we decide to override it.
//
// The main issue is that bridge.BridgeTracer currently supports injection /
// extraction of only single carrier type which is opentracing.HTTPHeadersCarrier. See:
//
// - https://github.com/open-telemetry/opentelemetry-go/blob/c2dc940e0b48e61712e4f8f6f2320d8fd4c9aac6/bridge/opentracing/bridge.go#L634-L638
// - https://github.com/open-telemetry/opentelemetry-go/blob/c2dc940e0b48e61712e4f8f6f2320d8fd4c9aac6/bridge/opentracing/bridge.go#L664-L668
type otelBridgeTracer struct{ bridge *otelbridge.BridgeTracer }

var _ opentracing.Tracer = &otelBridgeTracer{}

func (b *otelBridgeTracer) StartSpan(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	return b.bridge.StartSpan(operationName, opts...)
}

func (b *otelBridgeTracer) Inject(span opentracing.SpanContext, format interface{}, carrier interface{}) error {
	// Inject into a blank HTTPHeaders carrier first - we use this as a source for our
	// wrapped Inject implementation.
	otCarrier := opentracing.HTTPHeadersCarrier{}
	err := b.bridge.Inject(span, opentracing.HTTPHeaders, otCarrier)
	if err != nil {
		return err
	}

	// Regardless of format, inject context into the TextMapWriter if there is one. If we
	// do this, there is no need to pass this on to the underlying Inject implemenation
	if tmw, ok := carrier.(opentracing.TextMapWriter); ok {
		return otCarrier.ForeachKey(func(key, val string) error {
			tmw.Set(key, val)
			return nil
		})
	}

	// If we are receiving some other non-TextMapWriter type, pass it on and hope for the
	// best.
	return b.bridge.Inject(span, format, carrier)
}

func (b *otelBridgeTracer) Extract(format interface{}, carrier interface{}) (opentracing.SpanContext, error) {
	// Regardless of format, extract TextMapReader content into an HTTPHeadersCarrier
	if tmr, ok := carrier.(opentracing.TextMapReader); ok {
		otCarrier := opentracing.HTTPHeadersCarrier{}
		err := tmr.ForeachKey(func(key, val string) error {
			otCarrier.Set(key, val)
			return nil
		})
		if err != nil {
			return nil, err
		}

		return b.bridge.Extract(opentracing.HTTPHeaders, otCarrier)
	}

	return b.bridge.Extract(format, carrier)
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
