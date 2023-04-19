package shared

import (
	"context"
	"os"
	"runtime"
	"strings"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type readFileFn func(ctx context.Context, repoName api.RepoName, revision api.CommitID, fileName string) ([]byte, error)
type getRepoEmbeddingIndexFn func(ctx context.Context, repoName api.RepoName) (*embeddings.RepoEmbeddingIndex, error)
type getQueryEmbeddingFn func(ctx context.Context, query string) ([]float32, error)

func searchRepoEmbeddingIndex(
	ctx context.Context,
	logger log.Logger,
	params embeddings.EmbeddingsSearchParameters,
	readFile readFileFn,
	getRepoEmbeddingIndex getRepoEmbeddingIndexFn,
	getQueryEmbedding getQueryEmbeddingFn,
) (*embeddings.EmbeddingSearchResults, error) {
	embeddingIndex, err := getRepoEmbeddingIndex(ctx, params.RepoName)
	if err != nil {
		return nil, errors.Wrap(err, "getting repo embedding index")
	}

	embeddedQuery, err := getQueryEmbedding(ctx, params.Query)
	if err != nil {
		return nil, errors.Wrap(err, "getting query embedding")
	}

	opts := embeddings.SearchOptions{
		Debug:            params.Debug,
		UseDocumentRanks: params.UseDocumentRanks,
	}

	var codeResults, textResults []embeddings.EmbeddingSearchResult
	if params.CodeResultsCount > 0 && len(embeddingIndex.CodeIndex.Embeddings) > 0 {
		codeResults = searchEmbeddingIndex(ctx, logger, embeddingIndex.RepoName, embeddingIndex.Revision, &embeddingIndex.CodeIndex, readFile, embeddedQuery, params.CodeResultsCount, opts)
	}

	if params.TextResultsCount > 0 && len(embeddingIndex.TextIndex.Embeddings) > 0 {
		textResults = searchEmbeddingIndex(ctx, logger, embeddingIndex.RepoName, embeddingIndex.Revision, &embeddingIndex.TextIndex, readFile, embeddedQuery, params.TextResultsCount, opts)
	}

	return &embeddings.EmbeddingSearchResults{CodeResults: codeResults, TextResults: textResults}, nil
}

const SIMILARITY_SEARCH_MIN_ROWS_TO_SPLIT = 1000

func searchEmbeddingIndex(
	ctx context.Context,
	logger log.Logger,
	repoName api.RepoName,
	revision api.CommitID,
	index *embeddings.EmbeddingIndex,
	readFile readFileFn,
	query []float32,
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
	unfiltered []embeddings.EmbeddingSearchResult,
) []embeddings.EmbeddingSearchResult {
	filtered := unfiltered[:0]

	for idx, result := range unfiltered {
		fileContent, err := readFile(ctx, repoName, revision, result.FileName)
		if err != nil {
			if !os.IsNotExist(err) {
				logger.Error("error reading file", log.String("repoName", string(repoName)), log.String("revision", string(revision)), log.String("fileName", result.FileName), log.Error(err))
			}
			// scrub row just in case we leak it out
			unfiltered[idx] = embeddings.EmbeddingSearchResult{}
			continue
		}
		lines := strings.Split(string(fileContent), "\n")

		// Sanity check: check that startLine and endLine are within 0 and len(lines).
		result.StartLine = max(0, min(len(lines), result.StartLine))
		result.EndLine = max(0, min(len(lines), result.EndLine))

		result.Content = strings.Join(lines[result.StartLine:result.EndLine], "\n")

		filtered = append(filtered, result)
	}

	return filtered
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
