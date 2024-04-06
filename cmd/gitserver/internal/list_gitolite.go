package internal

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitolite"
	"github.com/sourcegraph/sourcegraph/internal/security"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var defaultGitolite = gitoliteFetcher{client: gitoliteClient{}}

type gitoliteFetcher struct {
	client gitoliteRepoLister
}

type gitoliteRepoLister interface {
	ListRepos(ctx context.Context, host string) ([]*gitolite.Repo, error)
}

// listRepos lists the repos of a Gitolite server reachable at the address in gitoliteHost
func (g gitoliteFetcher) listRepos(ctx context.Context, gitoliteHost string) ([]*gitolite.Repo, error) {
	var (
		repos []*gitolite.Repo
		err   error
	)

	// 🚨 SECURITY: If gitoliteHost is a non-empty string that fails hostname validation, return an error
	if gitoliteHost != "" && !security.ValidateRemoteAddr(gitoliteHost) {
		return nil, errors.New("invalid gitolite host")
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
