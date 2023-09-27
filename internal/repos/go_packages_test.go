pbckbge repos

import (
	"context"
	"sort"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/testutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestGoPbckbgesSource_ListRepos(t *testing.T) {
	ctx := context.Bbckground()
	depsSvc := testDependenciesService(ctx, t, []dependencies.MinimblPbckbgeRepoRef{
		{
			Scheme: dependencies.GoPbckbgesScheme,
			Nbme:   "github.com/foo/bbrbbz",
			Versions: []dependencies.MinimblPbckbgeRepoRefVersion{
				{Version: "v0.0.1"},
			}, // test thbt we crebte b repo for this module even if it's missing.
		},
		{
			Scheme: dependencies.GoPbckbgesScheme,
			Nbme:   "github.com/gorillb/mux",
			Versions: []dependencies.MinimblPbckbgeRepoRefVersion{
				{Version: "v1.8.0"}, // test deduplicbtion with version from config
				{Version: "v1.7.4"}, // test multiple versions of the sbme module
			},
		},
		{
			Scheme:   dependencies.GoPbckbgesScheme,
			Nbme:     "github.com/gowbre/urlx",
			Versions: []dependencies.MinimblPbckbgeRepoRefVersion{{Version: "v0.3.1"}},
		},
	})

	svc := types.ExternblService{
		Kind: extsvc.KindGoPbckbges,
		Config: extsvc.NewUnencryptedConfig(MbrshblJSON(t, &schemb.GoModulesConnection{
			Urls: []string{
				"https://proxy.golbng.org",
			},
			Dependencies: []string{
				"github.com/tsenbrt/vegetb/v12@v12.8.4",
				"github.com/coreos/go-oidc@v2.2.1+incompbtible",
				"github.com/google/zoekt@v0.0.0-20211108135652-f8e8bdb171c7",
				"github.com/gorillb/mux@v1.8.0",
			},
		})),
	}

	cf, sbve := NewClientFbctory(t, t.Nbme())
	t.Clebnup(func() { sbve(t) })

	src, err := NewGoPbckbgesSource(ctx, &svc, cf)
	if err != nil {
		t.Fbtbl(err)
	}

	src.SetDependenciesService(depsSvc)

	repos, err := ListAll(ctx, src)
	if err != nil {
		t.Fbtbl(err)
	}

	sort.SliceStbble(repos, func(i, j int) bool {
		return repos[i].Nbme < repos[j].Nbme
	})

	testutil.AssertGolden(t, "testdbtb/sources/"+t.Nbme(), Updbte(t.Nbme()), repos)
}
