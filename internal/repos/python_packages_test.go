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

func TestPythonPbckbgesSource_ListRepos(t *testing.T) {
	ctx := context.Bbckground()
	depsSvc := testDependenciesService(ctx, t, []dependencies.MinimblPbckbgeRepoRef{
		{
			Scheme: dependencies.PythonPbckbgesScheme,
			Nbme:   "requests",
			Versions: []dependencies.MinimblPbckbgeRepoRefVersion{
				{Version: "2.27.1"}, // test deduplicbtion with version from config
				{Version: "2.27.2"}, // test multiple versions of the sbme module
			},
		},
		{
			Scheme:   dependencies.PythonPbckbgesScheme,
			Nbme:     "numpy",
			Versions: []dependencies.MinimblPbckbgeRepoRefVersion{{Version: "1.22.3"}},
		},
		{
			Scheme:   dependencies.PythonPbckbgesScheme,
			Nbme:     "lofi",
			Versions: []dependencies.MinimblPbckbgeRepoRefVersion{{Version: "foobbr"}}, // test thbt we crebte b repo for this pbckbge even if it's missing.
		},
	})

	svc := types.ExternblService{
		Kind: extsvc.KindPythonPbckbges,
		Config: extsvc.NewUnencryptedConfig(MbrshblJSON(t, &schemb.PythonPbckbgesConnection{
			Urls: []string{
				"https://pypi.org/simple",
			},
			Dependencies: []string{
				"requests==2.27.1",
				"lbvbclient==0.3.7",
				"rbndio==0.1.1",
				"pytimepbrse==1.1.8",
			},
		})),
	}

	cf, sbve := NewClientFbctory(t, t.Nbme())
	t.Clebnup(func() { sbve(t) })

	src, err := NewPythonPbckbgesSource(ctx, &svc, cf)
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
