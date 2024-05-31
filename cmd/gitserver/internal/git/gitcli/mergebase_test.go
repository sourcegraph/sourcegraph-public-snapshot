package gitcli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestGitCLIBackend_MergeBase(t *testing.T) {
	ctx := context.Background()

	t.Run("resolves", func(t *testing.T) {
		// Prepare repo state:
		backend := BackendWithRepoCommands(t,
			"echo line1 > f",
			"git add f",
			"git commit -m foo --author='Foo Author <foo@sourcegraph.com>'",
			"git tag testbase",
			"git checkout -b b2",
			"echo line2 >> f",
			"git add f",
			"git commit -m foo --author='Foo Author <foo@sourcegraph.com>'",
			"git checkout master",
			"echo line3 > h",
			"git add h",
			"git commit -m qux --author='Foo Author <foo@sourcegraph.com>'",
		)

		base, err := backend.MergeBase(ctx, "master", "b2")
		require.NoError(t, err)
		require.Equal(t, api.CommitID("3580f4105887559aa530eb2b1744f7cad676578a"), base)
	})
	t.Run("orphan branches", func(t *testing.T) {
		// Prepare repo state:
		backend := BackendWithRepoCommands(t,
			"echo line1 > f",
			"git add f",
			"git commit -m foo --author='Foo Author <foo@sourcegraph.com>'",
			"git checkout --orphan b2",
			"echo line2 >> f",
			"git add f",
			"git commit -m foo --author='Foo Author <foo@sourcegraph.com>'",
			"git checkout master",
		)

		base, err := backend.MergeBase(ctx, "master", "b2")
		require.NoError(t, err)
		require.Equal(t, api.CommitID(""), base)
	})
	t.Run("not found revspec", func(t *testing.T) {
		// Prepare repo state:
		backend := BackendWithRepoCommands(t,
			"echo line1 > f",
			"git add f",
			"git commit -m foo --author='Foo Author <foo@sourcegraph.com>'",
			"git checkout -b b2",
			"echo line2 >> f",
			"git add f",
			"git commit -m foo --author='Foo Author <foo@sourcegraph.com>'",
			"git checkout master",
		)

		_, err := backend.MergeBase(ctx, "master", "notfound")
		require.Error(t, err)
		require.True(t, errors.HasTypeGeneric[*gitdomain.RevisionNotFoundError](err))

		_, err = backend.MergeBase(ctx, "notfound", "master")
		require.Error(t, err)
		require.True(t, errors.HasTypeGeneric[*gitdomain.RevisionNotFoundError](err))
	})
}
