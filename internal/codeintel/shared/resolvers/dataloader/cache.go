package dataloader

import (
	"context"
	"sync"
)

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

func (c *DoubleLockedCache[K, V]) SetAll(objs []V) {
	c.Lock()
	defer c.Unlock()
	c.internalSetAll(objs)
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

	c.internalSetAll(objs)
	obj, ok = c.cache[id]
	return obj, ok, nil
}

func (c *DoubleLockedCache[K, V]) internalSetAll(objs []V) {
	for _, obj := range objs {
		c.cache[obj.RecordID()] = obj
	}
}
