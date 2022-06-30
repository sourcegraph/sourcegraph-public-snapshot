package tracer

import (
	"context"
	"io"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/sourcegraph/log"
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

// TODO make this configurable?
const otelCollectorAddress = "localhost:30080"

// newOTelTracer creates an opentracing.Tracer that actually exports OpenTelemetry traces
// to an OpenTelemetry collector.
func newOTelTracer(opts *options) (opentracing.Tracer, io.Closer, error) {
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

	bridge, _ := otelbridge.NewTracerPair(provider.Tracer("global"))
	return bridge, &processorCloser{processor}, nil
}

func newOTelCollectorExporter(ctx context.Context) (oteltrace.SpanProcessor, error) {
	// If the OpenTelemetry Collector is running on a local cluster (minikube or
	// microk8s), it should be accessible through the NodePort service at the
	// `localhost:30080` endpoint. Otherwise, replace `localhost` with the
	// endpoint of your cluster. If you run the app inside k8s, then you can
	// probably connect directly to the service through dns
	//
	// TODO is this okay?
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
