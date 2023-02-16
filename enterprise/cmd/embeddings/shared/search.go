package shared

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/embed"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
)

const QUERY_EMBEDDING_RETRIES = 3

type readFileFn func(ctx context.Context, repoName api.RepoName, revision api.CommitID, fileName string) ([]byte, error)

type getRepoEmbeddingIndexFn func(ctx context.Context, repoEmbeddingIndexName embeddings.RepoEmbeddingIndexName) (*embeddings.RepoEmbeddingIndex, error)

func SearchRepoEmbeddingIndex(
	ctx context.Context,
	params embeddings.EmbeddingsSearchParameters,
	readFile readFileFn,
	getRepoEmbeddingIndex getRepoEmbeddingIndexFn,
) (*embeddings.EmbeddingSearchResults, error) {
	repoEmbeddingIndexName := embeddings.GetRepoEmbeddingIndexName(params.RepoName)
	embeddingIndex, err := getRepoEmbeddingIndex(ctx, repoEmbeddingIndexName)
	if err != nil {
		return nil, err
	}

	config := conf.Get().Embeddings
	embeddedQuery, err := embed.GetEmbeddingsWithRetries([]string{params.Query}, config, QUERY_EMBEDDING_RETRIES)
	if err != nil {
		return nil, err
	}

	var codeResults, textResults []embeddings.EmbeddingSearchResult
	if params.CodeResultsCount > 0 && embeddingIndex.CodeIndex != nil {
		codeResults = searchEmbeddingIndex(ctx, embeddingIndex.RepoName, embeddingIndex.Revision, embeddingIndex.CodeIndex, readFile, embeddedQuery, params.CodeResultsCount)
	}

	if params.TextResultsCount > 0 && embeddingIndex.TextIndex != nil {
		textResults = searchEmbeddingIndex(ctx, embeddingIndex.RepoName, embeddingIndex.Revision, embeddingIndex.TextIndex, readFile, embeddedQuery, params.TextResultsCount)
	}

	return &embeddings.EmbeddingSearchResults{CodeResults: codeResults, TextResults: textResults}, nil
}

func searchEmbeddingIndex(
	ctx context.Context,
	repoName api.RepoName,
	revision api.CommitID,
	index *embeddings.EmbeddingIndex[embeddings.RepoEmbeddingRowMetadata],
	readFile readFileFn,
	query []float32,
	nResults int,
) []embeddings.EmbeddingSearchResult {
	rows := index.SimilaritySearch(query, nResults)
	results := make([]embeddings.EmbeddingSearchResult, len(rows))
	for idx, row := range rows {
		fileContent, err := readFile(ctx, repoName, revision, row.FileName)
		if err != nil {
			continue
		}
		lines := strings.Split(string(fileContent), "\n")

		results[idx] = embeddings.EmbeddingSearchResult{
			FileName:  row.FileName,
			StartLine: row.StartLine,
			EndLine:   row.EndLine,
			// TODO: Sanity check: check that startline and endline are within 0 and len(lines)
			Content: strings.Join(lines[row.StartLine:row.EndLine], "\n"),
		}
	}

	return results
}
