package graphqlbackend

import (
	"context"
	"testing"

	graphql "github.com/neelance/graphql-go"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
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
	backend.Mocks.RepoTree.Get = func(ctx context.Context, op *sourcegraph.RepoTreeGetOp) (*sourcegraph.TreeEntry, error) {
		if op.Entry.RepoRev.Repo != 2 || op.Entry.RepoRev.CommitID != exampleCommitSHA1 || op.Entry.Path != "/foo" {
			t.Error("wrong arguments to RepoTree.Get")
		}
		return &sourcegraph.TreeEntry{
			BasicTreeEntry: &sourcegraph.BasicTreeEntry{
				Entries: []*sourcegraph.BasicTreeEntry{
					&sourcegraph.BasicTreeEntry{Name: "testDirectory", Type: sourcegraph.DirEntry},
					&sourcegraph.BasicTreeEntry{Name: "testFile", Type: sourcegraph.FileEntry},
				},
			},
		}, nil
	}

	graphql.RunTests(t, []*graphql.Test{
		{
			Schema: GraphQLSchema,
			Query: `
				{
					root {
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
				}
			`,
			ExpectedResult: `
				{
					"root": {
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
				}
			`,
		},
	})
}
