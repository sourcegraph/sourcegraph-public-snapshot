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
type getQueryEmbeddingFn func(query string) ([]float32, error)
type justSearchFn func(ctx context.Context, repoName api.RepoName, query []float32, nCode int32, nTxt int32) (string, []embeddings.RepoEmbeddingRowMetadata, []embeddings.RepoEmbeddingRowMetadata, error)

func searchRepoEmbeddingIndex(
	ctx context.Context,
	logger log.Logger,
	params embeddings.EmbeddingsSearchParameters,
	readFile readFileFn,
	getRepoEmbeddingIndex getRepoEmbeddingIndexFn,
	getQueryEmbedding getQueryEmbeddingFn,
	alternativeSearch justSearchFn,
) (*embeddings.EmbeddingSearchResults, error) {
	embeddingIndex, err := getRepoEmbeddingIndex(ctx, params.RepoName)
	if err != nil {
		return nil, errors.Wrap(err, "getting repo embedding index")
	}

	embeddedQuery, err := getQueryEmbedding(params.Query)
	if err != nil {
		return nil, errors.Wrap(err, "getting query embedding")
	}

	opts := embeddings.SearchOptions{
		Debug:            params.Debug,
		UseDocumentRanks: params.UseDocumentRanks,
	}

	rev, cod, txt, err := alternativeSearch(ctx, embeddingIndex.RepoName, embeddedQuery, int32(params.CodeResultsCount), int32(params.TextResultsCount))
	if err != nil {
		fmt.Printf("search failed %s: %s\n", embeddingIndex.RepoName, err)
	} else {
		fmt.Printf("Embeddings search with DB successful! found %d code and %d text\n", len(cod), len(txt))
	}
	var rCod, rTxt []embeddings.EmbeddingSearchResult
	for _, c := range cod {
		rCod = append(rCod, embeddings.EmbeddingSearchResult{RepoEmbeddingRowMetadata: c})
	}
	for _, t := range txt {
		rTxt = append(rTxt, embeddings.EmbeddingSearchResult{RepoEmbeddingRowMetadata: t})
	}
	fetchContent(ctx, logger, rCod, readFile, embeddingIndex.RepoName, api.CommitID(rev))
	fetchContent(ctx, logger, rTxt, readFile, embeddingIndex.RepoName, api.CommitID(rev))

	var codeResults, textResults []embeddings.EmbeddingSearchResult
	if params.CodeResultsCount > 0 && len(embeddingIndex.CodeIndex.Embeddings) > 0 {
		codeResults = searchEmbeddingIndex(ctx, logger, embeddingIndex.RepoName, embeddingIndex.Revision, &embeddingIndex.CodeIndex, readFile, embeddedQuery, params.CodeResultsCount, opts)
	}

	if params.TextResultsCount > 0 && len(embeddingIndex.TextIndex.Embeddings) > 0 {
		textResults = searchEmbeddingIndex(ctx, logger, embeddingIndex.RepoName, embeddingIndex.Revision, &embeddingIndex.TextIndex, readFile, embeddedQuery, params.TextResultsCount, opts)
	}
	// So that we compile while the return is commented out.
	codeResults = append(codeResults)
	textResults = append(textResults)

	//return &embeddings.EmbeddingSearchResults{CodeResults: codeResults, TextResults: textResults}, nil
	return &embeddings.EmbeddingSearchResults{CodeResults: rCod, TextResults: rTxt}, nil
}

func fetchContent(ctx context.Context,
	logger log.Logger, rows []embeddings.EmbeddingSearchResult, readFile readFileFn, repoName api.RepoName, revision api.CommitID) {
	for idx, row := range rows {
		fileContent, err := readFile(ctx, repoName, revision, row.FileName)
		if err != nil {
			if !os.IsNotExist(err) {
				logger.Error("error reading file", log.String("repoName", string(repoName)), log.String("revision", string(revision)), log.String("fileName", row.FileName), log.Error(err))
			}
			continue
		}
		lines := strings.Split(string(fileContent), "\n")

		// Sanity check: check that startLine and endLine are within 0 and len(lines).
		startLine := max(0, min(len(lines), row.StartLine))
		endLine := max(0, min(len(lines), row.EndLine))

		rows[idx].Content = strings.Join(lines[startLine:endLine], "\n")
	}
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
	rows := index.SimilaritySearch(query, nResults, embeddings.WorkerOptions{NumWorkers: numWorkers, MinRowsToSplit: SIMILARITY_SEARCH_MIN_ROWS_TO_SPLIT}, opts)

	// Hydrate content
	for idx, row := range rows {
		fileContent, err := readFile(ctx, repoName, revision, row.FileName)
		if err != nil {
			if !os.IsNotExist(err) {
				logger.Error("error reading file", log.String("repoName", string(repoName)), log.String("revision", string(revision)), log.String("fileName", row.FileName), log.Error(err))
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
