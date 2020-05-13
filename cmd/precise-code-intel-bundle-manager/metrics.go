package main

import (
	"github.com/dgraph-io/ristretto"
	"github.com/prometheus/client_golang/prometheus"
)

// MustRegisterCacheMonitor emits metrics for a ristretto cache.
func MustRegisterCacheMonitor(r prometheus.Registerer, cacheName string, capacity int, metrics *ristretto.Metrics) {
	cacheCapacity := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name:        "src_cache_capacity",
		Help:        "Capacity of the cache.",
		ConstLabels: prometheus.Labels{"cache": cacheName},
	}, func() float64 {
		return float64(capacity)
	})
	r.MustRegister(cacheCapacity)

	cacheCost := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name:        "src_cache_cost",
		Help:        "Current cost of the cache.",
		ConstLabels: prometheus.Labels{"cache": cacheName},
	}, func() float64 {
		return float64(metrics.CostAdded() - metrics.CostEvicted())
	})
	r.MustRegister(cacheCost)

	cacheHits := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name:        "src_cache_hits_total",
		Help:        "Total number of cache hits.",
		ConstLabels: prometheus.Labels{"cache": cacheName},
	}, func() float64 {
		return float64(metrics.Hits())
	})
	r.MustRegister(cacheHits)

	cacheMisses := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name:        "src_cache_misses_total",
		Help:        "Total number of cache misses.",
		ConstLabels: prometheus.Labels{"cache": cacheName},
	}, func() float64 {
		return float64(metrics.Misses())
	})
	r.MustRegister(cacheMisses)
}
