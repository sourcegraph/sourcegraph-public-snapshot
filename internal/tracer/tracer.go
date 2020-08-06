// Package tracer initializes distributed tracing and log15 behavior. It also updates distributed
// tracing behavior in response to changes in site configuration. When the Init function of this
// package is invoked, opentracing.SetGlobalTracer is called (and subsequently called again after
// every Sourcegraph site configuration change). Importing programs should not invoke
// opentracing.SetGlobalTracer anywhere else.
package tracer

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"sync"

	"github.com/fatih/color"
	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"go.uber.org/automaxprocs/maxprocs"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	jaegermetrics "github.com/uber/jaeger-lib/metrics"
)

var (
	logColors = map[log15.Lvl]color.Attribute{
		log15.LvlCrit:  color.FgRed,
		log15.LvlError: color.FgRed,
		log15.LvlWarn:  color.FgYellow,
		log15.LvlInfo:  color.FgCyan,
		log15.LvlDebug: color.Faint,
	}
	// We'd prefer these in caps, not lowercase, and don't need the 4-character alignment
	logLabels = map[log15.Lvl]string{
		log15.LvlCrit:  "CRITICAL",
		log15.LvlError: "ERROR",
		log15.LvlWarn:  "WARN",
		log15.LvlInfo:  "INFO",
		log15.LvlDebug: "DEBUG",
	}
)

func init() {
	// Tune GOMAXPROCS for kubernetes. All our binaries import this package,
	// so we tune for all of them.
	//
	// TODO it is surprising that we do this here. We should create a standard
	// import for sourcegraph binaries which would have less surprising
	// behaviour.
	if _, err := maxprocs.Set(); err != nil {
		log15.Error("automaxprocs failed", "error", err)
	}
}

func condensedFormat(r *log15.Record) []byte {
	colorAttr := logColors[r.Lvl]
	text := logLabels[r.Lvl]
	var msg bytes.Buffer
	if colorAttr != 0 {
		fmt.Print(color.New(colorAttr).Sprint(text) + " " + r.Msg)
	} else {
		fmt.Print(&msg, r.Msg)
	}
	if len(r.Ctx) > 0 {
		for i := 0; i < len(r.Ctx); i += 2 {
			// not as smart about printing things as log15's internal magic
			fmt.Fprintf(&msg, ", %s: %v", r.Ctx[i].(string), r.Ctx[i+1])
		}
	}
	msg.WriteByte('\n')
	return msg.Bytes()
}

// Options control the behavior of a tracer.
type Options struct {
	filters     []func(*log15.Record) bool
	serviceName string
}

// If this idiom seems strange:
// https://github.com/tmrts/go-patterns/blob/master/idiom/functional-options.md
type Option func(*Options)

func ServiceName(s string) Option {
	return func(o *Options) {
		o.serviceName = s
	}
}

func Filter(f func(*log15.Record) bool) Option {
	return func(o *Options) {
		o.filters = append(o.filters, f)
	}
}

func init() {
	// Enable colors by default but support https://no-color.org/
	color.NoColor = env.Get("NO_COLOR", "", "Disable colored output") != ""
}

func Init(options ...Option) {
	opts := &Options{}
	for _, setter := range options {
		setter(opts)
	}
	if opts.serviceName == "" {
		opts.serviceName = env.MyName
	}
	var handler log15.Handler
	switch env.LogFormat {
	case "condensed":
		handler = log15.StreamHandler(os.Stderr, log15.FormatFunc(condensedFormat))
	case "logfmt":
		fallthrough
	default:
		handler = log15.StreamHandler(os.Stderr, log15.LogfmtFormat())
	}
	for _, filter := range opts.filters {
		handler = log15.FilterHandler(filter, handler)
	}
	// Filter log output by level.
	lvl, err := log15.LvlFromString(env.LogLevel)
	if err == nil {
		handler = log15.LvlFilterHandler(lvl, handler)
	}
	log15.Root().SetHandler(log15.LvlFilterHandler(lvl, handler))

	initTracer(opts.serviceName)
}

type jaegerOpts struct {
	ServiceName string
	ExternalURL string
	Enabled     bool
	Debug       bool
}

// initTracer is a helper that should be called exactly once (from Init).
func initTracer(serviceName string) {
	globalTracer := newSwitchableTracer()
	opentracing.SetGlobalTracer(globalTracer)

	// Initially everything is disabled since we haven't read conf yet.
	oldOpts := jaegerOpts{
		ServiceName: serviceName,
		ExternalURL: conf.Get().ExternalURL,
		Enabled:     false,
		Debug:       false,
	}

	// Watch loop
	conf.Watch(func() {
		siteConfig := conf.Get()

		// Set sampling strategy
		samplingStrategy := ot.TraceNone
		shouldLog := false
		if tracingConfig := siteConfig.ObservabilityTracing; tracingConfig != nil {
			switch tracingConfig.Sampling {
			case "all":
				samplingStrategy = ot.TraceAll
			case "selective":
				samplingStrategy = ot.TraceSelective
			}
			shouldLog = tracingConfig.Debug
		} else if siteConfig.UseJaeger {
			samplingStrategy = ot.TraceAll
		}
		if tracePolicy := ot.GetTracePolicy(); tracePolicy != samplingStrategy {
			log15.Info("opentracing: TracePolicy", "oldValue", tracePolicy, "newValue", samplingStrategy)
		}
		ot.SetTracePolicy(samplingStrategy)

		opts := jaegerOpts{
			ServiceName: serviceName,
			ExternalURL: siteConfig.ExternalURL,
			Enabled:     samplingStrategy == ot.TraceAll || samplingStrategy == ot.TraceSelective,
			Debug:       shouldLog,
		}

		if opts == oldOpts {
			// Nothing changed
			return
		}

		oldOpts = opts

		tracer, urlFunc, closer, err := newTracer(&opts)
		if err != nil {
			log15.Warn("Could not initialize jaeger tracer", "error", err.Error())
			return
		}

		globalTracer.set(tracer, closer, opts.Debug)
		trace.SpanURL = urlFunc
	})
}

func newTracer(opts *jaegerOpts) (opentracing.Tracer, func(span opentracing.Span) string, io.Closer, error) {
	if !opts.Enabled {
		log15.Info("opentracing: Jaeger disabled")
		return opentracing.NoopTracer{}, trace.NoopSpanURL, nil, nil
	}

	log15.Info("opentracing: Jaeger enabled")
	cfg, err := jaegercfg.FromEnv()
	cfg.ServiceName = opts.ServiceName
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "jaegercfg.FromEnv failed")
	}
	if reflect.DeepEqual(cfg.Sampler, &jaegercfg.SamplerConfig{}) {
		// Default sampler configuration for when it is not specified via
		// JAEGER_SAMPLER_* env vars. In most cases, this is sufficient
		// enough to connect Sourcegraph to Jaeger without any env vars.
		cfg.Sampler.Type = jaeger.SamplerTypeConst
		cfg.Sampler.Param = 1
	}
	tracer, closer, err := cfg.NewTracer(
		jaegercfg.Logger(jaegerlog.StdLogger),
		jaegercfg.Metrics(jaegermetrics.NullFactory),
	)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "jaegercfg.NewTracer failed")
	}

	// We proxy jaeger so we can construct URLs to traces.
	jaegerURL := opts.ExternalURL + "/-/debug/jaeger/trace/"

	spanURL := func(span opentracing.Span) string {
		if span == nil {
			return tracingNotEnabledURL
		}
		spanCtx, ok := span.Context().(jaeger.SpanContext)
		if !ok {
			return tracingNotEnabledURL
		}
		return jaegerURL + spanCtx.TraceID().String()
	}

	return tracer, spanURL, closer, nil
}

// switchableTracer implements opentracing.Tracer. The underlying tracer used is switchable (set via
// the `set` method).
type switchableTracer struct {
	mu           sync.RWMutex
	tracer       opentracing.Tracer
	tracerCloser io.Closer
	log          bool
}

func newSwitchableTracer() *switchableTracer {
	return &switchableTracer{tracer: opentracing.NoopTracer{}}
}

func (t *switchableTracer) StartSpan(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.log {
		log15.Info("opentracing: StartSpan", "operationName", operationName, "tracer", fmt.Sprintf("%T", t.tracer))
	}
	return t.tracer.StartSpan(operationName, opts...)
}

func (t *switchableTracer) Inject(sm opentracing.SpanContext, format interface{}, carrier interface{}) error {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.log {
		log15.Info("opentracing: Inject", "tracer", fmt.Sprintf("%T", t.tracer))
	}
	return t.tracer.Inject(sm, format, carrier)
}

func (t *switchableTracer) Extract(format interface{}, carrier interface{}) (opentracing.SpanContext, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.log {
		log15.Info("opentracing: Extract", "tracer", fmt.Sprintf("%T", t.tracer))
	}
	return t.tracer.Extract(format, carrier)
}

func (t *switchableTracer) set(tracer opentracing.Tracer, tracerCloser io.Closer, log bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if tc := t.tracerCloser; tc != nil {
		// Close the old tracerCloser outside the critical zone
		go tc.Close()
	}

	t.tracerCloser = tracerCloser
	t.tracer = tracer
	t.log = log
}

func (t *switchableTracer) get() (tracer opentracing.Tracer, log bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.tracer, t.log
}

const tracingNotEnabledURL = "#tracing_not_enabled_for_this_request_add_?trace=1_to_url_to_enable"
