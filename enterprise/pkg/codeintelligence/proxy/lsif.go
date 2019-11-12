package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
)

var apiURL = url.URL{Scheme: "https", Host: "api.github.com"}

func ProxyHandler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = mux.Vars(r)["rest"]
		p.ServeHTTP(w, r)
	}
}

func UploadProxyHandler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		repoName := q.Get("repository")
		commit := q.Get("commit")

		repo, err := backend.Repos.GetByName(r.Context(), api.RepoName(repoName))
		if err != nil {
			http.Error(w, "Unknown repository.", http.StatusNotFound)
			return
		}

		_, err = backend.Repos.ResolveRev(r.Context(), repo, commit)
		if err != nil {
			http.Error(w, "Unknown commit.", http.StatusNotFound)
			return
		}

		if conf.Get().LsifEnforceAuth {
			err, status := enforceAuth(w, r, repoName)
			if err != nil {
				http.Error(w, err.Error(), status)
			}
		}

		r.URL.Path = "upload"
		p.ServeHTTP(w, r)
	}
}
