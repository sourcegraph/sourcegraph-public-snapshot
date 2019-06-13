package labels

import (
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/projects"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
)

func TestDB_LabelsObjects(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	resetMocks()
	ctx := dbtesting.TestContext(t)

	org1, err := db.Orgs.Create(ctx, "org1", nil)
	if err != nil {
		t.Fatal(err)
	}
	proj1, err := projects.TestCreateProject(ctx, "p1", 0, org1.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Create the thread.
	user, err := db.Users.Create(ctx, db.NewUser{Username: "u"})
	if err != nil {
		t.Fatal(err)
	}
	thread, err := db.DiscussionThreads.Create(ctx, &types.DiscussionThread{
		AuthorUserID: user.ID,
		Title:        "t",
	})
	if err != nil {
		t.Fatal(err)
	}

	label0, err := dbLabels{}.Create(ctx, &dbLabel{ProjectID: proj1, Name: "n0", Color: "h0"})
	if err != nil {
		t.Fatal(err)
	}
	label1, err := dbLabels{}.Create(ctx, &dbLabel{ProjectID: proj1, Name: "n1", Color: "h1"})
	if err != nil {
		t.Fatal(err)
	}

	{
		// List is empty initially.
		results, err := dbLabelsObjects{}.List(ctx, dbLabelsObjectsListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if len(results) != 0 {
			t.Error("want empty")
		}
	}

	// Add 2 labels.
	if err := (dbLabelsObjects{}).AddLabelsToThread(ctx, thread.ID, []int64{label0.ID, label1.ID}); err != nil {
		t.Fatal(err)
	}

	{
		// List threads by label.
		results, err := dbLabelsObjects{}.List(ctx, dbLabelsObjectsListOptions{LabelID: label0.ID})
		if err != nil {
			t.Fatal(err)
		}
		if want := []*dbObjectLabel{{Label: label0.ID, Thread: thread.ID}}; !reflect.DeepEqual(results, want) {
			t.Errorf("got %+v, want %+v", results, want)
		}
	}

	{
		// List labels by thread.
		results, err := dbLabelsObjects{}.List(ctx, dbLabelsObjectsListOptions{ThreadID: thread.ID})
		if err != nil {
			t.Fatal(err)
		}
		if want := []*dbObjectLabel{{Label: label0.ID, Thread: thread.ID}, {Label: label1.ID, Thread: thread.ID}}; !reflect.DeepEqual(results, want) {
			t.Errorf("got %+v, want %+v", results, want)
		}
	}

	// Remove 2 labels.
	if err := (dbLabelsObjects{}).RemoveLabelsFromThread(ctx, thread.ID, []int64{label0.ID, label1.ID}); err != nil {
		t.Fatal(err)
	}

	// Add back 1 label.
	if err := (dbLabelsObjects{}).AddLabelsToThread(ctx, thread.ID, []int64{label1.ID}); err != nil {
		t.Fatal(err)
	}

	{
		// List threads by label.
		results, err := dbLabelsObjects{}.List(ctx, dbLabelsObjectsListOptions{LabelID: label0.ID})
		if err != nil {
			t.Fatal(err)
		}
		if len(results) != 0 {
			t.Error("want empty")
		}
	}

	{
		// List labels by thread.
		results, err := dbLabelsObjects{}.List(ctx, dbLabelsObjectsListOptions{ThreadID: thread.ID})
		if err != nil {
			t.Fatal(err)
		}
		if want := []*dbObjectLabel{{Label: label1.ID, Thread: thread.ID}}; !reflect.DeepEqual(results, want) {
			t.Errorf("got %+v, want %+v", results, want)
		}
	}
}
