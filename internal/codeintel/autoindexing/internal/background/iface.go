package background

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

type DependenciesService interface {
	UpsertDependencyRepos(ctx context.Context, deps []dependencies.Repo) ([]dependencies.Repo, error)
}

type GitserverRepoStore interface {
	GetByNames(ctx context.Context, names ...api.RepoName) (map[api.RepoName]*types.GitserverRepo, error)
}

type AutoIndexingServiceForDepSchedulingShim struct {
	*Service
}

func (s *AutoIndexingServiceForDepSchedulingShim) QueueIndexesForPackage(ctx context.Context, pkg precise.Package) error {
	return s.Service.queueIndexesForPackage(ctx, pkg)
}

func (s *AutoIndexingServiceForDepSchedulingShim) InsertDependencyIndexingJob(ctx context.Context, uploadID int, externalServiceKind string, syncTime time.Time) (id int, err error) {
	return s.Service.insertDependencyIndexingJob(ctx, uploadID, externalServiceKind, syncTime)
}
