// An LRU cached aimed at high concurrency
package ccache

import (
	"hash/fnv"
	"sync/atomic"
	"time"
)

type LayeredCache[T any] struct {
	*Configuration[T]
	control
	list        *List[*Item[T]]
	buckets     []*layeredBucket[T]
	bucketMask  uint32
	size        int64
	deletables  chan *Item[T]
	promotables chan *Item[T]
}

// Create a new layered cache with the specified configuration.
// A layered cache used a two keys to identify a value: a primary key
// and a secondary key. Get, Set and Delete require both a primary and
// secondary key. However, DeleteAll requires only a primary key, deleting
// all values that share the same primary key.

// Layered Cache is useful as an HTTP cache, where an HTTP purge might
// delete multiple variants of the same resource:
// primary key = "user/44"
// secondary key 1 = ".json"
// secondary key 2 = ".xml"

// See ccache.Configure() for creating a configuration
func Layered[T any](config *Configuration[T]) *LayeredCache[T] {
	c := &LayeredCache[T]{
		list:          NewList[*Item[T]](),
		Configuration: config,
		control:       newControl(),
		bucketMask:    uint32(config.buckets) - 1,
		buckets:       make([]*layeredBucket[T], config.buckets),
		deletables:    make(chan *Item[T], config.deleteBuffer),
		promotables:   make(chan *Item[T], config.promoteBuffer),
	}
	for i := 0; i < int(config.buckets); i++ {
		c.buckets[i] = &layeredBucket[T]{
			buckets: make(map[string]*bucket[T]),
		}
	}
	go c.worker()
	return c
}

func (c *LayeredCache[T]) ItemCount() int {
	count := 0
	for _, b := range c.buckets {
		count += b.itemCount()
	}
	return count
}

// Get an item from the cache. Returns nil if the item wasn't found.
// This can return an expired item. Use item.Expired() to see if the item
// is expired and item.TTL() to see how long until the item expires (which
// will be negative for an already expired item).
func (c *LayeredCache[T]) Get(primary, secondary string) *Item[T] {
	item := c.bucket(primary).get(primary, secondary)
	if item == nil {
		return nil
	}
	if item.expires > time.Now().UnixNano() {
		select {
		case c.promotables <- item:
		default:
		}
	}
	return item
}

// Same as Get but does not promote the value. This essentially circumvents the
// "least recently used" aspect of this cache. To some degree, it's akin to a
// "peak"
func (c *LayeredCache[T]) GetWithoutPromote(primary, secondary string) *Item[T] {
	return c.bucket(primary).get(primary, secondary)
}

func (c *LayeredCache[T]) ForEachFunc(primary string, matches func(key string, item *Item[T]) bool) {
	c.bucket(primary).forEachFunc(primary, matches)
}

// Get the secondary cache for a given primary key. This operation will
// never return nil. In the case where the primary key does not exist, a
// new, underlying, empty bucket will be created and returned.
func (c *LayeredCache[T]) GetOrCreateSecondaryCache(primary string) *SecondaryCache[T] {
	primaryBkt := c.bucket(primary)
	bkt := primaryBkt.getSecondaryBucket(primary)
	primaryBkt.Lock()
	if bkt == nil {
		bkt = &bucket[T]{lookup: make(map[string]*Item[T])}
		primaryBkt.buckets[primary] = bkt
	}
	primaryBkt.Unlock()
	return &SecondaryCache[T]{
		bucket: bkt,
		pCache: c,
	}
}

// Used when the cache was created with the Track() configuration option.
// Avoid otherwise
func (c *LayeredCache[T]) TrackingGet(primary, secondary string) TrackedItem[T] {
	item := c.Get(primary, secondary)
	if item == nil {
		return nil
	}
	item.track()
	return item
}

// Set the value in the cache for the specified duration
func (c *LayeredCache[T]) TrackingSet(primary, secondary string, value T, duration time.Duration) TrackedItem[T] {
	return c.set(primary, secondary, value, duration, true)
}

// Set the value in the cache for the specified duration
func (c *LayeredCache[T]) Set(primary, secondary string, value T, duration time.Duration) {
	c.set(primary, secondary, value, duration, false)
}

// Replace the value if it exists, does not set if it doesn't.
// Returns true if the item existed an was replaced, false otherwise.
// Replace does not reset item's TTL nor does it alter its position in the LRU
func (c *LayeredCache[T]) Replace(primary, secondary string, value T) bool {
	item := c.bucket(primary).get(primary, secondary)
	if item == nil {
		return false
	}
	c.Set(primary, secondary, value, item.TTL())
	return true
}

// Attempts to get the value from the cache and calles fetch on a miss.
// If fetch returns an error, no value is cached and the error is returned back
// to the caller.
// Note that Fetch merely calls the public Get and Set functions. If you want
// a different Fetch behavior, such as thundering herd protection or returning
// expired items, implement it in your application.
func (c *LayeredCache[T]) Fetch(primary, secondary string, duration time.Duration, fetch func() (T, error)) (*Item[T], error) {
	item := c.Get(primary, secondary)
	if item != nil {
		return item, nil
	}
	value, err := fetch()
	if err != nil {
		return nil, err
	}
	return c.set(primary, secondary, value, duration, false), nil
}

// Remove the item from the cache, return true if the item was present, false otherwise.
func (c *LayeredCache[T]) Delete(primary, secondary string) bool {
	item := c.bucket(primary).delete(primary, secondary)
	if item != nil {
		c.deletables <- item
		return true
	}
	return false
}

// Deletes all items that share the same primary key
func (c *LayeredCache[T]) DeleteAll(primary string) bool {
	return c.bucket(primary).deleteAll(primary, c.deletables)
}

// Deletes all items that share the same primary key and prefix.
func (c *LayeredCache[T]) DeletePrefix(primary, prefix string) int {
	return c.bucket(primary).deletePrefix(primary, prefix, c.deletables)
}

// Deletes all items that share the same primary key and where the matches func evaluates to true.
func (c *LayeredCache[T]) DeleteFunc(primary string, matches func(key string, item *Item[T]) bool) int {
	return c.bucket(primary).deleteFunc(primary, matches, c.deletables)
}

func (c *LayeredCache[T]) set(primary, secondary string, value T, duration time.Duration, track bool) *Item[T] {
	item, existing := c.bucket(primary).set(primary, secondary, value, duration, track)
	if existing != nil {
		c.deletables <- existing
	}
	c.promote(item)
	return item
}

func (c *LayeredCache[T]) bucket(key string) *layeredBucket[T] {
	h := fnv.New32a()
	h.Write([]byte(key))
	return c.buckets[h.Sum32()&c.bucketMask]
}

func (c *LayeredCache[T]) halted(fn func()) {
	c.halt()
	defer c.unhalt()
	fn()
}

func (c *LayeredCache[T]) halt() {
	for _, bucket := range c.buckets {
		bucket.Lock()
	}
}

func (c *LayeredCache[T]) unhalt() {
	for _, bucket := range c.buckets {
		bucket.Unlock()
	}
}

func (c *LayeredCache[T]) promote(item *Item[T]) {
	c.promotables <- item
}

func (c *LayeredCache[T]) worker() {
	dropped := 0
	cc := c.control

	promoteItem := func(item *Item[T]) {
		if c.doPromote(item) && c.size > c.maxSize {
			dropped += c.gc()
		}
	}

	for {
		select {
		case item := <-c.promotables:
			promoteItem(item)
		case item := <-c.deletables:
			c.doDelete(item)
		case control := <-cc:
			switch msg := control.(type) {
			case controlStop:
				goto drain
			case controlGetDropped:
				msg.res <- dropped
				dropped = 0
			case controlSetMaxSize:
				c.maxSize = msg.size
				if c.size > c.maxSize {
					dropped += c.gc()
				}
				msg.done <- struct{}{}
			case controlClear:
				promotables := c.promotables
				for len(promotables) > 0 {
					<-promotables
				}
				deletables := c.deletables
				for len(deletables) > 0 {
					<-deletables
				}

				c.halted(func() {
					for _, bucket := range c.buckets {
						bucket.clear()
					}
					c.size = 0
					c.list = NewList[*Item[T]]()
				})
				msg.done <- struct{}{}
			case controlGetSize:
				msg.res <- c.size
			case controlGC:
				dropped += c.gc()
				msg.done <- struct{}{}
			case controlSyncUpdates:
				doAllPendingPromotesAndDeletes(c.promotables, promoteItem, c.deletables, c.doDelete)
				msg.done <- struct{}{}
			}
		}
	}

drain:
	for {
		select {
		case item := <-c.deletables:
			c.doDelete(item)
		default:
			return
		}
	}
}

func (c *LayeredCache[T]) doDelete(item *Item[T]) {
	if item.node == nil {
		item.promotions = -2
	} else {
		c.size -= item.size
		if c.onDelete != nil {
			c.onDelete(item)
		}
		c.list.Remove(item.node)
		item.node = nil
		item.promotions = -2
	}
}

func (c *LayeredCache[T]) doPromote(item *Item[T]) bool {
	// deleted before it ever got promoted
	if item.promotions == -2 {
		return false
	}
	if item.node != nil { //not a new item
		if item.shouldPromote(c.getsPerPromote) {
			c.list.MoveToFront(item.node)
			item.promotions = 0
		}
		return false
	}
	c.size += item.size
	item.node = c.list.Insert(item)
	return true
}

func (c *LayeredCache[T]) gc() int {
	node := c.list.Tail
	dropped := 0
	itemsToPrune := int64(c.itemsToPrune)

	if min := c.size - c.maxSize; min > itemsToPrune {
		itemsToPrune = min
	}

	for i := int64(0); i < itemsToPrune; i++ {
		if node == nil {
			return dropped
		}
		prev := node.Prev
		item := node.Value
		if c.tracking == false || atomic.LoadInt32(&item.refCount) == 0 {
			c.bucket(item.group).delete(item.group, item.key)
			c.size -= item.size
			c.list.Remove(node)
			if c.onDelete != nil {
				c.onDelete(item)
			}
			item.node = nil
			item.promotions = -2
			dropped += 1
		}
		node = prev
	}
	return dropped
}
