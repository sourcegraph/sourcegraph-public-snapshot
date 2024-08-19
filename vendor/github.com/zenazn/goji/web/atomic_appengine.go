// +build appengine

package web

func (rt *router) getMachine() *routeMachine {
	rt.lock.Lock()
	defer rt.lock.Unlock()
	return rt.machine
}

// We always hold the lock when calling setMachine.
func (rt *router) setMachine(m *routeMachine) {
	rt.machine = m
}
