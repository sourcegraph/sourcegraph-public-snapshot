package github

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/httpcache"
)

// namespacedCache is a Cache wrapper that prepends namespace + ":" to
// all keys before invoking the corresponding underlying Cache's
// method.
//
// It is used to, for example, store cached items for multiple users
// separately to avoid leaking private information (the user's OAuth2
// token is the namespace).
type namespacedCache struct {
	namespace string
	httpcache.Cache
}

func (c namespacedCache) Get(key string) (responseBytes []byte, ok bool) {
	return c.Cache.Get(c.namespace + ":" + key)
}

func (c namespacedCache) Set(key string, responseBytes []byte) {
	c.Cache.Set(c.namespace+":"+key, responseBytes)
}

func (c namespacedCache) Delete(key string) {
	c.Cache.Delete(c.namespace + ":" + key)
}

// cacheWithMetrics tracks the number of cache hits and misses returned from an
// httpcache.Cache in prometheus.
type cacheWithMetrics struct {
	cache   httpcache.Cache
	counter *prometheus.CounterVec
}

func (c *cacheWithMetrics) Get(key string) ([]byte, bool) {
	resp, ok := c.cache.Get(key)
	if ok {
		c.counter.WithLabelValues("hit").Inc()
	} else {
		c.counter.WithLabelValues("miss").Inc()
	}
	return resp, ok
}

func (c *cacheWithMetrics) Set(key string, resp []byte) {
	c.cache.Set(key, resp)
}

func (c *cacheWithMetrics) Delete(key string) {
	c.cache.Delete(key)
}
