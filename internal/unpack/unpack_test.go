package unpack

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestTgzFallback(t *testing.T) {
	tarBytes := makeTar(t, &fileInfo{path: "foo", contents: "bar", mode: 0655})

	t.Run("with-io-read-seeker", func(t *testing.T) {
		err := Tgz(bytes.NewReader(tarBytes), t.TempDir(), Opts{})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("without-io-read-seeker", func(t *testing.T) {
		err := Tgz(bytes.NewBuffer(tarBytes), t.TempDir(), Opts{})
		if err != nil {
			t.Fatal(err)
		}
	})
}

// TestUnpack tests general properties of all unpack functions.
func TestUnpack(t *testing.T) {
	type packer struct {
		name   string
		unpack func(io.Reader, string, Opts) error
		pack   func(testing.TB, ...*fileInfo) []byte
	}

	type testCase struct {
		packer
		name        string
		opts        Opts
		in          []*fileInfo
		out         []*fileInfo
		err         string
		errContains string
	}

	var testCases []testCase
	for _, p := range []packer{
		{"tar", Tar, makeTar},
		{"tgz", Tgz, makeTgz},
		{"zip", func(r io.Reader, dir string, opts Opts) error {
			br := r.(*bytes.Reader)
			return Zip(br, int64(br.Len()), dir, opts)
		}, makeZip},
	} {
		testCases = append(testCases, []testCase{
			{
				packer: p,
				name:   "filter",
				opts: Opts{
					Filter: func(path string, file fs.FileInfo) bool {
						return file.Size() <= 3 && (path == "bar" || path == "foo/bar")
					},
				},
				in: []*fileInfo{
					{path: "big", contents: "E_TOO_BIG", mode: 0655},
					{path: "bar/baz", contents: "bar", mode: 0655},
					{path: "bar", contents: "bar", mode: 0655},
					{path: "foo/bar", contents: "bar", mode: 0655},
				},
				out: []*fileInfo{
					{path: "bar", contents: "bar", mode: 0655, size: 3},
					{path: "foo", mode: fs.ModeDir | 0750},
					{path: "foo/bar", contents: "bar", mode: 0655, size: 3},
				},
			},
			{
				packer: p,
				name:   "empty-dirs",
				in: []*fileInfo{
					{path: "foo", mode: fs.ModeDir | 0740},
				},
				out: []*fileInfo{
					{path: "foo", mode: fs.ModeDir | 0740},
				},
			},
			{
				packer: p,
				name:   "illegal-file-path",
				in: []*fileInfo{
					{path: "../../etc/passwd", contents: "foo", mode: 0655},
				},
				err: "../../etc/passwd: illegal file path",
			},
			{
				packer: p,
				name:   "illegal-absolute-link-path",
				in: []*fileInfo{
					{path: "passwd", contents: "/etc/passwd", mode: fs.ModeSymlink},
				},
				err: "/etc/passwd: illegal link path",
			},
			{
				packer: p,
				name:   "illegal-relative-link-path",
				in: []*fileInfo{
					{path: "passwd", contents: "../../etc/passwd", mode: fs.ModeSymlink},
				},
				err: "../../etc/passwd: illegal link path",
			},
			{
				packer: p,
				name:   "skip-invalid",
				opts:   Opts{SkipInvalid: true},
				in: []*fileInfo{
					{path: "bar", contents: "bar", mode: 0655},
					{path: "../../etc/passwd", contents: "foo", mode: 0655},
					{path: "passwd", contents: "../../etc/passwd", mode: fs.ModeSymlink},
					{path: "passwd", contents: "/etc/passwd", mode: fs.ModeSymlink},
				},
				out: []*fileInfo{
					{path: "bar", contents: "bar", mode: 0655, size: 3},
				},
			},
			{
				packer: p,
				name:   "symbolic-link",
				in: []*fileInfo{
					{path: "bar", contents: "bar", mode: 0655},
					{path: "foo", contents: "bar", mode: fs.ModeSymlink},
				},
				out: []*fileInfo{
					{path: "bar", contents: "bar", mode: 0655, size: 3},
					{path: "foo", contents: "bar", mode: fs.ModeSymlink, size: 3},
				},
			},
			{
				packer: p,
				name:   "dir-permissions",
				in: []*fileInfo{
					{path: "dir", mode: fs.ModeDir},
					{path: "dir/file1", contents: "x", mode: 0000},
					{path: "dir/file2", contents: "x", mode: 0200},
					{path: "dir/file3", contents: "x", mode: 0400},
					{path: "dir/file4", contents: "x", mode: 0600},
				},
				out: []*fileInfo{
					{path: "dir", mode: fs.ModeDir | 0700},
					{path: "dir/file1", contents: "x", mode: 0600, size: 1},
					{path: "dir/file2", contents: "x", mode: 0600, size: 1},
					{path: "dir/file3", contents: "x", mode: 0600, size: 1},
					{path: "dir/file4", contents: "x", mode: 0600, size: 1},
				},
			},
			{
				packer: p,
				name:   "duplicates",
				in: []*fileInfo{
					{path: "bar", contents: "bar", mode: 0655},
					{path: "bar", contents: "bar", mode: 0655},
				},
				errContains: "/bar: file exists",
				out: []*fileInfo{
					{path: "bar", contents: "bar", mode: 0655, size: 3},
				},
			},
			{
				packer: p,
				name:   "skip-duplicates",
				opts:   Opts{SkipDuplicates: true},
				in: []*fileInfo{
					{path: "bar", contents: "bar", mode: 0655},
					{path: "bar", contents: "bar", mode: 0655},
				},
				out: []*fileInfo{
					{path: "bar", contents: "bar", mode: 0655, size: 3},
				},
			},
		}...)
	}

	for _, tc := range testCases {
		t.Run(path.Join(tc.packer.name, tc.name), func(t *testing.T) {
			dir := t.TempDir()

			err := tc.unpack(
				bytes.NewReader(tc.pack(t, tc.in...)),
				dir,
				tc.opts,
			)

			assertError(t, err, tc.err, tc.errContains)
			assertUnpack(t, dir, tc.out)
		})
	}
}

func makeZip(t testing.TB, files ...*fileInfo) []byte {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	for _, f := range files {
		h, err := zip.FileInfoHeader(f)
		if err != nil {
			t.Fatal(err)
		}

		h.Name = f.path
		fw, err := zw.CreateHeader(h)
		if err != nil {
			t.Fatal(err)
		}

		if len(f.contents) > 0 {
			if _, err := fw.Write([]byte(f.contents)); err != nil {
				t.Fatal(err)
			}
		}
	}

	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	return buf.Bytes()
}

func makeTgz(t testing.TB, files ...*fileInfo) []byte {
	var buf bytes.Buffer

	gzw := gzip.NewWriter(&buf)
	_, err := gzw.Write(makeTar(t, files...))
	if err != nil {
		t.Fatal(err)
	}

	if err = gzw.Close(); err != nil {
		t.Fatal(err)
	}

	return buf.Bytes()
}

func makeTar(t testing.TB, files ...*fileInfo) []byte {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	for _, f := range makeTarFiles(t, files...) {
		if err := tw.WriteHeader(f.Header); err != nil {
			t.Fatal(err)
		}

		if len(f.contents) > 0 && f.mode.IsRegular() {
			if _, err := tw.Write([]byte(f.contents)); err != nil {
				t.Fatal(err)
			}
		}
	}

	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}

	return buf.Bytes()
}

func assertError(t testing.TB, have error, want string, wantContains string) {
	if want == "" && wantContains != "" {
		haveMessage := fmt.Sprint(have)
		if !strings.Contains(haveMessage, wantContains) {
			t.Fatalf("error should contain %q, but doesn't: %q", wantContains, haveMessage)
		}
		return
	}

	if want == "" {
		want = "<nil>"
	}

	if diff := cmp.Diff(fmt.Sprint(have), want); diff != "" {
		t.Fatalf("error mismatch: %s", diff)
	}
}

func assertUnpack(t testing.TB, dir string, want []*fileInfo) {
	var have []*fileInfo
	_ = fs.WalkDir(os.DirFS(dir), ".", func(path string, d fs.DirEntry, err error) error {
		if path != "." {
			have = append(have, makeFileInfo(t, dir, path, d))
		}
		return nil
	})

	cmpOpts := []cmp.Option{
		cmp.AllowUnexported(fileInfo{}),
		cmpopts.IgnoreFields(fileInfo{}, "modtime"),
	}

	if diff := cmp.Diff(want, have, cmpOpts...); diff != "" {
		t.Errorf("files mismatch: %s", diff)
	}
}

type tarFile struct {
	*tar.Header
	*fileInfo
}

func makeTarFiles(t testing.TB, fs ...*fileInfo) []*tarFile {
	tfs := make([]*tarFile, 0, len(fs))
	for _, f := range fs {
		link := ""
		if f.mode&os.ModeSymlink != 0 {
			link = f.contents
		}

		header, err := tar.FileInfoHeader(f, link)
		if err != nil {
			t.Fatal(err)
		}

		header.Name = f.path
		tfs = append(tfs, &tarFile{Header: header, fileInfo: f})
	}
	return tfs
}

type fileInfo struct {
	path     string
	mode     fs.FileMode
	modtime  time.Time
	contents string
	size     int64
}

func makeFileInfo(t testing.TB, dir, path string, d fs.DirEntry) *fileInfo {
	info, err := d.Info()
	if err != nil {
		t.Fatal(err)
	}

	var (
		contents []byte
		mode     = info.Mode()
	)

	if !d.IsDir() {
		name := filepath.Join(dir, path)
		if mode&fs.ModeSymlink != 0 {
			link, err := os.Readlink(name)
			if err != nil {
				t.Fatal(err)
			}
			// Different OSes set different permissions in a symlink so we ignore them.
			mode = fs.ModeSymlink
			contents = []byte(link)
		} else if contents, err = os.ReadFile(name); err != nil {
			t.Fatal(err)
		}
	}

	return &fileInfo{
		path:     path,
		mode:     mode,
		modtime:  info.ModTime(),
		contents: string(contents),
		size:     int64(len(contents)),
	}
}

var _ fs.FileInfo = &fileInfo{}

func (f *fileInfo) Name() string { return path.Base(f.path) }
func (f *fileInfo) Size() int64 {
	if f.size != 0 {
		return f.size
	}
	return int64(len(f.contents))
}
func (f *fileInfo) Mode() fs.FileMode  { return f.mode }
func (f *fileInfo) ModTime() time.Time { return f.modtime }
func (f *fileInfo) IsDir() bool        { return f.mode.IsDir() }
func (f *fileInfo) Sys() any           { return nil }
