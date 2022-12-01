package service

import (
	"context"
	"io/fs"
	"os"
	"testing"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	bt "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/fileutil"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestGetRepoMetadata(t *testing.T) {
	ctx := context.Background()

	logger := logtest.Scoped(t)

	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	s := store.New(db, &observation.TestContext, nil)

	u := bt.CreateTestUser(t, db, false)

	rs, _ := bt.CreateTestRepos(t, ctx, db, 5)
	unsupported, _ := bt.CreateAWSCodeCommitTestRepos(t, ctx, db, 1)

	ctx = actor.WithActor(context.Background(), actor.FromUser(u.ID))

	// We'll set up a mock gitserver client. Interesting repos:
	//
	// rs[1]: is uncloned and cannot resolve the head to a commit.
	// rs[2]: errors trying to resolve the head commit.
	// rs[3]: has a .batchignore file.
	// rs[4]: errors stat-ing the .batchignore file.
	newGitserverClient := func(t *testing.T) gitserver.Client {
		client := gitserver.NewMockClient()

		client.HeadFunc.SetDefaultHook(func(ctx context.Context, name api.RepoName, _ authz.SubRepoPermissionChecker) (string, bool, error) {
			if name == rs[1].Name {
				return "", false, nil
			}
			if name == rs[2].Name {
				return "", false, errors.New("cannot resolve head!")
			}
			return "abcdef", true, nil
		})

		client.StatFunc.SetDefaultHook(func(ctx context.Context, _ authz.SubRepoPermissionChecker, name api.RepoName, _ api.CommitID, path string) (fs.FileInfo, error) {
			assert.Equal(t, batchIgnoreFilePath, path)
			if name == rs[3].Name {
				return &fileutil.FileInfo{Name_: batchIgnoreFilePath, Mode_: 0}, nil
			}
			if name == rs[4].Name {
				return nil, errors.New("error executing stat!")
			}
			return nil, os.ErrNotExist
		})

		t.Cleanup(func() {
			assert.Len(t, client.HeadFunc.History(), 6)
			assert.Len(t, client.StatFunc.History(), 4)
		})

		return client
	}

	assertExpectedMetadata := func(t *testing.T, client gitserver.Client) {
		t.Helper()

		t.Run("successful, not ignored", func(t *testing.T) {
			for _, repo := range []*types.Repo{rs[0], unsupported[0]} {
				meta, err := getRepoMetadata(ctx, s, client, repo)
				assert.NoError(t, err)
				assert.NotNil(t, meta)
				assert.False(t, meta.Ignored)
			}
		})

		t.Run("successful, ignored", func(t *testing.T) {
			meta, err := getRepoMetadata(ctx, s, client, rs[3])
			assert.NoError(t, err)
			assert.NotNil(t, meta)
			assert.True(t, meta.Ignored)
		})

		t.Run("failures", func(t *testing.T) {
			for _, repo := range []*types.Repo{rs[1], rs[2], rs[4]} {
				meta, err := getRepoMetadata(ctx, s, client, repo)
				assert.Error(t, err)
				assert.Nil(t, meta)
			}
		})
	}

	t.Run("database doesn't have repo metadata", func(t *testing.T) {
		client := newGitserverClient(t)
		assertExpectedMetadata(t, client)
	})

	t.Run("database now has repo metadata", func(t *testing.T) {
		// Use a new client that will error if anything is invoked other than the
		// initial head requests for rs[1], rs[2] or rs[4], which errored last time.
		client := gitserver.NewMockClient()
		client.HeadFunc.SetDefaultHook(func(ctx context.Context, name api.RepoName, _ authz.SubRepoPermissionChecker) (string, bool, error) {
			assert.Contains(t, []api.RepoName{rs[1].Name, rs[2].Name, rs[4].Name}, name)
			return "", false, errors.New("cannot resolve head!")
		})
		assertExpectedMetadata(t, client)
		assert.Len(t, client.HeadFunc.History(), 3)
	})

	t.Run("database has outdated repo metadata", func(t *testing.T) {
		// There are two different ways metadata can be detected as outdated, since
		// repo.updated_at is nullable: either the metadata was created before the
		// repo, or the metadata was updated before the repo was updated. We'll set
		// this up to test both by only setting repo.updated_at on a subset of
		// repos.
		err := s.Exec(ctx, sqlf.Sprintf("UPDATE repo SET updated_at = created_at WHERE id < 4"))
		require.NoError(t, err)

		err = s.Exec(ctx, sqlf.Sprintf(`
      UPDATE
        batch_changes_repo_metadata
      SET
        updated_at = (
          SELECT
            created_at - INTERVAL '1 hour'
          FROM
            repo
          WHERE
            repo.id = batch_changes_repo_metadata.repo_id
        )
    `))
		require.NoError(t, err)

		// Now we should get requests once more.
		client := newGitserverClient(t)
		assertExpectedMetadata(t, client)
	})
}
