package cache

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence"
)

// Cache is a map of active reader instances by key. This implementation ensures that there is only one
// handle open to a given reader, and that the initialization of the reader (which may run migrations)
// is only attempted once at a time. Readers which are idle for longer than a configured duration are
// closed in the background.
type Cache struct {
	opener ReaderOpener

	m       sync.Mutex
	entries map[string]*cacheEntry
}

// cacheEntry wraps a reader which may still be initializing.
type cacheEntry struct {
	reader         persistence.Reader
	err            error
	init           chan struct{}  // marks init completion
	draining       chan struct{}  // marks intent to close
	accessedInTick bool           // marks used since last eviction
	refCount       sync.WaitGroup // acts as concurrent use gauge
	closeOnce      sync.Once      // guards closing the entry
}

// ReaderOpener initializes a new reader for the given key.
type ReaderOpener func(key string) (persistence.Reader, error)

// Handler performs an operation on a reader. The invocation of this function is a critical section that
// locks the given reader argument so that it is not closed while in use.
type Handler func(reader persistence.Reader) error

// ErrReaderInitializationDeadlineExceeded occurs when a new reader takes too long to initialize.
var ErrReaderInitializationDeadlineExceeded = errors.New("reader initialization deadline exceeded")

// NewReaderCache initializes a new reader cache with the given max reader idle time and reader opener.
func NewReaderCache(maxReaderIdleTime time.Duration, opener ReaderOpener) *Cache {
	return newReaderCache(time.NewTicker(maxReaderIdleTime).C, opener)
}

// newReaderCache initializes a new reader cache with the given ticker and reader opener.
func newReaderCache(ch <-chan time.Time, opener ReaderOpener) *Cache {
	cache := &Cache{
		opener:  opener,
		entries: map[string]*cacheEntry{},
	}

	go func() {
		for range ch {
			cache.evict()
		}
	}()

	return cache
}

// WithReader calls the given function with a reader argument. If the reader has not yet initialized and
// does not do so before context deadline, an error is returned. The reader initialization is not canceled
// due to a context deadline here and will continue to run in the background until completion.
func (c *Cache) WithReader(ctx context.Context, key string, fn Handler) error {
	if entry, ok := c.getOrCreateActiveEntry(ctx, key); ok {
		defer entry.refCount.Done()

		select {
		case <-entry.init:
			if entry.err != nil {
				return entry.err
			}
			return fn(entry.reader)

		case <-ctx.Done():
		}
	}

	return ErrReaderInitializationDeadlineExceeded
}

const (
	minBackoff           = time.Millisecond * 1
	maxBackoff           = time.Millisecond * 250
	backoffIncreaseRatio = 1.5
	maxAttempts          = 100
)

// getOrCreateActiveEntry gets or creates a cache entry for the given key. If there exists an entry for
// the key, but that entry is marked as draining, we retry following an exponential backoff algorithm.
func (c *Cache) getOrCreateActiveEntry(ctx context.Context, key string) (*cacheEntry, bool) {
	backoff := minBackoff

	for attempts := maxAttempts; attempts > 0; attempts-- {
		if entry, ok := c.getOrCreateRawEntry(key); ok {
			return entry, true
		}

		select {
		case <-time.After(backoff):
			if backoff = time.Duration(float64(backoff) * backoffIncreaseRatio); backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue

		case <-ctx.Done():
			break
		}
	}

	return nil, false
}

// getOrCreateRawEntry gets or creates a cache entry for the given key and a boolean flag indicating
// whether or not the entry is "active". True indicates that the entry is not currently marked as
// draining and can be used. In this case the entry's wait group value is not modified and the entry's
// close procedure is unaffected.
func (c *Cache) getOrCreateRawEntry(key string) (*cacheEntry, bool) {
	c.m.Lock()
	defer c.m.Unlock()

	entry, ok := c.entries[key]
	if !ok {
		entry = c.makeEntry(key)
		c.entries[key] = entry
	}

	// Mark as in-use
	entry.refCount.Add(1)
	entry.accessedInTick = true

	select {
	case <-entry.draining:
		// Caller won't invoke handler, so we decrease our use count immediately
		entry.refCount.Done()
		return nil, false
	default:
	}

	return entry, true
}

// makeEntry creates a new empty cache entry. This will be
func (c *Cache) makeEntry(key string) *cacheEntry {
	entry := &cacheEntry{
		init:     make(chan struct{}),
		draining: make(chan struct{}),
	}

	go func() {
		// Mark as initialized
		defer close(entry.init)

		if entry.reader, entry.err = c.opener(key); entry.err != nil {
			// Do not hold on to cache entries that failed to initialize. In the case of a
			// transient error (e.g., a missing bundle file that is later uploaded), we do
			// not want a poison cache value occupying that key until a process restart.
			c.m.Lock()
			delete(c.entries, key)
			c.m.Unlock()
		}
	}()

	return entry
}

// evict iterates through all entries and determines those which have not been accessed since the
// last eviction pass. These entries are marked for removal which will occur once initialization
// finishes and any active readers have detached. After this pass, the accessed flag is false for
// all entries.
func (c *Cache) evict() {
	c.m.Lock()
	defer c.m.Unlock()

	for key, entry := range c.entries {
		if !entry.accessedInTick {
			c.remove(key, entry)
		}

		entry.accessedInTick = false
	}
}

// remove marks a cache entry as draining so that future readers are blocked. Once the cache entry
// has initialized and all active readers have detached, the underlying reader is closed and the
// entry is removed from the cache map.
func (c *Cache) remove(key string, entry *cacheEntry) {
	entry.closeOnce.Do(func() {
		go func() {
			close(entry.draining) // Reject future readers
			<-entry.init          // Wait until initialized
			entry.refCount.Wait() // Wait until there are no more readers

			if entry.reader != nil {
				// Release resources
				if err := entry.reader.Close(); err != nil {
					log15.Error("Failed to close reader", "key", key, "err", err)
				}
			}

			c.m.Lock()
			delete(c.entries, key)
			c.m.Unlock()
		}()
	})
}
