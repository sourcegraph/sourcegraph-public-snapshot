package a8n

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/a8n"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

func init() {
	dbtesting.DBNameSuffix = "a8nenterpriserdb"
}

func TestService(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := backend.WithAuthzBypass(context.Background())
	dbtesting.SetupGlobalTestDB(t)

	now := time.Now().UTC().Truncate(time.Microsecond)
	clock := func() time.Time {
		return now.UTC().Truncate(time.Microsecond)
	}

	u, err := db.Users.Create(ctx, db.NewUser{
		Email:                 "thorsten@sourcegraph.com",
		Username:              "thorsten",
		DisplayName:           "thorsten",
		Password:              "1234",
		EmailVerificationCode: "foobar",
	})
	if err != nil {
		t.Fatal(err)
	}

	store := NewStoreWithClock(dbconn.Global, clock)

	var rs []*repos.Repo
	for i := 0; i < 3; i++ {
		rs = append(rs, testRepo(i, github.ServiceType))
	}

	reposStore := repos.NewDBStore(dbconn.Global, sql.TxOptions{})
	err = reposStore.UpsertRepos(ctx, rs...)
	if err != nil {
		t.Fatal(err)
	}

	testPlan := &a8n.CampaignPlan{CampaignType: "test", Arguments: `{}`}
	err = store.CreateCampaignPlan(ctx, testPlan)
	if err != nil {
		t.Fatal(err)
	}

	campaignJobs := make([]*a8n.CampaignJob, 0, len(rs))
	for _, repo := range rs {
		campaignJob := &a8n.CampaignJob{
			CampaignPlanID: testPlan.ID,
			RepoID:         int32(repo.ID),
			Rev:            "deadbeef",
			BaseRef:        "refs/heads/master",
			Diff:           "cool diff",
			StartedAt:      now,
			FinishedAt:     now,
		}
		err := store.CreateCampaignJob(ctx, campaignJob)
		if err != nil {
			t.Fatal(err)
		}
		campaignJobs = append(campaignJobs, campaignJob)
	}

	campaign := &a8n.Campaign{
		Name:            "Testing Campaign",
		Description:     "Testing Campaign",
		AuthorID:        u.ID,
		NamespaceUserID: u.ID,
		CampaignPlanID:  testPlan.ID,
	}
	gitClient := &dummyGitserverClient{response: "testresponse", responseErr: nil}

	cf := httpcli.NewHTTPClientFactory()
	svc := NewServiceWithClock(store, gitClient, cf, clock)
	err = svc.CreateCampaign(ctx, campaign)
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.GetCampaign(ctx, GetCampaignOpts{ID: campaign.ID})
	if err != nil {
		t.Fatal(err)
	}

	haveJobs, _, err := store.ListChangesetJobs(ctx, ListChangesetJobsOpts{
		CampaignID: campaign.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(haveJobs) != len(campaignJobs) {
		t.Errorf("wrong number of ChangesetJobs: %d. want=%d", len(haveJobs), len(campaignJobs))
	}
}

type dummyGitserverClient struct {
	response    string
	responseErr error
}

func (d *dummyGitserverClient) CreateCommitFromPatch(ctx context.Context, req protocol.CreateCommitFromPatchRequest) (string, error) {
	return d.response, d.responseErr
}
