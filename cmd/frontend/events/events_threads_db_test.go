package events

import (
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/commentobjectdb"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike/testutil"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
)

func TestDB_EventsThreads(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	resetMocks()
	ctx := dbtesting.TestContext(t)

	// Create event.
	user1, err := db.Users.Create(ctx, db.NewUser{Username: "user1"})
	if err != nil {
		t.Fatal(err)
	}
	org1, err := db.Orgs.Create(ctx, "org1", nil)
	if err != nil {
		t.Fatal(err)
	}
	event0, err := dbEvents{}.Create(ctx,
		&dbEvent{NamespaceOrgID: org1.ID, Name: "n0"},
		commentobjectdb.DBObjectCommentFields{AuthorUserID: user1.ID, Body: "b0"},
	)
	if err != nil {
		t.Fatal(err)
	}

	// Create threads.
	if err := db.Repos.Upsert(ctx, api.InsertRepoOp{Name: "r", Enabled: true}); err != nil {
		t.Fatal(err)
	}
	repo, err := db.Repos.GetByName(ctx, "r")
	if err != nil {
		t.Fatal(err)
	}
	thread0, err := testutil.CreateThread(ctx, "t0", repo.ID, user1.ID)
	if err != nil {
		t.Fatal(err)
	}
	thread1, err := testutil.CreateThread(ctx, "t1", repo.ID, user1.ID)
	if err != nil {
		t.Fatal(err)
	}

	{
		// List is empty initially.
		results, err := dbEventsThreads{}.List(ctx, dbEventsThreadsListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if len(results) != 0 {
			t.Error("want empty")
		}
	}

	// Add 2 threads.
	if err := (dbEventsThreads{}).AddThreadsToEvent(ctx, event0.ID, []int64{thread0, thread1}); err != nil {
		t.Fatal(err)
	}

	{
		// List events by thread.
		results, err := dbEventsThreads{}.List(ctx, dbEventsThreadsListOptions{ThreadID: thread0})
		if err != nil {
			t.Fatal(err)
		}
		if want := []*dbEventThread{{Event: event0.ID, Thread: thread0}}; !reflect.DeepEqual(results, want) {
			t.Errorf("got %+v, want %+v", results, want)
		}
	}

	{
		// List threads by event.
		results, err := dbEventsThreads{}.List(ctx, dbEventsThreadsListOptions{EventID: event0.ID})
		if err != nil {
			t.Fatal(err)
		}
		if want := []*dbEventThread{{Event: event0.ID, Thread: thread0}, {Event: event0.ID, Thread: thread1}}; !reflect.DeepEqual(results, want) {
			t.Errorf("got %+v, want %+v", results, want)
		}
	}

	// Remove 2 labels.
	if err := (dbEventsThreads{}).RemoveThreadsFromEvent(ctx, event0.ID, []int64{thread0, thread1}); err != nil {
		t.Fatal(err)
	}

	// Add back 1 label.
	if err := (dbEventsThreads{}).AddThreadsToEvent(ctx, event0.ID, []int64{thread1}); err != nil {
		t.Fatal(err)
	}

	{
		// List events by thread.
		results, err := dbEventsThreads{}.List(ctx, dbEventsThreadsListOptions{ThreadID: thread0})
		if err != nil {
			t.Fatal(err)
		}
		if len(results) != 0 {
			t.Error("want empty")
		}
	}

	{
		// List threads by event.
		results, err := dbEventsThreads{}.List(ctx, dbEventsThreadsListOptions{EventID: event0.ID})
		if err != nil {
			t.Fatal(err)
		}
		if want := []*dbEventThread{{Event: event0.ID, Thread: thread1}}; !reflect.DeepEqual(results, want) {
			t.Errorf("got %+v, want %+v", results, want)
		}
	}
}
