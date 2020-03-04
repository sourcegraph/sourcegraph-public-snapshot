package app

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/ericchiang/k8s"
	"github.com/gorilla/mux"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/errorutil"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"gopkg.in/inconshreveable/log15.v2"
)

var grafanaURLFromEnv = env.Get("GRAFANA_SERVER_URL", "", "URL at which Grafana can be reached")

func addNoGrafanaHandler(r *mux.Router) {
	noGrafana := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `Grafana endpoint proxying: Please set env var GRAFANA_SERVER_URL`)
	})
	r.Handle("/grafana", adminOnly(noGrafana))
}

func addNoK8sClientHandler(r *mux.Router) {
	noHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `Cluster information not available`)
	})
	r.Handle("/", adminOnly(noHandler))
}

// addDebugHandlers registers the reverse proxies to each services debug
// endpoints.
func addDebugHandlers(r *mux.Router) {
	if len(debugserver.Services) > 0 {
		for _, svc := range debugserver.Services {
			addReverseProxyForService(svc, r)
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
		})
		r.Handle("/", adminOnly(index))
	} else {
		// TODO(uwedeportivo): check we're in cluster deployment
		client, err := k8s.NewInClusterClient()
		if err != nil {
			log15.Error("failed to create k8s client", "error", err)
			addNoK8sClientHandler(r)
			return
		}

		ci := newClusterInstrumenter(client)

		r.Handle("/", adminOnly(http.HandlerFunc(ci.ServeIndex)))
		r.PathPrefix("/cluster").Handler(adminOnly(errorutil.Handler(ci.ServeReverseProxy)))
	}

	// TODO(uwedeportivo): handle grafana in cluster deployment so no need for env var anymore
	if len(grafanaURLFromEnv) > 0 {
		grafanaURL, err := url.Parse(grafanaURLFromEnv)
		if err != nil {
			log.Printf("failed to parse GRAFANA_SERVER_URL=%s: %v",
				grafanaURLFromEnv, err)
			addNoGrafanaHandler(r)
		} else {
			addReverseProxyForService(debugserver.Service{
				Name:        "grafana",
				Host:        grafanaURL.Host,
				DefaultPath: "",
			}, r)
		}
	} else {
		addNoGrafanaHandler(r)
	}
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
