// +build !appengine

package web

import (
	"sync/atomic"
	"unsafe"
)

func (rt *router) getMachine() *routeMachine {
	ptr := (*unsafe.Pointer)(unsafe.Pointer(&rt.machine))
	sm := (*routeMachine)(atomic.LoadPointer(ptr))
	return sm
}
func (rt *router) setMachine(m *routeMachine) {
	ptr := (*unsafe.Pointer)(unsafe.Pointer(&rt.machine))
	atomic.StorePointer(ptr, unsafe.Pointer(m))
}
