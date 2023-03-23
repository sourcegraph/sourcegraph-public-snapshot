package shared

import (
	"context"
	"runtime"
	"strings"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

type readFileFn func(ctx context.Context, repoName api.RepoName, revision api.CommitID, fileName string) ([]byte, error)
type getRepoEmbeddingIndexFn func(ctx context.Context, repoName api.RepoName) (*embeddings.RepoEmbeddingIndex, error)
type getQueryEmbeddingFn func(query string) ([]float32, error)

func searchRepoEmbeddingIndex(
	ctx context.Context,
	params embeddings.EmbeddingsSearchParameters,
	readFile readFileFn,
	getRepoEmbeddingIndex getRepoEmbeddingIndexFn,
	getQueryEmbedding getQueryEmbeddingFn,
	debug bool,
) (*embeddings.EmbeddingSearchResults, error) {
	embeddingIndex, err := getRepoEmbeddingIndex(ctx, params.RepoName)
	if err != nil {
		return nil, err
	}

	embeddedQuery, err := getQueryEmbedding(params.Query)
	if err != nil {
		return nil, err
	}

	var codeResults, textResults []embeddings.EmbeddingSearchResult
	if params.CodeResultsCount > 0 && len(embeddingIndex.CodeIndex.Embeddings) > 0 {
		codeResults = searchEmbeddingIndex(ctx, embeddingIndex.RepoName, embeddingIndex.Revision, &embeddingIndex.CodeIndex, readFile, embeddedQuery, params.CodeResultsCount, debug)
	}

	if params.TextResultsCount > 0 && len(embeddingIndex.TextIndex.Embeddings) > 0 {
		textResults = searchEmbeddingIndex(ctx, embeddingIndex.RepoName, embeddingIndex.Revision, &embeddingIndex.TextIndex, readFile, embeddedQuery, params.TextResultsCount, debug)
	}

	return &embeddings.EmbeddingSearchResults{CodeResults: codeResults, TextResults: textResults}, nil
}

const SIMILARITY_SEARCH_MIN_ROWS_TO_SPLIT = 1000

func searchEmbeddingIndex(
	ctx context.Context,
	repoName api.RepoName,
	revision api.CommitID,
	index *embeddings.EmbeddingIndex[embeddings.RepoEmbeddingRowMetadata],
	readFile readFileFn,
	query []float32,
	nResults int,
	debug bool,
) []embeddings.EmbeddingSearchResult {
	numWorkers := runtime.GOMAXPROCS(0)
	res := index.SimilaritySearch(query, nResults, embeddings.WorkerOptions{NumWorkers: numWorkers, MinRowsToSplit: SIMILARITY_SEARCH_MIN_ROWS_TO_SPLIT}, debug)

	results := make([]embeddings.EmbeddingSearchResult, len(res.RowMetadata))
	for idx, row := range res.RowMetadata {
		fileContent, err := readFile(ctx, repoName, revision, row.FileName)
		if err != nil {
			continue
		}
		lines := strings.Split(string(fileContent), "\n")

		// Sanity check: check that startLine and endLine are within 0 and len(lines).
		startLine := max(0, min(len(lines), row.StartLine))
		endLine := max(0, min(len(lines), row.EndLine))

		results[idx] = embeddings.EmbeddingSearchResult{
			FileName:  row.FileName,
			StartLine: row.StartLine,
			EndLine:   row.EndLine,
			Content:   strings.Join(lines[startLine:endLine], "\n"),
		}
		if debug && len(res.Debug) > idx {
			results[idx].Debug = res.Debug[idx]
		}
	}

	return results
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
