package comments

import (
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike/testutil"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
)

func TestDB_CommentsThreads(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	resetMocks()
	ctx := dbtesting.TestContext(t)

	// Create comment.
	org1, err := db.Orgs.Create(ctx, "org1", nil)
	if err != nil {
		t.Fatal(err)
	}
	comment0, err := dbComments{}.Create(ctx, &DBComment{AuthorUserID: org1.ID, Body: "b0"})
	if err != nil {
		t.Fatal(err)
	}

	// Create threads.
	user, err := db.Users.Create(ctx, db.NewUser{Username: "u"})
	if err != nil {
		t.Fatal(err)
	}
	_ = user // TODO!(sqs)
	if err := db.Repos.Upsert(ctx, api.InsertRepoOp{Name: "r", Enabled: true}); err != nil {
		t.Fatal(err)
	}
	repo, err := db.Repos.GetByName(ctx, "r")
	if err != nil {
		t.Fatal(err)
	}
	thread0, err := testutil.CreateThread(ctx, "t0", repo.ID)
	if err != nil {
		t.Fatal(err)
	}
	thread1, err := testutil.CreateThread(ctx, "t1", repo.ID)
	if err != nil {
		t.Fatal(err)
	}

	{
		// List is empty initially.
		results, err := dbCommentsThreads{}.List(ctx, dbCommentsThreadsListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if len(results) != 0 {
			t.Error("want empty")
		}
	}

	// Add 2 threads.
	if err := (dbCommentsThreads{}).AddThreadsToComment(ctx, comment0.ID, []int64{thread0, thread1}); err != nil {
		t.Fatal(err)
	}

	{
		// List comments by thread.
		results, err := dbCommentsThreads{}.List(ctx, dbCommentsThreadsListOptions{ThreadID: thread0})
		if err != nil {
			t.Fatal(err)
		}
		if want := []*dbCommentThread{{Comment: comment0.ID, Thread: thread0}}; !reflect.DeepEqual(results, want) {
			t.Errorf("got %+v, want %+v", results, want)
		}
	}

	{
		// List threads by comment.
		results, err := dbCommentsThreads{}.List(ctx, dbCommentsThreadsListOptions{CommentID: comment0.ID})
		if err != nil {
			t.Fatal(err)
		}
		if want := []*dbCommentThread{{Comment: comment0.ID, Thread: thread0}, {Comment: comment0.ID, Thread: thread1}}; !reflect.DeepEqual(results, want) {
			t.Errorf("got %+v, want %+v", results, want)
		}
	}

	// Remove 2 labels.
	if err := (dbCommentsThreads{}).RemoveThreadsFromComment(ctx, comment0.ID, []int64{thread0, thread1}); err != nil {
		t.Fatal(err)
	}

	// Add back 1 label.
	if err := (dbCommentsThreads{}).AddThreadsToComment(ctx, comment0.ID, []int64{thread1}); err != nil {
		t.Fatal(err)
	}

	{
		// List comments by thread.
		results, err := dbCommentsThreads{}.List(ctx, dbCommentsThreadsListOptions{ThreadID: thread0})
		if err != nil {
			t.Fatal(err)
		}
		if len(results) != 0 {
			t.Error("want empty")
		}
	}

	{
		// List threads by comment.
		results, err := dbCommentsThreads{}.List(ctx, dbCommentsThreadsListOptions{CommentID: comment0.ID})
		if err != nil {
			t.Fatal(err)
		}
		if want := []*dbCommentThread{{Comment: comment0.ID, Thread: thread1}}; !reflect.DeepEqual(results, want) {
			t.Errorf("got %+v, want %+v", results, want)
		}
	}
}
