package shared

import (
	"context"

	"github.com/sourcegraph/go-diff/diff"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/gitserver"
)

type GitserverClient interface {
	CommitsExist(ctx context.Context, commits []gitserver.RepositoryCommit) ([]bool, error)
	DiffPath(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, sourceCommit, targetCommit, path string) ([]*diff.Hunk, error)
}
