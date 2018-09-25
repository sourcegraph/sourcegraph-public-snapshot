// +build !go1.3

package web

// This is an alternate implementation of Go 1.3's sync.Pool.

// Maximum size of the pool of spare middleware stacks
const cPoolSize = 32

type cPool chan *cStack

func makeCPool() *cPool {
	var p cPool = make(chan *cStack, cPoolSize)
	return &p
}

func (c cPool) alloc() *cStack {
	select {
	case cs := <-c:
		return cs
	default:
		return nil
	}
}

func (c cPool) release(cs *cStack) {
	select {
	case c <- cs:
	default:
	}
}
