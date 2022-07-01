package tracer

import (
	"fmt"
	"io"
	"reflect"
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/sourcegraph/log"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegermetrics "github.com/uber/jaeger-lib/metrics"
	"go.uber.org/automaxprocs/maxprocs"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// options control the behavior of a TracerType
type options struct {
	TracerType
	externalURL string
	debug       bool
	// these values are not configurable by site config
	serviceName string
	version     string
	env         string
}

type TracerType string

const (
	None        TracerType = "none"
	OpenTracing TracerType = "opentracing"
)

// isSetByUser returns true if the TracerType is one supported by the schema
// should be kept in sync with ObservabilityTracing.Type in schema/site.schema.json
func (t TracerType) isSetByUser() bool {
	switch t {
	case OpenTracing:
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

	opts := &options{}
	opts.serviceName = env.MyName
	if version.IsDev(version.Version()) {
		opts.env = "dev"
	}
	opts.version = version.Version()

	initTracer(logger, opts, c)
}

// initTracer is a helper that should be called exactly once (from Init).
func initTracer(logger log.Logger, opts *options, c conftypes.WatchableSiteConfig) {
	globalTracer := newSwitchableTracer(logger.Scoped("global", "the global tracer"))
	opentracing.SetGlobalTracer(globalTracer)

	// initial tracks if it's our first run of conf.Watch. This is used to
	// prevent logging "changes" when it's the first run.
	initial := true

	// Initially everything is disabled since we haven't read conf yet.
	oldOpts := options{
		serviceName: opts.serviceName,
		version:     opts.version,
		env:         opts.env,
		// the values below may change
		TracerType:  None,
		debug:       false,
		externalURL: "",
	}

	// Watch loop
	go c.Watch(func() {
		siteConfig := c.SiteConfig()

		samplingStrategy := policy.TraceNone
		shouldLog := false
		setTracer := None
		if tracingConfig := siteConfig.ObservabilityTracing; tracingConfig != nil {
			switch tracingConfig.Sampling {
			case "all":
				samplingStrategy = policy.TraceAll
				setTracer = OpenTracing
			case "selective":
				samplingStrategy = policy.TraceSelective
				setTracer = OpenTracing
			}
			if t := TracerType(tracingConfig.Type); t.isSetByUser() {
				setTracer = t
			}
			shouldLog = tracingConfig.Debug
		}
		if tracePolicy := policy.GetTracePolicy(); tracePolicy != samplingStrategy && !initial {
			log15.Info("opentracing: TracePolicy", "oldValue", tracePolicy, "newValue", samplingStrategy)
		}
		initial = false
		policy.SetTracePolicy(samplingStrategy)

		opts := options{
			externalURL: siteConfig.ExternalURL,
			TracerType:  setTracer,
			debug:       shouldLog,
			serviceName: opts.serviceName,
			version:     opts.version,
			env:         opts.env,
		}

		if opts == oldOpts {
			// Nothing changed
			return
		}
		prevTracer := oldOpts.TracerType
		oldOpts = opts

		t, closer, err := newTracer(logger, &opts, prevTracer)
		if err != nil {
			logger.Warn("Could not initialize tracer",
				log.String("tracer", string(opts.TracerType)),
				log.Error(err))
			return
		}
		globalTracer.set(t, closer, opts.debug)
	})
}

// TODO Use openTelemetry https://github.com/sourcegraph/sourcegraph/issues/27386
func newTracer(logger log.Logger, opts *options, prevTracer TracerType) (opentracing.Tracer, io.Closer, error) {
	if opts.TracerType == None {
		logger.Info("tracing disabled")
		return opentracing.NoopTracer{}, nil, nil
	}

	logger.Info("opentracing: enabled")
	cfg, err := jaegercfg.FromEnv()
	cfg.ServiceName = opts.serviceName
	if err != nil {
		return nil, nil, errors.Wrap(err, "jaegercfg.FromEnv failed")
	}
	cfg.Tags = append(cfg.Tags, opentracing.Tag{Key: "service.version", Value: opts.version}, opentracing.Tag{Key: "service.env", Value: opts.env})
	if reflect.DeepEqual(cfg.Sampler, &jaegercfg.SamplerConfig{}) {
		// Default sampler configuration for when it is not specified via
		// JAEGER_SAMPLER_* env vars. In most cases, this is sufficient
		// enough to connect Sourcegraph to Jaeger without any env vars.
		cfg.Sampler.Type = jaeger.SamplerTypeConst
		cfg.Sampler.Param = 1
	}
	tracer, closer, err := cfg.NewTracer(
		jaegercfg.Logger(jaegerLoggerShim{logger: logger.Scoped("jaeger", "Jaeger tracer")}),
		jaegercfg.Metrics(jaegermetrics.NullFactory),
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "jaegercfg.NewTracer failed")
	}

	return tracer, closer, nil
}

type jaegerLoggerShim struct {
	logger log.Logger
}

func (l jaegerLoggerShim) Error(msg string) { l.logger.Error(msg) }

func (l jaegerLoggerShim) Infof(msg string, args ...any) {
	l.logger.Info(fmt.Sprintf(msg, args...))
}

// move to OpenTelemetry https://github.com/sourcegraph/sourcegraph/issues/27386
// switchableTracer implements opentracing.Tracer. The underlying opentracer used is switchable (set via
// the `set` method).
type switchableTracer struct {
	mu           sync.RWMutex
	opentracer   opentracing.Tracer
	tracerCloser io.Closer

	log          bool
	logger       log.Logger
	parentLogger log.Logger // used to create logger
}

var _ opentracing.Tracer = &switchableTracer{}

// move to OpenTelemetry https://github.com/sourcegraph/sourcegraph/issues/27386
func newSwitchableTracer(logger log.Logger) *switchableTracer {
	var t opentracing.NoopTracer
	return &switchableTracer{
		opentracer:   t,
		logger:       logger.With(log.String("opentracer", fmt.Sprintf("%T", t))),
		parentLogger: logger,
	}
}

func (t *switchableTracer) StartSpan(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.log {
		t.logger.Info("opentracing: StartSpan",
			log.String("operationName", operationName))
	}
	return t.opentracer.StartSpan(operationName, opts...)
}

func (t *switchableTracer) Inject(sm opentracing.SpanContext, format any, carrier any) error {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.log {
		t.logger.Info("opentracing: Inject")
	}
	return t.opentracer.Inject(sm, format, carrier)
}

func (t *switchableTracer) Extract(format any, carrier any) (opentracing.SpanContext, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.log {
		t.logger.Info("opentracing: Extract")
	}
	return t.opentracer.Extract(format, carrier)
}

func (t *switchableTracer) set(tracer opentracing.Tracer, tracerCloser io.Closer, shouldLog bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if tc := t.tracerCloser; tc != nil {
		// Close the old tracerCloser outside the critical zone
		go tc.Close()
	}

	t.tracerCloser = tracerCloser
	t.opentracer = tracer
	t.log = shouldLog
	t.logger = t.parentLogger.With(log.String("opentracer", fmt.Sprintf("%T", t)))
}
