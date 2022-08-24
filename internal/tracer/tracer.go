package tracer

import (
	"context"
	"fmt"
	"io"
	"text/template"

	"github.com/opentracing/opentracing-go"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel"
	oteltracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/automaxprocs/maxprocs"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/hostname"
	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
	"github.com/sourcegraph/sourcegraph/internal/version"

	"github.com/sourcegraph/sourcegraph/internal/tracer/internal/exporters"
)

// options control the behavior of a TracerType
type options struct {
	TracerType
	externalURL string
	debug       bool
	// these values are not configurable by site config
	resource log.Resource
}

type TracerType string

const (
	None TracerType = "none"

	// Jaeger and openTracing should be treated as analagous - the 'opentracing' moniker
	// is for backwards compatibility only, 'jaeger' is more correct because we export
	// Jaeger traces in 'opentracing' mode because 'opentracing' itself is an implementation
	// detail, it does not have a wire protocol.
	Jaeger      TracerType = "jaeger"
	openTracing TracerType = "opentracing"

	// OpenTelemetry exports traces over OTLP.
	OpenTelemetry TracerType = "opentelemetry"
)

// isSetByUser returns true if the TracerType is one supported by the schema
// should be kept in sync with ObservabilityTracing.Type in schema/site.schema.json
func (t TracerType) isSetByUser() bool {
	switch t {
	case openTracing, Jaeger, OpenTelemetry:
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

	initTracer(logger, &options{resource: resource}, c)
}

// initTracer is a helper that should be called exactly once (from Init).
func initTracer(logger log.Logger, opts *options, c conftypes.WatchableSiteConfig) {
	// Initialize global, hot-swappable implementations of OpenTracing and OpenTelemetry
	// tracing.
	globalOTTracer := newSwitchableOTTracer(logger.Scoped("ot.global", "the global OpenTracing tracer"))
	opentracing.SetGlobalTracer(globalOTTracer)
	globalOTelTracerProvider := newSwitchableOtelTracerProvider(logger.Scoped("otel.global", "the global OpenTelemetry tracer"))
	otel.SetTracerProvider(globalOTelTracerProvider)

	// Initially everything is disabled since we haven't read conf yet. This variable is
	// also updated to compare against new version of configuration.
	oldOpts := options{
		resource: opts.resource,
		// the values below may change
		TracerType:  None,
		debug:       false,
		externalURL: "",
	}

	// Watch loop
	go c.Watch(func() {
		var (
			siteConfig = c.SiteConfig()
			debug      = false
			setTracer  = None
		)

		if tracingConfig := siteConfig.ObservabilityTracing; tracingConfig != nil {
			debug = tracingConfig.Debug

			// If sampling policy is set, update the strategy and set our tracer to be
			// Jaeger by default.
			previousPolicy := policy.GetTracePolicy()
			switch p := policy.TracePolicy(tracingConfig.Sampling); p {
			case policy.TraceAll, policy.TraceSelective:
				policy.SetTracePolicy(p)
				// enable the defualt tracer type. TODO in 4.0, this should be OpenTelemetry
				setTracer = Jaeger
			default:
				policy.SetTracePolicy(policy.TraceNone)
			}
			if newPolicy := policy.GetTracePolicy(); newPolicy != previousPolicy {
				logger.Info("updating TracePolicy",
					log.String("oldValue", string(previousPolicy)),
					log.String("newValue", string(newPolicy)))
			}

			// If the tracer type is configured, also set the tracer type
			if t := TracerType(tracingConfig.Type); t.isSetByUser() {
				setTracer = t
			}
		}

		opts := options{
			TracerType:  setTracer,
			externalURL: siteConfig.ExternalURL,
			debug:       debug,
			// Stays the same
			resource: oldOpts.resource,
		}
		if opts == oldOpts {
			// Nothing changed
			return
		}

		// update old opts for comparison
		oldOpts = opts

		// create new tracer providers
		tracerLogger := logger.With(
			log.String("tracerType", string(opts.TracerType)),
			log.Bool("debug", opts.debug))
		otImpl, otelImpl, closer, err := newTracer(tracerLogger, &opts)
		if err != nil {
			tracerLogger.Warn("failed to initialize tracer", log.Error(err))
			return
		}

		// update global tracers. for now, we let the OT tracer handle shutdown when
		// switching (we always switch in tandem, so this is fine)
		globalOTTracer.set(tracerLogger, otImpl, closer, opts.debug)
		globalOTelTracerProvider.set(otelImpl, opts.debug)
	})

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

// newTracer creates OpenTelemetry and OpenTracing tracers based on opts
func newTracer(logger log.Logger, opts *options) (opentracing.Tracer, oteltrace.TracerProvider, io.Closer, error) {
	logger.Debug("configuring tracer")

	var exporter oteltracesdk.SpanExporter
	var err error
	switch opts.TracerType {
	case Jaeger, openTracing:
		exporter, err = exporters.NewJaegerExporter()
	case OpenTelemetry:
		exporter, err = exporters.NewOTelCollectorExporter(context.Background(), logger)
	}

	if err != nil || exporter == nil {
		return opentracing.NoopTracer{}, trace.NewNoopTracerProvider(), nil, err
	}
	return newOTelBridgeTracer(logger, exporter, opts.resource, opts.debug)
}
