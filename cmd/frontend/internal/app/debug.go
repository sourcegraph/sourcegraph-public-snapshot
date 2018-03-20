package app

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/debugserver"
)

// addDebugHandlers registers the reverse proxies to each services debug
// endpoints.
func addDebugHandlers(r *mux.Router) {
	for _, svc := range debugserver.Services {
		svc := svc
		prefix := "/" + svc.Name
		r.PathPrefix(prefix).Handler(adminOnly(&httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.URL.Scheme = "http"
				req.URL.Host = svc.Host
				if i := strings.Index(req.URL.Path, prefix); i >= 0 {
					req.URL.Path = req.URL.Path[i+len(prefix):]
				}
			},
		}))
	}

	index := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, svc := range debugserver.Services {
			fmt.Fprintf(w, `<a href="%s/">%s</a><br>`, svc.Name, svc.Name)
		}
		fmt.Fprintf(w, `<a href="headers">headers</a><br>`)

		// We do not support data center yet
		if len(debugserver.Services) == 0 {
			fmt.Fprintf(w, `Instrumentation endpoint proxying for Sourcegraph Data-Center coming soon...<br>`)
		}
	})
	r.Handle("/", adminOnly(index))
}

// adminOnly is a HTTP middleware which only allows requests by admins.
func adminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := backend.CheckCurrentUserIsSiteAdmin(r.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
