package github

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
)

var githubRemainingGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	// TODO: New name?!
	Name: "TODO_src_github_rate_limit_remaining_v2",
	Help: "Number of calls to GitHub's API remaining before hitting the rate limit.",
}, []string{"resource", "name"})

var githubRatelimitWaitCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	// TODO: New name?!
	Name: "TODO_src_github_rate_limit_wait_duration_seconds",
	Help: "The amount of time spent waiting on the rate limit",
}, []string{"resource", "name"})

func collectRateLimitMonitorMetrics(monitor *ratelimit.Monitor, resource string) {
	monitor.SetCollector(&ratelimit.MetricsCollector{
		Remaining: func(n float64) {
			githubRemainingGauge.WithLabelValues(resource).Set(n)
		},
		WaitDuration: func(n time.Duration) {
			githubRatelimitWaitCounter.WithLabelValues(resource).Add(n.Seconds())
		},
	})
}
