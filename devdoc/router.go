package devdoc

import (
	"fmt"
	"net/url"

	"github.com/sourcegraph/mux"
)

const (
	RootRoute      = "devdoc.root"      // route name for root
	StaticRoute    = "devdoc.static"    // route name for static data files
	LibrariesRoute = "devdoc.libraries" // route name for libraries
	APIRoute       = "devdoc.api"       // route name for api
	EnableRoute    = "devdoc.enable"    // route name for enable
	CommunityRoute = "devdoc.community" // route name for community
)

// Router represents a router for the application.
type Router struct {
	r *mux.Router
}

// NewRouter creates a new URL router for a developer application.
func NewRouter(base *mux.Router) *Router {
	if base == nil {
		base = mux.NewRouter()
	}
	base.Path("/").Methods("GET").Name(RootRoute)
	base.PathPrefix("/static/").Methods("GET").Name(StaticRoute)
	base.PathPrefix("/libraries").Methods("GET").Name(LibrariesRoute)
	base.PathPrefix("/api").Methods("GET").Name(APIRoute)
	base.PathPrefix("/enable").Methods("GET").Name(EnableRoute)
	base.PathPrefix("/community").Methods("GET").Name(CommunityRoute)
	return &Router{r: base}
}

// URLTo constructs a URL to a given route.
func (r *Router) URLTo(route string) (*url.URL, error) {
	rt := r.r.Get(route)
	if rt == nil {
		return nil, fmt.Errorf("no such route: %q", route)
	}
	return rt.URL()
}
