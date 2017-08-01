package langserver

import (
	"os"
	"strconv"
	"sync"
	"sync/atomic"

	lru "github.com/hashicorp/golang-lru"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// typecheckCache is a process level cache for storing typechecked
	// values. Do not directly use this, instead use newTypecheckCache()
	typecheckCache = newLRU("SRC_TYPECHECK_CACHE_SIZE", 10)

	// symbolCache is a process level cache for storing symbols found. Do
	// not directly use this, instead use newSymbolCache()
	symbolCache = newLRU("SRC_SYMBOL_CACHE_SIZE", 500)

	// cacheID is used to prevent key conflicts between different
	// LangHandlers in the same process.
	cacheID int64

	typecheckCacheSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "golangserver",
		Subsystem: "typecheck",
		Name:      "cache_size",
		Help:      "Number of items in the typecheck cache",
	})
	typecheckCacheTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "golangserver",
		Subsystem: "typecheck",
		Name:      "cache_request_total",
		Help:      "Count of requests to cache.",
	}, []string{"type"})
	symbolCacheSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "golangserver",
		Subsystem: "symbol",
		Name:      "cache_size",
		Help:      "Number of items in the symbol cache",
	})
	symbolCacheTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "golangserver",
		Subsystem: "symbol",
		Name:      "cache_request_total",
		Help:      "Count of requests to cache.",
	}, []string{"type"})
)

func init() {
	prometheus.MustRegister(typecheckCacheSize)
	prometheus.MustRegister(typecheckCacheTotal)
	prometheus.MustRegister(symbolCacheSize)
	prometheus.MustRegister(symbolCacheTotal)
}

type cache interface {
	Get(key interface{}, fill func() interface{}) interface{}
	Purge()
}

func newTypecheckCache() *boundedCache {
	return &boundedCache{
		id:      nextCacheID(),
		c:       typecheckCache,
		size:    typecheckCacheSize,
		counter: typecheckCacheTotal,
	}
}

func newSymbolCache() *boundedCache {
	return &boundedCache{
		id:      nextCacheID(),
		c:       symbolCache,
		size:    symbolCacheSize,
		counter: symbolCacheTotal,
	}
}

type cacheKey struct {
	id int64
	k  interface{}
}

type cacheValue struct {
	ready chan struct{} // closed to broadcast readiness
	value interface{}
}

type boundedCache struct {
	mu      sync.Mutex
	id      int64
	c       *lru.Cache
	size    prometheus.Gauge
	counter *prometheus.CounterVec
}

func (c *boundedCache) Get(k interface{}, fill func() interface{}) interface{} {
	// c.c is already thread safe, but we need c.mu so we can insert a
	// cacheValue only once if we have a miss.
	c.mu.Lock()
	key := cacheKey{c.id, k}
	var v *cacheValue
	if vi, ok := c.c.Get(key); ok {
		// cache hit, wait until ready
		c.mu.Unlock()
		c.counter.WithLabelValues("hit").Inc()
		v = vi.(*cacheValue)
		<-v.ready
	} else {
		// cache miss. Add unready result to cache and fill
		v = &cacheValue{ready: make(chan struct{})}
		c.c.Add(key, v)
		c.mu.Unlock()
		c.size.Set(float64(c.c.Len()))
		c.counter.WithLabelValues("miss").Inc()

		defer close(v.ready)
		v.value = fill()
	}

	return v.value
}

func (c *boundedCache) Purge() {
	// c.c is a process level cache. We could increment c.id to make it seem
	// like we've purged the cache, but that would leave the objects in memory
	// (and typechecking objects can be several hundreds of MB). So instead,
	// we manually remove the objects from cache.
	c.mu.Lock()
	for _, key := range c.c.Keys() {
		if k := key.(cacheKey); k.id == c.id {
			c.c.Remove(key)
		}
	}
	c.mu.Unlock()
}

// newLRU returns an LRU based cache.
func newLRU(env string, defaultSize int) *lru.Cache {
	size := defaultSize
	if i, err := strconv.Atoi(os.Getenv(env)); err == nil && i > 0 {
		size = i
	}
	c, err := lru.New(size)
	if err != nil {
		// This should never happen since our size is always > 0
		panic(err)
	}
	return c
}

func nextCacheID() int64 {
	return atomic.AddInt64(&cacheID, 1)
}
