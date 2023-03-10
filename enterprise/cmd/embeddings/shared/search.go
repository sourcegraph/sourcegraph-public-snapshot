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
		codeResults = searchEmbeddingIndex(ctx, embeddingIndex.RepoName, embeddingIndex.Revision, &embeddingIndex.CodeIndex, readFile, embeddedQuery, params.CodeResultsCount)
	}

	if params.TextResultsCount > 0 && len(embeddingIndex.TextIndex.Embeddings) > 0 {
		textResults = searchEmbeddingIndex(ctx, embeddingIndex.RepoName, embeddingIndex.Revision, &embeddingIndex.TextIndex, readFile, embeddedQuery, params.TextResultsCount)
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
) []embeddings.EmbeddingSearchResult {
	numWorkers := runtime.GOMAXPROCS(0)
	rows := index.SimilaritySearch(query, nResults, embeddings.WorkerOptions{NumWorkers: numWorkers, MinRowsToSplit: SIMILARITY_SEARCH_MIN_ROWS_TO_SPLIT})

	results := make([]embeddings.EmbeddingSearchResult, len(rows))
	for idx, row := range rows {
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
