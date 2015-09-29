// Package trace provides functions that allows method calls
// to be traced (using, e.g., Appdash).
package trace

import (
	"fmt"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"src.sourcegraph.com/sourcegraph/vendored/github.com/resonancelabs/go-pub/instrument"
	tg_context "src.sourcegraph.com/sourcegraph/vendored/github.com/resonancelabs/go-pub/instrument/context"

	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/appdash"
	authpkg "src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/util/traceutil"
)

// Before is called before a method executes and is passed the server
// and method name and the argument. The returned context is passed
// when invoking the underlying method.
func Before(ctx context.Context, server, method string, arg interface{}) context.Context {
	spanID := traceutil.SpanIDFromContext(ctx)
	if spanID == (appdash.SpanID{}) {
		spanID = appdash.NewRootSpanID()
	} else {
		spanID = appdash.NewSpanID(spanID)
	}
	log15.Debug("gRPC "+server+"."+method+" before", "spanID", spanID)
	ctx = traceutil.NewContext(ctx, spanID)

	// Traceguide instrumentation
	ctx, span := tg_context.StartSpan(ctx)
	span.SetName(fmt.Sprintf("%s/%s", server, method))
	if arg != nil {
		span.Log(instrument.Printf("%s arg", method).Payload(arg))
	}
	span.Log(instrument.EventName("appdash_span_id").Payload(spanID))
	span.AddTraceJoinId("appdash_trace_id", spanID.Trace)

	return ctx
}

var metricLabels = []string{"method", "success"}
var requestCount = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: "src",
	Subsystem: "grpc",
	Name:      "client_requests_total",
	Help:      "Total number of requests sent to grpc endpoints.",
}, metricLabels)
var requestDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
	Namespace: "src",
	Subsystem: "grpc",
	Name:      "client_request_duration_nanoseconds",
	Help:      "Total time spent on grpc endpoints.",
}, metricLabels)
var requestHeartbeat = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "src",
	Subsystem: "grpc",
	Name:      "client_requests_last_timestamp_unixtime",
	Help:      "Last time a request finished for a grpc endpoint.",
}, metricLabels)

var userMetricLabels = []string{"uid", "service"}
var requestPerUser = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: "src",
	Subsystem: "grpc",
	Name:      "client_requests_per_user",
	Help:      "Total number of requests per user id.",
}, userMetricLabels)

func init() {
	prometheus.MustRegister(requestCount)
	prometheus.MustRegister(requestDuration)
	prometheus.MustRegister(requestHeartbeat)
	prometheus.MustRegister(requestPerUser)
}

// After is called after a method executes and is passed the elapsed
// execution time since the method's BeforeFunc was called and the
// error returned, if any.
func After(ctx context.Context, server, method string, arg interface{}, err error, elapsed time.Duration) {
	tg_context.FinishSpan(ctx)

	elapsed += time.Millisecond // HACK: make everything show up in the chart
	sr := time.Now().Add(-1 * elapsed)
	call := &traceutil.GRPCCall{
		Server:     server,
		Method:     method,
		Arg:        fmt.Sprintf("%#v", arg),
		ArgType:    fmt.Sprintf("%T", arg),
		ServerRecv: sr,
		ServerSend: time.Now(),
		Err:        fmt.Sprintf("%#v", err),
	}
	rec := traceutil.Recorder(ctx)
	rec.Name(server + "." + method)
	rec.Event(call)
	// TODO measure metrics on the server, rather than the client
	labels := prometheus.Labels{
		"method":  server + "." + method,
		"success": strconv.FormatBool(err == nil),
	}
	requestCount.With(labels).Inc()
	requestDuration.With(labels).Observe(float64(elapsed.Nanoseconds()))
	requestHeartbeat.With(labels).Set(float64(time.Now().Unix()))

	labels = prometheus.Labels{
		"uid":     strconv.Itoa(authpkg.ActorFromContext(ctx).UID),
		"service": server,
	}
	requestPerUser.With(labels).Inc()
	log15.Debug("gRPC "+server+"."+method+" after", "spanID", traceutil.SpanIDFromContext(ctx), "elapsed", elapsed)
}
