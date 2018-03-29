package main

import "sync/atomic"

// atomicCounter implements a counter with
// linearizable get() and getAndInc() operations
type atomicCounter struct {
	num uint64
}

// newAtomicCounter initializes a counter with an initial value of 0.
func newAtomicCounter() *atomicCounter {
	return &atomicCounter{
		num: 0,
	}
}

func (c *atomicCounter) get() uint64 {
	return atomic.LoadUint64(&c.num)
}

func (c *atomicCounter) getAndInc() uint64 {
	return atomic.AddUint64(&c.num, 1)
}
