package a8n

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/a8n"
	"github.com/sourcegraph/sourcegraph/internal/api"
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

	gitClient := &dummyGitserverClient{response: "testresponse", responseErr: nil}
	cf := httpcli.NewHTTPClientFactory()

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

	t.Run("CreateCampaignPlanFromPatches", func(t *testing.T) {
		const commit = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		repoResolveRevision := func(context.Context, *repos.Repo, string) (api.CommitID, error) {
			return commit, nil
		}

		svc := NewServiceWithClock(store, nil, repoResolveRevision, nil, clock)

		const patch = `diff f f
--- f
+++ f
@@ -1,1 +1,2 @@
+x
 y
`
		patches := []a8n.CampaignPlanPatch{
			{Repo: api.RepoID(rs[0].ID), BaseRevision: "b0", Patch: patch},
			{Repo: api.RepoID(rs[1].ID), BaseRevision: "b1", Patch: patch},
		}

		plan, err := svc.CreateCampaignPlanFromPatches(ctx, patches)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := store.GetCampaignPlan(ctx, GetCampaignPlanOpts{ID: plan.ID}); err != nil {
			t.Fatal(err)
		}

		jobs, _, err := store.ListCampaignJobs(ctx, ListCampaignJobsOpts{CampaignPlanID: plan.ID})
		if err != nil {
			t.Fatal(err)
		}
		for _, job := range jobs {
			job.ID = 0 // ignore database ID when checking for expected output
		}
		wantJobs := make([]*a8n.CampaignJob, len(patches))
		for i, patch := range patches {
			wantJobs[i] = &a8n.CampaignJob{
				CampaignPlanID: plan.ID,
				RepoID:         int32(patch.Repo),
				BaseRef:        patch.BaseRevision,
				Rev:            commit,
				Diff:           patch.Patch,
				StartedAt:      now,
				FinishedAt:     now,
				CreatedAt:      now,
				UpdatedAt:      now,
			}
		}
		if !cmp.Equal(jobs, wantJobs) {
			t.Error("jobs != wantJobs", cmp.Diff(jobs, wantJobs))
		}
	})

	t.Run("CreateCampaign", func(t *testing.T) {
		plan := &a8n.CampaignPlan{CampaignType: "test", Arguments: `{}`}
		err = store.CreateCampaignPlan(ctx, plan)
		if err != nil {
			t.Fatal(err)
		}

		campaignJobs := make([]*a8n.CampaignJob, 0, len(rs))
		for _, repo := range rs {
			campaignJob := testCampaignJob(plan.ID, repo.ID, now)
			err := store.CreateCampaignJob(ctx, campaignJob)
			if err != nil {
				t.Fatal(err)
			}
			campaignJobs = append(campaignJobs, campaignJob)
		}

		campaign := testCampaign(u.ID, plan.ID)

		svc := NewServiceWithClock(store, gitClient, nil, cf, clock)
		err = svc.CreateCampaign(ctx, campaign, false)
		if err != nil {
			t.Fatal(err)
		}

		if campaign.PublishedAt.IsZero() {
			t.Fatalf("PublishedAt is zero")
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
	})

	t.Run("CreateCampaignAsDraft", func(t *testing.T) {
		plan := &a8n.CampaignPlan{CampaignType: "test", Arguments: `{}`}
		err = store.CreateCampaignPlan(ctx, plan)
		if err != nil {
			t.Fatal(err)
		}

		for _, repo := range rs {
			campaignJob := testCampaignJob(plan.ID, repo.ID, now)
			err := store.CreateCampaignJob(ctx, campaignJob)
			if err != nil {
				t.Fatal(err)
			}
		}

		campaign := testCampaign(u.ID, plan.ID)

		svc := NewServiceWithClock(store, gitClient, nil, cf, clock)
		err = svc.CreateCampaign(ctx, campaign, true)
		if err != nil {
			t.Fatal(err)
		}

		if !campaign.PublishedAt.IsZero() {
			t.Fatalf("PublishedAt is not zero")
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

		if len(haveJobs) != 0 {
			t.Errorf("wrong number of ChangesetJobs: %d. want=%d", len(haveJobs), 0)
		}
	})

	t.Run("CreateChangesetJobForCampaignJob", func(t *testing.T) {
		plan := &a8n.CampaignPlan{CampaignType: "test", Arguments: `{}`}
		err = store.CreateCampaignPlan(ctx, plan)
		if err != nil {
			t.Fatal(err)
		}

		campaignJob := testCampaignJob(plan.ID, rs[0].ID, now)
		err := store.CreateCampaignJob(ctx, campaignJob)
		if err != nil {
			t.Fatal(err)
		}

		campaign := testCampaign(u.ID, plan.ID)
		err = store.CreateCampaign(ctx, campaign)
		if err != nil {
			t.Fatal(err)
		}

		svc := NewServiceWithClock(store, gitClient, nil, cf, clock)
		changesetJob, _, err := svc.CreateChangesetJobForCampaignJob(ctx, campaignJob.ID)
		if err != nil {
			t.Fatal(err)
		}

		haveJob, err := store.GetChangesetJob(ctx, GetChangesetJobOpts{
			CampaignID:    campaign.ID,
			CampaignJobID: campaignJob.ID,
		})
		if err != nil {
			t.Fatal(err)
		}

		if changesetJob.ID != haveJob.ID {
			t.Errorf("wrong changesetJob: %d. want=%d", changesetJob.ID, haveJob.ID)
		}

		// Try to create again, check that it's the same one
		changesetJob2, _, err := svc.CreateChangesetJobForCampaignJob(ctx, campaignJob.ID)
		if err != nil {
			t.Fatal(err)
		}

		if changesetJob2.ID != haveJob.ID {
			t.Errorf("wrong changesetJob: %d. want=%d", changesetJob2.ID, haveJob.ID)
		}
	})
}

func testCampaignJob(plan int64, repo uint32, t time.Time) *a8n.CampaignJob {
	return &a8n.CampaignJob{
		CampaignPlanID: plan,
		RepoID:         int32(repo),
		Rev:            "deadbeef",
		BaseRef:        "refs/heads/master",
		Diff:           "cool diff",
		StartedAt:      t,
		FinishedAt:     t,
	}
}

func testCampaign(user int32, plan int64) *a8n.Campaign {
	return &a8n.Campaign{
		Name:            "Testing Campaign",
		Description:     "Testing Campaign",
		AuthorID:        user,
		NamespaceUserID: user,
		CampaignPlanID:  plan,
	}
}

type dummyGitserverClient struct {
	response    string
	responseErr error
}

func (d *dummyGitserverClient) CreateCommitFromPatch(ctx context.Context, req protocol.CreateCommitFromPatchRequest) (string, error) {
	return d.response, d.responseErr
}
