package changesets

import (
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
)

func TestDB_Changesets(t *testing.T) {
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

	wantChangeset0 := &dbChangeset{RepositoryID: repo0.ID, Title: "t0", ExternalURL: strptr("u0")}
	changeset0, err := dbChangesets{}.Create(ctx, wantChangeset0)
	if err != nil {
		t.Fatal(err)
	}
	changeset1, err := dbChangesets{}.Create(ctx, &dbChangeset{
		RepositoryID: repo0.ID,
		Title:        "t1",
		ExternalURL:  strptr("u1"),
	})
	if err != nil {
		t.Fatal(err)
	}

	{
		// Check Create result.
		if changeset0.ID == 0 {
			t.Error("got ID == 0, want non-zero")
		}
		tmp := changeset0.ID
		changeset0.ID = 0 // for testing equality of all other fields
		if !reflect.DeepEqual(changeset0, wantChangeset0) {
			t.Errorf("got %+v, want %+v", changeset0, wantChangeset0)
		}
		changeset0.ID = tmp
	}

	{
		// Get a changeset.
		changeset, err := dbChangesets{}.GetByID(ctx, changeset0.ID)
		if err != nil {
			t.Fatal(err)
		}
		if changeset.ID == 0 {
			t.Error("got ID == 0, want non-zero")
		}
		changeset.ID = 0 // for testing equality of all other fields
		if !reflect.DeepEqual(changeset, wantChangeset0) {
			t.Errorf("got %+v, want %+v", changeset, wantChangeset0)
		}
	}

	{
		// List all changesets.
		ts, err := dbChangesets{}.List(ctx, dbChangesetsListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		const want = 2
		if len(ts) != want {
			t.Errorf("got %d changesets, want %d", len(ts), want)
		}
		count, err := dbChangesets{}.Count(ctx, dbChangesetsListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if count != want {
			t.Errorf("got %d, want %d", count, want)
		}
	}

	{
		// List repo0's changesets.
		ts, err := dbChangesets{}.List(ctx, dbChangesetsListOptions{RepositoryID: repo0.ID})
		if err != nil {
			t.Fatal(err)
		}
		const want = 2
		if len(ts) != want {
			t.Errorf("got %d changesets, want %d", len(ts), want)
		}
	}

	{
		// Query changesets.
		ts, err := dbChangesets{}.List(ctx, dbChangesetsListOptions{Query: "t1"})
		if err != nil {
			t.Fatal(err)
		}
		if want := []*dbChangeset{changeset1}; !reflect.DeepEqual(ts, want) {
			t.Errorf("got %+v, want %+v", ts, want)
		}
	}

	{
		// List changesets by IDs.
		ts, err := dbChangesets{}.List(ctx, dbChangesetsListOptions{ChangesetIDs: []int64{changeset0.ID}})
		if err != nil {
			t.Fatal(err)
		}
		if want := []*dbChangeset{changeset0}; !reflect.DeepEqual(ts, want) {
			t.Errorf("got %+v, want %+v", ts, want)
		}
	}

	{
		// List changesets by empty list of IDs.
		ts, err := dbChangesets{}.List(ctx, dbChangesetsListOptions{ChangesetIDs: []int64{}})
		if err != nil {
			t.Fatal(err)
		}
		if len(ts) != 0 {
			t.Errorf("got %+v, want empty", ts)
		}
	}

	{
		// Delete a changeset.
		if err := (dbChangesets{}).DeleteByID(ctx, changeset0.ID); err != nil {
			t.Fatal(err)
		}
		ts, err := dbChangesets{}.List(ctx, dbChangesetsListOptions{RepositoryID: repo0.ID})
		if err != nil {
			t.Fatal(err)
		}
		const want = 1
		if len(ts) != want {
			t.Errorf("got %d changesets, want %d", len(ts), want)
		}
	}
}

func strptr(s string) *string { return &s }
