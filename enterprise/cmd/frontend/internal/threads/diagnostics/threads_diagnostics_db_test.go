package diagnostics

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threads"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
)

func TestDB_ThreadsDiagnostics(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	resetMocks()
	ctx := dbtesting.TestContext(t)

	// Create threads.
	user1, err := db.Users.Create(ctx, db.NewUser{Username: "user1"})
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
	thread0, err := threads.TestCreateThread(ctx, "t0", repo.ID, user1.ID)
	if err != nil {
		t.Fatal(err)
	}

	{
		// List is empty initially.
		results, err := dbThreadDiagnosticEdges{}.List(ctx, dbThreadDiagnosticEdgesListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if len(results) != 0 {
			t.Error("want empty")
		}
	}

	// Create thread diagnostics.
	wantTD0 := dbThreadDiagnostic{ThreadID: thread0, Type: "t0", Data: json.RawMessage(`{"a": 1}`)}
	td0, err := dbThreadDiagnosticEdges{}.Create(ctx, wantTD0)
	if err != nil {
		t.Fatal(err)
	}
	wantTD0.ID = td0
	wantTD1 := dbThreadDiagnostic{ThreadID: thread0, Type: "t1", Data: json.RawMessage(`{"a": 2}`)}
	td1, err := (dbThreadDiagnosticEdges{}).Create(ctx, wantTD1)
	if err != nil {
		t.Fatal(err)
	}
	wantTD1.ID = td1

	{
		// List diagnostics by thread.
		results, err := dbThreadDiagnosticEdges{}.List(ctx, dbThreadDiagnosticEdgesListOptions{ThreadID: thread0})
		if err != nil {
			t.Fatal(err)
		}
		if want := []*dbThreadDiagnostic{&wantTD0, &wantTD1}; !reflect.DeepEqual(results, want) {
			t.Errorf("got %+v, want %+v", results, want)
		}
	}

	// Remove 1 diagnostic.
	if err := (dbThreadDiagnosticEdges{}).DeleteByIDInThread(ctx, td0, thread0); err != nil {
		t.Fatal(err)
	}

	{
		// List diagnostics by thread.
		results, err := dbThreadDiagnosticEdges{}.List(ctx, dbThreadDiagnosticEdgesListOptions{ThreadID: thread0})
		if err != nil {
			t.Fatal(err)
		}
		if want := []*dbThreadDiagnostic{&wantTD1}; !reflect.DeepEqual(results, want) {
			t.Errorf("got %+v, want %+v", results, want)
		}
	}
}
