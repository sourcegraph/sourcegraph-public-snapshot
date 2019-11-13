package proxy

import (
	"net/http"
	"net/http/httputil"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
)

func uploadProxyHandler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
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
