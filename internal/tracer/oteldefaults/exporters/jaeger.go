package exporters

import (
	"strings"

	jaegercfg "github.com/uber/jaeger-client-go/config"
	oteljaeger "go.opentelemetry.io/otel/exporters/jaeger"
	oteltracesdk "go.opentelemetry.io/otel/sdk/trace"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewJaegerExporter exports spans to a Jaeger collector or agent based on environment
// configuration.
//
// By default, prefer to use internal/tracer.Init to set up a global OpenTelemetry
// tracer and use that instead.
func NewJaegerExporter() (oteltracesdk.SpanExporter, error) {
	// Set configuration from jaegercfg package, to try and preserve back-compat with
	// existing behaviour.
	cfg, err := jaegercfg.FromEnv()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read Jaeger configuration from env")
	}
	var endpoint oteljaeger.EndpointOption
	switch {
	case cfg.Reporter.CollectorEndpoint != "":
		endpoint = oteljaeger.WithCollectorEndpoint(
			oteljaeger.WithEndpoint(cfg.Reporter.CollectorEndpoint),
			oteljaeger.WithUsername(cfg.Reporter.User),
			oteljaeger.WithPassword(cfg.Reporter.Password),
		)
	case cfg.Reporter.LocalAgentHostPort != "":
		hostport := strings.Split(cfg.Reporter.LocalAgentHostPort, ":")
		endpoint = oteljaeger.WithAgentEndpoint(
			oteljaeger.WithAgentHost(hostport[0]),
			oteljaeger.WithAgentPort(hostport[1]),
		)
	default:
		// Otherwise, oteljaeger defaults and env configuration
		endpoint = oteljaeger.WithAgentEndpoint()
	}

	// Create exporter for endpoint
	exporter, err := oteljaeger.New(endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create trace exporter")
	}
	return exporter, nil
}
