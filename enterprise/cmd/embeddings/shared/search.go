package shared

import (
	"context"
	"runtime"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const SIMILARITY_SEARCH_MIN_ROWS_TO_SPLIT = 1000

type readFileFn func(ctx context.Context, repoName api.RepoName, revision api.CommitID, fileName string) ([]byte, error)
type getRepoEmbeddingIndexFn func(ctx context.Context, repoName api.RepoName) (*embeddings.RepoEmbeddingIndex, error)
type getQueryEmbeddingFn func(ctx context.Context, query string) ([]float32, error)

func searchRepoEmbeddingIndexes(
	ctx context.Context,
	logger log.Logger,
	params embeddings.EmbeddingsMultiSearchParameters,
	readFile readFileFn,
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
			singleParams := embeddings.EmbeddingsSearchParameters{
				RepoName: repoName,
				// TODO assert len(multiParams.RepoNames) == len(multiParams.RepoIDs)
				RepoID:           params.RepoIDs[i],
				Query:            "",
				CodeResultsCount: 0,
				TextResultsCount: 0,
				UseDocumentRanks: false,
			}

			codeResults, textResults, err := weaviate.Search(ctx, singleParams)
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
