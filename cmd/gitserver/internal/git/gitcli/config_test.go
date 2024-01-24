package gitcli

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
)

func TestConfigGetSetUnset(t *testing.T) {
	rcf := wrexec.NewNoOpRecordingCommandFactory()
	reposDir := t.TempDir()
	testValue := "value"

	// Make a new bare repo on disk.
	p := filepath.Join(reposDir, "repo", ".git")
	require.NoError(t, os.MkdirAll(p, os.ModePerm))
	dir := common.GitDir(p)

	cmd := exec.Command("git", "--bare", "init", p)
	dir.Set(cmd)
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	git := NewBackend(logtest.Scoped(t), rcf, dir, "repo")

	ctx := context.Background()

	testGetSetUnset := func(testKey string) {
		// No config set should return empty value and no error:
		{
			val, err := git.Config().Get(ctx, testKey)
			require.NoError(t, err)
			require.Equal(t, "", val)
		}

		// Check that set, get, unset, get workflow works:
		{
			err := git.Config().Set(ctx, testKey, testValue)
			require.NoError(t, err)
			val, err := git.Config().Get(ctx, testKey)
			require.NoError(t, err)
			require.Equal(t, testValue, val)
			err = git.Config().Unset(ctx, testKey)
			require.NoError(t, err)
			// Unsetting twice should not error.
			err = git.Config().Unset(ctx, testKey)
			require.NoError(t, err)
			val, err = git.Config().Get(ctx, testKey)
			require.NoError(t, err)
			require.Equal(t, "", val)
		}

		// Check that concurrent writes aren't a problem:
		// TODO: Currently is a problem, need to fix.
		// {
		// 	p := pool.New().WithErrors()
		// 	for i := 0; i < 5; i++ {
		// 		p.Go(func() error {
		// 			for i := 0; i < 50; i++ {
		// 				if err := git.Config().Set(ctx, testKey, testValue); err != nil {
		// 					return err
		// 				}
		// 			}
		// 			return nil
		// 		})
		// 	}
		// 	require.NoError(t, p.Wait())
		// }
	}

	t.Run("one section", func(t *testing.T) {
		testGetSetUnset("sourcegraph.test")
	})

	t.Run("with subsection", func(t *testing.T) {
		testGetSetUnset("sourcegraph.test.section")
	})
}
