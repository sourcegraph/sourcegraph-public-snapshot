package idx

import (
	"context"
	"strings"

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
	var remoteURL string
	// If it is possible to 100% correctly determine it statically, use a fast path. This is used
	// to avoid a RepoLookup call for public GitHub.com repositories on Sourcegraph.com, which reduces
	// rate limit pressure significantly.
	if strings.HasPrefix(strings.ToLower(string(repo.URI)), "github.com/") {
		remoteURL = "https://" + string(repo.URI)
	} else {
		repoInfo, err := repoupdater.DefaultClient.RepoLookup(ctx, protocol.RepoLookupArgs{
			Repo:         repo.URI,
			ExternalRepo: repo.ExternalRepo,
		})
		if err != nil {
			return nil, "", err
		}
		remoteURL = repoInfo.Repo.VCS.URL
	}

	commit, err := gitcmd.Open(repoURI, remoteURL).ResolveRevision(ctx, spec)
	if err != nil {
		return nil, "", err
	}
	return repo, commit, nil
}
