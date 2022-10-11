package ranking

import (
	"context"
	"sort"
	"strings"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/schema"
)

type Service struct {
	store           store.Store
	uploadSvc       *uploads.Service
	gitserverClient GitserverClient
	getConf         conftypes.SiteConfigQuerier
	operations      *operations
	logger          log.Logger
}

func newService(
	store store.Store,
	uploadSvc *uploads.Service,
	gitserverClient GitserverClient,
	getConf conftypes.SiteConfigQuerier,
	observationContext *observation.Context,
) *Service {
	return &Service{
		store:           store,
		uploadSvc:       uploadSvc,
		gitserverClient: gitserverClient,
		getConf:         getConf,
		operations:      newOperations(observationContext),
		logger:          observationContext.Logger,
	}
}

// GetRepoRank returns a score vector for the given repository. Repositories are assumed to
// be ordered by each pairwise component of the resulting vector, lower scores coming earlier.
// We currently rank first by user-defined scores, then by GitHub star count.
func (s *Service) GetRepoRank(ctx context.Context, repoName api.RepoName) (_ []float64, err error) {
	_, _, endObservation := s.operations.getRepoRank.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	userRank := repoRankFromConfig(s.getConf.SiteConfig(), string(repoName))

	starRank, err := s.store.GetStarRank(ctx, repoName)
	if err != nil {
		return nil, err
	}

	return []float64{1 - squashRange(userRank), 1 - starRank}, nil
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
// assumed to be ordered by each pairwise component of the resulting vector, lower scores coming
// earlier. We currently rank documents by path name length and lexicographic order, while performing
// a few heuristics to sink generated, test, and vendor files lower in the ranking.
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
		ranks[path] = rank(path, float64(i)/n)
	}

	return ranks, nil
}

var testPattern = lazyregexp.New("test")

// copy pasta + modified
// https://github.com/sourcegraph/zoekt/blob/f89a534103a224663d23b4579959854dd7816942/build/builder.go#L872-L918
func rank(name string, nameRank float64) []float64 {
	generated := 0.0
	if strings.HasSuffix(name, "min.js") || strings.HasSuffix(name, "js.map") {
		generated = 1.0
	}

	vendor := 0.0
	if strings.Contains(name, "vendor/") || strings.Contains(name, "node_modules/") {
		vendor = 1.0
	}

	test := 0.0
	if testPattern.MatchString(name) {
		test = 1.0
	}

	// Smaller is earlier (=better).
	return []float64{
		// Prefer docs that are not generated
		generated,

		// Prefer docs that are not vendored
		vendor,

		// Prefer docs that are not tests
		test,

		// With short names
		squashRange(float64(len(name))),

		// // With many symbols
		// 1.0 - squashRange(len(d.Symbols)),

		// // With short content
		// squashRange(len(d.Content)),

		// // That is present is as many branches as possible
		// 1.0 - squashRange(len(d.Branches)),

		// Preserve original ordering.
		nameRank,
	}
}

// map [0,inf) to [0,1) monotonically
func squashRange(j float64) float64 {
	return j / (1 + j)
}
