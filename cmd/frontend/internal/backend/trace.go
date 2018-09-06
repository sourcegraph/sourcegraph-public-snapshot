package backend

import (
	"context"
	"fmt"
	"strconv"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/sourcegraph/sourcegraph/pkg/actor"
	tracepkg "github.com/sourcegraph/sourcegraph/pkg/trace"
)

var metricLabels = []string{"method", "success"}
var requestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "src",
	Subsystem: "backend",
	Name:      "client_request_duration_seconds",
	Help:      "Total time spent on backend endpoints.",
	Buckets:   tracepkg.UserLatencyBuckets,
}, metricLabels)
var requestGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "src",
	Subsystem: "backend",
	Name:      "client_requests",
	Help:      "Current number of requests running for a method.",
}, []string{"method"})

func init() {
	prometheus.MustRegister(requestDuration)
	prometheus.MustRegister(requestGauge)
}

func trace(ctx context.Context, server, method string, arg interface{}, err *error) (context.Context, func()) {
	requestGauge.WithLabelValues(server + "." + method).Inc()

	span, ctx := opentracing.StartSpanFromContext(ctx, server+"."+method)
	span.SetTag("Server", server)
	span.SetTag("Method", method)
	span.SetTag("Argument", fmt.Sprintf("%#v", arg))
	start := time.Now()

	done := func() {
		elapsed := time.Since(start)

		if err != nil && *err != nil {
			span.SetTag("Error", (*err).Error())
		}
		span.Finish()

		name := server + "." + method
		labels := prometheus.Labels{
			"method":  name,
			"success": strconv.FormatBool(err == nil),
		}
		requestDuration.With(labels).Observe(elapsed.Seconds())
		requestGauge.WithLabelValues(name).Dec()

		uid := actor.FromContext(ctx).UID
		errStr := ""
		if err != nil && *err != nil {
			errStr = (*err).Error()
		}
		log15.Debug("TRACE backend", "rpc", name, "uid", uid, "trace", tracepkg.SpanURL(span), "error", errStr, "duration", elapsed)
	}

	return ctx, done
}
