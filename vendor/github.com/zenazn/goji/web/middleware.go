package web

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

// mLayer is a single middleware stack layer. It contains a canonicalized
// middleware representation, as well as the original function as passed to us.
type mLayer struct {
	fn   func(*C, http.Handler) http.Handler
	orig interface{}
}

// mStack is an entire middleware stack. It contains a slice of middleware
// layers (outermost first) protected by a mutex, a cache of pre-built stack
// instances, and a final routing function.
type mStack struct {
	lock   sync.Mutex
	stack  []mLayer
	pool   *cPool
	router internalRouter
}

type internalRouter interface {
	route(*C, http.ResponseWriter, *http.Request)
}

/*
cStack is a cached middleware stack instance. Constructing a middleware stack
involves a lot of allocations: at the very least each layer will have to close
over the layer after (inside) it and a stack N levels deep will incur at least N
separate allocations. Instead of doing this on every request, we keep a pool of
pre-built stacks around for reuse.
*/
type cStack struct {
	C
	m    http.Handler
	pool *cPool
}

func (s *cStack) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.C = C{}
	s.m.ServeHTTP(w, r)
}
func (s *cStack) ServeHTTPC(c C, w http.ResponseWriter, r *http.Request) {
	s.C = c
	s.m.ServeHTTP(w, r)
}

func (m *mStack) appendLayer(fn interface{}) {
	ml := mLayer{orig: fn}
	switch f := fn.(type) {
	case func(http.Handler) http.Handler:
		ml.fn = func(c *C, h http.Handler) http.Handler {
			return f(h)
		}
	case func(*C, http.Handler) http.Handler:
		ml.fn = f
	default:
		log.Fatalf(`Unknown middleware type %v. Expected a function `+
			`with signature "func(http.Handler) http.Handler" or `+
			`"func(*web.C, http.Handler) http.Handler".`, fn)
	}
	m.stack = append(m.stack, ml)
}

func (m *mStack) findLayer(l interface{}) int {
	for i, middleware := range m.stack {
		if funcEqual(l, middleware.orig) {
			return i
		}
	}
	return -1
}

func (m *mStack) invalidate() {
	m.pool = makeCPool()
}

func (m *mStack) newStack() *cStack {
	cs := cStack{}
	router := m.router

	cs.m = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		router.route(&cs.C, w, r)
	})
	for i := len(m.stack) - 1; i >= 0; i-- {
		cs.m = m.stack[i].fn(&cs.C, cs.m)
	}

	return &cs
}

func (m *mStack) alloc() *cStack {
	p := m.pool
	cs := p.alloc()
	if cs == nil {
		cs = m.newStack()
	}

	cs.pool = p
	return cs
}

func (m *mStack) release(cs *cStack) {
	cs.C = C{}
	if cs.pool != m.pool {
		return
	}
	cs.pool.release(cs)
	cs.pool = nil
}

func (m *mStack) Use(middleware interface{}) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.appendLayer(middleware)
	m.invalidate()
}

func (m *mStack) Insert(middleware, before interface{}) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	i := m.findLayer(before)
	if i < 0 {
		return fmt.Errorf("web: unknown middleware %v", before)
	}

	m.appendLayer(middleware)
	inserted := m.stack[len(m.stack)-1]
	copy(m.stack[i+1:], m.stack[i:])
	m.stack[i] = inserted

	m.invalidate()
	return nil
}

func (m *mStack) Abandon(middleware interface{}) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	i := m.findLayer(middleware)
	if i < 0 {
		return fmt.Errorf("web: unknown middleware %v", middleware)
	}

	copy(m.stack[i:], m.stack[i+1:])
	m.stack = m.stack[:len(m.stack)-1 : len(m.stack)]

	m.invalidate()
	return nil
}
