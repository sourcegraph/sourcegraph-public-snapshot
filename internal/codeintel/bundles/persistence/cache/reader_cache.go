package cache

import (
	"container/list"
	"context"
	"errors"
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence"
)

// ErrReaderInitializationDeadlineExceeded occurs when a new reader takes too long to initialize.
var ErrReaderInitializationDeadlineExceeded = errors.New("reader initialization deadline exceeded")

// Cache is an LRU cache of reader instances by key. This implementation ensures that there is only one
// handle open to a given reader, and that the initialization of the reader (which may run migrations)
// is only attempted once at a time.
type Cache struct {
	size      int
	opener    ReaderOpener
	m         sync.Mutex               // guards mutation of entries/evictList
	entries   map[string]*list.Element // keys to evict list nodes
	evictList *list.List               // LRU ordering of cache entries
}

// ReaderOpener initializes a new reader for the given key.
type ReaderOpener func(key string) (persistence.Reader, error)

// Handler performs an operation on a reader. The invocation of this function is
// a critical section that locks the given reader argument so tha it is not closed while
// in use.
type Handler func(reader persistence.Reader) error

// cacheEntry wraps a reader which may still be initializing.
type cacheEntry struct {
	key    string
	reader persistence.Reader
	err    error
	init   chan struct{}  // marks availability of reader field
	once   sync.Once      // guards db.Close()
	wg     sync.WaitGroup // concurrent use ref count
}

// New initializes a new cache with the given capacity and reader opener.
func New(size int, opener ReaderOpener) *Cache {
	return &Cache{
		size:      size,
		opener:    opener,
		entries:   map[string]*list.Element{},
		evictList: list.New(),
	}
}

// WithReader calls the given function with a reader. If the reader has not yet initialized and does not
// do so before context deadline, an error is returned. The reader initialization is not canceled due to
// a context deadline here and will continue to run in the background until completion.
func (c *Cache) WithReader(ctx context.Context, key string, fn Handler) error {
	entry := c.entry(key)
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

// entry gets or creates a cache entry for the given key.
func (c *Cache) entry(key string) *cacheEntry {
	c.m.Lock()
	defer c.m.Unlock()

	if element, ok := c.entries[key]; ok {
		entry := element.Value.(*cacheEntry)
		entry.wg.Add(1)                  // Mark as in-use
		c.evictList.MoveToFront(element) // Update recency data
		return entry
	}

	entry := &cacheEntry{
		key:  key,
		init: make(chan struct{}),
	}
	entry.wg.Add(1)                         // Mark as in-use
	element := c.evictList.PushFront(entry) // Update recency data
	c.entries[key] = element

	go func() {
		// Mark as initialized
		defer close(entry.init)

		// Perform open procedure in a goroutine. We organize it this way so that
		// if multiple differnet goroutines want to use the the same reader they do
		// not all try to initialize it at once. In this situation, each goroutine
		// will get a reference to the same entry and will block on the init channel
		// being closed to signal that the handler cna be called.
		entry.reader, entry.err = c.opener(key)
		if entry.err != nil {
			// Immediately remove from cache on error so that we do not permanently
			// hold on to handles to missing bundles, as they may be uploaded in the
			// future.
			c.removeOnce(entry)
		}
	}()

	// Delete n elements from the back of the list where n is how many elements over
	// capacity we are.
	n := len(c.entries) - c.size
	for i, element := n, c.evictList.Back(); i > 0 && element != nil; i, element = i-1, element.Prev() {
		c.removeOnce(element.Value.(*cacheEntry))
	}

	return entry
}

// removeOnce calls remove once in a goroutine.
func (c *Cache) removeOnce(entry *cacheEntry) {
	entry.once.Do(func() {
		go func() { c.remove(entry) }()
	})
}

// removeOnce blocks until the entry is initialized and exclusive access is available.
// The entry is then removed from the cache and the entry's reader is closed.
func (c *Cache) remove(entry *cacheEntry) {
	<-entry.init    // Wait until initialized
	entry.wg.Wait() // Wait until there are no more readers

	c.m.Lock()
	defer c.m.Unlock()

	// Remove from cache
	element := c.entries[entry.key]
	c.evictList.Remove(element)
	delete(c.entries, entry.key)

	// Release resources
	if entry.reader != nil {
		if err := entry.reader.Close(); err != nil {
			log15.Error("Failed to close reader", "key", entry.key, "err", err)
		}
	}
}
