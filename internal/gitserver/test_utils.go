package gitserver

import (
	"encoding/json"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
	root, err := os.MkdirTemp("", name)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		os.RemoveAll(root)
	})
	return root
}

func MustParseTime(layout, value string) time.Time {
	tm, err := time.Parse(layout, value)
	if err != nil {
		panic(err.Error())
	}
	return tm
}

// MakeGitRepository calls initGitRepository to create a new Git repository and returns a handle to
// it.
func MakeGitRepository(t *testing.T, cmds ...string) api.RepoName {
	t.Helper()
	dir := InitGitRepository(t, cmds...)
	repo := api.RepoName(filepath.Base(dir))
	return repo
}

// InitGitRepository initializes a new Git repository and runs commands in a new
// temporary directory (returned as dir).
// It also sets ClientMocks.LocalGitCommandReposDir for successful run of local git commands.
func InitGitRepository(t *testing.T, cmds ...string) string {
	t.Helper()
	root := CreateRepoDir(t)
	remotes := filepath.Join(root, "remotes")
	if err := os.MkdirAll(remotes, 0700); err != nil {
		t.Fatal(err)
	}
	dir, err := os.MkdirTemp(remotes, strings.ReplaceAll(t.Name(), "/", "__"))
	if err != nil {
		t.Fatal(err)
	}

	// setting git repo which is needed for successful run of git command against local file system
	ClientMocks.LocalGitCommandReposDir = remotes

	cmds = append([]string{"git init"}, cmds...)
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
	c.Env = []string{"GIT_CONFIG=" + path.Join(dir, ".git", "config")}
	return c
}

func AsJSON(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(b)
}
