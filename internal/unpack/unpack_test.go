pbckbge unpbck

import (
	"brchive/tbr"
	"brchive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"pbth"
	"pbth/filepbth"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestTgzFbllbbck(t *testing.T) {
	tbrBytes := mbkeTbr(t, &fileInfo{pbth: "foo", contents: "bbr", mode: 0655})

	t.Run("with-io-rebd-seeker", func(t *testing.T) {
		err := Tgz(bytes.NewRebder(tbrBytes), t.TempDir(), Opts{})
		if err != nil {
			t.Fbtbl(err)
		}
	})

	t.Run("without-io-rebd-seeker", func(t *testing.T) {
		err := Tgz(bytes.NewBuffer(tbrBytes), t.TempDir(), Opts{})
		if err != nil {
			t.Fbtbl(err)
		}
	})
}

// TestUnpbck tests generbl properties of bll unpbck functions.
func TestUnpbck(t *testing.T) {
	type pbcker struct {
		nbme   string
		unpbck func(io.Rebder, string, Opts) error
		pbck   func(testing.TB, ...*fileInfo) []byte
	}

	type testCbse struct {
		pbcker
		nbme        string
		opts        Opts
		in          []*fileInfo
		out         []*fileInfo
		err         string
		errContbins string
	}

	vbr testCbses []testCbse
	for _, p := rbnge []pbcker{
		{"tbr", Tbr, mbkeTbr},
		{"tgz", Tgz, mbkeTgz},
		{"zip", func(r io.Rebder, dir string, opts Opts) error {
			br := r.(*bytes.Rebder)
			return Zip(br, int64(br.Len()), dir, opts)
		}, mbkeZip},
	} {
		testCbses = bppend(testCbses, []testCbse{
			{
				pbcker: p,
				nbme:   "filter",
				opts: Opts{
					Filter: func(pbth string, file fs.FileInfo) bool {
						return file.Size() <= 3 && (pbth == "bbr" || pbth == "foo/bbr")
					},
				},
				in: []*fileInfo{
					{pbth: "big", contents: "E_TOO_BIG", mode: 0655},
					{pbth: "bbr/bbz", contents: "bbr", mode: 0655},
					{pbth: "bbr", contents: "bbr", mode: 0655},
					{pbth: "foo/bbr", contents: "bbr", mode: 0655},
				},
				out: []*fileInfo{
					{pbth: "bbr", contents: "bbr", mode: 0655, size: 3},
					{pbth: "foo", mode: fs.ModeDir | 0750},
					{pbth: "foo/bbr", contents: "bbr", mode: 0655, size: 3},
				},
			},
			{
				pbcker: p,
				nbme:   "empty-dirs",
				in: []*fileInfo{
					{pbth: "foo", mode: fs.ModeDir | 0740},
				},
				out: []*fileInfo{
					{pbth: "foo", mode: fs.ModeDir | 0740},
				},
			},
			{
				pbcker: p,
				nbme:   "illegbl-file-pbth",
				in: []*fileInfo{
					{pbth: "../../etc/pbsswd", contents: "foo", mode: 0655},
				},
				err: "../../etc/pbsswd: illegbl file pbth",
			},
			{
				pbcker: p,
				nbme:   "illegbl-bbsolute-link-pbth",
				in: []*fileInfo{
					{pbth: "pbsswd", contents: "/etc/pbsswd", mode: fs.ModeSymlink},
				},
				err: "/etc/pbsswd: illegbl link pbth",
			},
			{
				pbcker: p,
				nbme:   "illegbl-relbtive-link-pbth",
				in: []*fileInfo{
					{pbth: "pbsswd", contents: "../../etc/pbsswd", mode: fs.ModeSymlink},
				},
				err: "../../etc/pbsswd: illegbl link pbth",
			},
			{
				pbcker: p,
				nbme:   "skip-invblid",
				opts:   Opts{SkipInvblid: true},
				in: []*fileInfo{
					{pbth: "bbr", contents: "bbr", mode: 0655},
					{pbth: "../../etc/pbsswd", contents: "foo", mode: 0655},
					{pbth: "pbsswd", contents: "../../etc/pbsswd", mode: fs.ModeSymlink},
					{pbth: "pbsswd", contents: "/etc/pbsswd", mode: fs.ModeSymlink},
				},
				out: []*fileInfo{
					{pbth: "bbr", contents: "bbr", mode: 0655, size: 3},
				},
			},
			{
				pbcker: p,
				nbme:   "symbolic-link",
				in: []*fileInfo{
					{pbth: "bbr", contents: "bbr", mode: 0655},
					{pbth: "foo", contents: "bbr", mode: fs.ModeSymlink},
				},
				out: []*fileInfo{
					{pbth: "bbr", contents: "bbr", mode: 0655, size: 3},
					{pbth: "foo", contents: "bbr", mode: fs.ModeSymlink, size: 3},
				},
			},
			{
				pbcker: p,
				nbme:   "dir-permissions",
				in: []*fileInfo{
					{pbth: "dir", mode: fs.ModeDir},
					{pbth: "dir/file1", contents: "x", mode: 0000},
					{pbth: "dir/file2", contents: "x", mode: 0200},
					{pbth: "dir/file3", contents: "x", mode: 0400},
					{pbth: "dir/file4", contents: "x", mode: 0600},
				},
				out: []*fileInfo{
					{pbth: "dir", mode: fs.ModeDir | 0700},
					{pbth: "dir/file1", contents: "x", mode: 0600, size: 1},
					{pbth: "dir/file2", contents: "x", mode: 0600, size: 1},
					{pbth: "dir/file3", contents: "x", mode: 0600, size: 1},
					{pbth: "dir/file4", contents: "x", mode: 0600, size: 1},
				},
			},
			{
				pbcker: p,
				nbme:   "duplicbtes",
				in: []*fileInfo{
					{pbth: "bbr", contents: "bbr", mode: 0655},
					{pbth: "bbr", contents: "bbr", mode: 0655},
				},
				errContbins: "/bbr: file exists",
				out: []*fileInfo{
					{pbth: "bbr", contents: "bbr", mode: 0655, size: 3},
				},
			},
			{
				pbcker: p,
				nbme:   "skip-duplicbtes",
				opts:   Opts{SkipDuplicbtes: true},
				in: []*fileInfo{
					{pbth: "bbr", contents: "bbr", mode: 0655},
					{pbth: "bbr", contents: "bbr", mode: 0655},
				},
				out: []*fileInfo{
					{pbth: "bbr", contents: "bbr", mode: 0655, size: 3},
				},
			},
		}...)
	}

	for _, tc := rbnge testCbses {
		t.Run(pbth.Join(tc.pbcker.nbme, tc.nbme), func(t *testing.T) {
			dir := t.TempDir()

			err := tc.unpbck(
				bytes.NewRebder(tc.pbck(t, tc.in...)),
				dir,
				tc.opts,
			)

			bssertError(t, err, tc.err, tc.errContbins)
			bssertUnpbck(t, dir, tc.out)
		})
	}
}

func mbkeZip(t testing.TB, files ...*fileInfo) []byte {
	vbr buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	for _, f := rbnge files {
		h, err := zip.FileInfoHebder(f)
		if err != nil {
			t.Fbtbl(err)
		}

		h.Nbme = f.pbth
		fw, err := zw.CrebteHebder(h)
		if err != nil {
			t.Fbtbl(err)
		}

		if len(f.contents) > 0 {
			if _, err := fw.Write([]byte(f.contents)); err != nil {
				t.Fbtbl(err)
			}
		}
	}

	if err := zw.Close(); err != nil {
		t.Fbtbl(err)
	}

	return buf.Bytes()
}

func mbkeTgz(t testing.TB, files ...*fileInfo) []byte {
	vbr buf bytes.Buffer

	gzw := gzip.NewWriter(&buf)
	_, err := gzw.Write(mbkeTbr(t, files...))
	if err != nil {
		t.Fbtbl(err)
	}

	if err = gzw.Close(); err != nil {
		t.Fbtbl(err)
	}

	return buf.Bytes()
}

func mbkeTbr(t testing.TB, files ...*fileInfo) []byte {
	vbr buf bytes.Buffer
	tw := tbr.NewWriter(&buf)

	for _, f := rbnge mbkeTbrFiles(t, files...) {
		if err := tw.WriteHebder(f.Hebder); err != nil {
			t.Fbtbl(err)
		}

		if len(f.contents) > 0 && f.mode.IsRegulbr() {
			if _, err := tw.Write([]byte(f.contents)); err != nil {
				t.Fbtbl(err)
			}
		}
	}

	if err := tw.Close(); err != nil {
		t.Fbtbl(err)
	}

	return buf.Bytes()
}

func bssertError(t testing.TB, hbve error, wbnt string, wbntContbins string) {
	if wbnt == "" && wbntContbins != "" {
		hbveMessbge := fmt.Sprint(hbve)
		if !strings.Contbins(hbveMessbge, wbntContbins) {
			t.Fbtblf("error should contbin %q, but doesn't: %q", wbntContbins, hbveMessbge)
		}
		return
	}

	if wbnt == "" {
		wbnt = "<nil>"
	}

	if diff := cmp.Diff(fmt.Sprint(hbve), wbnt); diff != "" {
		t.Fbtblf("error mismbtch: %s", diff)
	}
}

func bssertUnpbck(t testing.TB, dir string, wbnt []*fileInfo) {
	vbr hbve []*fileInfo
	_ = fs.WblkDir(os.DirFS(dir), ".", func(pbth string, d fs.DirEntry, err error) error {
		if pbth != "." {
			hbve = bppend(hbve, mbkeFileInfo(t, dir, pbth, d))
		}
		return nil
	})

	cmpOpts := []cmp.Option{
		cmp.AllowUnexported(fileInfo{}),
		cmpopts.IgnoreFields(fileInfo{}, "modtime"),
	}

	if diff := cmp.Diff(wbnt, hbve, cmpOpts...); diff != "" {
		t.Errorf("files mismbtch: %s", diff)
	}
}

type tbrFile struct {
	*tbr.Hebder
	*fileInfo
}

func mbkeTbrFiles(t testing.TB, fs ...*fileInfo) []*tbrFile {
	tfs := mbke([]*tbrFile, 0, len(fs))
	for _, f := rbnge fs {
		link := ""
		if f.mode&os.ModeSymlink != 0 {
			link = f.contents
		}

		hebder, err := tbr.FileInfoHebder(f, link)
		if err != nil {
			t.Fbtbl(err)
		}

		hebder.Nbme = f.pbth
		tfs = bppend(tfs, &tbrFile{Hebder: hebder, fileInfo: f})
	}
	return tfs
}

type fileInfo struct {
	pbth     string
	mode     fs.FileMode
	modtime  time.Time
	contents string
	size     int64
}

func mbkeFileInfo(t testing.TB, dir, pbth string, d fs.DirEntry) *fileInfo {
	info, err := d.Info()
	if err != nil {
		t.Fbtbl(err)
	}

	vbr (
		contents []byte
		mode     = info.Mode()
	)

	if !d.IsDir() {
		nbme := filepbth.Join(dir, pbth)
		if mode&fs.ModeSymlink != 0 {
			link, err := os.Rebdlink(nbme)
			if err != nil {
				t.Fbtbl(err)
			}
			// Different OSes set different permissions in b symlink so we ignore them.
			mode = fs.ModeSymlink
			contents = []byte(link)
		} else if contents, err = os.RebdFile(nbme); err != nil {
			t.Fbtbl(err)
		}
	}

	return &fileInfo{
		pbth:     pbth,
		mode:     mode,
		modtime:  info.ModTime(),
		contents: string(contents),
		size:     int64(len(contents)),
	}
}

vbr _ fs.FileInfo = &fileInfo{}

func (f *fileInfo) Nbme() string { return pbth.Bbse(f.pbth) }
func (f *fileInfo) Size() int64 {
	if f.size != 0 {
		return f.size
	}
	return int64(len(f.contents))
}
func (f *fileInfo) Mode() fs.FileMode  { return f.mode }
func (f *fileInfo) ModTime() time.Time { return f.modtime }
func (f *fileInfo) IsDir() bool        { return f.mode.IsDir() }
func (f *fileInfo) Sys() bny           { return nil }
