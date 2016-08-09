package langp

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/statsutil"
)

var (
	prepDuration, reqDuration   *prometheus.HistogramVec
	prepHeartbeat, reqHeartbeat *prometheus.GaugeVec
)

// InitMetrics initializes prometheus metrics for the given language which must
// be a lowercase prometheus-compatible name (e.g. "go" or "java").
func InitMetrics(language string) {
	namespace := "lang"
	// Workspace preparation metrics.
	prepLabels := []string{"repo", "status"}
	prepDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: language,
		Name:      "workspace_prep_seconds",
		Help:      language + " workspace preparation latencies in seconds.",
		Buckets:   statsutil.UserLatencyBuckets,
	}, prepLabels)
	prepHeartbeat = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: language,
		Name:      "workspace_prep_last_timestamp_unixtime",
		Help:      "last time a " + language + " workspace was prepared.",
	}, prepLabels)
	prometheus.MustRegister(prepDuration)
	prometheus.MustRegister(prepHeartbeat)

	// HTTP request metrics.
	reqLabels := []string{"method", "repo", "status"}
	reqDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: language,
		Name:      "request_duration_seconds",
		Help:      "The HTTP request latencies in seconds.",
		Buckets:   statsutil.UserLatencyBuckets,
	}, reqLabels)
	reqHeartbeat = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: language,
		Name:      "requests_last_timestamp_unixtime",
		Help:      "Last time a request finished for a http endpoint.",
	}, reqLabels)
	prometheus.MustRegister(reqDuration)
	prometheus.MustRegister(reqHeartbeat)
}

const (
	// prepStatusTimeout signals that workspace preparation was already under
	// way, and it took longer than our request was willing to wait, thus our
	// request timed out.
	prepStatusTimeout = "timeout"

	// prepStatusWaiting signals that workspace preparation was already under
	// way, and eventually finished (our request spent the whole time waiting).
	prepStatusWaiting = "waiting"

	// prepStatusOK signals that our request triggered workspace preparation
	// and finished it.
	prepStatusOK = "ok"
)

// observePrepare observes repository preparation as having began at the given
// start time (and finished now) for the given repository. The status parameter
// should be one of the prepStatus constants.
func observePrepare(start time.Time, repo, status string) {
	prepDuration.WithLabelValues(repo, status).Observe(time.Since(start).Seconds())
	prepHeartbeat.WithLabelValues(repo, status).Set(float64(time.Now().Unix()))
}

// observe observes an LSP request, recording information about it to
// prometheus.
//
// err should be the error which the LSP server responded with, if any.
//
// unresolved should be true if the request to the LSP server did not yield any
// results (i.e. an error did not occur, but no results were found).
func observe(start time.Time, method, repo string, err error, unresolved bool) {
	d := time.Since(start).Seconds()
	status := "ok"
	if err != nil {
		status = "error"
	} else if unresolved {
		status = "unresolved"
	}
	reqDuration.WithLabelValues(method, repo, status).Observe(d)
}
