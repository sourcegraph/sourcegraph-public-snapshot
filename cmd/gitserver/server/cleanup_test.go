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

const testRepoA = "testrepo-A"
const testRepoB = "testrepo-B"

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
	s.CleanupRepos()

	if _, err := os.Stat(repoA); os.IsNotExist(err) {
		t.Error("expected repoA not to be removed")
	}
	if _, err := os.Stat(repoB); err == nil {
		t.Error("expected repoB to be removed during clean up")
	}
}
