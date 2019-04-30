package server

import (
	"context"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
)

const (
	testRepoA = "testrepo-A"
	testRepoB = "testrepo-B"
	testRepoC = "testrepo-C"
)

func TestCleanupInactive(t *testing.T) {
	root, err := ioutil.TempDir("", "gitserver-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	repoA := path.Join(root, testRepoA, ".git")
	cmd := exec.Command("git", "--bare", "init", repoA)
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	repoC := path.Join(root, testRepoC, ".git")
	if err := os.MkdirAll(repoC, os.ModePerm); err != nil {
		t.Fatal(err)
	}

	s := &Server{ReposDir: root, DeleteStaleRepositories: true}
	s.Handler() // Handler as a side-effect sets up Server
	s.cleanupRepos()

	if _, err := os.Stat(repoA); os.IsNotExist(err) {
		t.Error("expected repoA not to be removed")
	}
	if _, err := os.Stat(repoC); err == nil {
		t.Error("expected corrupt repoC to be removed during clean up")
	}
}

func TestCleanupExpired(t *testing.T) {
	root, err := ioutil.TempDir("", "gitserver-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	repoA := path.Join(root, testRepoA, ".git")
	cmd := exec.Command("git", "--bare", "init", repoA)
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	repoB := path.Join(root, testRepoB, ".git")
	cmd = exec.Command("git", "--bare", "init", repoB)
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	remote := path.Join(root, testRepoC, ".git")
	cmd = exec.Command("git", "--bare", "init", remote)
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	origRepoRemoteURL := repoRemoteURL
	repoRemoteURL = func(ctx context.Context, dir string) (string, error) {
		return remote, nil
	}
	defer func() { repoRemoteURL = origRepoRemoteURL }()

	atime, err := os.Stat(filepath.Join(repoA, "HEAD"))
	if err != nil {
		t.Fatal(err)
	}
	cmd = exec.Command("git", "config", "--add", "sourcegraph.recloneTimestamp", strconv.FormatInt(time.Now().Add(-(2*repoTTL)).Unix(), 10))
	cmd.Dir = repoB
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	s := &Server{ReposDir: root}
	s.Handler() // Handler as a side-effect sets up Server
	s.cleanupRepos()

	fi, err := os.Stat(filepath.Join(repoA, "HEAD"))
	if err != nil {
		// repoA should still exist.
		t.Fatal(err)
	}
	if atime.ModTime().Before(fi.ModTime()) {
		// repoA should not have been recloned.
		t.Error("expected repoA to not be modified")
	}
	fi, err = os.Stat(repoB)
	if err != nil {
		// repoB should still exist after being recloned.
		t.Fatal(err)
	}
	// Expect the repo to be recloned hand have a recent mod time.
	ti := time.Now().Add(-repoTTL)
	if fi.ModTime().Before(ti) {
		t.Error("expected repoB to be recloned during clean up")
	}
}

func TestCleanupOldLocks(t *testing.T) {
	root, cleanup := tmpDir(t)
	defer cleanup()

	// Only recent lock files should remain.
	mkFiles(t, root,
		"github.com/foo/empty/.git/HEAD",

		"github.com/foo/freshconfiglock/.git/HEAD",
		"github.com/foo/freshconfiglock/.git/config.lock",

		"github.com/foo/freshpacked/.git/HEAD",
		"github.com/foo/freshpacked/.git/packed-refs.lock",

		"github.com/foo/staleconfiglock/.git/HEAD",
		"github.com/foo/staleconfiglock/.git/config.lock",

		"github.com/foo/stalepacked/.git/HEAD",
		"github.com/foo/stalepacked/.git/packed-refs.lock",

		"github.com/foo/refslock/.git/HEAD",
		"github.com/foo/refslock/.git/refs/heads/fresh",
		"github.com/foo/refslock/.git/refs/heads/fresh.lock",
		"github.com/foo/refslock/.git/refs/heads/stale",
		"github.com/foo/refslock/.git/refs/heads/stale.lock",
	)

	chtime := func(p string, age time.Duration) {
		err := os.Chtimes(filepath.Join(root, p), time.Now().Add(-age), time.Now().Add(-age))
		if err != nil {
			t.Fatal(err)
		}
	}
	chtime("github.com/foo/staleconfiglock/.git/config.lock", time.Hour)
	chtime("github.com/foo/stalepacked/.git/packed-refs.lock", 2*time.Hour)
	chtime("github.com/foo/refslock/.git/refs/heads/stale.lock", 2*time.Hour)

	s := &Server{ReposDir: root}
	s.Handler() // Handler as a side-effect sets up Server
	s.cleanupRepos()

	assertPaths(t, root,
		"github.com/foo/empty/.git/HEAD",
		"github.com/foo/empty/.git/info/attributes",

		"github.com/foo/freshconfiglock/.git/HEAD",
		"github.com/foo/freshconfiglock/.git/config.lock",
		"github.com/foo/freshconfiglock/.git/info/attributes",

		"github.com/foo/freshpacked/.git/HEAD",
		"github.com/foo/freshpacked/.git/packed-refs.lock",
		"github.com/foo/freshpacked/.git/info/attributes",

		"github.com/foo/staleconfiglock/.git/HEAD",
		"github.com/foo/staleconfiglock/.git/info/attributes",

		"github.com/foo/stalepacked/.git/HEAD",
		"github.com/foo/stalepacked/.git/info/attributes",

		"github.com/foo/refslock/.git/HEAD",
		"github.com/foo/refslock/.git/refs/heads/fresh",
		"github.com/foo/refslock/.git/refs/heads/fresh.lock",
		"github.com/foo/refslock/.git/refs/heads/stale",
		"github.com/foo/refslock/.git/info/attributes",
	)
}

func TestSetupAndClearTmp(t *testing.T) {
	root, cleanup := tmpDir(t)
	defer cleanup()

	s := &Server{ReposDir: root}

	// All non .git paths should become .git
	mkFiles(t, root,
		"github.com/foo/baz/.git/HEAD",
		"example.org/repo/.git/HEAD",

		// Needs to be deleted
		".tmp/foo",
		".tmp/baz/bam",

		// Older tmp cleanups that failed
		".tmp-old123/foo",
	)

	tmp, err := s.SetupAndClearTmp()
	if err != nil {
		t.Fatal(err)
	}

	// Straight after cleaning .tmp should be empty
	assertPaths(t, filepath.Join(root, ".tmp"), ".")

	// tmp should exist
	if info, err := os.Stat(tmp); err != nil {
		t.Fatal(err)
	} else if !info.IsDir() {
		t.Fatal("tmpdir is not a dir")
	}

	// tmp should be on the same mount as root, ie root is parent.
	if filepath.Dir(tmp) != root {
		t.Fatalf("tmp is not under root: tmp=%s root=%s", tmp, root)
	}

	// Wait until async cleaning is done
	for i := 0; i < 1000; i++ {
		found := false
		files, err := ioutil.ReadDir(s.ReposDir)
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range files {
			found = found || strings.HasPrefix(f.Name(), ".tmp-old")
		}
		if !found {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Only files should be the repo files
	assertPaths(t, root,
		"github.com/foo/baz/.git/HEAD",
		"example.org/repo/.git/HEAD",
		".tmp",
	)
}

func TestSetupAndClearTmp_Empty(t *testing.T) {
	root, cleanup := tmpDir(t)
	defer cleanup()

	s := &Server{ReposDir: root}

	_, err := s.SetupAndClearTmp()
	if err != nil {
		t.Fatal(err)
	}

	// No files, just the empty .tmp dir should exist
	assertPaths(t, root, ".tmp")
}

func TestRemoveRepoDirectory(t *testing.T) {
	root, cleanup := tmpDir(t)
	defer cleanup()

	mkFiles(t, root,
		"github.com/foo/baz/.git/HEAD",
		"github.com/foo/survior/.git/HEAD",
		"github.com/bam/bam/.git/HEAD",
		"example.com/repo/.git/HEAD",
	)
	s := &Server{
		ReposDir: root,
	}

	// Remove everything but github.com/foo/survior
	for _, d := range []string{
		"github.com/foo/baz/.git",
		"github.com/bam/bam/.git",
		"example.com/repo/.git",
	} {
		if err := s.removeRepoDirectory(filepath.Join(root, d)); err != nil {
			t.Fatalf("failed to remove %s: %s", d, err)
		}
	}

	assertPaths(t, root,
		"github.com/foo/survior/.git/HEAD",
		".tmp",
	)
}

func TestRemoveRepoDirectory_Empty(t *testing.T) {
	root, cleanup := tmpDir(t)
	defer cleanup()

	mkFiles(t, root,
		"github.com/foo/baz/.git/HEAD",
	)
	s := &Server{
		ReposDir: root,
	}

	if err := s.removeRepoDirectory(filepath.Join(root, "github.com/foo/baz/.git")); err != nil {
		t.Fatal(err)
	}

	assertPaths(t, root,
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

func TestFreeUpSpace(t *testing.T) {
	t.Run("no error if no space requested and no repos", func(t *testing.T) {
		s := &Server{}
		if err := s.freeUpSpace(0); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("error if space requested and no repos", func(t *testing.T) {
		s := &Server{}
		if err := s.freeUpSpace(1); err == nil {
			t.Fatal("want error")
		}
	})
	t.Run("oldest repo gets removed to free up space", func(t *testing.T) {
		// Set up.
		rd, err := ioutil.TempDir("", "freeUpSpace")
		if err != nil {
			t.Fatal(err)
		}
		r1 := filepath.Join(rd, "repo1")
		r2 := filepath.Join(rd, "repo2")
		if err := makeFakeRepo(r1, 1000); err != nil {
			t.Fatal(err)
		}
		if err := makeFakeRepo(r2, 1000); err != nil {
			t.Fatal(err)
		}
		m1, err := gitDirModTime(filepath.Join(r1, ".git"))
		if err != nil {
			t.Fatal(err)
		}
		m2, err := gitDirModTime(filepath.Join(r2, ".git"))
		if err != nil {
			t.Fatal(err)
		}
		if m1.Equal(m2) || m1.After(m2) {
			t.Fatalf("expected repo1 to be created before repo2, got mod times %v and %v", m1, m2)
		}

		// Run.
		s := Server{
			ReposDir: rd,
		}
		if err := s.freeUpSpace(1000); err != nil {
			t.Fatal(err)
		}

		// Check.
		files, err := ioutil.ReadDir(rd)
		if err != nil {
			t.Fatal(err)
		}
		if len(files) != 1 {
			t.Fatalf("got %d items in %s, want exactly 1", len(files), rd)
		}
		if files[0].Name() != "repo2" {
			t.Errorf("name of only item in repos dir is %q, want repo2", files[0].Name())
		}
		rds, err := dirSize(rd)
		if err != nil {
			t.Fatal(err)
		}
		wantSize := int64(1000)
		if rds > wantSize {
			t.Errorf("repo dir size is %d, want no more than %d", rds, wantSize)
		}
	})
}

func makeFakeRepo(d string, sizeBytes int) error {
	gd := filepath.Join(d, ".git")
	if err := os.MkdirAll(gd, 0700); err != nil {
		return errors.Wrap(err, "creating .git dir and any parents")
	}
	if err := ioutil.WriteFile(filepath.Join(gd, "HEAD"), nil, 0666); err != nil {
		return errors.Wrap(err, "creating HEAD file")
	}
	if err := ioutil.WriteFile(filepath.Join(gd, "space_eater"), make([]byte, sizeBytes), 0666); err != nil {
		return errors.Wrapf(err, "writing to space_eater file")
	}
	return nil
}
