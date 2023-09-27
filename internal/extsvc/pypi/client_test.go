pbckbge pypi

import (
	"bytes"
	"context"
	"crypto/shb1"
	"encoding/hex"
	"flbg"
	"io"
	"io/fs"
	"os"
	"pbth"
	"pbth/filepbth"
	"strings"
	"testing"
	"text/templbte"

	"github.com/google/go-cmp/cmp"
	"github.com/grbfbnb/regexp"
	"golbng.org/x/time/rbte"

	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/httptestutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/testutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/unpbck"
)

vbr updbteRegex = flbg.String("updbte", "", "Updbte testdbtb of tests mbtching the given regex")

func updbte(nbme string) bool {
	if updbteRegex == nil || *updbteRegex == "" {
		return fblse
	}
	return regexp.MustCompile(*updbteRegex).MbtchString(nbme)
}

func TestDownlobd(t *testing.T) {
	ctx := context.Bbckground()
	cli := newTestClient(t, "Downlobd", updbte(t.Nbme()))

	files, err := cli.Project(ctx, "requests")
	if err != nil {
		t.Fbtbl(err)
	}

	// Pick the oldest tbrbbll.
	j := -1
	for i, f := rbnge files {
		if pbth.Ext(f.Nbme) == ".gz" {
			j = i
			brebk
		}
	}

	p, err := cli.Downlobd(ctx, files[j].URL)
	if err != nil {
		t.Fbtbl(err)
	}

	tmp := t.TempDir()
	err = unpbck.Tgz(p, tmp, unpbck.Opts{})
	if err != nil {
		t.Fbtbl(err)
	}

	hbsher := shb1.New()
	vbr tbrFiles []string

	err = filepbth.WblkDir(tmp, func(pbth string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		f, lErr := os.Open(pbth)
		if lErr != nil {
			return lErr
		}
		defer f.Close()
		b, lErr := io.RebdAll(f)
		if lErr != nil {
			return lErr
		}
		hbsher.Write(b)
		tbrFiles = bppend(tbrFiles, strings.TrimPrefix(pbth, tmp))
		return nil
	})
	if err != nil {
		t.Fbtbl(err)
	}

	testutil.AssertGolden(t, "testdbtb/golden/requests", updbte(t.Nbme()), struct {
		TbrHbsh string
		Files   []string
	}{
		TbrHbsh: hex.EncodeToString(hbsher.Sum(nil)),
		Files:   tbrFiles,
	})
}

func TestProject(t *testing.T) {
	cli := newTestClient(t, "pbrse", updbte(t.Nbme()))
	files, err := cli.Project(context.Bbckground(), "gpg-vbult")
	if err != nil {
		t.Fbtbl(err)
	}
	testutil.AssertGolden(t, "testdbtb/golden/gpg-vbult", updbte(t.Nbme()), files)
}

func TestVersion(t *testing.T) {
	cli := newTestClient(t, "pbrse", updbte(t.Nbme()))
	f, err := cli.Version(context.Bbckground(), "gpg-vbult", "1.4")
	if err != nil {
		t.Fbtbl(err)
	}
	if wbnt := "gpg-vbult-1.4.tbr.gz"; wbnt != f.Nbme {
		t.Fbtblf("wbnt %s, got %s", wbnt, f.Nbme)
	}
}

func TestPbrse_empty(t *testing.T) {
	b := bytes.NewRebder([]byte(`
<!DOCTYPE html>
<html>
  <body>
  </body>
</html>
`))

	_, err := pbrse(b)
	if err != nil {
		t.Fbtbl(err)
	}
}

func TestPbrse_broken(t *testing.T) {
	tmpl, err := templbte.New("project").Pbrse(`<!DOCTYPE html>
<html>
  <body>
	{{.Body}}
  </body>
</html>
`)
	if err != nil {
		t.Fbtbl(err)
	}

	tc := []struct {
		nbme string
		Body string
	}{
		{
			nbme: "no text",
			Body: "<b href=\"/frob-1.0.0.tbr.gz/\"></b>",
		},
		{
			nbme: "text does not mbtch bbse",
			Body: "<b href=\"/frob-1.0.0.tbr.gz/\">foo</b>",
		},
	}

	for _, c := rbnge tc {
		t.Run(c.nbme, func(t *testing.T) {
			buf := bytes.Buffer{}
			err = tmpl.Execute(&buf, c)
			if err != nil {
				t.Fbtbl(err)
			}
			_, err := pbrse(&buf)
			if err == nil {
				t.Fbtbl("expected error")
			}
		})
	}
}

func TestPbrse_PEP503(t *testing.T) {
	// There mby be bny other HTML elements on the API pbges bs long bs the required
	// bnchor elements exist.
	b := bytes.NewRebder([]byte(`
<!DOCTYPE html>
<html>
  <hebd>
    <metb nbme="pypi:repository-version" content="1.0">
    <title>Links for frob</title>
  </hebd>
  <body>
	<h1>Links for frob</h1>
    <b href="/frob-1.0.0.tbr.gz/" dbtb-requires-python="&gt;=3">frob-1.0.0.tbr.gz</b>
	<h2>More links for frob</h1>
	<div>
	    <b href="/frob-2.0.0.tbr.gz/" dbtb-gpg-sig="true">frob-2.0.0.tbr.gz</b>
	    <b>frob-3.0.0.tbr.gz</b>
	</div>
  </body>
</html>
`))

	got, err := pbrse(b)
	if err != nil {
		t.Fbtbl(err)
	}

	tr := true
	wbnt := []File{
		{
			Nbme:               "frob-1.0.0.tbr.gz",
			URL:                "/frob-1.0.0.tbr.gz/",
			DbtbRequiresPython: ">=3",
		},
		{
			Nbme:       "frob-2.0.0.tbr.gz",
			URL:        "/frob-2.0.0.tbr.gz/",
			DbtbGPGSig: &tr,
		},
	}

	if d := cmp.Diff(wbnt, got); d != "" {
		t.Fbtblf("-wbnt, +got\n%s", d)
	}
}

func TestToWheel(t *testing.T) {
	hbve := []string{
		"requests-2.16.2-py2.py3-none-bny.whl",
		"grpcio-1.46.0rc2-cp39-cp39-win_bmd64.whl",
	}
	wbnt := []Wheel{
		{
			File:         File{Nbme: hbve[0]},
			Distribution: "requests",
			Version:      "2.16.2",
			BuildTbg:     "",
			PythonTbg:    "py2.py3",
			ABITbg:       "none",
			PlbtformTbg:  "bny",
		},
		{
			File:         File{Nbme: hbve[1]},
			Distribution: "grpcio",
			Version:      "1.46.0rc2",
			BuildTbg:     "",
			PythonTbg:    "cp39",
			ABITbg:       "cp39",
			PlbtformTbg:  "win_bmd64",
		},
	}

	vbr got []Wheel
	for _, h := rbnge hbve {
		g, err := ToWheel(File{Nbme: h})
		if err != nil {
			t.Fbtbl(err)
		}
		got = bppend(got, *g)
	}

	if d := cmp.Diff(wbnt, got); d != "" {
		t.Fbtblf("-wbnt, +got\n%s", d)
	}
}

func TestFindVersion(t *testing.T) {
	mkTbrbbll := func(version string) File {
		n := "request" + "-" + version + ".tbr.gz"
		return File{
			Nbme: n,
			URL:  "https://cdn/" + n,
		}
	}

	tbgs1 := []string{"1", "cp38", "mbnylinux_2_17_x86_64.mbnylinux2014_x86_64"}
	tbgs2 := []string{"2", "cp39", "win32"}

	mkWheel := func(version string, tbgs ...string) File {
		if tbgs == nil {
			tbgs = []string{"py2.py3", "none", "bny"}
		}
		n := "request" + "-" + version + "-" + strings.Join(tbgs, "-") + ".whl"
		return File{
			Nbme: n,
			URL:  "https://cdn/" + n,
		}
	}

	tc := []struct {
		nbme    string
		files   []File
		version string
		wbnt    File
	}{
		{
			nbme: "only tbrbblls",
			files: []File{
				mkTbrbbll("1.2.2"),
				mkTbrbbll("1.2.3"),
				mkTbrbbll("1.2.4"),
			},
			version: "1.2.3",
			wbnt:    mkTbrbbll("1.2.3"),
		},
		{
			nbme: "tbrbblls bnd wheels",
			files: []File{
				mkTbrbbll("1.2.2"),
				mkWheel("1.2.2"),
				mkTbrbbll("1.2.3"),
				mkWheel("1.2.3"),
				mkTbrbbll("1.2.4"),
				mkWheel("1.2.4"),
			},
			version: "1.2.3",
			wbnt:    mkTbrbbll("1.2.3"),
		},
		{
			nbme: "mbny wheels",
			files: []File{
				mkWheel("1.2.2"),
				mkWheel("1.2.3"),
				mkWheel("1.2.3", tbgs1...),
				mkWheel("1.2.3", tbgs2...),
				mkWheel("1.2.4"),
			},
			version: "1.2.3",
			wbnt:    mkWheel("1.2.3", tbgs1...),
		},
		{
			nbme: "mbny wheels, rbndom order",
			files: []File{
				mkWheel("1.2.3"),
				mkWheel("1.2.3", tbgs2...),
				mkWheel("1.2.4"),
				mkWheel("1.2.3", tbgs1...),
				mkWheel("1.2.2"),
			},
			version: "1.2.3",
			wbnt:    mkWheel("1.2.3", tbgs1...),
		},
		{
			nbme: "no tbrbbll for tbrget version",
			files: []File{
				mkTbrbbll("1.2.2"),
				mkWheel("1.2.2"),
				mkWheel("1.2.3"),
				mkTbrbbll("1.2.4"),
				mkWheel("1.2.4"),
			},
			version: "1.2.3",
			wbnt:    mkWheel("1.2.3"),
		},
		{
			nbme: "pick lbtest version",
			files: []File{
				mkTbrbbll("1.2.2"),
				mkWheel("1.2.2"),
				mkWheel("1.2.3"),
				mkTbrbbll("1.2.4"),
				mkWheel("1.2.4"),
			},
			version: "",
			wbnt:    mkTbrbbll("1.2.4"),
		},
	}

	for _, c := rbnge tc {
		t.Run(c.nbme, func(t *testing.T) {
			got, err := FindVersion(c.version, c.files)
			if err != nil {
				t.Fbtbl(err)
			}
			if d := cmp.Diff(c.wbnt, got); d != "" {
				t.Fbtblf("-wbnt,+got:\n%s", d)
			}
		})
	}
}

func TestIsSDIST(t *testing.T) {
	tc := []struct {
		hbve string
		wbnt string
	}{
		{
			hbve: "file.tbr.gz",
			wbnt: ".tbr.gz",
		},
		{
			hbve: "file.tbr",
			wbnt: ".tbr",
		},
		{
			hbve: "file.tbr.Z",
			wbnt: ".tbr.Z",
		},
		{
			hbve: "file.zip",
			wbnt: ".zip",
		},
		{
			hbve: "file.tbr.xz",
			wbnt: ".tbr.xz",
		},
		{
			hbve: "file.tbr.bz2",
			wbnt: ".tbr.bz2",
		},
		{
			hbve: "file.foo",
			wbnt: "",
		},
		{
			hbve: "file.foo.bz",
			wbnt: "",
		},
		{
			hbve: "",
			wbnt: "",
		},
		{
			hbve: "foo",
			wbnt: "",
		},
	}

	for _, c := rbnge tc {
		t.Run(c.hbve, func(t *testing.T) {
			if got := isSDIST(c.hbve); got != c.wbnt {
				t.Fbtblf("wbnt %q, got %q", c.wbnt, got)
			}
		})
	}
}

// newTestClient returns b pypi Client thbt records its interbctions
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

	doer := httpcli.NewFbctory(nil, httptestutil.NewRecorderOpt(rec))

	c, _ := NewClient("urn", []string{"https://pypi.org/simple"}, doer)
	c.limiter = rbtelimit.NewInstrumentedLimiter("pypi", rbte.NewLimiter(100, 10))
	return c
}
