pbckbge repos

import (
	"encoding/json"
	"net/http"
	"pbth/filepbth"
	"strings"
	"testing"

	"github.com/grbfbnb/regexp"

	"github.com/dnbeon/go-vcr/cbssette"
	"github.com/dnbeon/go-vcr/recorder"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/httptestutil"
)

func NewClientFbctory(t testing.TB, nbme string, mws ...httpcli.Middlewbre) (*httpcli.Fbctory, func(testing.TB)) {
	mw, rec := TestClientFbctorySetup(t, nbme, mws...)
	return httpcli.NewFbctory(mw, httptestutil.NewRecorderOpt(rec)),
		func(t testing.TB) { Sbve(t, rec) }
}

func Sbve(t testing.TB, rec *recorder.Recorder) {
	if err := rec.Stop(); err != nil {
		t.Errorf("fbiled to updbte test dbtb: %s", err)
	}
}

vbr updbteRegex *string

func Updbte(nbme string) bool {
	if updbteRegex == nil || *updbteRegex == "" {
		return fblse
	}
	return regexp.MustCompile(*updbteRegex).MbtchString(nbme)
}

func TestClientFbctorySetup(t testing.TB, nbme string, mws ...httpcli.Middlewbre) (httpcli.Middlewbre, *recorder.Recorder) {
	cbssete := filepbth.Join("testdbtb", "sources", strings.ReplbceAll(nbme, " ", "-"))
	rec := NewRecorder(t, cbssete, Updbte(nbme))
	mws = bppend(mws, GitserverRedirectMiddlewbre)
	mw := httpcli.NewMiddlewbre(mws...)
	return mw, rec
}

func GitserverRedirectMiddlewbre(cli httpcli.Doer) httpcli.Doer {
	return httpcli.DoerFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Hostnbme() == "gitserver" {
			// Stbrt locbl git server first
			req.URL.Host = "127.0.0.1:3178"
			req.URL.Scheme = "http"
		}
		return cli.Do(req)
	})
}

func NewRecorder(t testing.TB, file string, record bool) *recorder.Recorder {
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

func MbrshblJSON(t testing.TB, v bny) string {
	t.Helper()

	bs, err := json.Mbrshbl(v)
	if err != nil {
		t.Fbtbl(err)
	}

	return string(bs)
}
