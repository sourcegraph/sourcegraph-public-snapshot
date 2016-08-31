package langp

import (
	"context"
	"os"
	"time"

	lightstep "github.com/lightstep/lightstep-tracer-go"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/statsutil"
)

var (
	prepDuration, reqDuration, reqAndWorkspaceDuration *prometheus.HistogramVec
	prepHeartbeat, reqHeartbeat                        *prometheus.GaugeVec
)

// InitMetrics initializes prometheus metrics for the given language which must
// be a lowercase prometheus-compatible name (e.g. "go" or "java").
func InitMetrics(language string) {
	if t := os.Getenv("LIGHTSTEP_ACCESS_TOKEN"); t != "" {
		opentracing.InitGlobalTracer(lightstep.NewTracer(lightstep.Options{
			AccessToken: t,
		}))
	}

	namespace := "lang"
	// Workspace preparation metrics.
	prepLabels := []string{"type", "method", "repo", "status"}
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
	}, []string{"method", "repo", "status"})
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
	reqAndWorkspaceDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: language,
		Name:      "request_and_workspace_duration_seconds",
		Help:      "The HTTP request latencies, including workspace preparation, in seconds.",
		Buckets:   statsutil.UserLatencyBuckets,
	}, reqLabels)
	reqHeartbeat = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: language,
		Name:      "requests_last_timestamp_unixtime",
		Help:      "Last time a request finished for a http endpoint.",
	}, reqLabels)
	prometheus.MustRegister(reqDuration)
	prometheus.MustRegister(reqAndWorkspaceDuration)
	prometheus.MustRegister(reqHeartbeat)
}

const (
	// prepStatusNoWork signals that no work was needed to prepare the
	// workspace (i.e. it was already prepared).
	prepStatusNoWork = "no-work"

	// prepStatusTimeout signals that workspace preparation was already under
	// way, and it took longer than our request was willing to wait, thus our
	// request timed out.
	prepStatusTimeout = "timeout"

	// prepStatusWaiting signals that workspace preparation was already under
	// way, and eventually finished (our request spent the whole time waiting).
	prepStatusWaiting = "waiting"

	// prepStatusOK signals that our request triggered workspace preparation
	// and finished it successfully.
	prepStatusOK = "ok"

	// prepStatusError signals that our request failed in some way to prepare
	// the workspace.
	prepStatusError = "error"
)

const (
	// prepTypeRepo specifies that the preparation was for cloning/updating the
	// repository.
	prepTypeRepo = "repo"

	// prepTypeDeps specifies that the preparation was for fetching/updating
	// the workspace dependencies.
	prepTypeDeps = "deps"
)

// observePrepareRepo observes repository preparation as having began at the
// given start time (and finished now) for the given repository. The status
// parameter should be one of the prepStatus constants.
func observePrepareRepo(ctx context.Context, start time.Time, repo, status string) {
	method := ctx.Value(methodNameKey).(string)
	prepDuration.WithLabelValues("repo", method, repo, status).Observe(time.Since(start).Seconds())
	prepHeartbeat.WithLabelValues(method, repo, status).Set(float64(time.Now().Unix()))
}

// observePrepareDeps observes dependency preparation as having began at the given
// start time (and finished now) for the given repository.
func observePrepareDeps(ctx context.Context, start time.Time, repo, status string) {
	method := ctx.Value(methodNameKey).(string)
	prepDuration.WithLabelValues("deps", method, repo, status).Observe(time.Since(start).Seconds())
}

// observe observes a language processor request, recording information about it to
// prometheus.
//
// err should be the error which the LSP server responded with, if any.
//
// unresolved should be true if the request to the LSP server did not yield any
// results (i.e. an error did not occur, but no results were found).
func observe(ctx context.Context, start, workspaceStart time.Time, repo string, err error, unresolved bool) {
	method := ctx.Value(methodNameKey).(string)
	status := "ok"
	if err != nil {
		status = "error"
	} else if unresolved {
		status = "unresolved"
	}
	reqAndWorkspaceDuration.WithLabelValues(method, repo, status).Observe(time.Since(workspaceStart).Seconds())
	reqDuration.WithLabelValues(method, repo, status).Observe(time.Since(start).Seconds())
}
