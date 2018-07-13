package server

import (
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
	repoB := path.Join(root, testRepoB, ".git")
	cmd = exec.Command("git", "--bare", "init", repoB)
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	filepath.Walk(repoB, func(p string, _ os.FileInfo, _ error) error {
		// Rollback the mtime for these files to simulate an old repo.
		return os.Chtimes(p, time.Now().Add(-inactiveRepoTTL-time.Hour), time.Now().Add(-inactiveRepoTTL-time.Hour))
	})

	s := &Server{ReposDir: root}
	s.Handler() // Handler as a side-effect sets up Server
	s.cleanupRepos()

	if _, err := os.Stat(repoA); os.IsNotExist(err) {
		t.Error("expected repoA not to be removed")
	}
	if _, err := os.Stat(repoB); err == nil {
		t.Error("expected repoB to be removed during clean up")
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

	atime, err := os.Stat(repoA)
	if err != nil {
		t.Fatal(err)
	}
	filepath.Walk(repoB, func(p string, f os.FileInfo, _ error) error {
		if f.Name() == "HEAD" {
			return nil
		}
		// Rollback the mtime for these files to simulate an old repo.
		return os.Chtimes(p, time.Now().Add(-repoTTL-time.Hour), time.Now().Add(-repoTTL-time.Hour))
	})

	s := &Server{ReposDir: root}
	s.Handler() // Handler as a side-effect sets up Server
	s.cleanupRepos()

	fi, err := os.Stat(repoA)
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
	assertFiles(t, filepath.Join(root, ".tmp"))

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
	assertFiles(t, root,
		"github.com/foo/baz/.git/HEAD",
		"example.org/repo/.git/HEAD")
}

func TestSetupAndClearTmp_Empty(t *testing.T) {
	root, cleanup := tmpDir(t)
	defer cleanup()

	s := &Server{ReposDir: root}

	tmp, err := s.SetupAndClearTmp()
	if err != nil {
		t.Fatal(err)
	}

	// No files, just the empty .tmp dir should exist
	assertFiles(t, root)

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
}
