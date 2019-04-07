package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/gitolite"
	"github.com/sourcegraph/sourcegraph/schema"
)

func (s *Server) handleListGitolite(w http.ResponseWriter, r *http.Request) {
	defaultGitolite.listRepos(r.Context(), r.URL.Query().Get("gitolite"), w)
}

var defaultGitolite = gitoliteFetcher{client: gitoliteClient{}, config: config{}}

type gitoliteFetcher struct {
	client iGitoliteClient
	config iConfig
}

type iConfig interface {
	Gitolite(ctx context.Context) ([]*schema.GitoliteConnection, error)
}

type iGitoliteClient interface {
	ListRepos(ctx context.Context, host string) ([]*gitolite.Repo, error)
}

// listRepos iterates through all Gitolite configs and, for each, lists the repos for the Gitolite
// host.
func (g gitoliteFetcher) listRepos(ctx context.Context, gitoliteHost string, w http.ResponseWriter) {
	repos := make([]*gitolite.Repo, 0)

	config, err := g.config.Gitolite(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, gconf := range config {
		if gconf.Host != gitoliteHost {
			continue
		}
		rp, err := g.client.ListRepos(ctx, gconf.Host)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		repos = append(repos, rp...)
	}

	if err := json.NewEncoder(w).Encode(repos); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type config struct{}

func (c config) Gitolite(ctx context.Context) ([]*schema.GitoliteConnection, error) {
	return conf.GitoliteConfigs(ctx)
}

type gitoliteClient struct{}

func (c gitoliteClient) ListRepos(ctx context.Context, host string) ([]*gitolite.Repo, error) {
	return gitolite.NewClient(host).ListRepos(ctx)
}
