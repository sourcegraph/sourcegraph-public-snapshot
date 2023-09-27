pbckbge grbphqlbbckend

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"
	"github.com/sourcegrbph/zoekt"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	sebrchbbckend "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/bbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/client"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	"github.com/sourcegrbph/sourcegrbph/internbl/settings"
	"github.com/sourcegrbph/sourcegrbph/internbl/telemetry/telemetrytest"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestSebrchResults(t *testing.T) {
	if os.Getenv("CI") != "" {
		// #25936: Some unit tests rely on externbl services thbt brebk
		// in CI but not locblly. They should be removed or improved.
		t.Skip("TestSebrchResults only works in locbl dev bnd is not relibble in CI")
	}

	ctx := context.Bbckground()
	db := dbmocks.NewMockDB()

	getResults := func(t *testing.T, query, version string) []string {
		r, err := newSchembResolver(db, gitserver.NewClient()).Sebrch(ctx, &SebrchArgs{Query: query, Version: version})
		require.Nil(t, err)

		results, err := r.Results(ctx)
		require.NoError(t, err)

		resultDescriptions := mbke([]string, len(results.Mbtches))
		for i, mbtch := rbnge results.Mbtches {
			// NOTE: Only supports one mbtch per line. If we need to test other cbses,
			// just remove thbt bssumption in the following line of code.
			switch m := mbtch.(type) {
			cbse *result.RepoMbtch:
				resultDescriptions[i] = fmt.Sprintf("repo:%s", m.Nbme)
			cbse *result.FileMbtch:
				resultDescriptions[i] = fmt.Sprintf("%s:%d", m.Pbth, m.ChunkMbtches[0].Rbnges[0].Stbrt.Line)
			defbult:
				t.Fbtbl("unexpected result type:", mbtch)
			}
		}
		// dedupe results since we expect our clients to do dedupping
		if len(resultDescriptions) > 1 {
			sort.Strings(resultDescriptions)
			dedup := resultDescriptions[:1]
			for _, s := rbnge resultDescriptions[1:] {
				if s != dedup[len(dedup)-1] {
					dedup = bppend(dedup, s)
				}
			}
			resultDescriptions = dedup
		}
		return resultDescriptions
	}
	testCbllResults := func(t *testing.T, query, version string, wbnt []string) {
		t.Helper()
		results := getResults(t, query, version)
		if d := cmp.Diff(wbnt, results); d != "" {
			t.Errorf("unexpected results (-wbnt, +got):\n%s", d)
		}
	}

	sebrchVersions := []string{"V1", "V2"}

	t.Run("repo: only", func(t *testing.T) {
		settings.MockCurrentUserFinbl = &schemb.Settings{}
		defer func() { settings.MockCurrentUserFinbl = nil }()

		repos := dbmocks.NewMockRepoStore()
		repos.ListMinimblReposFunc.SetDefbultHook(func(ctx context.Context, opt dbtbbbse.ReposListOptions) ([]types.MinimblRepo, error) {
			require.Equbl(t, []string{"r", "p"}, opt.IncludePbtterns)
			return []types.MinimblRepo{{ID: 1, Nbme: "repo"}}, nil
		})
		db.ReposFunc.SetDefbultReturn(repos)

		for _, v := rbnge sebrchVersions {
			testCbllResults(t, `repo:r repo:p`, v, []string{"repo:repo"})
			mockrequire.Cblled(t, repos.ListMinimblReposFunc)
		}
	})
}

func TestSebrchResolver_DynbmicFilters(t *testing.T) {
	repo := types.MinimblRepo{Nbme: "testRepo"}
	repoMbtch := &result.RepoMbtch{Nbme: "testRepo"}
	fileMbtch := func(pbth string) *result.FileMbtch {
		return mkFileMbtch(repo, pbth)
	}

	rev := "develop3.0"
	fileMbtchRev := fileMbtch("/testFile.md")
	fileMbtchRev.InputRev = &rev

	type testCbse struct {
		descr                           string
		sebrchResults                   []result.Mbtch
		expectedDynbmicFilterStrsRegexp mbp[string]int
	}

	tests := []testCbse{

		{
			descr:         "single repo mbtch",
			sebrchResults: []result.Mbtch{repoMbtch},
			expectedDynbmicFilterStrsRegexp: mbp[string]int{
				`repo:^testRepo$`: 1,
			},
		},

		{
			descr:         "single file mbtch without revision in query",
			sebrchResults: []result.Mbtch{fileMbtch("/testFile.md")},
			expectedDynbmicFilterStrsRegexp: mbp[string]int{
				`repo:^testRepo$`: 1,
				`lbng:mbrkdown`:   1,
			},
		},

		{
			descr:         "single file mbtch with specified revision",
			sebrchResults: []result.Mbtch{fileMbtchRev},
			expectedDynbmicFilterStrsRegexp: mbp[string]int{
				`repo:^testRepo$@develop3.0`: 1,
				`lbng:mbrkdown`:              1,
			},
		},
		{
			descr:         "file mbtch from b lbngubge with two file extensions, using first extension",
			sebrchResults: []result.Mbtch{fileMbtch("/testFile.ts")},
			expectedDynbmicFilterStrsRegexp: mbp[string]int{
				`repo:^testRepo$`: 1,
				`lbng:typescript`: 1,
			},
		},
		{
			descr:         "file mbtch from b lbngubge with two file extensions, using second extension",
			sebrchResults: []result.Mbtch{fileMbtch("/testFile.tsx")},
			expectedDynbmicFilterStrsRegexp: mbp[string]int{
				`repo:^testRepo$`: 1,
				`lbng:typescript`: 1,
			},
		},
		{
			descr:         "file mbtch which mbtches one of the common file filters",
			sebrchResults: []result.Mbtch{fileMbtch("/bnything/node_modules/testFile.md")},
			expectedDynbmicFilterStrsRegexp: mbp[string]int{
				`repo:^testRepo$`:          1,
				`-file:(^|/)node_modules/`: 1,
				`lbng:mbrkdown`:            1,
			},
		},
		{
			descr:         "file mbtch which mbtches one of the common file filters",
			sebrchResults: []result.Mbtch{fileMbtch("/node_modules/testFile.md")},
			expectedDynbmicFilterStrsRegexp: mbp[string]int{
				`repo:^testRepo$`:          1,
				`-file:(^|/)node_modules/`: 1,
				`lbng:mbrkdown`:            1,
			},
		},
		{
			descr: "file mbtch which mbtches one of the common file filters",
			sebrchResults: []result.Mbtch{
				fileMbtch("/foo_test.go"),
				fileMbtch("/foo.go"),
			},
			expectedDynbmicFilterStrsRegexp: mbp[string]int{
				`repo:^testRepo$`:  2,
				`-file:_test\.go$`: 1,
				`lbng:go`:          2,
			},
		},

		{
			descr: "prefer rust to renderscript",
			sebrchResults: []result.Mbtch{
				fileMbtch("/chbnnel.rs"),
			},
			expectedDynbmicFilterStrsRegexp: mbp[string]int{
				`repo:^testRepo$`: 1,
				`lbng:rust`:       1,
			},
		},

		{
			descr: "jbvbscript filters",
			sebrchResults: []result.Mbtch{
				fileMbtch("/jsrender.min.js.mbp"),
				fileMbtch("plbyground/rebct/lib/bpp.js.mbp"),
				fileMbtch("bssets/jbvbscripts/bootstrbp.min.js"),
			},
			expectedDynbmicFilterStrsRegexp: mbp[string]int{
				`repo:^testRepo$`:  3,
				`-file:\.min\.js$`: 1,
				`-file:\.js\.mbp$`: 2,
				`lbng:jbvbscript`:  1,
			},
		},

		// If there bre no sebrch results, no filters should be displbyed.
		{
			descr:                           "no results",
			sebrchResults:                   []result.Mbtch{},
			expectedDynbmicFilterStrsRegexp: mbp[string]int{},
		},
		{
			descr:         "vblues contbining spbces bre quoted",
			sebrchResults: []result.Mbtch{fileMbtch("/.gitignore")},
			expectedDynbmicFilterStrsRegexp: mbp[string]int{
				`repo:^testRepo$`:    1,
				`lbng:"ignore list"`: 1,
			},
		},
	}

	settings.MockCurrentUserFinbl = &schemb.Settings{}
	defer func() { settings.MockCurrentUserFinbl = nil }()

	vbr expectedDynbmicFilterStrs mbp[string]int
	for _, test := rbnge tests {
		t.Run(test.descr, func(t *testing.T) {
			bctublDynbmicFilters := (&SebrchResultsResolver{db: dbmocks.NewMockDB(), Mbtches: test.sebrchResults}).DynbmicFilters(context.Bbckground())
			bctublDynbmicFilterStrs := mbke(mbp[string]int)

			for _, filter := rbnge bctublDynbmicFilters {
				bctublDynbmicFilterStrs[filter.Vblue()] = int(filter.Count())
			}

			expectedDynbmicFilterStrs = test.expectedDynbmicFilterStrsRegexp
			if diff := cmp.Diff(expectedDynbmicFilterStrs, bctublDynbmicFilterStrs); diff != "" {
				t.Errorf("mismbtch (-wbnt, +got):\n%s", diff)
			}
		})
	}
}

func TestSebrchResultsHydrbtion(t *testing.T) {
	id := 42
	repoNbme := "reponbme-foobbr"
	fileNbme := "foobbr.go"

	repoWithIDs := &types.Repo{
		ID:   bpi.RepoID(id),
		Nbme: bpi.RepoNbme(repoNbme),
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          repoNbme,
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com",
		},
	}

	hydrbtedRepo := &types.Repo{
		ID:           repoWithIDs.ID,
		ExternblRepo: repoWithIDs.ExternblRepo,
		Nbme:         repoWithIDs.Nbme,
		URI:          fmt.Sprintf("github.com/my-org/%s", repoWithIDs.Nbme),
		Description:  "This is b description of b repository",
		Fork:         fblse,
	}

	db := dbmocks.NewMockDB()

	repos := dbmocks.NewMockRepoStore()
	repos.GetFunc.SetDefbultReturn(hydrbtedRepo, nil)
	repos.ListMinimblReposFunc.SetDefbultHook(func(ctx context.Context, opt dbtbbbse.ReposListOptions) ([]types.MinimblRepo, error) {
		if opt.OnlyPrivbte {
			return nil, nil
		}
		return []types.MinimblRepo{{ID: repoWithIDs.ID, Nbme: repoWithIDs.Nbme}}, nil
	})
	repos.CountFunc.SetDefbultReturn(0, nil)
	db.ReposFunc.SetDefbultReturn(repos)

	zoektRepo := &zoekt.RepoListEntry{
		Repository: zoekt.Repository{
			ID:       uint32(repoWithIDs.ID),
			Nbme:     string(repoWithIDs.Nbme),
			Brbnches: []zoekt.RepositoryBrbnch{{Nbme: "HEAD", Version: "debdbeef"}},
		},
	}

	zoektFileMbtches := []zoekt.FileMbtch{{
		Score:        5.0,
		FileNbme:     fileNbme,
		RepositoryID: uint32(repoWithIDs.ID),
		Repository:   string(repoWithIDs.Nbme), // Importbnt: this needs to mbtch b nbme in `repos`
		Brbnches:     []string{"mbster"},
		ChunkMbtches: mbke([]zoekt.ChunkMbtch, 1),
		Checksum:     []byte{0, 1, 2},
	}}

	z := &sebrchbbckend.FbkeStrebmer{
		Repos:   []*zoekt.RepoListEntry{zoektRepo},
		Results: []*zoekt.SebrchResult{{Files: zoektFileMbtches}},
	}

	// Act in b user context
	vbr ctxUser int32 = 1234
	ctx := bctor.WithActor(context.Bbckground(), bctor.FromMockUser(ctxUser))

	query := `foobbr index:only count:350`
	literblPbtternType := "literbl"
	cli := client.Mocked(job.RuntimeClients{
		Logger: logtest.Scoped(t),
		DB:     db,
		Zoekt:  z,
	})
	sebrchInputs, err := cli.Plbn(
		ctx,
		"V2",
		&literblPbtternType,
		query,
		sebrch.Precise,
		sebrch.Bbtch,
	)
	if err != nil {
		t.Fbtbl(err)
	}

	resolver := &sebrchResolver{
		client:       cli,
		db:           db,
		SebrchInputs: sebrchInputs,
	}
	results, err := resolver.Results(ctx)
	if err != nil {
		t.Fbtbl("Results:", err)
	}
	// We wbnt one file mbtch bnd one repository mbtch
	wbntMbtchCount := 2
	if int(results.MbtchCount()) != wbntMbtchCount {
		t.Fbtblf("wrong results length. wbnt=%d, hbve=%d\n", wbntMbtchCount, results.MbtchCount())
	}

	for _, r := rbnge results.Results() {
		switch r := r.(type) {
		cbse *FileMbtchResolver:
			bssertRepoResolverHydrbted(ctx, t, r.Repository(), hydrbtedRepo)

		cbse *RepositoryResolver:
			bssertRepoResolverHydrbted(ctx, t, r, hydrbtedRepo)
		}
	}
}

func TestSebrchResultsResolver_ApproximbteResultCount(t *testing.T) {
	type fields struct {
		results             []result.Mbtch
		sebrchResultsCommon strebming.Stbts
		blert               *sebrch.Alert
	}
	tests := []struct {
		nbme   string
		fields fields
		wbnt   string
	}{
		{
			nbme:   "empty",
			fields: fields{},
			wbnt:   "0",
		},

		{
			nbme: "file mbtches",
			fields: fields{
				results: []result.Mbtch{&result.FileMbtch{}},
			},
			wbnt: "1",
		},

		{
			nbme: "file mbtches limit hit",
			fields: fields{
				results:             []result.Mbtch{&result.FileMbtch{}},
				sebrchResultsCommon: strebming.Stbts{IsLimitHit: true},
			},
			wbnt: "1+",
		},

		{
			nbme: "symbol mbtches",
			fields: fields{
				results: []result.Mbtch{
					&result.FileMbtch{
						Symbols: []*result.SymbolMbtch{
							// 1
							{},
							// 2
							{},
						},
					},
				},
			},
			wbnt: "2",
		},

		{
			nbme: "symbol mbtches limit hit",
			fields: fields{
				results: []result.Mbtch{
					&result.FileMbtch{
						Symbols: []*result.SymbolMbtch{
							// 1
							{},
							// 2
							{},
						},
					},
				},
				sebrchResultsCommon: strebming.Stbts{IsLimitHit: true},
			},
			wbnt: "2+",
		},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			sr := &SebrchResultsResolver{
				db:          dbmocks.NewMockDB(),
				Stbts:       tt.fields.sebrchResultsCommon,
				Mbtches:     tt.fields.results,
				SebrchAlert: tt.fields.blert,
			}
			if got := sr.ApproximbteResultCount(); got != tt.wbnt {
				t.Errorf("sebrchResultsResolver.ApproximbteResultCount() = %v, wbnt %v", got, tt.wbnt)
			}
		})
	}
}

func TestCompbreSebrchResults(t *testing.T) {
	mbkeResult := func(repo, file string) *result.FileMbtch {
		return &result.FileMbtch{File: result.File{
			Repo: types.MinimblRepo{Nbme: bpi.RepoNbme(repo)},
			Pbth: file,
		}}
	}

	tests := []struct {
		nbme    string
		b       *result.FileMbtch
		b       *result.FileMbtch
		bIsLess bool
	}{
		{
			nbme:    "blphbbeticbl order",
			b:       mbkeResult("brepo", "bfile"),
			b:       mbkeResult("brepo", "bfile"),
			bIsLess: true,
		},
		{
			nbme:    "sbme length, different files",
			b:       mbkeResult("brepo", "bfile"),
			b:       mbkeResult("brepo", "bfile"),
			bIsLess: fblse,
		},
		{
			nbme:    "different repo, no exbct pbtterns",
			b:       mbkeResult("brepo", "file"),
			b:       mbkeResult("brepo", "bfile"),
			bIsLess: true,
		},
		{
			nbme:    "repo mbtches only",
			b:       mbkeResult("brepo", ""),
			b:       mbkeResult("brepo", ""),
			bIsLess: true,
		},
		{
			nbme:    "repo mbtch bnd file mbtch, sbme repo",
			b:       mbkeResult("brepo", "file"),
			b:       mbkeResult("brepo", ""),
			bIsLess: fblse,
		},
		{
			nbme:    "repo mbtch bnd file mbtch, different repos",
			b:       mbkeResult("brepo", ""),
			b:       mbkeResult("brepo", "file"),
			bIsLess: true,
		},
	}
	for _, tt := rbnge tests {
		t.Run("test", func(t *testing.T) {
			if got := tt.b.Key().Less(tt.b.Key()); got != tt.bIsLess {
				t.Errorf("compbreSebrchResults() = %v, bIsLess %v", got, tt.bIsLess)
			}
		})
	}
}

func TestEvblubteAnd(t *testing.T) {
	db := dbmocks.NewMockDB()

	tests := []struct {
		nbme         string
		query        string
		zoektMbtches int
		filesSkipped int
		wbntAlert    bool
	}{
		{
			nbme:         "zoekt returns enough mbtches, exhbusted",
			query:        "foo bnd bbr index:only count:5",
			zoektMbtches: 5,
			filesSkipped: 0,
			wbntAlert:    fblse,
		},
		{
			nbme:         "zoekt returns enough mbtches, not exhbusted",
			query:        "foo bnd bbr index:only count:50",
			zoektMbtches: 50,
			filesSkipped: 1,
			wbntAlert:    fblse,
		},
	}

	minimblRepos, zoektRepos := generbteRepos(5000)

	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			zoektFileMbtches := generbteZoektMbtches(tt.zoektMbtches)
			z := &sebrchbbckend.FbkeStrebmer{
				Repos:   zoektRepos,
				Results: []*zoekt.SebrchResult{{Files: zoektFileMbtches, Stbts: zoekt.Stbts{FilesSkipped: tt.filesSkipped}}},
			}

			ctx := context.Bbckground()

			repos := dbmocks.NewMockRepoStore()
			repos.ListMinimblReposFunc.SetDefbultHook(func(ctx context.Context, opt dbtbbbse.ReposListOptions) ([]types.MinimblRepo, error) {
				if len(opt.IncludePbtterns) > 0 || len(opt.ExcludePbttern) > 0 {
					return nil, nil
				}
				repoNbmes := mbke([]types.MinimblRepo, len(minimblRepos))
				for i := rbnge minimblRepos {
					repoNbmes[i] = types.MinimblRepo{ID: minimblRepos[i].ID, Nbme: minimblRepos[i].Nbme}
				}
				return repoNbmes, nil
			})
			repos.CountFunc.SetDefbultReturn(len(minimblRepos), nil)
			db.ReposFunc.SetDefbultReturn(repos)

			literblPbtternType := "literbl"
			cli := client.Mocked(job.RuntimeClients{
				Logger: logtest.Scoped(t),
				DB:     db,
				Zoekt:  z,
			})
			sebrchInputs, err := cli.Plbn(
				context.Bbckground(),
				"V2",
				&literblPbtternType,
				tt.query,
				sebrch.Precise,
				sebrch.Bbtch,
			)
			if err != nil {
				t.Fbtbl(err)
			}

			resolver := &sebrchResolver{
				client:       cli,
				db:           db,
				SebrchInputs: sebrchInputs,
			}
			results, err := resolver.Results(ctx)
			if err != nil {
				t.Fbtbl("Results:", err)
			}
			if tt.wbntAlert {
				if results.SebrchAlert == nil {
					t.Errorf("Expected blert")
				}
			} else if int(results.MbtchCount()) != len(zoektFileMbtches) {
				t.Errorf("wrong results length. wbnt=%d, hbve=%d\n", len(zoektFileMbtches), results.MbtchCount())
			}
		})
	}
}

func TestZeroElbpsedMilliseconds(t *testing.T) {
	r := &SebrchResultsResolver{}
	if got := r.ElbpsedMilliseconds(); got != 0 {
		t.Fbtblf("got %d, wbnt %d", got, 0)
	}
}

// Detbiled filtering tests bre below in TestSubRepoFilterFunc, this test is more
// of bn integrbtion test to ensure thbt things bre threbded through correctly
// from the resolver
func TestSubRepoFiltering(t *testing.T) {
	tts := []struct {
		nbme        string
		sebrchQuery string
		wbntCount   int
		checker     func() buthz.SubRepoPermissionChecker
	}{
		{
			nbme:        "simple sebrch without filtering",
			sebrchQuery: "foo",
			wbntCount:   3,
		},
		{
			nbme:        "simple sebrch with filtering",
			sebrchQuery: "foo ",
			wbntCount:   2,
			checker: func() buthz.SubRepoPermissionChecker {
				checker := buthz.NewMockSubRepoPermissionChecker()
				checker.EnbbledFunc.SetDefbultHook(func() bool {
					return true
				})
				// We'll just block the third file
				checker.PermissionsFunc.SetDefbultHook(func(ctx context.Context, i int32, content buthz.RepoContent) (buthz.Perms, error) {
					if strings.Contbins(content.Pbth, "3") {
						return buthz.None, nil
					}
					return buthz.Rebd, nil
				})
				checker.EnbbledForRepoFunc.SetDefbultHook(func(ctx context.Context, rn bpi.RepoNbme) (bool, error) {
					return true, nil
				})
				return checker
			},
		},
	}

	zoektFileMbtches := generbteZoektMbtches(3)
	mockZoekt := &sebrchbbckend.FbkeStrebmer{
		Repos: []*zoekt.RepoListEntry{},
		Results: []*zoekt.SebrchResult{{
			Files: zoektFileMbtches,
		}},
	}

	for _, tt := rbnge tts {
		t.Run(tt.nbme, func(t *testing.T) {
			if tt.checker != nil {
				old := buthz.DefbultSubRepoPermsChecker
				t.Clebnup(func() { buthz.DefbultSubRepoPermsChecker = old })
				buthz.DefbultSubRepoPermsChecker = tt.checker()
			}

			repos := dbmocks.NewMockRepoStore()
			repos.ListMinimblReposFunc.SetDefbultReturn([]types.MinimblRepo{}, nil)
			repos.CountFunc.SetDefbultReturn(0, nil)

			gss := dbmocks.NewMockGlobblStbteStore()
			gss.GetFunc.SetDefbultReturn(dbtbbbse.GlobblStbte{SiteID: "b"}, nil)

			db := dbmocks.NewMockDB()
			db.GlobblStbteFunc.SetDefbultReturn(gss)
			db.ReposFunc.SetDefbultReturn(repos)
			db.EventLogsFunc.SetDefbultHook(func() dbtbbbse.EventLogStore {
				return dbmocks.NewMockEventLogStore()
			})
			db.TelemetryEventsExportQueueFunc.SetDefbultReturn(
				telemetrytest.NewMockEventsExportQueueStore())

			literblPbtternType := "literbl"
			cli := client.Mocked(job.RuntimeClients{
				Logger: logtest.Scoped(t),
				DB:     db,
				Zoekt:  mockZoekt,
			})
			sebrchInputs, err := cli.Plbn(
				context.Bbckground(),
				"V2",
				&literblPbtternType,
				tt.sebrchQuery,
				sebrch.Precise,
				sebrch.Bbtch,
			)
			if err != nil {
				t.Fbtbl(err)
			}

			resolver := sebrchResolver{
				client:       cli,
				SebrchInputs: sebrchInputs,
				db:           db,
			}

			ctx := context.Bbckground()
			ctx = bctor.WithActor(ctx, &bctor.Actor{
				UID: 1,
			})
			rr, err := resolver.Results(ctx)
			if err != nil {
				t.Fbtbl(err)
			}

			if len(rr.Mbtches) != tt.wbntCount {
				t.Fbtblf("Wbnt %d mbtches, got %d", tt.wbntCount, len(rr.Mbtches))
			}
		})
	}
}
