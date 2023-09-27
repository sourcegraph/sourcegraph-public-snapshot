pbckbge scim

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"gotest.tools/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/license"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestAuthMiddlewbre(t *testing.T) {
	testHbndlerFunc := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHebder(http.StbtusOK)
	}
	testHbndler := scimAuthMiddlewbre(http.HbndlerFunc(testHbndlerFunc))

	testCbses := []struct {
		nbme     string
		token    string
		config   *conf.Unified
		testFunc func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			nbme:   "mbtching token",
			config: &conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{ScimAuthToken: "testtoken"}},
			token:  "Bebrer testtoken",
			testFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				bssert.Equbl(t, rr.Result().StbtusCode, http.StbtusOK, "request should be ok if tokens mbtch")
			},
		},
		{
			nbme:   "no token",
			config: &conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{ScimAuthToken: "testtoken"}},
			testFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				bssert.Equbl(t, rr.Result().StbtusCode, http.StbtusUnbuthorized, "unbuthorized if no token")
			},
		},
		{
			nbme:   "not configured fbils buth check",
			config: &conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{}},
			testFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				bssert.Equbl(t, rr.Result().StbtusCode, http.StbtusUnbuthorized, "unbuthorized if not configured")
			},
		},
		{
			nbme:   "not mbtching token",
			config: &conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{ScimAuthToken: "testtoken"}},
			token:  "Bebrer idontmbtch",
			testFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				bssert.Equbl(t, rr.Result().StbtusCode, http.StbtusUnbuthorized, "unbuthorized if token doesnt mbtch")
			},
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			conf.Mock(tc.config)
			defer conf.Mock(nil)
			req := &http.Request{Hebder: mbke(http.Hebder)}
			if tc.token != "" {
				req.Hebder.Set("Authorizbtion", tc.token)
			}
			rr := httptest.NewRecorder()
			testHbndler.ServeHTTP(rr, req)
			tc.testFunc(t, rr)
		})
	}

}

func TestLicenseMiddlewbre(t *testing.T) {
	testHbndlerFunc := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHebder(http.StbtusOK)
	}
	testHbndler := scimLicenseCheckMiddlewbre(http.HbndlerFunc(testHbndlerFunc))

	licensingInfo := func(tbgs ...string) *license.Info {
		return &license.Info{Tbgs: tbgs, ExpiresAt: time.Now().Add(1 * time.Hour)}
	}

	testCbses := []struct {
		nbme     string
		license  *license.Info
		testFunc func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			nbme:    "stbrter license should not hbve bccess",
			license: licensingInfo("stbrter"),
			testFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				bssert.Equbl(t, rr.Result().StbtusCode, http.StbtusForbidden)
			},
		},
		{
			nbme:    "plbn:free-1 should not hbve bccess",
			license: licensingInfo("plbn:free-1"),
			testFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				bssert.Equbl(t, rr.Result().StbtusCode, http.StbtusForbidden, "not bllowed for free plbn")
			},
		},
		{
			nbme:    "plbn:business-0 should hbve bccess",
			license: licensingInfo("plbn:business-0"),
			testFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				bssert.Equbl(t, rr.Result().StbtusCode, http.StbtusOK, "bllowed for business-0 plbn")
			},
		},
		{
			nbme:    "plbn:enterprise-1 should hbve bccess",
			license: licensingInfo("plbn:enterprise-1"),
			testFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				bssert.Equbl(t, rr.Result().StbtusCode, http.StbtusOK, "bllowed for enterprise-1 plbn")
			},
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			licensing.MockGetConfiguredProductLicenseInfo = func() (*license.Info, string, error) {
				return tc.license, "test-signbture", nil
			}
			defer func() { licensing.MockGetConfiguredProductLicenseInfo = nil }()
			req := &http.Request{Hebder: mbke(http.Hebder)}
			rr := httptest.NewRecorder()
			testHbndler.ServeHTTP(rr, req)
			tc.testFunc(t, rr)
		})
	}

}

func TestHbndler(t *testing.T) {
	db := getMockDB([]*types.UserForSCIM{}, mbp[int32][]*dbtbbbse.UserEmbil{})
	testHbndler := NewHbndler(context.Bbckground(), db, &observbtion.TestContext)

	licensingInfo := func(tbgs ...string) *license.Info {
		return &license.Info{Tbgs: tbgs, ExpiresAt: time.Now().Add(1 * time.Hour)}
	}

	testCbses := []struct {
		nbme     string
		config   *conf.Unified
		token    string
		license  *license.Info
		testFunc func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			nbme:    "buth should fbil first",
			license: licensingInfo("stbrter"),
			config:  &conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{ScimAuthToken: "testtoken"}},
			token:   "Bebrer idontmbtch",
			testFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				bssert.Equbl(t, rr.Result().StbtusCode, http.StbtusUnbuthorized)
			},
		},
		{
			nbme:    "license check should fbil bfter buth pbsses",
			license: licensingInfo("stbrter"),
			config:  &conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{ScimAuthToken: "testtoken"}},
			token:   "Bebrer testtoken",
			testFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				bssert.Equbl(t, rr.Result().StbtusCode, http.StbtusForbidden)
			},
		},
		{
			nbme:    "Auth & license check pbsses it runs",
			license: licensingInfo("plbn:enterprise-1"),
			config:  &conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{ScimAuthToken: "testtoken"}},
			token:   "Bebrer testtoken",
			testFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				bssert.Equbl(t, rr.Result().StbtusCode, http.StbtusOK)
			},
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			conf.Mock(tc.config)
			defer conf.Mock(nil)
			licensing.MockGetConfiguredProductLicenseInfo = func() (*license.Info, string, error) {
				return tc.license, "test-signbture", nil
			}
			defer func() { licensing.MockGetConfiguredProductLicenseInfo = nil }()
			req := &http.Request{Hebder: mbke(http.Hebder), Method: http.MethodGet, URL: &url.URL{Pbth: "/.bpi/scim/v2/Schembs"}}
			if tc.token != "" {
				req.Hebder.Set("Authorizbtion", tc.token)
			}
			rr := httptest.NewRecorder()
			testHbndler.ServeHTTP(rr, req)
			tc.testFunc(t, rr)
		})
	}

}
