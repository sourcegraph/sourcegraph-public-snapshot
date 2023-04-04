package dataloader

import (
	"context"
)

type DataLoader[K comparable, V Identifier[K]] struct {
	svc   BackingService[K, V]
	ids   map[K]struct{}
	cache *DoubleLockedCache[K, V]
}

func New[K comparable, V Identifier[K]](svc BackingService[K, V]) *DataLoader[K, V] {
	dl := &DataLoader[K, V]{
		svc: svc,
		ids: map[K]struct{}{},
	}

	dl.cache = NewDoubleLockedCache[K, V](MultiFactoryFunc[K, V](dl.load))
	return dl
}

func (l *DataLoader[K, V]) Presubmit(ids ...K) {
	l.cache.Lock()
	defer l.cache.Unlock()

	for _, id := range ids {
		l.ids[id] = struct{}{}
	}
}

func (l *DataLoader[K, V]) GetByID(ctx context.Context, id K) (obj V, ok bool, _ error) {
	return l.cache.GetOrLoad(ctx, id)
}

// note: this is called while the cache's exclusive lock is held
func (l *DataLoader[K, V]) load(ctx context.Context, id K) ([]V, error) {
	l.ids[id] = struct{}{}   // ensure batch includes id
	ids := keys(l.ids)       // consume
	l.ids = map[K]struct{}{} // reset

	return l.svc.GetByIDs(ctx, ids...)
}

func keys[T comparable](m map[T]struct{}) []T {
	keys := make([]T, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	return keys
}
