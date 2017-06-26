package server

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
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

	repoA := path.Join(root, testRepoA)
	os.Mkdir(repoA, os.ModePerm)
	cmd := exec.Command("git", "--bare", "init")
	cmd.Dir = repoA
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	repoB := path.Join(root, testRepoB)
	os.Mkdir(repoB, os.ModePerm)
	cmd = exec.Command("git", "--bare", "init")
	cmd.Dir = repoB
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	filepath.Walk(repoB, func(p string, _ os.FileInfo, _ error) error {
		// Rollback the mtime for these files to simulate an old repo.
		return os.Chtimes(p, time.Now().Add(-inactiveRepoTTL-time.Hour), time.Now().Add(-inactiveRepoTTL-time.Hour))
	})

	s := &Server{ReposDir: root}
	s.cloning = make(map[string]struct{})
	s.CleanupRepos()

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

	repoA := path.Join(root, testRepoA)
	os.Mkdir(repoA, os.ModePerm)
	cmd := exec.Command("git", "--bare", "init")
	cmd.Dir = repoA
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	repoB := path.Join(root, testRepoB)
	os.Mkdir(repoB, os.ModePerm)
	cmd = exec.Command("git", "--bare", "init")
	cmd.Dir = repoB
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	remote := path.Join(root, testRepoC)
	os.Mkdir(remote, os.ModePerm)
	cmd = exec.Command("git", "--bare", "init")
	cmd.Dir = remote
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	restoreOriginMap := originMap[:]
	defer func() {
		originMap = restoreOriginMap
	}()
	originMap = append(originMap, prefixAndOrgin{Prefix: testRepoB, Origin: remote})

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
	s.cloning = make(map[string]struct{})
	s.CleanupRepos()

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
