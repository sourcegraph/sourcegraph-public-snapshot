package internal_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/actor"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/internal"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/types"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threads"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
)

func TestDB_Comments(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	internal.ResetMocks()
	ctx := dbtesting.TestContext(t)

	// for testing equality of all other fields
	norm := func(vs ...*internal.DBComment) {
		for _, v := range vs {
			v.ID = 0
			v.CreatedAt = time.Time{}
			v.UpdatedAt = time.Time{}
		}
	}

	user, err := db.Users.Create(ctx, db.NewUser{Username: "user"})
	if err != nil {
		t.Fatal(err)
	}

	wantComment0 := &internal.DBComment{
		Object: types.CommentObject{}, // empty for test
		Author: actor.DBColumns{UserID: user.ID},
		Body:   "b0",
	}
	comment0, err := internal.DBComments{}.Create(ctx, nil, wantComment0)
	if err != nil {
		t.Fatal(err)
	}
	comment1, err := internal.DBComments{}.Create(ctx, nil, &internal.DBComment{
		Object: types.CommentObject{}, // empty for test
		Author: actor.DBColumns{UserID: user.ID},
		Body:   "b1",
	})
	if err != nil {
		t.Fatal(err)
	}

	comment0ID := comment0.ID // needed later
	norm(comment0, comment1)

	{
		// Check Create result.
		if comment0ID == 0 {
			t.Error("got ID == 0, want non-zero")
		}
		if !reflect.DeepEqual(comment0, wantComment0) {
			t.Errorf("got %+v, want %+v", comment0, wantComment0)
		}
	}

	{
		// Get a comment.
		comment, err := internal.DBComments{}.GetByID(ctx, comment0ID)
		if err != nil {
			t.Fatal(err)
		}
		if comment.ID == 0 {
			t.Error("got ID == 0, want non-zero")
		}
		norm(comment)
		if !reflect.DeepEqual(comment, wantComment0) {
			t.Errorf("got %+v, want %+v", comment, wantComment0)
		}
	}

	{
		// List all comments.
		ts, err := internal.DBComments{}.List(ctx, internal.DBCommentsListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		const want = 2
		if len(ts) != want {
			t.Errorf("got %d comments, want %d", len(ts), want)
		}
		count, err := internal.DBComments{}.Count(ctx, internal.DBCommentsListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if count != want {
			t.Errorf("got %d, want %d", count, want)
		}
	}

	{
		// Query comments.
		ts, err := internal.DBComments{}.List(ctx, internal.DBCommentsListOptions{Query: "b1"})
		if err != nil {
			t.Fatal(err)
		}
		norm(ts...)
		if want := []*internal.DBComment{comment1}; !reflect.DeepEqual(ts, want) {
			t.Errorf("got %+v, want %+v", ts, want)
		}
	}

	{
		// Delete a comment.
		if err := (internal.DBComments{}).DeleteByID(ctx, comment0ID); err != nil {
			t.Fatal(err)
		}
		ts, err := internal.DBComments{}.List(ctx, internal.DBCommentsListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		const want = 1
		if len(ts) != want {
			t.Errorf("got %d comments, want %d", len(ts), want)
		}
	}
}

func TestDB_Thread_Comments(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	internal.ResetMocks()
	ctx := dbtesting.TestContext(t)

	user, err := db.Users.Create(ctx, db.NewUser{Username: "user"})
	if err != nil {
		t.Fatal(err)
	}

	if err := db.Repos.Upsert(ctx, api.InsertRepoOp{Name: "r", Enabled: true}); err != nil {
		t.Fatal(err)
	}
	repo, err := db.Repos.GetByName(ctx, "r")
	if err != nil {
		t.Fatal(err)
	}
	thread0, err := threads.TestCreateThread(ctx, "t0", repo.ID, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	thread1, err := threads.TestCreateThread(ctx, "t1", repo.ID, user.ID)
	if err != nil {
		t.Fatal(err)
	}

	{
		// List thread0's comments.
		ts, err := internal.DBComments{}.List(ctx, internal.DBCommentsListOptions{Object: types.CommentObject{ThreadID: thread0}})
		if err != nil {
			t.Fatal(err)
		}
		const want = 1
		if len(ts) != want {
			t.Errorf("got %d comments, want %d", len(ts), want)
		}
	}

	{
		// List thread1's comments.
		ts, err := internal.DBComments{}.List(ctx, internal.DBCommentsListOptions{Object: types.CommentObject{ThreadID: thread1}})
		if err != nil {
			t.Fatal(err)
		}
		const want = 1
		if len(ts) != want {
			t.Errorf("got %d comments, want %d", len(ts), want)
		}
	}
}

func strptr(s string) *string { return &s }
