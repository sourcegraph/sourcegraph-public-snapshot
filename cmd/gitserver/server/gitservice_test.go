package server

import (
	"bytes"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGitServiceHandler(t *testing.T) {
	root := tmpDir(t)
	repo := filepath.Join(root, "testrepo")

	// Setup a repo with a commit so we can add bad refs
	runCmd(t, root, "git", "init", repo)
	runCmd(t, repo, "sh", "-c", "echo hello world > hello.txt")
	runCmd(t, repo, "git", "add", "hello.txt")
	runCmd(t, repo, "git", "commit", "-m", "hello")

	ts := httptest.NewServer(&gitServiceHandler{
		Dir: func(s string) string {
			return filepath.Join(root, s, ".git")
		},
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
