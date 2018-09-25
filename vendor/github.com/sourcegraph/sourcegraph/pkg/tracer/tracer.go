package tracer

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/pkg/trace"
	log15 "gopkg.in/inconshreveable/log15.v2"

	lightstep "github.com/lightstep/lightstep-tracer-go"
	opentracing "github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	jaegermetrics "github.com/uber/jaeger-lib/metrics"
)

// Note: these configuration options require service restarts to take effect.
// Keep in sync with the pkg/conf.requireRestart list.
var (
	lightstepAccessToken         = conf.Get().LightstepAccessToken
	lightstepProject             = conf.Get().LightstepProject
	lightstepIncludeSensitive, _ = strconv.ParseBool(env.Get("LIGHTSTEP_INCLUDE_SENSITIVE", "", "send span logs to LightStep"))
	useJaeger                    = conf.Get().UseJaeger
	logColors                    = []int{5, 1, 3, 2, 6}
	// We'd prefer these in caps, not lowercase, and don't need the 4-character alignment
	logNames = []string{"CRIT", "ERROR", "WARN", "INFO", "DEBUG"}
)

func condensedFormat(r *log15.Record) []byte {
	color := 0
	text := "NOTE"
	if int(r.Lvl) >= 0 && int(r.Lvl) < len(logColors) {
		color = logColors[int(r.Lvl)]
		text = logNames[int(r.Lvl)]
	}
	var msg bytes.Buffer
	if color > 0 {
		fmt.Fprintf(&msg, "\x1b[3%dm%s\x1b[0m %s", color, text, r.Msg)
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
	if useJaeger {
		log15.Info("Distributed tracing enabled", "tracer", "jaeger")
		cfg := jaegercfg.Configuration{
			Sampler: &jaegercfg.SamplerConfig{
				Type:  jaeger.SamplerTypeConst,
				Param: 1,
			},
		}
		_, err := cfg.InitGlobalTracer(
			opts.serviceName,
			jaegercfg.Logger(jaegerlog.StdLogger),
			jaegercfg.Metrics(jaegermetrics.NullFactory),
		)
		if err != nil {
			log.Printf("Could not initialize jaeger tracer: %s", err.Error())
			return
		}
		trace.SpanURL = jaegerSpanURL
		return
	}

	if lightstepAccessToken != "" {
		log15.Info("Distributed tracing enabled", "tracer", "Lightstep")
		opentracing.InitGlobalTracer(lightstep.NewTracer(lightstep.Options{
			AccessToken: lightstepAccessToken,
			UseGRPC:     true,
			Tags: opentracing.Tags{
				lightstep.ComponentNameKey: opts.serviceName,
			},
			DropSpanLogs: !lightstepIncludeSensitive,
		}))
		trace.SpanURL = lightStepSpanURL

		// Ignore warnings from the tracer about SetTag calls with unrecognized value types. The
		// github.com/lightstep/lightstep-tracer-go package calls fmt.Sprintf("%#v", ...) on them, which is fine.
		defaultHandler := lightstep.NewEventLogOneError()
		lightstep.SetGlobalEventHandler(func(e lightstep.Event) {
			if _, ok := e.(lightstep.EventUnsupportedValue); ok {
				// ignore
			} else {
				defaultHandler(e)
			}
		})
	}
}

func lightStepSpanURL(span opentracing.Span) string {
	spanCtx := span.Context().(lightstep.SpanContext)
	t := span.(interface {
		Start() time.Time
	}).Start().UnixNano() / 1000
	return fmt.Sprintf("https://app.lightstep.com/%s/trace?span_guid=%x&at_micros=%d#span-%x", lightstepProject, spanCtx.SpanID, t, spanCtx.SpanID)
}

func jaegerSpanURL(span opentracing.Span) string {
	spanCtx := span.Context().(jaeger.SpanContext)
	return spanCtx.TraceID().String()
}
