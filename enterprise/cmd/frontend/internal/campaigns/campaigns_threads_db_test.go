package campaigns

import (
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/actor"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/commentobjectdb"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threads"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
)

func TestDB_CampaignsThreads(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	resetMocks()
	ctx := dbtesting.TestContext(t)

	// Create campaign.
	user1, err := db.Users.Create(ctx, db.NewUser{Username: "user1"})
	if err != nil {
		t.Fatal(err)
	}
	org1, err := db.Orgs.Create(ctx, "org1", nil)
	if err != nil {
		t.Fatal(err)
	}
	campaign0, err := dbCampaigns{}.Create(ctx,
		&dbCampaign{NamespaceOrgID: org1.ID, Name: "n0"},
		commentobjectdb.DBObjectCommentFields{Author: actor.DBColumns{UserID: user1.ID}, Body: "b0"},
	)
	if err != nil {
		t.Fatal(err)
	}

	// Create threads.
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
	thread1, err := threads.TestCreateThread(ctx, "t1", repo.ID, user1.ID)
	if err != nil {
		t.Fatal(err)
	}

	{
		// List is empty initially.
		results, err := dbCampaignsThreads{}.List(ctx, dbCampaignsThreadsListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if len(results) != 0 {
			t.Error("want empty")
		}
	}

	// Add 2 threads.
	if err := (dbCampaignsThreads{}).AddThreadsToCampaign(ctx, campaign0.ID, []int64{thread0, thread1}); err != nil {
		t.Fatal(err)
	}

	{
		// List campaigns by thread.
		results, err := dbCampaignsThreads{}.List(ctx, dbCampaignsThreadsListOptions{ThreadID: thread0})
		if err != nil {
			t.Fatal(err)
		}
		if want := []*dbCampaignThread{{Campaign: campaign0.ID, Thread: thread0}}; !reflect.DeepEqual(results, want) {
			t.Errorf("got %+v, want %+v", results, want)
		}
	}

	{
		// List threads by campaign.
		results, err := dbCampaignsThreads{}.List(ctx, dbCampaignsThreadsListOptions{CampaignID: campaign0.ID})
		if err != nil {
			t.Fatal(err)
		}
		if want := []*dbCampaignThread{{Campaign: campaign0.ID, Thread: thread0}, {Campaign: campaign0.ID, Thread: thread1}}; !reflect.DeepEqual(results, want) {
			t.Errorf("got %+v, want %+v", results, want)
		}
	}

	// Remove 2 threads.
	if err := (dbCampaignsThreads{}).RemoveThreadsFromCampaign(ctx, campaign0.ID, []int64{thread0, thread1}); err != nil {
		t.Fatal(err)
	}

	// Add back 1 thread.
	if err := (dbCampaignsThreads{}).AddThreadsToCampaign(ctx, campaign0.ID, []int64{thread1}); err != nil {
		t.Fatal(err)
	}

	{
		// List campaigns by thread.
		results, err := dbCampaignsThreads{}.List(ctx, dbCampaignsThreadsListOptions{ThreadID: thread0})
		if err != nil {
			t.Fatal(err)
		}
		if len(results) != 0 {
			t.Error("want empty")
		}
	}

	{
		// List threads by campaign.
		results, err := dbCampaignsThreads{}.List(ctx, dbCampaignsThreadsListOptions{CampaignID: campaign0.ID})
		if err != nil {
			t.Fatal(err)
		}
		if want := []*dbCampaignThread{{Campaign: campaign0.ID, Thread: thread1}}; !reflect.DeepEqual(results, want) {
			t.Errorf("got %+v, want %+v", results, want)
		}
	}
}
