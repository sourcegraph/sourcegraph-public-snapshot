package projects

import (
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
)

func TestDB_Projects(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	resetMocks()
	ctx := dbtesting.TestContext(t)

	org1, err := db.Orgs.Create(ctx, "org1", nil)
	if err != nil {
		t.Fatal(err)
	}
	org2, err := db.Orgs.Create(ctx, "org2", nil)
	if err != nil {
		t.Fatal(err)
	}

	wantProject0 := &dbProject{NamespaceOrgID: org1.ID, Name: "n0"}
	project0, err := dbProjects{}.Create(ctx, wantProject0)
	if err != nil {
		t.Fatal(err)
	}
	project1, err := dbProjects{}.Create(ctx, &dbProject{NamespaceOrgID: org1.ID, Name: "n1"})
	if err != nil {
		t.Fatal(err)
	}
	{
		// Check Create result.
		if project0.ID == 0 {
			t.Error("got ID == 0, want non-zero")
		}
		tmp := project0.ID
		project0.ID = 0 // for testing equality of all other fields
		if !reflect.DeepEqual(project0, wantProject0) {
			t.Errorf("got %+v, want %+v", project0, wantProject0)
		}
		project0.ID = tmp
	}

	{
		// Get a project.
		project, err := dbProjects{}.GetByID(ctx, project0.ID)
		if err != nil {
			t.Fatal(err)
		}
		if project.ID == 0 {
			t.Error("got ID == 0, want non-zero")
		}
		project.ID = 0 // for testing equality of all other fields
		if !reflect.DeepEqual(project, wantProject0) {
			t.Errorf("got %+v, want %+v", project, wantProject0)
		}
	}

	{
		// List all projects.
		ts, err := dbProjects{}.List(ctx, dbProjectsListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		const want = 2
		if len(ts) != want {
			t.Errorf("got %d projects, want %d", len(ts), want)
		}
		count, err := dbProjects{}.Count(ctx, dbProjectsListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if count != want {
			t.Errorf("got %d, want %d", count, want)
		}
	}

	{
		// List proj1's projects.
		ts, err := dbProjects{}.List(ctx, dbProjectsListOptions{NamespaceOrgID: org1.ID})
		if err != nil {
			t.Fatal(err)
		}
		const want = 2
		if len(ts) != want {
			t.Errorf("got %d projects, want %d", len(ts), want)
		}
	}

	{
		// List proj2's projects.
		ts, err := dbProjects{}.List(ctx, dbProjectsListOptions{NamespaceOrgID: org2.ID})
		if err != nil {
			t.Fatal(err)
		}
		const want = 0
		if len(ts) != want {
			t.Errorf("got %d projects, want %d", len(ts), want)
		}
	}

	{
		// Query projects.
		ts, err := dbProjects{}.List(ctx, dbProjectsListOptions{Query: "n1"})
		if err != nil {
			t.Fatal(err)
		}
		if want := []*dbProject{project1}; !reflect.DeepEqual(ts, want) {
			t.Errorf("got %+v, want %+v", ts, want)
		}
	}

	{
		// Delete a project.
		if err := (dbProjects{}).DeleteByID(ctx, project0.ID); err != nil {
			t.Fatal(err)
		}
		ts, err := dbProjects{}.List(ctx, dbProjectsListOptions{NamespaceOrgID: org1.ID})
		if err != nil {
			t.Fatal(err)
		}
		const want = 1
		if len(ts) != want {
			t.Errorf("got %d projects, want %d", len(ts), want)
		}
	}
}

func strptr(s string) *string { return &s }
