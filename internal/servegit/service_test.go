package servegit

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// numTestCommits determines the number of files/commits/tags to create for
// the local test repo. The value of 25 causes clonev1 and clonev2 to use gzip
// compression but shallow to be uncompressed. The value of 10 does not
// trigger this same behavior.
const numTestCommits = 25

func TestGitServiceHandler(t *testing.T) {
	root := tmpDir(t)
	repo := filepath.Join(root, "testrepo")

	// Setup a repo with a commit so we can add bad refs
	runCmd(t, root, "git", "init", repo)

	for i := 0; i < numTestCommits; i++ {
		runCmd(t, repo, "sh", "-c", fmt.Sprintf("echo hello world > hello-%d.txt", i+1))
		runCmd(t, repo, "git", "add", fmt.Sprintf("hello-%d.txt", i+1))
		runCmd(t, repo, "git", "commit", "-m", fmt.Sprintf("c%d", i+1))
		runCmd(t, repo, "git", "tag", fmt.Sprintf("v%d", i+1))
	}

	ts := httptest.NewServer(&gitServiceHandler{
		Dir: func(s string) string {
			return filepath.Join(root, s, ".git")
		},
		Debug: discardLogger,
	})
	defer ts.Close()

	t.Run("404", func(t *testing.T) {
		c := exec.Command("git", "clone", ts.URL+"/doesnotexist")
		c.Dir = tmpDir(t)
		b, err := c.CombinedOutput()
		if !bytes.Contains(b, []byte("repository not found")) {
			t.Fatal("expected clone to fail with repository not found", string(b), err)
		}
	})

	cloneURL := ts.URL + "/testrepo"

	t.Run("clonev1", func(t *testing.T) {
		runCmd(t, tmpDir(t), "git", "-c", "protocol.version=1", "clone", cloneURL)
	})

	cloneV2 := []struct {
		Name string
		Args []string
	}{{
		"clonev2",
		[]string{},
	}, {
		"shallow",
		[]string{"--depth=1"},
	}}

	for _, tc := range cloneV2 {
		t.Run(tc.Name, func(t *testing.T) {
			args := []string{"-c", "protocol.version=2", "clone"}
			args = append(args, tc.Args...)
			args = append(args, cloneURL)

			c := exec.Command("git", args...)
			c.Dir = tmpDir(t)
			c.Env = []string{
				"GIT_TRACE_PACKET=1",
			}
			b, err := c.CombinedOutput()
			if err != nil {
				t.Fatalf("command failed: %s\nOutput: %s", err, b)
			}

			// This is the same test done by git's tests for checking if the
			// server is using protocol v2.
			if !bytes.Contains(b, []byte("git< version 2")) {
				t.Fatalf("protocol v2 not used by server. Output:\n%s", b)
			}
		})
	}
}

func runCmd(t *testing.T, dir string, cmd string, arg ...string) string {
	t.Helper()
	c := exec.Command(cmd, arg...)
	c.Dir = dir
	c.Env = []string{
		"GIT_COMMITTER_NAME=a",
		"GIT_COMMITTER_EMAIL=a@a.com",
		"GIT_AUTHOR_NAME=a",
		"GIT_AUTHOR_EMAIL=a@a.com",
	}
	b, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s failed: %s\nOutput: %s", cmd, strings.Join(arg, " "), err, b)
	}
	return string(b)
}

func tmpDir(t *testing.T) string {
	t.Helper()
	dir, err := ioutil.TempDir("", filepath.Base(t.Name()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}
