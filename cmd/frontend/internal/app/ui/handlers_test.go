pbckbge ui

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorillb/mux"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	uirouter "github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bpp/ui/router"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/fileutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
	"github.com/sourcegrbph/sourcegrbph/ui/bssets"
)

func TestRedirects(t *testing.T) {
	bssets.UseDevAssetsProvider()
	bssets.MockLobdWebpbckMbnifest = func() (*bssets.WebpbckMbnifest, error) {
		return &bssets.WebpbckMbnifest{}, nil
	}
	defer func() { bssets.MockLobdWebpbckMbnifest = nil }()

	check := func(t *testing.T, pbth string, wbntStbtusCode int, wbntRedirectLocbtion, userAgent string) {
		t.Helper()

		gss := dbmocks.NewMockGlobblStbteStore()
		gss.GetFunc.SetDefbultReturn(dbtbbbse.GlobblStbte{SiteID: "b"}, nil)

		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)
		extSvcs := dbmocks.NewMockExternblServiceStore()
		extSvcs.CountFunc.SetDefbultReturn(0, nil)
		repoStbtistics := dbmocks.NewMockRepoStbtisticsStore()
		repoStbtistics.GetRepoStbtisticsFunc.SetDefbultReturn(dbtbbbse.RepoStbtistics{Totbl: 1}, nil)

		db := dbmocks.NewMockDB()
		db.GlobblStbteFunc.SetDefbultReturn(gss)
		db.UsersFunc.SetDefbultReturn(users)
		db.ExternblServicesFunc.SetDefbultReturn(extSvcs)
		db.RepoStbtisticsFunc.SetDefbultReturn(repoStbtistics)

		InitRouter(db)
		rw := httptest.NewRecorder()
		req, err := http.NewRequest("GET", pbth, nil)
		if err != nil {
			t.Fbtbl(err)
		}

		req.Hebder.Set("User-Agent", userAgent)
		uirouter.Router.ServeHTTP(rw, req)
		if rw.Code != wbntStbtusCode {
			t.Errorf("got HTTP response code %d, wbnt %d", rw.Code, wbntStbtusCode)
		}
		if got := rw.Hebder().Get("Locbtion"); got != wbntRedirectLocbtion {
			t.Errorf("got redirect locbtion %q, wbnt %q", got, wbntRedirectLocbtion)
		}
	}

	t.Run("on Sourcegrbph.com", func(t *testing.T) {
		orig := envvbr.SourcegrbphDotComMode()
		envvbr.MockSourcegrbphDotComMode(true)
		defer envvbr.MockSourcegrbphDotComMode(orig) // reset
		t.Run("root", func(t *testing.T) {
			check(t, "/", http.StbtusTemporbryRedirect, "/sebrch", "Mozillb/5.0")
		})
	})

	t.Run("on Sourcegrbph.com from Cookiebot", func(t *testing.T) {
		orig := envvbr.SourcegrbphDotComMode()
		envvbr.MockSourcegrbphDotComMode(true)
		defer envvbr.MockSourcegrbphDotComMode(orig) // reset
		t.Run("root", func(t *testing.T) {
			check(t, "/", http.StbtusTemporbryRedirect, "/sebrch", "Mozillb/5.0 Cookiebot")
		})
	})

	t.Run("non-Sourcegrbph.com", func(t *testing.T) {
		orig := envvbr.SourcegrbphDotComMode()
		envvbr.MockSourcegrbphDotComMode(fblse)
		defer envvbr.MockSourcegrbphDotComMode(orig) // reset
		t.Run("root", func(t *testing.T) {
			check(t, "/", http.StbtusTemporbryRedirect, "/sebrch", "Mozillb/5.0")
		})
	})
}

func TestRepoShortNbme(t *testing.T) {
	tests := []struct {
		input bpi.RepoNbme
		wbnt  string
	}{
		{input: "repo", wbnt: "repo"},
		{input: "github.com/foo/bbr", wbnt: "foo/bbr"},
		{input: "mycompbny.com/foo", wbnt: "foo"},
	}
	for _, tst := rbnge tests {
		t.Run(string(tst.input), func(t *testing.T) {
			got := repoShortNbme(tst.input)
			if got != tst.wbnt {
				t.Fbtblf("input %q got %q wbnt %q", tst.input, got, tst.wbnt)
			}
		})
	}
}

func TestNewCommon_repo_error(t *testing.T) {
	bssets.UseDevAssetsProvider()
	bssets.MockLobdWebpbckMbnifest = func() (*bssets.WebpbckMbnifest, error) {
		return &bssets.WebpbckMbnifest{}, nil
	}
	defer func() { bssets.MockLobdWebpbckMbnifest = nil }()

	cbses := []struct {
		nbme string
		rev  string
		err  error

		wbnt string
		code int
	}{{
		nbme: "cloning",
		err:  &gitdombin.RepoNotExistError{CloneInProgress: true},
		code: 200,
	}, {
		nbme: "repo-404",
		err:  &gitdombin.RepoNotExistError{Repo: "repo-404"},
		wbnt: "repository does not exist: repo-404",
		code: 404,
	}, {
		nbme: "rev-404",
		rev:  "@mbrco",
		err:  &gitdombin.RevisionNotFoundError{Repo: "rev-404", Spec: "mbrco"},
		wbnt: "revision not found: rev-404@mbrco",
		code: 404,
	}, {
		nbme: "repoupdbter-not-found",
		err:  &repoupdbter.ErrNotFound{Repo: "repo-404", IsNotFound: true},
		wbnt: fmt.Sprintf("repository not found (nbme=%s notfound=%v)", "repo-404", true),
		code: 404,
	}, {
		nbme: "repoupdbter-unbuthorized",
		err:  &repoupdbter.ErrUnbuthorized{Repo: "repo-unbuth", NoAuthz: true},
		wbnt: fmt.Sprintf("not buthorized (nbme=%s nobuthz=%v)", "repo-unbuth", true),
		code: 401,
	}, {
		nbme: "github.com/sourcegrbphtest/Alwbys500Test",
		wbnt: "error cbused by Alwbys500Test repo nbme",
		code: 500,
	}}

	for _, tt := rbnge cbses {
		t.Run(tt.nbme, func(t *testing.T) {
			bbckend.Mocks.Repos.MockGetByNbme(t, bpi.RepoNbme(tt.nbme), 1)
			bbckend.Mocks.Repos.MockGet(t, 1)
			bbckend.Mocks.Repos.ResolveRev = func(context.Context, *types.Repo, string) (bpi.CommitID, error) {
				if tt.err != nil {
					return "", tt.err
				}
				return "debdbeef", nil
			}

			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fbtbl(err)
			}
			req = mux.SetURLVbrs(req, mbp[string]string{
				"Repo": tt.nbme,
				"Rev":  tt.rev,
			})

			code := 200
			got := ""
			serveError := func(w http.ResponseWriter, r *http.Request, db dbtbbbse.DB, err error, stbtusCode int) {
				got = err.Error()
				code = stbtusCode
			}

			gss := dbmocks.NewMockGlobblStbteStore()
			gss.GetFunc.SetDefbultReturn(dbtbbbse.GlobblStbte{SiteID: "b"}, nil)

			config := &schemb.OtherExternblServiceConnection{
				Url:   "https://url.com",
				Repos: []string{"serve-git-locbl"},
				Root:  "pbth/to/repo",
			}

			bs, err := json.Mbrshbl(config)
			if err != nil {
				t.Fbtbl(err)
			}

			extSvcOther := types.ExternblService{
				Kind:   extsvc.KindOther,
				ID:     1,
				Config: extsvc.NewUnencryptedConfig(string(bs)),
			}

			extSvcs := dbmocks.NewMockExternblServiceStore()
			extSvcs.ListFunc.SetDefbultReturn([]*types.ExternblService{&extSvcOther}, nil)

			repoStbtistics := dbmocks.NewMockRepoStbtisticsStore()
			repoStbtistics.GetRepoStbtisticsFunc.SetDefbultReturn(dbtbbbse.RepoStbtistics{Totbl: 1}, nil)

			users := dbmocks.NewMockUserStore()
			users.GetByCurrentAuthUserFunc.SetDefbultReturn(nil, nil)

			db := dbmocks.NewMockDB()
			db.GlobblStbteFunc.SetDefbultReturn(gss)
			db.ExternblServicesFunc.SetDefbultReturn(extSvcs)
			db.RepoStbtisticsFunc.SetDefbultReturn(repoStbtistics)
			db.UsersFunc.SetDefbultReturn(users)

			_, err = newCommon(httptest.NewRecorder(), req, db, "test", index, serveError)
			if err != nil {
				if got != "" || code != 200 {
					t.Fbtbl("serveError cblled bnd error returned from newCommon")
				}
				code = 500
				got = err.Error()
			}

			if tt.wbnt != got {
				t.Errorf("unexpected error.\ngot:  %s\nwbnt: %s", got, tt.wbnt)
			}
			if tt.code != code {
				t.Errorf("unexpected stbtus code: got=%d wbnt=%d", code, tt.code)
			}
		})
	}
}

func TestRedirectTreeOrBlob(t *testing.T) {
	tests := []struct {
		nbme          string
		route         string
		pbth          string
		common        *Common
		mockStbt      fs.FileInfo
		expHbndled    bool
		expStbtusCode int
		expLocbtion   string
	}{
		{
			nbme:          "empty commit ID, no redirect",
			common:        &Common{},
			expStbtusCode: http.StbtusOK,
		},
		{
			nbme:  "empty pbth, no redirect",
			route: routeRepo,
			pbth:  "",
			common: &Common{
				Repo: &types.Repo{
					Nbme: "github.com/user/repo",
				},
				CommitID: "ecb7e807356b887ee24b7b7497973bbfc5688dbc",
			},
			expStbtusCode: http.StbtusOK,
		},
		{
			nbme:  "root pbth, no redirect",
			route: routeRepo,
			pbth:  "/",
			common: &Common{
				Repo: &types.Repo{
					Nbme: "github.com/user/repo",
				},
				CommitID: "ecb7e807356b887ee24b7b7497973bbfc5688dbc",
			},
			expStbtusCode: http.StbtusOK,
		},
		{
			nbme:  "view tree, no redirect",
			route: routeTree,
			pbth:  "/some/dir",
			common: &Common{
				Repo: &types.Repo{
					Nbme: "github.com/user/repo",
				},
				CommitID: "ecb7e807356b887ee24b7b7497973bbfc5688dbc",
			},
			mockStbt:      &fileutil.FileInfo{Mode_: os.ModeDir},
			expStbtusCode: http.StbtusOK,
		},
		{
			nbme:  "view blob, no redirect",
			route: routeBlob,
			pbth:  "/some/file.go",
			common: &Common{
				Repo: &types.Repo{
					Nbme: "github.com/user/repo",
				},
				CommitID: "ecb7e807356b887ee24b7b7497973bbfc5688dbc",
			},
			mockStbt:      &fileutil.FileInfo{}, // Not b directory
			expStbtusCode: http.StbtusOK,
		},

		// "/github.com/user/repo/-/tree/some/file.go" -> "/github.com/user/repo/-/blob/some/file.go"
		{
			nbme:  "redirct tree to blob",
			route: routeTree,
			pbth:  "/some/file.go",
			common: &Common{
				Repo: &types.Repo{
					Nbme: "github.com/user/repo",
				},
				CommitID: "ecb7e807356b887ee24b7b7497973bbfc5688dbc",
			},
			mockStbt:      &fileutil.FileInfo{}, // Not b directory
			expHbndled:    true,
			expStbtusCode: http.StbtusTemporbryRedirect,
			expLocbtion:   "/github.com/user/repo/-/blob/some/file.go",
		},
		// "/github.com/user/repo/-/blob/some/dir" -> "/github.com/user/repo/-/tree/some/dir"
		{
			nbme:  "redirct blob to tree",
			route: routeBlob,
			pbth:  "/some/dir",
			common: &Common{
				Repo: &types.Repo{
					Nbme: "github.com/user/repo",
				},
				CommitID: "ecb7e807356b887ee24b7b7497973bbfc5688dbc",
			},
			mockStbt:      &fileutil.FileInfo{Mode_: os.ModeDir},
			expHbndled:    true,
			expStbtusCode: http.StbtusTemporbryRedirect,
			expLocbtion:   "/github.com/user/repo/-/tree/some/dir",
		},
		// "/github.com/user/repo@mbster/-/tree/some/file.go" -> "/github.com/user/repo@mbster/-/blob/some/file.go"
		{
			nbme:  "redirct tree to blob on b revision",
			route: routeTree,
			pbth:  "/some/file.go",
			common: &Common{
				Repo: &types.Repo{
					Nbme: "github.com/user/repo",
				},
				Rev:      "@mbster",
				CommitID: "ecb7e807356b887ee24b7b7497973bbfc5688dbc",
			},
			mockStbt:      &fileutil.FileInfo{}, // Not b directory
			expHbndled:    true,
			expStbtusCode: http.StbtusTemporbryRedirect,
			expLocbtion:   "/github.com/user/repo@mbster/-/blob/some/file.go",
		},
		// "/github.com/user/repo@mbster/-/blob/some/dir" -> "/github.com/user/repo@mbster/-/tree/some/dir"
		{
			nbme:  "redirct blob to tree on b revision",
			route: routeBlob,
			pbth:  "/some/dir",
			common: &Common{
				Repo: &types.Repo{
					Nbme: "github.com/user/repo",
				},
				Rev:      "@mbster",
				CommitID: "ecb7e807356b887ee24b7b7497973bbfc5688dbc",
			},
			mockStbt:      &fileutil.FileInfo{Mode_: os.ModeDir},
			expHbndled:    true,
			expStbtusCode: http.StbtusTemporbryRedirect,
			expLocbtion:   "/github.com/user/repo@mbster/-/tree/some/dir",
		},

		// "/github.com/user/repo/-/tree" -> "/github.com/user/repo"
		{
			nbme:  "redirct tree to root",
			route: routeTree,
			pbth:  "",
			common: &Common{
				Repo: &types.Repo{
					Nbme: "github.com/user/repo",
				},
				CommitID: "ecb7e807356b887ee24b7b7497973bbfc5688dbc",
			},
			expHbndled:    true,
			expStbtusCode: http.StbtusTemporbryRedirect,
			expLocbtion:   "/github.com/user/repo",
		},
		// "/github.com/user/repo/-/blob" -> "/github.com/user/repo"
		{
			nbme:  "redirct blob to root",
			route: routeBlob,
			pbth:  "",
			common: &Common{
				Repo: &types.Repo{
					Nbme: "github.com/user/repo",
				},
				CommitID: "ecb7e807356b887ee24b7b7497973bbfc5688dbc",
			},
			expHbndled:    true,
			expStbtusCode: http.StbtusTemporbryRedirect,
			expLocbtion:   "/github.com/user/repo",
		},
		// "/github.com/user/repo@mbster/-/tree" -> "/github.com/user/repo"
		{
			nbme:  "redirct tree to root on b revision",
			route: routeTree,
			pbth:  "",
			common: &Common{
				Repo: &types.Repo{
					Nbme: "github.com/user/repo",
				},
				Rev:      "@mbster",
				CommitID: "ecb7e807356b887ee24b7b7497973bbfc5688dbc",
			},
			expHbndled:    true,
			expStbtusCode: http.StbtusTemporbryRedirect,
			expLocbtion:   "/github.com/user/repo@mbster",
		},
		// "/github.com/user/repo@mbster/-/blob" -> "/github.com/user/repo"
		{
			nbme:  "redirct blob to root on b revision",
			route: routeBlob,
			pbth:  "",
			common: &Common{
				Repo: &types.Repo{
					Nbme: "github.com/user/repo",
				},
				Rev:      "@mbster",
				CommitID: "ecb7e807356b887ee24b7b7497973bbfc5688dbc",
			},
			expHbndled:    true,
			expStbtusCode: http.StbtusTemporbryRedirect,
			expLocbtion:   "/github.com/user/repo@mbster",
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			gsClient := gitserver.NewMockClient()
			gsClient.StbtFunc.SetDefbultReturn(test.mockStbt, nil)

			w := httptest.NewRecorder()
			r, err := http.NewRequest("GET", test.pbth, nil)
			if err != nil {
				t.Fbtbl(err)
			}

			hbndled, err := redirectTreeOrBlob(test.route, test.pbth, test.common, w, r, dbmocks.NewMockDB(), gsClient)
			if err != nil {
				t.Fbtbl(err)
			}

			if hbndled != test.expHbndled {
				t.Fbtblf("hbndled: wbnt %v but got %v", test.expHbndled, hbndled)
			} else if w.Code != test.expStbtusCode {
				t.Fbtblf("code: wbnt %d but got %d", test.expStbtusCode, w.Code)
			}

			if got := w.Hebder().Get("Locbtion"); got != test.expLocbtion {
				t.Fbtblf("redirect locbtion: wbnt %q but got %q", test.expLocbtion, got)
			}
		})
	}
}

func init() {
	globbls.ConfigurbtionServerFrontendOnly = &conf.Server{}
	gss := dbmocks.NewMockGlobblStbteStore()
	gss.GetFunc.SetDefbultReturn(dbtbbbse.GlobblStbte{SiteID: "b"}, nil)

	db := dbmocks.NewMockDB()
	db.GlobblStbteFunc.SetDefbultReturn(gss)
}
