package github

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var githubRemainingGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	// _v2 since we have an older metric defined in github-proxy
	Name: "src_github_rate_limit_remaining_v2",
	Help: "Number of calls to GitHub's API remaining before hitting the rate limit.",
}, []string{"resource", "name"})

var githubRatelimitWaitCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "src_github_rate_limit_wait_duration_seconds",
	Help: "The amount of time spent waiting on the rate limit",
}, []string{"resource", "name"})
