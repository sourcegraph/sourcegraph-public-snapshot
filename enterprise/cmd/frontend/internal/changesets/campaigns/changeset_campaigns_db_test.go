package campaigns

import (
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/projects"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
)

func TestDB_ChangesetCampaigns(t *testing.T) {
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
	proj2, err := projects.TestCreateProject(ctx, "p2", 0, org1.ID)
	if err != nil {
		t.Fatal(err)
	}

	wantChangesetCampaign0 := &dbChangesetCampaign{ProjectID: proj1, Name: "n0", Description: strptr("d0")}
	changesetCampaign0, err := dbChangesetCampaigns{}.Create(ctx, wantChangesetCampaign0)
	if err != nil {
		t.Fatal(err)
	}
	changesetCampaign1, err := dbChangesetCampaigns{}.Create(ctx, &dbChangesetCampaign{ProjectID: proj1, Name: "n1", Description: strptr("d1")})
	if err != nil {
		t.Fatal(err)
	}
	{
		// Check Create result.
		if changesetCampaign0.ID == 0 {
			t.Error("got ID == 0, want non-zero")
		}
		tmp := changesetCampaign0.ID
		changesetCampaign0.ID = 0 // for testing equality of all other fields
		if !reflect.DeepEqual(changesetCampaign0, wantChangesetCampaign0) {
			t.Errorf("got %+v, want %+v", changesetCampaign0, wantChangesetCampaign0)
		}
		changesetCampaign0.ID = tmp
	}

	{
		// Get a changesetCampaign.
		changesetCampaign, err := dbChangesetCampaigns{}.GetByID(ctx, changesetCampaign0.ID)
		if err != nil {
			t.Fatal(err)
		}
		if changesetCampaign.ID == 0 {
			t.Error("got ID == 0, want non-zero")
		}
		changesetCampaign.ID = 0 // for testing equality of all other fields
		if !reflect.DeepEqual(changesetCampaign, wantChangesetCampaign0) {
			t.Errorf("got %+v, want %+v", changesetCampaign, wantChangesetCampaign0)
		}
	}

	{
		// List all campaigns.
		ts, err := dbChangesetCampaigns{}.List(ctx, dbChangesetCampaignsListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		const want = 2
		if len(ts) != want {
			t.Errorf("got %d campaigns, want %d", len(ts), want)
		}
		count, err := dbChangesetCampaigns{}.Count(ctx, dbChangesetCampaignsListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if count != want {
			t.Errorf("got %d, want %d", count, want)
		}
	}

	{
		// List proj1's campaigns.
		ts, err := dbChangesetCampaigns{}.List(ctx, dbChangesetCampaignsListOptions{ProjectID: proj1})
		if err != nil {
			t.Fatal(err)
		}
		const want = 2
		if len(ts) != want {
			t.Errorf("got %d campaigns, want %d", len(ts), want)
		}
	}

	{
		// List proj2's campaigns.
		ts, err := dbChangesetCampaigns{}.List(ctx, dbChangesetCampaignsListOptions{ProjectID: proj2})
		if err != nil {
			t.Fatal(err)
		}
		const want = 0
		if len(ts) != want {
			t.Errorf("got %d campaigns, want %d", len(ts), want)
		}
	}

	{
		// Query campaigns.
		ts, err := dbChangesetCampaigns{}.List(ctx, dbChangesetCampaignsListOptions{Query: "n1"})
		if err != nil {
			t.Fatal(err)
		}
		if want := []*dbChangesetCampaign{changesetCampaign1}; !reflect.DeepEqual(ts, want) {
			t.Errorf("got %+v, want %+v", ts, want)
		}
	}

	{
		// Delete a changesetCampaign.
		if err := (dbChangesetCampaigns{}).DeleteByID(ctx, changesetCampaign0.ID); err != nil {
			t.Fatal(err)
		}
		ts, err := dbChangesetCampaigns{}.List(ctx, dbChangesetCampaignsListOptions{ProjectID: proj1})
		if err != nil {
			t.Fatal(err)
		}
		const want = 1
		if len(ts) != want {
			t.Errorf("got %d campaigns, want %d", len(ts), want)
		}
	}
}

func strptr(s string) *string { return &s }
