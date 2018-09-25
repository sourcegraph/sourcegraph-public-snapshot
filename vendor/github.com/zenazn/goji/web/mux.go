package web

import (
	"net/http"
)

/*
Mux is an HTTP multiplexer, much like net/http's ServeMux.

Routes may be added using any of the various HTTP-method-specific functions.
When processing a request, when iterating in insertion order the first route
that matches both the request's path and method is used.

There are two other differences worth mentioning between web.Mux and
http.ServeMux. First, string patterns (i.e., Sinatra-like patterns) must match
exactly: the "rooted subtree" behavior of ServeMux is not implemented. Secondly,
unlike ServeMux, Mux does not support Host-specific patterns.

If you require any of these features, remember that you are free to mix and
match muxes at any part of the stack.

In order to provide a sane API, many functions on Mux take interface{}'s. This
is obviously not a very satisfying solution, but it's probably the best we can
do for now. Instead of duplicating documentation on each method, the types
accepted by those functions are documented here.

A middleware (the untyped parameter in Use() and Insert()) must be one of the
following types:
	- func(http.Handler) http.Handler
	- func(c *web.C, http.Handler) http.Handler

All of the route-adding functions on Mux take two untyped parameters: pattern
and handler. Pattern will be passed to ParsePattern, which takes a web.Pattern,
a string, or a regular expression (more information can be found in the
ParsePattern documentation). Handler must be one of the following types:
	- http.Handler
	- web.Handler
	- func(w http.ResponseWriter, r *http.Request)
	- func(c web.C, w http.ResponseWriter, r *http.Request)
*/
type Mux struct {
	ms mStack
	rt router
}

// New creates a new Mux without any routes or middleware.
func New() *Mux {
	mux := Mux{
		ms: mStack{
			stack: make([]mLayer, 0),
			pool:  makeCPool(),
		},
		rt: router{
			routes:   make([]route, 0),
			notFound: parseHandler(http.NotFound),
		},
	}
	mux.ms.router = &mux.rt
	return &mux
}

// ServeHTTP processes HTTP requests. It make Muxes satisfy net/http.Handler.
func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stack := m.ms.alloc()
	stack.ServeHTTP(w, r)
	m.ms.release(stack)
}

// ServeHTTPC creates a context dependent request with the given Mux. Satisfies
// the web.Handler interface.
func (m *Mux) ServeHTTPC(c C, w http.ResponseWriter, r *http.Request) {
	stack := m.ms.alloc()
	stack.ServeHTTPC(c, w, r)
	m.ms.release(stack)
}

// Middleware Stack functions

// Append the given middleware to the middleware stack. See the documentation
// for type Mux for a list of valid middleware types.
//
// No attempt is made to enforce the uniqueness of middlewares. It is illegal to
// call this function concurrently with active requests.
func (m *Mux) Use(middleware interface{}) {
	m.ms.Use(middleware)
}

// Insert the given middleware immediately before a given existing middleware in
// the stack. See the documentation for type Mux for a list of valid middleware
// types. Returns an error if no middleware has the name given by "before."
//
// No attempt is made to enforce the uniqueness of middlewares. If the insertion
// point is ambiguous, the first (outermost) one is chosen. It is illegal to
// call this function concurrently with active requests.
func (m *Mux) Insert(middleware, before interface{}) error {
	return m.ms.Insert(middleware, before)
}

// Remove the given middleware from the middleware stack. Returns an error if
// no such middleware can be found.
//
// If the name of the middleware to delete is ambiguous, the first (outermost)
// one is chosen. It is illegal to call this function concurrently with active
// requests.
func (m *Mux) Abandon(middleware interface{}) error {
	return m.ms.Abandon(middleware)
}

// Router functions

/*
Dispatch to the given handler when the pattern matches, regardless of HTTP
method. See the documentation for type Mux for a description of what types are
accepted for pattern and handler.

This method is commonly used to implement sub-routing: an admin application, for
instance, can expose a single handler that is attached to the main Mux by
calling Handle("/admin*", adminHandler) or similar. Note that this function
doesn't strip this prefix from the path before forwarding it on (e.g., the
handler will see the full path, including the "/admin" part), but this
functionality can easily be performed by an extra middleware layer.
*/
func (m *Mux) Handle(pattern interface{}, handler interface{}) {
	m.rt.handleUntyped(pattern, mALL, handler)
}

// Dispatch to the given handler when the pattern matches and the HTTP method is
// CONNECT. See the documentation for type Mux for a description of what types
// are accepted for pattern and handler.
func (m *Mux) Connect(pattern interface{}, handler interface{}) {
	m.rt.handleUntyped(pattern, mCONNECT, handler)
}

// Dispatch to the given handler when the pattern matches and the HTTP method is
// DELETE. See the documentation for type Mux for a description of what types
// are accepted for pattern and handler.
func (m *Mux) Delete(pattern interface{}, handler interface{}) {
	m.rt.handleUntyped(pattern, mDELETE, handler)
}

// Dispatch to the given handler when the pattern matches and the HTTP method is
// GET. See the documentation for type Mux for a description of what types are
// accepted for pattern and handler.
//
// All GET handlers also transparently serve HEAD requests, since net/http will
// take care of all the fiddly bits for you. If you wish to provide an alternate
// implementation of HEAD, you should add a handler explicitly and place it
// above your GET handler.
func (m *Mux) Get(pattern interface{}, handler interface{}) {
	m.rt.handleUntyped(pattern, mGET|mHEAD, handler)
}

// Dispatch to the given handler when the pattern matches and the HTTP method is
// HEAD. See the documentation for type Mux for a description of what types are
// accepted for pattern and handler.
func (m *Mux) Head(pattern interface{}, handler interface{}) {
	m.rt.handleUntyped(pattern, mHEAD, handler)
}

// Dispatch to the given handler when the pattern matches and the HTTP method is
// OPTIONS. See the documentation for type Mux for a description of what types
// are accepted for pattern and handler.
func (m *Mux) Options(pattern interface{}, handler interface{}) {
	m.rt.handleUntyped(pattern, mOPTIONS, handler)
}

// Dispatch to the given handler when the pattern matches and the HTTP method is
// PATCH. See the documentation for type Mux for a description of what types are
// accepted for pattern and handler.
func (m *Mux) Patch(pattern interface{}, handler interface{}) {
	m.rt.handleUntyped(pattern, mPATCH, handler)
}

// Dispatch to the given handler when the pattern matches and the HTTP method is
// POST. See the documentation for type Mux for a description of what types are
// accepted for pattern and handler.
func (m *Mux) Post(pattern interface{}, handler interface{}) {
	m.rt.handleUntyped(pattern, mPOST, handler)
}

// Dispatch to the given handler when the pattern matches and the HTTP method is
// PUT. See the documentation for type Mux for a description of what types are
// accepted for pattern and handler.
func (m *Mux) Put(pattern interface{}, handler interface{}) {
	m.rt.handleUntyped(pattern, mPUT, handler)
}

// Dispatch to the given handler when the pattern matches and the HTTP method is
// TRACE. See the documentation for type Mux for a description of what types are
// accepted for pattern and handler.
func (m *Mux) Trace(pattern interface{}, handler interface{}) {
	m.rt.handleUntyped(pattern, mTRACE, handler)
}

// Set the fallback (i.e., 404) handler for this mux. See the documentation for
// type Mux for a description of what types are accepted for handler.
//
// As a convenience, the context environment variable "goji.web.validMethods"
// (also available as the constant ValidMethodsKey) will be set to the list of
// HTTP methods that could have been routed had they been provided on an
// otherwise identical request.
func (m *Mux) NotFound(handler interface{}) {
	m.rt.notFound = parseHandler(handler)
}

// Compile the list of routes into bytecode. This only needs to be done once
// after all the routes have been added, and will be called automatically for
// you (at some performance cost on the first request) if you do not call it
// explicitly.
func (m *Mux) Compile() {
	m.rt.compile()
}
