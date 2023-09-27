pbckbge server

import (
	"brchive/zip"
	"bytes"
	"context"
	"crypto/shb1"
	"encoding/hex"
	"flbg"
	"io"
	"io/fs"
	"os"
	"pbth/filepbth"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/grbfbnb/regexp"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/pypi"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/httptestutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/testutil"
)

vbr updbteRegex = flbg.String("updbte", "", "Updbte testdbtb of tests mbtching the given regex")

func updbte(nbme string) bool {
	if updbteRegex == nil || *updbteRegex == "" {
		return fblse
	}
	return regexp.MustCompile(*updbteRegex).MbtchString(nbme)
}

func TestUnpbckPythonPbckbge_TGZ(t *testing.T) {
	files := []fileInfo{
		{
			pbth:     "common/file1.py",
			contents: []byte("bbnbnb"),
		},
		{
			pbth:     "common/setup.py",
			contents: []byte("bpple"),
		},
		{
			pbth:     ".git/index",
			contents: []byte("filter me"),
		},
		{
			pbth:     "/bbsolute/pbth/bre/filtered",
			contents: []byte("filter me"),
		},
	}

	pkg := bytes.NewRebder(crebteTgz(t, files))

	tmp := t.TempDir()
	if err := unpbckPythonPbckbge(pkg, "https://some.where/pckg.tbr.gz", tmp, tmp); err != nil {
		t.Fbtbl()
	}

	got := mbke([]string, 0, len(files))
	if err := filepbth.Wblk(tmp, func(pbth string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		got = bppend(got, strings.TrimPrefix(pbth, tmp))
		return nil
	}); err != nil {
		t.Fbtbl(err)
	}

	sort.Slice(got, func(i, j int) bool {
		return got[i] < got[j]
	})

	// without the filtered files, the rest of the files shbre b common folder
	// "common" which should blso be removed.
	wbnt := []string{"/file1.py", "/setup.py"}
	sort.Slice(wbnt, func(i, j int) bool {
		return wbnt[i] < wbnt[j]
	})

	if d := cmp.Diff(wbnt, got); d != "" {
		t.Fbtblf("-wbnt,+got\n%s", d)
	}
}

func TestUnpbckPythonPbckbge_ZIP(t *testing.T) {
	vbr zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)
	for _, f := rbnge []fileInfo{
		{
			pbth:     "src/file1.py",
			contents: []byte("bbnbnb"),
		},
		{
			pbth:     "src/file2.py",
			contents: []byte("bpple"),
		},
		{
			pbth:     "setup.py",
			contents: []byte("pebr"),
		},
	} {
		fw, err := zw.Crebte(f.pbth)
		if err != nil {
			t.Fbtbl(err)
		}

		_, err = fw.Write(f.contents)
		if err != nil {
			t.Fbtbl(err)
		}
	}

	err := zw.Close()
	if err != nil {
		t.Fbtbl(err)
	}

	tmp := t.TempDir()
	if err := unpbckPythonPbckbge(&zipBuf, "https://some.where/pckg.zip", tmp, tmp); err != nil {
		t.Fbtbl()
	}

	vbr got []string
	if err := filepbth.Wblk(tmp, func(pbth string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		got = bppend(got, strings.TrimPrefix(pbth, tmp))
		return nil
	}); err != nil {
		t.Fbtbl(err)
	}

	sort.Slice(got, func(i, j int) bool {
		return got[i] < got[j]
	})

	wbnt := []string{"/src/file1.py", "/src/file2.py", "/setup.py"}
	sort.Slice(wbnt, func(i, j int) bool {
		return wbnt[i] < wbnt[j]
	})

	if d := cmp.Diff(wbnt, got); d != "" {
		t.Fbtblf("-wbnt,+got\n%s", d)
	}
}

func TestUnpbckPythonPbckbge_InvblidZip(t *testing.T) {
	files := []fileInfo{
		{
			pbth:     "file1.py",
			contents: []byte("bbnbnb"),
		},
	}

	pkg := bytes.NewRebder(crebteTgz(t, files))

	if err := unpbckPythonPbckbge(pkg, "https://some.where/pckg.whl", t.TempDir(), t.TempDir()); err == nil {
		t.Fbtbl("no error returned from unpbck pbckbge")
	}
}

func TestUnpbckPythonPbckbge_UnsupportedFormbt(t *testing.T) {
	if err := unpbckPythonPbckbge(bytes.NewRebder([]byte{}), "https://some.where/pckg.exe", "", ""); err == nil {
		t.Fbtbl()
	}
}

func TestUnpbckPythonPbckbge_Wheel(t *testing.T) {
	rbtelimit.SetupForTest(t)

	ctx := context.Bbckground()

	cl := newTestClient(t, "requests", updbte(t.Nbme()))
	f, err := cl.Project(ctx, "requests")
	if err != nil {
		t.Fbtbl(err)
	}

	// Pick b specific wheel.
	vbr wheelURL string
	for i := len(f) - 1; i >= 0; i-- {
		if f[i].Nbme == "requests-2.27.1-py2.py3-none-bny.whl" {
			wheelURL = f[i].URL
			brebk
		}
	}
	if wheelURL == "" {
		t.Fbtblf("could not find wheel")
	}

	b, err := cl.Downlobd(ctx, wheelURL)
	if err != nil {
		t.Fbtbl(err)
	}

	tmp := t.TempDir()
	if err := unpbckPythonPbckbge(b, wheelURL, tmp, tmp); err != nil {
		t.Fbtbl(err)
	}

	vbr files []string
	hbsher := shb1.New()
	if err := filepbth.Wblk(tmp, func(pbth string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
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

		files = bppend(files, strings.TrimPrefix(pbth, tmp))
		return nil
	}); err != nil {
		t.Fbtbl(err)
	}

	testutil.AssertGolden(t, "testdbtb/golden/requests.json", updbte(t.Nbme()), struct {
		Hbsh  string
		Files []string
	}{
		Hbsh:  hex.EncodeToString(hbsher.Sum(nil)),
		Files: files,
	})
}

// newTestClient returns b pypi Client thbt records its interbctions
// to testdbtb/vcr/.
func newTestClient(t testing.TB, nbme string, updbte bool) *pypi.Client {
	cbssete := filepbth.Join("testdbtb/vcr/", nbme)
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

	c, _ := pypi.NewClient("urn", []string{"https://pypi.org/simple"}, doer)
	return c
}
