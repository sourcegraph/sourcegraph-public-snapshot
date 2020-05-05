package main

import (
	"github.com/dgraph-io/ristretto"
	"github.com/prometheus/client_golang/prometheus"
)

// registerer exists so we can override it in tests
var registerer = prometheus.DefaultRegisterer

// MustRegisterRistrettoMonitor exports three prometheus metrics
// "src_ristretto_cache_hits{cache=$cache}",
// "src_ristretto_cache_misses{cache=$cache}", and
// "src_ristretto_cache_cost{cache=$cache}".
//
// It is safe to call this function more than once for the same cache name.
func MustRegisterRistrettoMonitor(cacheName string, metrics *ristretto.Metrics) {
	mustRegisterOnce(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name:        "src_ristretto_cache_hits",
		Help:        "Total number of cache hits.",
		ConstLabels: prometheus.Labels{"cache": cacheName},
	}, func() float64 {
		return float64(metrics.Hits())
	}))

	mustRegisterOnce(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name:        "src_ristretto_cache_misses",
		Help:        "Total number of cache misses.",
		ConstLabels: prometheus.Labels{"cache": cacheName},
	}, func() float64 {
		return float64(metrics.Misses())
	}))

	mustRegisterOnce(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name:        "src_ristretto_cache_cost",
		Help:        "Current cost of the cache.",
		ConstLabels: prometheus.Labels{"cache": cacheName},
	}, func() float64 {
		return float64(metrics.CostAdded() - metrics.CostEvicted())
	}))
}

func mustRegisterOnce(c prometheus.Collector) {
	err := registerer.Register(c)
	if err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); ok {
			return
		}
		panic(err)
	}
}
