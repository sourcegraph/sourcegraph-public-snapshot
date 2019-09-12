package app

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gorilla/mux"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/pkg/debugserver"
	"github.com/sourcegraph/sourcegraph/pkg/env"
)

var grafanaURLFromEnv = env.Get("SRC_GRAFANA_SERVER_URL", "", "URL at which Grafana can be reached")

// addDebugHandlers registers the reverse proxies to each services debug
// endpoints.
func addDebugHandlers(r *mux.Router) {
	for _, svc := range debugserver.Services {
		addReverseProxyForService(svc, r)
	}

	// TODO(uwedeportivo): revisit this to make it less brittle (for example derive values from one declared place)
	// NOTE: changing the link names or paths for the prometheus or grafana entries in this index
	// requires making changes in other places so they can run behind a proxy:
	//    - docker-images/grafana/config/grafana.ini: value for root_url key
	//    - docker-images/grafana/config/grafana-single-container.ini: value for root_url key
	//    - cmd/server/shared/shared.go: prometheus entry in Procfile, value for --web.external-url
	//    - dev/prometheus.sh: value for --web.external-url
	//  other deploy contexts also need appropriate adjustments

	if len(grafanaURLFromEnv) > 0 {
		grafanaURL, err := url.Parse(grafanaURLFromEnv)
		if err != nil {
			log.Printf("failed to parse SRC_GRAFANA_SERVER_URL=%s: %v. won't generate Grafana link",
				grafanaURLFromEnv, err)
			// setting to empty string so no link gets created for it below
			grafanaURLFromEnv = ""
		} else {
			addReverseProxyForService(debugserver.Service{
				Name:        "grafana",
				Host:        grafanaURL.Host,
				DefaultPath: "",
			}, r)
		}
	}

	index := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, svc := range debugserver.Services {
			path := "/"
			if svc.DefaultPath != "" {
				path = svc.DefaultPath
			}
			fmt.Fprintf(w, `<a href="%s%s">%s</a><br>`, svc.Name, path, svc.Name)
		}
		fmt.Fprintf(w, `<a href="headers">headers</a><br>`)

		// We do not support cluster deployments yet.
		if len(debugserver.Services) == 0 {
			fmt.Fprintf(w, `Instrumentation endpoint proxying for Sourcegraph cluster deployments is not yet available<br>`)
		}

		if len(grafanaURLFromEnv) > 0 {
			fmt.Fprintf(w, `<a target="_blank" href="grafana">grafana</a><br>`)
		}
	})
	r.Handle("/", adminOnly(index))
}

// addReverseProxyForService registers a reverse proxy for the specified service.
func addReverseProxyForService(svc debugserver.Service, r *mux.Router) {
	prefix := "/" + svc.Name
	r.PathPrefix(prefix).Handler(adminOnly(&httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = "http"
			req.URL.Host = svc.Host
			if i := strings.Index(req.URL.Path, prefix); i >= 0 {
				req.URL.Path = req.URL.Path[i+len(prefix):]
			}
		},
		ErrorLog: log.New(env.DebugOut, fmt.Sprintf("%s debug proxy: ", svc.Name), log.LstdFlags),
	}))
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
