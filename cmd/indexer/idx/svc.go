package idx

import (
	"context"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

func resolveRevision(ctx context.Context, repoURI api.RepoURI, spec string) (*api.Repo, api.CommitID, error) {
	if spec == "" {
		spec = "HEAD"
	}
	repo, err := api.InternalClient.ReposGetByURI(ctx, repoURI)
	if err != nil {
		return nil, "", err
	}

	commit, err := git.ResolveRevision(ctx, gitserver.Repo{Name: repo.URI}, nil, spec, nil)
	if err != nil {
		return nil, "", err
	}
	return repo, commit, nil
}
