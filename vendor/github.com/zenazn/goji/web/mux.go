package web

import (
	"net/http"
)

/*
Mux is an HTTP multiplexer, much like net/http's ServeMux. It functions as both
a middleware stack and as an HTTP router.

Middleware provide a great abstraction for actions that must be performed on
every request, such as request logging and authentication. To append, insert,
and remove middleware, you can call the Use, Insert, and Abandon functions
respectively.

Routes may be added using any of the HTTP verb functions (Get, Post, etc.), or
through the generic Handle function. Goji's routing algorithm is very simple:
routes are processed in the order they are added, and the first matching route
will be executed. Routes match if their HTTP method and Pattern both match.
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

// ServeHTTP processes HTTP requests. Satisfies net/http.Handler.
func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stack := m.ms.alloc()
	stack.ServeHTTP(w, r)
	m.ms.release(stack)
}

// ServeHTTPC creates a context dependent request with the given Mux. Satisfies
// the Handler interface.
func (m *Mux) ServeHTTPC(c C, w http.ResponseWriter, r *http.Request) {
	stack := m.ms.alloc()
	stack.ServeHTTPC(c, w, r)
	m.ms.release(stack)
}

// Middleware Stack functions

// Use appends the given middleware to the middleware stack.
//
// No attempt is made to enforce the uniqueness of middlewares. It is illegal to
// call this function concurrently with active requests.
func (m *Mux) Use(middleware MiddlewareType) {
	m.ms.Use(middleware)
}

// Insert inserts the given middleware immediately before a given existing
// middleware in the stack. Returns an error if "before" cannot be found in the
// current stack.
//
// No attempt is made to enforce the uniqueness of middlewares. If the insertion
// point is ambiguous, the first (outermost) one is chosen. It is illegal to
// call this function concurrently with active requests.
func (m *Mux) Insert(middleware, before MiddlewareType) error {
	return m.ms.Insert(middleware, before)
}

// Abandon removes the given middleware from the middleware stack. Returns an
// error if no such middleware can be found.
//
// If the name of the middleware to delete is ambiguous, the first (outermost)
// one is chosen. It is illegal to call this function concurrently with active
// requests.
func (m *Mux) Abandon(middleware MiddlewareType) error {
	return m.ms.Abandon(middleware)
}

// Router functions

type routerMiddleware struct {
	m *Mux
	c *C
	h http.Handler
}

func (rm routerMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if rm.c.Env == nil {
		rm.c.Env = make(map[interface{}]interface{}, 1)
	}
	rm.c.Env[MatchKey] = rm.m.rt.getMatch(rm.c, w, r)
	rm.h.ServeHTTP(w, r)
}

/*
Router is a middleware that performs routing and stores the resulting Match in
Goji's environment. If a routing Match is present at the end of the middleware
stack, that Match is used instead of re-routing.

This middleware is especially useful to create post-routing middleware, e.g. a
request logger which prints which pattern or handler was selected, or an
authentication middleware which only applies to certain routes.

If you use nested Muxes with explicit routing, you should be aware that the
explicit routing information set by an outer Mux can be picked up by an inner
Mux, inadvertently causing an infinite routing loop. If you use both explicit
routing and nested Muxes, you should be sure to unset MatchKey before the inner
Mux performs routing (or attach a Router to the inner Mux as well).
*/
func (m *Mux) Router(c *C, h http.Handler) http.Handler {
	return routerMiddleware{m, c, h}
}

/*
Handle dispatches to the given handler when the pattern matches, regardless of
HTTP method.

This method is commonly used to implement sub-routing: an admin application, for
instance, can expose a single handler that is attached to the main Mux by
calling Handle("/admin/*", adminHandler) or similar. Note that this function
doesn't strip this prefix from the path before forwarding it on (e.g., the
handler will see the full path, including the "/admin/" part), but this
functionality can easily be performed by an extra middleware layer.
*/
func (m *Mux) Handle(pattern PatternType, handler HandlerType) {
	m.rt.handleUntyped(pattern, mALL, handler)
}

// Connect dispatches to the given handler when the pattern matches and the HTTP
// method is CONNECT.
func (m *Mux) Connect(pattern PatternType, handler HandlerType) {
	m.rt.handleUntyped(pattern, mCONNECT, handler)
}

// Delete dispatches to the given handler when the pattern matches and the HTTP
// method is DELETE.
func (m *Mux) Delete(pattern PatternType, handler HandlerType) {
	m.rt.handleUntyped(pattern, mDELETE, handler)
}

// Get dispatches to the given handler when the pattern matches and the HTTP
// method is GET.
//
// All GET handlers also transparently serve HEAD requests, since net/http will
// take care of all the fiddly bits for you. If you wish to provide an alternate
// implementation of HEAD, you should add a handler explicitly and place it
// above your GET handler.
func (m *Mux) Get(pattern PatternType, handler HandlerType) {
	m.rt.handleUntyped(pattern, mGET|mHEAD, handler)
}

// Head dispatches to the given handler when the pattern matches and the HTTP
// method is HEAD.
func (m *Mux) Head(pattern PatternType, handler HandlerType) {
	m.rt.handleUntyped(pattern, mHEAD, handler)
}

// Options dispatches to the given handler when the pattern matches and the HTTP
// method is OPTIONS.
func (m *Mux) Options(pattern PatternType, handler HandlerType) {
	m.rt.handleUntyped(pattern, mOPTIONS, handler)
}

// Patch dispatches to the given handler when the pattern matches and the HTTP
// method is PATCH.
func (m *Mux) Patch(pattern PatternType, handler HandlerType) {
	m.rt.handleUntyped(pattern, mPATCH, handler)
}

// Post dispatches to the given handler when the pattern matches and the HTTP
// method is POST.
func (m *Mux) Post(pattern PatternType, handler HandlerType) {
	m.rt.handleUntyped(pattern, mPOST, handler)
}

// Put dispatches to the given handler when the pattern matches and the HTTP
// method is PUT.
func (m *Mux) Put(pattern PatternType, handler HandlerType) {
	m.rt.handleUntyped(pattern, mPUT, handler)
}

// Trace dispatches to the given handler when the pattern matches and the HTTP
// method is TRACE.
func (m *Mux) Trace(pattern PatternType, handler HandlerType) {
	m.rt.handleUntyped(pattern, mTRACE, handler)
}

// NotFound sets the fallback (i.e., 404) handler for this mux.
//
// As a convenience, the context environment variable "goji.web.validMethods"
// (also available as the constant ValidMethodsKey) will be set to the list of
// HTTP methods that could have been routed had they been provided on an
// otherwise identical request.
func (m *Mux) NotFound(handler HandlerType) {
	m.rt.notFound = parseHandler(handler)
}

// Compile compiles the list of routes into bytecode. This only needs to be done
// once after all the routes have been added, and will be called automatically
// for you (at some performance cost on the first request) if you do not call it
// explicitly.
func (m *Mux) Compile() {
	m.rt.compile()
}
