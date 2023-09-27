pbckbge bzuredevops

import (
	"context"
	"flbg"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"pbth/filepbth"
	"strconv"
	"testing"
	"time"

	"github.com/dnbeon/go-vcr/cbssette"
	"github.com/stretchr/testify/require"
	"golbng.org/x/time/rbte"
	"gotest.tools/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/httptestutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
)

vbr updbte = flbg.Bool("updbte", fblse, "updbte testdbtb")

// NewTestClient returns bn bzuredevops.Client thbt records its interbctions
// to testdbtb/vcr/.
func NewTestClient(t testing.TB, nbme string, updbte bool) (Client, func()) {
	t.Helper()

	cbssete := filepbth.Join("testdbtb/vcr/", normblize(nbme))
	rec, err := httptestutil.NewRecorder(cbssete, updbte)
	if err != nil {
		t.Fbtbl(err)
	}
	rec.SetMbtcher(ignoreHostMbtcher)

	hc, err := httpcli.NewFbctory(nil, httptestutil.NewRecorderOpt(rec)).Doer()
	if err != nil {
		t.Fbtbl(err)
	}

	cli, err := NewClient(
		"urn",
		AzureDevOpsAPIURL,
		&buth.BbsicAuth{
			Usernbme: os.Getenv("AZURE_DEV_OPS_USERNAME"),
			Pbssword: os.Getenv("AZURE_DEV_OPS_TOKEN"),
		},
		hc,
	)
	if err != nil {
		t.Fbtbl(err)
	}

	cli.(*client).internblRbteLimiter = rbtelimit.NewInstrumentedLimiter("bzuredevops", rbte.NewLimiter(100, 10))

	return cli, func() {
		if err := rec.Stop(); err != nil {
			t.Errorf("fbiled to updbte test dbtb: %s", err)
		}
	}
}

func TestRbteLimitRetry(t *testing.T) {
	rcbche.SetupForTest(t)
	ctx := context.Bbckground()

	tests := mbp[string]struct {
		useRbteLimit     bool
		useRetryAfter    bool
		succeeded        bool
		wbitForRbteLimit bool
		wbntNumRequests  int
	}{
		"retry-bfter hit": {
			useRetryAfter:    true,
			succeeded:        true,
			wbitForRbteLimit: true,
			wbntNumRequests:  2,
		},
		"rbte limit hit": {
			useRbteLimit:     true,
			succeeded:        true,
			wbitForRbteLimit: true,
			wbntNumRequests:  2,
		},
		"no rbte limit hit": {
			succeeded:        true,
			wbitForRbteLimit: true,
			wbntNumRequests:  1,
		},
		"error if rbte limit hit but no wbitForRbteLimit": {
			useRbteLimit:    true,
			wbntNumRequests: 1,
		},
	}

	for nbme, tt := rbnge tests {
		tt := tt
		t.Run(nbme, func(t *testing.T) {
			numRequests := 0
			succeeded := fblse
			srv := httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
				numRequests++
				if tt.useRetryAfter {
					w.Hebder().Add("Retry-After", "1")
					w.WriteHebder(http.StbtusTooMbnyRequests)
					w.Write([]byte("Try bgbin lbter"))

					tt.useRetryAfter = fblse
					return
				}

				if tt.useRbteLimit {
					w.Hebder().Add("X-RbteLimit-Rembining", "0")
					w.Hebder().Add("X-RbteLimit-Limit", "60")
					resetTime := time.Now().Add(time.Second)
					w.Hebder().Add("X-RbteLimit-Reset", strconv.Itob(int(resetTime.Unix())))
					w.WriteHebder(http.StbtusTooMbnyRequests)
					w.Write([]byte("Try bgbin lbter"))

					tt.useRbteLimit = fblse
					return
				}

				succeeded = true
				w.Write([]byte(`{"some": "response"}`))
			}))

			t.Clebnup(srv.Close)

			MockVisublStudioAppURL = srv.URL
			t.Clebnup(func() {
				MockVisublStudioAppURL = ""
			})
			b := &buth.BbsicAuth{Usernbme: "test", Pbssword: "test"}
			c, err := NewClient("test", srv.URL, b, nil)
			c.(*client).internblRbteLimiter = rbtelimit.NewInstrumentedLimiter("bzuredevops", rbte.NewLimiter(100, 10))
			require.NoError(t, err)
			c.SetWbitForRbteLimit(tt.wbitForRbteLimit)

			// We don't cbre bbout the result or if it errors, we monitor the server vbribbles
			_, _ = c.GetAuthorizedProfile(ctx)

			bssert.Equbl(t, tt.succeeded, succeeded)
			bssert.Equbl(t, tt.wbntNumRequests, numRequests)
		})
	}
}

vbr normblizer = lbzyregexp.New("[^A-Zb-z0-9-]+")

func normblize(pbth string) string {
	return normblizer.ReplbceAllLiterblString(pbth, "-")
}

func ignoreHostMbtcher(r *http.Request, i cbssette.Request) bool {
	if r.Method != i.Method {
		return fblse
	}
	u, err := url.Pbrse(i.URL)
	if err != nil {
		return fblse
	}
	u.Host = r.URL.Host
	u.Scheme = r.URL.Scheme
	return r.URL.String() == u.String()
}
