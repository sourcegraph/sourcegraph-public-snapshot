package events

import (
	"reflect"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/commentobjectdb"
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
			v.PrimaryCommentID = 0
			v.CreatedAt = time.Time{}
			v.UpdatedAt = time.Time{}
		}
	}

	user1, err := db.Users.Create(ctx, db.NewUser{Username: "user1"})
	if err != nil {
		t.Fatal(err)
	}
	org1, err := db.Orgs.Create(ctx, "org1", nil)
	if err != nil {
		t.Fatal(err)
	}

	wantEvent0 := &dbEvent{NamespaceUserID: user1.ID, Name: "n0"}
	event0, err := dbEvents{}.Create(ctx,
		wantEvent0,
		commentobjectdb.DBObjectCommentFields{AuthorUserID: user1.ID, Body: "b0"},
	)
	if err != nil {
		t.Fatal(err)
	}
	event0PrimaryCommentID := event0.PrimaryCommentID // needed below but is zeroed out by norm
	event1, err := dbEvents{}.Create(ctx,
		&dbEvent{NamespaceUserID: user1.ID, Name: "n1"},
		commentobjectdb.DBObjectCommentFields{AuthorUserID: user1.ID, Body: "b0"},
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
		// Get a event.
		event, err := dbEvents{}.GetByID(ctx, event0.ID)
		if err != nil {
			t.Fatal(err)
		}
		if event.ID == 0 {
			t.Error("got ID == 0, want non-zero")
		}
		norm(event)
		if !reflect.DeepEqual(event, wantEvent0) {
			t.Errorf("got %+v, want %+v", event, wantEvent0)
		}
	}

	{
		// Get the event primary comment.
		comment, err := comments.DBGetByID(ctx, event0PrimaryCommentID)
		if err != nil {
			t.Fatal(err)
		}
		if comment.Object.EventID != event0.ID {
			t.Errorf("got %d, want %d", comment.Object.EventID, event0.ID)
		}
		if want := "b0"; comment.Body != want {
			t.Errorf("got %q, want %q", comment.Body, want)
		}
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
		// List user1's events.
		ts, err := dbEvents{}.List(ctx, dbEventsListOptions{NamespaceUserID: user1.ID})
		if err != nil {
			t.Fatal(err)
		}
		const want = 2
		if len(ts) != want {
			t.Errorf("got %d events, want %d", len(ts), want)
		}
	}

	{
		// List proj2's events.
		ts, err := dbEvents{}.List(ctx, dbEventsListOptions{NamespaceOrgID: org1.ID})
		if err != nil {
			t.Fatal(err)
		}
		const want = 0
		if len(ts) != want {
			t.Errorf("got %d events, want %d", len(ts), want)
		}
	}

	{
		// Query events.
		ts, err := dbEvents{}.List(ctx, dbEventsListOptions{Query: "n1"})
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
		// Delete a event.
		if err := (dbEvents{}).DeleteByID(ctx, event0.ID); err != nil {
			t.Fatal(err)
		}
		ts, err := dbEvents{}.List(ctx, dbEventsListOptions{NamespaceUserID: user1.ID})
		if err != nil {
			t.Fatal(err)
		}
		const want = 1
		if len(ts) != want {
			t.Errorf("got %d events, want %d", len(ts), want)
		}
	}
}

func strptr(s string) *string { return &s }
