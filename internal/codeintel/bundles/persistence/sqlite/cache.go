package sqlite

import "github.com/dgraph-io/ristretto"

// Cache is a LFU cache that holds the results of values deserialized by the reader.
type Cache interface {
	// Get returns the value (if any) and a boolean representing whether the
	// value was found or not.
	Get(key interface{}) (interface{}, bool)

	// Set attempts to add the key-value item to the cache with the given cost. If it
	// returns false, then the value as dropped and the item isn't added to the cache.
	Set(key, value interface{}, cost int64) bool
}

// NewCache creates a cache instance with the given maximum capacity.
func NewCache(size int64) (Cache, error) {
	return ristretto.NewCache(&ristretto.Config{
		NumCounters: size * 10,
		MaxCost:     size,
		BufferItems: 64,
	})
}
