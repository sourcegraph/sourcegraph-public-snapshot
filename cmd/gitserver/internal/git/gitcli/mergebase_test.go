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
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))

		_, err = backend.MergeBase(ctx, "notfound", "master")
		require.Error(t, err)
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))
	})
}

func TestGitCLIBackend_MergeBaseOctopus(t *testing.T) {
	ctx := context.Background()

	t.Run("resolves", func(t *testing.T) {
		// Prepare repo state:
		// Structure:
		// 1 - 5 master
		// \ - 2 b2
		//  \ - 3 - 4 b3
		// Expected merge base of {master, b2, b3}: 1 (aka testbase)
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

			"git checkout -b b3",
			"echo line3 >> f",
			"git add f",
			"git commit -m foo --author='Foo Author <foo@sourcegraph.com>'",
			"echo line4 >> f",
			"git add f",
			"git commit -m foo --author='Foo Author <foo@sourcegraph.com>'",

			"git checkout master",
			"echo line2 > f",
			"git add f",
			"git commit -m foo --author='Foo Author <foo@sourcegraph.com>'",
		)

		wantSHA, err := backend.ResolveRevision(ctx, "testbase")
		require.NoError(t, err)
		require.NotEmpty(t, wantSHA)

		base, err := backend.MergeBaseOctopus(ctx, "master", "b2", "b3")
		require.NoError(t, err)
		require.Equal(t, wantSHA, base)
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
			"git checkout --orphan b3",
			"echo line3 >> f",
			"git add f",
			"git commit -m foo --author='Foo Author <foo@sourcegraph.com>'",
			"git checkout master",
		)

		base, err := backend.MergeBaseOctopus(ctx, "master", "b2", "b3")
		require.NoError(t, err)
		require.Equal(t, api.CommitID(""), base)
	})
	t.Run("not found revspec", func(t *testing.T) {
		// Prepare repo state:
		backend := BackendWithRepoCommands(t,
			"echo line1 > f",
			"git add f",
			"git commit -m foo --author='Foo Author <foo@sourcegraph.com>'",
		)

		// Last revspec not found
		_, err := backend.MergeBaseOctopus(ctx, "master", "notfound")
		require.Error(t, err)
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))
		require.Equal(t, "notfound", err.(*gitdomain.RevisionNotFoundError).Spec)

		// First revspec not found
		_, err = backend.MergeBaseOctopus(ctx, "notfound", "master")
		require.Error(t, err)
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))
		require.Equal(t, "notfound", err.(*gitdomain.RevisionNotFoundError).Spec)
	})
	t.Run("less than two revspecs", func(t *testing.T) {
		// Prepare repo state:
		backend := BackendWithRepoCommands(t)

		_, err := backend.MergeBaseOctopus(ctx, "master")
		require.Error(t, err)
		require.ErrorContains(t, err, "at least two revspecs must be given")
	})
}
