package vfsutil

import (
	"path"
	"reflect"
	"sort"
	"testing"
)

func TestSortedPathsStat(t *testing.T) {
	paths := sortedPaths{
		"/a",
		"/bb",
		"/b/a",
		"/b/b",
		"/b/c/d",
		"/b-c", // corner case: - is before /
		"/b/c-d",
		"/c",
	}
	sort.Strings(paths)

	// Test stat works for all file paths
	for _, p := range paths {
		fi, err := paths.Stat(p)
		if err != nil {
			t.Errorf("failed to stat %s: %s", p, err)
			continue
		}
		if fi.IsDir() {
			t.Errorf("path should not be a dir: %s", p)
		}
		if fi.Name() != path.Base(p) {
			t.Errorf("path %s name is %v, expected %v", p, fi.Name(), path.Base(p))
		}
	}

	// Test stat works on all dirs
	for _, p := range paths {
		for p != "/" {
			p = path.Dir(p)
			fi, err := paths.Stat(p)
			if err != nil {
				t.Errorf("failed to stat %s: %s", p, err)
				continue
			}
			if !fi.IsDir() {
				t.Errorf("path should be a dir: %s", p)
			}
			if fi.Name() != path.Base(p) {
				t.Errorf("path %s name is %v, expected %v", p, fi.Name(), path.Base(p))
			}
		}
	}

	// Test stat fails on missing paths
	missing := []string{
		"/aa",
		"/a-",
		"/b/aa",
		"/b/a/a",
		"/b/c-",
		"/b-",
		"/z",
	}
	for _, p := range missing {
		_, err := paths.Stat(p)
		if err == nil {
			t.Errorf("path found, but should not exist: %s", p)
		}
	}
}

func TestSortedPathsReadDir(t *testing.T) {
	paths := sortedPaths{
		"/a",
		"/bb",
		"/b/a",
		"/b/b",
		"/b/c/d",
		"/b-c", // corner case: - is before /
		"/b/c-d",
		"/c",
	}
	sort.Strings(paths)

	// Test ReadDir fails for all file paths
	for _, p := range paths {
		_, err := paths.ReadDir(p)
		if err == nil {
			t.Errorf("expected ReadDir to fail for file %s", p)
		}
	}

	// Check ReadDir works on all the dirs
	type child struct {
		Name  string
		IsDir bool
	}
	cases := map[string][]child{
		"/": []child{
			{Name: "a"},
			{Name: "b-c"},
			{Name: "b", IsDir: true},
			{Name: "bb"},
			{Name: "c"},
		},
		"/b": []child{
			{Name: "a"},
			{Name: "b"},
			{Name: "c-d"},
			{Name: "c", IsDir: true},
		},
		"/b/c": []child{{Name: "d"}},
	}
	for p, want := range cases {
		fis, err := paths.ReadDir(p)
		if err != nil {
			t.Errorf("ReadDir failed on %s: %s", p, err)
			continue
		}
		// We are expecting the response to be sorted
		got := make([]child, len(fis))
		for i, fi := range fis {
			got[i] = child{Name: fi.Name(), IsDir: fi.IsDir()}
		}
		if !reflect.DeepEqual(want, got) {
			t.Errorf("ReadDir(%s)\ngot  %#+v\nwant %#+v", p, got, want)
		}
	}

	// Test ReadDir fails on missing dirs
	missing := []string{
		"/aa",
		"/a-",
		"/b/aa",
		"/b/a/a",
		"/b/c-",
		"/b-",
		"/z",
	}
	for _, p := range missing {
		_, err := paths.ReadDir(p)
		if err == nil {
			t.Errorf("ReadDir(%s) should fail on missing path", p)
		}
	}
}
