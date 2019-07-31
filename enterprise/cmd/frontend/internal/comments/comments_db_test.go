package comments

import (
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
)

func TestDB_Comments(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	resetMocks()
	ctx := dbtesting.TestContext(t)

	user1, err := db.Users.Create(ctx, db.NewUser{Username: "user1"})
	if err != nil {
		t.Fatal(err)
	}
	org1, err := db.Orgs.Create(ctx, "org1", nil)
	if err != nil {
		t.Fatal(err)
	}

	wantComment0 := &dbComment{NamespaceUserID: user1.ID, Name: "n0", Description: strptr("d0")}
	comment0, err := dbComments{}.Create(ctx, wantComment0)
	if err != nil {
		t.Fatal(err)
	}
	comment1, err := dbComments{}.Create(ctx, &dbComment{NamespaceUserID: user1.ID, Name: "n1", Description: strptr("d1")})
	if err != nil {
		t.Fatal(err)
	}
	{
		// Check Create result.
		if comment0.ID == 0 {
			t.Error("got ID == 0, want non-zero")
		}
		tmp := comment0.ID
		comment0.ID = 0 // for testing equality of all other fields
		if !reflect.DeepEqual(comment0, wantComment0) {
			t.Errorf("got %+v, want %+v", comment0, wantComment0)
		}
		comment0.ID = tmp
	}

	{
		// Get a comment.
		comment, err := dbComments{}.GetByID(ctx, comment0.ID)
		if err != nil {
			t.Fatal(err)
		}
		if comment.ID == 0 {
			t.Error("got ID == 0, want non-zero")
		}
		comment.ID = 0 // for testing equality of all other fields
		if !reflect.DeepEqual(comment, wantComment0) {
			t.Errorf("got %+v, want %+v", comment, wantComment0)
		}
	}

	{
		// List all comments.
		ts, err := dbComments{}.List(ctx, dbCommentsListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		const want = 2
		if len(ts) != want {
			t.Errorf("got %d comments, want %d", len(ts), want)
		}
		count, err := dbComments{}.Count(ctx, dbCommentsListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if count != want {
			t.Errorf("got %d, want %d", count, want)
		}
	}

	{
		// List user1's comments.
		ts, err := dbComments{}.List(ctx, dbCommentsListOptions{NamespaceUserID: user1.ID})
		if err != nil {
			t.Fatal(err)
		}
		const want = 2
		if len(ts) != want {
			t.Errorf("got %d comments, want %d", len(ts), want)
		}
	}

	{
		// List proj2's comments.
		ts, err := dbComments{}.List(ctx, dbCommentsListOptions{NamespaceOrgID: org1.ID})
		if err != nil {
			t.Fatal(err)
		}
		const want = 0
		if len(ts) != want {
			t.Errorf("got %d comments, want %d", len(ts), want)
		}
	}

	{
		// Query comments.
		ts, err := dbComments{}.List(ctx, dbCommentsListOptions{Query: "n1"})
		if err != nil {
			t.Fatal(err)
		}
		if want := []*dbComment{comment1}; !reflect.DeepEqual(ts, want) {
			t.Errorf("got %+v, want %+v", ts, want)
		}
	}

	{
		// Delete a comment.
		if err := (dbComments{}).DeleteByID(ctx, comment0.ID); err != nil {
			t.Fatal(err)
		}
		ts, err := dbComments{}.List(ctx, dbCommentsListOptions{NamespaceUserID: user1.ID})
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
