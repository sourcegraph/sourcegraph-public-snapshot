package shared

import (
	"context"
	"runtime"

	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	similaritySearchMinRowsToSplit = 1000
	queryEmbeddingRetries          = 3
)

type (
	getRepoEmbeddingIndexFn func(ctx context.Context, repoID api.RepoID, repoName api.RepoName) (*embeddings.RepoEmbeddingIndex, error)
	getQueryEmbeddingFn     func(ctx context.Context, model string) ([]float32, string, error)
)

func searchRepoEmbeddingIndexes(
	ctx context.Context,
	params embeddings.EmbeddingsSearchParameters,
	getRepoEmbeddingIndex getRepoEmbeddingIndexFn,
	getQueryEmbedding getQueryEmbeddingFn,
) (_ *embeddings.EmbeddingCombinedSearchResults, err error) {
	tr, ctx := trace.New(ctx, "searchRepoEmbeddingIndexes", params.Attrs()...)
	defer tr.EndWithErr(&err)

	floatQuery, queryModel, err := getQueryEmbedding(ctx, params.Query)
	if err != nil {
		return nil, err
	}
	embeddedQuery := embeddings.Quantize(floatQuery, nil)

	workerOpts := embeddings.WorkerOptions{
		NumWorkers:     runtime.GOMAXPROCS(0),
		MinRowsToSplit: similaritySearchMinRowsToSplit,
	}

	searchOpts := embeddings.SearchOptions{
		UseDocumentRanks: params.UseDocumentRanks,
	}

	searchRepo := func(repoID api.RepoID, repoName api.RepoName) (codeResults, textResults []embeddings.EmbeddingSearchResult, err error) {
		tr, ctx := trace.New(ctx, "searchRepo",
			attribute.String("repoName", string(repoName)),
		)
		defer tr.EndWithErr(&err)

		embeddingIndex, err := getRepoEmbeddingIndex(ctx, repoID, repoName)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "getting repo embedding index for repo %q", repoName)
		}

		if !embeddingIndex.IsModelCompatible(queryModel) {
			return nil, nil, errors.Newf("embeddings model in config (%s) does not match the embeddings model for the"+
				" index (%s). Embedding index for repo %q must be reindexed with the new model",
				queryModel, embeddingIndex.EmbeddingsModel, repoName)
		}

		codeResults = embeddingIndex.CodeIndex.SimilaritySearch(embeddedQuery, params.CodeResultsCount, workerOpts, searchOpts, embeddingIndex.RepoName, embeddingIndex.Revision)
		textResults = embeddingIndex.TextIndex.SimilaritySearch(embeddedQuery, params.TextResultsCount, workerOpts, searchOpts, embeddingIndex.RepoName, embeddingIndex.Revision)
		return codeResults, textResults, nil
	}

	var result embeddings.EmbeddingCombinedSearchResults
	for i, repoName := range params.RepoNames {
		codeResults, textResults, err := searchRepo(params.RepoIDs[i], repoName)
		if err != nil {
			return nil, err
		}
		result.CodeResults.MergeTruncate(codeResults, params.CodeResultsCount)
		result.TextResults.MergeTruncate(textResults, params.TextResultsCount)
	}

	return &result, nil
}
