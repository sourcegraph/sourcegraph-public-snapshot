package authzchecked

import (
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/store"
)

// RepoConfigs wraps base's methods with authorization checks.
func RepoConfigs(base store.RepoConfigs) store.RepoConfigs { return &repoConfigs{base} }

// repoConfigs adds authorization checks to an underlying RepoConfigs.
type repoConfigs struct {
	noauthz store.RepoConfigs
}

func (s *repoConfigs) Get(ctx context.Context, repo string) (*sourcegraph.RepoConfig, error) {
	if err := auth.CheckRepo(ctx, repo, auth.Read); err != nil {
		return nil, err
	}
	return s.noauthz.Get(ctx, repo)
}

func (s *repoConfigs) Update(ctx context.Context, repo string, settings sourcegraph.RepoConfig) error {
	if err := auth.CheckRepo(ctx, repo, auth.Admin); err != nil {
		return err
	}
	return s.noauthz.Update(ctx, repo, settings)
}
