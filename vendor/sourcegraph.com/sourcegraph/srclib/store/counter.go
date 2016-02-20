package store

import "sync/atomic"

// counter is a simple thread-safe integer.
type counter struct {
	count *int64
}

// increment increments the counter by one.
func (c counter) increment() {
	atomic.AddInt64(c.count, 1)
}

// get returns the counter's current value.
func (c *counter) get() int {
	return int(atomic.LoadInt64(c.count))
}

// set sets the counter value.
func (c *counter) set(i int) {
	atomic.StoreInt64(c.count, int64(i))
}
