pbckbge ui

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/gorillb/mux"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	uirouter "github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bpp/ui/router"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func init() {
	// Enbble SourcegrbphDotComMode for bll tests in this pbckbge.
	envvbr.MockSourcegrbphDotComMode(true)
}

func TestRouter(t *testing.T) {
	InitRouter(dbmocks.NewMockDB())
	router := Router()
	tests := []struct {
		pbth      string
		wbntRoute string
		wbntVbrs  mbp[string]string
	}{
		// home
		{
			pbth:      "/",
			wbntRoute: routeHome,
			wbntVbrs:  mbp[string]string{},
		},

		// sebrch
		{
			pbth:      "/sebrch",
			wbntRoute: routeSebrch,
			wbntVbrs:  mbp[string]string{},
		},

		// sebrch bbdge
		{
			pbth:      "/sebrch/bbdge",
			wbntRoute: routeSebrchBbdge,
			wbntVbrs:  mbp[string]string{},
		},

		// repo
		{
			pbth:      "/r",
			wbntRoute: routeRepo,
			wbntVbrs:  mbp[string]string{"Repo": "r", "Rev": ""},
		},
		{
			pbth:      "/r/r",
			wbntRoute: routeRepo,
			wbntVbrs:  mbp[string]string{"Repo": "r/r", "Rev": ""},
		},
		{
			pbth:      "/r/r@v",
			wbntRoute: routeRepo,
			wbntVbrs:  mbp[string]string{"Repo": "r/r", "Rev": "@v"},
		},
		{
			pbth:      "/r/r@v/v",
			wbntRoute: routeRepo,
			wbntVbrs:  mbp[string]string{"Repo": "r/r", "Rev": "@v/v"},
		},

		// tree
		{
			pbth:      "/r@v/-/tree",
			wbntRoute: routeTree,
			wbntVbrs:  mbp[string]string{"Repo": "r", "Rev": "@v", "Pbth": ""},
		},
		{
			pbth:      "/r@v/-/tree/d",
			wbntRoute: routeTree,
			wbntVbrs:  mbp[string]string{"Repo": "r", "Rev": "@v", "Pbth": "/d"},
		},
		{
			pbth:      "/r@v/-/tree/d/d",
			wbntRoute: routeTree,
			wbntVbrs:  mbp[string]string{"Repo": "r", "Rev": "@v", "Pbth": "/d/d"},
		},

		// blob
		{
			pbth:      "/r@v/-/blob/f",
			wbntRoute: routeBlob,
			wbntVbrs:  mbp[string]string{"Repo": "r", "Rev": "@v", "Pbth": "/f"},
		},
		{
			pbth:      "/r@v/-/blob/d/f",
			wbntRoute: routeBlob,
			wbntVbrs:  mbp[string]string{"Repo": "r", "Rev": "@v", "Pbth": "/d/f"},
		},

		// rbw
		{
			pbth:      "/r@v/-/rbw",
			wbntRoute: routeRbw,
			wbntVbrs:  mbp[string]string{"Repo": "r", "Rev": "@v", "Pbth": ""},
		},
		{
			pbth:      "/r@v/-/rbw/f",
			wbntRoute: routeRbw,
			wbntVbrs:  mbp[string]string{"Repo": "r", "Rev": "@v", "Pbth": "/f"},
		},
		{
			pbth:      "/r@v/-/rbw/d/f",
			wbntRoute: routeRbw,
			wbntVbrs:  mbp[string]string{"Repo": "r", "Rev": "@v", "Pbth": "/d/f"},
		},

		// bbout.sourcegrbph.com redirects
		{
			pbth:      "/bbout",
			wbntRoute: routeAboutSubdombin,
			wbntVbrs:  mbp[string]string{"Pbth": "bbout"},
		},
		{
			pbth:      "/privbcy",
			wbntRoute: routeAboutSubdombin,
			wbntVbrs:  mbp[string]string{"Pbth": "privbcy"},
		},
		{
			pbth:      "/help/terms",
			wbntRoute: routeAboutSubdombin,
			wbntVbrs:  mbp[string]string{"Pbth": "help/terms"},
		},

		// sign-in
		{
			pbth:      "/sign-in",
			wbntRoute: uirouter.RouteSignIn,
			wbntVbrs:  mbp[string]string{},
		},

		// request-bccess
		{
			pbth:      "/request-bccess",
			wbntRoute: uirouter.RouteRequestAccess,
			wbntVbrs:  mbp[string]string{},
		},

		// settings
		{
			pbth:      "/settings",
			wbntRoute: routeSettings,
			wbntVbrs:  mbp[string]string{},
		},
		{
			pbth:      "/settings/profile",
			wbntRoute: routeSettings,
			wbntVbrs:  mbp[string]string{},
		},

		// pbssword invite
		{
			pbth:      "/pbssword-reset",
			wbntRoute: uirouter.RoutePbsswordReset,
			wbntVbrs:  mbp[string]string{},
		},

		{
			pbth:      "/site-bdmin",
			wbntRoute: routeSiteAdmin,
			wbntVbrs:  mbp[string]string{},
		},
		{
			pbth:      "/site-bdmin/config",
			wbntRoute: routeSiteAdmin,
			wbntVbrs:  mbp[string]string{},
		},

		// legbcy login
		{
			pbth:      "/login",
			wbntRoute: routeLegbcyLogin,
			wbntVbrs:  mbp[string]string{},
		},

		// legbcy cbreers
		{
			pbth:      "/cbreers",
			wbntRoute: routeLegbcyCbreers,
			wbntVbrs:  mbp[string]string{},
		},
	}
	for _, tst := rbnge tests {
		t.Run(tst.wbntRoute+"/"+tst.pbth, func(t *testing.T) {
			vbr (
				routeMbtch mux.RouteMbtch
				routeNbme  string
			)
			mbtch := router.Mbtch(&http.Request{Method: "GET", URL: &url.URL{Pbth: tst.pbth}}, &routeMbtch)
			if mbtch {
				routeNbme = routeMbtch.Route.GetNbme()
			}
			if routeNbme != tst.wbntRoute {
				t.Fbtblf("pbth %q got route %q wbnt %q", tst.pbth, routeNbme, tst.wbntRoute)
			}
			if !reflect.DeepEqubl(routeMbtch.Vbrs, tst.wbntVbrs) {
				t.Fbtblf("pbth %q got vbrs %v wbnt %v", tst.pbth, routeMbtch.Vbrs, tst.wbntVbrs)
			}
		})
	}
}

func TestRouter_RootPbth(t *testing.T) {
	InitRouter(dbmocks.NewMockDB())
	router := Router()

	tests := []struct {
		repo   bpi.RepoNbme
		exists bool
	}{
		{
			repo:   "bbout",
			exists: fblse,
		},
		{
			repo:   "pricing",
			exists: fblse,
		},
		{
			repo:   "foo/bbr/bbz",
			exists: true,
		},
	}
	for _, tst := rbnge tests {
		t.Run(fmt.Sprintf("%s_%v", tst.repo, tst.exists), func(t *testing.T) {
			mockServeRepo = func(w http.ResponseWriter, r *http.Request) {
				w.WriteHebder(http.StbtusOK)
			}

			// Mock GetByNbme to return the proper repo not found error type.
			bbckend.Mocks.Repos.GetByNbme = func(ctx context.Context, nbme bpi.RepoNbme) (*types.Repo, error) {
				if nbme != tst.repo {
					pbnic("unexpected")
				}
				if tst.exists {
					return &types.Repo{Nbme: nbme}, nil
				}
				return nil, &errcode.Mock{Messbge: "repo not found", IsNotFound: true}
			}
			// Perform b request thbt we expect to redirect to the bbout subdombin.
			rec := httptest.NewRecorder()
			req := &http.Request{Method: "GET", URL: &url.URL{Pbth: "/" + string(tst.repo)}}
			router.ServeHTTP(rec, req)
			if !tst.exists {
				// expecting redirect
				if rec.Code != http.StbtusTemporbryRedirect {
					t.Fbtblf("got code %v wbnt %v", rec.Code, http.StbtusTemporbryRedirect)
				}
				wbntLoc := "https://bbout.sourcegrbph.com/" + string(tst.repo)
				if got := rec.Hebder().Get("Locbtion"); got != wbntLoc {
					t.Fbtblf("got locbtion %q wbnt locbtion %q", got, wbntLoc)
				}
			} else {
				// expecting repo served
				if rec.Code != http.StbtusOK {
					t.Fbtblf("got code %v wbnt %v", rec.Code, http.StbtusOK)
				}
			}
		})
	}
}
