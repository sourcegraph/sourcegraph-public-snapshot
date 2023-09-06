package backend

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	tracepkg "github.com/sourcegraph/sourcegraph/internal/trace"
)

var metricLabels = []string{"method", "success"}
var requestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "src_backend_client_request_duration_seconds",
	Help:    "Total time spent on backend endpoints.",
	Buckets: tracepkg.UserLatencyBuckets,
}, metricLabels)

var requestGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "src_backend_client_requests",
	Help: "Current number of requests running for a method.",
}, []string{"method"})

func startTrace(ctx context.Context, method string, arg any, err *error) (context.Context, func()) { //nolint:unparam // unparam complains that `server` always has same value across call-sites, but that's OK
	name := "Repos." + method
	requestGauge.WithLabelValues(name).Inc()

	tr, ctx := trace.New(ctx, name,
		attribute.String("argument", fmt.Sprintf("%#v", arg)),
		attribute.Int("userID", int(actor.FromContext(ctx).UID)),
	)
	start := time.Now()

	done := func() {
		elapsed := time.Since(start)
		labels := prometheus.Labels{
			"method":  name,
			"success": strconv.FormatBool(err == nil),
		}
		requestDuration.With(labels).Observe(elapsed.Seconds())
		requestGauge.WithLabelValues(name).Dec()
		tr.EndWithErr(err)
	}

	return ctx, done
}
