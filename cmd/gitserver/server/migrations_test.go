package server

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

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

		// We will chmod the basedir so that HEAD can't be statted. This is to
		// test that our migration doesn't fail due to
		// it. https://github.com/sourcegraph/sourcegraph/issues/12234
		"naughty.com/repo/HEAD",
	)
	if err := os.Chmod(filepath.Join(root, "naughty.com/repo"), 0400); err != nil {
		t.Fatal(err)
	}
	(&Server{
		ReposDir: root,
		locker:   &RepositoryLocker{},
	}).migrateGitDir()
	assertPaths(t, root,
		"github.com/foo/bar/.git/HEAD",
		"github.com/foo/baz/.git/HEAD",

		"example.com/repo/.git/HEAD",

		"example.org/repo/.git/HEAD",
		"example.org/repo/subrepo/.git/HEAD",
		"example.org/repo/subrepo2/.git/HEAD",

		"example.gov/repo/.git/HEAD",
		"example.gov/repo/.git/.git/HEAD",

		"naughty.com/repo/.git/HEAD",

		".tmp",
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

// assertPaths checks that all paths under want exist. It excludes non-empty directories
func assertPaths(t *testing.T, root string, want ...string) {
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
			if empty, err := isEmptyDir(path); err != nil {
				t.Fatal(err)
			} else if !empty {
				return nil
			}
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

func isEmptyDir(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}
