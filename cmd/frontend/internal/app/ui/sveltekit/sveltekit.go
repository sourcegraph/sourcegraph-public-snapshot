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
    "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/ui/sveltekit/tags"
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
	enabledRoutes []route
	enabled bool
}

type RouteRegistry struct {
	globalRoutes []string
}

func (r *RouteRegistry) MiddlewareFunc() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			ff := featureflag.FromContext(ctx)

			enabledRoutes := make([]route, 0, len(routes))
			enabled := false

			availabilityMask := tags.EnableAlways
			if ff.GetBoolOr("web-next", false) {
				availabilityMask |= tags.EnableOptIn
			}
			if ff.GetBoolOr("web-next-rollout", false) {
				availabilityMask |= tags.EnableRollout
			}

			for _, route := range routes {
				if route.Tag&availabilityMask != 0 {
					enabledRoutes = append(enabledRoutes, route)

					if (!enabled && r.matches(req, route)) {
						enabled = true
					}
				}
			}

			value := &contextValue{enabledRoutes: enabledRoutes, registry: r, enabled: enabled}
			next.ServeHTTP(w, req.WithContext(context.WithValue(req.Context(), contextKey{}, value)))
		})
	}
}

func (r *RouteRegistry) AddGlobalRoute(route string) {
	r.globalRoutes = append(r.globalRoutes, route)
}

func (r *RouteRegistry) matches(req *http.Request, route route) bool {
	if route.Pattern.MatchString(req.URL.Path) {
		// TODO: check RepoRoot tag
		return true
	}
	return false
}

func NewRouteRegistry() *RouteRegistry {
	return &RouteRegistry{
	}
}

func fromContext(ctx context.Context) *contextValue {
	v := ctx.Value(contextKey{})
	if v == nil {
		return nil
	}
	return v.(*contextValue)
}

// Enabled returns true if the route is configured to be supported by useSvelteKit
func Enabled(r *http.Request) bool {
	ctx := fromContext(r.Context())
	if ctx == nil {
		return false
	}
	return ctx.enabled
}

// SvelteKitJSContext is the context object that is passed to the client apps
// to determine which routes are enabled for SvelteKit.
func GetJSContext(r *http.Request) JSContext {
	ctx := fromContext(r.Context())

	if ctx == nil {
		return JSContext{}
	}

	ff := featureflag.FromContext(r.Context())

	enabledRoutes := make([]string, 0, len(ctx.enabledRoutes))
	availableRoutes := make([]string, 0, len(routes))

	for _, route := range ctx.enabledRoutes {
		enabledRoutes = append(enabledRoutes, route.Pattern.String())
	}

	if ff.GetBoolOr("web-next-toggle", false) {
		for _, route := range routes {
			availableRoutes = append(availableRoutes, route.Pattern.String())
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
