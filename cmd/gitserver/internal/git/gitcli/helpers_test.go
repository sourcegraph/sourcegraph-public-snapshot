package gitcli

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

func TestGitRepo(t *testing.T) {
	submoduleRepoPath := prepareGitRepo(t,
		// Add a basic file:
		"echo Hello World from the submodule > README.md",
		"git add README.md",
		`git commit -m "Add file"`,
	)

	repoPath := prepareGitRepo(t,
		// Add a basic file:
		"echo Hello World > README.md",
		"git add README.md",
		`git commit -m "Add file"`,
		// Add a nested file:
		"mkdir folder",
		"echo Hello Folder > folder/README.md",
		"git add folder",
		`git commit -m "Add folder"`,
		// Add a submodule:
		fmt.Sprintf("git -c protocol.file.allow=always submodule add %s sub", submoduleRepoPath),
		"git add .",
		"git commit -m Submodule",
	)

	headSHA := runGitCmd(t, repoPath, "git rev-parse HEAD")

	t.Run("read basic file", func(t *testing.T) {
		r, err := readObj(common.GitDir(repoPath), headSHA, "README.md")
		require.NoError(t, err)
		c, err := io.ReadAll(r)
		require.NoError(t, err)
		require.Equal(t, "Hello World\n", string(c))
	})

	t.Run("read nested file", func(t *testing.T) {
		r, err := readObj(common.GitDir(repoPath), headSHA, "folder/README.md")
		require.NoError(t, err)
		c, err := io.ReadAll(r)
		require.NoError(t, err)
		require.Equal(t, "Hello Folder\n", string(c))
	})

	t.Run("attempt to read folder", func(t *testing.T) {
		_, err := readObj(common.GitDir(repoPath), headSHA, "folder")
		require.Error(t, err)
		require.True(t, os.IsNotExist(err))
	})

	t.Run("attempt to read submodule", func(t *testing.T) {
		_, err := readObj(common.GitDir(repoPath), headSHA, "sub")
		require.Error(t, err)
		require.Equal(t, io.EOF, err)
	})
}

func prepareGitRepo(t *testing.T, cmds ...string) string {
	repoPath := t.TempDir()
	for _, cmd := range append([]string{"git init --initial-branch=master"}, cmds...) {
		runGitCmd(t, repoPath, cmd)
	}
	return repoPath
}

func runGitCmd(t *testing.T, repoPath string, cmd string) string {
	out, err := gitserver.CreateGitCommand(repoPath, "bash", "-c", cmd).CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run git command. Output was:\n\n%s", out)
	}
	return string(out)
}
