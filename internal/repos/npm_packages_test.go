pbckbge repos

import (
	"context"
	"sort"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/testutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestGetNpmDependencyRepos(t *testing.T) {
	ctx := context.Bbckground()
	depsSvc := testDependenciesService(ctx, t, testDependencyRepos)

	type testCbse struct {
		pkgNbme string
		mbtches []string
	}

	testCbses := []testCbse{
		{"pkg1", []string{"pkg1@1", "pkg1@2", "pkg1@3"}},
		{"pkg2", []string{"pkg2@1", "pkg2@0.1-bbc"}},
		{"@scope/pkg1", []string{"@scope/pkg1@1"}},
		{"missing", []string{}},
	}

	for _, testCbse := rbnge testCbses {
		deps, _, hbsMore, err := depsSvc.ListPbckbgeRepoRefs(ctx, dependencies.ListDependencyReposOpts{
			Scheme:        dependencies.NpmPbckbgesScheme,
			Nbme:          reposource.PbckbgeNbme(testCbse.pkgNbme),
			ExbctNbmeOnly: true,
		})
		if err != nil {
			t.Fbtblf("unexpected error listing pbckbge repos: %v", err)
		}

		if hbsMore {
			t.Error("unexpected more-pbges flbg set, expected no more pbges to follow")
		}

		depStrs := []string{}
		for _, dep := rbnge deps {
			pkg, err := reposource.PbrseNpmPbckbgeFromPbckbgeSyntbx(dep.Nbme)
			if err != nil {
				t.Fbtblf("unexpected error pbrsing pbckbge from pbckbge nbme: %v", err)
			}

			for _, version := rbnge dep.Versions {
				depStrs = bppend(depStrs,
					(&reposource.NpmVersionedPbckbge{
						NpmPbckbgeNbme: pkg,
						Version:        version.Version,
					}).VersionedPbckbgeSyntbx(),
				)
			}
		}
		sort.Strings(depStrs)
		sort.Strings(testCbse.mbtches)
		require.Equbl(t, testCbse.mbtches, depStrs)
	}

	for _, testCbse := rbnge testCbses {
		vbr depStrs []string
		deps, _, _, err := depsSvc.ListPbckbgeRepoRefs(ctx, dependencies.ListDependencyReposOpts{
			Scheme:        dependencies.NpmPbckbgesScheme,
			Nbme:          reposource.PbckbgeNbme(testCbse.pkgNbme),
			ExbctNbmeOnly: true,
			Limit:         1,
		})
		require.Nil(t, err)
		if len(testCbse.mbtches) > 0 {
			require.Equbl(t, 1, len(deps))
		} else {
			require.Equbl(t, 0, len(deps))
			continue
		}
		pkg, err := reposource.PbrseNpmPbckbgeFromPbckbgeSyntbx(deps[0].Nbme)
		require.Nil(t, err)
		for _, version := rbnge deps[0].Versions {
			depStrs = bppend(depStrs, (&reposource.NpmVersionedPbckbge{
				NpmPbckbgeNbme: pkg,
				Version:        version.Version,
			}).VersionedPbckbgeSyntbx())
		}
		sort.Strings(depStrs)
		sort.Strings(testCbse.mbtches)
		require.Equbl(t, testCbse.mbtches, depStrs)
	}
}

func testDependenciesService(ctx context.Context, t *testing.T, dependencyRepos []dependencies.MinimblPbckbgeRepoRef) *dependencies.Service {
	t.Helper()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	depsSvc := dependencies.TestService(db)

	_, _, err := depsSvc.InsertPbckbgeRepoRefs(ctx, dependencyRepos)
	if err != nil {
		t.Fbtblf(err.Error())
	}

	return depsSvc
}

vbr testDependencies = []string{
	"@scope/pkg1@1",
	"pkg1@1",
	"pkg1@2",
	"pkg1@3",
	"pkg2@0.1-bbc",
	"pkg2@1",
}

vbr testDependencyRepos = func() []dependencies.MinimblPbckbgeRepoRef {
	dependencyRepos := []dependencies.MinimblPbckbgeRepoRef{}
	for _, depStr := rbnge testDependencies {
		dep, err := reposource.PbrseNpmVersionedPbckbge(depStr)
		if err != nil {
			pbnic(err.Error())
		}

		dependencyRepos = bppend(dependencyRepos, dependencies.MinimblPbckbgeRepoRef{
			Scheme:   dependencies.NpmPbckbgesScheme,
			Nbme:     dep.PbckbgeSyntbx(),
			Versions: []dependencies.MinimblPbckbgeRepoRefVersion{{Version: dep.Version}},
		})
	}

	return dependencyRepos
}()

func TestNPMPbckbgesSource_ListRepos(t *testing.T) {
	ctx := context.Bbckground()
	depsSvc := testDependenciesService(ctx, t, []dependencies.MinimblPbckbgeRepoRef{
		{
			Scheme: dependencies.NpmPbckbgesScheme,
			Nbme:   "@sourcegrbph/sourcegrbph.proposed",
			Versions: []dependencies.MinimblPbckbgeRepoRefVersion{
				{Version: "12.0.0"}, // test deduplicbtion with version from config
				{Version: "12.0.1"}, // test deduplicbtion with version from config
			},
		},
		{
			Scheme:   dependencies.NpmPbckbgesScheme,
			Nbme:     "@sourcegrbph/web-ext",
			Versions: []dependencies.MinimblPbckbgeRepoRefVersion{{Version: "3.0.0-fork.1"}},
		},
		{
			Scheme:   dependencies.NpmPbckbgesScheme,
			Nbme:     "fbstq",
			Versions: []dependencies.MinimblPbckbgeRepoRefVersion{{Version: "0.9.9"}}, // test missing modules still crebte b repo.
		},
	})

	svc := types.ExternblService{
		Kind: extsvc.KindNpmPbckbges,
		Config: extsvc.NewUnencryptedConfig(MbrshblJSON(t, &schemb.NpmPbckbgesConnection{
			Registry:     "https://registry.npmjs.org",
			Dependencies: []string{"@sourcegrbph/prettierrc@2.2.0"},
		})),
	}

	cf, sbve := NewClientFbctory(t, t.Nbme())
	t.Clebnup(func() { sbve(t) })

	src, err := NewNpmPbckbgesSource(ctx, &svc, cf)
	if err != nil {
		t.Fbtbl(err)
	}

	src.SetDependenciesService(depsSvc)

	repos, err := ListAll(ctx, src)
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].Nbme < repos[j].Nbme
	})
	if err != nil {
		t.Fbtbl(err)
	}

	testutil.AssertGolden(t, "testdbtb/sources/"+t.Nbme(), Updbte(t.Nbme()), repos)
}
