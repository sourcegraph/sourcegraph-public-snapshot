package proxy

import (
	"context"
	"encoding/json"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/go-lsp/lspext"
	"github.com/sourcegraph/sourcegraph/pkg/rcache"
)

var cache = rcache.New("xlang")

var (
	cacheGetTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "xlang",
		Subsystem: "cache",
		Name:      "get_total",
		Help:      "Total number of gets for a mode.",
	}, []string{"mode", "type"})
	cacheSetTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "xlang",
		Subsystem: "cache",
		Name:      "set_total",
		Help:      "Total number of sets for a mode.",
	}, []string{"mode"})
	cacheGetTotalBytes = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "xlang",
		Subsystem: "cache",
		Name:      "get_total_bytes",
		Help:      "Total number of bytes fetched from the cache for a mode.",
	}, []string{"mode"})
	cacheSetTotalBytes = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "xlang",
		Subsystem: "cache",
		Name:      "set_total_bytes",
		Help:      "Total number of bytes set to the cache for a mode.",
	}, []string{"mode"})
)

func init() {
	prometheus.MustRegister(cacheGetTotal)
	prometheus.MustRegister(cacheSetTotal)
	prometheus.MustRegister(cacheGetTotalBytes)
	prometheus.MustRegister(cacheSetTotalBytes)
}

func (c *serverProxyConn) handleCacheGet(ctx context.Context, p lspext.CacheGetParams) (*json.RawMessage, error) {
	key := "xlang:" + c.id.mode + ":" + p.Key
	b, ok := cache.Get(key)
	if ok {
		cacheGetTotal.WithLabelValues(c.id.mode, "hit").Inc()
		cacheGetTotalBytes.WithLabelValues(c.id.mode).Add(float64(len(b)))
	} else {
		cacheGetTotal.WithLabelValues(c.id.mode, "miss").Inc()
		// cache miss requires a null return
		b = []byte("null")
	}
	m := json.RawMessage(b)
	return &m, nil
}

func (c *serverProxyConn) handleCacheSet(ctx context.Context, p lspext.CacheSetParams) {
	key := "xlang:" + c.id.mode + ":" + p.Key
	if p.Value == nil {
		// notification, so we can't return an error value
		return
	}
	b := []byte(*p.Value)
	cacheSetTotal.WithLabelValues(c.id.mode).Inc()
	cacheSetTotalBytes.WithLabelValues(c.id.mode).Add(float64(len(b)))
	cache.Set(key, b)
}
