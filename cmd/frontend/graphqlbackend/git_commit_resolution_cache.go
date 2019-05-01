package graphqlbackend

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type resolutionCacheEntry struct {
	t     time.Time
	value string
}

type resolutionCache struct {
	// ttl indicates how long before cache entries expire. There is no limit on
	// the size of the cache except the effective # of repositories on the
	// Sourcegraph instance.
	ttl time.Duration

	// cacheEntries, if non-nil, is used to record the number of entries in the cache.
	cacheEntries prometheus.Histogram

	// workerInterval, if non-zero, specifies the interval at which the worker
	// checks for entries to evict. Defaults ttl / 2.
	workerInterval time.Duration

	m sync.Map
}

func (r *resolutionCache) Set(k, v string) {
	r.m.Store(k, resolutionCacheEntry{
		t:     time.Now(),
		value: v,
	})
}

func (r *resolutionCache) Get(k string) (string, bool) {
	v, ok := r.m.Load(k)
	if !ok {
		return "", false
	}
	e := v.(resolutionCacheEntry)
	if time.Since(e.t) >= r.ttl {
		// entry has expired
		r.m.Delete(k)
		return "", false
	}
	return e.value, true
}

func (r *resolutionCache) startWorker() *resolutionCache {
	if r.workerInterval == 0 {
		r.workerInterval = r.ttl / 2
	}
	go func() {
		for {
			time.Sleep(r.workerInterval)
			size := 0
			r.m.Range(func(key, value interface{}) bool {
				size++
				e := value.(resolutionCacheEntry)
				if time.Since(e.t) >= r.ttl {
					// entry has expired
					r.m.Delete(key)
				}
				return true
			})
			r.cacheEntries.Observe(float64(size))
		}
	}()
	return r
}

var (
	// oidResolutionCache is used to cache Git commit OID resolution. This is
	// used because OID resolution happens extremely often (e.g. multiple times
	// per search result).
	oidResolutionCache = (&resolutionCache{
		ttl: 60 * time.Second,
		cacheEntries: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: "src",
			Subsystem: "graphql",
			Name:      "git_commit_oid_resolution_cache_entries",
			Help:      "Total number of entries in the in-memory Git commit OID resolution cache.",
		}),
	}).startWorker()

	oidResolutionCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "graphql",
		Name:      "git_commit_oid_resolution_cache_hit",
		Help:      "Counts cache hits and misses for Git commit OID resolution.",
	}, []string{"type"})

	oidResolutionDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "src",
		Subsystem: "graphql",
		Name:      "git_commit_oid_resolution_duration_seconds",
		Help:      "Total time spent performing uncached Git commit OID resolution.",
	})

	oidResolutionCacheLookupDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "src",
		Subsystem: "graphql",
		Name:      "git_commit_oid_resolution_cache_lookup_duration_seconds",
		Help:      "Total time spent performing cache lookups for Git commit OID resolution.",
	})
)

func init() {
	prometheus.MustRegister(oidResolutionCache.cacheEntries)
	prometheus.MustRegister(oidResolutionCounter)
	prometheus.MustRegister(oidResolutionDuration)
	prometheus.MustRegister(oidResolutionCacheLookupDuration)
}
