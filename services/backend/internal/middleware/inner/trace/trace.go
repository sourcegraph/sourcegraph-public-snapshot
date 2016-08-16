// Package trace provides functions that allows method calls
// to be traced (using, e.g., Appdash).
package trace

import (
	"fmt"
	"strconv"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/prometheus/client_golang/prometheus"

	"context"

	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	authpkg "sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/statsutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
)

// prepareArg prepares the gRPC method arg for logging/tracing. For
// example, it does not log/trace arg if it is a very long byte slice
// (as it often is for git transport ops).
func prepareArg(server, method string, arg interface{}) interface{} {
	switch arg := arg.(type) {
	case *sourcegraph.ReceivePackOp:
		return &sourcegraph.ReceivePackOp{Repo: arg.Repo, Data: []byte("OMITTED"), AdvertiseRefs: arg.AdvertiseRefs}
	case *sourcegraph.UploadPackOp:
		return &sourcegraph.UploadPackOp{Repo: arg.Repo, Data: []byte("OMITTED"), AdvertiseRefs: arg.AdvertiseRefs}
	}
	return arg
}

// Before is called before a method executes and is passed the server
// and method name and the argument. The returned context is passed
// when invoking the underlying method.
func Before(ctx context.Context, server, method string, arg interface{}) context.Context {
	parent := traceutil.SpanIDFromContext(ctx)
	var spanID appdash.SpanID
	if parent == (appdash.SpanID{}) {
		spanID = appdash.NewRootSpanID()
	} else {
		spanID = appdash.NewSpanID(parent)
	}
	ctx = traceutil.NewContext(ctx, spanID)
	requestGauge.WithLabelValues(server + "." + method).Inc()
	return ctx
}

var metricLabels = []string{"method", "success"}
var requestCount = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: "src",
	Subsystem: "grpc",
	Name:      "client_requests_total",
	Help:      "Total number of requests sent to grpc endpoints.",
}, metricLabels)
var requestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "src",
	Subsystem: "grpc",
	Name:      "client_request_duration_seconds",
	Help:      "Total time spent on grpc endpoints.",
	Buckets:   statsutil.UserLatencyBuckets,
}, metricLabels)
var requestGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "src",
	Subsystem: "grpc",
	Name:      "client_requests",
	Help:      "Current number of requests running for a method.",
}, []string{"method"})

func init() {
	prometheus.MustRegister(requestCount)
	prometheus.MustRegister(requestDuration)
	prometheus.MustRegister(requestGauge)
}

// After is called after a method executes and is passed the elapsed
// execution time since the method's BeforeFunc was called and the
// error returned, if any.
func After(ctx context.Context, server, method string, arg interface{}, err error, elapsed time.Duration) {
	sr := time.Now().Add(-1 * elapsed)
	// HACK: make everything show up in the chart
	if elapsed < time.Millisecond {
		sr = time.Now().Add(-time.Millisecond)
	}
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}
	name := server + "." + method
	call := &traceutil.GRPCCall{
		Server:     server,
		Method:     method,
		Arg:        fmt.Sprintf("%#v", prepareArg(server, method, arg)),
		ArgType:    fmt.Sprintf("%T", arg),
		ServerRecv: sr,
		ServerSend: time.Now(),
		Err:        errStr,
	}
	rec := traceutil.Recorder(ctx)
	rec.Name(server + "." + method)
	rec.Event(call)

	labels := prometheus.Labels{
		"method":  name,
		"success": strconv.FormatBool(err == nil),
	}
	requestCount.With(labels).Inc()
	requestDuration.With(labels).Observe(elapsed.Seconds())
	requestGauge.WithLabelValues(name).Dec()

	uid := strconv.Itoa(authpkg.ActorFromContext(ctx).UID)
	log15.Debug("TRACE gRPC", "rpc", name, "uid", uid, "spanID", traceutil.SpanIDFromContext(ctx), "error", errStr, "duration", elapsed)
}
