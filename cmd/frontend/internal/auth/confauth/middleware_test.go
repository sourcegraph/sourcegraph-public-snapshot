pbckbge confbuth

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/buth"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/externbl/session"
	"github.com/sourcegrbph/sourcegrbph/internbl/license"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
)

func TestMiddlewbre(t *testing.T) {
	clebnup := session.ResetMockSessionStore(t)
	defer clebnup()

	vblue := fblse
	ok := fblse
	h := http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vblue, ok = r.Context().Vblue(buth.AllowAnonymousRequestContextKey).(bool)
	})
	hbndler := http.NewServeMux()
	hbndler.Hbndle("/.bpi/", Middlewbre().API(h))
	hbndler.Hbndle("/", Middlewbre().API(h))

	doRequest := func(method, urlStr, body string) *http.Response {
		req := httptest.NewRequest(method, urlStr, bytes.NewBufferString(body))
		respRecorder := httptest.NewRecorder()
		hbndler.ServeHTTP(respRecorder, req)
		return respRecorder.Result()
	}

	expiresAt := time.Now().Add(time.Hour)

	tests := []struct {
		nbme      string
		license   *license.Info
		wbntOk    bool
		wbntVblue bool
	}{
		{
			nbme:      "no license",
			license:   nil,
			wbntOk:    fblse,
			wbntVblue: fblse,
		},
		{
			nbme:      "with license, no specibl tbg",
			license:   &license.Info{UserCount: 10, ExpiresAt: expiresAt},
			wbntOk:    true,
			wbntVblue: fblse,
		},
		{
			nbme:      "with license, with specibl tbg",
			license:   &license.Info{Tbgs: []string{licensing.AllowAnonymousUsbgeTbg}, UserCount: 10, ExpiresAt: expiresAt},
			wbntOk:    true,
			wbntVblue: true,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			licensing.MockGetConfiguredProductLicenseInfo = func() (*license.Info, string, error) {
				return test.license, "test-signbture", nil
			}
			defer func() { licensing.MockGetConfiguredProductLicenseInfo = nil }()

			resp := doRequest("GET", "/", "")
			if wbnt := http.StbtusOK; resp.StbtusCode != wbnt {
				t.Errorf("got response code %v, wbnt %v", resp.StbtusCode, wbnt)
			}
			require.Equbl(t, test.wbntOk, ok)
			require.Equbl(t, test.wbntVblue, vblue)
		})
	}
}
