package testutil

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"sourcegraph.com/sourcegraph/rwvfs"
)

func Write(t *testing.T, fs rwvfs.FileSystem) {
	label := fmt.Sprintf("%T", fs)
	fs.Mkdir("/foo")
	defer fs.Remove("/foo")
	path := "/foo/bar"

	w, err := fs.Create(path)
	if err != nil {
		t.Fatalf("%s: WriterOpen: %s", label, err)
	}

	input := []byte("qux")
	_, err = w.Write(input)
	if err != nil {
		t.Fatalf("%s: Write: %s", label, err)
	}

	err = w.Close()
	if err != nil {
		t.Fatalf("%s: w.Close: %s", label, err)
	}

	var r io.ReadCloser
	r, err = fs.Open(path)
	if err != nil {
		t.Fatalf("%s: Open: %s", label, err)
	}
	var output []byte
	output, err = ioutil.ReadAll(r)
	if err != nil {
		t.Fatalf("%s: ReadAll: %s", label, err)
	}
	if !bytes.Equal(output, input) {
		t.Errorf("%s: got output %q, want %q", label, output, input)
	}

	IsFile(t, label, fs, path)
	IsDir(t, label, fs, "/foo")
	infos, err := fs.ReadDir("/foo")
	if err != nil {
		t.Fatalf("%s: ReadDir: %s", label, err)
	}
	if len(infos) != 1 || infos[0].Name() != "bar" {
		t.Fatalf("%s: ReadDir: got %v, want file 'bar'", label, infos)
	}

	r, err = fs.Open(path)
	if err != nil {
		t.Fatalf("%s: Open: %s", label, err)
	}
	output, err = ioutil.ReadAll(r)
	if err != nil {
		t.Fatalf("%s: ReadAll: %s", label, err)
	}
	if !bytes.Equal(output, input) {
		t.Errorf("%s: got output %q, want %q", label, output, input)
	}

	err = fs.Remove(path)
	if err != nil {
		t.Errorf("%s: Remove(%q): %s", label, path, err)
	}
	PathDoesNotExist(t, label, fs, path)
}

func Mkdir(t *testing.T, fs rwvfs.FileSystem) {
	label := fmt.Sprintf("%T", fs)

	if strings.Contains(label, "subFS") {
		if err := fs.Mkdir("/"); err != nil && !os.IsExist(err) {
			t.Fatalf("%s: subFS Mkdir(/): %s", label, err)
		}
	}
	if strings.Contains(label, "mapFS") {
		if err := fs.Mkdir("/"); err != nil && !os.IsExist(err) {
			t.Fatalf("%s: mapFS Mkdir(/): %s", label, err)
		}
	}

	fi, err := fs.Stat(".")
	if err != nil {
		t.Fatalf("%s: Stat(.): %s", label, err)
	}
	if !fi.Mode().IsDir() {
		t.Fatalf("%s: got Stat(.) FileMode %o, want IsDir", label, fi.Mode())
	}

	fi, err = fs.Stat("/")
	if err != nil {
		t.Fatalf("%s: Stat(/): %s", label, err)
	}
	if !fi.Mode().IsDir() {
		t.Fatalf("%s: got Stat(/) FileMode %o, want IsDir", label, fi.Mode())
	}

	if _, err := fs.ReadDir("."); err != nil {
		t.Fatalf("%s: ReadDir(.): %s", label, err)
	}
	if _, err := fs.ReadDir("/"); err != nil {
		t.Fatalf("%s: ReadDir(/): %s", label, err)
	}

	fis, err := fs.ReadDir("/")
	if err != nil {
		t.Fatalf("%s: ReadDir(/): %s", label, err)
	}
	if len(fis) != 0 {
		t.Fatalf("%s: ReadDir(/): got %d file infos (%v), want none (is it including .?)", label, len(fis), fis)
	}

	err = fs.Mkdir("dir0")
	if err != nil {
		t.Fatalf("%s: Mkdir(dir0): %s", label, err)
	}
	IsDir(t, label, fs, "dir0")
	IsDir(t, label, fs, "/dir0")

	err = fs.Mkdir("/dir1")
	if err != nil {
		t.Fatalf("%s: Mkdir(/dir1): %s", label, err)
	}
	IsDir(t, label, fs, "dir1")
	IsDir(t, label, fs, "/dir1")

	err = fs.Mkdir("/dir1")
	if !os.IsExist(err) {
		t.Errorf("%s: Mkdir(/dir1) again: got err %v, want os.IsExist-satisfying error", label, err)
	}

	err = fs.Mkdir("/parent-doesnt-exist/dir2")
	if !os.IsNotExist(err) {
		t.Errorf("%s: Mkdir(/parent-doesnt-exist/dir2): got error %v, want os.IsNotExist-satisfying error", label, err)
	}

	err = fs.Remove("/dir1")
	if err != nil {
		t.Errorf("%s: Remove(/dir1): %s", label, err)
	}
	PathDoesNotExist(t, label, fs, "/dir1")
}

func MkdirAll(t *testing.T, fs rwvfs.FileSystem) {
	label := fmt.Sprintf("%T", fs)

	err := rwvfs.MkdirAll(fs, "/a/b/c")
	if err != nil {
		t.Fatalf("%s: MkdirAll: %s", label, err)
	}
	IsDir(t, label, fs, "/a")
	IsDir(t, label, fs, "/a/b")
	IsDir(t, label, fs, "/a/b/c")

	err = rwvfs.MkdirAll(fs, "/a/b/c")
	if err != nil {
		t.Fatalf("%s: MkdirAll again: %s", label, err)
	}
}

func Glob(t *testing.T, fs rwvfs.FileSystem) {
	label := fmt.Sprintf("%T", fs)

	files := []string{"x/y/0.txt", "x/y/1.txt", "x/2.txt"}
	for _, file := range files {
		err := rwvfs.MkdirAll(fs, filepath.Dir(file))
		if err != nil {
			t.Fatalf("%s: MkdirAll: %s", label, err)
		}
		w, err := fs.Create(file)
		if err != nil {
			t.Errorf("%s: Create(%q): %s", label, file, err)
			return
		}
		w.Close()
	}

	globTests := []struct {
		prefix  string
		pattern string
		matches []string
	}{
		{"", "x/y/*.txt", []string{"x/y/0.txt", "x/y/1.txt"}},
		{"x/y", "x/y/*.txt", []string{"x/y/0.txt", "x/y/1.txt"}},
		{"", "x/*", []string{"x/y", "x/2.txt"}},
	}
	for _, test := range globTests {
		matches, err := rwvfs.Glob(rwvfs.Walkable(fs), test.prefix, test.pattern)
		if err != nil {
			t.Errorf("%s: Glob(prefix=%q, pattern=%q): %s", label, test.prefix, test.pattern, err)
			continue
		}
		sort.Strings(test.matches)
		sort.Strings(matches)
		if !reflect.DeepEqual(matches, test.matches) {
			t.Errorf("%s: Glob(prefix=%q, pattern=%q): got %v, want %v", label, test.prefix, test.pattern, matches, test.matches)
		}
	}
}

func IsDir(t *testing.T, label string, fs rwvfs.FileSystem, path string) {
	fi, err := fs.Stat(path)
	if err != nil {
		t.Fatalf("%s: Stat(%q): %s", label, path, err)
	}

	if fi == nil {
		t.Fatalf("%s: FileInfo (%q) == nil", label, path)
	}

	if !fi.IsDir() {
		t.Errorf("%s: got fs.Stat(%q) IsDir() == false, want true", label, path)
	}
}

func IsFile(t *testing.T, label string, fs rwvfs.FileSystem, path string) {
	fi, err := fs.Stat(path)
	if err != nil {
		t.Fatalf("%s: Stat(%q): %s", label, path, err)
	}

	if !fi.Mode().IsRegular() {
		t.Errorf("%s: got fs.Stat(%q) Mode().IsRegular() == false, want true", label, path)
	}
}

func PathDoesNotExist(t *testing.T, label string, fs rwvfs.FileSystem, path string) {
	fi, err := fs.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		t.Errorf("%s: Stat(%q): want os.IsNotExist-satisfying error, got %q", label, path, err)
	} else if err == nil {
		t.Errorf("%s: Stat(%q): want file to not exist, got existing file with FileInfo %+v", label, path, fi)
	}
}
