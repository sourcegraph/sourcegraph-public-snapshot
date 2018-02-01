package graphqlbackend

import (
	"context"
	"os"
	"testing"

	"github.com/neelance/graphql-go/gqltesting"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	vcstest "sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/testing"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/util"
)

func TestTree(t *testing.T) {
	resetMocks()
	db.Mocks.Repos.MockGetByURI(t, "github.com/gorilla/mux", 2)
	backend.Mocks.Repos.ResolveRev = func(ctx context.Context, repo *types.Repo, rev string) (api.CommitID, error) {
		if repo.ID != 2 || rev != exampleCommitSHA1 {
			t.Error("wrong arguments to Repos.ResolveRev")
		}
		return exampleCommitSHA1, nil
	}
	backend.Mocks.Repos.MockGetCommit_Return_NoCheck(t, &vcs.Commit{ID: exampleCommitSHA1})

	mockRepo := vcstest.MockRepository{}
	mockRepo.ReadDir_ = func(ctx context.Context, commit api.CommitID, name string, recurse bool) ([]os.FileInfo, error) {
		if string(commit) != exampleCommitSHA1 || name != "/foo" {
			t.Error("wrong arguments to RepoTree.Get")
		}
		return []os.FileInfo{
			&util.FileInfo{Name_: "testDirectory", Mode_: os.ModeDir},
			&util.FileInfo{Name_: "testFile", Mode_: 0},
		}, nil
	}
	backend.Mocks.Repos.MockVCS(t, "github.com/gorilla/mux", mockRepo)

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: GraphQLSchema,
			Query: `
				{
					repository(uri: "github.com/gorilla/mux") {
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
