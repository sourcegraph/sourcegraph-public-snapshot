package gitcli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestGitCLIBackend_ResolveRevision(t *testing.T) {
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
			"git tag v1.0.0",
			"echo $(git cat-file commit f372e36a91bc35e5d99df8be435bdcb1f0660bc5) > /tmp/catfile.test",
		)

		commit, err := backend.ResolveRevision(ctx, "HEAD")
		require.NoError(t, err)
		require.Equal(t, api.CommitID("f372e36a91bc35e5d99df8be435bdcb1f0660bc5"), commit)
		// @ is an alias for HEAD.
		commit, err = backend.ResolveRevision(ctx, "@")
		require.NoError(t, err)
		require.Equal(t, api.CommitID("f372e36a91bc35e5d99df8be435bdcb1f0660bc5"), commit)

		// Empty resolves HEAD, too:
		commit, err = backend.ResolveRevision(ctx, "")
		require.NoError(t, err)
		require.Equal(t, api.CommitID("f372e36a91bc35e5d99df8be435bdcb1f0660bc5"), commit)

		// Resolve commit:
		commit, err = backend.ResolveRevision(ctx, "f372e36a91bc35e5d99df8be435bdcb1f0660bc5")
		require.NoError(t, err)
		require.Equal(t, api.CommitID("f372e36a91bc35e5d99df8be435bdcb1f0660bc5"), commit)
		// Unknown commit:
		_, err = backend.ResolveRevision(ctx, "dfcb84e522cab3c0b307a70917604c6d3da00dc8")
		require.Error(t, err)
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))

		// Resolve abbrev commit:
		commit, err = backend.ResolveRevision(ctx, "f372e36")
		require.NoError(t, err)
		require.Equal(t, api.CommitID("f372e36a91bc35e5d99df8be435bdcb1f0660bc5"), commit)
		// Unknown abbrev commit:
		_, err = backend.ResolveRevision(ctx, "dfcb84e5")
		require.Error(t, err)
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))

		// Resolve ref:
		commit, err = backend.ResolveRevision(ctx, "refs/heads/master")
		require.NoError(t, err)
		require.Equal(t, api.CommitID("f372e36a91bc35e5d99df8be435bdcb1f0660bc5"), commit)
		commit, err = backend.ResolveRevision(ctx, "heads/master")
		require.NoError(t, err)
		require.Equal(t, api.CommitID("f372e36a91bc35e5d99df8be435bdcb1f0660bc5"), commit)
		commit, err = backend.ResolveRevision(ctx, "master")
		require.NoError(t, err)
		require.Equal(t, api.CommitID("f372e36a91bc35e5d99df8be435bdcb1f0660bc5"), commit)
		commit, err = backend.ResolveRevision(ctx, "v1.0.0")
		require.NoError(t, err)
		require.Equal(t, api.CommitID("f372e36a91bc35e5d99df8be435bdcb1f0660bc5"), commit)

		// Unknown ref:
		_, err = backend.ResolveRevision(ctx, "refs/heads/notfound")
		require.Error(t, err)
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))
		_, err = backend.ResolveRevision(ctx, "notfound")
		require.Error(t, err)
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))

		// Resolve object that is not a commit: (this is the tree object of f372e36a91bc35e5d99df8be435bdcb1f0660bc5)
		_, err = backend.ResolveRevision(ctx, "92cb0143f5166452f2d45ed974a818749bc4a13f")
		require.Error(t, err)
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))

		// :file1 gets the object ID of the file called file1 at HEAD.
		// We don't allow that, since it leaks the existence of the file.
		_, err = backend.ResolveRevision(ctx, ":file1")
		require.Error(t, err)
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))

		// HEAD:file1 gets the object ID of the file called file1 at HEAD.
		// We don't allow that, since it leaks the existence of the file.
		_, err = backend.ResolveRevision(ctx, "HEAD:file1")
		require.Error(t, err)
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))

		// :/foo gets a commit by commit message, but we don't want that.
		_, err = backend.ResolveRevision(ctx, ":/foo")
		require.Error(t, err)
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))
		// HEAD^{/foo} is the same as the above.
		// TODO: This currently passes, but it shouldn't need to.
		// _, err = backend.ResolveRevision(ctx, "HEAD^{/foo}")
		// require.Error(t, err)
		// require.True(t, errors.HasType(err, &gitdomain.RevisionNotFoundError{}))

		// Ranges:
		commit, err = backend.ResolveRevision(ctx, "master..b2")
		require.NoError(t, err)
		require.Equal(t, api.CommitID("a8994413dc8109087150c7932b162a4713e6d59a"), commit)
		// Not found range:
		_, err = backend.ResolveRevision(ctx, "master..notfound")
		require.Error(t, err)
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))
	})

	t.Run("HEAD in empty repo", func(t *testing.T) {
		backend := BackendWithRepoCommands(t)

		_, err := backend.ResolveRevision(ctx, "HEAD")
		require.Error(t, err)
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))
	})
}
