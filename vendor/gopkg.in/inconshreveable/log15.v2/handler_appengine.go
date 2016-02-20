// +build appengine

package log15

import "sync"

// swapHandler wraps another handler that may be swapped out
// dynamically at runtime in a thread-safe fashion.
type swapHandler struct {
	handler interface{}
	lock    sync.RWMutex
}

func (h *swapHandler) Log(r *Record) error {
	h.lock.RLock()
	defer h.lock.RUnlock()

	return h.handler.(Handler).Log(r)
}

func (h *swapHandler) Swap(newHandler Handler) {
	h.lock.Lock()
	defer h.lock.Unlock()

	h.handler = newHandler
}
