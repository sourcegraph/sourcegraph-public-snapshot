package local

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/ext/github"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
)

// This file deals with remote repos (e.g., GitHub repos) that are not
// persisted locally.

var getGitHubRepo = (&github.Repos{}).Get

func (s *repos) Resolve(ctx context.Context, op *sourcegraph.RepoResolveOp) (*sourcegraph.RepoResolution, error) {
	// First, look up locally.
	repo, err := store.ReposFromContext(ctx).Get(ctx, op.Path)
	if err == nil {
		repoSpec := repo.RepoSpec()
		return &sourcegraph.RepoResolution{Result: &sourcegraph.RepoResolution_Repo{Repo: &repoSpec}}, nil
	} else if errcode.GRPC(err) == codes.NotFound {
		// Next, see if it's a GitHub repo.
		repo, err := getGitHubRepo(ctx, op.Path)
		if err == nil {
			return &sourcegraph.RepoResolution{
				Result: &sourcegraph.RepoResolution_RemoteRepo{RemoteRepo: repo},
			}, nil
		} else if errcode.GRPC(err) == codes.NotFound {
			return nil, grpc.Errorf(codes.NotFound, "repo %q not found locally or remotely", op.Path)
		} else {
			return nil, err
		}
	}
	return nil, err
}
