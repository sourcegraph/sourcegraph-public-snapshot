package shared

import (
	"context"
	"runtime"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const SIMILARITY_SEARCH_MIN_ROWS_TO_SPLIT = 1000

type getRepoEmbeddingIndexFn func(ctx context.Context, repoName api.RepoName) (*embeddings.RepoEmbeddingIndex, error)
type getQueryEmbeddingFn func(ctx context.Context, query string) ([]float32, error)

func searchRepoEmbeddingIndexes(
	ctx context.Context,
	params embeddings.EmbeddingsSearchParameters,
	getRepoEmbeddingIndex getRepoEmbeddingIndexFn,
	getQueryEmbedding getQueryEmbeddingFn,
	weaviate *weaviateClient,
) (*embeddings.EmbeddingCombinedSearchResults, error) {
	floatQuery, err := getQueryEmbedding(ctx, params.Query)
	if err != nil {
		return nil, errors.Wrap(err, "getting query embedding")
	}
	embeddedQuery := embeddings.Quantize(floatQuery)

	workerOpts := embeddings.WorkerOptions{
		NumWorkers:     runtime.GOMAXPROCS(0),
		MinRowsToSplit: SIMILARITY_SEARCH_MIN_ROWS_TO_SPLIT,
	}

	searchOpts := embeddings.SearchOptions{
		UseDocumentRanks: params.UseDocumentRanks,
	}

	var result embeddings.EmbeddingCombinedSearchResults

	for i, repoName := range params.RepoNames {
		if weaviate.Use(ctx) {
			codeResults, textResults, err := weaviate.Search(ctx, repoName, params.RepoIDs[i], params.Query, params.CodeResultsCount, params.TextResultsCount)
			if err != nil {
				return nil, err
			}

			result.CodeResults.MergeTruncate(codeResults, params.CodeResultsCount)
			result.TextResults.MergeTruncate(textResults, params.TextResultsCount)
			continue
		}

		embeddingIndex, err := getRepoEmbeddingIndex(ctx, repoName)
		if err != nil {
			return nil, errors.Wrapf(err, "getting repo embedding index for repo %q", repoName)
		}

		codeResults := embeddingIndex.CodeIndex.SimilaritySearch(embeddedQuery, params.CodeResultsCount, workerOpts, searchOpts, embeddingIndex.RepoName, embeddingIndex.Revision)
		textResults := embeddingIndex.TextIndex.SimilaritySearch(embeddedQuery, params.TextResultsCount, workerOpts, searchOpts, embeddingIndex.RepoName, embeddingIndex.Revision)

		result.CodeResults.MergeTruncate(codeResults, params.CodeResultsCount)
		result.TextResults.MergeTruncate(textResults, params.TextResultsCount)

	}

	return &result, nil
}
