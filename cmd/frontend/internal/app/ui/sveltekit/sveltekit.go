package sveltekit

import (
	"bytes"
	"context"
	"html/template"
	"io"
	"net/http"
	"strings"
	"sync"
	"regexp"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/ui/assets"
    "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/ui/sveltekit/tags"
)

var namedCaptureGroup = regexp.MustCompile(`\(\?P<[^>]+>`)

// JSContext is the metadata sent to the client apps
type JSContext struct {
	// EnabledRoutes is a list of routes for which the SvelteKit app should be rendered.
	// On the client this list is used to determine whether to perform client side navigation
	// or to perform a full page reload.
	EnabledRoutes []int `json:"enabledRoutes"`

	// Whether or not to allow the user to toggle the SvelteKit app on or off.
	ShowToggle bool `json:"showToggle"`

	// KnownRoutes is a list of all routes that are known to the server. The client app
	// needs this list to determine whether a pathname is a repsitory route or not, because
	// the repository route is a catch-all route.
	KnownRoutes []string `json:"knownRoutes"`
}

type contextKey struct{}

type contextValue struct {
	enabledRoutes []int
	knownRoutes   []string
	enabled bool
}

func (r *svelteKitRoute) isRepoRoot() bool {
	return r.Tag&tags.RepoRoot != 0
}

func (r *svelteKitRoute) matches(req *http.Request) bool {
	return r.Pattern.MatchString(req.URL.Path)
}

// RegisterSvelteKit registers a middleware that determines which routes are enabled for SvelteKit.
// It also extends the request context with inormation that is sent to the client apps via JSContext.
func RegisterSvelteKit(r *mux.Router, repoRootRoute *mux.Route) {
	var knownRoutes []string

	r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		regexp, err := route.GetPathRegexp()
		if err != nil {
			return nil
		}

		// Ignore the repository root route, as it is a catch all route and would
		// defeat the purpose of tracking all *other* routes.
		if route == repoRootRoute {
			return nil
		}

		// routes can contain named capture groups which don't work in JavaScript.
		// We replace them with non-capturing groups.
		knownRoutes = append(knownRoutes, namedCaptureGroup.ReplaceAllString(regexp, "(?:"))
		return nil
	})

	r.Use(func (next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			ff := featureflag.FromContext(ctx)
			route := mux.CurrentRoute(req)

			enabledRoutes := make([]int, 0, len(svelteKitRoutes))
			enabled := false

			availabilityMask := tags.EnableAlways
			if ff.GetBoolOr("web-next", false) {
				availabilityMask |= tags.EnableOptIn
			}
			if ff.GetBoolOr("web-next-rollout", false) {
				availabilityMask |= tags.EnableRollout
			}

			for i, skr := range svelteKitRoutes {
				if skr.Tag&availabilityMask != 0 {
					enabledRoutes = append(enabledRoutes, i)

					if !enabled {
						// The repo root is essentially a catch-all route. If it matches we need
						// to make sure that the server also interprets it as the repo root route.
						if skr.matches(req) && (!skr.isRepoRoot() || route == repoRootRoute) {
							enabled = true
						}
					}
				}
			}

			value := &contextValue{enabledRoutes: enabledRoutes, knownRoutes: knownRoutes, enabled: enabled}
			next.ServeHTTP(w, req.WithContext(context.WithValue(req.Context(), contextKey{}, value)))
		})
	})
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

	jsCtx := JSContext{
		ShowToggle: ff.GetBoolOr("web-next-toggle", false),
		KnownRoutes: ctx.knownRoutes,
	}

	for _, routeIndex := range ctx.enabledRoutes {
		jsCtx.EnabledRoutes = append(jsCtx.EnabledRoutes, routeIndex)
	}

	return jsCtx
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
