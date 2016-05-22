package localstore

import (
	"strings"

	"gopkg.in/inconshreveable/log15.v2"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/github"
)

const gitHubPublicRepoQuery = `SELECT repo.* FROM repo
				WHERE ((NOT blocked)) AND ((NOT fork)) AND (NOT private)
				ORDER BY repo.updated_at desc NULLS LAST
				LIMIT $1 OFFSET $2`

// GitHubRepoGetter is useful for mocking the GitHub API functionality.
type GitHubRepoGetter interface {
	Get(context.Context, string) (*sourcegraph.RemoteRepo, error)
}

var repoGetter GitHubRepoGetter = &github.Repos{}

// listAllPublicGitHubPublic is a special case repos.List specifically for use by the sitemap.
//
// KLUDGE: Normally, we would not want to return any repos with a URI that starts with github.com/
// because we can't guarantee that the metadata in the database currently reflects the
// actual state of the repo (specifically whether it is public or private). This function is
// safe because we explicitly call out to the github API to check for each repo's public/private
// status, but should be regarded as a hack and a better solution must be sought.
func (r *repos) listAllGitHubPublic(ctx context.Context, opt *sourcegraph.RepoListOptions) ([]*sourcegraph.Repo, error) {
	var dbRepos []*dbRepo
	_, err := appDBH(ctx).Select(&dbRepos, gitHubPublicRepoQuery, opt.PerPageOrDefault(), opt.Offset())
	if err != nil {
		return nil, err
	}

	repos := toRepos(dbRepos)

	return removePrivateGitHubRepos(ctx, repos)
}

func removePrivateGitHubRepos(ctx context.Context, repos []*sourcegraph.Repo) ([]*sourcegraph.Repo, error) {
	var publicRepos []*sourcegraph.Repo
	for _, repo := range repos {
		if strings.HasPrefix(strings.ToLower(repo.URI), "github.com/") {
			r, err := repoGetter.Get(ctx, repo.URI)
			if err != nil {
				if grpc.Code(err) != codes.Unauthenticated &&
					grpc.Code(err) != codes.PermissionDenied &&
					grpc.Code(err) != codes.NotFound {
					// Unexpected error, log it.
					log15.Error("Fetching GitHub repo failed (indicates that the GitHub repo in our DB has been renamed, deleted, or made private on GitHub)", "error", err)
				}

				continue
			}

			if !r.Private {
				publicRepos = append(publicRepos, repo)
			}
		} else {
			publicRepos = append(publicRepos, repo)
		}
	}

	return publicRepos, nil
}
