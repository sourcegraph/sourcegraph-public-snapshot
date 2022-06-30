package tracer

import (
	"context"
	"io"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel"
	otelbridge "go.opentelemetry.io/otel/bridge/opentracing"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	oteltrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/hostname"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// If the OpenTelemetry Collector is running on a local cluster (minikube or
// microk8s), it should be accessible through the NodePort service at the
// `localhost:30080` endpoint. Otherwise, replace `localhost` with the
// endpoint of your cluster. If you run the app inside k8s, then you can
// probably connect directly to the service through dns
//
// TODO make this configurable?
const otelCollectorAddress = "localhost:4317"

// newOTelTracer creates an opentracing.Tracer that actually exports OpenTelemetry traces
// to an OpenTelemetry collector.
func newOTelTracer(logger log.Logger, opts *options) (opentracing.Tracer, io.Closer, error) {
	logger = logger.Scoped("otel", "OpenTelemetry tracer")

	processor, err := newOTelCollectorExporter(context.Background())
	if err != nil {
		return nil, nil, err
	}

	// TODO should this be prop-drilled? this is the same as the fields used to initialize
	// sourcegraph/log - but we might be able to get away with setting this up here,
	// since this is an internal package.
	resource := log.Resource{
		Name:       env.MyName,
		Version:    version.Version(),
		InstanceID: hostname.Get(),
	}

	provider := oteltrace.NewTracerProvider(
		oteltrace.WithResource(newResource(resource)),
		oteltrace.WithSpanProcessor(processor),
		// Sampling doesn't have to be configured, oteltrace falls back to AlwaysSample
	)

	// Set up OpenTelemetry package
	// otel.SetTracerProvider(provider)

	// Set up bridge
	bridge, _ := otelbridge.NewTracerPair(provider.Tracer("global"))

	// Set up logging
	otelLogger := logger.AddCallerSkip(1) // no additional scope needed, this is already otel scope
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) { otelLogger.Warn("error encountered", log.Error(err)) }))
	bridgeLogger := logger.AddCallerSkip(1).Scoped("bridge", "OpenTracing to OpenTelemetry compatibility layer")
	bridge.SetWarningHandler(func(msg string) { bridgeLogger.Warn(msg) })

	// Done
	return &bridgeTracerWrapper{bridge}, &processorCloser{processor}, nil
}

func newOTelCollectorExporter(ctx context.Context) (oteltrace.SpanProcessor, error) {
	conn, err := grpc.DialContext(ctx, otelCollectorAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create gRPC connection to collector")
	}

	// Set up a trace exporter
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create trace exporter")
	}

	return oteltrace.NewBatchSpanProcessor(traceExporter), nil
}

type processorCloser struct{ oteltrace.SpanProcessor }

var _ io.Closer = &processorCloser{}

func (p processorCloser) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return p.Shutdown(ctx)
}

func newResource(r log.Resource) *resource.Resource {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(r.Name),
		semconv.ServiceNamespaceKey.String(r.Namespace),
		semconv.ServiceInstanceIDKey.String(r.InstanceID),
		semconv.ServiceVersionKey.String(r.Version))
}

// bridgeTracerWrapper wraps bridge.BridgeTracer with extended Inject/Extract support for
// opentracing.TextMap which is used in the codebase and similar carriers. It is adapted
// from the 'thanos-io/thanos' project
// https://sourcegraph.com/github.com/thanos-io/thanos/-/blob/pkg/tracing/migration/bridge.go?L53:6#tab=references
// but behaves differently, disregarding the provided format if we decide to override it.
//
// The main issue is that bridge.BridgeTracer currently supports injection /
// extraction of only single carrier type which is opentracing.HTTPHeadersCarrier.
// (see https://github.com/open-telemetry/opentelemetry-go/blob/main/bridge/opentracing/bridge.go#L626)
type bridgeTracerWrapper struct{ bt *otelbridge.BridgeTracer }

var _ opentracing.Tracer = &bridgeTracerWrapper{}

func (b *bridgeTracerWrapper) StartSpan(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	return b.bt.StartSpan(operationName, opts...)
}

func (b *bridgeTracerWrapper) Inject(sm opentracing.SpanContext, format interface{}, carrier interface{}) error {
	// Inject HTTPHeaders carrier
	otCarrier := opentracing.HTTPHeadersCarrier{}
	err := b.bt.Inject(sm, opentracing.HTTPHeaders, otCarrier)
	if err != nil {
		return err
	}

	// Regardless of format, inject context into the TextMapWriter if there is one
	if tmw, ok := carrier.(opentracing.TextMapWriter); ok {
		if err := otCarrier.ForeachKey(func(key, val string) error {
			tmw.Set(key, val)
			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

func (b *bridgeTracerWrapper) Extract(format interface{}, carrier interface{}) (opentracing.SpanContext, error) {
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

		return b.bt.Extract(opentracing.HTTPHeaders, otCarrier)
	}

	return b.bt.Extract(format, carrier)
}
