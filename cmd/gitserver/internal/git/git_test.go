package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
)

func TestMakeBareRepo(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	require.NoError(t, MakeBareRepo(ctx, dir))

	// Now verify we created a valid repo.
	c := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	c.Dir = dir
	out, err := c.CombinedOutput()
	require.NoError(t, err)
	require.Equal(t, "HEAD\n", string(out))
}

func TestRemoveBadRefs(t *testing.T) {
	dir := t.TempDir()
	gitDir := common.GitDir(filepath.Join(dir, ".git"))

	cmd := func(name string, arg ...string) string {
		t.Helper()
		return runCmd(t, dir, name, arg...)
	}
	wantCommit := makeSingleCommitRepo(cmd)

	for _, name := range []string{"HEAD", "head", "Head", "HeAd"} {
		// Tag
		cmd("git", "tag", name)

		if dontWant := cmd("git", "rev-parse", "HEAD"); dontWant == wantCommit {
			t.Logf("WARNING: git tag %s failed to produce ambiguous output: %s", name, dontWant)
		}

		if err := RemoveBadRefs(context.Background(), gitDir); err != nil {
			t.Fatal(err)
		}

		if got := cmd("git", "rev-parse", "HEAD"); got != wantCommit {
			t.Fatalf("git tag %s failed to be removed: %s", name, got)
		}

		// Ref
		if err := os.WriteFile(filepath.Join(dir, ".git", "refs", "heads", name), []byte(wantCommit), 0o600); err != nil {
			t.Fatal(err)
		}

		if dontWant := cmd("git", "rev-parse", "HEAD"); dontWant == wantCommit {
			t.Logf("WARNING: git ref %s failed to produce ambiguous output: %s", name, dontWant)
		}

		if err := RemoveBadRefs(context.Background(), gitDir); err != nil {
			t.Fatal(err)
		}

		if got := cmd("git", "rev-parse", "HEAD"); got != wantCommit {
			t.Fatalf("git ref %s failed to be removed: %s", name, got)
		}
	}
}

// makeSingleCommitRepo make create a new repo with a single commit and returns
// the HEAD SHA
func makeSingleCommitRepo(cmd func(string, ...string) string) string {
	// Setup a repo with a commit so we can see if we can clone it.
	cmd("git", "init", ".")
	cmd("sh", "-c", "echo hello world > hello.txt")
	return addCommitToRepo(cmd)
}

// addCommitToRepo adds a commit to the repo at the current path.
func addCommitToRepo(cmd func(string, ...string) string) string {
	// Setup a repo with a commit so we can see if we can clone it.
	cmd("git", "add", "hello.txt")
	cmd("git", "commit", "-m", "hello")
	return cmd("git", "rev-parse", "HEAD")
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
