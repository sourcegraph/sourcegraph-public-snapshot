package backend

import (
	"strings"

	gogithub "github.com/sourcegraph/go-github/github"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/github"
)

// This file deals with remote repos (e.g., GitHub repos) that are not
// persisted locally.

var getGitHubRepo = (&github.Repos{}).Get

func (s *repos) Resolve(ctx context.Context, op *sourcegraph.RepoResolveOp) (*sourcegraph.RepoResolution, error) {
	// First, look up locally.
	repo, err := store.ReposFromContext(ctx).GetByURI(ctx, op.Path)
	if err == nil {
		return &sourcegraph.RepoResolution{Repo: repo.URI, CanonicalPath: repo.URI}, nil
	} else if errcode.GRPC(err) == codes.NotFound {
		// Next, see if it's a GitHub repo.
		repo, err := getGitHubRepo(ctx, op.Path)
		if err == nil {
			// If canonical location differs, try looking up locally at canonical location.
			if canonicalPath := "github.com/" + repo.Owner + "/" + repo.Name; op.Path != canonicalPath {
				if repo, err := store.ReposFromContext(ctx).GetByURI(ctx, canonicalPath); err == nil {
					return &sourcegraph.RepoResolution{Repo: repo.URI, CanonicalPath: repo.URI}, nil
				}
			}

			if op.Remote {
				return &sourcegraph.RepoResolution{RemoteRepo: repo}, nil
			}
			return nil, grpc.Errorf(codes.NotFound, "resolved repo not found locally: %s", op.Path)
		} else if errcode.GRPC(err) == codes.NotFound {
			if strings.HasPrefix(op.Path, "gopkg.in/") && op.Remote {
				return &sourcegraph.RepoResolution{
					RemoteRepo: &sourcegraph.RemoteRepo{HTTPCloneURL: "https://" + op.Path},
				}, nil
			}
			return nil, grpc.Errorf(codes.NotFound, "repo %q not found locally or remotely", op.Path)
		} else {
			return nil, err
		}
	}
	return nil, err
}

func (s *repos) ListRemote(ctx context.Context, opt *sourcegraph.ReposListRemoteOptions) (*sourcegraph.RemoteRepoList, error) {
	repos, err := (&github.Repos{}).ListAccessible(ctx, &gogithub.RepositoryListOptions{
		ListOptions: gogithub.ListOptions{
			PerPage: int(opt.ListOptions.PerPage),
			Page:    int(opt.ListOptions.Page),
		},
	})
	if err != nil {
		return nil, err
	}
	return &sourcegraph.RemoteRepoList{RemoteRepos: repos}, nil
}
