package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

func uploadProxyHandler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		repoName := q.Get("repository")
		commit := q.Get("commit")

		repo, err := backend.Repos.GetByName(r.Context(), api.RepoName(repoName))
		if err != nil {
			if errcode.IsNotFound(err) {
				http.Error(w, fmt.Sprintf("unknown repository %q", repoName), http.StatusNotFound)
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = backend.Repos.ResolveRev(r.Context(), repo, commit)
		if err != nil {
			if gitserver.IsRevisionNotFound(err) {
				http.Error(w, fmt.Sprintf("unknown commit %q", commit), http.StatusNotFound)
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if conf.Get().LsifEnforceAuth {
			// 🚨 SECURITY: Ensure we return before proxying to the lsif-server upload
			// endpoint. This endpoint is unprotected, so we need to make sure the user
			// provides a valid token proving contributor access to the repository.
			if err, status := enforceAuth(w, r, repoName); err != nil {
				http.Error(w, err.Error(), status)
				return
			}
		}

		r.URL.Path = "upload"
		p.ServeHTTP(w, r)
	}
}
