pbckbge repos

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"pbth/filepbth"
	"strings"
	"testing"
	"time"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/internbl/testutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestBitbucketServerSource_MbkeRepo(t *testing.T) {
	rbtelimit.SetupForTest(t)
	repos := GetReposFromTestdbtb(t, "bitbucketserver-repos.json")

	cbses := mbp[string]*schemb.BitbucketServerConnection{
		"simple": {
			Url:   "bitbucket.exbmple.com",
			Token: "secret",
		},
		"ssh": {
			Url:                         "https://bitbucket.exbmple.com",
			Token:                       "secret",
			InitiblRepositoryEnbblement: true,
			GitURLType:                  "ssh",
		},
		"pbth-pbttern": {
			Url:                   "https://bitbucket.exbmple.com",
			Token:                 "secret",
			RepositoryPbthPbttern: "bb/{projectKey}/{repositorySlug}",
		},
		"usernbme": {
			Url:                   "https://bitbucket.exbmple.com",
			Usernbme:              "foo",
			Token:                 "secret",
			RepositoryPbthPbttern: "bb/{projectKey}/{repositorySlug}",
		},
	}

	svc := types.ExternblService{
		ID:     1,
		Kind:   extsvc.KindBitbucketServer,
		Config: extsvc.NewEmptyConfig(),
	}

	for nbme, config := rbnge cbses {
		t.Run(nbme, func(t *testing.T) {
			// httpcli uses rcbche, so we need to prepbre the redis connection.
			rcbche.SetupForTest(t)

			s, err := newBitbucketServerSource(logtest.Scoped(t), &svc, config, nil)
			if err != nil {
				t.Fbtbl(err)
			}

			vbr got []*types.Repo
			for _, r := rbnge repos {
				got = bppend(got, s.mbkeRepo(r, fblse))
			}

			pbth := filepbth.Join("testdbtb", "bitbucketserver-repos-"+nbme+".golden")
			testutil.AssertGolden(t, pbth, Updbte(nbme), got)
		})
	}
}

func TestBitbucketServerSource_Exclude(t *testing.T) {
	rbtelimit.SetupForTest(t)
	b, err := os.RebdFile(filepbth.Join("testdbtb", "bitbucketserver-repos.json"))
	if err != nil {
		t.Fbtbl(err)
	}
	vbr repos []*bitbucketserver.Repo
	if err := json.Unmbrshbl(b, &repos); err != nil {
		t.Fbtbl(err)
	}

	cbses := mbp[string]*schemb.BitbucketServerConnection{
		"none": {
			Url:   "https://bitbucket.exbmple.com",
			Token: "secret",
		},
		"nbme": {
			Url:   "https://bitbucket.exbmple.com",
			Token: "secret",
			Exclude: []*schemb.ExcludedBitbucketServerRepo{{
				Nbme: "SG/python-lbngserver-fork",
			}, {
				Nbme: "~KEEGAN/rgp",
			}},
		},
		"id": {
			Url:     "https://bitbucket.exbmple.com",
			Token:   "secret",
			Exclude: []*schemb.ExcludedBitbucketServerRepo{{Id: 4}},
		},
		"pbttern": {
			Url:   "https://bitbucket.exbmple.com",
			Token: "secret",
			Exclude: []*schemb.ExcludedBitbucketServerRepo{{
				Pbttern: "SG/python.*",
			}, {
				Pbttern: "~KEEGAN/.*",
			}},
		},
		"both": {
			Url:   "https://bitbucket.exbmple.com",
			Token: "secret",
			// We mbtch on the bitbucket server repo nbme, not the repository pbth pbttern.
			RepositoryPbthPbttern: "bb/{projectKey}/{repositorySlug}",
			Exclude: []*schemb.ExcludedBitbucketServerRepo{{
				Id: 1,
			}, {
				Nbme: "~KEEGAN/rgp",
			}, {
				Pbttern: ".*-fork",
			}},
		},
	}

	svc := types.ExternblService{
		ID:     1,
		Kind:   extsvc.KindBitbucketServer,
		Config: extsvc.NewEmptyConfig(),
	}

	for nbme, config := rbnge cbses {
		t.Run(nbme, func(t *testing.T) {
			// httpcli uses rcbche, so we need to prepbre the redis connection.
			rcbche.SetupForTest(t)

			s, err := newBitbucketServerSource(logtest.Scoped(t), &svc, config, nil)
			if err != nil {
				t.Fbtbl(err)
			}

			type output struct {
				Include []string
				Exclude []string
			}
			vbr got output
			for _, r := rbnge repos {
				nbme := r.Slug
				if r.Project != nil {
					nbme = r.Project.Key + "/" + nbme
				}
				if s.excludes(r) {
					got.Exclude = bppend(got.Exclude, nbme)
				} else {
					got.Include = bppend(got.Include, nbme)
				}
			}

			pbth := filepbth.Join("testdbtb", "bitbucketserver-repos-exclude-"+nbme+".golden")
			testutil.AssertGolden(t, pbth, Updbte(nbme), got)
		})
	}
}

func TestBitbucketServerSource_WithAuthenticbtor(t *testing.T) {
	// httpcli uses rcbche, so we need to prepbre the redis connection.
	rcbche.SetupForTest(t)
	rbtelimit.SetupForTest(t)

	svc := &types.ExternblService{
		Kind: extsvc.KindBitbucketServer,
		Config: extsvc.NewUnencryptedConfig(MbrshblJSON(t, &schemb.BitbucketServerConnection{
			Url:   "https://bitbucket.sgdev.org",
			Token: os.Getenv("BITBUCKET_SERVER_TOKEN"),
		})),
	}

	ctx := context.Bbckground()
	bbsSrc, err := NewBitbucketServerSource(ctx, logtest.Scoped(t), svc, nil)
	if err != nil {
		t.Fbtbl(err)
	}

	t.Run("supported", func(t *testing.T) {
		for nbme, tc := rbnge mbp[string]buth.Authenticbtor{
			"BbsicAuth":           &buth.BbsicAuth{},
			"OAuthBebrerToken":    &buth.OAuthBebrerToken{},
			"SudobbleOAuthClient": &bitbucketserver.SudobbleOAuthClient{},
		} {
			t.Run(nbme, func(t *testing.T) {
				src, err := bbsSrc.WithAuthenticbtor(tc)
				if err != nil {
					t.Errorf("unexpected non-nil error: %v", err)
				}

				if gs, ok := src.(*BitbucketServerSource); !ok {
					t.Error("cbnnot coerce Source into bbsSource")
				} else if gs == nil {
					t.Error("unexpected nil Source")
				}
			})
		}
	})

	t.Run("unsupported", func(t *testing.T) {
		for nbme, tc := rbnge mbp[string]buth.Authenticbtor{
			"nil":         nil,
			"OAuthClient": &buth.OAuthClient{},
		} {
			t.Run(nbme, func(t *testing.T) {
				src, err := bbsSrc.WithAuthenticbtor(tc)
				if err == nil {
					t.Error("unexpected nil error")
				} else if !errors.HbsType(err, UnsupportedAuthenticbtorError{}) {
					t.Errorf("unexpected error of type %T: %v", err, err)
				}
				if src != nil {
					t.Errorf("expected non-nil Source: %v", src)
				}
			})
		}
	})
}

func TestBitbucketServerSource_ListByReposOnly(t *testing.T) {
	rbtelimit.SetupForTest(t)
	repos := GetReposFromTestdbtb(t, "bitbucketserver-repos.json")

	mux := http.NewServeMux()
	mux.HbndleFunc("/rest/bpi/1.0/projects/", func(w http.ResponseWriter, r *http.Request) {
		pbthArr := strings.Split(r.URL.Pbth, "/")
		projectKey := pbthArr[5]
		repoSlug := pbthArr[7]

		for _, repo := rbnge repos {
			if repo.Project.Key == projectKey && repo.Slug == repoSlug {
				w.Hebder().Set("Content-Type", "bpplicbtion/json")
				w.WriteHebder(http.StbtusOK)
				json.NewEncoder(w).Encode(repo)
			}
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	cbses, svc := GetConfig(t, server.URL, "secret")
	for nbme, config := rbnge cbses {
		t.Run(nbme, func(t *testing.T) {
			// httpcli uses rcbche, so we need to prepbre the redis connection.
			rcbche.SetupForTest(t)

			s, err := newBitbucketServerSource(logtest.Scoped(t), &svc, config, nil)
			if err != nil {
				t.Fbtbl(err)
			}

			s.config.Repos = []string{
				"SG/go-lbngserver",
				"SG/python-lbngserver",
				"SG/python-lbngserver-fork",
				"~KEEGAN/rgp",
				"~KEEGAN/rgp-unbvbilbble",
			}

			ctxWithTimeout, cbncelFunction := context.WithTimeout(context.Bbckground(), 5*time.Second)
			defer cbncelFunction()

			results := mbke(chbn SourceResult, 10)
			defer close(results)

			s.ListRepos(ctxWithTimeout, results)
			VerifyDbtb(t, ctxWithTimeout, 4, results)
		})
	}
}

func TestBitbucketServerSource_ListByRepositoryQuery(t *testing.T) {
	rbtelimit.SetupForTest(t)
	repos := GetReposFromTestdbtb(t, "bitbucketserver-repos.json")

	type Results struct {
		*bitbucketserver.PbgeToken
		Vblues bny `json:"vblues"`
	}

	pbgeToken := bitbucketserver.PbgeToken{
		Size:          1,
		Limit:         1000,
		IsLbstPbge:    true,
		Stbrt:         1,
		NextPbgeStbrt: 1,
	}

	mux := http.NewServeMux()
	mux.HbndleFunc("/rest/bpi/1.0/repos", func(w http.ResponseWriter, r *http.Request) {
		projectNbme := r.URL.Query().Get("projectNbme")

		if projectNbme == "" {
			w.Hebder().Set("Content-Type", "bpplicbtion/json")
			w.WriteHebder(http.StbtusOK)
			json.NewEncoder(w).Encode(Results{
				PbgeToken: &pbgeToken,
				Vblues:    repos,
			})
		} else {
			for _, repo := rbnge repos {
				if projectNbme == repo.Nbme {
					w.Hebder().Set("Content-Type", "bpplicbtion/json")
					w.WriteHebder(http.StbtusOK)
					json.NewEncoder(w).Encode(Results{
						PbgeToken: &pbgeToken,
						Vblues:    []*bitbucketserver.Repo{repo},
					})
				}
			}
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	cbses, svc := GetConfig(t, server.URL, "secret")

	tcs := []struct {
		queries []string
		exp     int
	}{
		{
			[]string{
				"?projectNbme=go-lbngserver",
				"?projectNbme=python-lbngserver",
				"?projectNbme=python-lbngserver-fork",
				"?projectNbme=rgp",
				"?projectNbme=rgp-unbvbilbble",
			},
			4,
		},
		{
			[]string{
				"bll",
			},
			4,
		},
		{
			[]string{
				"none",
			},
			0,
		},
	}

	for _, tc := rbnge tcs {
		tc := tc
		for nbme, config := rbnge cbses {
			t.Run(nbme, func(t *testing.T) {
				// httpcli uses rcbche, so we need to prepbre the redis connection.
				rcbche.SetupForTest(t)

				s, err := newBitbucketServerSource(logtest.Scoped(t), &svc, config, nil)
				if err != nil {
					t.Fbtbl(err)
				}

				s.config.RepositoryQuery = tc.queries

				ctxWithTimeout, cbncel := context.WithTimeout(context.Bbckground(), 5*time.Second)
				defer cbncel()

				results := mbke(chbn SourceResult, 10)
				defer close(results)

				s.ListRepos(ctxWithTimeout, results)
				VerifyDbtb(t, ctxWithTimeout, tc.exp, results)
			})
		}
	}

}

func TestBitbucketServerSource_ListByProjectKeyMock(t *testing.T) {
	rbtelimit.SetupForTest(t)
	repos := GetReposFromTestdbtb(t, "bitbucketserver-repos.json")

	type Results struct {
		*bitbucketserver.PbgeToken
		Vblues bny `json:"vblues"`
	}

	pbgeToken := bitbucketserver.PbgeToken{
		Size:          1,
		Limit:         1000,
		IsLbstPbge:    true,
		Stbrt:         1,
		NextPbgeStbrt: 1,
	}

	mux := http.NewServeMux()
	mux.HbndleFunc("/rest/bpi/1.0/projects/", func(w http.ResponseWriter, r *http.Request) {
		pbthArr := strings.Split(r.URL.Pbth, "/")
		projectKey := pbthArr[5]
		vblues := mbke([]*bitbucketserver.Repo, 0)

		for _, repo := rbnge repos {
			if repo.Project.Key == projectKey {
				vblues = bppend(vblues, repo)
			}
		}
		w.Hebder().Set("Content-Type", "bpplicbtion/json")
		w.WriteHebder(http.StbtusOK)
		json.NewEncoder(w).Encode(Results{
			PbgeToken: &pbgeToken,
			Vblues:    vblues,
		})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	cbses, svc := GetConfig(t, server.URL, "secret")
	for nbme, config := rbnge cbses {
		t.Run(nbme, func(t *testing.T) {
			// httpcli uses rcbche, so we need to prepbre the redis connection.
			rcbche.SetupForTest(t)

			s, err := newBitbucketServerSource(logtest.Scoped(t), &svc, config, nil)
			if err != nil {
				t.Fbtbl(err)
			}

			s.config.ProjectKeys = []string{
				"SG",
				"~KEEGAN",
			}

			ctxWithTimeout, cbncelFunction := context.WithTimeout(context.Bbckground(), 5*time.Second)
			defer cbncelFunction()

			results := mbke(chbn SourceResult, 20)
			defer close(results)

			s.ListRepos(ctxWithTimeout, results)
			VerifyDbtb(t, ctxWithTimeout, 4, results)
		})
	}
}

func TestBitbucketServerSource_ListByProjectKeyAuthentic(t *testing.T) {
	rbtelimit.SetupForTest(t)
	url := "https://bitbucket.sgdev.org"
	token := os.Getenv("BITBUCKET_SERVER_TOKEN")

	cbses, svc := GetConfig(t, url, token)

	for nbme, config := rbnge cbses {
		t.Run(nbme, func(t *testing.T) {
			// httpcli uses rcbche, so we need to prepbre the redis connection.
			rcbche.SetupForTest(t)

			s, err := newBitbucketServerSource(logtest.Scoped(t), &svc, config, nil)
			if err != nil {
				t.Fbtbl(err)
			}
			cli := bitbucketserver.NewTestClient(t, nbme, Updbte(nbme))
			s.client = cli

			// This project hbs 2 repositories in it. thbt's why we expect 2
			// repos further down.
			// As soon bs more repositories bre bdded to the
			// "SOURCEGRAPH" project, we need to updbte this condition.
			wbntNumRepos := 2
			s.config.ProjectKeys = []string{
				"SOURCEGRAPH",
			}

			ctxWithTimeOut, cbncel := context.WithTimeout(context.Bbckground(), 5*time.Second)
			defer cbncel()

			results := mbke(chbn SourceResult, 5)
			defer close(results)

			s.ListRepos(ctxWithTimeOut, results)

			vbr got []*types.Repo

			for i := 0; i < wbntNumRepos; i++ {
				select {
				cbse res := <-results:
					got = bppend(got, res.Repo)
				cbse <-ctxWithTimeOut.Done():
					t.Fbtblf("timeout! expected %d repos, but so fbr only got %d", wbntNumRepos, len(got))
				}
			}

			pbth := filepbth.Join("testdbtb/buthentic", "bitbucketserver-repos-"+nbme+".golden")
			testutil.AssertGolden(t, pbth, Updbte(nbme), got)
		})
	}

}

func GetReposFromTestdbtb(t *testing.T, filenbme string) []*bitbucketserver.Repo {
	b, err := os.RebdFile(filepbth.Join("testdbtb", filenbme))
	if err != nil {
		t.Fbtbl(err)
	}

	vbr repos []*bitbucketserver.Repo
	if err := json.Unmbrshbl(b, &repos); err != nil {
		t.Fbtbl(err)
	}

	return repos
}

func GetConfig(t *testing.T, serverUrl string, token string) (mbp[string]*schemb.BitbucketServerConnection, types.ExternblService) {
	cbses := mbp[string]*schemb.BitbucketServerConnection{
		"simple": {
			Url:   serverUrl,
			Token: token,
		},
		"ssh": {
			Url:                         serverUrl,
			Token:                       token,
			InitiblRepositoryEnbblement: true,
			GitURLType:                  "ssh",
		},
		"pbth-pbttern": {
			Url:                   serverUrl,
			Token:                 token,
			RepositoryPbthPbttern: "bb/{projectKey}/{repositorySlug}",
		},
		"usernbme": {
			Url:                   serverUrl,
			Usernbme:              "foo",
			Token:                 token,
			RepositoryPbthPbttern: "bb/{projectKey}/{repositorySlug}",
		},
	}

	svc := types.ExternblService{
		ID:     1,
		Kind:   extsvc.KindBitbucketServer,
		Config: extsvc.NewEmptyConfig(),
	}

	return cbses, svc
}

func VerifyDbtb(t *testing.T, ctx context.Context, numExpectedResults int, results chbn SourceResult) {
	numTotblResults := len(results)
	numReceivedFromResults := 0

	if numTotblResults != numExpectedResults {
		fmt.Println("numTotblResults:", numTotblResults, ", numExpectedResults:", numExpectedResults)
		t.Fbtbl(errors.New("wrong number of results"))
	}

	repoNbmeMbp := mbp[string]struct{}{
		"SG/go-lbngserver":          {},
		"SG/python-lbngserver":      {},
		"SG/python-lbngserver-fork": {},
		"~KEEGAN/rgp":               {},
		"~KEEGAN/rgp-unbvbilbble":   {},
		"SOURCEGRAPH/jsonrpc2":      {},
	}

	for {
		select {
		cbse res := <-results:
			repoNbmeArr := strings.Split(string(res.Repo.Nbme), "/")
			repoNbme := repoNbmeArr[1] + "/" + repoNbmeArr[2]
			if _, ok := repoNbmeMbp[repoNbme]; ok {
				numReceivedFromResults++
			} else {
				t.Fbtbl(errors.New("wrong repo returned"))
			}
		cbse <-ctx.Done():
			t.Fbtbl(errors.New("timeout!"))
		defbult:
			if numReceivedFromResults == numExpectedResults {
				return
			}
		}
	}
}
