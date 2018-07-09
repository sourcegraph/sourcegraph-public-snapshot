package server

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestMigrate(t *testing.T) {
	empty, cleanup := tmpDir(t)
	defer cleanup()

	// Empty should only create the manifest
	migrate(empty)
	assertFiles(t, empty, "gitserver.json")
	migrate(empty) // no-op
	assertFiles(t, empty, "gitserver.json")
}

func TestMigrateGitDir(t *testing.T) {
	root, cleanup := tmpDir(t)
	defer cleanup()

	// All non .git paths should become .git
	mkFiles(t, root,
		"github.com/foo/bar/HEAD",
		"github.com/foo/baz/.git/HEAD",

		"example.com/repo/HEAD",

		"example.org/repo/.git/HEAD",
		"example.org/repo/subrepo/.git/HEAD",
		"example.org/repo/subrepo2/HEAD",

		// This shouldn't happen, but ensure we don't fail if a repos is at
		// both possible locations
		"example.gov/repo/HEAD",
		"example.gov/repo/.git/HEAD",
	)
	if err := migrateGitDir(root); err != nil {
		t.Fatal(err)
	}
	assertFiles(t, root,
		"github.com/foo/bar/.git/HEAD",
		"github.com/foo/baz/.git/HEAD",

		"example.com/repo/.git/HEAD",

		"example.org/repo/.git/HEAD",
		"example.org/repo/subrepo/.git/HEAD",
		"example.org/repo/subrepo2/.git/HEAD",

		"example.gov/repo/.git/HEAD",
		"example.gov/repo/.git/.git/HEAD",
	)
}

func tmpDir(t *testing.T) (string, func()) {
	t.Helper()
	dir, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	return dir, func() { os.RemoveAll(dir) }
}

func mkFiles(t *testing.T, root string, paths ...string) {
	t.Helper()
	for _, p := range paths {
		if err := os.MkdirAll(filepath.Join(root, filepath.Dir(p)), os.ModePerm); err != nil {
			t.Fatal(err)
		}
		fd, err := os.OpenFile(filepath.Join(root, p), os.O_RDONLY|os.O_CREATE, 0666)
		if err != nil {
			t.Fatal(err)
		}
		if err := fd.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func assertFiles(t *testing.T, root string, want ...string) {
	t.Helper()
	notfound := make(map[string]struct{})
	for _, p := range want {
		notfound[p] = struct{}{}
	}
	var unwanted []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if _, ok := notfound[rel]; ok {
			delete(notfound, rel)
		} else {
			unwanted = append(unwanted, rel)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	if len(notfound) > 0 {
		var paths []string
		for p := range notfound {
			paths = append(paths, p)
		}
		sort.Strings(paths)
		t.Errorf("did not find expected paths: %s", strings.Join(paths, " "))
	}
	if len(unwanted) > 0 {
		sort.Strings(unwanted)
		t.Errorf("found unexpected paths: %s", strings.Join(unwanted, " "))
	}
}
