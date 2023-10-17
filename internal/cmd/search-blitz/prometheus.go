package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var Buckets = []float64{.01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 15, 30, 45, 60, 80, 100}

var durationSearchSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "search_blitz_duration_seconds",
	Help:    "e2e duration search-blitz where client is either stream or batch",
	Buckets: Buckets,
}, []string{"query_name", "client"})

var firstResultSearchSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "search_blitz_first_result_seconds",
	Help:    "e2e time to first result search-blitz where client is either stream or batch",
	Buckets: Buckets,
}, []string{"query_name", "client"})

var matchCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "search_blitz_match_count",
	Help: "the match count where client is either stream or batch",
}, []string{"query_name", "client"})
