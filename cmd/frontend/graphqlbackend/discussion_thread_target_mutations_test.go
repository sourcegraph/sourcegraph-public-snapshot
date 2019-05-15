package graphqlbackend

import (
	"context"
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func TestDiscussionThreadTargetInput_validate(t *testing.T) {
	const wantErr = "exactly 1 field in DiscussionThreadTargetInput must be non-null"
	if err := (&discussionThreadTargetInput{}).validate(); err == nil || err.Error() != wantErr {
		t.Fatalf("got error %v, want %q", err, wantErr)
	}
}

func TestDiscussionsMutations_AddTargetToThread(t *testing.T) {
	resetMocks()
	db.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) { return &types.User{}, nil }
	mockViewerCanUseDiscussions = func() error { return nil }
	defer func() { mockViewerCanUseDiscussions = nil }()
	const (
		wantRepoID   = api.RepoID(1)
		wantThreadID = 123
		wantPath     = "foo/bar"
	)
	db.Mocks.Repos.Get = func(_ context.Context, id api.RepoID) (*types.Repo, error) {
		if id != wantRepoID {
			t.Errorf("got repo ID %v, want %v", id, wantRepoID)
		}
		return &types.Repo{ID: wantRepoID}, nil
	}
	db.Mocks.DiscussionThreads.AddTarget = func(tr *types.DiscussionThreadTargetRepo) (*types.DiscussionThreadTargetRepo, error) {
		if tr.RepoID != wantRepoID {
			t.Errorf("got repo ID %v, want %v", tr.RepoID, wantRepoID)
		}
		if tr.ThreadID != wantThreadID {
			t.Errorf("got ThreadID %v, want %v", tr.ThreadID, wantThreadID)
		}
		if tr.Path == nil || *tr.Path != wantPath {
			var path string
			if tr.Path != nil {
				path = *tr.Path
			}
			t.Errorf("got Path %v, want %v", path, wantPath)
		}
		return &types.DiscussionThreadTargetRepo{Path: strptr(wantPath)}, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: GraphQLSchema,
			Query: `
				mutation($target: DiscussionThreadTargetInput!) {
					discussions {
						addTargetToThread(threadID: "123", target: $target) {
							__typename
							... on DiscussionThreadTargetRepo {
								path
							}
						}
					}
				}
			`,
			Variables: map[string]interface{}{
				"target": map[string]interface{}{
					"repo": map[string]interface{}{
						"repositoryID": string(marshalRepositoryID(wantRepoID)),
						"path":         wantPath,
					},
				},
			},
			ExpectedResult: `
				{
					"discussions": {
						"addTargetToThread": {
							"__typename": "DiscussionThreadTargetRepo",
							"path": "` + wantPath + `"
						}
					}
				}
			`,
		},
	})
}

func TestDiscussionsMutations_UpdateTargetInThread(t *testing.T) {
	resetMocks()
	db.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) { return &types.User{}, nil }
	mockViewerCanUseDiscussions = func() error { return nil }
	defer func() { mockViewerCanUseDiscussions = nil }()
	const wantTargetID = 123

	t.Run("remove", func(t *testing.T) {
		db.Mocks.DiscussionThreads.RemoveTarget = func(targetID int64) error {
			if targetID != wantTargetID {
				t.Errorf("got targetID %v, want %v", targetID, wantTargetID)
			}
			return nil
		}
		defer func() { db.Mocks.DiscussionThreads.RemoveTarget = nil }()
		gqltesting.RunTests(t, []*gqltesting.Test{
			{
				Schema: GraphQLSchema,
				Query: `
					mutation($targetID: ID!) {
						discussions {
							updateTargetInThread(targetID: $targetID, remove: true) {
								__typename
							}
						}
					}
			`,
				Variables: map[string]interface{}{"targetID": string(marshalDiscussionThreadTargetID(wantTargetID))},
				ExpectedResult: `
				{
					"discussions": {
						"updateTargetInThread": null
					}
				}
			`,
			},
		})
	})
}
