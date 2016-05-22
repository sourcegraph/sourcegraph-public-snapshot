package httpctx

import (
	"fmt"
	"net/http"

	gcontext "github.com/gorilla/context"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
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
func Base(ctx context.Context) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx2 := FromRequestOrNil(r)
			if ctx2 == nil {
				ctx2 = ctx
			}
			SetForRequest(r, ctx)
			defer gcontext.Clear(r)
			next.ServeHTTP(w, r)
		})
	}
}

// RouteName returns the named route that r is routed to, which is
// taken from the manually provided name (using SetRouteName) or the
// mux router, in that order.
func RouteName(r *http.Request) string {
	name, err := RouteNameOrError(r)
	if err != nil {
		panic(err)
	}
	return name
}

// RouteNameOrError returns the current route name. If no such name is set, it
// returns an error. This should only be used in rare cases over RouteName,
// since the RouteName should always be set. A reasonable use-case is via
// Middlewares which do not rely on mux.
func RouteNameOrError(r *http.Request) (string, error) {
	if name, ok := gcontext.GetOk(r, routeNameKey); ok {
		return name.(string), nil
	}
	if route := mux.CurrentRoute(r); route != nil {
		return route.GetName(), nil
	}
	return "", fmt.Errorf("no route name set for request %s", r.URL.String())
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
