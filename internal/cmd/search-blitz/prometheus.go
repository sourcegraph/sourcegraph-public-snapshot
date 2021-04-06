package main

import "github.com/prometheus/client_golang/prometheus"

var UserLatencyBuckets = []float64{100, 200, 300, 400, 500, 1000, 2000, 5000, 10000, 30000}

var durationSearchHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "search_blitz_duration_ms",
	Help:    "e2e duration search-blitz where search_type is either stream or batch",
	Buckets: UserLatencyBuckets,
}, []string{"type", "search_type"})

func init() {
	prometheus.MustRegister(durationSearchHistogram)
}
