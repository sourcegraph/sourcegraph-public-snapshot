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
	metricLabels    = []string{"mutation", "route", "success"}
	requestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "src_graphql_request_duration_seconds",
		Help:    "GraphQL request latencies in seconds.",
		Buckets: trace.UserLatencyBuckets,
	}, metricLabels)
)

func instrumentGraphQL(data traceData) {
	duration := time.Since(data.execStart)
	labels := prometheus.Labels{
		"route":    data.requestName,
		"success":  strconv.FormatBool(len(data.queryErrors) == 0),
		"mutation": strconv.FormatBool(strings.Contains(data.queryParams.Query, "mutation")),
	}
	requestDuration.With(labels).Observe(duration.Seconds())
}
