package ccache

import "time"

type SecondaryCache[T any] struct {
	bucket *bucket[T]
	pCache *LayeredCache[T]
}

// Get the secondary key.
// The semantics are the same as for LayeredCache.Get
func (s *SecondaryCache[T]) Get(secondary string) *Item[T] {
	return s.bucket.get(secondary)
}

// Set the secondary key to a value.
// The semantics are the same as for LayeredCache.Set
func (s *SecondaryCache[T]) Set(secondary string, value T, duration time.Duration) *Item[T] {
	item, existing := s.bucket.set(secondary, value, duration, false)
	if existing != nil {
		s.pCache.deletables <- existing
	}
	s.pCache.promote(item)
	return item
}

// Fetch or set a secondary key.
// The semantics are the same as for LayeredCache.Fetch
func (s *SecondaryCache[T]) Fetch(secondary string, duration time.Duration, fetch func() (T, error)) (*Item[T], error) {
	item := s.Get(secondary)
	if item != nil {
		return item, nil
	}
	value, err := fetch()
	if err != nil {
		return nil, err
	}
	return s.Set(secondary, value, duration), nil
}

// Delete a secondary key.
// The semantics are the same as for LayeredCache.Delete
func (s *SecondaryCache[T]) Delete(secondary string) bool {
	item := s.bucket.delete(secondary)
	if item != nil {
		s.pCache.deletables <- item
		return true
	}
	return false
}

// Replace a secondary key.
// The semantics are the same as for LayeredCache.Replace
func (s *SecondaryCache[T]) Replace(secondary string, value T) bool {
	item := s.Get(secondary)
	if item == nil {
		return false
	}
	s.Set(secondary, value, item.TTL())
	return true
}

// Track a secondary key.
// The semantics are the same as for LayeredCache.TrackingGet
func (c *SecondaryCache[T]) TrackingGet(secondary string) TrackedItem[T] {
	item := c.Get(secondary)
	if item == nil {
		return nil
	}
	item.track()
	return item
}
