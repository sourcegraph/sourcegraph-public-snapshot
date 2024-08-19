package ccache

import (
	"sync"
	"time"
)

type layeredBucket[T any] struct {
	sync.RWMutex
	buckets map[string]*bucket[T]
}

func (b *layeredBucket[T]) itemCount() int {
	count := 0
	b.RLock()
	defer b.RUnlock()
	for _, b := range b.buckets {
		count += b.itemCount()
	}
	return count
}

func (b *layeredBucket[T]) get(primary, secondary string) *Item[T] {
	bucket := b.getSecondaryBucket(primary)
	if bucket == nil {
		return nil
	}
	return bucket.get(secondary)
}

func (b *layeredBucket[T]) getSecondaryBucket(primary string) *bucket[T] {
	b.RLock()
	bucket, exists := b.buckets[primary]
	b.RUnlock()
	if exists == false {
		return nil
	}
	return bucket
}

func (b *layeredBucket[T]) set(primary, secondary string, value T, duration time.Duration, track bool) (*Item[T], *Item[T]) {
	b.Lock()
	bkt, exists := b.buckets[primary]
	if exists == false {
		bkt = &bucket[T]{lookup: make(map[string]*Item[T])}
		b.buckets[primary] = bkt
	}
	b.Unlock()
	item, existing := bkt.set(secondary, value, duration, track)
	item.group = primary
	return item, existing
}

func (b *layeredBucket[T]) delete(primary, secondary string) *Item[T] {
	b.RLock()
	bucket, exists := b.buckets[primary]
	b.RUnlock()
	if exists == false {
		return nil
	}
	return bucket.delete(secondary)
}

func (b *layeredBucket[T]) deletePrefix(primary, prefix string, deletables chan *Item[T]) int {
	b.RLock()
	bucket, exists := b.buckets[primary]
	b.RUnlock()
	if exists == false {
		return 0
	}
	return bucket.deletePrefix(prefix, deletables)
}

func (b *layeredBucket[T]) deleteFunc(primary string, matches func(key string, item *Item[T]) bool, deletables chan *Item[T]) int {
	b.RLock()
	bucket, exists := b.buckets[primary]
	b.RUnlock()
	if exists == false {
		return 0
	}
	return bucket.deleteFunc(matches, deletables)
}

func (b *layeredBucket[T]) deleteAll(primary string, deletables chan *Item[T]) bool {
	b.RLock()
	bucket, exists := b.buckets[primary]
	b.RUnlock()
	if exists == false {
		return false
	}

	bucket.Lock()
	defer bucket.Unlock()

	if l := len(bucket.lookup); l == 0 {
		return false
	}
	for key, item := range bucket.lookup {
		delete(bucket.lookup, key)
		deletables <- item
	}
	return true
}

func (b *layeredBucket[T]) forEachFunc(primary string, matches func(key string, item *Item[T]) bool) {
	b.RLock()
	bucket, exists := b.buckets[primary]
	b.RUnlock()
	if exists {
		bucket.forEachFunc(matches)
	}
}

// we expect the caller to have acquired a write lock
func (b *layeredBucket[T]) clear() {
	for _, bucket := range b.buckets {
		bucket.clear()
	}
	b.buckets = make(map[string]*bucket[T])
}
