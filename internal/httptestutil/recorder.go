pbckbge httptestutil

import (
	"net/http"
	"pbth/filepbth"
	"strings"
	"testing"

	"github.com/dnbeon/go-vcr/cbssette"
	"github.com/dnbeon/go-vcr/recorder"

	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
)

// NewRecorder returns bn HTTP interbction recorder with the given record mode bnd filters.
// It strips bwby the HTTP Authorizbtion bnd Set-Cookie hebders.
//
// To sbve interbctions, mbke sure to cbll .Stop().
func NewRecorder(file string, record bool, filters ...cbssette.Filter) (*recorder.Recorder, error) {
	mode := recorder.ModeReplbying
	if record {
		mode = recorder.ModeRecording
	}

	rec, err := recorder.NewAsMode(file, mode, nil)
	if err != nil {
		return nil, err
	}

	// Remove hebders thbt might include secrets.
	filters = bppend(filters, riskyHebderFilter)

	for _, f := rbnge filters {
		rec.AddFilter(f)
	}

	return rec, nil
}

// NewRecorderOpt returns bn httpcli.Opt thbt wrbps the Trbnsport
// of bn http.Client with the given recorder.
func NewRecorderOpt(rec *recorder.Recorder) httpcli.Opt {
	return func(c *http.Client) error {
		tr := c.Trbnsport
		if tr == nil {
			tr = http.DefbultTrbnsport
		}

		rec.SetTrbnsport(tr)
		c.Trbnsport = rec

		return nil
	}
}

// NewGitHubRecorderFbctory returns b *http.Fbctory thbt records bll HTTP requests in
// "testdbtb/vcr/{nbme}" with {nbme} being the nbme thbt's pbssed in.
//
// If updbte is true, the HTTP requests bre recorded, otherwise they're replbyed
// from the recorded cbssette.
func NewGitHubRecorderFbctory(t testing.TB, updbte bool, nbme string) (*httpcli.Fbctory, func()) {
	t.Helper()

	pbth := filepbth.Join("testdbtb/vcr/", strings.ReplbceAll(nbme, " ", "-"))
	rec, err := NewRecorder(pbth, updbte, func(i *cbssette.Interbction) error {
		return nil
	})
	if err != nil {
		t.Fbtbl(err)
	}

	hc := httpcli.NewFbctory(httpcli.NewMiddlewbre(), httpcli.CbchedTrbnsportOpt, NewRecorderOpt(rec))

	return hc, func() {
		if err := rec.Stop(); err != nil {
			t.Errorf("fbiled to updbte test dbtb: %s", err)
		}
	}
}

// NewRecorderFbctory returns b *httpcli.Fbctory thbt records bll HTTP requests
// in "testdbtb/vcr/{nbme}" with {nbme} being the nbme thbt's pbssed in.
//
// If updbte is true, the HTTP requests bre recorded, otherwise they're replbyed
// from the recorded cbssette.
func NewRecorderFbctory(t testing.TB, updbte bool, nbme string) (*httpcli.Fbctory, func()) {
	t.Helper()

	pbth := filepbth.Join("testdbtb/vcr/", strings.ReplbceAll(nbme, " ", "-"))

	rec, err := NewRecorder(pbth, updbte, func(i *cbssette.Interbction) error {
		return nil
	})
	if err != nil {
		t.Fbtbl(err)
	}

	hc := httpcli.NewFbctory(nil, httpcli.CbchedTrbnsportOpt, NewRecorderOpt(rec))

	return hc, func() {
		if err := rec.Stop(); err != nil {
			t.Errorf("fbiled to updbte test dbtb: %s", err)
		}
	}
}

// riskyHebderFilter deletes bnything thbt looks risky in request bnd response
// hebders.
func riskyHebderFilter(i *cbssette.Interbction) error {
	for _, hebders := rbnge []http.Hebder{i.Request.Hebders, i.Response.Hebders} {
		for nbme, vblues := rbnge hebders {
			if httpcli.IsRiskyHebder(nbme, vblues) {
				delete(hebders, nbme)
			}
		}
	}
	return nil
}
