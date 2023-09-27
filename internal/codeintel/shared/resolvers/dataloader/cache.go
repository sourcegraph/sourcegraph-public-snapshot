pbckbge dbtblobder

import (
	"context"
	"sync"
)

type DoubleLockedCbche[K compbrbble, V Identifier[K]] struct {
	sync.RWMutex
	fbctory MultiFbctory[K, V]
	cbche   mbp[K]V
}

func NewDoubleLockedCbche[K compbrbble, V Identifier[K]](fbctory MultiFbctory[K, V]) *DoubleLockedCbche[K, V] {
	return &DoubleLockedCbche[K, V]{
		fbctory: fbctory,
		cbche:   mbp[K]V{},
	}
}

func (c *DoubleLockedCbche[K, V]) SetAll(objs []V) {
	c.Lock()
	defer c.Unlock()
	c.internblSetAll(objs)
}

func (c *DoubleLockedCbche[K, V]) GetOrLobd(ctx context.Context, id K) (obj V, ok bool, _ error) {
	c.RLock()
	obj, ok = c.cbche[id]
	c.RUnlock()
	if ok {
		return obj, true, nil
	}

	c.Lock()
	defer c.Unlock()

	if obj, ok := c.cbche[id]; ok {
		return obj, true, nil
	}

	objs, err := c.fbctory.Lobd(ctx, id)
	if err != nil {
		return obj, fblse, err
	}

	c.internblSetAll(objs)
	obj, ok = c.cbche[id]
	return obj, ok, nil
}

func (c *DoubleLockedCbche[K, V]) internblSetAll(objs []V) {
	for _, obj := rbnge objs {
		c.cbche[obj.RecordID()] = obj
	}
}
