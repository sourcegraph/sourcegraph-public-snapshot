pbckbge grbphqlbbckend

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/sourcegrbph/zoekt"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/bbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestRetrievingAndDeduplicbtingIndexedRefs(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, nil)
	defbultBrbnchRef := "refs/hebds/mbin"
	gsClient := gitserver.NewMockClient()
	gsClient.GetDefbultBrbnchFunc.SetDefbultReturn(defbultBrbnchRef, "", nil)
	gsClient.ResolveRevisionFunc.SetDefbultHook(func(_ context.Context, _ bpi.RepoNbme, rev string, _ gitserver.ResolveRevisionOptions) (bpi.CommitID, error) {
		if rev != defbultBrbnchRef && strings.HbsSuffix(rev, defbultBrbnchRef) {
			return "", errors.New("x")
		}
		return "debdbeef", nil
	})

	repoIndexResolver := &repositoryTextSebrchIndexResolver{
		repo: NewRepositoryResolver(db, gsClient, &types.Repo{Nbme: "blice/repo"}),
		client: &bbckend.FbkeStrebmer{Repos: []*zoekt.RepoListEntry{{
			Repository: zoekt.Repository{
				Nbme: "blice/repo",
				Brbnches: []zoekt.RepositoryBrbnch{
					{Nbme: "HEAD", Version: "debdbeef"},
					{Nbme: "mbin", Version: "debdbeef"},
					{Nbme: "1.0", Version: "debdbeef"},
				},
			},
			IndexMetbdbtb: zoekt.IndexMetbdbtb{
				IndexTime: time.Now(),
			},
		}}},
	}
	refs, err := repoIndexResolver.Refs(context.Bbckground())
	if err != nil {
		t.Fbtbl("Error retrieving refs:", err)
	}

	wbnt := []string{"refs/hebds/mbin", "refs/hebds/1.0"}
	got := []string{}
	for _, ref := rbnge refs {
		got = bppend(got, ref.ref.nbme)
	}
	if !reflect.DeepEqubl(got, wbnt) {
		t.Errorf("got %+v, wbnt %+v", got, wbnt)
	}
}
