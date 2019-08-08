package labels

import (
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threads"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
)

func TestDB_LabelsObjects(t *testing.T) {
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

	// Create the thread.
	user, err := db.Users.Create(ctx, db.NewUser{Username: "u"})
	if err != nil {
		t.Fatal(err)
	}
	thread, err := threads.TestCreateThread(ctx, "t0", repo0.ID, user.ID)
	if err != nil {
		t.Fatal(err)
	}

	label0, err := dbLabels{}.Create(ctx, &dbLabel{RepositoryID: int64(repo0.ID), Name: "n0", Color: "h0"})
	if err != nil {
		t.Fatal(err)
	}
	label1, err := dbLabels{}.Create(ctx, &dbLabel{RepositoryID: int64(repo0.ID), Name: "n1", Color: "h1"})
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
	if err := (dbLabelsObjects{}).AddLabelsToThread(ctx, thread, []int64{label0.ID, label1.ID}); err != nil {
		t.Fatal(err)
	}

	{
		// List threads by label.
		results, err := dbLabelsObjects{}.List(ctx, dbLabelsObjectsListOptions{LabelID: label0.ID})
		if err != nil {
			t.Fatal(err)
		}
		if want := []*dbObjectLabel{{Label: label0.ID, Thread: thread}}; !reflect.DeepEqual(results, want) {
			t.Errorf("got %+v, want %+v", results, want)
		}
	}

	{
		// List labels by thread.
		results, err := dbLabelsObjects{}.List(ctx, dbLabelsObjectsListOptions{ThreadID: thread})
		if err != nil {
			t.Fatal(err)
		}
		if want := []*dbObjectLabel{{Label: label0.ID, Thread: thread}, {Label: label1.ID, Thread: thread}}; !reflect.DeepEqual(results, want) {
			t.Errorf("got %+v, want %+v", results, want)
		}
	}

	// Remove 2 labels.
	if err := (dbLabelsObjects{}).RemoveLabelsFromThread(ctx, thread, []int64{label0.ID, label1.ID}); err != nil {
		t.Fatal(err)
	}

	// Add back 1 label.
	if err := (dbLabelsObjects{}).AddLabelsToThread(ctx, thread, []int64{label1.ID}); err != nil {
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
		results, err := dbLabelsObjects{}.List(ctx, dbLabelsObjectsListOptions{ThreadID: thread})
		if err != nil {
			t.Fatal(err)
		}
		if want := []*dbObjectLabel{{Label: label1.ID, Thread: thread}}; !reflect.DeepEqual(results, want) {
			t.Errorf("got %+v, want %+v", results, want)
		}
	}
}
