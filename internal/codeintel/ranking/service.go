package ranking

import (
	"context"
	"sort"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Service struct {
	store           store.Store
	uploadSvc       *uploads.Service
	gitserverClient GitserverClient
	operations      *operations
	logger          log.Logger
}

func newService(
	store store.Store,
	uploadSvc *uploads.Service,
	gitserverClient GitserverClient,
	observationContext *observation.Context,
) *Service {
	return &Service{
		store:           store,
		uploadSvc:       uploadSvc,
		gitserverClient: gitserverClient,
		operations:      newOperations(observationContext),
		logger:          observationContext.Logger,
	}
}

func (s *Service) GetRepoRank(ctx context.Context, repoName api.RepoName) (_ float64, err error) {
	_, _, endObservation := s.operations.getRepoRank.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.GetStarRank(ctx, repoName)
}

var allPathsPattern = lazyregexp.New(".*")

func (s *Service) GetDocumentRanks(ctx context.Context, repoName api.RepoName) (_ map[string]float64, err error) {
	_, _, endObservation := s.operations.getDocumentRanks.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	paths, err := s.gitserverClient.ListFilesForRepo(ctx, repoName, "HEAD", allPathsPattern.Re())
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)

	n := float64(len(paths))
	ranks := make(map[string]float64, len(paths))
	for i, path := range paths {
		ranks[path] = 1 - (float64(i) / n)
	}

	return ranks, nil
}
