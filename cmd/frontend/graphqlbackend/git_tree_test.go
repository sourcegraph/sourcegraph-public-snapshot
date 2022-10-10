package graphqlbackend

import (
	"context"
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/fileutil"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestGitTree(t *testing.T) {
	db := database.NewMockDB()
	gsClient := setupGitserverClient(t)
	tests := []*Test{
		{
			Schema: mustParseGraphQLSchemaWithClient(t, db, gsClient),
			Query: `
				{
					repository(name: "github.com/gorilla/mux") {
						commit(rev: "` + exampleCommitSHA1 + `") {
							tree(path: "foo bar") {
								directories {
									name
									path
									url
								}
								files {
									name
									path
									url
								}
							}
						}
					}
				}
			`,
			ExpectedResult: `
{
  "repository": {
    "commit": {
      "tree": {
        "directories": [
          {
            "name": "Geoffrey's random queries.32r242442bf",
            "path": "foo bar/Geoffrey's random queries.32r242442bf",
            "url": "/github.com/gorilla/mux@1234567890123456789012345678901234567890/-/tree/foo%20bar/Geoffrey%27s%20random%20queries.32r242442bf"
          },
          {
            "name": "testDirectory",
            "path": "foo bar/testDirectory",
            "url": "/github.com/gorilla/mux@1234567890123456789012345678901234567890/-/tree/foo%20bar/testDirectory"
          }
        ],
        "files": [
          {
            "name": "% token.4288249258.sql",
            "path": "foo bar/% token.4288249258.sql",
            "url": "/github.com/gorilla/mux@1234567890123456789012345678901234567890/-/blob/foo%20bar/%25%20token.4288249258.sql"
          },
          {
            "name": "testFile",
            "path": "foo bar/testFile",
            "url": "/github.com/gorilla/mux@1234567890123456789012345678901234567890/-/blob/foo%20bar/testFile"
          }
        ]
      }
    }
  }
}
			`,
		},
	}
	testGitTree(t, db, tests)
}

func setupGitserverClient(t *testing.T) gitserver.Client {
	t.Helper()
	gsClient := gitserver.NewMockClient()
	gsClient.ReadDirFunc.SetDefaultHook(func(_ context.Context, _ authz.SubRepoPermissionChecker, _ api.RepoName, commit api.CommitID, name string, recurse bool) ([]fs.FileInfo, error) {
		assert.Equal(t, api.CommitID(exampleCommitSHA1), commit)
		assert.Equal(t, "foo bar", name)
		assert.False(t, recurse)
		return []fs.FileInfo{
			&fileutil.FileInfo{Name_: name + "/testDirectory", Mode_: os.ModeDir},
			&fileutil.FileInfo{Name_: name + "/Geoffrey's random queries.32r242442bf", Mode_: os.ModeDir},
			&fileutil.FileInfo{Name_: name + "/testFile", Mode_: 0},
			&fileutil.FileInfo{Name_: name + "/% token.4288249258.sql", Mode_: 0},
		}, nil
	})
	gsClient.StatFunc.SetDefaultHook(func(_ context.Context, _ authz.SubRepoPermissionChecker, _ api.RepoName, commit api.CommitID, path string) (fs.FileInfo, error) {
		assert.Equal(t, api.CommitID(exampleCommitSHA1), commit)
		assert.Equal(t, "foo bar", path)
		return &fileutil.FileInfo{Name_: path, Mode_: os.ModeDir}, nil
	})
	return gsClient
}

func testGitTree(t *testing.T, db *database.MockDB, tests []*Test) {
	externalServices := database.NewMockExternalServiceStore()
	externalServices.ListFunc.SetDefaultReturn(nil, nil)

	repos := database.NewMockRepoStore()
	repos.GetFunc.SetDefaultReturn(&types.Repo{ID: 2, Name: "github.com/gorilla/mux"}, nil)
	repos.GetByNameFunc.SetDefaultReturn(&types.Repo{ID: 2, Name: "github.com/gorilla/mux"}, nil)

	db.ExternalServicesFunc.SetDefaultReturn(externalServices)
	db.ReposFunc.SetDefaultReturn(repos)

	backend.Mocks.Repos.ResolveRev = func(ctx context.Context, repo *types.Repo, rev string) (api.CommitID, error) {
		assert.Equal(t, api.RepoID(2), repo.ID)
		assert.Equal(t, exampleCommitSHA1, rev)
		return exampleCommitSHA1, nil
	}
	backend.Mocks.Repos.MockGetCommit_Return_NoCheck(t, &gitdomain.Commit{ID: exampleCommitSHA1})
	defer func() {
		backend.Mocks = backend.MockServices{}
		gitserver.ResetMocks()
	}()

	RunTests(t, tests)
}
