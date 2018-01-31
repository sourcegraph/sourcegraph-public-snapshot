package idx

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/repoupdater"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/gitcmd"
)

func resolveRevision(ctx context.Context, repoURI api.RepoURI, spec string) (*api.Repo, api.CommitID, error) {
	if spec == "" {
		spec = "HEAD"
	}
	repo, err := api.InternalClient.ReposGetByURI(ctx, repoURI)
	if err != nil {
		return nil, "", err
	}
	// Get the repository's remote URL so that gitserver can update the repository from the remote, in case
	// we're asked to index a commit that isn't yet on gitserver.
	repoInfo, err := repoupdater.DefaultClient.RepoLookup(ctx, protocol.RepoLookupArgs{
		Repo:         repo.URI,
		ExternalRepo: repo.ExternalRepo,
	})
	if err != nil {
		return nil, "", err
	}
	commit, err := gitcmd.Open(repoURI, repoInfo.Repo.VCS.URL).ResolveRevision(ctx, spec)
	if err != nil {
		return nil, "", err
	}
	return repo, commit, nil
}
