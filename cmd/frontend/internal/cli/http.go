package cli

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/NYTimes/gziphandler"
	gcontext "github.com/gorilla/context"
	"github.com/gorilla/mux"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/assets"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/cli/middleware"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/globals"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi/router"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	tracepkg "sourcegraph.com/sourcegraph/sourcegraph/pkg/trace"
)

// newExternalHTTPHandler creates and returns the HTTP handler that serves the app and API pages to
// external clients.
func newExternalHTTPHandler(ctx context.Context) (http.Handler, error) {
	sm := http.NewServeMux()
	sm.Handle("/.api/", gziphandler.GzipHandler(httpapi.NewHandler(router.New(mux.NewRouter().PathPrefix("/.api/").Subrouter()))))
	sm.Handle("/", handlerutil.NewHandlerWithCSRFProtection(app.NewHandler(), globals.AppURL.Scheme == "https"))
	assets.Mount(sm)

	var h http.Handler = sm
	h = middleware.SourcegraphComGoGetHandler(h)
	h = middleware.BlackHole(h)
	h = tracepkg.Middleware(h)
	h = secureHeadersMiddleware(h)
	// 🚨 SECURITY: Verify user identity if required
	var err error
	h, err = auth.NewAuthHandler(ctx, h, appURL)
	if err != nil {
		return nil, err
	}
	// Don't leak memory through gorilla/session items stored in context
	h = gcontext.ClearHandler(h)
	return h, nil
}

func secureHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// headers for security
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Frame-Options", "DENY")
		if v, _ := strconv.ParseBool(enableHSTS); v {
			w.Header().Set("Strict-Transport-Security", "max-age=8640000")
		}
		// no cache by default
		w.Header().Set("Cache-Control", "no-cache, max-age=0")

		// CORS
		// If the headerOrigin is the development or production Chrome Extension explictly set the Allow-Control-Allow-Origin
		// to the incoming header URL. Otherwise use the configured CORS origin.
		headerOrigin := r.Header.Get("Origin")
		isExtensionRequest := (headerOrigin == devExtension || headerOrigin == prodExtension) && !disableBrowserExtension
		if corsOrigin := conf.Get().CorsOrigin; corsOrigin != "" || isExtensionRequest {
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			allowOrigin := corsOrigin
			if isExtensionRequest || isAllowedOrigin(headerOrigin, strings.Fields(corsOrigin)) {
				allowOrigin = headerOrigin
			}

			w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
			if r.Method == "OPTIONS" {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With, X-Sourcegraph-Client, Content-Type")
				w.WriteHeader(http.StatusOK)
				return // do not invoke next handler
			}
		}

		next.ServeHTTP(w, r)
	})
}
