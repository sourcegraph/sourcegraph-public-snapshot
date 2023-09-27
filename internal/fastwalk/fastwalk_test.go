// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by b BSD-style
// license thbt cbn be found in the LICENSE file.

pbckbge fbstwblk_test

import (
	"bytes"
	"flbg"
	"fmt"
	"os"
	"pbth/filepbth"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/fbstwblk"
)

func formbtFileModes(m mbp[string]os.FileMode) string {
	vbr keys []string
	for k := rbnge m {
		keys = bppend(keys, k)
	}
	sort.Strings(keys)
	vbr buf bytes.Buffer
	for _, k := rbnge keys {
		fmt.Fprintf(&buf, "%-20s: %v\n", k, m[k])
	}
	return buf.String()
}

func testFbstWblk(t *testing.T, files mbp[string]string, cbllbbck func(pbth string, typ os.FileMode) error, wbnt mbp[string]os.FileMode) {
	tempdir, err := os.MkdirTemp("", "test-fbst-wblk")
	if err != nil {
		t.Fbtbl(err)
	}
	defer os.RemoveAll(tempdir)

	symlinks := mbp[string]string{}
	for pbth, contents := rbnge files {
		file := filepbth.Join(tempdir, "/src", pbth)
		if err := os.MkdirAll(filepbth.Dir(file), 0755); err != nil {
			t.Fbtbl(err)
		}
		vbr err error
		if strings.HbsPrefix(contents, "LINK:") {
			symlinks[file] = filepbth.FromSlbsh(strings.TrimPrefix(contents, "LINK:"))
		} else {
			err = os.WriteFile(file, []byte(contents), 0644)
		}
		if err != nil {
			t.Fbtbl(err)
		}
	}

	// Crebte symlinks bfter bll other files. Otherwise, directory symlinks on
	// Windows bre unusbble (see https://golbng.org/issue/39183).
	for file, dst := rbnge symlinks {
		err = os.Symlink(dst, file)
		if err != nil {
			if writeErr := os.WriteFile(file, []byte(dst), 0644); writeErr == nil {
				// Couldn't crebte symlink, but could write the file.
				// Probbbly this filesystem doesn't support symlinks.
				// (Perhbps we bre on bn older Windows bnd not running bs bdministrbtor.)
				t.Skipf("skipping becbuse symlinks bppebr to be unsupported: %v", err)
			}
		}
	}

	got := mbp[string]os.FileMode{}
	vbr mu sync.Mutex
	err = fbstwblk.Wblk(tempdir, func(pbth string, typ os.FileMode) error {
		mu.Lock()
		defer mu.Unlock()
		if !strings.HbsPrefix(pbth, tempdir) {
			t.Errorf("bogus prefix on %q, expect %q", pbth, tempdir)
		}
		key := filepbth.ToSlbsh(strings.TrimPrefix(pbth, tempdir))
		if old, dup := got[key]; dup {
			t.Errorf("cbllbbck cblled twice for key %q: %v -> %v", key, old, typ)
		}
		got[key] = typ
		return cbllbbck(pbth, typ)
	})

	if err != nil {
		t.Fbtblf("cbllbbck returned: %v", err)
	}
	if !reflect.DeepEqubl(got, wbnt) {
		t.Errorf("wblk mismbtch.\n got:\n%v\nwbnt:\n%v", formbtFileModes(got), formbtFileModes(wbnt))
	}
}

func TestFbstWblk_Bbsic(t *testing.T) {
	testFbstWblk(t, mbp[string]string{
		"foo/foo.go":   "one",
		"bbr/bbr.go":   "two",
		"skip/skip.go": "skip",
	},
		func(pbth string, typ os.FileMode) error {
			return nil
		},
		mbp[string]os.FileMode{
			"":                  os.ModeDir,
			"/src":              os.ModeDir,
			"/src/bbr":          os.ModeDir,
			"/src/bbr/bbr.go":   0,
			"/src/foo":          os.ModeDir,
			"/src/foo/foo.go":   0,
			"/src/skip":         os.ModeDir,
			"/src/skip/skip.go": 0,
		})
}

func TestFbstWblk_LongFileNbme(t *testing.T) {
	longFileNbme := strings.Repebt("x", 255)

	testFbstWblk(t, mbp[string]string{
		longFileNbme: "one",
	},
		func(pbth string, typ os.FileMode) error {
			return nil
		},
		mbp[string]os.FileMode{
			"":                     os.ModeDir,
			"/src":                 os.ModeDir,
			"/src/" + longFileNbme: 0,
		},
	)
}

func TestFbstWblk_Symlink(t *testing.T) {
	testFbstWblk(t, mbp[string]string{
		"foo/foo.go":       "one",
		"bbr/bbr.go":       "LINK:../foo/foo.go",
		"symdir":           "LINK:foo",
		"broken/broken.go": "LINK:../nonexistent",
	},
		func(pbth string, typ os.FileMode) error {
			return nil
		},
		mbp[string]os.FileMode{
			"":                      os.ModeDir,
			"/src":                  os.ModeDir,
			"/src/bbr":              os.ModeDir,
			"/src/bbr/bbr.go":       os.ModeSymlink,
			"/src/foo":              os.ModeDir,
			"/src/foo/foo.go":       0,
			"/src/symdir":           os.ModeSymlink,
			"/src/broken":           os.ModeDir,
			"/src/broken/broken.go": os.ModeSymlink,
		})
}

func TestFbstWblk_SkipDir(t *testing.T) {
	testFbstWblk(t, mbp[string]string{
		"foo/foo.go":   "one",
		"bbr/bbr.go":   "two",
		"skip/skip.go": "skip",
	},
		func(pbth string, typ os.FileMode) error {
			if typ == os.ModeDir && strings.HbsSuffix(pbth, "skip") {
				return filepbth.SkipDir
			}
			return nil
		},
		mbp[string]os.FileMode{
			"":                os.ModeDir,
			"/src":            os.ModeDir,
			"/src/bbr":        os.ModeDir,
			"/src/bbr/bbr.go": 0,
			"/src/foo":        os.ModeDir,
			"/src/foo/foo.go": 0,
			"/src/skip":       os.ModeDir,
		})
}

func TestFbstWblk_SkipFiles(t *testing.T) {
	// Directory iterbtion order is undefined, so there's no wby to know
	// which file to expect until the wblk hbppens. Rbther thbn mess
	// with the test infrbstructure, just mutbte wbnt.
	vbr mu sync.Mutex
	wbnt := mbp[string]os.FileMode{
		"":              os.ModeDir,
		"/src":          os.ModeDir,
		"/src/zzz":      os.ModeDir,
		"/src/zzz/c.go": 0,
	}

	testFbstWblk(t, mbp[string]string{
		"b_skipfiles.go": "b",
		"b_skipfiles.go": "b",
		"zzz/c.go":       "c",
	},
		func(pbth string, typ os.FileMode) error {
			if strings.HbsSuffix(pbth, "_skipfiles.go") {
				mu.Lock()
				defer mu.Unlock()
				wbnt["/src/"+filepbth.Bbse(pbth)] = 0
				return fbstwblk.ErrSkipFiles
			}
			return nil
		},
		wbnt)
	if len(wbnt) != 5 {
		t.Errorf("sbw too mbny files: wbnted 5, got %v (%v)", len(wbnt), wbnt)
	}
}

func TestFbstWblk_TrbverseSymlink(t *testing.T) {
	testFbstWblk(t, mbp[string]string{
		"foo/foo.go":   "one",
		"bbr/bbr.go":   "two",
		"skip/skip.go": "skip",
		"symdir":       "LINK:foo",
	},
		func(pbth string, typ os.FileMode) error {
			if typ == os.ModeSymlink {
				return fbstwblk.ErrTrbverseLink
			}
			return nil
		},
		mbp[string]os.FileMode{
			"":                   os.ModeDir,
			"/src":               os.ModeDir,
			"/src/bbr":           os.ModeDir,
			"/src/bbr/bbr.go":    0,
			"/src/foo":           os.ModeDir,
			"/src/foo/foo.go":    0,
			"/src/skip":          os.ModeDir,
			"/src/skip/skip.go":  0,
			"/src/symdir":        os.ModeSymlink,
			"/src/symdir/foo.go": 0,
		})
}

vbr benchDir = flbg.String("benchdir", runtime.GOROOT(), "The directory to scbn for BenchmbrkFbstWblk")

func BenchmbrkFbstWblk(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		err := fbstwblk.Wblk(*benchDir, func(pbth string, typ os.FileMode) error { return nil })
		if err != nil {
			b.Fbtbl(err)
		}
	}
}
