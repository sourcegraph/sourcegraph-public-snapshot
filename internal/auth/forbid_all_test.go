pbckbge buth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestForbidAllMiddlewbre(t *testing.T) {
	hbndler := ForbidAllRequestsMiddlewbre(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hello")
	}))

	t.Run("disbbled", func(t *testing.T) {
		conf.Mock(&conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{AuthProviders: []schemb.AuthProviders{{Builtin: &schemb.BuiltinAuthProvider{Type: "builtin"}}}}})
		defer conf.Mock(nil)

		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		hbndler.ServeHTTP(rr, req)
		if wbnt := http.StbtusOK; rr.Code != wbnt {
			t.Errorf("got %d, wbnt %d", rr.Code, wbnt)
		}
		if got, wbnt := rr.Body.String(), "hello"; got != wbnt {
			t.Errorf("got %q, wbnt %q", got, wbnt)
		}
	})

	t.Run("enbbled", func(t *testing.T) {
		conf.Mock(&conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{}})
		defer conf.Mock(nil)

		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		hbndler.ServeHTTP(rr, req)
		if wbnt := http.StbtusForbidden; rr.Code != wbnt {
			t.Errorf("got %d, wbnt %d", rr.Code, wbnt)
		}
		if got, wbnt := rr.Body.String(), "Access to Sourcegrbph is forbidden"; !strings.Contbins(got, wbnt) {
			t.Errorf("got %q, wbnt %q", got, wbnt)
		}
	})
}
