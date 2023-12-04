package ttlcache

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sourcegraph/log"
)

// Cache is a cache that expires entries after a given expiration time.
type Cache[K comparable, V any] struct {
	reapOnce sync.Once // reapOnce ensures that the background reaper is only started once.

	reapContext    context.Context    // reapContext is the context used for the background reaper.
	reapCancelFunc context.CancelFunc // reapCancelFunc is the cancel function for reapContext.

	reapInterval time.Duration // reapInterval is the interval at which the cache will reap expired entries.
	ttl          time.Duration // ttl is the expiration duration for entries in the cache.

	newEntryFunc   func(K) V  // newEntryFunc is the routine that runs when a cache miss occurs.
	expirationFunc func(K, V) // expirationFunc is the callback to be called when an entry expires in the cache.

	logger log.Logger // logger is the logger used by the cache.

	sizeWarningThreshold uint // sizeWarningThreshold is the number of entries in the cache before a warning is logged.

	mu      sync.RWMutex
	entries map[K]*entry[V] // entries is the map of entries in the cache.

	clock clock // clock is the clock used to determine the current time.
}

type entry[V any] struct {
	lastUsed atomic.Pointer[time.Time]
	value    V
}

// New returns a new Cache with the provided newEntryFunc and options.
//
// newEntryFunc is the routine that runs when a cache miss occurs. The returned value is stored
// in the cache.
//
// By default, the cache will reap expired entries every minute and entries will
// expire after 10 minutes.
func New[K comparable, V any](newEntryFunc func(K) V, options ...Option[K, V]) *Cache[K, V] {
	ctx, cancel := context.WithCancel(context.Background())

	cache := Cache[K, V]{
		reapContext:    ctx,
		reapCancelFunc: cancel,

		reapInterval: 1 * time.Minute,
		ttl:          10 * time.Minute,

		newEntryFunc:   newEntryFunc,
		expirationFunc: func(k K, v V) {},

		logger: log.Scoped("ttlcache"),

		sizeWarningThreshold: 0,

		entries: make(map[K]*entry[V]),

		clock: productionClock{},
	}

	for _, option := range options {
		option(&cache)
	}

	return &cache
}

// Option is a function that configures a Cache.
type Option[K comparable, V any] func(*Cache[K, V])

// WithReapInterval sets the interval at which the cache will reap expired entries.
func WithReapInterval[K comparable, V any](interval time.Duration) Option[K, V] {
	return func(c *Cache[K, V]) {
		c.reapInterval = interval
	}
}

// WithTTL sets the expiration duration for entries in the cache.
//
// On each key access via Get(), the entry's expiration time is reset to now() + ttl.
//
// If the entry is not accessed before it expires, the reaper background goroutine will remove it from the cache.
func WithTTL[K comparable, V any](ttl time.Duration) Option[K, V] {
	return func(c *Cache[K, V]) {
		c.ttl = ttl
	}
}

// WithExpirationFunc sets the callback to be called when an entry expires.
func WithExpirationFunc[K comparable, V any](onExpiration func(K, V)) Option[K, V] {
	return func(c *Cache[K, V]) {
		c.expirationFunc = onExpiration
	}
}

// WithLogger sets the logger to be used by the cache.
func WithLogger[K comparable, V any](logger log.Logger) Option[K, V] {
	return func(c *Cache[K, V]) {
		c.logger = logger
	}
}

// WithSizeWarningThreshold sets the number of entries that can be in the cache before a warning is logged.
func WithSizeWarningThreshold[K comparable, V any](threshold uint) Option[K, V] {
	return func(c *Cache[K, V]) {
		c.sizeWarningThreshold = threshold
	}
}

// Get returns the value for the given key. If the key is not in the cache, it
// will be added using the newEntryFunc and returned to the caller.
func (c *Cache[K, V]) Get(key K) V {
	now := c.clock.Now()

	c.mu.RLock()

	// Fast path: check if the entry is already in the cache.
	e, ok := c.entries[key]
	if ok {
		e.lastUsed.Store(&now)
		value := e.value

		c.mu.RUnlock()
		return value
	}
	c.mu.RUnlock()

	// Slow path: lock the entire cache and check again.

	c.mu.Lock()
	defer c.mu.Unlock()

	// Did another goroutine already create the entry?
	e, ok = c.entries[key]
	if ok {
		e.lastUsed.Store(&now)
		return e.value
	}

	// Nobody created one, add a new one.
	e = &entry[V]{}
	e.lastUsed.Store(&now)
	e.value = c.newEntryFunc(key)

	c.entries[key] = e

	if c.sizeWarningThreshold > 0 && (len(c.entries) > int(c.sizeWarningThreshold)) {
		c.logger.Warn("cache is large", log.Int("size", len(c.entries)))
	}

	return e.value
}

// StartReaper starts the reaper goroutine. Every reapInterval, the reaper will
// remove entries that have not been accessed since now() - ttl.
//
// shutdown can be called to stop the reaper. After shutdown is called, the
// reaper will not be restarted.
func (c *Cache[K, V]) StartReaper() {
	c.reapOnce.Do(func() {
		c.logger.Info("starting reaper",
			log.Duration("reapInterval", c.reapInterval),
			log.Duration("ttl", c.ttl))

		go func() {
			ticker := time.NewTicker(c.reapInterval)
			defer ticker.Stop()

			for {
				select {
				case <-c.reapContext.Done():
					return
				case <-ticker.C:
					c.reap()
				}
			}
		}()
	})
}

// reap removes all entries that have not been accessed since ttl, and calls
// the expirationFunc for each entry that is removed.
func (c *Cache[K, V]) reap() {
	now := c.clock.Now()
	earliestAllowed := now.Add(-c.ttl)

	getExpiredEntries := func() map[K]V {
		expired := make(map[K]V)

		for key, entry := range c.entries {
			lastUsed := entry.lastUsed.Load()
			if lastUsed == nil {
				lastUsed = &time.Time{}
			}

			if lastUsed.Before(earliestAllowed) {
				expired[key] = entry.value
			}
		}

		return expired
	}

	// First, find all the entries that have expired.
	// We do this under a read lock to avoid blocking other goroutines
	// from accessing the cache.

	c.mu.RLock()
	possiblyExpired := getExpiredEntries()
	c.mu.RUnlock()

	// If there are no entries to delete, we're done.
	if len(possiblyExpired) == 0 {
		return
	}

	// If there are entries to delete, only now do we need to acquire
	// the write lock to delete them.

	c.mu.Lock()

	beforeLength := len(c.entries)

	// We need to check again to make sure that the entries are still
	// expired. It's possible that another goroutine has updated the
	// entries in between releasing an acquiring the locks.

	actuallyExpired := getExpiredEntries()

	// Go through the list of expired entries and delete them from the cache.
	for k := range actuallyExpired {
		delete(c.entries, k)
	}

	afterLength := len(c.entries)

	removedEntries := beforeLength - afterLength
	if removedEntries > 0 {
		c.logger.Debug("reaped entries",
			log.Int("removedEntries", removedEntries),
			log.Int("remainingEntries", afterLength))
	}

	c.mu.Unlock()

	// Call the expiration function for each entry that was deleted.
	// We do this outside of the lock to avoid blocking other goroutines
	// from accessing the cache.
	//
	// This is safe because these entries are no longer visible in the cache.

	for k, v := range actuallyExpired {
		c.expirationFunc(k, v)
	}
}

// Shutdown stops the background reaper. This function has no effect if the cache
// has already been shut down.
func (c *Cache[K, V]) Shutdown() {
	c.reapCancelFunc()
}

// clock is an interface for getting the current time. This is useful for testing.
type clock interface {
	Now() time.Time
}

type productionClock struct{}

func (productionClock) Now() time.Time {
	return time.Now()
}

var _ clock = productionClock{}
