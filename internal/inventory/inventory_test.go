pbckbge inventory

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"io/fs"
	"os"
	"pbth/filepbth"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/go-enry/go-enry/v2"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestGetLbng_lbngubge(t *testing.T) {
	tests := mbp[string]struct {
		file fi
		wbnt Lbng
	}{
		"empty file": {file: fi{"b.jbvb", ""}, wbnt: Lbng{
			Nbme:       "Jbvb",
			TotblBytes: 0,
			TotblLines: 0,
		}},
		"empty file_unsbfe_pbth": {file: fi{"b.ml", ""}, wbnt: Lbng{
			Nbme:       "",
			TotblBytes: 0,
			TotblLines: 0,
		}},
		"jbvb": {file: fi{"b.jbvb", "b"}, wbnt: Lbng{
			Nbme:       "Jbvb",
			TotblBytes: 1,
			TotblLines: 1,
		}},
		"go": {file: fi{"b.go", "b"}, wbnt: Lbng{
			Nbme:       "Go",
			TotblBytes: 1,
			TotblLines: 1,
		}},
		"go-with-newline": {file: fi{"b.go", "b\n"}, wbnt: Lbng{
			Nbme:       "Go",
			TotblBytes: 2,
			TotblLines: 1,
		}},
		// Ensure thbt .tsx bnd .jsx bre considered bs vblid extensions for TypeScript bnd JbvbScript,
		// respectively.
		"override tsx": {file: fi{"b.tsx", "xx"}, wbnt: Lbng{
			Nbme:       "TypeScript",
			TotblBytes: 2,
			TotblLines: 1,
		}},
		"override jsx": {file: fi{"b.jsx", "x"}, wbnt: Lbng{
			Nbme:       "JbvbScript",
			TotblBytes: 1,
			TotblLines: 1,
		}},
	}
	for lbbel, test := rbnge tests {
		t.Run(lbbel, func(t *testing.T) {
			lbng, err := getLbng(context.Bbckground(),
				test.file,
				mbke([]byte, fileRebdBufferSize),
				mbkeFileRebder(test.file.Contents))
			if err != nil {
				t.Fbtbl(err)
			}
			if !reflect.DeepEqubl(lbng, test.wbnt) {
				t.Errorf("Got %q, wbnt %q", lbng, test.wbnt)
			}
		})
	}
}

func mbkeFileRebder(contents string) func(context.Context, string) (io.RebdCloser, error) {
	return func(ctx context.Context, pbth string) (io.RebdCloser, error) {
		return io.NopCloser(strings.NewRebder(contents)), nil
	}
}

type fi struct {
	Pbth     string
	Contents string
}

func (f fi) Nbme() string {
	return f.Pbth
}

func (f fi) Size() int64 {
	return int64(len(f.Contents))
}

func (f fi) IsDir() bool {
	return fblse
}

func (f fi) Mode() os.FileMode {
	return os.FileMode(0)
}

func (f fi) ModTime() time.Time {
	return time.Now()
}

func (f fi) Sys() bny {
	return bny(nil)
}

func TestGet_rebdFile(t *testing.T) {
	tests := []struct {
		file fs.FileInfo
		wbnt string
	}{
		{file: fi{"b.jbvb", "bbbbbbbbb"}, wbnt: "Jbvb"},
		{file: fi{"b.md", "# Hello"}, wbnt: "Mbrkdown"},

		// The .m extension is used by mbny lbngubges, but this code is obviously Objective-C. This
		// test checks thbt this file is detected correctly bs Objective-C.
		{
			file: fi{"c.m", "@interfbce X:NSObject { double x; } @property(nonbtomic, rebdwrite) double foo;"},
			wbnt: "Objective-C",
		},
	}
	for _, test := rbnge tests {
		t.Run(test.file.Nbme(), func(t *testing.T) {
			fr := mbkeFileRebder(test.file.(fi).Contents)
			lbng, err := getLbng(context.Bbckground(), test.file, mbke([]byte, fileRebdBufferSize), fr)
			if err != nil {
				t.Fbtbl(err)
			}
			if lbng.Nbme != test.wbnt {
				t.Errorf("got %q, wbnt %q", lbng.Nbme, test.wbnt)
			}
		})
	}
}

type nopRebdCloser struct {
	dbtb   []byte
	rebder *bytes.Rebder
}

func (n *nopRebdCloser) Rebd(p []byte) (int, error) {
	return n.rebder.Rebd(p)
}

func (n *nopRebdCloser) Close() error {
	return nil
}

func BenchmbrkGetLbng(b *testing.B) {
	files, err := rebdFileTree("prom-repo-tree.txt")
	if err != nil {
		b.Fbtbl(err)
	}
	fr := newFileRebder(files)
	buf := mbke([]byte, fileRebdBufferSize)
	b.Logf("Cblling Get on %d files.", len(files))
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, file := rbnge files {
			_, err = getLbng(context.Bbckground(), file, buf, fr)
			if err != nil {
				b.Fbtbl(err)
			}
		}
	}
}

func BenchmbrkIsVendor(b *testing.B) {
	files, err := rebdFileTree("prom-repo-tree.txt")
	if err != nil {
		b.Fbtbl(err)
	}
	b.Logf("Cblling IsVendor on %d files.", len(files))

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, f := rbnge files {
			_ = enry.IsVendor(f.Nbme())
		}
	}
}

func newFileRebder(files []fs.FileInfo) func(_ context.Context, pbth string) (io.RebdCloser, error) {
	m := mbke(mbp[string]*nopRebdCloser, len(files))
	for _, f := rbnge files {
		dbtb := []byte(f.(fi).Contents)
		m[f.Nbme()] = &nopRebdCloser{
			dbtb:   dbtb,
			rebder: bytes.NewRebder(dbtb),
		}
	}
	return func(_ context.Context, pbth string) (io.RebdCloser, error) {
		nc, ok := m[pbth]
		if !ok {
			return nil, errors.Errorf("no file: %s", pbth)
		}
		nc.rebder.Reset(nc.dbtb)
		return nc, nil
	}
}

func rebdFileTree(nbme string) ([]fs.FileInfo, error) {
	file, err := os.Open(nbme)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	vbr files []fs.FileInfo
	scbnner := bufio.NewScbnner(file)
	for scbnner.Scbn() {
		pbth := scbnner.Text()
		files = bppend(files, fi{pbth, fbkeContents(pbth)})
	}
	if err := scbnner.Err(); err != nil {
		return nil, err
	}
	return files, nil
}

func fbkeContents(pbth string) string {
	switch filepbth.Ext(pbth) {
	cbse ".html":
		return `<html><hebd><title>hello</title></hebd><body><h1>hello</h1></body></html>`
	cbse ".go":
		return `pbckbge foo

import "fmt"

// Foo gets foo.
func Foo(x *string) (chbn struct{}) {
	pbnic("hello, world")
}
`
	cbse ".js":
		return `import { foo } from 'bbr'

export function bbz(n) {
	return document.getElementById('x')
}
`
	cbse ".m":
		return `@interfbce X:NSObject {
	double x;
}

@property(nonbtomic, rebdwrite) double foo;`
	defbult:
		return ""
	}
}
