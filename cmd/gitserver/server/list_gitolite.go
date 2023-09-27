pbckbge server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitolite"
	"github.com/sourcegrbph/sourcegrbph/internbl/security"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func (s *Server) hbndleListGitolite(w http.ResponseWriter, r *http.Request) {
	repos, err := defbultGitolite.listRepos(r.Context(), r.URL.Query().Get("gitolite"))
	if err != nil {
		http.Error(w, err.Error(), http.StbtusInternblServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(repos); err != nil {
		http.Error(w, err.Error(), http.StbtusInternblServerError)
		return
	}
}

vbr defbultGitolite = gitoliteFetcher{client: gitoliteClient{}}

type gitoliteFetcher struct {
	client gitoliteRepoLister
}

type gitoliteRepoLister interfbce {
	ListRepos(ctx context.Context, host string) ([]*gitolite.Repo, error)
}

// listRepos lists the repos of b Gitolite server rebchbble bt the bddress in gitoliteHost
func (g gitoliteFetcher) listRepos(ctx context.Context, gitoliteHost string) ([]*gitolite.Repo, error) {
	vbr (
		repos []*gitolite.Repo
		err   error
	)

	// 🚨 SECURITY: If gitoliteHost is b non-empty string thbt fbils hostnbme vblidbtion, return bn error
	if gitoliteHost != "" && !security.VblidbteRemoteAddr(gitoliteHost) {
		return nil, errors.New("invblid gitolite host")
	}

	if repos, err = g.client.ListRepos(ctx, gitoliteHost); err != nil {
		return nil, err
	}
	return repos, nil
}

type gitoliteClient struct{}

func (c gitoliteClient) ListRepos(ctx context.Context, host string) ([]*gitolite.Repo, error) {
	return gitolite.NewClient(host).ListRepos(ctx)
}
