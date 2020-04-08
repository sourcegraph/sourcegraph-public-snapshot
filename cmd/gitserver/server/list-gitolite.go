package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitolite"
)

func (s *Server) handleListGitolite(w http.ResponseWriter, r *http.Request) {
	defaultGitolite.listRepos(r.Context(), r.URL.Query().Get("gitolite"), w)
}

var defaultGitolite = gitoliteFetcher{client: gitoliteClient{}}

type gitoliteFetcher struct {
	client iGitoliteClient
}

type iGitoliteClient interface {
	ListRepos(ctx context.Context, host string) ([]*gitolite.Repo, error)
}

// listRepos lists the repos of a Gitolite server reachable at the address in gitoliteHost
func (g gitoliteFetcher) listRepos(ctx context.Context, gitoliteHost string, w http.ResponseWriter) {
	var (
		repos = []*gitolite.Repo{}
		err   error
	)

	if gitoliteHost != "" {
		if repos, err = g.client.ListRepos(ctx, gitoliteHost); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err = json.NewEncoder(w).Encode(repos); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type gitoliteClient struct{}

func (c gitoliteClient) ListRepos(ctx context.Context, host string) ([]*gitolite.Repo, error) {
	return gitolite.NewClient(host).ListRepos(ctx)
}
