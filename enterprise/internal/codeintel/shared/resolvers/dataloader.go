package sharedresolvers

import (
	"context"
	"sync"
)

type Identifier[T any] interface {
	RecordID() T
}

type MultiFactory[K, V any] interface {
	Load(ctx context.Context, id K) ([]V, error)
}

type MultiFactoryFunc[K, V any] func(ctx context.Context, id K) ([]V, error)

func (f MultiFactoryFunc[K, V]) Load(ctx context.Context, id K) ([]V, error) {
	return f(ctx, id)
}

func MultiFactoryFromFactoryFunc[K, V any](f func(ctx context.Context, id K) (V, error)) MultiFactory[K, V] {
	return MultiFactoryFunc[K, V](func(ctx context.Context, id K) ([]V, error) {
		v, err := f(ctx, id)
		if err != nil {
			return nil, err
		}

		return []V{v}, nil
	})
}

func MultiFactoryFromFallibleFactoryFunc[K, V any](f func(ctx context.Context, id K) (*V, error)) MultiFactory[K, V] {
	return MultiFactoryFunc[K, V](func(ctx context.Context, id K) ([]V, error) {
		v, err := f(ctx, id)
		if err != nil || v == nil {
			return nil, err
		}

		return []V{*v}, nil
	})
}

type DoubleLockedCache[K comparable, V Identifier[K]] struct {
	sync.RWMutex
	factory MultiFactory[K, V]
	cache   map[K]V
}

func NewDoubleLockedCache[K comparable, V Identifier[K]](factory MultiFactory[K, V]) *DoubleLockedCache[K, V] {
	return &DoubleLockedCache[K, V]{
		factory: factory,
		cache:   map[K]V{},
	}
}

func (c *DoubleLockedCache[K, V]) GetOrLoad(ctx context.Context, id K) (obj V, ok bool, _ error) {
	c.RLock()
	obj, ok = c.cache[id]
	c.RUnlock()
	if ok {
		return obj, true, nil
	}

	c.Lock()
	defer c.Unlock()

	if obj, ok := c.cache[id]; ok {
		return obj, true, nil
	}

	objs, err := c.factory.Load(ctx, id)
	if err != nil {
		return obj, false, err
	}
	for _, obj := range objs {
		c.cache[obj.RecordID()] = obj
	}

	obj, ok = c.cache[id]
	return obj, ok, nil
}

type DataLoader[K comparable, V Identifier[K]] struct {
	svc   DataLoaderBackingService[K, V]
	ids   map[K]struct{}
	cache *DoubleLockedCache[K, V]
}

type DataLoaderBackingService[K comparable, V Identifier[K]] interface {
	GetByIDs(ctx context.Context, ids ...K) ([]V, error)
}

type DataLoaderBackingServiceFunc[K, V any] func(ctx context.Context, ids ...K) ([]V, error)

func (f DataLoaderBackingServiceFunc[K, V]) GetByIDs(ctx context.Context, ids ...K) ([]V, error) {
	return f(ctx, ids...)
}

func NewDataLoader[K comparable, V Identifier[K]](svc DataLoaderBackingService[K, V]) *DataLoader[K, V] {
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
