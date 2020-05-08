package main

import (
	"github.com/dgraph-io/ristretto"
	"github.com/prometheus/client_golang/prometheus"
)

// MustRegisterCacheMonitor emits metrics for a ristretto cache.
func MustRegisterCacheMonitor(r prometheus.Registerer, cacheName string, metrics *ristretto.Metrics) {
	cacheCost := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name:        "src_cache_cost",
		Help:        "Current cost of the cache.",
		ConstLabels: prometheus.Labels{"cache": cacheName},
	}, func() float64 {
		return float64(metrics.CostAdded() - metrics.CostEvicted())
	})

	cacheHits := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name:        "src_cache_hits",
		Help:        "Total number of cache hits.",
		ConstLabels: prometheus.Labels{"cache": cacheName},
	}, func() float64 {
		return float64(metrics.Hits())
	})

	cacheMisses := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name:        "src_cache_misses",
		Help:        "Total number of cache misses.",
		ConstLabels: prometheus.Labels{"cache": cacheName},
	}, func() float64 {
		return float64(metrics.Misses())
	})

	r.MustRegister(cacheCost)
	r.MustRegister(cacheHits)
	r.MustRegister(cacheMisses)
}
