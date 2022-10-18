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

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/hostname"
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

	initTracer(logger, &options{resource: resource}, c)
}

// initTracer is a helper that should be called exactly once (from Init).
func initTracer(logger log.Logger, opts *options, c conftypes.WatchableSiteConfig) {
	// Initialize global, hot-swappable implementations of OpenTelemetry and OpenTracing
	// tracing.
	globalOTelTracerProvider := newSwitchableOtelTracerProvider(logger.Scoped("otel.global", "the global OpenTelemetry tracer"))
	otel.SetTracerProvider(globalOTelTracerProvider)
	globalOTTracer := newSwitchableOTTracer(logger.Scoped("ot.global", "the global OpenTracing tracer"))
	opentracing.SetGlobalTracer(globalOTTracer)

	// Initially everything is disabled since we haven't read conf yet. This variable is
	// also updated to compare against new version of configuration.
	go c.Watch(newConfWatcher(logger, c, globalOTelTracerProvider, globalOTTracer, options{
		resource:    opts.resource,
		TracerType:  None,
		debug:       false,
		externalURL: "",
	}))

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

// newTracer creates OpenTelemetry and OpenTracing tracers based on opts. It always returns
// valid tracers.
func newTracer(logger log.Logger, opts *options) (opentracing.Tracer, oteltrace.TracerProvider, io.Closer, error) {
	logger.Debug("configuring tracer")

	var exporter oteltracesdk.SpanExporter
	var err error
	switch opts.TracerType {
	case OpenTelemetry:
		exporter, err = exporters.NewOTLPExporter(context.Background(), logger)

	case Jaeger:
		exporter, err = exporters.NewJaegerExporter()

	default:
		err = errors.Newf("unknown tracer type %q", opts.TracerType)
	}

	if err != nil || exporter == nil {
		return opentracing.NoopTracer{}, trace.NewNoopTracerProvider(), nil, err
	}
	return newOTelBridgeTracer(logger, exporter, opts.resource, opts.debug)
}
