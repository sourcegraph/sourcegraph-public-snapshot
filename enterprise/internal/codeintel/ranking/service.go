package ranking

import (
	"context"
	"sort"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/schema"
)

type Service struct {
	store           store.Store
	uploadSvc       *uploads.Service
	gitserverClient GitserverClient
	symbolsClient   SymbolsClient
	getConf         conftypes.SiteConfigQuerier
	resultsBucket   *storage.BucketHandle
	operations      *operations
	logger          log.Logger
}

func newService(
	store store.Store,
	uploadSvc *uploads.Service,
	gitserverClient GitserverClient,
	symbolsClient SymbolsClient,
	getConf conftypes.SiteConfigQuerier,
	resultsBucket *storage.BucketHandle,
	observationContext *observation.Context,
) *Service {
	return &Service{
		store:           store,
		uploadSvc:       uploadSvc,
		gitserverClient: gitserverClient,
		symbolsClient:   symbolsClient,
		getConf:         getConf,
		resultsBucket:   resultsBucket,
		operations:      newOperations(observationContext),
		logger:          observationContext.Logger,
	}
}

// GetRepoRank returns a rank vector for the given repository. Repositories are assumed to
// be ordered by each pairwise component of the resulting vector, higher ranks coming earlier.
// We currently rank first by user-defined scores, then by GitHub star count.
func (s *Service) GetRepoRank(ctx context.Context, repoName api.RepoName) (_ []float64, err error) {
	_, _, endObservation := s.operations.getRepoRank.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	userRank := repoRankFromConfig(s.getConf.SiteConfig(), string(repoName))

	starRank, err := s.store.GetStarRank(ctx, repoName)
	if err != nil {
		return nil, err
	}

	return []float64{squashRange(userRank), starRank}, nil
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

// GetDocumentRank returns a map from paths within the given repo to their rank vector. Paths are
// assumed to be ordered by each pairwise component of the resulting vector, higher ranks coming
// earlier. We currently rank documents by path name length and lexicographic order, while performing
// a few heuristics to sink generated, test, and vendor files lower in the ranking.
//
// Rank vector index labels:
//   - precision                   [0 to 1]
//   - generated                   [0 or 1]
//   - vendor                      [0 or 1]
//   - test                        [0 or 1]
//   - global document rank        [0 to 1] (=0 w/o pagerank)
//   - name length                 [0 to 1] (=1 w/  pagerank)
//   - lexicographic order in repo [0 to 1] (=1 w/  pagerank)
func (s *Service) GetDocumentRanks(ctx context.Context, repoName api.RepoName) (_ map[string][]float64, err error) {
	_, _, endObservation := s.operations.getDocumentRanks.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	ranks := map[string][]float64{}
	documentRanks, ok, err := s.store.GetDocumentRanks(ctx, repoName)
	if err != nil {
		return nil, err
	}
	if ok {
		for path, rank := range documentRanks {
			ranks[path] = []float64{
				rank[0],                             // precision level (0, 1]
				1 - boolRank(isPathGenerated(path)), // rank generated paths lower
				1 - boolRank(isPathVendored(path)),  // rank vendored paths lower
				1 - boolRank(isPathTest(path)),      // rank test paths lower
				squashRange(rank[1]),                // global document rank
				1,                                   // name length
				1,                                   // lexicographic order in repo
			}
		}
	}

	paths, err := s.gitserverClient.ListFilesForRepo(ctx, repoName, "HEAD", allPathsPattern.Re())
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)

	for i, path := range paths {
		if _, ok := ranks[path]; ok {
			continue
		}

		ranks[path] = []float64{
			0,                                     // imprecise
			1 - boolRank(isPathGenerated(path)),   // rank generated paths lower
			1 - boolRank(isPathVendored(path)),    // rank vendored paths lower
			1 - boolRank(isPathTest(path)),        // rank test paths lower
			0,                                     // no global document rank
			1.0 - squashRange(float64(len(path))), // name length (prefer short names)
			1.0 - float64(i)/float64(len(paths)),  // lexicographic order in repo
		}
	}

	return ranks, nil
}

func (s *Service) LastUpdatedAt(ctx context.Context, repoIDs []api.RepoID) (map[api.RepoID]time.Time, error) {
	return s.store.LastUpdatedAt(ctx, repoIDs)
}

func (s *Service) UpdatedAfter(ctx context.Context, t time.Time) ([]api.RepoName, error) {
	return s.store.UpdatedAfter(ctx, t)
}

func isPathGenerated(path string) bool {
	return strings.HasSuffix(path, "min.js") || strings.HasSuffix(path, "js.map")
}

func isPathVendored(path string) bool {
	return strings.Contains(path, "vendor/") || strings.Contains(path, "node_modules/")
}

var testPattern = lazyregexp.New("test")

func isPathTest(path string) bool {
	return testPattern.MatchString(path)
}

// Converts a boolean to a [0, 1] rank (where true is ordered before false).
func boolRank(v bool) float64 {
	if v {
		return 1.0
	}

	return 0.0
}

// squashRange maps a value in the range [0, inf) to a value in the range
// [0, 1) monotonically (i.e., (a < b) <-> (squashRange(a) < squashRange(b))).
func squashRange(j float64) float64 {
	return j / (1 + j)
}
