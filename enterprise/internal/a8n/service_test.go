package a8n

import (
	"context"
	"database/sql"
	"fmt"
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

		campaign := testCampaign(u.ID, plan.ID)
		svc := NewServiceWithClock(store, gitClient, nil, cf, clock)

		// Without CampaignJobs it should fail
		err = svc.CreateCampaign(ctx, campaign, false)
		if err != ErrNoCampaignJobs {
			t.Fatal("CreateCampaign did not produce expected error")
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

		// With CampaignJobs it should succeed
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

	t.Run("UpdateCampaignWithNewCampaignPlan", func(t *testing.T) {
		oldPlan := &a8n.CampaignPlan{CampaignType: "test", Arguments: `{}`}
		err = store.CreateCampaignPlan(ctx, oldPlan)
		if err != nil {
			t.Fatal(err)
		}

		// Create 3 CampaignJobs
		oldCampaignJobs := make([]*a8n.CampaignJob, 0, 3)
		oldCampaignJobsByID := make(map[int64]*a8n.CampaignJob)
		for _, repo := range rs[:3] {
			campaignJob := testCampaignJob(oldPlan.ID, repo.ID, now)
			err := store.CreateCampaignJob(ctx, campaignJob)
			if err != nil {
				t.Fatal(err)
			}
			oldCampaignJobs = append(oldCampaignJobs, campaignJob)
			oldCampaignJobsByID[campaignJob.ID] = campaignJob
		}

		campaign := testCampaign(u.ID, oldPlan.ID)

		svc := NewServiceWithClock(store, gitClient, nil, cf, clock)
		// This creates the Campaign and 3 ChangesetJobs
		err = svc.CreateCampaign(ctx, campaign, false)
		if err != nil {
			t.Fatal(err)
		}

		oldChangesetJobs, _, err := store.ListChangesetJobs(ctx, ListChangesetJobsOpts{
			CampaignID: campaign.ID,
			Limit:      -1,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Sanity check
		if len(oldChangesetJobs) != len(oldCampaignJobs) {
			t.Fatalf("wrong number of changeset jobs. want=%d, have=%d", len(oldCampaignJobs), len(oldChangesetJobs))
		}

		// Now we do what RunChangesetJobs would do, expect that we don't
		// actually create pull requests on the code host
		oldChangesets := make([]*a8n.Changeset, 0, len(oldCampaignJobs))
		for _, changesetJob := range oldChangesetJobs {
			campaignJob, ok := oldCampaignJobsByID[changesetJob.CampaignJobID]
			if !ok {
				t.Fatal("no CampaignJob found for ChangesetJob")
			}

			changeset := &a8n.Changeset{
				RepoID:              campaignJob.RepoID,
				CampaignIDs:         []int64{changesetJob.CampaignID},
				ExternalServiceType: "github",
				ExternalID:          fmt.Sprintf("ext-id-%d", changesetJob.ID),
			}

			err = store.CreateChangesets(ctx, changeset)
			if err != nil {
				t.Fatal(err)
			}

			oldChangesets = append(oldChangesets, changeset)

			changesetJob.ChangesetID = changeset.ID
			changesetJob.StartedAt = now
			changesetJob.FinishedAt = now

			err := store.UpdateChangesetJob(ctx, changesetJob)
			if err != nil {
				t.Fatal(err)
			}
		}

		// Advance clock
		oldTime := now
		now = now.Add(5 * time.Second)

		// Creating new CampaignPlan
		newPlan := &a8n.CampaignPlan{CampaignType: "test", Arguments: `{}`}
		err = store.CreateCampaignPlan(ctx, newPlan)
		if err != nil {
			t.Fatal(err)
		}

		// Create 2 new CampaignJobs, so that one will be closed
		newCampaignJobs := make([]*a8n.CampaignJob, 2)
		// TODO: Change this so we use 3 CampaignJobs but only 2/3 reference
		// existing repos and 1/3 another repo

		// First one has same RepoID, same Rev, same BaseRef, same Diff
		newCampaignJobs[0] = oldCampaignJobs[0].Clone()
		newCampaignJobs[0].CampaignPlanID = newPlan.ID

		// Second one has same RepoID, same Rev, same BaseRef, but different Diff
		newCampaignJobs[1] = oldCampaignJobs[1].Clone()
		newCampaignJobs[1].CampaignPlanID = newPlan.ID
		newCampaignJobs[1].Diff = "different diff"

		oldCampaignJobToBeDeleted := oldCampaignJobs[2]

		for _, j := range newCampaignJobs {
			err := store.CreateCampaignJob(ctx, j)
			if err != nil {
				t.Fatal(err)
			}
		}

		// Now we update the Campaign to use the new plan
		newName := "new name"
		newDescription := "new description"
		args := UpdateCampaignArgs{
			Campaign:    campaign.ID,
			Name:        &newName,
			Description: &newDescription,
			Plan:        &newPlan.ID,
		}
		updatedCampaign, err := svc.UpdateCampaign(ctx, args)
		if err != nil {
			t.Fatal(err)
		}

		if updatedCampaign.Name != newName {
			t.Fatalf("campaign name not updated. want=%q, have=%q", newName, updatedCampaign.Name)
		}
		if updatedCampaign.Description != newDescription {
			t.Fatalf("campaign description not updated. want=%q, have=%q", newDescription, updatedCampaign.Description)
		}

		if updatedCampaign.CampaignPlanID != newPlan.ID {
			t.Fatalf("campaign CampaignPlanID not updated. want=%q, have=%q", newPlan.ID, updatedCampaign.CampaignPlanID)
		}

		// Check that we now have 2 instead of 3 ChangesetJobs
		newChangesetJobs, _, err := store.ListChangesetJobs(ctx, ListChangesetJobsOpts{
			CampaignID: campaign.ID,
			Limit:      -1,
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(newChangesetJobs) != len(newCampaignJobs) {
			t.Fatalf("wrong number of new ChangesetJobs. want=%d, have=%d", len(newCampaignJobs), len(newChangesetJobs))
		}

		newChangesetJobsByCampaignJobID := make(map[int64]*a8n.ChangesetJob)
		for _, j := range newChangesetJobs {
			if j.ID == oldCampaignJobToBeDeleted.ID {
				t.Errorf("ChangesetJob should have been deleted")
			}
			newChangesetJobsByCampaignJobID[j.CampaignJobID] = j
		}

		// CampaignJob has same diff, so ChangesetJob should not be reset
		unmodified, ok := newChangesetJobsByCampaignJobID[newCampaignJobs[0].ID]
		if !ok {
			t.Fatal("ChangesetJob not found")
		}
		if unmodified.StartedAt != oldTime {
			t.Fatalf("ChangesetJob StartedAt changed. want=%v, have=%v", oldTime, unmodified.StartedAt)
		}

		if unmodified.FinishedAt != oldTime {
			t.Fatalf("ChangesetJob FinishedAt changed. want=%v, have=%v", oldTime, unmodified.FinishedAt)
		}

		// CampaignJob has new diff, so ChangesetJob should be reset
		modified, ok := newChangesetJobsByCampaignJobID[newCampaignJobs[1].ID]
		if !ok {
			t.Fatal("ChangesetJob not found")
		}
		if !modified.StartedAt.IsZero() {
			t.Fatalf("ChangesetJob StartedAt not reset. have=%v", modified.StartedAt)
		}

		if !modified.FinishedAt.IsZero() {
			t.Fatalf("ChangesetJob FinishedAt not reset. have=%v", modified.FinishedAt)
		}

		// Check Changeset with RepoID == CampaignJobToBeDeleted.RepoID is
		// detached from campaign
		var oldChangesetToBeDetached *a8n.Changeset
		for _, c := range oldChangesets {
			if c.RepoID == oldCampaignJobToBeDeleted.RepoID {
				oldChangesetToBeDetached = c
				break
			}
		}
		if oldChangesetToBeDetached == nil {
			t.Fatalf("could not find old changeset to be detached")
		}
		changeset, err := store.GetChangeset(ctx, GetChangesetOpts{ID: oldChangesetToBeDetached.ID})
		if err != nil {
			t.Fatal(err)
		}
		if len(changeset.CampaignIDs) != 0 {
			t.Fatalf("old changeset still attached to campaign")
		}
		for _, changesetID := range campaign.ChangesetIDs {
			if changesetID == changeset.ID {
				t.Fatalf("old changeset still attached to campaign")
			}
		}

		// TODO: Check that it's closed on Codehost (which we can't here
		// without mocking a lot of things, so we might need to do that
		// somewhere else)
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
