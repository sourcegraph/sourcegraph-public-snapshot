package httpctx

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type key int

const (
	routeNameKey key = iota
)

// Base is a middleware that sets a context.Context on each HTTP request.
func Base(ctx context.Context) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(ctx)) // TODO this is bad, we're losing the request's own context
		})
	}
}

// RouteName returns the named route that r is routed to, which is
// taken from the manually provided name (using WithRouteName) or the
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
	if name := r.Context().Value(routeNameKey); name != nil {
		return name.(string), nil
	}
	if route := mux.CurrentRoute(r); route != nil {
		return route.GetName(), nil
	}
	return "", fmt.Errorf("no route name set for request %s", r.URL.String())
}

// WithRouteName sets the route name that RouteName will return for
// r. It overrides (or can be used in lieu of) the route name from the
// mux router.
//
// Routes that are not routed via the mux router should use
// WithRouteName to set their route name.
func WithRouteName(r *http.Request, name string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), routeNameKey, name))
}
