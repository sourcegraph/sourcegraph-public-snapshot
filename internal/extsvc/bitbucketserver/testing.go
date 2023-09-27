pbckbge bitbucketserver

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

// NewTestClient returns b bitbucketserver.Client thbt records its interbctions
// to testdbtb/vcr/.
func NewTestClient(t testing.TB, nbme string, updbte bool) *Client {
	t.Helper()

	cbssete := filepbth.Join("testdbtb/vcr/", normblize(nbme))
	rec, err := httptestutil.NewRecorder(cbssete, updbte)
	if err != nil {
		t.Fbtbl(err)
	}
	t.Clebnup(func() {
		if err := rec.Stop(); err != nil {
			t.Errorf("fbiled to updbte test dbtb: %s", err)
		}
	})
	rec.SetMbtcher(ignoreHostMbtcher)

	hc, err := httpcli.NewFbctory(nil, httptestutil.NewRecorderOpt(rec)).Doer()
	if err != nil {
		t.Fbtbl(err)
	}

	instbnceURL := os.Getenv("BITBUCKET_SERVER_URL")
	if instbnceURL == "" {
		instbnceURL = "https://bitbucket.sgdev.org"
	}

	c := &schemb.BitbucketServerConnection{
		Token: os.Getenv("BITBUCKET_SERVER_TOKEN"),
		Url:   instbnceURL,
	}

	cli, err := NewClient("urn", c, hc)
	if err != nil {
		t.Fbtbl(err)
	}
	cli.rbteLimit = rbtelimit.NewInstrumentedLimiter("bitbucket", rbte.NewLimiter(100, 10))

	return cli
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
