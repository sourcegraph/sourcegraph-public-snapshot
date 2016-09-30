// Package trace provides functions that allows method calls
// to be traced (using, e.g., Appdash).
package trace

import (
	"context"
	"fmt"
	"strconv"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	log15 "gopkg.in/inconshreveable/log15.v2"

	authpkg "sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/statsutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
)

// Before is called before a method executes and is passed the server
// and method name and the argument. The returned context is passed
// when invoking the underlying method.
func Before(ctx context.Context, server, method string, arg interface{}) context.Context {
	requestGauge.WithLabelValues(server + "." + method).Inc()

	_, ctx = opentracing.StartSpanFromContext(ctx, server+"."+method)
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
	span := opentracing.SpanFromContext(ctx)
	span.SetTag("Server", server)
	span.SetTag("Method", method)
	span.SetTag("Argument", fmt.Sprintf("%#v", arg))
	if err != nil {
		span.SetTag("Error", err.Error())
	}
	span.Finish()

	name := server + "." + method
	labels := prometheus.Labels{
		"method":  name,
		"success": strconv.FormatBool(err == nil),
	}
	requestCount.With(labels).Inc()
	requestDuration.With(labels).Observe(elapsed.Seconds())
	requestGauge.WithLabelValues(name).Dec()

	uid := authpkg.ActorFromContext(ctx).UID
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}
	log15.Debug("TRACE gRPC", "rpc", name, "uid", uid, "trace", traceutil.SpanURL(span), "error", errStr, "duration", elapsed)
}
