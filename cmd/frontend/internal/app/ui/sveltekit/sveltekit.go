package sveltekit

import (
	"bytes"
	"context"
	"html/template"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/ui/assets"
)

// JSContext is the metadata sent to the client apps
type JSContext struct {
	// EnabledRoutes is a list of routes for which the SvelteKit app should be rendered.
	// On the client this list is used to determine whether to perform client side navigation
	// or to perform a full page reload.
	EnabledRoutes []string `json:"enabledRoutes"`

	// AvailableRoutes is a list of all routes that are supported by SvelteKit.
	// This is used in the React app to determine whether to show the button to enable SvelteKit.
	AvailableRoutes []string `json:"availableRoutes"`
}

type contextKey struct{}

type contextValue struct {
	registry *RouteRegistry
	routeMap map[string]struct{}
}

type Availablity int

const (
	// EnableAlways always renders the SvelteKit app for this route
	EnableAlways Availablity = 1 << iota
	// EnableRollout renders the SvelteKit app for this route when the "web-next-rollout" feature flag is enabled
	EnableRollout
	// EnableOptIn renders the SvelteKit app for this route when the "web-next" feature flag is enabled.
	EnableOptIn Availablity = 1<<iota | EnableRollout
)

type RouteRegistry struct {
	routes map[string]Availablity
}

// Register registers a route to be supported by SvelteKit
// availability determines when the server should serve the SvelteKit app for this route
// Default: When the feature flag "web-next" is enabled
// Rollout: When the feature flag "web-next-rollout" is enabled
// Always: Always serve the SvelteKit app for this route
func (r *RouteRegistry) Register(route *mux.Route, enabled Availablity) *mux.Route {
	if enabled == 0 {
		// Default to EnableDefault
		enabled = EnableOptIn
	}
	r.routes[route.GetName()] = enabled
	return route
}

func (r *RouteRegistry) MiddlewareFunc() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			ff := featureflag.FromContext(ctx)

			routeMap := make(map[string]struct{})
			availabilityMask := EnableAlways
			if ff.GetBoolOr("web-next", false) {
				availabilityMask |= EnableOptIn
			}
			if ff.GetBoolOr("web-next-rollout", false) {
				availabilityMask |= EnableRollout
			}

			for route, availability := range r.routes {
				if availability&availabilityMask != 0 {
					routeMap[route] = struct{}{}
				}
			}

			value := &contextValue{routeMap: routeMap, registry: r}
			next.ServeHTTP(w, req.WithContext(context.WithValue(req.Context(), contextKey{}, value)))
		})
	}
}

func NewRouteRegistry() *RouteRegistry {
	return &RouteRegistry{
		routes: make(map[string]Availablity),
	}
}

func fromContext(ctx context.Context) *contextValue {
	v := ctx.Value(contextKey{})
	if v == nil {
		return nil
	}
	return v.(*contextValue)
}

func getRouteMap(ctx context.Context) map[string]struct{} {
	return fromContext(ctx).routeMap
}

// Enabled returns true if the route is configured to be supported by useSvelteKit
func Enabled(r *http.Request) bool {
	routeMap := getRouteMap(r.Context())
	route := mux.CurrentRoute(r)
	if route == nil {
		return false
	}
	_, enabled := routeMap[route.GetName()]
	return enabled
}

// SvelteKitJSContext is the context object that is passed to the client apps
// to determine which routes are enabled for SvelteKit.
func GetJSContext(r *http.Request) JSContext {
	ctx := fromContext(r.Context())

	if ctx == nil {
		return JSContext{}
	}

	ff := featureflag.FromContext(r.Context())

	enabledRoutes := make([]string, 0, len(ctx.routeMap))
	availableRoutes := make([]string, 0, len(ctx.registry.routes))

	for route := range ctx.routeMap {
		enabledRoutes = append(enabledRoutes, route)
	}

	if ff.GetBoolOr("web-next-toggle", false) {
		for route, availablity := range ctx.registry.routes {
			if availablity&EnableOptIn != 0 {
				availableRoutes = append(availableRoutes, route)
			}
		}
	}

	return JSContext{EnabledRoutes: enabledRoutes, AvailableRoutes: availableRoutes}
}

func loadSvelteKitTemplate() (*template.Template, error) {
	fileName := "_sk/index.html"
	file, err := assets.Provider.Assets().Open(fileName)
	if err != nil {
		return nil, errors.Errorf("failed to open %s: %w", fileName, err)
	}
	defer file.Close()
	buf := new(strings.Builder)
	io.Copy(buf, file)

	tmpl, err := template.New(fileName).Parse(buf.String())
	if err != nil {
		return nil, errors.Errorf("failed to parse template %s: %w", fileName, err)
	}
	return tmpl, nil

}

var loadCachedSvelteKitTemplate = sync.OnceValues(loadSvelteKitTemplate)

// RenderTemplate writes SvelteKit's fallback page to the provided writer
func RenderTemplate(dst io.Writer, data any) error {
	tmpl, err := loadCachedSvelteKitTemplate()
	if env.InsecureDev {
		// Load "fresh" template in dev mode
		tmpl, err = loadSvelteKitTemplate()
	}
	if err != nil {
		return err
	}

	// Write to a buffer to avoid a partially written response going to w
	// when an error would occur. Otherwise, our error page template rendering
	// will be corrupted.
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}
	_, err = buf.WriteTo(dst)
	return err
}
