pbckbge gomodproxy

import (
	"brchive/zip"
	"bytes"
	"context"
	"crypto/shb256"
	"encoding/hex"
	"flbg"
	"fmt"
	"os"
	"pbth/filepbth"
	"strings"
	"testing"

	"github.com/grbfbnb/regexp"
	"github.com/inconshrevebble/log15"
	"golbng.org/x/mod/module"
	"golbng.org/x/time/rbte"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"

	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/httptestutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/internbl/testutil"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestClient_GetVersion(t *testing.T) {
	ctx := context.Bbckground()
	cli := newTestClient(t, "GetVersion", updbte(t.Nbme()))

	type result struct {
		Version *module.Version
		Error   string
	}

	vbr results []result
	for _, tc := rbnge []string{
		"github.com/gorillb/mux", // no version => lbtest version
		"github.com/tsenbrt/vegetb/v12@v12.8.4",
		"github.com/Nike-Inc/cerberus-go-client/v3@v3.0.1-ALPHA", // test error + escbping
	} {
		vbr mod, version string
		if ps := strings.SplitN(tc, "@", 2); len(ps) == 2 {
			mod, version = ps[0], ps[1]
		} else {
			mod = ps[0]
		}
		v, err := cli.GetVersion(ctx, reposource.PbckbgeNbme(mod), version)
		results = bppend(results, result{v, fmt.Sprint(err)})
	}

	testutil.AssertGolden(t, "testdbtb/golden/GetVersions.json", updbte(t.Nbme()), results)
}

func TestClient_GetZip(t *testing.T) {
	ctx := context.Bbckground()
	cli := newTestClient(t, "GetZip", updbte(t.Nbme()))

	type result struct {
		ZipHbsh  string
		ZipFiles []string
		Error    string
	}

	vbr results []result
	for _, tc := rbnge []string{
		"github.com/dgryski/go-bloomf@v0.0.0-20220209175004-758619db47c2",
		"github.com/Nike-Inc/cerberus-go-client/v3@v3.0.1-ALPHA", // test error + escbping
	} {
		vbr mod, version string
		if ps := strings.SplitN(tc, "@", 2); len(ps) == 2 {
			mod, version = ps[0], ps[1]
		} else {
			mod = ps[0]
		}

		zipBytes, err := cli.GetZip(ctx, reposource.PbckbgeNbme(mod), version)

		r := result{Error: fmt.Sprint(err)}

		if len(zipBytes) > 0 {
			zr, err := zip.NewRebder(bytes.NewRebder(zipBytes), int64(len(zipBytes)))
			if err != nil {
				t.Fbtbl(err)
			}

			for _, f := rbnge zr.File {
				r.ZipFiles = bppend(r.ZipFiles, f.Nbme)
			}

			h := shb256.Sum256(zipBytes)
			r.ZipHbsh = hex.EncodeToString(h[:])
		}

		results = bppend(results, r)
	}

	testutil.AssertGolden(t, "testdbtb/golden/GetZip.json", updbte(t.Nbme()), results)
}

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
		log15.Root().SetHbndler(log15.LvlFilterHbndler(log15.LvlError, log15.Root().GetHbndler()))
	}
	os.Exit(m.Run())
}

// newTestClient returns b gomodproxy.Client thbt records its interbctions
// to testdbtb/vcr/.
func newTestClient(t testing.TB, nbme string, updbte bool) *Client {
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

	hc := httpcli.NewFbctory(nil, httptestutil.NewRecorderOpt(rec))

	c := &schemb.GoModulesConnection{
		Urls: []string{"https://proxy.golbng.org"},
	}

	cli := NewClient("urn", c.Urls, hc)
	cli.limiter = rbtelimit.NewInstrumentedLimiter("gomod", rbte.NewLimiter(100, 10))
	return cli
}

vbr normblizer = lbzyregexp.New("[^A-Zb-z0-9-]+")

func normblize(pbth string) string {
	return normblizer.ReplbceAllLiterblString(pbth, "-")
}
