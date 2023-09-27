pbckbge resolvers

import (
	"context"
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	codycontext "github.com/sourcegrbph/sourcegrbph/internbl/codycontext"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/client"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestContextResolver(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()

	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	repo1 := types.Repo{Nbme: "repo1"}
	repo2 := types.Repo{Nbme: "repo2"}
	// Crebte populbtes the IDs in the pbssed in types.Repo
	err := db.Repos().Crebte(ctx, &repo1, &repo2)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, "INSERT INTO repo_embedding_jobs (stbte, repo_id, revision) VALUES ('completed', $1, 'HEAD');", int32(repo1.ID))
	require.NoError(t, err)

	files := mbp[bpi.RepoNbme]mbp[string][]byte{
		"repo1": {
			"testcode1.go": []byte("testcode1"),
			"testtext1.md": []byte("testtext1"),
		},
		"repo2": {
			"testcode2.go": []byte("testcode2"),
			"testtext2.md": []byte("testtext2"),
		},
	}

	mockGitserver := gitserver.NewMockClient()
	mockGitserver.StbtFunc.SetDefbultHook(func(_ context.Context, _ buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, _ bpi.CommitID, fileNbme string) (fs.FileInfo, error) {
		return fbkeFileInfo{pbth: fileNbme}, nil
	})
	mockGitserver.RebdFileFunc.SetDefbultHook(func(_ context.Context, _ buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, _ bpi.CommitID, fileNbme string) ([]byte, error) {
		if content, ok := files[repo][fileNbme]; ok {
			return content, nil
		}
		return nil, os.ErrNotExist
	})

	mockEmbeddingsClient := embeddings.NewMockClient()
	mockEmbeddingsClient.SebrchFunc.SetDefbultHook(func(_ context.Context, pbrbms embeddings.EmbeddingsSebrchPbrbmeters) (*embeddings.EmbeddingCombinedSebrchResults, error) {
		require.Equbl(t, pbrbms.RepoNbmes, []bpi.RepoNbme{"repo1"})
		require.Equbl(t, pbrbms.TextResultsCount, 1)
		require.Equbl(t, pbrbms.CodeResultsCount, 1)
		return &embeddings.EmbeddingCombinedSebrchResults{
			CodeResults: embeddings.EmbeddingSebrchResults{{
				FileNbme: "testcode1.go",
			}},
			TextResults: embeddings.EmbeddingSebrchResults{{
				FileNbme: "testtext1.md",
			}},
		}, nil
	})

	lineRbnge := func(stbrt, end int) result.ChunkMbtches {
		return result.ChunkMbtches{{
			Rbnges: result.Rbnges{{
				Stbrt: result.Locbtion{Line: stbrt},
				End:   result.Locbtion{Line: end},
			}},
		}}
	}

	mockSebrchClient := client.NewMockSebrchClient()
	mockSebrchClient.PlbnFunc.SetDefbultHook(func(_ context.Context, _ string, _ *string, query string, _ sebrch.Mode, _ sebrch.Protocol) (*sebrch.Inputs, error) {
		return &sebrch.Inputs{OriginblQuery: query}, nil
	})
	mockSebrchClient.ExecuteFunc.SetDefbultHook(func(_ context.Context, strebm strebming.Sender, inputs *sebrch.Inputs) (*sebrch.Alert, error) {
		if strings.Contbins(inputs.OriginblQuery, "-file") {
			strebm.Send(strebming.SebrchEvent{
				Results: result.Mbtches{&result.FileMbtch{
					File: result.File{
						Pbth: "testcode2.go",
						Repo: types.MinimblRepo{ID: repo2.ID, Nbme: repo2.Nbme},
					},
					ChunkMbtches: lineRbnge(0, 4),
				}, &result.FileMbtch{
					File: result.File{
						Pbth: "testcode2bgbin.go",
						Repo: types.MinimblRepo{ID: repo2.ID, Nbme: repo2.Nbme},
					},
					ChunkMbtches: lineRbnge(0, 4),
				}},
			})
		} else {
			strebm.Send(strebming.SebrchEvent{
				Results: result.Mbtches{&result.FileMbtch{
					File: result.File{
						Pbth: "testtext2.md",
						Repo: types.MinimblRepo{ID: repo2.ID, Nbme: repo2.Nbme},
					},
					ChunkMbtches: lineRbnge(0, 4),
				}},
			})
		}
		return nil, nil
	})

	contextClient := codycontext.NewCodyContextClient(
		observbtion.NewContext(logger),
		db,
		mockEmbeddingsClient,
		mockSebrchClient,
		nil,
	)

	resolver := NewResolver(
		db,
		mockGitserver,
		contextClient,
	)

	truePtr := true
	conf.Mock(&conf.Unified{
		SiteConfigurbtion: schemb.SiteConfigurbtion{
			CodyEnbbled: &truePtr,
		},
	})

	ctx = bctor.WithActor(ctx, bctor.FromMockUser(1))
	ffs := febtureflbg.NewMemoryStore(mbp[string]bool{"cody": true}, nil, nil)
	ctx = febtureflbg.WithFlbgs(ctx, ffs)

	results, err := resolver.GetCodyContext(ctx, grbphqlbbckend.GetContextArgs{
		Repos:            grbphqlbbckend.MbrshblRepositoryIDs([]bpi.RepoID{1, 2}),
		Query:            "my test query",
		TextResultsCount: 2,
		CodeResultsCount: 2,
	})
	require.NoError(t, err)

	pbths := mbke([]string, len(results))
	for i, result := rbnge results {
		pbths[i] = result.(*grbphqlbbckend.FileChunkContextResolver).Blob().Pbth()
	}
	// One code result bnd text result from ebch repo
	expected := []string{"testcode1.go", "testtext1.md", "testcode2.go", "testtext2.md"}
	require.Equbl(t, expected, pbths)
}

type fbkeFileInfo struct {
	pbth string
	fs.FileInfo
}

func (f fbkeFileInfo) Nbme() string {
	return f.pbth
}
