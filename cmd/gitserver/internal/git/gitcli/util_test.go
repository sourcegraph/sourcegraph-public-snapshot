package gitcli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/git"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
)

func BackendWithRepoCommands(t *testing.T, cmds ...string) git.GitBackend {
	rcf := wrexec.NewNoOpRecordingCommandFactory()

	dir := RepoWithCommands(t, cmds...)

	return NewBackend(logtest.Scoped(t), rcf, dir, api.RepoName(t.Name()))
}

func RepoWithCommands(t *testing.T, cmds ...string) common.GitDir {
	reposDir := t.TempDir()

	// Make a new bare repo on disk.
	p := filepath.Join(reposDir, "repo")
	require.NoError(t, os.MkdirAll(p, os.ModePerm))
	dir := common.GitDir(filepath.Join(p, ".git"))

	// Prepare repo state:
	for _, cmd := range append(
		append([]string{"git init --initial-branch=master ."}, cmds...),
		// Promote the repo to a bare repo.
		"git config --bool core.bare true",
	) {
		out, err := gitserver.CreateGitCommand(p, "bash", "-c", cmd).CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to run git command %v. Output was:\n\n%s", cmd, out)
		}
	}

	return dir
}
