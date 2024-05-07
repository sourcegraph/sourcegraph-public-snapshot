package gitcli

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/git"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestGitCLIBackend_RawDiff(t *testing.T) {
	var f1Diff = []byte(`diff --git f f
index a29bdeb434d874c9b1d8969c40c42161b03fafdc..c0d0fb45c382919737f8d0c20aaf57cf89b74af8 100644
--- f
+++ f
@@ -1 +1,2 @@
 line1
+line2
`)
	var f2Diff = []byte(`diff --git f2 f2
new file mode 100644
index 0000000000000000000000000000000000000000..8a6a2d098ecaf90105f1cf2fa90fc4608bb08067
--- /dev/null
+++ f2
@@ -0,0 +1 @@
+line2
`)
	ctx := context.Background()

	// Prepare repo state:
	backend := BackendWithRepoCommands(t,
		"echo line1 > f",
		"git add f",
		"git commit -m foo --author='Foo Author <foo@sourcegraph.com>'",
		"git tag testbase",
		"echo line2 >> f",
		"git add f",
		"git commit -m foo --author='Foo Author <foo@sourcegraph.com>'",
	)

	t.Run("streams diff", func(t *testing.T) {
		r, err := backend.RawDiff(ctx, "testbase", "HEAD", git.GitDiffComparisonTypeOnlyInHead)
		require.NoError(t, err)
		diff, err := io.ReadAll(r)
		require.NoError(t, err)
		require.NoError(t, r.Close())
		require.Equal(t, string(f1Diff), string(diff))
	})
	t.Run("streams diff for path", func(t *testing.T) {
		// Prepare repo state:
		backend := BackendWithRepoCommands(t,
			"echo line1 > f",
			"git add f",
			"git commit -m foo --author='Foo Author <foo@sourcegraph.com>'",
			"git tag testbase",
			"echo line2 >> f2",
			"git add f2",
			"git commit -m foo --author='Foo Author <foo@sourcegraph.com>'",
		)

		r, err := backend.RawDiff(ctx, "testbase", "HEAD", git.GitDiffComparisonTypeOnlyInHead, "f2")
		require.NoError(t, err)
		diff, err := io.ReadAll(r)
		require.NoError(t, err)
		require.NoError(t, r.Close())
		// We expect only a diff for f2, not for f.
		require.Equal(t, string(f2Diff), string(diff))
	})
	t.Run("not found revspec", func(t *testing.T) {
		// Prepare repo state:
		backend := BackendWithRepoCommands(t,
			"echo line1 > f",
			"git add f",
			"git commit -m foo --author='Foo Author <foo@sourcegraph.com>'",
			"git tag test",
		)

		_, err := backend.RawDiff(ctx, "unknown", "test", git.GitDiffComparisonTypeOnlyInHead)
		require.Error(t, err)
		require.True(t, errors.HasType(err, &gitdomain.RevisionNotFoundError{}))

		_, err = backend.RawDiff(ctx, "test", "unknown", git.GitDiffComparisonTypeOnlyInHead)
		require.Error(t, err)
		require.True(t, errors.HasType(err, &gitdomain.RevisionNotFoundError{}))
	})
	// Verify that if the context is canceled, the reader returns an error.
	t.Run("context cancelation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
		t.Cleanup(cancel)

		r, err := backend.RawDiff(ctx, "testbase", "HEAD", git.GitDiffComparisonTypeOnlyInHead)
		require.NoError(t, err)

		cancel()

		_, err = io.ReadAll(r)
		require.Error(t, err)
		require.True(t, errors.Is(err, context.Canceled), "unexpected error: %v", err)

		require.True(t, errors.Is(r.Close(), context.Canceled), "unexpected error: %v", err)
	})
}
