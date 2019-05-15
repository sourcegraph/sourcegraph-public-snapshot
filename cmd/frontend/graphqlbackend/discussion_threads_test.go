package graphqlbackend

import (
	"context"
	"reflect"
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

func TestDiscussionThread_Target(t *testing.T) {
	resetMocks()
	mockViewerCanUseDiscussions = func() error { return nil }
	defer func() { mockViewerCanUseDiscussions = nil }()
	const wantThreadID = 123
	db.Mocks.DiscussionThreads.List = func(_ context.Context, opt *db.DiscussionThreadsListOptions) ([]*types.DiscussionThread, error) {
		return []*types.DiscussionThread{{ID: wantThreadID}}, nil
	}

	t.Run("no target", func(t *testing.T) {
		db.Mocks.DiscussionThreads.ListTargets = func(threadID int64) ([]*types.DiscussionThreadTargetRepo, error) {
			if threadID != wantThreadID {
				t.Fatalf("got threadID %v, want %v", threadID, wantThreadID)
			}
			return []*types.DiscussionThreadTargetRepo{}, nil
		}
		gqltesting.RunTests(t, []*gqltesting.Test{
			{
				Schema: GraphQLSchema,
				Query: `
				{
					discussionThreads {
						nodes {
							target {
								__typename
								... on DiscussionThreadTargetRepo {
									path
								}
							}
						}
					}
				}
			`,
				ExpectedResult: `
				{
					"discussionThreads": {
						"nodes": [
							{
								"target": null
							}
						]
					}
				}
			`,
			},
		})
	})
	t.Run("first target from multiple", func(t *testing.T) {
		db.Mocks.DiscussionThreads.ListTargets = func(threadID int64) ([]*types.DiscussionThreadTargetRepo, error) {
			if threadID != wantThreadID {
				t.Fatalf("got threadID %v, want %v", threadID, wantThreadID)
			}
			return []*types.DiscussionThreadTargetRepo{{Path: strptr("foo/bar")}, {Path: strptr("qux")}}, nil
		}
		gqltesting.RunTests(t, []*gqltesting.Test{
			{
				Schema: GraphQLSchema,
				Query: `
				{
					discussionThreads {
						nodes {
							target {
								__typename
								... on DiscussionThreadTargetRepo {
									path
								}
							}
						}
					}
				}
			`,
				ExpectedResult: `
				{
					"discussionThreads": {
						"nodes": [
							{
								"target": {
									"__typename": "DiscussionThreadTargetRepo",
									"path": "foo/bar"
								}
							}
						]
					}
				}
			`,
			},
		})
	})
}

func TestDiscussionThread_Get(t *testing.T) {
	resetMocks()
	mockViewerCanUseDiscussions = func() error { return nil }
	defer func() { mockViewerCanUseDiscussions = nil }()
	const (
		wantThreadID        = 123
		wantThreadGraphQLID = "RGlzY3Vzc2lvblRocmVhZDoiM2Yi"
	)
	db.Mocks.DiscussionThreads.Get = func(threadID int64) (*types.DiscussionThread, error) {
		return &types.DiscussionThread{ID: wantThreadID}, nil
	}

	t.Run("by ID", func(t *testing.T) {
		gqltesting.RunTests(t, []*gqltesting.Test{
			{
				Schema: GraphQLSchema,
				Query: `
                                query ($id: ID!) {
                                        node(id: $id) {
                                                __typename
												id
                                        }
                                }
                        `,
				Variables: map[string]interface{}{"id": wantThreadGraphQLID},
				ExpectedResult: `
                                {
                                        "node": {
											"__typename": "DiscussionThread",
											"id": "` + wantThreadGraphQLID + `"
                                        }
                                }
                        `,
			},
		})
	})
}

func TestDiscussionSelectionRelativeTo(t *testing.T) {
	i32 := func(i int32) *int32 {
		return &i
	}
	tests := []struct {
		name         string
		oldSelection *types.DiscussionThreadTargetRepo
		newContent   string
		want         *discussionSelectionRangeResolver
	}{
		{
			name: "added_content_before",
			oldSelection: &types.DiscussionThreadTargetRepo{
				StartLine: i32(3), StartCharacter: i32(0), EndLine: i32(4), EndCharacter: i32(1),
				LinesBefore: &[]string{"0", "1", "2"},
				Lines:       &[]string{"3"},
				LinesAfter:  &[]string{"4", "5", "6"},
			},
			newContent: "a\nb\nc\n0\n1\n2\n3\n4\n5\n6",
			want:       &discussionSelectionRangeResolver{startLine: 6, startCharacter: 0, endLine: 7, endCharacter: 1},
		},
		{
			name: "added_content_after",
			oldSelection: &types.DiscussionThreadTargetRepo{
				StartLine: i32(3), StartCharacter: i32(0), EndLine: i32(4), EndCharacter: i32(1),
				LinesBefore: &[]string{"0", "1", "2"},
				Lines:       &[]string{"3"},
				LinesAfter:  &[]string{"4", "5", "6"},
			},
			newContent: "0\n1\n2\n3\n4\n5\n6\na\nb\nc",
			want:       &discussionSelectionRangeResolver{startLine: 3, startCharacter: 0, endLine: 4, endCharacter: 1},
		},
		{
			name: "added_content_before_and_after",
			oldSelection: &types.DiscussionThreadTargetRepo{
				StartLine: i32(3), StartCharacter: i32(0), EndLine: i32(4), EndCharacter: i32(1),
				LinesBefore: &[]string{"0", "1", "2"},
				Lines:       &[]string{"3"},
				LinesAfter:  &[]string{"4", "5", "6"},
			},
			newContent: "a\nb\nc\n0\n1\n2\n3\n4\n5\n6\na\nb\nc",
			want:       &discussionSelectionRangeResolver{startLine: 6, startCharacter: 0, endLine: 7, endCharacter: 1},
		},
		{
			name: "removed_content_before",
			oldSelection: &types.DiscussionThreadTargetRepo{
				StartLine: i32(3), StartCharacter: i32(0), EndLine: i32(4), EndCharacter: i32(1),
				LinesBefore: &[]string{"0", "1", "2"},
				Lines:       &[]string{"3"},
				LinesAfter:  &[]string{"4", "5", "6"},
			},
			newContent: "3\n4\n5\n6",
			want:       &discussionSelectionRangeResolver{startLine: 0, startCharacter: 0, endLine: 1, endCharacter: 1},
		},
		{
			name: "removed_content_after",
			oldSelection: &types.DiscussionThreadTargetRepo{
				StartLine: i32(3), StartCharacter: i32(0), EndLine: i32(4), EndCharacter: i32(1),
				LinesBefore: &[]string{"0", "1", "2"},
				Lines:       &[]string{"3"},
				LinesAfter:  &[]string{"4", "5", "6"},
			},
			newContent: "0\n1\n2\n3\n",
			want:       &discussionSelectionRangeResolver{startLine: 0, startCharacter: 0, endLine: 1, endCharacter: 1},
		},
		{
			name: "no_match",
			oldSelection: &types.DiscussionThreadTargetRepo{
				StartLine: i32(3), StartCharacter: i32(0), EndLine: i32(4), EndCharacter: i32(1),
				LinesBefore: &[]string{"0", "1", "2"},
				Lines:       &[]string{"3"},
				LinesAfter:  &[]string{"4", "5", "6"},
			},
			newContent: "0\n2\n3\n1\n",
			want:       nil,
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			got := discussionSelectionRelativeTo(tst.oldSelection, tst.newContent)
			if !reflect.DeepEqual(got, tst.want) {
				t.Logf("got  %+v\n", got)
				t.Fatalf("want %+v\n", tst.want)
			}
		})
	}
}

func TestDiscussionsMutations_UpdateThread(t *testing.T) {
	resetMocks()
	db.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) { return &types.User{}, nil }
	mockViewerCanUseDiscussions = func() error { return nil }
	defer func() { mockViewerCanUseDiscussions = nil }()
	const (
		wantThreadID = 123
		wantTitle    = "b"
	)
	db.Mocks.DiscussionThreads.Get = func(threadID int64) (*types.DiscussionThread, error) {
		if threadID != wantThreadID {
			t.Errorf("got threadID %v, want %v", threadID, wantThreadID)
		}
		return &types.DiscussionThread{}, nil
	}
	db.Mocks.DiscussionThreads.Update = func(_ context.Context, threadID int64, opts *db.DiscussionThreadsUpdateOptions) (*types.DiscussionThread, error) {
		if threadID != wantThreadID {
			t.Errorf("got threadID %v, want %v", threadID, wantThreadID)
		}
		if opts == nil || opts.Title == nil || *opts.Title != wantTitle {
			var title string
			if opts != nil && opts.Title != nil {
				title = *opts.Title
			}
			t.Errorf("got title %v, want %v", title, wantTitle)
		}
		return &types.DiscussionThread{ID: wantThreadID, Title: wantTitle}, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: backend.WithAuthzBypass(context.Background()),
			Schema:  GraphQLSchema,
			Query: `
                                mutation($title: String!) {
                                        discussions {
                                                updateThread(input: {threadID: "RGlzY3Vzc2lvblRocmVhZDoiM2Yi", title: $title}) {
                                                        title
                                                }
                                        }
                                }
                        `,
			Variables: map[string]interface{}{"title": wantTitle},
			ExpectedResult: `
                                {
                                        "discussions": {
                                                "updateThread": {
                                                        "title": "` + wantTitle + `"
                                                }
                                        }
                                }
                        `,
		},
	})
}
