package gitcli

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestGitCLIBackend_RevAtTime(t *testing.T) {
	ctx := context.Background()

	t.Run("basic", func(t *testing.T) {
		backend := BackendWithRepoCommands(t,
			"echo line1 > f",
			"git add f",
			// 2f340b97899b26d3f889ec8048666559e4dca830
			"GIT_COMMITTER_DATE='2021-01-01T00:00:00Z' git commit -m foo",
			"git tag testbase",

			"git checkout -b b2",
			"echo line2 >> f",
			"git add f",
			// b30e87d684eb20c32a15f36fe4f191c3b78542d7
			"GIT_COMMITTER_DATE='2022-01-01T00:00:00Z' git commit -m bar",

			"git checkout master",
			"echo line3 > h",
			"git add h",
			// 26cbcb4ee4cb649293f22afc0cc9d3baed5eede5
			"GIT_COMMITTER_DATE='2023-01-01T00:00:00Z' git commit -m qux",
			"git tag v1.0.0",

			// ebadaea713c06e387c83947bbe6662475a366ffe
			"GIT_COMMITTER_DATE='2024-01-01T00:00:00Z' git merge b2",
		)

		// Target the first commit on master
		commit, err := backend.RevAtTime(ctx, "HEAD", time.Date(2021, 6, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		require.Equal(t, api.CommitID("2f340b97899b26d3f889ec8048666559e4dca830"), commit)

		// Target the second commit on master
		commit, err = backend.RevAtTime(ctx, "HEAD", time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		require.Equal(t, api.CommitID("26cbcb4ee4cb649293f22afc0cc9d3baed5eede5"), commit)

		// A date before the first commit returns an empty string
		commit, err = backend.RevAtTime(ctx, "HEAD", time.Date(1996, 6, 28, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		require.Equal(t, api.CommitID(""), commit)

		// If we traversed merged branches, this would target d5fe. Since we use --first-parent,
		// this targets the root commit.
		commit, err = backend.RevAtTime(ctx, "HEAD", time.Date(2022, 6, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		require.Equal(t, api.CommitID("2f340b97899b26d3f889ec8048666559e4dca830"), commit)

		// If we target the b2 branch specifically though, that's on the --first-parent history
		commit, err = backend.RevAtTime(ctx, "b2", time.Date(2022, 6, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		require.Equal(t, api.CommitID("b30e87d684eb20c32a15f36fe4f191c3b78542d7"), commit)

		// Targeting in the future is fine
		commit, err = backend.RevAtTime(ctx, "master", time.Date(2048, 6, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		require.Equal(t, api.CommitID("ebadaea713c06e387c83947bbe6662475a366ffe"), commit)

		// Invalid rev returns a rev not found error
		_, err = backend.RevAtTime(ctx, "noexist", time.Date(2048, 6, 1, 0, 0, 0, 0, time.UTC))
		require.Error(t, err)
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))

		// Invalid OID returns a rev not found error
		_, err = backend.RevAtTime(ctx, "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef", time.Date(2048, 6, 1, 0, 0, 0, 0, time.UTC))
		require.Error(t, err)
		require.True(t, errors.HasType[*gitdomain.RevisionNotFoundError](err))
	})

	t.Run("out of order commit date", func(t *testing.T) {
		backend := BackendWithRepoCommands(t,
			"echo line1 > f",
			"git add f",
			"GIT_COMMITTER_DATE='2022-01-01T00:00:00Z' git commit -m foo",
			"git tag testbase",

			"echo line2 >> f",
			"git add f",
			// 3cb885dd34db9637f7906427ca1c65d5e0568bfc
			"GIT_COMMITTER_DATE='2021-01-01T00:00:00Z' git commit -m bar",
		)

		// It's not possible to target the root commit because the commit on top of it
		// will always be returned first since it has an earlier committer date.
		commit, err := backend.RevAtTime(ctx, "HEAD", time.Date(2022, 6, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		require.Equal(t, api.CommitID("3cb885dd34db9637f7906427ca1c65d5e0568bfc"), commit)

		commit, err = backend.RevAtTime(ctx, "HEAD", time.Date(2021, 6, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		require.Equal(t, api.CommitID("3cb885dd34db9637f7906427ca1c65d5e0568bfc"), commit)

		commit, err = backend.RevAtTime(ctx, "HEAD", time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		require.Equal(t, api.CommitID(""), commit)
	})
}
