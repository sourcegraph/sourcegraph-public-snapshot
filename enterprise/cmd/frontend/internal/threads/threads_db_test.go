package threads

import (
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
)

func TestDB_Threads(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	resetMocks()
	ctx := dbtesting.TestContext(t)

	if err := db.Repos.Upsert(ctx, api.InsertRepoOp{Name: "r", Enabled: true}); err != nil {
		t.Fatal(err)
	}
	repo0, err := db.Repos.GetByName(ctx, "r")
	if err != nil {
		t.Fatal(err)
	}

	wantThread0 := &dbThread{DBThreadCommon: DBThreadCommon{RepositoryID: repo0.ID, Title: "t0", ExternalURL: strptr("u0")}}
	thread0, err := dbThreads{}.Create(ctx, wantThread0)
	if err != nil {
		t.Fatal(err)
	}
	thread1, err := dbThreads{}.Create(ctx, &dbThread{
		DBThreadCommon: DBThreadCommon{
			RepositoryID: repo0.ID,
			Title:        "t1",
			ExternalURL:  strptr("u1"),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	{
		// Check Create result.
		if thread0.ID == 0 {
			t.Error("got ID == 0, want non-zero")
		}
		tmp := thread0.ID
		thread0.ID = 0 // for testing equality of all other fields
		if !reflect.DeepEqual(thread0, wantThread0) {
			t.Errorf("got %+v, want %+v", thread0, wantThread0)
		}
		thread0.ID = tmp
	}

	{
		// Get a thread.
		thread, err := dbThreads{}.GetByID(ctx, thread0.ID)
		if err != nil {
			t.Fatal(err)
		}
		if thread.ID == 0 {
			t.Error("got ID == 0, want non-zero")
		}
		thread.ID = 0 // for testing equality of all other fields
		if !reflect.DeepEqual(thread, wantThread0) {
			t.Errorf("got %+v, want %+v", thread, wantThread0)
		}
	}

	{
		// List all threads.
		ts, err := dbThreads{}.List(ctx, dbThreadsListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		const want = 2
		if len(ts) != want {
			t.Errorf("got %d threads, want %d", len(ts), want)
		}
		count, err := dbThreads{}.Count(ctx, dbThreadsListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if count != want {
			t.Errorf("got %d, want %d", count, want)
		}
	}

	{
		// List repo0's threads.
		ts, err := dbThreads{}.List(ctx, dbThreadsListOptions{common: DBThreadsListOptionsCommon{RepositoryID: repo0.ID}})
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
		ts, err := dbThreads{}.List(ctx, dbThreadsListOptions{common: DBThreadsListOptionsCommon{Query: "t1"}})
		if err != nil {
			t.Fatal(err)
		}
		if want := []*dbThread{thread1}; !reflect.DeepEqual(ts, want) {
			t.Errorf("got %+v, want %+v", ts, want)
		}
	}

	{
		// List threads by IDs.
		ts, err := dbThreads{}.List(ctx, dbThreadsListOptions{common: DBThreadsListOptionsCommon{ThreadIDs: []int64{thread0.ID}}})
		if err != nil {
			t.Fatal(err)
		}
		if want := []*dbThread{thread0}; !reflect.DeepEqual(ts, want) {
			t.Errorf("got %+v, want %+v", ts, want)
		}
	}

	{
		// List threads by empty list of IDs.
		ts, err := dbThreads{}.List(ctx, dbThreadsListOptions{common: DBThreadsListOptionsCommon{ThreadIDs: []int64{}}})
		if err != nil {
			t.Fatal(err)
		}
		if len(ts) != 0 {
			t.Errorf("got %+v, want empty", ts)
		}
	}

	{
		// Delete a thread.
		if err := (dbThreads{}).DeleteByID(ctx, thread0.ID); err != nil {
			t.Fatal(err)
		}
		ts, err := dbThreads{}.List(ctx, dbThreadsListOptions{common: DBThreadsListOptionsCommon{RepositoryID: repo0.ID}})
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
