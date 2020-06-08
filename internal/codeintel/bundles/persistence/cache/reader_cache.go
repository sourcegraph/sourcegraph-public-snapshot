package cache

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence"
)

// Cache is a map of active reader instances by key. This implementation ensures that there is only one
// handle open to a given reader, and that the initialization of the reader (which may run migrations)
// is only attempted once at a time. Readers which are idle for longer than a configured duration are
// closed in the background.
type Cache struct {
	ttl    time.Duration
	opener ReaderOpener

	m       sync.Mutex
	entries map[string]*cacheEntry
}

// cacheEntry wraps a reader which may still be initializing.
type cacheEntry struct {
	reader          persistence.Reader
	err             error
	init            chan struct{} // signals init completion
	drained         chan struct{} // signals complete removal from cache
	removeOnce      sync.Once     // guards entry remove procedure
	refCount        uint32        // number of active readers
	refCountChanged *sync.Cond    // broadcast when refCount decreases

	// expiry denotes the time when the entry can be removed from the cache.
	//
	// 🔐 - the cache mutex must be held when reading or modifying this value.
	expiry time.Time

	// draining marks the intent to close this entry. No additional readers
	// of this entry should be allowed after this value is set to true.
	//
	// 🔐 - the cache mutex must be held when reading or modifying this value.
	draining bool
}

// ReaderOpener initializes a new reader for the given key.
type ReaderOpener func(key string) (persistence.Reader, error)

// Handler performs an operation on a reader. The invocation of this function is a critical section that
// locks the given reader argument so that it is not closed while in use.
type Handler func(reader persistence.Reader) error

// ErrReaderInitializationDeadlineExceeded occurs when a new reader takes too long to initialize.
var ErrReaderInitializationDeadlineExceeded = errors.New("reader initialization deadline exceeded")

// NewReaderCache initializes a new reader cache with the given TTL and reader opener.
func NewReaderCache(ttl time.Duration, opener ReaderOpener) *Cache {
	return newReaderCache(ttl, time.NewTicker(ttl).C, opener)
}

// newReaderCache initializes a new reader cache with the given ticker and reader opener.
func newReaderCache(ttl time.Duration, ch <-chan time.Time, opener ReaderOpener) *Cache {
	cache := &Cache{
		ttl:     ttl,
		opener:  opener,
		entries: map[string]*cacheEntry{},
	}

	go func() {
		for now := range ch {
			cache.evict(now.UTC())
		}
	}()

	return cache
}

// WithReader calls the given function with a reader argument. If the reader has not yet initialized and
// does not do so before context deadline, an error is returned. The reader initialization is not canceled
// due to a context deadline here and will continue to run in the background until completion.
func (c *Cache) WithReader(ctx context.Context, key string, fn Handler) error {
	if entry, ok := c.getOrCreateActiveEntry(ctx, key); ok {
		defer func() {
			// Decrement refcount
			if atomic.AddUint32(&entry.refCount, ^uint32(0)) == 0 {
				// Wake routines waiting for this value to go to zero
				entry.refCountChanged.Broadcast()

				// No more readers, update expiry
				entry.expiry = time.Now().UTC().Add(c.ttl)
			}
		}()

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

// getOrCreateActiveEntry gets or creates a cache entry for the given key. If there exists an entry for
// the key, but that entry is marked as draining, we wait until that entry has completely drained and then
// try again.
func (c *Cache) getOrCreateActiveEntry(ctx context.Context, key string) (*cacheEntry, bool) {
loop:
	for {
		entry, ok := c.getOrCreateRawEntry(key)
		if ok {
			return entry, true
		}

		select {
		case <-entry.drained:
			continue loop
		case <-ctx.Done():
			return nil, false
		}
	}
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

	if !entry.draining {
		// Mark as in-use and disable expiry while held
		atomic.AddUint32(&entry.refCount, 1)
		entry.expiry = time.Time{}
	}

	return entry, !entry.draining
}

// makeEntry creates a new empty cache entry. This will be
func (c *Cache) makeEntry(key string) *cacheEntry {
	entry := &cacheEntry{
		init:            make(chan struct{}),
		drained:         make(chan struct{}),
		refCountChanged: sync.NewCond(&sync.Mutex{}),
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

// evict iterates through all entries and determines those which have expired. These entries are marked
// for removal which will occur once initialization finishes and any active readers have detached. Generally,
// no readers should be active, but may occur when a request around the expiry races.
func (c *Cache) evict(now time.Time) {
	c.m.Lock()
	defer c.m.Unlock()

	for key, entry := range c.entries {
		if !entry.expiry.IsZero() && entry.expiry.Before(now) {
			c.remove(key, entry)
		}
	}
}

// remove marks a cache entry as draining so that future readers are blocked. Once the cache entry
// has initialized and all active readers have detached, the underlying reader is closed and the
// entry is removed from the cache map.
func (c *Cache) remove(key string, entry *cacheEntry) {
	entry.removeOnce.Do(func() {
		go func() {
			// Release any readers waiting on this entry to drain before creating a new one
			defer close(entry.drained)

			// Reject future readers. We need to do this while holding the cache lock so that we don't
			// introduce a race between refCount.Add(1) in an invocation of getOrCreateRawEntry and the
			// wait call below.
			c.m.Lock()
			entry.draining = true
			c.m.Unlock()

			// Wait until initialized
			<-entry.init

			// Wait until there are no more readers
			for atomic.LoadUint32(&entry.refCount) != 0 {
				entry.refCountChanged.Wait()
			}

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
