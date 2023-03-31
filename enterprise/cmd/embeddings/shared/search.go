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
type getQueryEmbeddingFn func(query string) ([]float32, error)

func searchRepoEmbeddingIndex(
	ctx context.Context,
	logger log.Logger,
	params embeddings.EmbeddingsSearchParameters,
	readFile readFileFn,
	getRepoEmbeddingIndex getRepoEmbeddingIndexFn,
	getQueryEmbedding getQueryEmbeddingFn,
	debug bool,
) (*embeddings.EmbeddingSearchResults, error) {
	embeddingIndex, err := getRepoEmbeddingIndex(ctx, params.RepoName)
	if err != nil {
		return nil, errors.Wrap(err, "getting repo embedding index")
	}

	embeddedQuery, err := getQueryEmbedding(params.Query)
	if err != nil {
		return nil, errors.Wrap(err, "getting query embedding")
	}

	var codeResults, textResults []embeddings.EmbeddingSearchResult
	if params.CodeResultsCount > 0 && len(embeddingIndex.CodeIndex.Embeddings) > 0 {
		codeResults = searchEmbeddingIndex(ctx, logger, embeddingIndex.RepoName, embeddingIndex.Revision, &embeddingIndex.CodeIndex, readFile, embeddedQuery, params.CodeResultsCount, debug)
	}

	if params.TextResultsCount > 0 && len(embeddingIndex.TextIndex.Embeddings) > 0 {
		textResults = searchEmbeddingIndex(ctx, logger, embeddingIndex.RepoName, embeddingIndex.Revision, &embeddingIndex.TextIndex, readFile, embeddedQuery, params.TextResultsCount, debug)
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
	debug bool,
) []embeddings.EmbeddingSearchResult {
	numWorkers := runtime.GOMAXPROCS(0)
	rows := index.SimilaritySearch(query, nResults, embeddings.WorkerOptions{NumWorkers: numWorkers, MinRowsToSplit: SIMILARITY_SEARCH_MIN_ROWS_TO_SPLIT}, debug)

	// Hydrate content
	for idx, row := range rows {
		fileContent, err := readFile(ctx, repoName, revision, row.FileName)
		if err != nil {
			if !os.IsNotExist(err) {
				logger.Error("error reading file", log.Error(err))
			}
			continue
		}
		lines := strings.Split(string(fileContent), "\n")

		// Sanity check: check that startLine and endLine are within 0 and len(lines).
		startLine := max(0, min(len(lines), row.StartLine))
		endLine := max(0, min(len(lines), row.EndLine))

		rows[idx].Content = strings.Join(lines[startLine:endLine], "\n")
	}

	return rows
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
