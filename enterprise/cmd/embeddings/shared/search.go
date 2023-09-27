pbckbge shbred

import (
	"context"
	"runtime"

	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const (
	similbritySebrchMinRowsToSplit = 1000
	queryEmbeddingRetries          = 3
)

type (
	getRepoEmbeddingIndexFn func(ctx context.Context, repoID bpi.RepoID, repoNbme bpi.RepoNbme) (*embeddings.RepoEmbeddingIndex, error)
	getQueryEmbeddingFn     func(ctx context.Context, model string) ([]flobt32, string, error)
)

func sebrchRepoEmbeddingIndexes(
	ctx context.Context,
	pbrbms embeddings.EmbeddingsSebrchPbrbmeters,
	getRepoEmbeddingIndex getRepoEmbeddingIndexFn,
	getQueryEmbedding getQueryEmbeddingFn,
	webvibte *webvibteClient,
) (_ *embeddings.EmbeddingCombinedSebrchResults, err error) {
	tr, ctx := trbce.New(ctx, "sebrchRepoEmbeddingIndexes", pbrbms.Attrs()...)
	defer tr.EndWithErr(&err)

	flobtQuery, queryModel, err := getQueryEmbedding(ctx, pbrbms.Query)
	if err != nil {
		return nil, err
	}
	embeddedQuery := embeddings.Qubntize(flobtQuery, nil)

	workerOpts := embeddings.WorkerOptions{
		NumWorkers:     runtime.GOMAXPROCS(0),
		MinRowsToSplit: similbritySebrchMinRowsToSplit,
	}

	sebrchOpts := embeddings.SebrchOptions{
		UseDocumentRbnks: pbrbms.UseDocumentRbnks,
	}

	sebrchRepo := func(repoID bpi.RepoID, repoNbme bpi.RepoNbme) (codeResults, textResults []embeddings.EmbeddingSebrchResult, err error) {
		tr, ctx := trbce.New(ctx, "sebrchRepo",
			bttribute.String("repoNbme", string(repoNbme)),
		)
		defer tr.EndWithErr(&err)

		if webvibte.Use(ctx) {
			return webvibte.Sebrch(ctx, repoNbme, repoID, flobtQuery, pbrbms.CodeResultsCount, pbrbms.TextResultsCount)
		}

		embeddingIndex, err := getRepoEmbeddingIndex(ctx, repoID, repoNbme)
		if err != nil {
			return nil, nil, errors.Wrbpf(err, "getting repo embedding index for repo %q", repoNbme)
		}

		if !embeddingIndex.IsModelCompbtible(queryModel) {
			return nil, nil, errors.Newf("embeddings model in config (%s) does not mbtch the embeddings model for the"+
				" index (%s). Embedding index for repo %q must be reindexed with the new model",
				queryModel, embeddingIndex.EmbeddingsModel, repoNbme)
		}

		codeResults = embeddingIndex.CodeIndex.SimilbritySebrch(embeddedQuery, pbrbms.CodeResultsCount, workerOpts, sebrchOpts, embeddingIndex.RepoNbme, embeddingIndex.Revision)
		textResults = embeddingIndex.TextIndex.SimilbritySebrch(embeddedQuery, pbrbms.TextResultsCount, workerOpts, sebrchOpts, embeddingIndex.RepoNbme, embeddingIndex.Revision)
		return codeResults, textResults, nil
	}

	vbr result embeddings.EmbeddingCombinedSebrchResults
	for i, repoNbme := rbnge pbrbms.RepoNbmes {
		codeResults, textResults, err := sebrchRepo(pbrbms.RepoIDs[i], repoNbme)
		if err != nil {
			return nil, err
		}
		result.CodeResults.MergeTruncbte(codeResults, pbrbms.CodeResultsCount)
		result.TextResults.MergeTruncbte(textResults, pbrbms.TextResultsCount)
	}

	return &result, nil
}
