package gitserver

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

// CreateRepoDir creates a repo directory for testing purposes.
// This includes creating a tmp dir and deleting it after test finishes running.
func CreateRepoDir(t *testing.T) string {
	return CreateRepoDirWithName(t, "")
}

// CreateRepoDirWithName creates a repo directory with a given name for testing purposes.
// This includes creating a tmp dir and deleting it after test finishes running.
func CreateRepoDirWithName(t *testing.T, name string) string {
	t.Helper()
	if name == "" {
		name = t.Name()
	}
	name = strings.ReplaceAll(name, "/", "-")
	root, err := os.MkdirTemp("", name)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		os.RemoveAll(root)
	})
	return root
}

// MakeGitRepositoryAndReturnDir calls initGitRepository to create a new Git repository and returns
// the repo name and directory.
func MakeGitRepositoryAndReturnDir(t *testing.T, cmds ...string) (api.RepoName, string) {
	t.Helper()
	dir := InitGitRepository(t, cmds...)
	repo := api.RepoName(filepath.Base(dir))
	return repo, dir
}

// InitGitRepository initializes a new Git repository and runs commands in a new
// temporary directory (returned as dir).
// It also sets ClientMocks.LocalGitCommandReposDir for successful run of local git commands.
func InitGitRepository(t *testing.T, cmds ...string) string {
	t.Helper()
	root := CreateRepoDir(t)
	remotes := filepath.Join(root, "remotes")
	if err := os.MkdirAll(remotes, 0o700); err != nil {
		t.Fatal(err)
	}
	dir, err := os.MkdirTemp(remotes, strings.ReplaceAll(t.Name(), "/", "__"))
	if err != nil {
		t.Fatal(err)
	}

	cmds = append([]string{"git init --initial-branch=master"}, cmds...)
	for _, cmd := range cmds {
		out, err := CreateGitCommand(dir, "bash", "-c", cmd).CombinedOutput()
		if err != nil {
			t.Fatalf("Command %q failed. Output was:\n\n%s", cmd, out)
		}
	}
	return dir
}

func CreateGitCommand(dir, name string, args ...string) *exec.Cmd {
	c := exec.Command(name, args...)
	c.Dir = dir
	c.Env = []string{
		"GIT_CONFIG=" + path.Join(dir, ".git", "config"),
		"GIT_COMMITTER_NAME=a",
		"GIT_COMMITTER_EMAIL=a@a.com",
		"GIT_COMMITTER_DATE=2006-01-02T15:04:05Z",
		"GIT_AUTHOR_NAME=a",
		"GIT_AUTHOR_EMAIL=a@a.com",
		"GIT_AUTHOR_DATE=2006-01-02T15:04:05Z",
	}
	if systemPath, ok := os.LookupEnv("PATH"); ok {
		c.Env = append(c.Env, "PATH="+systemPath)
	}
	return c
}
