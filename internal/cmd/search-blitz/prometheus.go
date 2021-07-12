package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var Buckets = []float64{.01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 15, 30, 45, 60, 80, 100}

var durationSearchSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "search_blitz_duration_seconds",
	Help:    "e2e duration search-blitz where search_type is either stream or batch",
	Buckets: Buckets,
}, []string{"group", "query_name", "client"})

var firstResultSearchSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "search_blitz_first_result_seconds",
	Help:    "e2e time to first result search-blitz where search_type is either stream or batch",
	Buckets: Buckets,
}, []string{"group", "query_name", "client"})

var fetchDurationSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "search_blitz_trace_fetch_seconds",
	Help:    "The time taken to fetch a trace from Jaeger",
	Buckets: Buckets,
}, []string{"error", "attempts"})
