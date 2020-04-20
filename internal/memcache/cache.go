package memcache

import (
	"container/list"
	"errors"
	"fmt"
	"sync"
)

// Cache is a generic in-memory LRU cache.
type Cache interface {
	// Size returns the current sum of all cache entry sizes.
	Size() int

	// GetOrCreate returns the value cached at the given key or creates a new
	// value and adds it to the cache.
	GetOrCreate(key interface{}, factory ValueFactory) (interface{}, error)
}

// ValueFactory constructs a value and returns its weight within the cache.
type ValueFactory func() (interface{}, int, error)

// EvictCallback is invoked when an entry is pushed from the cache.
type EvictCallback func(key interface{}, value interface{})

// cache implements the Cache interface. This implementation is based off of
// github.com/hashicorp/golang-lru, but allows cache values to report their
// own size instead of using the length of evictList as the sole heuristic.
type cache struct {
	size      int                           // sum of all cache entry sizes
	maxSize   int                           // upper bound on the sum of all cache entry sizes
	cacheMu   sync.Mutex                    // protcts evictList and items
	evictList *list.List                    // list of cacheEntires ordered by recency
	items     map[interface{}]*list.Element // map from key to the node in the evictList
	onEvict   EvictCallback
}

type cacheEntry struct {
	key   interface{}
	value interface{}
	size  int
}

// New creates a new cache bounded by maxSize.
func New(maxSize int) (Cache, error) {
	return NewWithEvict(maxSize, nil)
}

// NewWithEvict creates a new cache bounded by maxSize with the given eviction callback.
func NewWithEvict(maxSize int, onEvict EvictCallback) (Cache, error) {
	if maxSize <= 0 {
		return nil, errors.New("must provide a positive size")
	}

	return &cache{
		maxSize:   maxSize,
		evictList: list.New(),
		items:     map[interface{}]*list.Element{},
		onEvict:   onEvict,
	}, nil
}

// Size returns the current sum of all cache entry sizes.
func (c *cache) Size() int {
	return c.size
}

// GetOrCreate returns the value cached at the given key or creates a new value and adds it
// to the cache. This method is goroutine-safe.
func (c *cache) GetOrCreate(key interface{}, factory ValueFactory) (interface{}, error) {
	value, exists := c.get(key)
	if exists {
		return value, nil
	}

	value, size, err := factory()
	if err != nil {
		return nil, err
	}

	if size <= 0 {
		c.evict(key, value)
		return nil, fmt.Errorf("must provide a positive cache entry size")
	}

	c.add(key, value, size)
	return value, nil
}

// get returns the value stored at the given key.
func (c *cache) get(key interface{}) (interface{}, bool) {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()

	entry, exists := c.items[key]
	if !exists {
		// cache miss
		return nil, false
	}

	// cache hit, update recency data
	kv := entry.Value.(*cacheEntry)
	c.evictList.MoveToFront(entry)
	return kv.value, true
}

// add inserts a key/value pair into the cache.
func (c *cache) add(key, value interface{}, size int) {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()

	if entry, exists := c.items[key]; exists {
		kv := entry.Value.(*cacheEntry)
		oldValue, oldSize := kv.value, kv.size
		kv.value, kv.size = value, size
		c.size += (size - oldSize)
		c.evictList.MoveToFront(entry)
		c.evict(key, oldValue)
		return
	}

	kv := &cacheEntry{key: key, value: value, size: size}
	entry := c.evictList.PushFront(kv)
	c.items[key] = entry
	c.size += size

	for c.size > c.maxSize {
		// Get the least recently used cache entry. This value is
		// guaranteed to exist as we can't empty the evictList without
		// getting back to a zero size, which is strictly less than
		// any valid value of the cache max size.
		entry := c.evictList.Back()
		kv := entry.Value.(*cacheEntry)
		c.evictList.Remove(entry)
		delete(c.items, kv.key)
		c.size -= kv.size
		c.evict(kv.key, kv.value)
	}
}

// evict invokes the evict callback, if set, on key and value.
func (c *cache) evict(key, value interface{}) {
	if c.onEvict != nil {
		c.onEvict(key, value)
	}
}
