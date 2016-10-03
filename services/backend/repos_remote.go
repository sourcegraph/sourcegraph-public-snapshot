package backend

import (
	"strings"

	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/github"
)

// This file deals with remote repos (e.g., GitHub repos) that are not
// persisted locally.

func (s *repos) Resolve(ctx context.Context, op *sourcegraph.RepoResolveOp) (*sourcegraph.RepoResolution, error) {
	if Mocks.Repos.Resolve != nil {
		return Mocks.Repos.Resolve(ctx, op)
	}

	ctx = context.WithValue(ctx, github.GitHubTrackingContextKey, "Repos.Resolve")
	// First, look up locally.
	if repo, err := localstore.Repos.GetByURI(ctx, op.Path); err == nil {
		return &sourcegraph.RepoResolution{Repo: repo.ID, CanonicalPath: op.Path}, nil
	} else if errcode.GRPC(err) != codes.NotFound {
		return nil, err
	}

	// Next, see if it's a GitHub repo.
	if repo, err := github.ReposFromContext(ctx).Get(ctx, op.Path); err == nil {
		// If canonical location differs, try looking up locally at canonical location.
		if canonicalPath := "github.com/" + repo.Owner + "/" + repo.Name; op.Path != canonicalPath {
			if repo, err := localstore.Repos.GetByURI(ctx, canonicalPath); err == nil {
				return &sourcegraph.RepoResolution{Repo: repo.ID, CanonicalPath: canonicalPath}, nil
			}
		}

		if op.Remote {
			return &sourcegraph.RepoResolution{RemoteRepo: repo}, nil
		}
		return nil, grpc.Errorf(codes.NotFound, "resolved repo not found locally: %s", op.Path)
	} else if errcode.GRPC(err) != codes.NotFound {
		return nil, err
	}

	// Try some remote aliases.
	if op.Remote {
		switch {
		case strings.HasPrefix(op.Path, "gopkg.in/"):
			return &sourcegraph.RepoResolution{
				RemoteRepo: &sourcegraph.Repo{HTTPCloneURL: "https://" + op.Path},
			}, nil
		}
	}

	// Not found anywhere.
	return nil, grpc.Errorf(codes.NotFound, "repo %q not found locally or remotely", op.Path)
}
