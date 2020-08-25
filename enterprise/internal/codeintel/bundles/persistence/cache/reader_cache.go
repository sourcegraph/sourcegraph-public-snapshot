package cache

import (
	"context"
	"sync"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence"
)

// ReaderCache is a cache of reader instances with very short TTL values. The primary purpose of this
// cache is to ensure that multiple concurrent requests for the same reader do not launch multiple
// expensive migration operations, which occur when the reader of an older bundle is first opened.
// Often, this results in one or more readers failing with a transaction error or a locked database.
//
// Readers in this cache are closed soon after they are first used, but any concurrent requests coming
// in for the same reader will share the same cache entry and both will wait on the same reader to
// complete initialization. It is guaranteed that a reader will remain open while there are active
// users of that reader instance.
type ReaderCache interface {
	// WithReader calls the given handler function with a reader argument. If the reader has not yet
	// initialized and does not do so before context deadline, an error is returned. The reader
	// initialization is not canceled due to a context deadline here and will continue to run in the
	// background until completion.
	WithReader(ctx context.Context, key string, f HandlerFunc) error
}

// ReaderOpener initializes a new reader for the given key.
type ReaderOpener func(key string) (persistence.Reader, error)

// Handler performs an operation on a reader. The invocation of this function is a critical section that
// locks the given reader argument so that it is not closed while in use.
type HandlerFunc func(reader persistence.Reader) error

type readerCache struct {
	opener ReaderOpener

	m       sync.RWMutex
	readers map[string]*readerCacheEntry
}

// NewReaderCache creates a new reader cache with the given reader opener.
func NewReaderCache(opener ReaderOpener) ReaderCache {
	return newReaderCache(opener)
}

func newReaderCache(opener ReaderOpener) *readerCache {
	return &readerCache{
		opener:  opener,
		readers: map[string]*readerCacheEntry{},
	}
}

// WithReader calls the given handler function with a reader argument. If the reader has not yet
// initialized and does not do so before context deadline, an error is returned. The reader
// initialization is not canceled due to a context deadline here and will continue to run in the
// background until completion.
func (c *readerCache) WithReader(ctx context.Context, key string, f HandlerFunc) error {
	entry, err := c.getLockedEntry(ctx, key)
	if err != nil {
		return err
	}
	defer entry.dec()

	select {
	case <-makeInitChannel(entry):
	case <-ctx.Done():
		return ctx.Err()
	}

	if entry.err != nil {
		return entry.err
	}

	return f(entry.reader)
}

// getLockedEntry gets a non-disposed cache entry or creates a new one. The caller must ensure
// that the returned entry's refcount is decremented after use.
func (c *readerCache) getLockedEntry(ctx context.Context, key string) (*readerCacheEntry, error) {
	for {
		if entry, locked := c.getOrCreateEntry(key); locked || entry.inc() {
			return entry, nil
		}

		select {
		case <-time.After(time.Millisecond):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// getOrCreateEntry retrieves an existing reader cache entry for the given key. If one does not
// exist, a new one is created. This method also returns a boolean flag indicating whether or not
// the refCount was incremented as part of construction. If this value is true, the caller is
// expected to decrement the entry's refcount after use (but not increment it).
func (c *readerCache) getOrCreateEntry(key string) (*readerCacheEntry, bool) {
	if entry := c.getEntry(key); entry != nil {
		return entry, false
	}

	return c.createEntry(key)
}

// getEntry attempts to return an existing entry for the given key. If such an entry does not exist,
// this method returns nil.
func (c *readerCache) getEntry(key string) *readerCacheEntry {
	c.m.RLock()
	defer c.m.RUnlock()
	return c.readers[key]
}

// createEntry attempts to create a new entry for the given key if such an entry does not already exist.
// This method also returns a boolean flag indicating whether or not the entry was just created. Newly
// created entries have a refcount of zero (and are already marked as in-use for the calling function).
func (c *readerCache) createEntry(key string) (*readerCacheEntry, bool) {
	c.m.Lock()
	defer c.m.Unlock()

	// Ensure entry doesn't exist after acquiring write lock
	if entry, ok := c.readers[key]; ok {
		return entry, false
	}

	// Create new entry and make it immediately visible to all other requests
	entry := newReaderCacheEntry()
	c.readers[key] = entry

	go func() {
		// Initialize the reader in the background. This allows multiple concurrent requests
		// for the same reader to share the same entry while it is initializing, which an take
		// a non-negligible amount of time for large, outdated bundles.
		entry.init(func() (persistence.Reader, error) { return c.opener(key) })

		// Block until the entry is unused, then close the underlying reader
		entry.closeOnceUnused()

		// Remove entry from map so that it's inaccessible to future requests
		c.removeEntry(key)
	}()

	return entry, true
}

// removeEntry removes the given key from the cache's reader map.
func (c *readerCache) removeEntry(key string) {
	c.m.Lock()
	delete(c.readers, key)
	c.m.Unlock()
}

// makeInitChannel creates a channel that closes once the given entry has initialized.
func makeInitChannel(entry *readerCacheEntry) <-chan struct{} {
	ch := make(chan struct{})

	go func() {
		defer close(ch)
		entry.waitUntilInitialized()
	}()

	return ch
}

// readerCacheEntry wraps a reader with its initialization state and its ref count.
type readerCacheEntry struct {
	m           *sync.Mutex        // protects all fields
	reader      persistence.Reader // shared reader instance
	err         error              // construction error
	initialized bool               // set when reader/err fields are set
	disposed    bool               // set when entry is no longer usable
	refCount    uint32             // number of references to entry

	// cond wraps the entry's mutex and broadcasts to waiting
	// goroutines when initialized is first set to true and when
	// refCount is decremented.
	cond *sync.Cond
}

// newReaderCacheEntry creates a new cache entry with a refcount of one.
func newReaderCacheEntry() *readerCacheEntry {
	m := &sync.Mutex{}

	return &readerCacheEntry{
		m:        m,
		cond:     sync.NewCond(m),
		refCount: 1,
	}
}

// init calls the given function to construct a reader, then updates the reader, err,
// and initialized fields of the entry.
func (e *readerCacheEntry) init(openReaderFunc func() (persistence.Reader, error)) {
	reader, err := openReaderFunc()

	e.m.Lock()
	e.reader = reader
	e.err = err
	e.initialized = true
	e.m.Unlock()

	// wake users waiting for initialization to complete
	e.cond.Broadcast()
}

// waitUntilInitialized blocks until the entry has completed initialization.
func (e *readerCacheEntry) waitUntilInitialized() {
	e.m.Lock()
	defer e.m.Unlock()

	for !e.initialized {
		e.cond.Wait()
	}
}

// closeOnceUnused blocks until the refcount of the entry goes to zero. The entry's disposed
// flag is set to indicate to future users that the underlying reader is no longer usable. The
// entry's reader is then closed (if it was initialized successfully).
func (e *readerCacheEntry) closeOnceUnused() {
	e.m.Lock()
	for e.refCount > 0 {
		e.cond.Wait()
	}
	e.disposed = true
	e.m.Unlock()

	if e.reader != nil {
		e.reader.Close(nil)
	}
}

// inc attempts to increase the refcount by one. If the entry is already disposed, this
// method returns false and the entry is not modified.
func (e *readerCacheEntry) inc() bool {
	e.m.Lock()
	defer e.m.Unlock()

	if e.disposed {
		return false
	}

	e.refCount++
	return true
}

// dec decreases the refcount by one.
func (e *readerCacheEntry) dec() {
	e.m.Lock()
	e.refCount--
	e.m.Unlock()
	e.cond.Broadcast()
}
