package httpctx

import (
	"net/http"

	"golang.org/x/net/context"

	gcontext "github.com/gorilla/context"
	"github.com/sourcegraph/mux"
)

type key int

const (
	contextKey key = iota
	routeNameKey
)

// SetForRequest sets the context for the HTTP request. It will be
// available for the lifetime of the HTTP request.
//
// Typically this function is called in the following pattern:
//   ctx := FromRequest(r)
//   ctx = modifyContextInSomeWay(ctx)
//   SetForRequest(r, ctx)
//
// If calling this function on a request for which previously there
// was no context, it is the caller's responsibility to clean up the
// gorilla context (e.g., by calling `defer gcontext.Clear(r)`).
func SetForRequest(r *http.Request, ctx context.Context) {
	gcontext.Set(r, contextKey, ctx)
}

// FromRequest returns the context for the specified HTTP request. If
// no context exists, it panics.
func FromRequest(r *http.Request) context.Context {
	if ctx, ok := gcontext.GetOk(r, contextKey); ok {
		return ctx.(context.Context)
	}
	panic("no context set for HTTP request")
}

// FromRequestOrNil returns the context for the specified HTTP
// request, or a nil context.Background() if none is set.
func FromRequestOrNil(r *http.Request) context.Context {
	if ctx, ok := gcontext.GetOk(r, contextKey); ok {
		return ctx.(context.Context)
	}
	return nil
}

// Base is a middleware that sets a context.Context on each HTTP request.
func Base(ctx context.Context) func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	return func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		ctx2 := FromRequestOrNil(r)
		if ctx2 == nil {
			ctx2 = ctx
		}
		SetForRequest(r, ctx)
		next(w, r)
	}
}

// RouteName returns the named route that r is routed to, which is
// taken from the manually provided name (using SetRouteName) or the
// mux router, in that order.
func RouteName(r *http.Request) string {
	if name, ok := gcontext.GetOk(r, routeNameKey); ok {
		return name.(string)
	}
	if route := mux.CurrentRoute(r); route != nil {
		return route.GetName()
	}
	panic("no route name set for request " + r.URL.String())
}

// SetRouteName sets the route name that RouteName will return for
// r. It overrides (or can be used in lieu of) the route name from the
// mux router.
//
// Routes that are not routed via the mux router should use
// SetRouteName to set their route name.
func SetRouteName(r *http.Request, name string) {
	gcontext.Set(r, routeNameKey, name)
}
