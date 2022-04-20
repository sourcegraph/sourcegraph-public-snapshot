package repos

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
)

type DependenciesService interface {
	ListDependencyRepos(ctx context.Context, opts dependencies.ListDependencyReposOpts) ([]dependencies.Repo, error)
}
