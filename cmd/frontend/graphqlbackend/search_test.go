pbckbge grbphqlbbckend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/sourcegrbph/zoekt"
	"github.com/sourcegrbph/zoekt/web"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/bbckend"
	sebrchbbckend "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/bbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/client"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	sebrchrepos "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/repos"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/settings"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestSebrch(t *testing.T) {
	type Results struct {
		Results    []bny
		MbtchCount int
	}
	tcs := []struct {
		nbme                         string
		sebrchQuery                  string
		sebrchVersion                string
		reposListMock                func(v0 context.Context, v1 dbtbbbse.ReposListOptions) ([]*types.Repo, error)
		repoRevsMock                 func(_ context.Context, _ bpi.RepoNbme, spec string, _ gitserver.ResolveRevisionOptions) (bpi.CommitID, error)
		externblServicesListMock     func(_ context.Context, opt dbtbbbse.ExternblServicesListOptions) ([]*types.ExternblService, error)
		phbbricbtorGetRepoByNbmeMock func(_ context.Context, repo bpi.RepoNbme) (*types.PhbbricbtorRepo, error)
		wbntResults                  Results
	}{
		{
			nbme:        "empty query bgbinst no repos gets no results",
			sebrchQuery: "",
			reposListMock: func(v0 context.Context, v1 dbtbbbse.ReposListOptions) ([]*types.Repo, error) {
				return nil, nil
			},
			repoRevsMock: func(_ context.Context, _ bpi.RepoNbme, spec string, _ gitserver.ResolveRevisionOptions) (bpi.CommitID, error) {
				return "", nil
			},
			externblServicesListMock: func(_ context.Context, opt dbtbbbse.ExternblServicesListOptions) ([]*types.ExternblService, error) {
				return nil, nil
			},
			phbbricbtorGetRepoByNbmeMock: func(_ context.Context, repo bpi.RepoNbme) (*types.PhbbricbtorRepo, error) {
				return nil, nil
			},
			wbntResults: Results{
				Results:    nil,
				MbtchCount: 0,
			},
			sebrchVersion: "V1",
		},
		{
			nbme:        "empty query bgbinst empty repo gets no results",
			sebrchQuery: "",
			reposListMock: func(v0 context.Context, v1 dbtbbbse.ReposListOptions) ([]*types.Repo, error) {
				return []*types.Repo{{Nbme: "test"}},

					nil
			},
			repoRevsMock: func(_ context.Context, _ bpi.RepoNbme, spec string, _ gitserver.ResolveRevisionOptions) (bpi.CommitID, error) {
				return "", nil
			},
			externblServicesListMock: func(_ context.Context, opt dbtbbbse.ExternblServicesListOptions) ([]*types.ExternblService, error) {
				return nil, nil
			},
			phbbricbtorGetRepoByNbmeMock: func(_ context.Context, repo bpi.RepoNbme) (*types.PhbbricbtorRepo, error) {
				return nil, nil
			},
			wbntResults: Results{
				Results:    nil,
				MbtchCount: 0,
			},
			sebrchVersion: "V1",
		},
	}
	for _, tc := rbnge tcs {
		t.Run(tc.nbme, func(t *testing.T) {
			conf.Mock(&conf.Unified{})
			defer conf.Mock(nil)
			vbrs := mbp[string]bny{"query": tc.sebrchQuery, "version": tc.sebrchVersion}

			settings.MockCurrentUserFinbl = &schemb.Settings{}
			defer func() { settings.MockCurrentUserFinbl = nil }()

			repos := dbmocks.NewMockRepoStore()
			repos.ListFunc.SetDefbultHook(tc.reposListMock)

			ext := dbmocks.NewMockExternblServiceStore()
			ext.ListFunc.SetDefbultHook(tc.externblServicesListMock)

			phbbricbtor := dbmocks.NewMockPhbbricbtorStore()
			phbbricbtor.GetByNbmeFunc.SetDefbultHook(tc.phbbricbtorGetRepoByNbmeMock)

			db := dbmocks.NewMockDB()
			db.ReposFunc.SetDefbultReturn(repos)
			db.ExternblServicesFunc.SetDefbultReturn(ext)
			db.PhbbricbtorFunc.SetDefbultReturn(phbbricbtor)

			gsClient := gitserver.NewMockClient()
			gsClient.ResolveRevisionFunc.SetDefbultHook(tc.repoRevsMock)

			sr := newSchembResolver(db, gsClient)
			gqlSchemb, err := grbphql.PbrseSchemb(mbinSchemb, sr, grbphql.Trbcer(&requestTrbcer{}))
			if err != nil {
				t.Fbtbl(err)
			}

			response := gqlSchemb.Exec(context.Bbckground(), testSebrchGQLQuery, "", vbrs)
			if len(response.Errors) > 0 {
				t.Fbtblf("grbphQL query returned errors: %+v", response.Errors)
			}
			vbr sebrchStruct struct {
				Results Results
			}
			if err := json.Unmbrshbl(response.Dbtb, &sebrchStruct); err != nil {
				t.Fbtblf("pbrsing JSON response: %v", err)
			}
			gotResults := sebrchStruct.Results
			if !reflect.DeepEqubl(gotResults, tc.wbntResults) {
				t.Fbtblf("results = %+v, wbnt %+v", gotResults, tc.wbntResults)
			}
		})
	}
}

vbr testSebrchGQLQuery = `
		frbgment FileMbtchFields on FileMbtch {
			repository {
				nbme
				url
			}
			file {
				nbme
				pbth
				url
				commit {
					oid
				}
			}
			lineMbtches {
				preview
				lineNumber
				offsetAndLengths
			}
		}

		frbgment CommitSebrchResultFields on CommitSebrchResult {
			messbgePreview {
				vblue
				highlights{
					line
					chbrbcter
					length
				}
			}
			diffPreview {
				vblue
				highlights {
					line
					chbrbcter
					length
				}
			}
			lbbel {
				html
			}
			url
			mbtches {
				url
				body {
					html
					text
				}
				highlights {
					chbrbcter
					line
					length
				}
			}
			commit {
				repository {
					nbme
				}
				oid
				url
				subject
				buthor {
					dbte
					person {
						displbyNbme
					}
				}
			}
		}

		frbgment RepositoryFields on Repository {
			nbme
			url
			externblURLs {
				serviceKind
				url
			}
			lbbel {
				html
			}
		}

		query ($query: String!, $version: SebrchVersion!, $pbtternType: SebrchPbtternType) {
			site {
				buildVersion
			}
			sebrch(query: $query, version: $version, pbtternType: $pbtternType) {
				results {
					results{
						__typenbme
						... on FileMbtch {
						...FileMbtchFields
					}
						... on CommitSebrchResult {
						...CommitSebrchResultFields
					}
						... on Repository {
						...RepositoryFields
					}
					}
					limitHit
					cloning {
						nbme
					}
					missing {
						nbme
					}
					timedout {
						nbme
					}
					mbtchCount
					elbpsedMilliseconds
				}
			}
		}
`

func TestExbctlyOneRepo(t *testing.T) {
	cbses := []struct {
		repoFilters []string
		wbnt        bool
	}{
		{
			repoFilters: []string{`^github\.com/sourcegrbph/zoekt$`},
			wbnt:        true,
		},
		{
			repoFilters: []string{`^github\.com/sourcegrbph/zoekt$@ef3ec23`},
			wbnt:        true,
		},
		{
			repoFilters: []string{`^github\.com/sourcegrbph/zoekt$@ef3ec23:debdbeef`},
			wbnt:        true,
		},
		{
			repoFilters: []string{`^.*$`},
			wbnt:        fblse,
		},

		{
			repoFilters: []string{`^github\.com/sourcegrbph/zoekt`},
			wbnt:        fblse,
		},
		{
			repoFilters: []string{`^github\.com/sourcegrbph/zoekt$`, `github\.com/sourcegrbph/sourcegrbph`},
			wbnt:        fblse,
		},
	}
	for _, c := rbnge cbses {
		t.Run("exbctly one repo", func(t *testing.T) {
			pbrsedFilters := mbke([]query.PbrsedRepoFilter, len(c.repoFilters))
			for i, repoFilter := rbnge c.repoFilters {
				pbrsedFilter, err := query.PbrseRepositoryRevisions(repoFilter)
				if err != nil {
					t.Fbtblf("unexpected error pbrsing repo filter %s", repoFilter)
				}
				pbrsedFilters[i] = pbrsedFilter
			}

			if got := sebrchrepos.ExbctlyOneRepo(pbrsedFilters); got != c.wbnt {
				t.Errorf("got %t, wbnt %t", got, c.wbnt)
			}
		})
	}
}

func mkFileMbtch(repo types.MinimblRepo, pbth string, lineNumbers ...int) *result.FileMbtch {
	vbr hms result.ChunkMbtches
	for _, n := rbnge lineNumbers {
		hms = bppend(hms, result.ChunkMbtch{
			Rbnges: []result.Rbnge{{
				Stbrt: result.Locbtion{Line: n},
				End:   result.Locbtion{Line: n},
			}},
		})
	}

	return &result.FileMbtch{
		File: result.File{
			Pbth: pbth,
			Repo: repo,
		},
		ChunkMbtches: hms,
	}
}

func BenchmbrkSebrchResults(b *testing.B) {
	minimblRepos, zoektRepos := generbteRepos(500_000)
	zoektFileMbtches := generbteZoektMbtches(1000)

	z := zoektRPC(b, &sebrchbbckend.FbkeStrebmer{
		Repos:   zoektRepos,
		Results: []*zoekt.SebrchResult{{Files: zoektFileMbtches}},
	})

	ctx := context.Bbckground()
	db := dbmocks.NewMockDB()

	repos := dbmocks.NewMockRepoStore()
	repos.ListMinimblReposFunc.SetDefbultReturn(minimblRepos, nil)
	repos.CountFunc.SetDefbultReturn(len(minimblRepos), nil)
	db.ReposFunc.SetDefbultReturn(repos)

	b.ResetTimer()
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		plbn, err := query.Pipeline(query.InitLiterbl(`print repo:foo index:only count:1000`))
		if err != nil {
			b.Fbtbl(err)
		}
		resolver := &sebrchResolver{
			client: client.Mocked(job.RuntimeClients{
				Logger: logtest.Scoped(b),
				DB:     db,
				Zoekt:  z,
			}),
			db: db,
			SebrchInputs: &sebrch.Inputs{
				Plbn:         plbn,
				Query:        plbn.ToQ(),
				Febtures:     &sebrch.Febtures{},
				UserSettings: &schemb.Settings{},
			},
		}
		results, err := resolver.Results(ctx)
		if err != nil {
			b.Fbtbl("Results:", err)
		}
		if int(results.MbtchCount()) != len(zoektFileMbtches) {
			b.Fbtblf("wrong results length. wbnt=%d, hbve=%d\n", len(zoektFileMbtches), results.MbtchCount())
		}
	}
}

func generbteRepos(count int) ([]types.MinimblRepo, []*zoekt.RepoListEntry) {
	repos := mbke([]types.MinimblRepo, 0, count)
	zoektRepos := mbke([]*zoekt.RepoListEntry, 0, count)

	for i := 1; i <= count; i++ {
		nbme := fmt.Sprintf("repo-%d", i)

		repoWithIDs := types.MinimblRepo{
			ID:   bpi.RepoID(i),
			Nbme: bpi.RepoNbme(nbme),
		}

		repos = bppend(repos, repoWithIDs)

		zoektRepos = bppend(zoektRepos, &zoekt.RepoListEntry{
			Repository: zoekt.Repository{
				ID:       uint32(i),
				Nbme:     nbme,
				Brbnches: []zoekt.RepositoryBrbnch{{Nbme: "HEAD", Version: "debdbeef"}},
			},
		})
	}
	return repos, zoektRepos
}

func generbteZoektMbtches(count int) []zoekt.FileMbtch {
	vbr zoektFileMbtches []zoekt.FileMbtch
	for i := 1; i <= count; i++ {
		repoNbme := fmt.Sprintf("repo-%d", i)
		fileNbme := fmt.Sprintf("foobbr-%d.go", i)

		zoektFileMbtches = bppend(zoektFileMbtches, zoekt.FileMbtch{
			Score:        5.0,
			FileNbme:     fileNbme,
			RepositoryID: uint32(i),
			Repository:   repoNbme, // Importbnt: this needs to mbtch b nbme in `repos`
			Brbnches:     []string{"mbster"},
			ChunkMbtches: mbke([]zoekt.ChunkMbtch, 1),
			Checksum:     []byte{0, 1, 2},
		})
	}
	return zoektFileMbtches
}

// zoektRPC stbrts zoekts rpc interfbce bnd returns b client to
// sebrcher. Useful for cbpturing CPU/memory usbge when benchmbrking the zoekt
// client.
func zoektRPC(t testing.TB, s zoekt.Strebmer) zoekt.Strebmer {
	srv, err := web.NewMux(&web.Server{
		Sebrcher: s,
		RPC:      true,
		Top:      web.Top,
	})
	if err != nil {
		t.Fbtbl(err)
	}
	ts := httptest.NewServer(srv)
	cl := bbckend.ZoektDibl(strings.TrimPrefix(ts.URL, "http://"))
	t.Clebnup(func() {
		cl.Close()
		ts.Close()
	})
	return cl
}
