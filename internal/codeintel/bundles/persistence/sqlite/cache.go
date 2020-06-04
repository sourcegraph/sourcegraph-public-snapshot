package sqlite

import "github.com/dgraph-io/ristretto"

type Cache interface {
	Get(key interface{}) (interface{}, bool)
	Set(key, value interface{}, cost int64) bool
}

func NewCache(size int64) (Cache, error) {
	return ristretto.NewCache(&ristretto.Config{
		NumCounters: size * 10,
		MaxCost:     size,
		BufferItems: 64,
		Metrics:     true,
	})
}
