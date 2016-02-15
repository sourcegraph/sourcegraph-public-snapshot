package toolchain

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"sourcegraph.com/sourcegraph/srclib"
)

func TestList_program(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "srclib-toolchain-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	defer func(orig string) {
		srclib.Path = orig
	}(srclib.Path)
	srclib.Path = tmpdir

	var extension string
	if runtime.GOOS == "windows" {
		extension = ".exe"
	} else {
		extension = ""
	}
	files := map[string]os.FileMode{
		// ok
		filepath.Join("a", "a", ".bin", "a"+extension): 0700,
		filepath.Join("a", "a", "Srclibtoolchain"):     0700,

		// not executable
		filepath.Join("b", "b", ".bin", "z"+extension): 0600,
		filepath.Join("b", "b", "Srclibtoolchain"):     0600,

		// not in .bin
		filepath.Join("c", "c", "c"+extension):     0700,
		filepath.Join("c", "c", "Srclibtoolchain"): 0700,
	}
	for f, mode := range files {
		f = filepath.Join(tmpdir, f)
		if err := os.MkdirAll(filepath.Dir(f), 0700); err != nil {
			t.Fatal(err)
		}
		if err := ioutil.WriteFile(f, nil, mode); err != nil {
			t.Fatal(err)
		}
	}

	// Put a file symlink in srclib DIR path.
	oldp := filepath.Join(tmpdir, "a", "a", ".bin", "a"+extension)
	newp := filepath.Join(tmpdir, "link")
	if err := os.Symlink(oldp, newp); err != nil {
		t.Fatal(err)
	}

	toolchains, err := List()
	if err != nil {
		t.Fatal(err)
	}

	got := toolchainPathsWithProgram(toolchains)
	want := []string{path.Join("a", "a")}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got toolchains %v, want %v", got, want)
	}
}

func toolchainPathsWithProgram(toolchains []*Info) []string {
	paths := make([]string, 0, len(toolchains))
	for _, toolchain := range toolchains {
		if toolchain.Program != "" {
			paths = append(paths, toolchain.Path)
		}
	}
	return paths
}
