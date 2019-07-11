package httpapi

import (
	"net/http"
	"net/http/httputil"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
)

func proxyHandlerLSIF(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = mux.Vars(r)["rest"]
		p.ServeHTTP(w, r)
	}
}

func proxyHandlerLSIFUpload(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := backend.CheckCurrentUserIsSiteAdmin(r.Context()); err != nil {
			http.Error(w, "Only admins are allowed to upload LSIF data.", http.StatusUnauthorized)
			return
		}
		r.URL.Path = "upload"
		p.ServeHTTP(w, r)
	}
}
