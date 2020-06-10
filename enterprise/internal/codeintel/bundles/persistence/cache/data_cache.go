package cache

import "github.com/dgraph-io/ristretto"

// DataCache is a LRU cache that holds the results of values deserialized by a reader.
type DataCache interface {
	// Get returns the value (if any) and a boolean representing whether the value was
	// found or not.
	Get(key interface{}) (interface{}, bool)

	// Set attempts to add the key-value item to the cache with the given cost. If it
	// returns false, then the value as dropped and the item isn't added to the cache.
	Set(key, value interface{}, cost int64) bool
}

// NewDataCache creates a data cache instance with the given maximum capacity.
func NewDataCache(size int) (DataCache, error) {
	return ristretto.NewCache(&ristretto.Config{
		NumCounters: int64(size) * 10,
		MaxCost:     int64(size),
		BufferItems: 64,
	})
}
