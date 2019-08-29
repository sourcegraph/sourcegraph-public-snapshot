package commitstatuses

import (
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
)

func TestDB_CommitStatusContexts(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	resetMocks()
	ctx := dbtesting.TestContext(t)

	if err := db.Repos.Upsert(ctx, api.InsertRepoOp{Name: "r0", Enabled: true}); err != nil {
		t.Fatal(err)
	}
	repo0, err := db.Repos.GetByName(ctx, "r0")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Repos.Upsert(ctx, api.InsertRepoOp{Name: "r1", Enabled: true}); err != nil {
		t.Fatal(err)
	}
	repo1, err := db.Repos.GetByName(ctx, "r1")
	if err != nil {
		t.Fatal(err)
	}

	wantCommitStatusContext0 := &dbCommitStatusContext{RepositoryID: repo0.ID, CommitOID: "oid0", Context: "c0"}
	commitStatusContext0, err := dbCommitStatusContexts{}.Create(ctx, wantCommitStatusContext0)
	if err != nil {
		t.Fatal(err)
	}
	commitStatus1, err := dbCommitStatusContexts{}.Create(ctx, &dbCommitStatusContext{RepositoryID: repo0.ID, CommitOID: "oid1", Context: "c1"})
	if err != nil {
		t.Fatal(err)
	}
	{
		// Check Create result.
		if commitStatusContext0.ID == 0 {
			t.Error("got ID == 0, want non-zero")
		}
		tmp := commitStatusContext0.ID
		commitStatusContext0.ID = 0 // for testing equality of all other fields
		if !reflect.DeepEqual(commitStatusContext0, wantCommitStatusContext0) {
			t.Errorf("got %+v, want %+v", commitStatusContext0, wantCommitStatusContext0)
		}
		commitStatusContext0.ID = tmp
	}

	{
		// Get a commitStatus.
		commitStatus, err := dbCommitStatusContexts{}.GetByID(ctx, commitStatusContext0.ID)
		if err != nil {
			t.Fatal(err)
		}
		if commitStatus.ID == 0 {
			t.Error("got ID == 0, want non-zero")
		}
		commitStatus.ID = 0 // for testing equality of all other fields
		if !reflect.DeepEqual(commitStatus, wantCommitStatusContext0) {
			t.Errorf("got %+v, want %+v", commitStatus, wantCommitStatusContext0)
		}
	}

	{
		ts, err := dbCommitStatusContexts{}.List(ctx, dbCommitStatusContextsListOptions{RepositoryID: repo1.ID, CommitOID: "c1"})
		if err != nil {
			t.Fatal(err)
		}
		if want := []*dbCommitStatusContext{commitStatus1}; !reflect.DeepEqual(ts, want) {
			t.Errorf("got %+v, want %+v", ts, want)
		}
	}

	{
		// List by other commit ID.
		ts, err := dbCommitStatusContexts{}.List(ctx, dbCommitStatusContextsListOptions{RepositoryID: repo1.ID, CommitOID: "x"})
		if err != nil {
			t.Fatal(err)
		}
		const want = 0
		if len(ts) != want {
			t.Errorf("got %d commitStatuses, want %d", len(ts), want)
		}
	}
}

func strptr(s string) *string { return &s }
