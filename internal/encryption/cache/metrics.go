package cache

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	hitTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "src_encryption_cache_hit_total",
			Help: "Total number of cache hits in encryption/cache",
		},
		[]string{},
	)
	missTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "src_encryption_cache_miss_total",
			Help: "Total number of cache misses in encryption/cache",
		},
		[]string{},
	)
	loadSuccessTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "src_encryption_cache_load_success_total",
			Help: "Total number of successful cache loads in encryption/cache",
		},
		[]string{},
	)
	loadErrorTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "src_encryption_cache_load_error_total",
			Help: "Total number of failed cache loads in encryption/cache",
		},
		[]string{},
	)
	evictTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "src_encryption_cache_eviction_total",
			Help: "Total number of cache evictions in encryption/cache",
		},
		[]string{},
	)
)
