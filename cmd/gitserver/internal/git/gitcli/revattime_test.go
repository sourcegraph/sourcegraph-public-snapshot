package gitcli

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

func TestGitCLIBackend_RevAtTime(t *testing.T) {
	ctx := context.Background()

	t.Run("basic", func(t *testing.T) {
		backend := BackendWithRepoCommands(t,
			"echo line1 > f",
			"git add f",
			// 1ba45c7103f4df182d7ce361567c61308ebae8ed
			"GIT_COMMITTER_DATE='2021-01-01T00:00:00' git commit -m foo",
			"git tag testbase",

			"git checkout -b b2",
			"echo line2 >> f",
			"git add f",
			// 51bb862781503cde4634c75144e3939c1561761b
			"GIT_COMMITTER_DATE='2022-01-01T00:00:00' git commit -m bar",

			"git checkout master",
			"echo line3 > h",
			"git add h",
			// d00ae3010469c160d508f932afa58ca7c690abbb
			"GIT_COMMITTER_DATE='2023-01-01T00:00:00' git commit -m qux",
			"git tag v1.0.0",

			// 47e7692f76bd092c3408c49362c65ff2e1c5eb79
			"GIT_COMMITTER_DATE='2024-01-01T00:00:00' git merge b2",
		)

		// Target the first commit on master
		commit, err := backend.RevAtTime(ctx, "HEAD", time.Date(2021, 6, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		require.Equal(t, api.CommitID("1ba45c7103f4df182d7ce361567c61308ebae8ed"), commit)

		// Target the second commit on master
		commit, err = backend.RevAtTime(ctx, "HEAD", time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		require.Equal(t, api.CommitID("d00ae3010469c160d508f932afa58ca7c690abbb"), commit)

		// A date before the first commit returns an empty string
		commit, err = backend.RevAtTime(ctx, "HEAD", time.Date(1996, 6, 28, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		require.Equal(t, api.CommitID(""), commit)

		// If we traversed merged branches, this would target d5fe. Since we use --first-parent,
		// this targets the root commit.
		commit, err = backend.RevAtTime(ctx, "HEAD", time.Date(2022, 6, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		require.Equal(t, api.CommitID("1ba45c7103f4df182d7ce361567c61308ebae8ed"), commit)

		// If we target the b2 branch specifically though, that's on the --first-parent history
		commit, err = backend.RevAtTime(ctx, "b2", time.Date(2022, 6, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		require.Equal(t, api.CommitID("51bb862781503cde4634c75144e3939c1561761b"), commit)

		// Targeting in the future is fine
		commit, err = backend.RevAtTime(ctx, "master", time.Date(2048, 6, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		require.Equal(t, api.CommitID("47e7692f76bd092c3408c49362c65ff2e1c5eb79"), commit)
	})

	t.Run("out of order commit date", func(t *testing.T) {
		backend := BackendWithRepoCommands(t,
			"echo line1 > f",
			"git add f",
			// 42b1f173d788ad508388cb97f645a733d314d004
			"GIT_COMMITTER_DATE='2022-01-01T00:00:00' git commit -m foo",
			"git tag testbase",

			"echo line2 >> f",
			"git add f",
			// 442e6591b3940bdf66be81afcba2906ea5dde703
			"GIT_COMMITTER_DATE='2021-01-01T00:00:00' git commit -m bar",
		)

		// It's not possible to target the root commit because the commit on top of it
		// will always be returned first since it has an earlier committer date.
		commit, err := backend.RevAtTime(ctx, "HEAD", time.Date(2022, 6, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		require.Equal(t, api.CommitID("442e6591b3940bdf66be81afcba2906ea5dde703"), commit)

		commit, err = backend.RevAtTime(ctx, "HEAD", time.Date(2021, 6, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		require.Equal(t, api.CommitID("442e6591b3940bdf66be81afcba2906ea5dde703"), commit)

		commit, err = backend.RevAtTime(ctx, "HEAD", time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		require.Equal(t, api.CommitID(""), commit)
	})
}
