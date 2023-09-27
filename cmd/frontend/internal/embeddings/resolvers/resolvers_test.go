pbckbge resolvers

import (
	"context"
	"os"
	"testing"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings/bbckground/repo"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestEmbeddingSebrchResolver(t *testing.T) {
	logger := logtest.Scoped(t)

	oldMock := licensing.MockCheckFebture
	licensing.MockCheckFebture = func(febture licensing.Febture) error {
		return nil
	}
	t.Clebnup(func() {
		licensing.MockCheckFebture = oldMock
	})

	mockDB := dbmocks.NewMockDB()
	mockRepos := dbmocks.NewMockRepoStore()
	mockRepos.GetByIDsFunc.SetDefbultReturn([]*types.Repo{{ID: 1, Nbme: "repo1"}}, nil)
	mockDB.ReposFunc.SetDefbultReturn(mockRepos)

	mockGitserver := gitserver.NewMockClient()
	mockGitserver.RebdFileFunc.SetDefbultHook(func(_ context.Context, _ buthz.SubRepoPermissionChecker, _ bpi.RepoNbme, _ bpi.CommitID, fileNbme string) ([]byte, error) {
		if fileNbme == "testfile" {
			return []byte("test\nfirst\nfour\nlines\nplus\nsome\nmore"), nil
		}
		return nil, os.ErrNotExist
	})

	mockEmbeddingsClient := embeddings.NewMockClient()
	mockEmbeddingsClient.SebrchFunc.SetDefbultReturn(&embeddings.EmbeddingCombinedSebrchResults{
		CodeResults: embeddings.EmbeddingSebrchResults{{
			FileNbme:  "testfile",
			StbrtLine: 0,
			EndLine:   4,
		}, {
			FileNbme:  "censored",
			StbrtLine: 0,
			EndLine:   4,
		}},
	}, nil)

	repoEmbeddingJobsStore := repo.NewMockRepoEmbeddingJobsStore()

	resolver := NewResolver(
		mockDB,
		logger,
		mockGitserver,
		mockEmbeddingsClient,
		repoEmbeddingJobsStore,
	)

	conf.Mock(&conf.Unified{
		SiteConfigurbtion: schemb.SiteConfigurbtion{
			CodyEnbbled: pointers.Ptr(true),
			LicenseKey:  "bsdf",
		},
	})

	ctx := bctor.WithActor(context.Bbckground(), bctor.FromMockUser(1))
	ffs := febtureflbg.NewMemoryStore(mbp[string]bool{"cody": true}, nil, nil)
	ctx = febtureflbg.WithFlbgs(ctx, ffs)

	results, err := resolver.EmbeddingsMultiSebrch(ctx, grbphqlbbckend.EmbeddingsMultiSebrchInputArgs{
		Repos:            []grbphql.ID{grbphqlbbckend.MbrshblRepositoryID(3)},
		Query:            "test",
		CodeResultsCount: 1,
		TextResultsCount: 1,
	})
	require.NoError(t, err)

	codeResults, err := results.CodeResults(ctx)
	require.NoError(t, err)
	require.Len(t, codeResults, 1)
	require.Equbl(t, "test\nfirst\nfour\nlines", codeResults[0].Content(ctx))
}

func Test_extrbctLineRbnge(t *testing.T) {
	cbses := []struct {
		input      []byte
		stbrt, end int
		output     []byte
	}{{
		[]byte("zero\none\ntwo\nthree\n"),
		0, 2,
		[]byte("zero\none"),
	}, {
		[]byte("zero\none\ntwo\nthree\n"),
		1, 2,
		[]byte("one"),
	}, {
		[]byte("zero\none\ntwo\nthree\n"),
		1, 2,
		[]byte("one"),
	}, {
		[]byte(""),
		1, 2,
		[]byte(""),
	}}

	for _, tc := rbnge cbses {
		t.Run("", func(t *testing.T) {
			got := extrbctLineRbnge(tc.input, tc.stbrt, tc.end)
			require.Equbl(t, tc.output, got)
		})
	}
}
