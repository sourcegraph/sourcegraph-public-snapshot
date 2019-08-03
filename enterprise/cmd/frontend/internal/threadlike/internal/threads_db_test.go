package internal

import (
	"reflect"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	commentobjectdb "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/commentobjectdb"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
)

func TestDB_Threads(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ResetMocks()
	ctx := dbtesting.TestContext(t)

	// for testing equality of all other fields
	norm := func(vs ...*DBThread) {
		for _, v := range vs {
			v.ID = 0
			v.PrimaryCommentID = 0
			v.CreatedAt = time.Time{}
			v.UpdatedAt = time.Time{}
		}
	}

	user, err := db.Users.Create(ctx, db.NewUser{Username: "user"})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Repos.Upsert(ctx, api.InsertRepoOp{Name: "r", Enabled: true}); err != nil {
		t.Fatal(err)
	}
	repo0, err := db.Repos.GetByName(ctx, "r")
	if err != nil {
		t.Fatal(err)
	}

	wantThread0 := &DBThread{Type: DBThreadTypeThread, RepositoryID: repo0.ID, Title: "t0", ExternalURL: strptr("u0")}
	thread0, err := DBThreads{}.Create(ctx, wantThread0, commentobjectdb.DBObjectCommentFields{AuthorUserID: user.ID})
	if err != nil {
		t.Fatal(err)
	}
	thread0ID := thread0.ID // needed later
	thread1, err := DBThreads{}.Create(ctx, &DBThread{
		Type:         "THREAD",
		RepositoryID: repo0.ID,
		Title:        "t1",
		ExternalURL:  strptr("u1"),
	}, commentobjectdb.DBObjectCommentFields{AuthorUserID: user.ID})
	if err != nil {
		t.Fatal(err)
	}
	norm(thread0, thread1)

	{
		// Check Create result.
		if thread0ID == 0 {
			t.Error("got ID == 0, want non-zero")
		}
		if !reflect.DeepEqual(thread0, wantThread0) {
			t.Errorf("got %+v, want %+v", thread0, wantThread0)
		}
	}

	{
		// Get a thread.
		thread, err := DBThreads{}.GetByID(ctx, thread0ID)
		if err != nil {
			t.Fatal(err)
		}
		if thread.ID == 0 {
			t.Error("got ID == 0, want non-zero")
		}
		norm(thread)
		if !reflect.DeepEqual(thread, wantThread0) {
			t.Errorf("got %+v, want %+v", thread, wantThread0)
		}
	}

	{
		// List all threads.
		ts, err := DBThreads{}.List(ctx, DBThreadsListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		const want = 2
		if len(ts) != want {
			t.Errorf("got %d threads, want %d", len(ts), want)
		}
		count, err := DBThreads{}.Count(ctx, DBThreadsListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if count != want {
			t.Errorf("got %d, want %d", count, want)
		}
	}

	{
		// List repo0's threads.
		ts, err := DBThreads{}.List(ctx, DBThreadsListOptions{RepositoryID: repo0.ID})
		if err != nil {
			t.Fatal(err)
		}
		const want = 2
		if len(ts) != want {
			t.Errorf("got %d threads, want %d", len(ts), want)
		}
	}

	{
		// Query threads.
		ts, err := DBThreads{}.List(ctx, DBThreadsListOptions{Query: "t1"})
		if err != nil {
			t.Fatal(err)
		}
		norm(ts...)
		if want := []*DBThread{thread1}; !reflect.DeepEqual(ts, want) {
			t.Errorf("got %+v, want %+v", ts, want)
		}
	}

	{
		// List threads by IDs.
		ts, err := DBThreads{}.List(ctx, DBThreadsListOptions{ThreadIDs: []int64{thread0ID}})
		if err != nil {
			t.Fatal(err)
		}
		norm(ts...)
		if want := []*DBThread{thread0}; !reflect.DeepEqual(ts, want) {
			t.Errorf("got %+v, want %+v", ts, want)
		}
	}

	{
		// List threads by empty list of IDs.
		ts, err := DBThreads{}.List(ctx, DBThreadsListOptions{ThreadIDs: []int64{}})
		if err != nil {
			t.Fatal(err)
		}
		if len(ts) != 0 {
			t.Errorf("got %+v, want empty", ts)
		}
	}

	{
		// Delete a thread.
		if err := (DBThreads{}).DeleteByID(ctx, thread0ID); err != nil {
			t.Fatal(err)
		}
		ts, err := DBThreads{}.List(ctx, DBThreadsListOptions{RepositoryID: repo0.ID})
		if err != nil {
			t.Fatal(err)
		}
		const want = 1
		if len(ts) != want {
			t.Errorf("got %d threads, want %d", len(ts), want)
		}
	}
}

func strptr(s string) *string { return &s }
