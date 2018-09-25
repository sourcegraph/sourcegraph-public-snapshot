// +build go1.3

package web

import "sync"

type cPool sync.Pool

func makeCPool() *cPool {
	return &cPool{}
}

func (c *cPool) alloc() *cStack {
	cs := (*sync.Pool)(c).Get()
	if cs == nil {
		return nil
	}
	return cs.(*cStack)
}

func (c *cPool) release(cs *cStack) {
	(*sync.Pool)(c).Put(cs)
}
