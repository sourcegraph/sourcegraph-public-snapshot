package server

import (
	"archive/zip"
	"bytes"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestUnpackPythonPackage_TGZ(t *testing.T) {
	files := []fileInfo{
		{
			path:     "common/file1.py",
			contents: []byte("banana"),
		},
		{
			path:     "common/setup.py",
			contents: []byte("apple"),
		},
		{
			path:     ".git/index",
			contents: []byte("filter me"),
		},
		{
			path:     "/absolute/path/are/filtered",
			contents: []byte("filter me"),
		},
	}

	pkg := createTgz(t, files)

	tmp := t.TempDir()
	if err := unpackPythonPackage(pkg, "https://some.where/pckg.tar.gz", tmp); err != nil {
		t.Fatal()
	}

	got := make([]string, 0, len(files))
	if err := filepath.Walk(tmp, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		got = append(got, strings.TrimPrefix(path, tmp))
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	sort.Slice(got, func(i, j int) bool {
		return got[i] < got[j]
	})

	// without the filtered files, the rest of the files share a common folder
	// "common" which should also be removed.
	want := []string{"/file1.py", "/setup.py"}
	sort.Slice(want, func(i, j int) bool {
		return want[i] < want[j]
	})

	if d := cmp.Diff(want, got); d != "" {
		t.Fatalf("-want,+got\n%s", d)
	}
}

func TestUnpackPythonPackage_ZIP(t *testing.T) {
	var zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)
	for _, f := range []fileInfo{
		{
			path:     "src/file1.py",
			contents: []byte("banana"),
		},
		{
			path:     "src/file2.py",
			contents: []byte("apple"),
		},
		{
			path:     "setup.py",
			contents: []byte("pear"),
		},
	} {
		fw, err := zw.Create(f.path)
		if err != nil {
			t.Fatal(err)
		}

		_, err = fw.Write(f.contents)
		if err != nil {
			t.Fatal(err)
		}
	}

	err := zw.Close()
	if err != nil {
		t.Fatal(err)
	}

	tmp := t.TempDir()
	if err := unpackPythonPackage(zipBuf.Bytes(), "https://some.where/pckg.zip", tmp); err != nil {
		t.Fatal()
	}

	var got []string
	if err := filepath.Walk(tmp, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		got = append(got, strings.TrimPrefix(path, tmp))
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	sort.Slice(got, func(i, j int) bool {
		return got[i] < got[j]
	})

	want := []string{"/src/file1.py", "/src/file2.py", "/setup.py"}
	sort.Slice(want, func(i, j int) bool {
		return want[i] < want[j]
	})

	if d := cmp.Diff(want, got); d != "" {
		t.Fatalf("-want,+got\n%s", d)
	}
}

func TestUnpackPythonPackage_InvalidZip(t *testing.T) {
	files := []fileInfo{
		{
			path:     "file1.py",
			contents: []byte("banana"),
		},
	}

	pkg := createTgz(t, files)

	if err := unpackPythonPackage(pkg, "https://some.where/pckg.whl", t.TempDir()); err == nil {
		t.Fatal()
	}
}

func TestUnpackPythonPackage_UnsupportedFormat(t *testing.T) {
	if err := unpackPythonPackage([]byte{}, "https://some.where/pckg.exe", ""); err == nil {
		t.Fatal()
	}
}
