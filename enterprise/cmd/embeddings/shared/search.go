package shared

import (
	"context"
	"fmt"
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
	weaviate *weaviateClient,
) (*embeddings.EmbeddingSearchResults, error) {
	if weaviate.Use(ctx) {
		return weaviate.Search(ctx, params)
	}

	embeddingIndex, err := getRepoEmbeddingIndex(ctx, params.RepoName)
	if err != nil {
		return nil, errors.Wrapf(err, "getting repo embedding index for repo %q", params.RepoName)
	}

	floatQuery, err := getQueryEmbedding(ctx, params.Query)
	if err != nil {
		return nil, errors.Wrap(err, "getting query embedding")
	}
	embeddedQuery := embeddings.Quantize(floatQuery)

	opts := embeddings.SearchOptions{
		Debug:            params.Debug,
		UseDocumentRanks: params.UseDocumentRanks,
	}

	codeResults := searchEmbeddingIndex(ctx, logger, embeddingIndex.RepoName, embeddingIndex.Revision, &embeddingIndex.CodeIndex, readFile, embeddedQuery, params.CodeResultsCount, opts)
	textResults := searchEmbeddingIndex(ctx, logger, embeddingIndex.RepoName, embeddingIndex.Revision, &embeddingIndex.TextIndex, readFile, embeddedQuery, params.TextResultsCount, opts)

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

		content := strings.Join(lines[result.StartLine:result.EndLine], "\n")

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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
