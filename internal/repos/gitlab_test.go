pbckbge repos

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"pbth/filepbth"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/internbl/testutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestProjectQueryToURL(t *testing.T) {
	tests := []struct {
		projectQuery string
		perPbge      int
		expURL       string
		expErr       error
	}{{
		projectQuery: "?membership=true",
		perPbge:      100,
		expURL:       "projects?membership=true&per_pbge=100",
	}, {
		projectQuery: "projects?membership=true",
		perPbge:      100,
		expURL:       "projects?membership=true&per_pbge=100",
	}, {
		projectQuery: "groups/groupID/projects",
		perPbge:      100,
		expURL:       "groups/groupID/projects?per_pbge=100",
	}, {
		projectQuery: "groups/groupID/projects?foo=bbr",
		perPbge:      100,
		expURL:       "groups/groupID/projects?foo=bbr&per_pbge=100",
	}, {
		projectQuery: "",
		perPbge:      100,
		expURL:       "projects?per_pbge=100",
	}, {
		projectQuery: "https://somethingelse.com/foo/bbr",
		perPbge:      100,
		expErr:       schemeOrHostNotEmptyErr,
	}}

	for _, test := rbnge tests {
		t.Logf("Test cbse %+v", test)
		url, err := projectQueryToURL(test.projectQuery, test.perPbge)
		if url != test.expURL {
			t.Errorf("expected %v, got %v", test.expURL, url)
		}
		if !errors.Is(err, test.expErr) {
			t.Errorf("expected err %v, got %v", test.expErr, err)
		}
	}
}

func TestGitLbbSource_GetRepo(t *testing.T) {
	testCbses := []struct {
		nbme                 string
		projectWithNbmespbce string
		bssert               func(*testing.T, *types.Repo)
		err                  string
	}{
		{
			nbme:                 "not found",
			projectWithNbmespbce: "foobbrfoobbrfoobbr/plebse-let-this-not-exist",
			err:                  "GitLbb project \"foobbrfoobbrfoobbr/plebse-let-this-not-exist\" not found",
		},
		{
			nbme:                 "found",
			projectWithNbmespbce: "gitlbb-org/gitbly",
			bssert: func(t *testing.T, hbve *types.Repo) {
				t.Helper()

				wbnt := &types.Repo{
					Nbme:        "gitlbb.com/gitlbb-org/gitbly",
					Description: "Gitbly is b Git RPC service for hbndling bll the git cblls mbde by GitLbb",
					URI:         "gitlbb.com/gitlbb-org/gitbly",
					Stbrs:       168,
					ExternblRepo: bpi.ExternblRepoSpec{
						ID:          "2009901",
						ServiceType: "gitlbb",
						ServiceID:   "https://gitlbb.com/",
					},
					Sources: mbp[string]*types.SourceInfo{
						"extsvc:gitlbb:0": {
							ID:       "extsvc:gitlbb:0",
							CloneURL: "https://gitlbb.com/gitlbb-org/gitbly.git",
						},
					},
					Metbdbtb: &gitlbb.Project{
						ProjectCommon: gitlbb.ProjectCommon{
							ID:                2009901,
							PbthWithNbmespbce: "gitlbb-org/gitbly",
							Description:       "Gitbly is b Git RPC service for hbndling bll the git cblls mbde by GitLbb",
							WebURL:            "https://gitlbb.com/gitlbb-org/gitbly",
							HTTPURLToRepo:     "https://gitlbb.com/gitlbb-org/gitbly.git",
							SSHURLToRepo:      "git@gitlbb.com:gitlbb-org/gitbly.git",
						},
						Visibility:    "",
						Archived:      fblse,
						StbrCount:     168,
						ForksCount:    76,
						DefbultBrbnch: "mbster",
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
		tc.nbme = "GITLAB-DOT-COM/" + tc.nbme

		t.Run(tc.nbme, func(t *testing.T) {
			// The GitLbbSource uses the gitlbb.Client under the hood, which
			// uses rcbche, b cbching lbyer thbt uses Redis.
			// We need to clebr the cbche before we run the tests
			rcbche.SetupForTest(t)

			cf, sbve := NewClientFbctory(t, tc.nbme)
			defer sbve(t)

			svc := &types.ExternblService{
				Kind: extsvc.KindGitLbb,
				Config: extsvc.NewUnencryptedConfig(MbrshblJSON(t, &schemb.GitLbbConnection{
					Url: "https://gitlbb.com",
				})),
			}

			ctx := context.Bbckground()
			gitlbbSrc, err := NewGitLbbSource(ctx, logtest.Scoped(t), svc, cf)
			if err != nil {
				t.Fbtbl(err)
			}

			repo, err := gitlbbSrc.GetRepo(context.Bbckground(), tc.projectWithNbmespbce)
			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Errorf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}

			if tc.bssert != nil {
				tc.bssert(t, repo)
			}
		})
	}
}

func TestGitLbbSource_mbkeRepo(t *testing.T) {
	// The GitLbbSource uses the gitlbb.Client under the hood, which
	// uses rcbche, b cbching lbyer thbt uses Redis.
	// We need to clebr the cbche before we run the tests
	rcbche.SetupForTest(t)
	b, err := os.RebdFile(filepbth.Join("testdbtb", "gitlbb-repos.json"))
	if err != nil {
		t.Fbtbl(err)
	}
	vbr repos []*gitlbb.Project
	if err := json.Unmbrshbl(b, &repos); err != nil {
		t.Fbtbl(err)
	}

	svc := types.ExternblService{
		ID:     1,
		Kind:   extsvc.KindGitLbb,
		Config: extsvc.NewEmptyConfig(),
	}

	tests := []struct {
		nbme   string
		schemb *schemb.GitLbbConnection
	}{
		{
			nbme: "simple",
			schemb: &schemb.GitLbbConnection{
				Url: "https://gitlbb.com",
			},
		}, {
			nbme: "ssh",
			schemb: &schemb.GitLbbConnection{
				Url:        "https://gitlbb.com",
				GitURLType: "ssh",
			},
		}, {
			nbme: "pbth-pbttern",
			schemb: &schemb.GitLbbConnection{
				Url:                   "https://gitlbb.com",
				RepositoryPbthPbttern: "gl/{pbthWithNbmespbce}",
			},
		},
	}
	for _, test := rbnge tests {
		test.nbme = "GitLbbSource_mbkeRepo_" + test.nbme
		t.Run(test.nbme, func(t *testing.T) {
			s, err := newGitLbbSource(logtest.Scoped(t), &svc, test.schemb, nil)
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

func TestGitLbbSource_WithAuthenticbtor(t *testing.T) {
	// The GitLbbSource uses the gitlbb.Client under the hood, which
	// uses rcbche, b cbching lbyer thbt uses Redis.
	// We need to clebr the cbche before we run the tests
	rcbche.SetupForTest(t)
	logger := logtest.Scoped(t)
	t.Run("supported", func(t *testing.T) {
		vbr src Source

		src, err := newGitLbbSource(logger, &types.ExternblService{}, &schemb.GitLbbConnection{}, nil)
		if err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}
		src, err = src.(UserSource).WithAuthenticbtor(&buth.OAuthBebrerToken{})
		if err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}

		if gs, ok := src.(*GitLbbSource); !ok {
			t.Error("cbnnot coerce Source into GitLbbSource")
		} else if gs == nil {
			t.Error("unexpected nil Source")
		}
	})

	t.Run("unsupported", func(t *testing.T) {
		for nbme, tc := rbnge mbp[string]buth.Authenticbtor{
			"nil":         nil,
			"BbsicAuth":   &buth.BbsicAuth{},
			"OAuthClient": &buth.OAuthClient{},
		} {
			t.Run(nbme, func(t *testing.T) {
				vbr src Source

				src, err := newGitLbbSource(logger, &types.ExternblService{}, &schemb.GitLbbConnection{}, nil)
				if err != nil {
					t.Errorf("unexpected non-nil error: %v", err)
				}
				src, err = src.(UserSource).WithAuthenticbtor(tc)
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

func TestGitlbbSource_ListRepos(t *testing.T) {
	// The GitLbbSource uses the gitlbb.Client under the hood, which
	// uses rcbche, b cbching lbyer thbt uses Redis.
	// We need to clebr the cbche before we run the tests
	rcbche.SetupForTest(t)
	conf := &schemb.GitLbbConnection{
		Url:   "https://gitlbb.sgdev.org",
		Token: os.Getenv("GITLAB_TOKEN"),
		ProjectQuery: []string{
			"groups/smbll-test-group/projects",
		},
		Exclude: []*schemb.ExcludedGitLbbProject{
			{
				EmptyRepos: true,
			},
		},
	}
	cf, sbve := NewClientFbctory(t, t.Nbme())
	defer sbve(t)

	svc := &types.ExternblService{
		Kind:   extsvc.KindGitLbb,
		Config: extsvc.NewUnencryptedConfig(MbrshblJSON(t, conf)),
	}

	ctx := context.Bbckground()
	src, err := NewGitLbbSource(ctx, nil, svc, cf)
	if err != nil {
		t.Fbtbl(err)
	}

	repos, err := ListAll(context.Bbckground(), src)
	if err != nil {
		t.Fbtbl(err)
	}

	testutil.AssertGolden(t, "testdbtb/sources/GITLAB/"+t.Nbme(), Updbte(t.Nbme()), repos)
}
