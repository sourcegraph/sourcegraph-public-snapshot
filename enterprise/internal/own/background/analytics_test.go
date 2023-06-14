package background

import (
	"context"
	"os"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

func (f fakeGitServer) LsFiles(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, commit api.CommitID, pathspecs ...gitdomain.Pathspec) ([]string, error) {
	return f.files, nil
}

func (f fakeGitServer) ResolveRevision(ctx context.Context, repo api.RepoName, spec string, opt gitserver.ResolveRevisionOptions) (api.CommitID, error) {
	return api.CommitID(""), nil
}

func (f fakeGitServer) ReadFile(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, commit api.CommitID, name string) ([]byte, error) {
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
	obsCtx := observation.TestContextTB(t)
	logger := obsCtx.Logger
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
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
	checker.EnabledForRepoIDFunc.SetDefaultReturn(false, nil)
	err = newAnalyticsIndexer(client, db, nil, logger).indexRepo(ctx, repoID, checker)
	require.NoError(t, err)

	totalFileCount, err := db.RepoPaths().AggregateFileCount(ctx, database.TreeLocationOpts{})
	require.NoError(t, err)
	assert.Equal(t, int32(5), totalFileCount)

	codeownedCount, err := db.OwnershipStats().QueryAggregateCounts(ctx, database.TreeLocationOpts{})
	require.NoError(t, err)
	assert.Equal(t, database.PathAggregateCounts{CodeownedFileCount: 3}, codeownedCount)
}

func TestAnalyticsIndexerSkipsReposWithSubRepoPerms(t *testing.T) {
	obsCtx := observation.TestContextTB(t)
	logger := obsCtx.Logger
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
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
	err = newAnalyticsIndexer(client, db, nil, logger).indexRepo(ctx, repoID, checker)
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
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
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
	err = newAnalyticsIndexer(client, db, nil, logger).indexRepo(ctx, repoID, checker)
	require.NoError(t, err)

	totalFileCount, err := db.RepoPaths().AggregateFileCount(ctx, database.TreeLocationOpts{})
	require.NoError(t, err)
	assert.Equal(t, int32(5), totalFileCount)

	codeownedCount, err := db.OwnershipStats().QueryAggregateCounts(ctx, database.TreeLocationOpts{})
	require.NoError(t, err)
	assert.Equal(t, database.PathAggregateCounts{CodeownedFileCount: 0}, codeownedCount)
}
