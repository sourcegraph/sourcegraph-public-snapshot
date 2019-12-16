package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/lsifserver/client"
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
		root := q.Get("root")
		ctx := r.Context()

		if !ensureRepoAndCommitExist(ctx, w, repoName, commit) {
			return
		}

		// 🚨 SECURITY: Ensure we return before proxying to the lsif-server upload
		// endpoint. This endpoint is unprotected, so we need to make sure the user
		// provides a valid token proving contributor access to the repository.
		if conf.Get().LsifEnforceAuth && !enforceAuth(ctx, w, r, repoName) {
			return
		}

		jobID, queued, err := client.DefaultClient.Upload(ctx, &struct {
			RepoName string
			Commit   graphqlbackend.GitObjectID
			Root     string
			Blocking *bool
			MaxWait  *int32
			Body     io.ReadCloser
		}{
			RepoName: repoName,
			Commit:   graphqlbackend.GitObjectID(commit),
			Root:     root,
			Body:     r.Body,
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		payload, err := json.Marshal(map[string]string{"id": jobID})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if queued {
			w.WriteHeader(http.StatusAccepted)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		_, _ = w.Write(payload)
	}
}

func ensureRepoAndCommitExist(ctx context.Context, w http.ResponseWriter, repoName, commit string) bool {
	repo, err := backend.Repos.GetByName(ctx, api.RepoName(repoName))
	if err != nil {
		if errcode.IsNotFound(err) {
			http.Error(w, fmt.Sprintf("unknown repository %q", repoName), http.StatusNotFound)
			return false
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return false
	}

	if _, err := backend.Repos.ResolveRev(ctx, repo, commit); err != nil {
		if gitserver.IsRevisionNotFound(err) {
			http.Error(w, fmt.Sprintf("unknown commit %q", commit), http.StatusNotFound)
			return false
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return false
	}

	return true
}
