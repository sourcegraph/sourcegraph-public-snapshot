package graphqlbackend

import (
	"context"
	"os"
	"testing"

	"github.com/neelance/graphql-go/gqltesting"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	vcstest "sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/testing"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/util"
)

func TestTree(t *testing.T) {
	resetMocks()
	localstore.Mocks.Repos.MockGetByURI(t, "github.com/gorilla/mux", 2)
	backend.Mocks.Repos.ResolveRev = func(ctx context.Context, op *sourcegraph.ReposResolveRevOp) (*sourcegraph.ResolvedRev, error) {
		if op.Repo != 2 || op.Rev != exampleCommitSHA1 {
			t.Error("wrong arguments to Repos.ResolveRev")
		}
		return &sourcegraph.ResolvedRev{
			CommitID: exampleCommitSHA1,
		}, nil
	}

	mockRepo := vcstest.MockRepository{}
	mockRepo.ReadDir_ = func(ctx context.Context, commit vcs.CommitID, name string, recurse bool) ([]os.FileInfo, error) {
		if string(commit) != exampleCommitSHA1 || name != "/foo" {
			t.Error("wrong arguments to RepoTree.Get")
		}
		return []os.FileInfo{
			&util.FileInfo{Name_: "testDirectory", Mode_: os.ModeDir},
			&util.FileInfo{Name_: "testFile", Mode_: 0},
		}, nil
	}
	localstore.Mocks.RepoVCS.Open = func(ctx context.Context, repo int32) (vcs.Repository, error) {
		return mockRepo, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: GraphQLSchema,
			Query: `
				{
					repository(uri: "github.com/gorilla/mux") {
						commit(rev: "` + exampleCommitSHA1 + `") {
							commit {
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
				}
			`,
			ExpectedResult: `
				{
					"repository": {
						"commit": {
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
				}
			`,
		},
	})
}
