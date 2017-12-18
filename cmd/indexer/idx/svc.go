package idx

import (
	"context"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/gitcmd"
)

func resolveRevision(ctx context.Context, repoURI string, spec string) (*sourcegraph.Repo, string, error) {
	if spec == "" {
		spec = "HEAD"
	}
	repo, err := sourcegraph.InternalClient.ReposGetByURI(ctx, repoURI)
	if err != nil {
		return nil, "", err
	}
	commit, err := gitcmd.Open(repo).ResolveRevision(ctx, spec)
	if err != nil {
		return nil, "", err
	}
	return repo, string(commit), nil
}
