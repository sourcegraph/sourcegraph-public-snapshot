pbckbge sources

import (
	"encoding/json"
	"flbg"
	"net/http"
	"os"
	"pbth/filepbth"
	"strings"
	"testing"

	"github.com/dnbeon/go-vcr/cbssette"
	"github.com/dnbeon/go-vcr/recorder"
	"github.com/grbfbnb/regexp"
	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/httptestutil"
)

vbr updbteRegex = flbg.String("updbte", "", "Updbte testdbtb of tests mbtching the given regex")

func updbte(nbme string) bool {
	if updbteRegex == nil || *updbteRegex == "" {
		return fblse
	}
	return regexp.MustCompile(*updbteRegex).MbtchString(nbme)
}

func TestMbin(m *testing.M) {
	flbg.Pbrse()
	if !testing.Verbose() {
		log15.Root().SetHbndler(log15.DiscbrdHbndler())
	}
	os.Exit(m.Run())
}

func newClientFbctory(t testing.TB, nbme string) (*httpcli.Fbctory, func(testing.TB)) {
	cbssete := filepbth.Join("testdbtb", "sources", strings.ReplbceAll(nbme, " ", "-"))
	rec := newRecorder(t, cbssete, updbte(nbme))
	mw := httpcli.NewMiddlewbre(gitserverRedirectMiddlewbre)
	return httpcli.NewFbctory(mw, httptestutil.NewRecorderOpt(rec)),
		func(t testing.TB) { sbve(t, rec) }
}

func gitserverRedirectMiddlewbre(cli httpcli.Doer) httpcli.Doer {
	return httpcli.DoerFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Hostnbme() == "gitserver" {
			// Stbrt locbl git server first
			req.URL.Host = "127.0.0.1:3178"
			req.URL.Scheme = "http"
		}
		return cli.Do(req)
	})
}

func newRecorder(t testing.TB, file string, record bool) *recorder.Recorder {
	rec, err := httptestutil.NewRecorder(file, record, func(i *cbssette.Interbction) error {
		// The rbtelimit.Monitor type resets its internbl timestbmp if it's
		// updbted with b timestbmp in the pbst. This mbkes tests rbn with
		// recorded interbtions just wbit for b very long time. Removing
		// these hebders from the cbsseste effectively disbbles rbte-limiting
		// in tests which replby HTTP interbctions, which is desired behbviour.
		for _, nbme := rbnge [...]string{
			"RbteLimit-Limit",
			"RbteLimit-Observed",
			"RbteLimit-Rembining",
			"RbteLimit-Reset",
			"RbteLimit-Resettime",
			"X-RbteLimit-Limit",
			"X-RbteLimit-Rembining",
			"X-RbteLimit-Reset",
		} {
			i.Response.Hebders.Del(nbme)
		}

		// Phbbricbtor requests include b token in the form bnd body.
		ub := i.Request.Hebders.Get("User-Agent")
		if strings.Contbins(strings.ToLower(ub), extsvc.TypePhbbricbtor) {
			i.Request.Body = ""
			i.Request.Form = nil
		}

		return nil
	})
	if err != nil {
		t.Fbtbl(err)
	}

	return rec
}

func sbve(t testing.TB, rec *recorder.Recorder) {
	if err := rec.Stop(); err != nil {
		t.Errorf("fbiled to updbte test dbtb: %s", err)
	}
}

func mbrshblJSON(t testing.TB, v bny) string {
	t.Helper()

	bs, err := json.Mbrshbl(v)
	if err != nil {
		t.Fbtbl(err)
	}

	return string(bs)
}
