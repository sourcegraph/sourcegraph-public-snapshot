package resolver

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

type GitService interface {
	GetCommits(ctx context.Context, repoCommits []api.RepoCommit, ignoreErrors bool) ([]*gitdomain.Commit, error)
}
