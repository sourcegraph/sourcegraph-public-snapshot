package cache

import (
	"container/list"
	"context"
	"errors"
	"sync"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence"
)

// ErrReaderInitializationDeadlineExceeded occurs when a new reader takes too long to initialize.
var ErrReaderInitializationDeadlineExceeded = errors.New("reader initialization deadline exceeded")

// Cache is an LRU cache of reader instances by key. This implementation ensures that there is only one
// handle open to a given reader, and that the initialization of the reader (which may run migrations)
// is only attempted once at a time.
type Cache struct {
	size   int
	opener ReaderOpener

	m         sync.Mutex
	entries   map[string]*list.Element // keys to evict list nodes
	evictList *list.List               // LRU ordering of cache entries
}

// ReaderOpener initializes a new reader for the given key.
type ReaderOpener func(key string) (persistence.Reader, error)

// Handler performs an operation on a reader. The invocation of this function is a critical section that
// locks the given reader argument so that it is not closed while in use.
type Handler func(reader persistence.Reader) error

// New initializes a new cache with the given capacity and reader opener.
func New(size int, opener ReaderOpener) *Cache {
	return &Cache{
		size:      size,
		opener:    opener,
		entries:   map[string]*list.Element{},
		evictList: list.New(),
	}
}

// WithReader calls the given function with a reader argument. If the reader has not yet initialized and
// does not do so before context deadline, an error is returned. The reader initialization is not canceled
// due to a context deadline here and will continue to run in the background until completion.
func (c *Cache) WithReader(ctx context.Context, key string, fn Handler) error {
	entry, ok := c.getOrCreateActiveEntry(ctx, key)
	if !ok {
		return ErrReaderInitializationDeadlineExceeded
	}
	defer entry.wg.Done()

	select {
	case <-entry.init:
		if entry.err != nil {
			return entry.err
		}
		return fn(entry.reader)

	case <-ctx.Done():
		return ErrReaderInitializationDeadlineExceeded
	}
}

const (
	MinBackoff           = time.Millisecond * 1
	MaxBackoff           = time.Millisecond * 250
	BackoffIncreaseRatio = 1.5
	MaxAttempts          = 100
)

// getOrCreateActiveEntry gets or creates a cache entry for the given key. If there exists an entry for
// the key, but that entry is marked as draining, we retry following an exponential backoff algorithm.
func (c *Cache) getOrCreateActiveEntry(ctx context.Context, key string) (*cacheEntry, bool) {
	backoff := MinBackoff

	for attempts := MaxAttempts; attempts > 0; attempts-- {
		if entry, draining := c.getOrCreateRawEntry(key); !draining {
			return entry, true
		}

		select {
		case <-time.After(backoff):
			if backoff = time.Duration(float64(backoff) * BackoffIncreaseRatio); backoff > MaxBackoff {
				backoff = MaxBackoff
			}
			continue

		case <-ctx.Done():
			break
		}
	}

	return nil, false
}

// getOrCreateRawEntry gets or creates a cache entry for the given key and a boolean flag indicating
// whether or not the entry is "active". False indicates that the entry was marked as draining. In this
// case the entry's wait group value is not modified and the entry's close procedure is unaffected.
func (c *Cache) getOrCreateRawEntry(key string) (*cacheEntry, bool) {
	c.m.Lock()
	defer c.m.Unlock()

	if entry, exists, draining := c.getEntry(key); exists {
		return entry, draining
	}

	return c.createEntry(key), false
}

// getEntry attempts to return an existing entry for the given key. This function returns boolen flags
// indicating whether an entry exists, and whether or not that entry is currently draining. This function
// assumes that the cache's mutex is held by the caller.
func (c *Cache) getEntry(key string) (_ *cacheEntry, exists, draining bool) {
	element, ok := c.entries[key]
	if !ok {
		return nil, false, false
	}

	entry := element.Value.(*cacheEntry)
	entry.wg.Add(1)                  // Mark as in-use
	c.evictList.MoveToFront(element) // Update recency data

	select {
	case <-entry.draining:
		// We optimistically added one to the wait group above. If the entry is draining at this
		// point, it is not safe for use as the close procedure may have already unblocked on the
		// wait group. Undo our mark here and try to get a new entry.
		entry.wg.Done()
		return nil, true, true
	default:
	}

	return entry, true, false
}

// createEntry creates a new cache entry for the given key. This function assumes that the cache's mutex is
// held by the caller.
func (c *Cache) createEntry(key string) *cacheEntry {
	entry := newCacheEntry(key)
	entry.wg.Add(1)                         // Mark as in-use
	element := c.evictList.PushFront(entry) // Update recency data
	c.entries[key] = element

	// Perform the open procedure in a goroutine. The close of the init channel will signal
	// all consumers of the cache that the entry's underlying reader is now ready for use.
	// This allows us to have several consumers concurrently waiting on the same cache entry
	// to initialize.
	//
	// Previous constructions of this cache did not guarantee this property:
	//   (1) holding a lock during initialization stops independent readers for initializing
	//       concurrently, which can severely decrease request throughput
	//   (2) not synchronizing during initialization can allow the same reader to be
	//       initialized multiple times as there is no "pending" entry in the cache for the
	//       second request to the same key to wait on.
	go func() {
		// Mark as initialized
		defer close(entry.init)

		entry.reader, entry.err = c.opener(key)
		if entry.err != nil {
			// Do not hold on to cache entries that failed to initialize. In the case of a
			// transient error (e.g., a missing bundle file that is later uploaded), we do
			// not want a poison cache value occupying that key until a process restart.
			c.remove(entry)
		}
	}()

	// We just created a new entry so we may now be over capcity. Remove as many entries
	// as necessary to get us back under our maximum size. This may call end up calling
	// remove on an entry that is already draining, but is safe to do as that function
	// is idempotent and is guaranteed to remove the underlying entry at some point in
	// the future.
	c.evict(len(c.entries) - c.size)

	return entry
}

// evict attempts to remove n elements from the back of the list. This function assumes that the cache's
// mutex is held by the caller.
func (c *Cache) evict(n int) {
	for element := c.evictList.Back(); element != nil; element = element.Prev() {
		if n <= 0 {
			return
		}

		n--
		c.remove(element.Value.(*cacheEntry))
	}
}

// remove marks the entry as draining so that future readers will not attempt to hold it,
// then waits for all current readers to drain. The entry is then removed from the cache
// mapping. This function is safe to call on the same entry multiple times.
func (c *Cache) remove(entry *cacheEntry) {
	entry.once.Do(func() {
		go func() {
			entry.close()

			c.m.Lock()
			defer c.m.Unlock()

			element := c.entries[entry.key]
			c.evictList.Remove(element)
			delete(c.entries, entry.key)
		}()
	})
}

// cacheEntry wraps a reader which may still be initializing.
type cacheEntry struct {
	key      string
	reader   persistence.Reader
	err      error
	init     chan struct{}  // marks init completion
	draining chan struct{}  // marks intent to close
	wg       sync.WaitGroup // concurrent use ref count
	once     sync.Once      // guards close function
}

func newCacheEntry(key string) *cacheEntry {
	return &cacheEntry{
		key:      key,
		init:     make(chan struct{}),
		draining: make(chan struct{}),
	}
}

// close marks the entry as draining then blocks until initialized and until all readers
// have finished. If the reader was initialized successfully, it is closed and all underlying
// resources are released.
func (e *cacheEntry) close() {
	close(e.draining) // Reject future readers
	<-e.init          // Wait until initialized
	e.wg.Wait()       // Wait until there are no more readers

	if e.reader == nil {
		return
	}

	// Release resources
	if err := e.reader.Close(); err != nil {
		log15.Error("Failed to close reader", "key", e.key, "err", err)
	}
}
