package main

import "github.com/prometheus/client_golang/prometheus"

var UserLatencyBuckets = []float64{200, 500, 1000, 2000, 5000, 10000, 30000}

var durationSearchHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "duration",
	Help:    "e2e duration",
	Buckets: UserLatencyBuckets,
}, []string{"type"})

func init() {
	prometheus.MustRegister(durationSearchHistogram)
}
