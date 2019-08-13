package campaigns

import (
	"reflect"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/actor"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/commentobjectdb"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
)

func TestDB_Campaigns(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	resetMocks()
	ctx := dbtesting.TestContext(t)

	// for testing equality of all other fields
	norm := func(vs ...*dbCampaign) {
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

	wantCampaign0 := &dbCampaign{NamespaceUserID: user1.ID, Name: "n0"}
	campaign0, err := dbCampaigns{}.Create(ctx,
		wantCampaign0,
		commentobjectdb.DBObjectCommentFields{Author: actor.DBColumns{UserID: user1.ID}, Body: "b0"},
	)
	if err != nil {
		t.Fatal(err)
	}
	campaign0PrimaryCommentID := campaign0.PrimaryCommentID // needed below but is zeroed out by norm
	campaign1, err := dbCampaigns{}.Create(ctx,
		&dbCampaign{NamespaceUserID: user1.ID, Name: "n1"},
		commentobjectdb.DBObjectCommentFields{Author: actor.DBColumns{UserID: user1.ID}, Body: "b0"},
	)
	if err != nil {
		t.Fatal(err)
	}
	{
		// Check Create result.
		if campaign0.ID == 0 {
			t.Error("got ID == 0, want non-zero")
		}
		tmp := campaign0.ID
		norm(campaign0)
		if !reflect.DeepEqual(campaign0, wantCampaign0) {
			t.Errorf("got %+v, want %+v", campaign0, wantCampaign0)
		}
		campaign0.ID = tmp
	}

	{
		// Get a campaign.
		campaign, err := dbCampaigns{}.GetByID(ctx, campaign0.ID)
		if err != nil {
			t.Fatal(err)
		}
		if campaign.ID == 0 {
			t.Error("got ID == 0, want non-zero")
		}
		norm(campaign)
		if !reflect.DeepEqual(campaign, wantCampaign0) {
			t.Errorf("got %+v, want %+v", campaign, wantCampaign0)
		}
	}

	{
		// Get the campaign primary comment.
		comment, err := comments.DBGetByID(ctx, campaign0PrimaryCommentID)
		if err != nil {
			t.Fatal(err)
		}
		if comment.Object.CampaignID != campaign0.ID {
			t.Errorf("got %d, want %d", comment.Object.CampaignID, campaign0.ID)
		}
		if want := "b0"; comment.Body != want {
			t.Errorf("got %q, want %q", comment.Body, want)
		}
	}

	{
		// List all campaigns.
		ts, err := dbCampaigns{}.List(ctx, dbCampaignsListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		const want = 2
		if len(ts) != want {
			t.Errorf("got %d campaigns, want %d", len(ts), want)
		}
		count, err := dbCampaigns{}.Count(ctx, dbCampaignsListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if count != want {
			t.Errorf("got %d, want %d", count, want)
		}
	}

	{
		// List user1's campaigns.
		ts, err := dbCampaigns{}.List(ctx, dbCampaignsListOptions{NamespaceUserID: user1.ID})
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
		ts, err := dbCampaigns{}.List(ctx, dbCampaignsListOptions{NamespaceOrgID: org1.ID})
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
		ts, err := dbCampaigns{}.List(ctx, dbCampaignsListOptions{Query: "n1"})
		if err != nil {
			t.Fatal(err)
		}
		norm(ts...)
		norm(campaign1)
		if want := []*dbCampaign{campaign1}; !reflect.DeepEqual(ts, want) {
			t.Errorf("got %+v, want %+v", ts, want)
		}
	}

	{
		// Delete a campaign.
		if err := (dbCampaigns{}).DeleteByID(ctx, campaign0.ID); err != nil {
			t.Fatal(err)
		}
		ts, err := dbCampaigns{}.List(ctx, dbCampaignsListOptions{NamespaceUserID: user1.ID})
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
