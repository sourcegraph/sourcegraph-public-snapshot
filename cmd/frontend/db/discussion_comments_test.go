package db

import (
	"testing"
	"time"

	dbtesting "github.com/sourcegraph/sourcegraph/cmd/frontend/db/testing"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

// TODO(slimsag:discussions): future: test that DiscussionCommentsListOptions.AuthorUserID works
// TODO(slimsag:discussions): future: test that DiscussionCommentsListOptions.ThreadID works

func TestDiscussionComments_Create(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := dbtesting.TestContext(t)

	user, err := Users.Create(ctx, NewUser{
		Email:                 "a@a.com",
		Username:              "u",
		Password:              "p",
		EmailVerificationCode: "c",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create a repository to comply with the postgres repo constraint.
	if err := Repos.Upsert(ctx, api.InsertRepoOp{URI: "myrepo", Description: "", Fork: false, Enabled: true}); err != nil {
		t.Fatal(err)
	}
	repo, err := Repos.GetByURI(ctx, "myrepo")
	if err != nil {
		t.Fatal(err)
	}

	// Create the thread.
	thread, err := DiscussionThreads.Create(ctx, &types.DiscussionThread{
		AuthorUserID: user.ID,
		Title:        "Hello world!",
		TargetRepo: &types.DiscussionThreadTargetRepo{
			RepoID:   repo.ID,
			Path:     strPtr("foo/bar/mux.go"),
			Branch:   strPtr("master"),
			Revision: strPtr("0c1a96370c1a96370c1a96370c1a96370c1a9637"),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create the comment.
	comment, err := DiscussionComments.Create(ctx, &types.DiscussionComment{
		ThreadID:     thread.ID,
		AuthorUserID: user.ID,
		Contents:     "What do you think of Hello World as a Service?",
	})
	if err != nil {
		t.Fatal(err)
	}
	if comment.CreatedAt == (time.Time{}) {
		t.Fatal("expected CreatedAt to be set, got zero value time")
	}
}
