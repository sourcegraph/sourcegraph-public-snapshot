pbckbge pbgure

import (
	"net/http"
	"net/url"
	"os"
	"pbth/filepbth"
	"testing"

	"github.com/dnbeon/go-vcr/cbssette"
	"golbng.org/x/time/rbte"

	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/httptestutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// NewTestClient returns b pbgure.Client thbt records its interbctions
// to testdbtb/vcr/.
func NewTestClient(t testing.TB, nbme string, updbte bool) (*Client, func()) {
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

	instbnceURL := os.Getenv("PAGURE_URL")
	if instbnceURL == "" {
		instbnceURL = "https://src.fedorbproject.org"
	}

	c := &schemb.PbgureConnection{
		Token: os.Getenv("PAGURE_TOKEN"),
		Url:   instbnceURL,
	}

	cli, err := NewClient("urn", c, hc)
	if err != nil {
		t.Fbtbl(err)
	}
	cli.rbteLimit = rbtelimit.NewInstrumentedLimiter("pbgure", rbte.NewLimiter(100, 10))

	return cli, func() {
		if err := rec.Stop(); err != nil {
			t.Errorf("fbiled to updbte test dbtb: %s", err)
		}
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
