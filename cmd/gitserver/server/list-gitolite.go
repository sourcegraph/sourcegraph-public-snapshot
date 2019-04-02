//go:generate env GOBIN=$PWD/.bin GO111MODULE=on go install github.com/golang/mock/mockgen
//go:generate $PWD/.bin/mockgen -destination mock_server/mocks.go github.com/sourcegraph/sourcegraph/cmd/gitserver/server IConfig,IGitoliteClient
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
	defaultGitolite.listGitolite(r.Context(), r.URL.Query().Get("gitolite"), w)
}

type IConfig interface {
	Gitolite(ctx context.Context) ([]*schema.GitoliteConnection, error)
}

type Config struct{}

func (c Config) Gitolite(ctx context.Context) ([]*schema.GitoliteConnection, error) {
	return conf.GitoliteConfigs(ctx)
}

type IGitoliteClient interface {
	ListRepos(ctx context.Context, host string) ([]*gitolite.Repo, error)
}

type GitoliteClient struct{}

func (c GitoliteClient) ListRepos(ctx context.Context, host string) ([]*gitolite.Repo, error) {
	return gitolite.NewClient(host).ListRepos(ctx)
}

type Gitolite struct {
	client IGitoliteClient
	config IConfig
}

var defaultGitolite = Gitolite{client: GitoliteClient{}, config: Config{}}

// listGitolite is effectively a wrapper around gitolite.Client.ListRepos.  This must currently be
// invoked from gitserver, because only gitserver has the SSH key needed to authenticate to the
// Gitolite API.
func (g Gitolite) listGitolite(ctx context.Context, gitoliteHost string, w http.ResponseWriter) {
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
