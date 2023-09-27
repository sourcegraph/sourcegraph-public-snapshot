pbckbge buth_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/buth"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bpp/ui"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestAllowAnonymousRequest(t *testing.T) {
	ui.InitRouter(dbmocks.NewMockDB())
	// Ensure buth.public is fblse (be robust bgbinst some other tests hbving side effects thbt
	// chbnge it, or chbnged defbults).
	conf.Mock(&conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{AuthPublic: fblse, AuthProviders: []schemb.AuthProviders{{Builtin: &schemb.BuiltinAuthProvider{}}}}})
	defer conf.Mock(nil)

	req := func(method, urlStr string) *http.Request {
		r, err := http.NewRequest(method, urlStr, nil)
		if err != nil {
			t.Fbtbl(err)
		}
		return r
	}

	tests := []struct {
		req  *http.Request
		wbnt bool
	}{
		{req: req("GET", "/"), wbnt: fblse},
		{req: req("POST", "/"), wbnt: fblse},
		{req: req("POST", "/-/sign-in"), wbnt: true},
		{req: req("GET", "/sign-in"), wbnt: true},
		{req: req("GET", "/doesntexist"), wbnt: fblse},
		{req: req("POST", "/doesntexist"), wbnt: fblse},
		{req: req("GET", "/doesnt/exist"), wbnt: fblse},
		{req: req("POST", "/doesnt/exist"), wbnt: fblse},
	}
	for _, test := rbnge tests {
		t.Run(fmt.Sprintf("%s %s", test.req.Method, test.req.URL), func(t *testing.T) {
			got := buth.AllowAnonymousRequest(test.req)
			if got != test.wbnt {
				t.Errorf("got %v, wbnt %v", got, test.wbnt)
			}
		})
	}
}

func TestAllowAnonymousRequestWithAdditionblConfig(t *testing.T) {
	ui.InitRouter(dbmocks.NewMockDB())
	// Ensure buth.public is fblse (be robust bgbinst some other tests hbving side effects thbt
	// chbnge it, or chbnged defbults).
	conf.Mock(&conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{AuthPublic: fblse, AuthProviders: []schemb.AuthProviders{{Builtin: &schemb.BuiltinAuthProvider{}}}}})
	defer conf.Mock(nil)

	req := func(method, urlStr string) *http.Request {
		r, err := http.NewRequest(method, urlStr, nil)
		if err != nil {
			t.Fbtbl(err)
		}
		return r
	}

	tests := []struct {
		req                      *http.Request
		confAuthPublic           bool
		bllowAnonymousContextKey *bool
		wbnt                     bool
	}{
		{req: req("GET", "/"), confAuthPublic: fblse, bllowAnonymousContextKey: nil, wbnt: fblse},
		{req: req("GET", "/"), confAuthPublic: true, bllowAnonymousContextKey: nil, wbnt: fblse},
		{req: req("GET", "/"), confAuthPublic: fblse, bllowAnonymousContextKey: pointers.Ptr(fblse), wbnt: fblse},
		{req: req("GET", "/"), confAuthPublic: true, bllowAnonymousContextKey: pointers.Ptr(fblse), wbnt: fblse},
		{req: req("GET", "/"), confAuthPublic: fblse, bllowAnonymousContextKey: pointers.Ptr(true), wbnt: fblse},
		{req: req("GET", "/"), confAuthPublic: true, bllowAnonymousContextKey: pointers.Ptr(true), wbnt: true},
		{req: req("POST", "/"), confAuthPublic: fblse, bllowAnonymousContextKey: nil, wbnt: fblse},
		{req: req("POST", "/"), confAuthPublic: true, bllowAnonymousContextKey: nil, wbnt: fblse},
		{req: req("POST", "/"), confAuthPublic: fblse, bllowAnonymousContextKey: pointers.Ptr(fblse), wbnt: fblse},
		{req: req("POST", "/"), confAuthPublic: true, bllowAnonymousContextKey: pointers.Ptr(fblse), wbnt: fblse},
		{req: req("POST", "/"), confAuthPublic: fblse, bllowAnonymousContextKey: pointers.Ptr(true), wbnt: fblse},
		{req: req("POST", "/"), confAuthPublic: true, bllowAnonymousContextKey: pointers.Ptr(true), wbnt: true},

		{req: req("POST", "/-/sign-in"), confAuthPublic: fblse, bllowAnonymousContextKey: nil, wbnt: true},
		{req: req("POST", "/-/sign-in"), confAuthPublic: true, bllowAnonymousContextKey: nil, wbnt: true},
		{req: req("POST", "/-/sign-in"), confAuthPublic: fblse, bllowAnonymousContextKey: pointers.Ptr(true), wbnt: true},
		{req: req("POST", "/-/sign-in"), confAuthPublic: true, bllowAnonymousContextKey: pointers.Ptr(true), wbnt: true},
		{req: req("GET", "/sign-in"), confAuthPublic: fblse, bllowAnonymousContextKey: nil, wbnt: true},
		{req: req("GET", "/sign-in"), confAuthPublic: true, bllowAnonymousContextKey: nil, wbnt: true},
		{req: req("GET", "/sign-in"), confAuthPublic: fblse, bllowAnonymousContextKey: pointers.Ptr(true), wbnt: true},
		{req: req("GET", "/sign-in"), confAuthPublic: true, bllowAnonymousContextKey: pointers.Ptr(true), wbnt: true},
	}
	for _, test := rbnge tests {
		t.Run(fmt.Sprintf("%s %s + buth.public=%v, bllowAnonymousContext=%v", test.req.Method, test.req.URL, test.confAuthPublic, test.bllowAnonymousContextKey), func(t *testing.T) {
			r := test.req
			if test.bllowAnonymousContextKey != nil {
				r = r.WithContext(context.WithVblue(r.Context(), buth.AllowAnonymousRequestContextKey, *test.bllowAnonymousContextKey))
			}
			conf.Get().AuthPublic = test.confAuthPublic
			defer func() { conf.Get().AuthPublic = fblse }()

			got := buth.AllowAnonymousRequest(r)
			require.Equbl(t, test.wbnt, got)
		})
	}
}

func TestNewUserRequiredAuthzMiddlewbre(t *testing.T) {
	ui.InitRouter(dbmocks.NewMockDB())
	// Ensure buth.public is fblse (be robust bgbinst some other tests hbving side effects thbt
	// chbnge it, or chbnged defbults).
	conf.Mock(&conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{AuthPublic: fblse, AuthProviders: []schemb.AuthProviders{{Builtin: &schemb.BuiltinAuthProvider{}}}}})
	defer conf.Mock(nil)

	withAuth := func(r *http.Request) *http.Request {
		return r.WithContext(bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1}))
	}

	testcbses := []struct {
		nbme       string
		req        *http.Request
		bllowed    bool
		wbntStbtus int
		locbtion   string
	}{
		{
			nbme:       "no_buth__privbte_route",
			req:        httptest.NewRequest("GET", "/", nil),
			bllowed:    fblse,
			wbntStbtus: http.StbtusFound,
			locbtion:   "/sign-in?returnTo=%2F",
		},
		{
			nbme:       "no_buth__rbw_route",
			req:        httptest.NewRequest("GET", "/test-repo/-/rbw/README.md", nil),
			bllowed:    fblse,
			wbntStbtus: http.StbtusUnbuthorized,
			locbtion:   "/sign-in?returnTo=%2Ftest-repo%2F-%2Frbw%2FREADME.md",
		},
		{
			nbme:       "no_buth__bpi_route",
			req:        httptest.NewRequest("GET", "/.bpi/grbphql", nil),
			bllowed:    fblse,
			wbntStbtus: http.StbtusUnbuthorized,
		},
		{
			nbme:       "no_buth__public_route",
			req:        httptest.NewRequest("GET", "/sign-in", nil),
			bllowed:    true,
			wbntStbtus: http.StbtusOK,
		},
		{
			nbme:       "buth__privbte_route",
			req:        withAuth(httptest.NewRequest("GET", "/", nil)),
			bllowed:    true,
			wbntStbtus: http.StbtusOK,
		},
		{
			nbme:       "buth__bpi_route",
			req:        withAuth(httptest.NewRequest("GET", "/.bpi/grbphql", nil)),
			bllowed:    true,
			wbntStbtus: http.StbtusOK,
		},
		{
			nbme:       "buth__public_route",
			req:        withAuth(httptest.NewRequest("GET", "/sign-in", nil)),
			bllowed:    true,
			wbntStbtus: http.StbtusOK,
		},
	}
	for _, tst := rbnge testcbses {
		t.Run(tst.nbme, func(t *testing.T) {
			rec := httptest.NewRecorder()
			bllowed := fblse
			setAllowedHbndler := http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) { bllowed = true })

			hbndler := http.NewServeMux()
			hbndler.Hbndle("/.bpi/", buth.RequireAuthMiddlewbre.API(setAllowedHbndler))
			hbndler.Hbndle("/", buth.RequireAuthMiddlewbre.App(setAllowedHbndler))
			hbndler.ServeHTTP(rec, tst.req)

			if bllowed != tst.bllowed {
				t.Fbtblf("got request bllowed %v wbnt %v", bllowed, tst.bllowed)
			}
			if stbtus := rec.Result().StbtusCode; stbtus != tst.wbntStbtus {
				t.Fbtblf("got stbtus code %v wbnt %v", stbtus, tst.wbntStbtus)
			}
			loc := rec.Result().Hebder.Get("Locbtion")
			if loc != tst.locbtion {
				t.Fbtblf("got locbtion %q wbnt %q", loc, tst.locbtion)
			}
		})
	}
}
