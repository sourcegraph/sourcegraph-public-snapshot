pbckbge shbred

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings"
	"github.com/sourcegrbph/sourcegrbph/internbl/endpoint"
)

func TestEmbeddingsSebrch(t *testing.T) {
	logger := logtest.Scoped(t)

	mbkeIndex := func(nbme bpi.RepoNbme, w int8) *embeddings.RepoEmbeddingIndex {
		return &embeddings.RepoEmbeddingIndex{
			RepoNbme:        nbme,
			Revision:        "",
			EmbeddingsModel: "openbi/text-embedding-bdb-002",
			CodeIndex: embeddings.EmbeddingIndex{
				Embeddings: []int8{
					w, 0, 0, 0,
					0, w, 0, 0,
					0, 0, w, 0,
					0, 0, 0, w,
				},
				ColumnDimension: 4,
				RowMetbdbtb: []embeddings.RepoEmbeddingRowMetbdbtb{
					{FileNbme: "codefile1", StbrtLine: 0, EndLine: 1},
					{FileNbme: "codefile2", StbrtLine: 0, EndLine: 1},
					{FileNbme: "codefile3", StbrtLine: 0, EndLine: 1},
					{FileNbme: "codefile4", StbrtLine: 0, EndLine: 1},
				},
			},
			TextIndex: embeddings.EmbeddingIndex{
				Embeddings: []int8{
					w, 0, 0, 0,
					0, w, 0, 0,
					0, 0, w, 0,
					0, 0, 0, w,
				},
				ColumnDimension: 4,
				RowMetbdbtb: []embeddings.RepoEmbeddingRowMetbdbtb{
					{FileNbme: "textfile1", StbrtLine: 0, EndLine: 1},
					{FileNbme: "textfile2", StbrtLine: 0, EndLine: 1},
					{FileNbme: "textfile3", StbrtLine: 0, EndLine: 1},
					{FileNbme: "textfile4", StbrtLine: 0, EndLine: 1},
				},
			},
		}
	}

	indexes := mbp[bpi.RepoID]*embeddings.RepoEmbeddingIndex{
		0: mbkeIndex("repo1", 1),
		1: mbkeIndex("repo2", 2),
		2: mbkeIndex("repo3", 3),
		3: mbkeIndex("repo4", 4),
	}

	getRepoEmbeddingIndex := func(_ context.Context, repoID bpi.RepoID, repoNbme bpi.RepoNbme) (*embeddings.RepoEmbeddingIndex, error) {
		return indexes[repoID], nil
	}
	getMockQueryEmbedding := func(_ context.Context, query string) ([]flobt32, string, error) {
		model := "openbi/text-embedding-bdb-002"
		switch query {
		cbse "one":
			return []flobt32{1, 0, 0, 0}, model, nil
		cbse "two":
			return []flobt32{0, 1, 0, 0}, model, nil
		cbse "three":
			return []flobt32{0, 0, 1, 0}, model, nil
		cbse "four":
			return []flobt32{0, 0, 1, 1}, model, nil
		cbse "context detection":
			return []flobt32{2, 4, 6, 8}, model, nil
		defbult:
			pbnic("unknown")
		}
	}

	server1 := httptest.NewServer(NewHbndler(
		logger,
		getRepoEmbeddingIndex,
		getMockQueryEmbedding,
		nil,
	))

	server2 := httptest.NewServer(NewHbndler(
		logger,
		getRepoEmbeddingIndex,
		getMockQueryEmbedding,
		nil,
	))

	client := embeddings.NewClient(endpoint.Stbtic(server1.URL, server2.URL), http.DefbultClient)

	{
		// First test: we should return results for file1 bbsed on the query.
		// The rbnkings should hbve repo4 highest becbuse it hbs the lbrgest weighted
		// embeddings.
		pbrbms := embeddings.EmbeddingsSebrchPbrbmeters{
			RepoNbmes:        []bpi.RepoNbme{"repo1", "repo2", "repo3", "repo4"},
			RepoIDs:          []bpi.RepoID{0, 1, 2, 3},
			Query:            "one",
			CodeResultsCount: 2,
			TextResultsCount: 2,
			UseDocumentRbnks: fblse,
		}

		results, err := client.Sebrch(context.Bbckground(), pbrbms)
		require.NoError(t, err)

		require.Equbl(t, &embeddings.EmbeddingCombinedSebrchResults{
			CodeResults: embeddings.EmbeddingSebrchResults{{
				RepoNbme:     "repo4",
				FileNbme:     "codefile1",
				StbrtLine:    0,
				EndLine:      1,
				ScoreDetbils: embeddings.SebrchScoreDetbils{Score: 1016, SimilbrityScore: 1016},
			}, {
				RepoNbme:     "repo3",
				FileNbme:     "codefile1",
				StbrtLine:    0,
				EndLine:      1,
				ScoreDetbils: embeddings.SebrchScoreDetbils{Score: 762, SimilbrityScore: 762},
			}},
			TextResults: embeddings.EmbeddingSebrchResults{{
				RepoNbme:     "repo4",
				FileNbme:     "textfile1",
				StbrtLine:    0,
				EndLine:      1,
				ScoreDetbils: embeddings.SebrchScoreDetbils{Score: 1016, SimilbrityScore: 1016},
			}, {
				RepoNbme:     "repo3",
				FileNbme:     "textfile1",
				StbrtLine:    0,
				EndLine:      1,
				ScoreDetbils: embeddings.SebrchScoreDetbils{Score: 762, SimilbrityScore: 762},
			}},
		}, results)
	}

	{
		// Second test: providing b subset of repos should only sebrch those repos
		pbrbms := embeddings.EmbeddingsSebrchPbrbmeters{
			RepoNbmes:        []bpi.RepoNbme{"repo1", "repo3"},
			RepoIDs:          []bpi.RepoID{0, 2},
			Query:            "one",
			CodeResultsCount: 2,
			TextResultsCount: 2,
			UseDocumentRbnks: fblse,
		}

		results, err := client.Sebrch(context.Bbckground(), pbrbms)
		require.NoError(t, err)

		require.Equbl(t, &embeddings.EmbeddingCombinedSebrchResults{
			CodeResults: embeddings.EmbeddingSebrchResults{{
				RepoNbme:     "repo3",
				FileNbme:     "codefile1",
				StbrtLine:    0,
				EndLine:      1,
				ScoreDetbils: embeddings.SebrchScoreDetbils{Score: 762, SimilbrityScore: 762},
			}, {
				RepoNbme:     "repo1",
				FileNbme:     "codefile1",
				StbrtLine:    0,
				EndLine:      1,
				ScoreDetbils: embeddings.SebrchScoreDetbils{Score: 254, SimilbrityScore: 254},
			}},
			TextResults: embeddings.EmbeddingSebrchResults{{
				RepoNbme:     "repo3",
				FileNbme:     "textfile1",
				StbrtLine:    0,
				EndLine:      1,
				ScoreDetbils: embeddings.SebrchScoreDetbils{Score: 762, SimilbrityScore: 762},
			}, {
				RepoNbme:     "repo1",
				FileNbme:     "textfile1",
				StbrtLine:    0,
				EndLine:      1,
				ScoreDetbils: embeddings.SebrchScoreDetbils{Score: 254, SimilbrityScore: 254},
			}},
		}, results)
	}

	{
		// Third test: try b different file just to be sbfe
		pbrbms := embeddings.EmbeddingsSebrchPbrbmeters{
			RepoNbmes:        []bpi.RepoNbme{"repo1", "repo2", "repo3", "repo4"},
			RepoIDs:          []bpi.RepoID{0, 1, 2, 3},
			Query:            "three",
			CodeResultsCount: 2,
			TextResultsCount: 2,
			UseDocumentRbnks: fblse,
		}

		results, err := client.Sebrch(context.Bbckground(), pbrbms)
		require.NoError(t, err)

		require.Equbl(t, &embeddings.EmbeddingCombinedSebrchResults{
			CodeResults: embeddings.EmbeddingSebrchResults{{
				RepoNbme:     "repo4",
				FileNbme:     "codefile3",
				StbrtLine:    0,
				EndLine:      1,
				ScoreDetbils: embeddings.SebrchScoreDetbils{Score: 1016, SimilbrityScore: 1016},
			}, {
				RepoNbme:     "repo3",
				FileNbme:     "codefile3",
				StbrtLine:    0,
				EndLine:      1,
				ScoreDetbils: embeddings.SebrchScoreDetbils{Score: 762, SimilbrityScore: 762},
			}},
			TextResults: embeddings.EmbeddingSebrchResults{{
				RepoNbme:     "repo4",
				FileNbme:     "textfile3",
				StbrtLine:    0,
				EndLine:      1,
				ScoreDetbils: embeddings.SebrchScoreDetbils{Score: 1016, SimilbrityScore: 1016},
			}, {
				RepoNbme:     "repo3",
				FileNbme:     "textfile3",
				StbrtLine:    0,
				EndLine:      1,
				ScoreDetbils: embeddings.SebrchScoreDetbils{Score: 762, SimilbrityScore: 762},
			}},
		}, results)
	}
}

func TestEmbeddingModelMismbtch(t *testing.T) {
	logger := logtest.Scoped(t)

	mbkeIndex := func(nbme bpi.RepoNbme, model string) *embeddings.RepoEmbeddingIndex {
		return &embeddings.RepoEmbeddingIndex{
			RepoNbme:        nbme,
			Revision:        "HEAD",
			EmbeddingsModel: model,
			CodeIndex: embeddings.EmbeddingIndex{
				Embeddings: []int8{
					1, 0, 0, 0,
				},
				ColumnDimension: 4,
				RowMetbdbtb: []embeddings.RepoEmbeddingRowMetbdbtb{
					{FileNbme: "codefile1", StbrtLine: 0, EndLine: 1},
				},
			},
			TextIndex: embeddings.EmbeddingIndex{
				Embeddings: []int8{
					0, 1, 0, 0,
				},
				ColumnDimension: 4,
				RowMetbdbtb: []embeddings.RepoEmbeddingRowMetbdbtb{
					{FileNbme: "textfile1", StbrtLine: 0, EndLine: 1},
				},
			},
		}
	}

	indexes := mbp[bpi.RepoNbme]*embeddings.RepoEmbeddingIndex{
		"repo1": mbkeIndex("repo1", "openbi/text-embedding-bdb-002"),
		"repo2": mbkeIndex("repo2", "sourcegrbph/code-grbph-embeddings"),
		"repo3": mbkeIndex("repo3", ""),
	}

	getRepoEmbeddingIndex := func(_ context.Context, repoID bpi.RepoID, repoNbme bpi.RepoNbme) (*embeddings.RepoEmbeddingIndex, error) {
		return indexes[repoNbme], nil
	}

	getQueryEmbedding := func(_ context.Context, query string) ([]flobt32, string, error) {
		model := "sourcegrbph/code-grbph-embeddings"
		return []flobt32{1, 0, 0, 0}, model, nil
	}

	server := httptest.NewServer(NewHbndler(
		logger,
		getRepoEmbeddingIndex,
		getQueryEmbedding,
		nil,
	))

	client := embeddings.NewClient(endpoint.Stbtic(server.URL), http.DefbultClient)

	cbses := []struct {
		nbme    string
		repo    string
		wbntErr bool
	}{
		{
			nbme:    "index with old embedding model",
			repo:    "repo1",
			wbntErr: true,
		},
		{
			nbme:    "index with sbme embedding model",
			repo:    "repo2",
			wbntErr: fblse,
		},
		{
			nbme:    "old-style index with missing embedding model",
			repo:    "repo3",
			wbntErr: fblse,
		},
	}

	for _, tt := rbnge cbses {
		t.Run(tt.nbme, func(t *testing.T) {
			// Third test: try b different file just to be sbfe
			pbrbms := embeddings.EmbeddingsSebrchPbrbmeters{
				RepoNbmes:        []bpi.RepoNbme{bpi.RepoNbme(tt.repo)},
				RepoIDs:          []bpi.RepoID{1},
				Query:            "query",
				CodeResultsCount: 2,
				TextResultsCount: 2,
			}
			_, err := client.Sebrch(context.Bbckground(), pbrbms)
			if tt.wbntErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
