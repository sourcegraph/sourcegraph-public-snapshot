package httpapi

import (
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/sourcegraph/internal/trace"
)

var (
	metricLabels    = []string{"mutation", "route", "success", "source"}
	requestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "src_graphql_request_duration_seconds",
		Help:    "GraphQL request latencies in seconds.",
		Buckets: trace.UserLatencyBuckets,
	}, metricLabels)
)

func instrumentGraphQL(data traceData) {
	labels := prometheus.Labels{
		"route":    data.requestName,
		"source":   data.requestSource,
		"success":  strconv.FormatBool(len(data.queryErrors) == 0),
		"mutation": strconv.FormatBool(strings.Contains(data.queryParams.Query, "mutation")),
	}
	duration := time.Since(data.execStart)
	requestDuration.With(labels).Observe(duration.Seconds())
}
