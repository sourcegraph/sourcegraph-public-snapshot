package background

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/rcache"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type fakeGitServer struct {
	gitserver.Client
	files        []string
	fileContents map[string]string
}

func (f fakeGitServer) LsFiles(ctx context.Context, repo api.RepoName, commit api.CommitID, pathspecs ...gitdomain.Pathspec) ([]string, error) {
	return f.files, nil
}

func (f fakeGitServer) ResolveRevision(ctx context.Context, repo api.RepoName, spec string) (api.CommitID, error) {
	return api.CommitID(""), nil
}

func (f fakeGitServer) ReadFile(ctx context.Context, repo api.RepoName, commit api.CommitID, name string) ([]byte, error) {
	if f.fileContents == nil {
		return nil, os.ErrNotExist
	}
	contents, ok := f.fileContents[name]
	if !ok {
		return nil, os.ErrNotExist
	}
	return []byte(contents), nil
}

func TestAnalyticsIndexerSuccess(t *testing.T) {
	rcache.SetupForTest(t)
	obsCtx := observation.TestContextTB(t)
	logger := obsCtx.Logger
	db := database.NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()
	user, err := db.Users().Create(ctx, database.NewUser{Username: "test"})
	require.NoError(t, err)
	var repoID api.RepoID = 1
	require.NoError(t, db.Repos().Create(ctx, &types.Repo{Name: "repo", ID: repoID}))
	client := fakeGitServer{
		files: []string{
			"notOwned.go",
			"alsoNotOwned.go",
			"owned/file1.go",
			"owned/file2.go",
			"owned/file3.go",
			"assigned.go",
		},
		fileContents: map[string]string{
			"CODEOWNERS": "/owned/* @owner",
		},
	}
	checker := authz.NewMockSubRepoPermissionChecker()
	checker.EnabledFunc.SetDefaultReturn(true)
	checker.EnabledForRepoIDFunc.SetDefaultReturn(false, nil)
	require.NoError(t, db.AssignedOwners().Insert(ctx, user.ID, repoID, "owned/file1.go", user.ID))
	require.NoError(t, db.AssignedOwners().Insert(ctx, user.ID, repoID, "assigned.go", user.ID))
	require.NoError(t, newAnalyticsIndexer(client, db, rcache.New("test_own_signal"), logger).indexRepo(ctx, repoID, checker))

	totalFileCount, err := db.RepoPaths().AggregateFileCount(ctx, database.TreeLocationOpts{})
	require.NoError(t, err)
	assert.Equal(t, int32(len(client.files)), totalFileCount)

	gotCounts, err := db.OwnershipStats().QueryAggregateCounts(ctx, database.TreeLocationOpts{})
	require.NoError(t, err)
	// We don't really need to compare time here.
	defaultTime := time.Time{}
	gotCounts.UpdatedAt = defaultTime
	wantCounts := database.PathAggregateCounts{
		CodeownedFileCount:         3,
		AssignedOwnershipFileCount: 2,
		TotalOwnedFileCount:        4,
		UpdatedAt:                  defaultTime,
	}
	assert.Equal(t, wantCounts, gotCounts)
}

func TestAnalyticsIndexerSkipsReposWithSubRepoPerms(t *testing.T) {
	rcache.SetupForTest(t)
	obsCtx := observation.TestContextTB(t)
	logger := obsCtx.Logger
	db := database.NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()
	var repoID api.RepoID = 1
	err := db.Repos().Create(ctx, &types.Repo{Name: "repo", ID: repoID})
	require.NoError(t, err)
	client := fakeGitServer{
		files: []string{"notOwned.go", "alsoNotOwned.go", "owned/file1.go", "owned/file2.go", "owned/file3.go"},
		fileContents: map[string]string{
			"CODEOWNERS": "/owned/* @owner",
		},
	}
	checker := authz.NewMockSubRepoPermissionChecker()
	checker.EnabledFunc.SetDefaultReturn(true)
	checker.EnabledForRepoIDFunc.SetDefaultReturn(true, nil)
	err = newAnalyticsIndexer(client, db, rcache.New("test_own_signal"), logger).indexRepo(ctx, repoID, checker)
	require.NoError(t, err)

	totalFileCount, err := db.RepoPaths().AggregateFileCount(ctx, database.TreeLocationOpts{})
	require.NoError(t, err)
	assert.Equal(t, int32(0), totalFileCount)

	codeownedCount, err := db.OwnershipStats().QueryAggregateCounts(ctx, database.TreeLocationOpts{})
	require.NoError(t, err)
	assert.Equal(t, database.PathAggregateCounts{CodeownedFileCount: 0}, codeownedCount)
}

func TestAnalyticsIndexerNoCodeowners(t *testing.T) {
	rcache.SetupForTest(t)
	obsCtx := observation.TestContextTB(t)
	logger := obsCtx.Logger
	db := database.NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()
	var repoID api.RepoID = 1
	err := db.Repos().Create(ctx, &types.Repo{Name: "repo", ID: repoID})
	require.NoError(t, err)
	client := fakeGitServer{
		files: []string{"notOwned.go", "alsoNotOwned.go", "owned/file1.go", "owned/file2.go", "owned/file3.go"},
	}
	checker := authz.NewMockSubRepoPermissionChecker()
	checker.EnabledFunc.SetDefaultReturn(true)
	checker.EnabledForRepoIDFunc.SetDefaultReturn(false, nil)
	err = newAnalyticsIndexer(client, db, rcache.New("test_own_signal"), logger).indexRepo(ctx, repoID, checker)
	require.NoError(t, err)

	totalFileCount, err := db.RepoPaths().AggregateFileCount(ctx, database.TreeLocationOpts{})
	require.NoError(t, err)
	assert.Equal(t, int32(5), totalFileCount)

	codeownedCount, err := db.OwnershipStats().QueryAggregateCounts(ctx, database.TreeLocationOpts{})
	defaultTime := time.Time{}
	codeownedCount.UpdatedAt = defaultTime
	require.NoError(t, err)
	assert.Equal(t, database.PathAggregateCounts{CodeownedFileCount: 0, UpdatedAt: defaultTime}, codeownedCount)
}
