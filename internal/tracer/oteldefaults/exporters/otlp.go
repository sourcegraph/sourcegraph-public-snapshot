package exporters

import (
	"context"

	"github.com/grafana/regexp"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	oteltracesdk "go.opentelemetry.io/otel/sdk/trace"

	"github.com/sourcegraph/sourcegraph/internal/otlpenv"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewOTLPTraceExporter exports trace spans to an OpenTelemetry collector via the
// OpenTelemetry protocol (OTLP) based on environment configuration.
//
// By default, prefer to use internal/tracer.Init to set up a global OpenTelemetry
// tracer and use that instead.
func NewOTLPTraceExporter(ctx context.Context, logger log.Logger) (oteltracesdk.SpanExporter, error) {
	endpoint := otlpenv.GetEndpoint()
	if endpoint == "" {
		// OTEL_EXPORTER_OTLP_ENDPOINT has been explicitly set to ""
		return nil, errors.Newf("please configure an exporter endpoint with OTEL_EXPORTER_OTLP_ENDPOINT")
	}

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
	return traceExporter, nil
}

var httpSchemeRegexp = regexp.MustCompile(`(?i)^http://|https://`)

func trimSchema(endpoint string) string {
	return httpSchemeRegexp.ReplaceAllString(endpoint, "")
}
