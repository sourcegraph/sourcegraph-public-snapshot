// +build go1.3

package stack

import (
	"sync"
)

var pcStackPool = sync.Pool{
	New: func() interface{} { return make([]uintptr, 1000) },
}

func poolBuf() []uintptr {
	return pcStackPool.Get().([]uintptr)
}

func putPoolBuf(p []uintptr) {
	pcStackPool.Put(p)
}
