package ranking

import (
	"context"
	"sort"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/schema"
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

// GetRepoRank returns a score vector for the given repository. Repositories are assumed to
// be ordered by each pairwise component of the resulting vector, higher scores coming earlier.
// We currently rank first by user-defined scores, then by GitHub star count.
func (s *Service) GetRepoRank(ctx context.Context, repoName api.RepoName) (_ []float64, err error) {
	_, _, endObservation := s.operations.getRepoRank.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	siteConfig := conf.Get().SiteConfiguration
	userRank := repoRankFromConfig(siteConfig, string(repoName))

	starRank, err := s.store.GetStarRank(ctx, repoName)
	if err != nil {
		return nil, err
	}

	return []float64{userRank, starRank}, nil
}

// copy pasta
// https://github.com/sourcegraph/sourcegraph/blob/942c417363b07c9e0a6377456f1d6a80a94efb99/cmd/frontend/internal/httpapi/search.go#L172
func repoRankFromConfig(siteConfig schema.SiteConfiguration, repoName string) float64 {
	val := 0.0
	if siteConfig.ExperimentalFeatures == nil || siteConfig.ExperimentalFeatures.Ranking == nil {
		return val
	}
	scores := siteConfig.ExperimentalFeatures.Ranking.RepoScores
	if len(scores) == 0 {
		return val
	}
	// try every "directory" in the repo name to assign it a value, so a repoName like
	// "github.com/sourcegraph/zoekt" will have "github.com", "github.com/sourcegraph",
	// and "github.com/sourcegraph/zoekt" tested.
	for i := 0; i < len(repoName); i++ {
		if repoName[i] == '/' {
			val += scores[repoName[:i]]
		}
	}
	val += scores[repoName]
	return val
}

var allPathsPattern = lazyregexp.New(".*")

// GetDocumentRank returns a map from paths within the given repo to their score vectors. Paths are
// assumed to be ordered by each pairwise component of the resulting vector, higher scores coming
// earlier. We currently rank lexicographically.
func (s *Service) GetDocumentRanks(ctx context.Context, repoName api.RepoName) (_ map[string][]float64, err error) {
	_, _, endObservation := s.operations.getDocumentRanks.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	paths, err := s.gitserverClient.ListFilesForRepo(ctx, repoName, "HEAD", allPathsPattern.Re())
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)

	n := float64(len(paths))
	ranks := make(map[string][]float64, len(paths))
	for i, path := range paths {
		ranks[path] = []float64{(float64(i) / n)}
	}

	return ranks, nil
}
