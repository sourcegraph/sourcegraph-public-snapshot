package gitcli

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGitCLIBackend_LatestCommitTimestamp(t *testing.T) {
	ctx := context.Background()

	// Prepare repo state:
	backend := BackendWithRepoCommands(t,
		"echo line1 > f",
		"git add f",
		`GIT_COMMITTER_DATE="2015-01-01 00:00 Z" git commit -m foo --author='Foo Author <foo@sourcegraph.com>'`,
	)

	have, err := backend.LatestCommitTimestamp(ctx)
	require.NoError(t, err)
	require.Equal(t, time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC), have)
}
