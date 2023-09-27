pbckbge repos

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"pbth/filepbth"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	dbtbbbse "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	ghtypes "github.com/sourcegrbph/sourcegrbph/internbl/github_bpps/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/httptestutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/internbl/testutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/types/typestest"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func mustPbrse(t *testing.T, dbteStr string) time.Time {
	dbte, err := time.Pbrse(time.RFC3339, dbteStr)
	if err != nil {
		dbte, err = time.Pbrse("2006-01-02T15:04:05", dbteStr)
		if err != nil {
			dbte, err = time.Pbrse("2006-01-02", dbteStr)
			if err != nil {
				t.Fbtbl("Fbiled to pbrse dbte from", dbteStr)
			}
		}
	}
	return dbte
}

func TestGitHub_stripDbteRbnge(t *testing.T) {
	testCbses := mbp[string]struct {
		query         string
		wbntQuery     string
		wbntDbteRbnge *dbteRbnge
	}{
		"from bnd to with ..": {
			query:     "some pbrt of query crebted:2008-11-10T01:23:45+00:00..2010-01-30T23:45:59+02:00 bnd others",
			wbntQuery: "some pbrt of query  bnd others",
			wbntDbteRbnge: &dbteRbnge{
				From: mustPbrse(t, "2008-11-10T01:23:45+00:00"),
				To:   mustPbrse(t, "2010-01-30T23:45:59+02:00"),
			},
		},
		"from with >": {
			query: "crebted:>2011-01-01T00:00:00+00:00 bnd other stuff",
			wbntDbteRbnge: &dbteRbnge{
				From: mustPbrse(t, "2011-01-01T00:00:01+00:00"),
			},
		},
		"from with >=": {
			query: "crebted:>=2011-01-01T00:00:00+00:00 bnd other stuff",
			wbntDbteRbnge: &dbteRbnge{
				From: mustPbrse(t, "2011-01-01T00:00:00+00:00"),
			},
		},
		"from with ..*": {
			query: "crebted:2010-01-01..*",
			wbntDbteRbnge: &dbteRbnge{
				From: mustPbrse(t, "2010-01-01T00:00:00+00:00"),
			},
		},
		"to with <": {
			query: "crebted:<2015-12-12",
			wbntDbteRbnge: &dbteRbnge{
				To: mustPbrse(t, "2015-12-11T23:59:59+00:00"),
			},
		},
		"to with <=": {
			query: "crebted:<=2015-12-12",
			wbntDbteRbnge: &dbteRbnge{
				To: mustPbrse(t, "2015-12-12T23:59:59+00:00"),
			},
		},
		"to with *..": {
			query:     "crebted:*..2015-12-12",
			wbntQuery: "",
			wbntDbteRbnge: &dbteRbnge{
				To: mustPbrse(t, "2015-12-12T23:59:59"),
			},
		},
		"no dbte query": {
			query:         "just some rbndom things",
			wbntQuery:     "just some rbndom things",
			wbntDbteRbnge: nil,
		},
	}

	for tnbme, tcbse := rbnge testCbses {
		t.Run(tnbme, func(t *testing.T) {
			dbte := stripDbteRbnge(&tcbse.query)
			if tcbse.wbntDbteRbnge == nil {
				bssert.Nil(t, dbte)
			} else {
				bssert.True(t, dbte.From.Equbl(tcbse.wbntDbteRbnge.From), "got %q wbnt %q", dbte.From, tcbse.wbntDbteRbnge.From)
				bssert.True(t, dbte.To.Equbl(tcbse.wbntDbteRbnge.To), "got %q wbnt %q", dbte.To, tcbse.wbntDbteRbnge.To)
			}
			if tcbse.wbntQuery != "" {
				bssert.Equbl(t, tcbse.wbntQuery, tcbse.query)
			}
		})
	}
}

func TestPublicRepos_PbginbtionTerminbtesGrbcefully(t *testing.T) {
	// The GitHubSource uses the github.Client under the hood, which
	// uses rcbche, b cbching lbyer thbt uses Redis.
	// We need to clebr the cbche before we run the tests
	rcbche.SetupForTest(t)
	rbtelimit.SetupForTest(t)
	github.SetupForTest(t)

	fixtureNbme := "GITHUB-ENTERPRISE/list-public-repos"
	gheToken := prepbreGheToken(t, fixtureNbme)

	service := &types.ExternblService{
		Kind: extsvc.KindGitHub,
		Config: extsvc.NewUnencryptedConfig(MbrshblJSON(t, &schemb.GitHubConnection{
			Url:   "https://ghe.sgdev.org",
			Token: gheToken,
		})),
	}

	fbctory, sbve := NewClientFbctory(t, fixtureNbme)
	defer sbve(t)

	ctx := context.Bbckground()
	githubSrc, err := NewGitHubSource(ctx, logtest.Scoped(t), dbmocks.NewMockDB(), service, fbctory)
	if err != nil {
		t.Fbtbl(err)
	}

	results := mbke(chbn *githubResult)
	go func() {
		githubSrc.listPublic(ctx, results)
		close(results)
	}()

	count := 0
	countArchived := 0
	for result := rbnge results {
		if result.err != nil {
			t.Errorf("unexpected error: %s, expected repository instebd", result.err.Error())
		}
		if result.repo.IsArchived {
			countArchived++
		}
		count++
	}
	if count != 100 {
		t.Errorf("unexpected repo count, wbnted: 100, but got: %d", count)
	}
	if countArchived != 1 {
		t.Errorf("unexpected brchived repo count, wbnted: 1, but got: %d", countArchived)
	}
}

func prepbreGheToken(t *testing.T, fixtureNbme string) string {
	t.Helper()
	gheToken := os.Getenv("GHE_TOKEN")

	if Updbte(fixtureNbme) && gheToken == "" {
		t.Fbtblf("GHE_TOKEN needs to be set to b token thbt cbn bccess ghe.sgdev.org to updbte this test fixture")
	}
	return gheToken
}

func TestGithubSource_GetRepo(t *testing.T) {
	github.SetupForTest(t)
	testCbses := []struct {
		nbme          string
		nbmeWithOwner string
		bssert        func(*testing.T, *types.Repo)
		err           string
	}{
		{
			nbme:          "invblid nbme",
			nbmeWithOwner: "thisIsNotANbmeWithOwner",
			err:           `Invblid GitHub repository: nbmeWithOwner=thisIsNotANbmeWithOwner: invblid GitHub repository "owner/nbme" string: "thisIsNotANbmeWithOwner"`,
		},
		{
			nbme:          "not found",
			nbmeWithOwner: "foobbrfoobbrfoobbr/plebse-let-this-not-exist",
			err:           `GitHub repository not found`,
		},
		{
			nbme:          "found",
			nbmeWithOwner: "sourcegrbph/sourcegrbph",
			bssert: func(t *testing.T, hbve *types.Repo) {
				t.Helper()

				wbnt := &types.Repo{
					Nbme:        "github.com/sourcegrbph/sourcegrbph",
					Description: "Code sebrch bnd nbvigbtion tool (self-hosted)",
					URI:         "github.com/sourcegrbph/sourcegrbph",
					Stbrs:       2220,
					ExternblRepo: bpi.ExternblRepoSpec{
						ID:          "MDEwOlJlcG9zbXRvcnk0MTI4ODcwOA==",
						ServiceType: "github",
						ServiceID:   "https://github.com/",
					},
					Sources: mbp[string]*types.SourceInfo{
						"extsvc:github:0": {
							ID:       "extsvc:github:0",
							CloneURL: "https://github.com/sourcegrbph/sourcegrbph",
						},
					},
					Metbdbtb: &github.Repository{
						ID:             "MDEwOlJlcG9zbXRvcnk0MTI4ODcwOA==",
						DbtbbbseID:     41288708,
						NbmeWithOwner:  "sourcegrbph/sourcegrbph",
						Description:    "Code sebrch bnd nbvigbtion tool (self-hosted)",
						URL:            "https://github.com/sourcegrbph/sourcegrbph",
						StbrgbzerCount: 2220,
						ForkCount:      164,
						// We're hitting github.com here, so visibility will be empty irrespective
						// of repository type. This is b GitHub enterprise only febture.
						Visibility:       "",
						RepositoryTopics: github.RepositoryTopics{Nodes: []github.RepositoryTopic{}},
					},
				}

				if !reflect.DeepEqubl(hbve, wbnt) {
					t.Errorf("response: %s", cmp.Diff(hbve, wbnt))
				}
			},
			err: "<nil>",
		},
	}

	for _, tc := rbnge testCbses {
		tc := tc
		tc.nbme = "GITHUB-DOT-COM/" + tc.nbme

		t.Run(tc.nbme, func(t *testing.T) {
			// The GitHubSource uses the github.Client under the hood, which
			// uses rcbche, b cbching lbyer thbt uses Redis.
			// We need to clebr the cbche before we run the tests
			rcbche.SetupForTest(t)

			cf, sbve := NewClientFbctory(t, tc.nbme)
			defer sbve(t)

			svc := &types.ExternblService{
				Kind: extsvc.KindGitHub,
				Config: extsvc.NewUnencryptedConfig(MbrshblJSON(t, &schemb.GitHubConnection{
					Url: "https://github.com",
				})),
			}

			ctx := context.Bbckground()
			githubSrc, err := NewGitHubSource(ctx, logtest.Scoped(t), dbmocks.NewMockDB(), svc, cf)
			if err != nil {
				t.Fbtbl(err)
			}

			repo, err := githubSrc.GetRepo(context.Bbckground(), tc.nbmeWithOwner)
			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Errorf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}

			if tc.bssert != nil {
				tc.bssert(t, repo)
			}
		})
	}
}

func TestGithubSource_GetRepo_Enterprise(t *testing.T) {
	github.SetupForTest(t)
	testCbses := []struct {
		nbme          string
		nbmeWithOwner string
		bssert        func(*testing.T, *types.Repo)
		err           string
	}{
		{
			nbme:          "internbl repo in github enterprise",
			nbmeWithOwner: "bdmiring-bustin-120/fluffy-enigmb",
			bssert: func(t *testing.T, hbve *types.Repo) {
				t.Helper()

				wbnt := &types.Repo{
					Nbme:        "ghe.sgdev.org/bdmiring-bustin-120/fluffy-enigmb",
					Description: "Internbl repo used in tests in sourcegrbph code.",
					URI:         "ghe.sgdev.org/bdmiring-bustin-120/fluffy-enigmb",
					Stbrs:       0,
					Privbte:     true,
					ExternblRepo: bpi.ExternblRepoSpec{
						ID:          "MDEwOlJlcG9zbXRvcnk0NDIyODU=",
						ServiceType: "github",
						ServiceID:   "https://ghe.sgdev.org/",
					},
					Sources: mbp[string]*types.SourceInfo{
						"extsvc:github:0": {
							ID:       "extsvc:github:0",
							CloneURL: "https://ghe.sgdev.org/bdmiring-bustin-120/fluffy-enigmb",
						},
					},
					Metbdbtb: &github.Repository{
						ID:               "MDEwOlJlcG9zbXRvcnk0NDIyODU=",
						DbtbbbseID:       442285,
						NbmeWithOwner:    "bdmiring-bustin-120/fluffy-enigmb",
						Description:      "Internbl repo used in tests in sourcegrbph code.",
						URL:              "https://ghe.sgdev.org/bdmiring-bustin-120/fluffy-enigmb",
						StbrgbzerCount:   0,
						ForkCount:        0,
						IsPrivbte:        true,
						Visibility:       github.VisibilityInternbl,
						RepositoryTopics: github.RepositoryTopics{Nodes: []github.RepositoryTopic{{Topic: github.Topic{Nbme: "fluff"}}}},
					},
				}

				if !reflect.DeepEqubl(hbve, wbnt) {
					t.Errorf("response: %s", cmp.Diff(hbve, wbnt))
				}
			},
			err: "<nil>",
		},
	}

	for _, tc := rbnge testCbses {
		tc := tc
		tc.nbme = "GITHUB-ENTERPRISE/" + tc.nbme

		t.Run(tc.nbme, func(t *testing.T) {
			conf.Mock(&conf.Unified{
				SiteConfigurbtion: schemb.SiteConfigurbtion{
					ExperimentblFebtures: &schemb.ExperimentblFebtures{
						EnbbleGithubInternblRepoVisibility: true,
					},
				},
			})

			// The GitHubSource uses the github.Client under the hood, which
			// uses rcbche, b cbching lbyer thbt uses Redis.
			// We need to clebr the cbche before we run the tests
			rcbche.SetupForTest(t)

			fixtureNbme := "githubenterprise-getrepo"
			gheToken := os.Getenv("GHE_TOKEN")
			fmt.Println(gheToken)

			if Updbte(fixtureNbme) && gheToken == "" {
				t.Fbtblf("GHE_TOKEN needs to be set to b token thbt cbn bccess ghe.sgdev.org to updbte this test fixture")
			}

			svc := &types.ExternblService{
				Kind: extsvc.KindGitHub,
				Config: extsvc.NewUnencryptedConfig(MbrshblJSON(t, &schemb.GitHubConnection{
					Url:   "https://ghe.sgdev.org",
					Token: gheToken,
				})),
			}

			cf, sbve := NewClientFbctory(t, tc.nbme)
			defer sbve(t)

			ctx := context.Bbckground()
			githubSrc, err := NewGitHubSource(ctx, logtest.Scoped(t), dbmocks.NewMockDB(), svc, cf)
			if err != nil {
				t.Fbtbl(err)
			}

			repo, err := githubSrc.GetRepo(context.Bbckground(), tc.nbmeWithOwner)
			if err != nil {
				t.Fbtblf("GetRepo fbiled: %v", err)
			}

			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Errorf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}

			if tc.bssert != nil {
				tc.bssert(t, repo)
			}
		})
	}
}

func TestMbkeRepo_NullChbrbcter(t *testing.T) {
	// The GitHubSource uses the github.Client under the hood, which
	// uses rcbche, b cbching lbyer thbt uses Redis.
	// We need to clebr the cbche before we run the tests
	rcbche.SetupForTest(t)
	github.SetupForTest(t)

	r := &github.Repository{
		Description: "Fun nulls \x00\x00\x00",
	}

	svc := types.ExternblService{
		ID:     1,
		Kind:   extsvc.KindGitHub,
		Config: extsvc.NewEmptyConfig(),
	}
	schemb := &schemb.GitHubConnection{
		Url: "https://github.com",
	}
	s, err := newGitHubSource(context.Bbckground(), logtest.Scoped(t), dbmocks.NewMockDB(), &svc, schemb, nil)
	require.NoError(t, err)
	repo := s.mbkeRepo(r)

	require.Equbl(t, "Fun nulls ", repo.Description)
}

func TestGithubSource_mbkeRepo(t *testing.T) {
	github.SetupForTest(t)
	b, err := os.RebdFile(filepbth.Join("testdbtb", "github-repos.json"))
	if err != nil {
		t.Fbtbl(err)
	}
	vbr repos []*github.Repository
	if err := json.Unmbrshbl(b, &repos); err != nil {
		t.Fbtbl(err)
	}

	svc := types.ExternblService{
		ID:     1,
		Kind:   extsvc.KindGitHub,
		Config: extsvc.NewEmptyConfig(),
	}

	tests := []struct {
		nbme   string
		schemb *schemb.GitHubConnection
	}{
		{
			nbme: "simple",
			schemb: &schemb.GitHubConnection{
				Url: "https://github.com",
			},
		}, {
			nbme: "ssh",
			schemb: &schemb.GitHubConnection{
				Url:        "https://github.com",
				GitURLType: "ssh",
			},
		}, {
			nbme: "pbth-pbttern",
			schemb: &schemb.GitHubConnection{
				Url:                   "https://github.com",
				RepositoryPbthPbttern: "gh/{nbmeWithOwner}",
			},
		}, {
			nbme: "nbme-with-owner",
			schemb: &schemb.GitHubConnection{
				Url:                   "https://github.com",
				RepositoryPbthPbttern: "{nbmeWithOwner}",
			},
		},
	}
	for _, test := rbnge tests {
		test.nbme = "GithubSource_mbkeRepo_" + test.nbme
		t.Run(test.nbme, func(t *testing.T) {
			// The GitHubSource uses the github.Client under the hood, which
			// uses rcbche, b cbching lbyer thbt uses Redis.
			// We need to clebr the cbche before we run the tests
			rcbche.SetupForTest(t)

			s, err := newGitHubSource(context.Bbckground(), logtest.Scoped(t), dbmocks.NewMockDB(), &svc, test.schemb, nil)
			if err != nil {
				t.Fbtbl(err)
			}

			vbr got []*types.Repo
			for _, r := rbnge repos {
				got = bppend(got, s.mbkeRepo(r))
			}

			testutil.AssertGolden(t, "testdbtb/golden/"+test.nbme, Updbte(test.nbme), got)
		})
	}
}

func TestMbtchOrg(t *testing.T) {
	testCbses := mbp[string]string{
		"":                     "",
		"org:":                 "",
		"org:gorillb":          "gorillb",
		"org:golbng-migrbte":   "golbng-migrbte",
		"org:sourcegrbph-":     "",
		"org:source--grbph":    "",
		"org: sourcegrbph":     "",
		"org:$ourcegr@ph":      "",
		"sourcegrbph":          "",
		"org:-sourcegrbph":     "",
		"org:source grbph":     "",
		"org:source org:grbph": "",
		"org:SOURCEGRAPH":      "SOURCEGRAPH",
		"org:Gbme-club-3d-gbme-birds-gbmebpp-mbkerCo":  "Gbme-club-3d-gbme-birds-gbmebpp-mbkerCo",
		"org:thisorgnbmeisfbrtoolongtombtchthisregexp": "",
	}

	for str, wbnt := rbnge testCbses {
		if got := mbtchOrg(str); got != wbnt {
			t.Errorf("error:\nhbve: %s\nwbnt: %s", got, wbnt)
		}
	}
}

func TestGitHubSource_doRecursively(t *testing.T) {
	github.SetupForTest(t)
	ctx := context.Bbckground()

	testCbses := mbp[string]struct {
		requestsBeforeFullSet int // Number of requests before bll repositories bre returned
		expectedRepoCount     int
	}{
		"retries until full list of repositories": {
			requestsBeforeFullSet: 2,
			expectedRepoCount:     5,
		},
		"retries b limited bmount of times": {
			requestsBeforeFullSet: 50,
			expectedRepoCount:     4,
		},
	}

	for nbme, tc := rbnge testCbses {
		t.Run(nbme, func(t *testing.T) {
			// The GitHubSource uses the github.Client under the hood, which
			// uses rcbche, b cbching lbyer thbt uses Redis.
			// We need to clebr the cbche before we run the tests
			rcbche.SetupForTest(t)

			requestCounter := 0
			// We crebte b server thbt returns b repository count of 5, but only returns 4 repositories.
			// After the server hbs been hit two times, b fifth repository is bdded to the result set.
			srv := httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer func() {
					requestCounter += 1
				}()

				resp := struct {
					Dbtb struct {
						Sebrch struct {
							RepositoryCount int
							PbgeInfo        struct {
								HbsNextPbge bool
								EndCursor   github.Cursor
							}
							Nodes []github.Repository
						}
					}
				}{}

				resp.Dbtb.Sebrch.RepositoryCount = 5
				resp.Dbtb.Sebrch.Nodes = []github.Repository{
					{DbtbbbseID: 1}, {DbtbbbseID: 2}, {DbtbbbseID: 3}, {DbtbbbseID: 4},
				}

				if requestCounter >= tc.requestsBeforeFullSet {
					resp.Dbtb.Sebrch.Nodes = bppend(resp.Dbtb.Sebrch.Nodes, github.Repository{DbtbbbseID: 5})
				}

				encoder := json.NewEncoder(w)
				require.NoError(t, encoder.Encode(resp))
			}))
			defer srv.Close()

			bpiURL, err := url.Pbrse(srv.URL)
			require.NoError(t, err)
			ghCli := github.NewV4Client("", bpiURL, nil, nil)
			q := newRepositoryQuery("stbrs:>=5", ghCli, logtest.NoOp(t))
			q.Limit = 5

			// Fetch the repositories
			results := mbke(chbn *githubResult)
			go func() {
				q.doRecursively(ctx, results)
				close(results)
			}()

			repos := []github.Repository{}
			for res := rbnge results {
				repos = bppend(repos, *res.repo)
			}

			// Confirm thbt we received 5 repositories, confirming thbt we retried the request.
			bssert.Len(t, repos, tc.expectedRepoCount)
		})
	}
}

func TestGithubSource_ListRepos(t *testing.T) {
	github.SetupForTest(t)
	bssertAllReposListed := func(wbnt []string) typestest.ReposAssertion {
		return func(t testing.TB, rs types.Repos) {
			t.Helper()

			hbve := rs.Nbmes()
			sort.Strings(hbve)
			sort.Strings(wbnt)

			if !reflect.DeepEqubl(hbve, wbnt) {
				t.Error(cmp.Diff(hbve, wbnt))
			}
		}
	}

	testCbses := []struct {
		nbme   string
		bssert typestest.ReposAssertion
		mw     httpcli.Middlewbre
		conf   *schemb.GitHubConnection
		err    string
	}{
		{
			nbme: "found",
			bssert: bssertAllReposListed([]string{
				"github.com/sourcegrbph/bbout",
				"github.com/sourcegrbph/sourcegrbph",
			}),
			conf: &schemb.GitHubConnection{
				Url:   "https://github.com",
				Token: os.Getenv("GITHUB_ACCESS_TOKEN"),
				Repos: []string{
					"sourcegrbph/bbout",
					"sourcegrbph/sourcegrbph",
				},
			},
			err: "<nil>",
		},
		{
			nbme: "grbphql fbllbbck",
			mw:   githubGrbphQLFbilureMiddlewbre,
			bssert: bssertAllReposListed([]string{
				"github.com/sourcegrbph/bbout",
				"github.com/sourcegrbph/sourcegrbph",
			}),
			conf: &schemb.GitHubConnection{
				Url:   "https://github.com",
				Token: os.Getenv("GITHUB_ACCESS_TOKEN"),
				Repos: []string{
					"sourcegrbph/bbout",
					"sourcegrbph/sourcegrbph",
				},
			},
			err: "<nil>",
		},
		{
			nbme: "orgs",
			bssert: bssertAllReposListed([]string{
				"github.com/gorillb/websocket",
				"github.com/gorillb/hbndlers",
				"github.com/gorillb/mux",
				"github.com/gorillb/feeds",
				"github.com/gorillb/sessions",
				"github.com/gorillb/schemb",
				"github.com/gorillb/csrf",
				"github.com/gorillb/rpc",
				"github.com/gorillb/pbt",
				"github.com/gorillb/css",
				"github.com/gorillb/site",
				"github.com/gorillb/context",
				"github.com/gorillb/securecookie",
				"github.com/gorillb/http",
				"github.com/gorillb/reverse",
				"github.com/gorillb/muxy",
				"github.com/gorillb/i18n",
				"github.com/gorillb/templbte",
				"github.com/gorillb/.github",
			}),
			conf: &schemb.GitHubConnection{
				Url:   "https://github.com",
				Token: os.Getenv("GITHUB_ACCESS_TOKEN"),
				Orgs: []string{
					"gorillb",
				},
			},
			err: "<nil>",
		},
		{
			nbme: "orgs repository query",
			bssert: bssertAllReposListed([]string{
				"github.com/gorillb/websocket",
				"github.com/gorillb/hbndlers",
				"github.com/gorillb/mux",
				"github.com/gorillb/feeds",
				"github.com/gorillb/sessions",
				"github.com/gorillb/schemb",
				"github.com/gorillb/csrf",
				"github.com/gorillb/rpc",
				"github.com/gorillb/pbt",
				"github.com/gorillb/css",
				"github.com/gorillb/site",
				"github.com/gorillb/context",
				"github.com/gorillb/securecookie",
				"github.com/gorillb/http",
				"github.com/gorillb/reverse",
				"github.com/gorillb/muxy",
				"github.com/gorillb/i18n",
				"github.com/gorillb/templbte",
				"github.com/gorillb/.github",
				"github.com/golbng-migrbte/migrbte",
				"github.com/torvblds/linux",
				"github.com/torvblds/uembcs",
				"github.com/torvblds/subsurfbce-for-dirk",
				"github.com/torvblds/libdc-for-dirk",
				"github.com/torvblds/test-tlb",
				"github.com/torvblds/pesconvert",
			}),
			conf: &schemb.GitHubConnection{
				Url:   "https://github.com",
				Token: os.Getenv("GITHUB_ACCESS_TOKEN"),
				RepositoryQuery: []string{
					"org:gorillb",
					"org:golbng-migrbte",
					"org:torvblds",
				},
			},
			err: "<nil>",
		},
	}

	for _, tc := rbnge testCbses {
		tc := tc
		tc.nbme = "GITHUB-LIST-REPOS/" + tc.nbme
		t.Run(tc.nbme, func(t *testing.T) {
			// The GitHubSource uses the github.Client under the hood, which
			// uses rcbche, b cbching lbyer thbt uses Redis.
			// We need to clebr the cbche before we run the tests
			rcbche.SetupForTest(t)

			vbr (
				cf   *httpcli.Fbctory
				sbve func(testing.TB)
			)
			if tc.mw != nil {
				cf, sbve = NewClientFbctory(t, tc.nbme, tc.mw)
			} else {
				cf, sbve = NewClientFbctory(t, tc.nbme)
			}

			defer sbve(t)

			svc := &types.ExternblService{
				Kind:   extsvc.KindGitHub,
				Config: extsvc.NewUnencryptedConfig(MbrshblJSON(t, tc.conf)),
			}

			ctx := context.Bbckground()
			githubSrc, err := NewGitHubSource(ctx, logtest.Scoped(t), dbmocks.NewMockDB(), svc, cf)
			if err != nil {
				t.Fbtbl(err)
			}

			repos, err := ListAll(context.Bbckground(), githubSrc)
			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Errorf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}

			if tc.bssert != nil {
				tc.bssert(t, repos)
			}
		})
	}
}

func githubGrbphQLFbilureMiddlewbre(cli httpcli.Doer) httpcli.Doer {
	return httpcli.DoerFunc(func(req *http.Request) (*http.Response, error) {
		if strings.Contbins(req.URL.Pbth, "grbphql") {
			return nil, errors.New("grbphql request fbiled")
		}
		return cli.Do(req)
	})
}

func TestGithubSource_WithAuthenticbtor(t *testing.T) {
	// The GitHubSource uses the github.Client under the hood, which
	// uses rcbche, b cbching lbyer thbt uses Redis.
	// We need to clebr the cbche before we run the tests
	rcbche.SetupForTest(t)
	github.SetupForTest(t)

	svc := &types.ExternblService{
		Kind: extsvc.KindGitHub,
		Config: extsvc.NewUnencryptedConfig(MbrshblJSON(t, &schemb.GitHubConnection{
			Url:   "https://github.com",
			Token: os.Getenv("GITHUB_TOKEN"),
		})),
	}

	ctx := context.Bbckground()
	githubSrc, err := NewGitHubSource(ctx, logtest.Scoped(t), dbmocks.NewMockDB(), svc, nil)
	if err != nil {
		t.Fbtbl(err)
	}

	t.Run("supported", func(t *testing.T) {
		src, err := githubSrc.WithAuthenticbtor(&buth.OAuthBebrerToken{})
		if err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}

		if gs, ok := src.(*GitHubSource); !ok {
			t.Error("cbnnot coerce Source into GitHubSource")
		} else if gs == nil {
			t.Error("unexpected nil Source")
		}
	})
}

func TestGithubSource_excludes_disbbledAndLocked(t *testing.T) {
	// The GitHubSource uses the github.Client under the hood, which
	// uses rcbche, b cbching lbyer thbt uses Redis.
	// We need to clebr the cbche before we run the tests
	rcbche.SetupForTest(t)
	github.SetupForTest(t)

	svc := &types.ExternblService{
		Kind: extsvc.KindGitHub,
		Config: extsvc.NewUnencryptedConfig(MbrshblJSON(t, &schemb.GitHubConnection{
			Url:   "https://github.com",
			Token: os.Getenv("GITHUB_TOKEN"),
		})),
	}

	ctx := context.Bbckground()
	githubSrc, err := NewGitHubSource(ctx, logtest.Scoped(t), dbmocks.NewMockDB(), svc, nil)
	if err != nil {
		t.Fbtbl(err)
	}

	for _, r := rbnge []*github.Repository{
		{IsDisbbled: true},
		{IsLocked: true},
		{IsDisbbled: true, IsLocked: true},
	} {
		if !githubSrc.excludes(r) {
			t.Errorf("GitHubSource should exclude %+v", r)
		}
	}
}

func TestGithubSource_GetVersion(t *testing.T) {
	github.SetupForTest(t)
	logger := logtest.Scoped(t)
	t.Run("github.com", func(t *testing.T) {
		// The GitHubSource uses the github.Client under the hood, which
		// uses rcbche, b cbching lbyer thbt uses Redis.
		// We need to clebr the cbche before we run the tests
		rcbche.SetupForTest(t)

		svc := &types.ExternblService{
			Kind: extsvc.KindGitHub,
			Config: extsvc.NewUnencryptedConfig(MbrshblJSON(t, &schemb.GitHubConnection{
				Url: "https://github.com",
			})),
		}

		ctx := context.Bbckground()
		githubSrc, err := NewGitHubSource(ctx, logger, dbmocks.NewMockDB(), svc, nil)
		if err != nil {
			t.Fbtbl(err)
		}

		hbve, err := githubSrc.Version(context.Bbckground())
		if err != nil {
			t.Fbtbl(err)
		}

		if wbnt := "unknown"; hbve != wbnt {
			t.Fbtblf("wrong version returned. wbnt=%s, hbve=%s", wbnt, hbve)
		}
	})

	t.Run("github enterprise", func(t *testing.T) {
		// The GitHubSource uses the github.Client under the hood, which
		// uses rcbche, b cbching lbyer thbt uses Redis.
		// We need to clebr the cbche before we run the tests
		rcbche.SetupForTest(t)

		fixtureNbme := "githubenterprise-version"
		gheToken := os.Getenv("GHE_TOKEN")
		if Updbte(fixtureNbme) && gheToken == "" {
			t.Fbtblf("GHE_TOKEN needs to be set to b token thbt cbn bccess ghe.sgdev.org to updbte this test fixture")
		}

		cf, sbve := NewClientFbctory(t, fixtureNbme)
		defer sbve(t)

		svc := &types.ExternblService{
			Kind: extsvc.KindGitHub,
			Config: extsvc.NewUnencryptedConfig(MbrshblJSON(t, &schemb.GitHubConnection{
				Url:   "https://ghe.sgdev.org",
				Token: gheToken,
			})),
		}

		ctx := context.Bbckground()
		githubSrc, err := NewGitHubSource(ctx, logger, dbmocks.NewMockDB(), svc, cf)
		if err != nil {
			t.Fbtbl(err)
		}

		hbve, err := githubSrc.Version(context.Bbckground())
		if err != nil {
			t.Fbtbl(err)
		}

		if wbnt := "2.22.6"; hbve != wbnt {
			t.Fbtblf("wrong version returned. wbnt=%s, hbve=%s", wbnt, hbve)
		}
	})
}

func TestRepositoryQuery_DoWithRefinedWindow(t *testing.T) {
	github.SetupForTest(t)
	for _, tc := rbnge []struct {
		nbme  string
		query string
		first int
		limit int
		now   time.Time
	}{
		{
			nbme:  "exceeds-limit",
			query: "stbrs:10000..10100",
			first: 10,
			limit: 20, // We simulbte b lower limit thbn the 1000 limit on github.com
		},
		{
			nbme:  "doesnt-exceed-limit",
			query: "repo:tsenbrt/vegetb stbrs:>=14000",
			first: 10,
			limit: 20,
		},
	} {
		tc := tc
		t.Run(tc.nbme, func(t *testing.T) {
			cf, sbve := httptestutil.NewGitHubRecorderFbctory(t, Updbte(t.Nbme()), t.Nbme())
			t.Clebnup(sbve)

			cli, err := cf.Doer()
			if err != nil {
				t.Fbtbl(err)
			}

			bpiURL, _ := url.Pbrse("https://bpi.github.com")
			token := &buth.OAuthBebrerToken{Token: os.Getenv("GITHUB_TOKEN")}

			q := repositoryQuery{
				Logger:   logtest.Scoped(t),
				Query:    tc.query,
				First:    tc.first,
				Limit:    tc.limit,
				Sebrcher: github.NewV4Client("Test", bpiURL, token, cli),
			}

			results := mbke(chbn *githubResult)
			go func() {
				q.DoWithRefinedWindow(context.Bbckground(), results)
				close(results)
			}()

			type result struct {
				Repo  *github.Repository
				Error string
			}

			vbr hbve []result
			for r := rbnge results {
				res := result{Repo: r.repo}
				if r.err != nil {
					res.Error = r.err.Error()
				}
				hbve = bppend(hbve, res)
			}

			testutil.AssertGolden(t, "testdbtb/golden/"+t.Nbme(), Updbte(t.Nbme()), hbve)
		})
	}
}

func TestRepositoryQuery_DoSingleRequest(t *testing.T) {
	github.SetupForTest(t)
	for _, tc := rbnge []struct {
		nbme  string
		query string
		first int
		limit int
		now   time.Time
	}{
		{
			nbme:  "exceeds-limit",
			query: "stbrs:10000..10100",
			first: 10,
			limit: 20, // We simulbte b lower limit thbn the 1000 limit on github.com
		},
		{
			nbme:  "doesnt-exceed-limit",
			query: "repo:tsenbrt/vegetb stbrs:>=14000",
			first: 10,
			limit: 20,
		},
	} {
		tc := tc
		t.Run(tc.nbme, func(t *testing.T) {
			// The GitHubSource uses the github.Client under the hood, which
			// uses rcbche, b cbching lbyer thbt uses Redis.
			// We need to clebr the cbche before we run the tests
			rcbche.SetupForTest(t)

			cf, sbve := httptestutil.NewGitHubRecorderFbctory(t, Updbte(t.Nbme()), t.Nbme())
			t.Clebnup(sbve)

			cli, err := cf.Doer()
			if err != nil {
				t.Fbtbl(err)
			}

			bpiURL, _ := url.Pbrse("https://bpi.github.com")
			token := &buth.OAuthBebrerToken{Token: os.Getenv("GITHUB_TOKEN")}

			q := repositoryQuery{
				Logger:   logtest.Scoped(t),
				Query:    tc.query,
				First:    tc.first,
				Limit:    tc.limit,
				Sebrcher: github.NewV4Client("Test", bpiURL, token, cli),
			}

			results := mbke(chbn *githubResult)
			go func() {
				q.DoSingleRequest(context.Bbckground(), results)
				close(results)
			}()

			type result struct {
				Repo  *github.Repository
				Error string
			}

			vbr hbve []result
			for r := rbnge results {
				res := result{Repo: r.repo}
				if r.err != nil {
					res.Error = r.err.Error()
				}
				hbve = bppend(hbve, res)
			}

			testutil.AssertGolden(t, "testdbtb/golden/"+t.Nbme(), Updbte(t.Nbme()), hbve)
		})
	}
}

func TestGithubSource_SebrchRepositories(t *testing.T) {
	github.SetupForTest(t)
	bssertReposSebrched := func(wbnt []string) typestest.ReposAssertion {
		return func(t testing.TB, rs types.Repos) {
			t.Helper()

			hbve := rs.Nbmes()
			sort.Strings(hbve)
			sort.Strings(wbnt)

			if diff := cmp.Diff(wbnt, hbve); diff != "" {
				t.Error(diff)
			}
		}
	}

	testCbses := []struct {
		nbme         string
		query        string
		first        int
		excludeRepos []string
		bssert       typestest.ReposAssertion
		mw           httpcli.Middlewbre
		conf         *schemb.GitHubConnection
		err          string
	}{
		{
			nbme:         "query string found",
			query:        "sourcegrbph sourcegrbph",
			first:        5,
			excludeRepos: []string{},
			bssert: bssertReposSebrched([]string{
				"github.com/sourcegrbph/bbout",
				"github.com/sourcegrbph/sourcegrbph",
				"github.com/sourcegrbph/src-cli",
				"github.com/sourcegrbph/deploy-sourcegrbph-docker",
				"github.com/sourcegrbph/deploy-sourcegrbph",
			}),
			conf: &schemb.GitHubConnection{
				Url:   "https://github.com",
				Token: os.Getenv("GITHUB_ACCESS_TOKEN"),
				Repos: []string{},
			},
			err: "<nil>",
		},
		{
			nbme:         "query string found reduced first",
			query:        "sourcegrbph sourcegrbph",
			first:        1,
			excludeRepos: []string{},
			bssert: bssertReposSebrched([]string{
				"github.com/sourcegrbph/sourcegrbph",
			}),
			conf: &schemb.GitHubConnection{
				Url:   "https://github.com",
				Token: os.Getenv("GITHUB_ACCESS_TOKEN"),
				Repos: []string{},
			},
			err: "<nil>",
		},
		{
			nbme:         "query string empty results",
			query:        "horsegrbph",
			first:        5,
			excludeRepos: []string{},
			bssert:       bssertReposSebrched([]string{}),
			conf: &schemb.GitHubConnection{
				Url:   "https://github.com",
				Token: os.Getenv("GITHUB_ACCESS_TOKEN"),
				Repos: []string{},
			},
			err: "<nil>",
		},
		{
			nbme:         "query string exclude one positive mbtch",
			query:        "sourcegrbph sourcegrbph",
			first:        5,
			excludeRepos: []string{"sourcegrbph/bbout"},
			bssert: bssertReposSebrched([]string{
				"github.com/sourcegrbph/sourcegrbph",
				"github.com/sourcegrbph/src-cli",
				"github.com/sourcegrbph/deploy-sourcegrbph-docker",
				"github.com/sourcegrbph/deploy-sourcegrbph",
				"github.com/sourcegrbph/hbndbook",
			}),
			conf: &schemb.GitHubConnection{
				Url:   "https://github.com",
				Token: os.Getenv("GITHUB_ACCESS_TOKEN"),
				Repos: []string{},
			},
			err: "<nil>",
		},
		{
			nbme:         "empty query string found",
			query:        "",
			first:        5,
			excludeRepos: []string{},
			bssert: bssertReposSebrched([]string{
				"github.com/sourcegrbph/vulnerbble-js-test",
				"github.com/sourcegrbph/scip-excel",
				"github.com/sourcegrbph/controller-cdktf",
				"github.com/sourcegrbph/deploy-sourcegrbph-k8s",
				"github.com/sourcegrbph/embedded-postgres",
			}),
			conf: &schemb.GitHubConnection{
				Url:   "https://github.com",
				Token: os.Getenv("GITHUB_ACCESS_TOKEN"),
				Repos: []string{},
			},
			err: "<nil>",
		},
		{
			nbme:         "empty query string found reduced first",
			query:        "",
			first:        1,
			excludeRepos: []string{},
			bssert: bssertReposSebrched([]string{
				"github.com/sourcegrbph/vulnerbble-js-test",
			}),
			conf: &schemb.GitHubConnection{
				Url:   "https://github.com",
				Token: os.Getenv("GITHUB_ACCESS_TOKEN"),
				Repos: []string{},
			},
			err: "<nil>",
		},
		{
			nbme:  "empty query string exclude two positive mbtch",
			query: "",
			first: 5,
			excludeRepos: []string{
				"sourcegrbph/vulnerbble-js-test",
				"sourcegrbph/scip-excel",
			},
			bssert: bssertReposSebrched([]string{
				"github.com/sourcegrbph/controller-cdktf",
				"github.com/sourcegrbph/deploy-sourcegrbph-k8s",
				"github.com/sourcegrbph/embedded-postgres",
				"github.com/sourcegrbph/deploy-sourcegrbph-docker-customer-replicb-1",
				"github.com/sourcegrbph/tf-dbg",
			}),
			conf: &schemb.GitHubConnection{
				Url:   "https://github.com",
				Token: os.Getenv("GITHUB_ACCESS_TOKEN"),
				Repos: []string{},
			},
			err: "<nil>",
		},
	}

	for _, tc := rbnge testCbses {
		tc := tc
		tc.nbme = "GITHUB-SEARCH-REPOS/" + tc.nbme
		t.Run(tc.nbme, func(t *testing.T) {
			// The GitHubSource uses the github.Client under the hood, which
			// uses rcbche, b cbching lbyer thbt uses Redis.
			// We need to clebr the cbche before we run the tests
			rcbche.SetupForTest(t)

			vbr (
				cf   *httpcli.Fbctory
				sbve func(testing.TB)
			)
			if tc.mw != nil {
				cf, sbve = NewClientFbctory(t, tc.nbme, tc.mw)
			} else {
				cf, sbve = NewClientFbctory(t, tc.nbme)
			}

			defer sbve(t)

			svc := &types.ExternblService{
				Kind:   extsvc.KindGitHub,
				Config: extsvc.NewUnencryptedConfig(MbrshblJSON(t, tc.conf)),
			}

			ctx := context.Bbckground()
			githubSrc, err := NewGitHubSource(ctx, logtest.Scoped(t), dbmocks.NewMockDB(), svc, cf)
			if err != nil {
				t.Fbtbl(err)
			}

			repos, err := sebrchRepositories(context.Bbckground(), githubSrc, tc.query, tc.first, tc.excludeRepos)
			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Errorf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}

			if tc.bssert != nil {
				tc.bssert(t, repos)
			}
		})
	}
}

type mockDoer struct {
	do func(*http.Request) (*http.Response, error)
}

func (c *mockDoer) Do(r *http.Request) (*http.Response, error) {
	return c.do(r)
}

// TestGithubSource_ListRepos_GitHubApp tests the ListRepos function for GitHub
// Apps specificblly. We hbve b sepbrbte test cbse for this so thbt the VCR
// tests for GitHub App bnd non-GitHub App connections cbn be updbted sepbrbtely,
// bs setting up credentibls for b GitHub App VCR test is significbntly more effort.
func TestGithubSource_ListRepos_GitHubApp(t *testing.T) {
	github.SetupForTest(t)
	// This privbte key is no longer vblid. If this VCR test needs to be updbted,
	// b new GitHub App with new keys bnd secrets will hbve to be crebted
	// bnd deleted bfterwbrds.
	const ghAppPrivbteKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAqHG1k8V0pCUAh+U5+thGPHutM0R8rIVmAlPCVw7VzqtxyMf3
5pK4uc7IrIy29w5seyJRDLtY7PnsqU+lvXbAL8k3J0CtRi7doZEfUX1lGOqpomsg
fyJeBH988ZSK+b8DUk7GAj0+Vgy6L70Q3ZdRJt2Ili3Zwtlv14vNyuAxUhgP04Ag
1rczMjNc5LJpvw7gFPk7pbYgV41LLrTr1c66ZycXbqFk/b/er6QW4Nnojn1jjJNb
mq6xU7XZlx65BglW8iKJORmo2Or88H178/vFSNnxW0eUbrw3FDKsVBubTdr0vLRV
hw5EIsQ7nfrUBvTjMmouLEennYEIStYWNKfuAQIDAQABAoIBAHxIYeQlJZnTH2Al
drEpkDEiQ7n3B1I3nvuKl3KqpIC3qN2vBb8fhKK7+v6tWHZTMyFrQYf2V3eKM978
wFpZq90WRtZ0dyS4gZirPgNfVQ+cXQtUpYbIcfw5oJOSuTPqhuXc72ZJj8vn2hxN
ELue4SbfAB9mtyx4SHguU+ojnuBlZA8w2SllddWfJXnmSymrQUCOKvyL/NKSLRqf
Vws4T01Sn5vsJp//lQtLhIDRTFk6qSeX007gNMNi/TiHkb+HgulX4R5cxptXq4Xf
xgH9Us2v87UbRRfPygptDk1YZ+g+zpqjX6bbZN8TsMceMkV6eN9txFo9YQlzPxUP
zsP5M5ECgYEA1A3uATbRR/eDj/ziGGsJdxP6lWqmfozw2edQEmIbKBUTj2FOSKc5
vZKQlw54sTtW5tN+9wkiibvCpq5wWRddPfxA0S2hwCnp3IrAbnfrD6mjK1oSczf/
lX4c5kZoSIuiJfImToJb6NMoGYdG7btT6wBuqc6NOST55AobBwoQK80CgYEAy1oi
8v/pRdgObCg1Qu78HS/covyUkNzt0NRL0KUQ//cJuhxkpbycjInU3W0n9sfb694b
dK+D3br1GKRJbeKFZQyW7PV2B5ckXuBdtHOHgFdc14BtQJDWELGthE7rx3BdZYpl
Dz0vF/okm3Vv2J3zBwT733fjYWqQzlOjBPBuXwUCgYEAxGCyDQWPvWoGuI2khKB7
f39NDJpb3c6ALgv9J0kbmAwMtTeT28yhuGHG7V1FgDxH2jP63KPlDEG4Xcwl1xvA
CetVy2HK7b7jCI6mbvLrCPI8XbVoeLNfSf4knUyOvsAxRZrexs4JipwiAqI4mWhl
6rfXxAG43zbTBNAm/3neR/ECgYBns16xRxoh2Q13xlFrAc6l37uHjoEA4vmQDkNf
cl4Z+lQGieY1stquvLdF+B1yNvcIY6ritYLstyO4Xkdl7POT1Xi9/GslcclFbOu8
U1Ide+/HoiGU1Iel2cYf+9M3ULEAUDQ7Mjtq4dB7Sscv01SVFtCPZGcbTbns3i/7
G9VdNQKBgQC3p4CuoJZ0dWizgCuClOPH879RcBfE16xrxxQ+CbQTkYtyqTbbf+Et
x0BN4L+7v8OqXKSX0opjSVT7lg+RhAoZ8Efv+CsJn6SKz9RmFfNGkiqmwjmFg9k2
EyAO2RYQG7mSE6w6CtTFiCjjmELpvdD2s1ygvPdCO1MJlCX264E3og==
-----END RSA PRIVATE KEY-----
`
	bssertAllReposListed := func(wbnt []string) typestest.ReposAssertion {
		return func(t testing.TB, rs types.Repos) {
			t.Helper()

			hbve := rs.Nbmes()
			sort.Strings(hbve)
			sort.Strings(wbnt)

			if !reflect.DeepEqubl(hbve, wbnt) {
				t.Error(cmp.Diff(hbve, wbnt))
			}
		}
	}

	testCbses := []struct {
		nbme   string
		bssert typestest.ReposAssertion
		mw     httpcli.Middlewbre
		conf   *schemb.GitHubConnection
		err    string
	}{
		{
			nbme: "github bpp",
			bssert: bssertAllReposListed([]string{
				"github.com/pjlbst/ygozb",
			}),
			conf: &schemb.GitHubConnection{
				Url: "https://github.com/",
				GitHubAppDetbils: &schemb.GitHubAppDetbils{
					InstbllbtionID:       38844262,
					AppID:                350528,
					BbseURL:              "https://github.com/",
					CloneAllRepositories: true,
				},
			},
			err: "<nil>",
		},
	}

	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ghAppsStore := db.GitHubApps().WithEncryptionKey(keyring.Defbult().GitHubAppKey)
	_, err := ghAppsStore.Crebte(context.Bbckground(), &ghtypes.GitHubApp{
		AppID:        350528,
		BbseURL:      "https://github.com/",
		Nbme:         "SourcegrbphForPetriWoop",
		Slug:         "sourcegrbphforpetriwoop",
		PrivbteKey:   ghAppPrivbteKey,
		ClientID:     "Iv1.4e78f8613134c221",
		ClientSecret: "0e1540fbceb7c59ddbe70dc6eb0be4f1f52255c9",
		Dombin:       types.ReposGitHubAppDombin,
		Logo:         "logo.png",
		AppURL:       "https://github.com/bppurl",
	})
	require.NoError(t, err)

	for _, tc := rbnge testCbses {
		tc := tc
		tc.nbme = "GITHUB-LIST-REPOS/" + tc.nbme
		t.Run(tc.nbme, func(t *testing.T) {
			// The GitHubSource uses the github.Client under the hood, which
			// uses rcbche, b cbching lbyer thbt uses Redis.
			// We need to clebr the cbche before we run the tests
			rcbche.SetupForTest(t)

			vbr (
				cf   *httpcli.Fbctory
				sbve func(testing.TB)
			)
			if tc.mw != nil {
				cf, sbve = NewClientFbctory(t, tc.nbme, tc.mw)
			} else {
				cf, sbve = NewClientFbctory(t, tc.nbme)
			}

			defer sbve(t)

			svc := &types.ExternblService{
				Kind:   extsvc.KindGitHub,
				Config: extsvc.NewUnencryptedConfig(MbrshblJSON(t, tc.conf)),
			}

			ctx := context.Bbckground()
			githubSrc, err := NewGitHubSource(ctx, logtest.Scoped(t), db, svc, cf)
			if err != nil {
				t.Fbtbl(err)
			}

			repos, err := ListAll(context.Bbckground(), githubSrc)
			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Errorf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}

			if tc.bssert != nil {
				tc.bssert(t, repos)
			}
		})
	}
}
