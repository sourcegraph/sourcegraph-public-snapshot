package shared

import (
	"context"
	"runtime"
	"sort"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type readFileFn func(ctx context.Context, repoName api.RepoName, revision api.CommitID, fileName string) ([]byte, error)
type getRepoEmbeddingIndexFn func(ctx context.Context, repoName api.RepoName) (*embeddings.RepoEmbeddingIndex, error)
type getQueryEmbeddingFn func(ctx context.Context, query string) ([]float32, error)

func searchRepoEmbeddingIndexes(
	ctx context.Context,
	logger log.Logger,
	multiParams embeddings.EmbeddingsMultiSearchParameters,
	readFile readFileFn,
	getRepoEmbeddingIndex getRepoEmbeddingIndexFn,
	getQueryEmbedding getQueryEmbeddingFn,
	weaviate *weaviateClient,
) (*embeddings.EmbeddingSearchResults, error) {
	floatQuery, err := getQueryEmbedding(ctx, multiParams.Query)
	if err != nil {
		return nil, errors.Wrap(err, "getting query embedding")
	}
	embeddedQuery := embeddings.Quantize(floatQuery)

	workerOpts := embeddings.WorkerOptions{
		NumWorkers:     runtime.GOMAXPROCS(0),
		MinRowsToSplit: SIMILARITY_SEARCH_MIN_ROWS_TO_SPLIT,
	}

	searchOpts := embeddings.SearchOptions{
		Debug:            multiParams.Debug,
		UseDocumentRanks: multiParams.UseDocumentRanks,
	}

	aggregatedCodeResults := newResultAggregator(multiParams.CodeResultsCount)
	aggregatedTextResults := newResultAggregator(multiParams.TextResultsCount)

	for i := range multiParams.RepoNames {
		// TODO assert len(multiParams.RepoNames) == len(multiParams.RepoIDs)
		params := embeddings.EmbeddingsSearchParameters{
			RepoName:         multiParams.RepoNames[i],
			RepoID:           multiParams.RepoIDs[i],
			Query:            "",
			CodeResultsCount: 0,
			TextResultsCount: 0,
			UseDocumentRanks: false,
		}

		if weaviate.Use(ctx) {
			codeResults, textResults, revision, err := weaviate.Search(ctx, params)
			if err != nil {
				return nil, err
			}

			aggregatedCodeResults.Add(params.RepoName, revision, codeResults...)
			aggregatedTextResults.Add(params.RepoName, revision, textResults...)
			continue
		}

		embeddingIndex, err := getRepoEmbeddingIndex(ctx, params.RepoName)
		if err != nil {
			return nil, errors.Wrapf(err, "getting repo embedding index for repo %q", params.RepoName)
		}

		codeResults := embeddingIndex.CodeIndex.SimilaritySearch(embeddedQuery, params.CodeResultsCount, workerOpts, searchOpts, embeddingIndex.RepoName, embeddingIndex.Revision)
		aggregatedCodeResults.Add(embeddingIndex.RepoName, embeddingIndex.Revision, codeResults...)

		textResults := embeddingIndex.TextIndex.SimilaritySearch(embeddedQuery, params.TextResultsCount, workerOpts, searchOpts, embeddingIndex.RepoName, embeddingIndex.Revision)
		aggregatedTextResults.Add(embeddingIndex.RepoName, embeddingIndex.Revision, textResults...)
	}

	return &embeddings.EmbeddingSearchResults{
		CodeResults: aggregatedCodeResults.results,
		TextResults: aggregatedTextResults.results,
	}, nil
}

func newResultAggregator(maxResults int) resultAggregator {
	return resultAggregator{
		results:    make([]embeddings.EmbeddingSearchResult, maxResults*2),
		maxResults: maxResults,
	}
}

type resultAggregator struct {
	results    []embeddings.EmbeddingSearchResult
	maxResults int
}

func (a *resultAggregator) Add(repoName api.RepoName, revision api.CommitID, srs ...embeddings.EmbeddingSearchResult) {
	// Note: this leaves some performance on the table because we know the new
	// results are sorted. We could instead merge here. Given that the number of
	// results we are usually fetching is quite small (<100), I don't expect
	// re-sorting for ever repo to be a problem.
	a.results = append(a.results, srs...)
	sort.Slice(a.results, func(i, j int) bool { return a.results[i].Score() > a.results[j].Score() })
	a.results = a.results[:min(a.maxResults, len(a.results))]
}

const SIMILARITY_SEARCH_MIN_ROWS_TO_SPLIT = 1000

func searchEmbeddingIndex(
	ctx context.Context,
	logger log.Logger,
	repoName api.RepoName,
	revision api.CommitID,
	index *embeddings.EmbeddingIndex,
	readFile readFileFn,
	query []int8,
	nResults int,
	opts embeddings.SearchOptions,
) []embeddings.EmbeddingSearchResult {
	numWorkers := runtime.GOMAXPROCS(0)
	results := index.SimilaritySearch(query, nResults, embeddings.WorkerOptions{NumWorkers: numWorkers, MinRowsToSplit: SIMILARITY_SEARCH_MIN_ROWS_TO_SPLIT}, opts)

	return filterAndHydrateContent(
		ctx,
		logger,
		repoName,
		revision,
		readFile,
		opts.Debug,
		results,
	)
}

// filterAndHydrateContent will mutate unfiltered to populate the Content
// field. If we fail to read a file (eg permission issues) we will remove the
// result. As such the returned slice should be used.
func filterAndHydrateContent(
	ctx context.Context,
	logger log.Logger,
	repoName api.RepoName,
	revision api.CommitID,
	readFile readFileFn,
	debug bool,
	unfiltered []embeddings.SimilaritySearchResult,
) []embeddings.EmbeddingSearchResult {
	filtered := make([]embeddings.EmbeddingSearchResult, 0, len(unfiltered))

	for _, result := range unfiltered {
		fileContent, err := readFile(ctx, repoName, revision, result.FileName)
		if err != nil {
			if !os.IsNotExist(err) {
				logger.Error("error reading file", log.String("repoName", string(repoName)), log.String("revision", string(revision)), log.String("fileName", result.FileName), log.Error(err))
			}
			continue
		}
		lines := strings.Split(string(fileContent), "\n")

		// Sanity check: check that startLine and endLine are within 0 and len(lines).
		startLine := max(0, min(len(lines), result.StartLine))
		endLine := max(0, min(len(lines), result.EndLine))

		content := strings.Join(lines[startLine:endLine], "\n")

		var debugString string
		if debug {
			debugString = fmt.Sprintf("score:%d, similarity:%d, rank:%d", result.Score(), result.SimilarityScore, result.RankScore)
		}

		filtered = append(filtered, embeddings.EmbeddingSearchResult{
			RepoName: repoName,
			Revision: revision,
			RepoEmbeddingRowMetadata: embeddings.RepoEmbeddingRowMetadata{
				FileName:  result.FileName,
				StartLine: startLine,
				EndLine:   endLine,
			},
			Debug:   debugString,
			Content: content,
		})
	}

	return filtered
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
