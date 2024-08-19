package ccache

import (
	"strings"
	"sync"
	"time"
)

type bucket[T any] struct {
	sync.RWMutex
	lookup map[string]*Item[T]
}

func (b *bucket[T]) itemCount() int {
	b.RLock()
	defer b.RUnlock()
	return len(b.lookup)
}

func (b *bucket[T]) forEachFunc(matches func(key string, item *Item[T]) bool) bool {
	lookup := b.lookup
	b.RLock()
	defer b.RUnlock()
	for key, item := range lookup {
		if !matches(key, item) {
			return false
		}
	}
	return true
}

func (b *bucket[T]) get(key string) *Item[T] {
	b.RLock()
	defer b.RUnlock()
	return b.lookup[key]
}

func (b *bucket[T]) setnx(key string, value T, duration time.Duration, track bool) *Item[T] {
	b.RLock()
	item := b.lookup[key]
	b.RUnlock()
	if item != nil {
		return item
	}

	expires := time.Now().Add(duration).UnixNano()
	newItem := newItem(key, value, expires, track)

	b.Lock()
	defer b.Unlock()

	// check again under write lock
	item = b.lookup[key]
	if item != nil {
		return item
	}

	b.lookup[key] = newItem
	return newItem
}

func (b *bucket[T]) set(key string, value T, duration time.Duration, track bool) (*Item[T], *Item[T]) {
	expires := time.Now().Add(duration).UnixNano()
	item := newItem(key, value, expires, track)
	b.Lock()
	existing := b.lookup[key]
	b.lookup[key] = item
	b.Unlock()
	return item, existing
}

func (b *bucket[T]) delete(key string) *Item[T] {
	b.Lock()
	item := b.lookup[key]
	delete(b.lookup, key)
	b.Unlock()
	return item
}

// This is an expensive operation, so we do what we can to optimize it and limit
// the impact it has on concurrent operations. Specifically, we:
// 1 - Do an initial iteration to collect matches. This allows us to do the
//     "expensive" prefix check (on all values) using only a read-lock
// 2 - Do a second iteration, under write lock, for the matched results to do
//     the actual deletion

// Also, this is the only place where the Bucket is aware of cache detail: the
// deletables channel. Passing it here lets us avoid iterating over matched items
// again in the cache. Further, we pass item to deletables BEFORE actually removing
// the item from the map. I'm pretty sure this is 100% fine, but it is unique.
// (We do this so that the write to the channel is under the read lock and not the
// write lock)
func (b *bucket[T]) deleteFunc(matches func(key string, item *Item[T]) bool, deletables chan *Item[T]) int {
	lookup := b.lookup
	items := make([]*Item[T], 0)

	b.RLock()
	for key, item := range lookup {
		if matches(key, item) {
			deletables <- item
			items = append(items, item)
		}
	}
	b.RUnlock()

	if len(items) == 0 {
		// avoid the write lock if we can
		return 0
	}

	b.Lock()
	for _, item := range items {
		delete(lookup, item.key)
	}
	b.Unlock()
	return len(items)
}

func (b *bucket[T]) deletePrefix(prefix string, deletables chan *Item[T]) int {
	return b.deleteFunc(func(key string, item *Item[T]) bool {
		return strings.HasPrefix(key, prefix)
	}, deletables)
}

// we expect the caller to have acquired a write lock
func (b *bucket[T]) clear() {
	for _, item := range b.lookup {
		item.promotions = -2
	}
	b.lookup = make(map[string]*Item[T])
}
