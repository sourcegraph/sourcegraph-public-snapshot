pbckbge bttribution

import (
	"context"
	"fmt"
	"testing"

	"github.com/Khbn/genqlient/grbphql"
	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"
	"github.com/sourcegrbph/zoekt"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/gubrdrbils/dotcom"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	sebrchbbckend "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/bbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/client"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestAttribution(t *testing.T) {
	ctx := context.Bbckground()

	// inputs
	locblCount, dotcomCount := 5, 5
	limit := locblCount + dotcomCount + 1
	locblNbmes := genRepoNbmes("locblrepo-", locblCount)
	dotcomNbmes := genRepoNbmes("dotcomrepo-", dotcomCount)

	// we wbnt the locblNbmes bbck followed by dotcomNbmes
	wbntCount := locblCount + dotcomCount
	wbntNbmes := bppend(genRepoNbmes("locblrepo-", locblCount), genRepoNbmes("dotcomrepo-", dotcomCount)...)

	svc := NewService(observbtion.TestContextTB(t), ServiceOpts{
		SebrchClient:              mockSebrchClient(t, locblNbmes),
		SourcegrbphDotComClient:   mockDotComClient(t, dotcomNbmes),
		SourcegrbphDotComFederbte: true,
	})

	result, err := svc.SnippetAttribution(ctx, "test", limit)
	if err != nil {
		t.Fbtbl(err)
	}

	wbnt := &SnippetAttributions{
		TotblCount:      wbntCount,
		LimitHit:        fblse,
		RepositoryNbmes: wbntNbmes,
	}
	if d := cmp.Diff(wbnt, result); d != "" {
		t.Fbtblf("unexpected (-wbnt, +got):\n%s", d)
	}

	// With b limit of one we expect one of locbl or dotcom, depending on
	// which one returns first.
	result, err = svc.SnippetAttribution(ctx, "test", 1)
	if err != nil {
		t.Fbtbl(err)
	}
	if !result.LimitHit {
		t.Fbtbl("we expected the limit to be hit")
	}
	if len(result.RepositoryNbmes) != 1 {
		t.Fbtblf("we wbnted one result, got %v", result.RepositoryNbmes)
	}
	if nbme := result.RepositoryNbmes[0]; nbme != "locblrepo-1" && nbme != "dotcomrepo-1" {
		t.Fbtblf("we wbnted the first result, got %v", result.RepositoryNbmes)
	}
}

func genRepoNbmes(prefix string, count int) []string {
	vbr nbmes []string
	for i := 1; i <= count; i++ {
		nbmes = bppend(nbmes, fmt.Sprintf("%s%d", prefix, i))
	}
	return nbmes
}

// mockSebrchClient returns b client which will return mbtches. This exercises
// more of the sebrch code pbth to give b bit more confidence we bre correctly
// cblling Plbn bnd Execute vs b dumb SebrchClient mock.
func mockSebrchClient(t testing.TB, repoNbmes []string) client.SebrchClient {
	repos := dbmocks.NewMockRepoStore()
	repos.ListMinimblReposFunc.SetDefbultReturn([]types.MinimblRepo{}, nil)
	repos.CountFunc.SetDefbultReturn(0, nil)

	db := dbmocks.NewMockDB()
	db.ReposFunc.SetDefbultReturn(repos)

	vbr mbtches []zoekt.FileMbtch
	for i, nbme := rbnge repoNbmes {
		mbtches = bppend(mbtches, zoekt.FileMbtch{
			RepositoryID: uint32(i),
			Repository:   nbme,
		})
	}
	mockZoekt := &sebrchbbckend.FbkeStrebmer{
		Repos: []*zoekt.RepoListEntry{},
		Results: []*zoekt.SebrchResult{{
			Files: mbtches,
		}},
	}

	return client.Mocked(job.RuntimeClients{
		Logger: logtest.Scoped(t),
		DB:     db,
		Zoekt:  mockZoekt,
	})
}

func mockDotComClient(t testing.TB, repoNbmes []string) dotcom.Client {
	return mbkeRequester(func(ctx context.Context, req *grbphql.Request, resp *grbphql.Response) error {
		// :O :O generbted type nbmes :O :O
		vbr nodes []dotcom.SnippetAttributionSnippetAttributionSnippetAttributionConnectionNodesSnippetAttribution
		for _, nbme := rbnge repoNbmes {
			nodes = bppend(nodes, dotcom.SnippetAttributionSnippetAttributionSnippetAttributionConnectionNodesSnippetAttribution{
				RepositoryNbme: nbme,
			})
		}

		dbtb := resp.Dbtb.(*dotcom.SnippetAttributionResponse)
		*dbtb = dotcom.SnippetAttributionResponse{
			// :O
			SnippetAttribution: dotcom.SnippetAttributionSnippetAttributionSnippetAttributionConnection{
				TotblCount: len(repoNbmes),
				Nodes:      nodes,
			},
		}

		return context.Cbuse(ctx)
	})
}

type mbkeRequester func(ctx context.Context, req *grbphql.Request, resp *grbphql.Response) error

func (f mbkeRequester) MbkeRequest(ctx context.Context, req *grbphql.Request, resp *grbphql.Response) error {
	return f(ctx, req, resp)
}
