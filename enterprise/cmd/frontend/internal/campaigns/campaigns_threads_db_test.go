package campaigns

import (
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike/testutil"
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
	org1, err := db.Orgs.Create(ctx, "org1", nil)
	if err != nil {
		t.Fatal(err)
	}
	campaign0, err := dbCampaigns{}.Create(ctx, &dbCampaign{NamespaceOrgID: org1.ID, Name: "n0"})
	if err != nil {
		t.Fatal(err)
	}

	// Create threads.
	user, err := db.Users.Create(ctx, db.NewUser{Username: "u"})
	if err != nil {
		t.Fatal(err)
	}
	_ = user // TODO!(sqs)
	if err := db.Repos.Upsert(ctx, api.InsertRepoOp{Name: "r", Enabled: true}); err != nil {
		t.Fatal(err)
	}
	repo, err := db.Repos.GetByName(ctx, "r")
	if err != nil {
		t.Fatal(err)
	}
	thread0, err := testutil.CreateThread(ctx, "t0", repo.ID)
	if err != nil {
		t.Fatal(err)
	}
	thread1, err := testutil.CreateThread(ctx, "t1", repo.ID)
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

	// Remove 2 labels.
	if err := (dbCampaignsThreads{}).RemoveThreadsFromCampaign(ctx, campaign0.ID, []int64{thread0, thread1}); err != nil {
		t.Fatal(err)
	}

	// Add back 1 label.
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
