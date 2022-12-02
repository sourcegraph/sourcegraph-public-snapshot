package service

import (
	"context"
	"io/fs"
	"os"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	bt "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
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

func TestGetRepoStatus(t *testing.T) {
	ctx := context.Background()

	logger := logtest.Scoped(t)

	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	s := store.New(db, &observation.TestContext, nil)

	repos, _ := bt.CreateTestRepos(t, ctx, db, 5)
	unsupported, _ := bt.CreateAWSCodeCommitTestRepos(t, ctx, db, 1)

	heads := map[api.RepoName]api.CommitID{}
	for _, repo := range append(repos, unsupported...) {
		heads[repo.Name] = api.CommitID(repo.Name + "-commit")
	}

	// We'll set up a mock gitserver client. Interesting repos:
	//
	// rs[1]: is uncloned and cannot resolve the head to a commit.
	// rs[2]: errors trying to resolve the head commit.
	// rs[3]: has a .batchignore file.
	// rs[4]: errors stat-ing the .batchignore file.
	newGitserverClient := func(t *testing.T) *gitserver.MockClient {
		client := gitserver.NewMockClient()

		client.HeadFunc.SetDefaultHook(func(ctx context.Context, name api.RepoName, _ authz.SubRepoPermissionChecker) (string, bool, error) {
			if name == repos[1].Name {
				return "", false, nil
			}
			if name == repos[2].Name {
				return "", false, errors.New("cannot resolve head!")
			}
			commit, ok := heads[name]
			return string(commit), ok, nil
		})

		client.StatFunc.SetDefaultHook(func(ctx context.Context, _ authz.SubRepoPermissionChecker, name api.RepoName, commit api.CommitID, path string) (fs.FileInfo, error) {
			assert.Equal(t, heads[name], commit)
			assert.Equal(t, batchIgnoreFilePath, path)
			if name == repos[3].Name {
				return &fileutil.FileInfo{Name_: batchIgnoreFilePath, Mode_: 0}, nil
			}
			if name == repos[4].Name {
				return nil, errors.New("error executing stat!")
			}
			return nil, os.ErrNotExist
		})

		return client
	}

	assertExpectedStatuses := func(t *testing.T, client gitserver.Client) {
		t.Helper()

		t.Run("successful, not ignored", func(t *testing.T) {
			for _, repo := range []*types.Repo{repos[0], unsupported[0]} {
				rs, err := getRepoStatus(ctx, s, client, repo)
				assert.NoError(t, err)
				assert.NotNil(t, rs)
				assert.False(t, rs.Ignored)
			}
		})

		t.Run("successful, ignored", func(t *testing.T) {
			rs, err := getRepoStatus(ctx, s, client, repos[3])
			assert.NoError(t, err)
			assert.NotNil(t, rs)
			assert.True(t, rs.Ignored)
		})

		t.Run("failures", func(t *testing.T) {
			for _, repo := range []*types.Repo{repos[1], repos[2], repos[4]} {
				rs, err := getRepoStatus(ctx, s, client, repo)
				assert.Error(t, err)
				assert.Nil(t, rs)
			}
		})
	}

	t.Run("empty cache", func(t *testing.T) {
		client := newGitserverClient(t)
		assertExpectedStatuses(t, client)
	})

	t.Run("primed cache", func(t *testing.T) {
		// Use a new client that has a failed assertion if an unexpected repo hits
		// the stat endpoint.
		client := newGitserverClient(t)
		client.StatFunc.SetDefaultHook(func(ctx context.Context, _ authz.SubRepoPermissionChecker, name api.RepoName, commit api.CommitID, path string) (fs.FileInfo, error) {
			assert.Contains(t, []api.RepoName{repos[1].Name, repos[2].Name, repos[4].Name}, name)
			return nil, errors.New("still don't want the cache populated")
		})
		assertExpectedStatuses(t, client)
	})

	t.Run("outdated cache", func(t *testing.T) {
		for name, commit := range heads {
			heads[name] = commit + api.CommitID("-new")
		}

		// Now we should get requests once more.
		client := newGitserverClient(t)
		assertExpectedStatuses(t, client)
	})
}
