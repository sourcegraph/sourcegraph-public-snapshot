package tracer

import (
	"fmt"
	"sync/atomic"
	"text/template"

	"github.com/opentracing/opentracing-go"
	"github.com/sourcegraph/log"
	jaegerpropagator "go.opentelemetry.io/contrib/propagators/jaeger"
	otpropagator "go.opentelemetry.io/contrib/propagators/ot"
	"go.opentelemetry.io/otel"
	otelbridge "go.opentelemetry.io/otel/bridge/opentracing"
	w3cpropagator "go.opentelemetry.io/otel/propagation"
	oteltracesdk "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/automaxprocs/maxprocs"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/hostname"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

// options control the behavior of a TracerType
type options struct {
	TracerType
	externalURL string
	// these values are not configurable by site config
	resource log.Resource
}

type TracerType string

const (
	None TracerType = "none"

	// Jaeger exports traces over the Jaeger thrift protocol.
	Jaeger TracerType = "jaeger"

	// OpenTelemetry exports traces over OTLP.
	OpenTelemetry TracerType = "opentelemetry"
)

// DefaultTracerType is the default tracer type if not explicitly set by the user and
// some trace policy is enabled.
const DefaultTracerType = OpenTelemetry

// isSetByUser returns true if the TracerType is one supported by the schema
// should be kept in sync with ObservabilityTracing.Type in schema/site.schema.json
func (t TracerType) isSetByUser() bool {
	switch t {
	case Jaeger, OpenTelemetry:
		return true
	}
	return false
}

// Init should be called from the main function of service
func Init(logger log.Logger, c conftypes.WatchableSiteConfig) {
	// Tune GOMAXPROCS for kubernetes. All our binaries import this package,
	// so we tune for all of them.
	//
	// TODO it is surprising that we do this here. We should create a standard
	// import for sourcegraph binaries which would have less surprising
	// behaviour.
	if _, err := maxprocs.Set(); err != nil {
		logger.Error("automaxprocs failed", log.Error(err))
	}

	// Resource mirrors the initialization used by our OpenTelemetry logger.
	resource := log.Resource{
		Name:       env.MyName,
		Version:    version.Version(),
		InstanceID: hostname.Get(),
	}

	// Additionally set a dev namespace
	if version.IsDev(version.Version()) {
		resource.Namespace = "dev"
	}

	// Set up initial configurations
	debugMode := &atomic.Bool{}
	provider := newOtelTracerProvider(resource)

	// Create and set up global tracers from provider. We will be making updates to these
	// tracers through the debugMode ref and underlying provider.
	otTracer, otelTracerProvider := newBridgeTracers(logger, provider, debugMode)
	opentracing.SetGlobalTracer(otTracer)
	otel.SetTracerProvider(otelTracerProvider)

	// Initially everything is disabled since we haven't read conf yet - start a goroutine
	// that watches for updates to configure the undelrying provider and debugMode.
	go c.Watch(newConfWatcher(logger, c, provider, newOtelSpanProcessor, debugMode))

	// Contribute validation for tracing package
	conf.ContributeWarning(func(c conftypes.SiteConfigQuerier) conf.Problems {
		tracing := c.SiteConfig().ObservabilityTracing
		if tracing == nil || tracing.UrlTemplate == "" {
			return nil
		}
		if _, err := template.New("").Parse(tracing.UrlTemplate); err != nil {
			return conf.NewSiteProblems(fmt.Sprintf("observability.tracing.traceURL is not a valid template: %s", err.Error()))
		}
		return nil
	})
}

// newBridgeTracers creates an opentracing.Tracer that exports all OpenTracing traces,
// allowing us to continue leveraging the OpenTracing API (which is a predecessor to
// OpenTelemetry tracing) without making changes to existing tracing code. The returned
// opentracing.Tracer and oteltrace.TracerProvider should be set as global defaults for
// their respective libraries.
//
// All configuration should be sourced directly from the environment using the specification
// laid out in https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/exporter.md
func newBridgeTracers(logger log.Logger, provider *oteltracesdk.TracerProvider, debugMode *atomic.Bool) (opentracing.Tracer, oteltrace.TracerProvider) {
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

	// Set up otBridgeTracer for converting OpenTracing API calls to OpenTelemetry, and
	// otelTracerProvider for the inverse.
	//
	// TODO: Unfortunately, this wrapped tracer provider discards the name provided to
	// the Tracer() constructor on it, since it uses a fixed tracer that we provide it.
	// This is implemented in https://github.com/sourcegraph/sourcegraph/pull/40945
	otBridgeTracer, otelTracerProvider := otelbridge.NewTracerPair(provider.Tracer("internal/tracer/otel"))
	otBridgeTracer.SetTextMapPropagator(compositePropagator)

	// Set up logging
	otelLogger := logger.AddCallerSkip(2).Scoped("otel", "OpenTelemetry library")
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		if debugMode.Load() {
			otelLogger.Warn("error encountered", log.Error(err))
		} else {
			otelLogger.Debug("error encountered", log.Error(err))
		}
	}))
	bridgeLogger := logger.AddCallerSkip(2).Scoped("ot.bridge", "OpenTracing to OpenTelemetry compatibility layer")
	otBridgeTracer.SetWarningHandler(func(msg string) {
		if debugMode.Load() {
			bridgeLogger.Warn(msg)
		} else {
			bridgeLogger.Debug(msg)
		}
	})

	// Wrap each tracer in additional logging
	return newLoggedOTTracer(logger, otBridgeTracer, debugMode),
		newLoggedOtelTracerProvider(logger, otelTracerProvider, debugMode)
}
