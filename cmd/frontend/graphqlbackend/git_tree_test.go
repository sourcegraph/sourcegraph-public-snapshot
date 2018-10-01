package graphqlbackend

import (
	"context"
	"os"
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/util"
)

func TestGitTree(t *testing.T) {
	resetMocks()
	db.Mocks.Repos.MockGetByURI(t, "github.com/gorilla/mux", 2)
	backend.Mocks.Repos.ResolveRev = func(ctx context.Context, repo *types.Repo, rev string) (api.CommitID, error) {
		if repo.ID != 2 || rev != exampleCommitSHA1 {
			t.Error("wrong arguments to Repos.ResolveRev")
		}
		return exampleCommitSHA1, nil
	}
	backend.Mocks.Repos.MockGetCommit_Return_NoCheck(t, &git.Commit{ID: exampleCommitSHA1})

	git.Mocks.Stat = func(commit api.CommitID, path string) (os.FileInfo, error) {
		if string(commit) != exampleCommitSHA1 || path != "/foo" {
			t.Error("wrong arguments to Stat")
		}
		return &util.FileInfo{Name_: "", Mode_: os.ModeDir}, nil
	}
	git.Mocks.ReadDir = func(commit api.CommitID, name string, recurse bool) ([]os.FileInfo, error) {
		if string(commit) != exampleCommitSHA1 || name != "/foo" {
			t.Error("wrong arguments to RepoTree.Get")
		}
		return []os.FileInfo{
			&util.FileInfo{Name_: "testDirectory", Mode_: os.ModeDir},
			&util.FileInfo{Name_: "testFile", Mode_: 0},
		}, nil
	}
	defer git.ResetMocks()

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: GraphQLSchema,
			Query: `
				{
					repository(name: "github.com/gorilla/mux") {
						commit(rev: "` + exampleCommitSHA1 + `") {
							tree(path: "/foo") {
								directories {
									name
								}
								files {
									name
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
									{"name": "testDirectory"}
								],
								"files": [
									{"name": "testFile"}
								]
							}
						}
					}
				}
			`,
		},
	})
}
