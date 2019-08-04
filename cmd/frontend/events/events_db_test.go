package events

import (
	"reflect"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
)

func TestDB_Events(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	resetMocks()
	ctx := dbtesting.TestContext(t)

	// for testing equality of all other fields
	norm := func(vs ...*dbEvent) {
		for _, v := range vs {
			v.ID = 0
			v.CreatedAt = time.Time{}
		}
	}

	user1, err := db.Users.Create(ctx, db.NewUser{Username: "user1"})
	if err != nil {
		t.Fatal(err)
	}

	wantEvent0 := &dbEvent{Type: "t0", ActorUserID: user1.ID}
	event0, err := dbEvents{}.Create(ctx, nil, wantEvent0)
	if err != nil {
		t.Fatal(err)
	}
	event1, err := dbEvents{}.Create(ctx, nil,
		&dbEvent{Type: "t1", ActorUserID: user1.ID, Objects: Objects{User: user1.ID}},
	)
	if err != nil {
		t.Fatal(err)
	}
	{
		// Check Create result.
		if event0.ID == 0 {
			t.Error("got ID == 0, want non-zero")
		}
		tmp := event0.ID
		norm(event0)
		if !reflect.DeepEqual(event0, wantEvent0) {
			t.Errorf("got %+v, want %+v", event0, wantEvent0)
		}
		event0.ID = tmp
	}

	{
		// List all events.
		ts, err := dbEvents{}.List(ctx, dbEventsListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		const want = 2
		if len(ts) != want {
			t.Errorf("got %d events, want %d", len(ts), want)
		}
		count, err := dbEvents{}.Count(ctx, dbEventsListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if count != want {
			t.Errorf("got %d, want %d", count, want)
		}
	}

	{
		// Query by object.
		ts, err := dbEvents{}.List(ctx, dbEventsListOptions{Objects: Objects{User: user1.ID}})
		if err != nil {
			t.Fatal(err)
		}
		norm(ts...)
		norm(event1)
		if want := []*dbEvent{event1}; !reflect.DeepEqual(ts, want) {
			t.Errorf("got %+v, want %+v", ts, want)
		}
	}

	{
		// Delete by object.
		if err := (dbEvents{}).Delete(ctx, nil, dbEventsListOptions{Objects: Objects{User: user1.ID}}); err != nil {
			t.Fatal(err)
		}
		n, err := dbEvents{}.Count(ctx, dbEventsListOptions{Objects: Objects{User: user1.ID}})
		if err != nil {
			t.Fatal(err)
		}
		if n != 0 {
			t.Errorf("got %d events, want 0 after deleting", n)
		}
	}
}

func strptr(s string) *string { return &s }
