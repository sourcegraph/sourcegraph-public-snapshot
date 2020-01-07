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
	for i := 0; i < 4; i++ {
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
}

func TestService_UpdateCampaignWithNewCampaignPlanID(t *testing.T) {
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
	for i := 0; i < 4; i++ {
		rs = append(rs, testRepo(i, github.ServiceType))
	}

	reposStore := repos.NewDBStore(dbconn.Global, sql.TxOptions{})
	err = reposStore.UpsertRepos(ctx, rs...)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name                                string
		oldCampaignJobs                     func(currentPlanID int64) []*a8n.CampaignJob
		newCampaignJobs                     func(newPlanID int64, oldCampaignJobs []*a8n.CampaignJob) []*a8n.CampaignJob
		wantCampaignJobsWithoutChangesetJob func(oldCampaignJobs []*a8n.CampaignJob) []*a8n.CampaignJob
		wantUnmodifiedChangesetJobs         func(changesetJobs []*a8n.ChangesetJob, newCampaignJobs []*a8n.CampaignJob) []*a8n.ChangesetJob
		wantModifiedChangesetJobs           func(changesetJobs []*a8n.ChangesetJob, newCampaignJobs []*a8n.CampaignJob) (jobs []*a8n.ChangesetJob)
		wantCreatedChangesetJobs            func(changesetJobs []*a8n.ChangesetJob, newCampaignJobs []*a8n.CampaignJob) (jobs []*a8n.ChangesetJob)
	}{
		{
			name: "1 unmodified",
			oldCampaignJobs: func(plan int64) (jobs []*a8n.CampaignJob) {
				jobs = append(jobs, testCampaignJob(plan, rs[0].ID, now))
				return jobs
			},
			newCampaignJobs: func(plan int64, oldCampaignJobs []*a8n.CampaignJob) []*a8n.CampaignJob {
				newJobs := make([]*a8n.CampaignJob, 1)
				newJobs[0] = oldCampaignJobs[0].Clone()
				newJobs[0].CampaignPlanID = plan
				return newJobs
			},
			wantCampaignJobsWithoutChangesetJob: func(oldCampaignJobs []*a8n.CampaignJob) []*a8n.CampaignJob {
				return []*a8n.CampaignJob{}
			},
			wantUnmodifiedChangesetJobs: func(changesetJobs []*a8n.ChangesetJob, newCampaignJobs []*a8n.CampaignJob) (jobs []*a8n.ChangesetJob) {
				for _, j := range changesetJobs {
					// CampaignJob has same diff, so ChangesetJob should not be reset
					if j.CampaignJobID == newCampaignJobs[0].ID {
						jobs = append(jobs, j)
					}
				}
				return jobs
			},
			wantModifiedChangesetJobs: func(changesetJobs []*a8n.ChangesetJob, newCampaignJobs []*a8n.CampaignJob) (jobs []*a8n.ChangesetJob) {
				return jobs
			},
			wantCreatedChangesetJobs: func(changesetJobs []*a8n.ChangesetJob, newCampaignJobs []*a8n.CampaignJob) (jobs []*a8n.ChangesetJob) {
				return jobs
			},
		},
		{
			name: "1 unmodified, 1 modified, 1 new changeset",
			oldCampaignJobs: func(plan int64) (jobs []*a8n.CampaignJob) {
				for _, repo := range rs[:3] {
					jobs = append(jobs, testCampaignJob(plan, repo.ID, now))
				}
				return jobs
			},
			newCampaignJobs: func(plan int64, oldCampaignJobs []*a8n.CampaignJob) []*a8n.CampaignJob {
				newJobs := make([]*a8n.CampaignJob, 3)
				// First one has same RepoID, same Rev, same BaseRef, same Diff
				newJobs[0] = oldCampaignJobs[0].Clone()
				newJobs[0].CampaignPlanID = plan

				// Second one has same RepoID, same Rev, same BaseRef, but different Diff
				newJobs[1] = oldCampaignJobs[1].Clone()
				newJobs[1].CampaignPlanID = plan
				newJobs[1].Diff = "different diff"

				// Third one has new RepoID (we only created 3 CampaignJobs, but rs has
				// 4 entries)
				newJobs[2] = testCampaignJob(plan, rs[len(rs)-1].ID, now)

				return newJobs
			},
			wantCampaignJobsWithoutChangesetJob: func(oldCampaignJobs []*a8n.CampaignJob) []*a8n.CampaignJob {
				return []*a8n.CampaignJob{
					// Doesn't have a matching CampaignJob in newCampaignJobs
					oldCampaignJobs[2],
				}
			},
			wantUnmodifiedChangesetJobs: func(changesetJobs []*a8n.ChangesetJob, newCampaignJobs []*a8n.CampaignJob) (jobs []*a8n.ChangesetJob) {
				for _, j := range changesetJobs {
					// CampaignJob has same diff, so ChangesetJob should not be reset
					if j.CampaignJobID == newCampaignJobs[0].ID {
						jobs = append(jobs, j)
					}
				}
				return jobs
			},
			wantModifiedChangesetJobs: func(changesetJobs []*a8n.ChangesetJob, newCampaignJobs []*a8n.CampaignJob) (jobs []*a8n.ChangesetJob) {
				for _, j := range changesetJobs {
					// CampaignJob has new diff, so ChangesetJob should be reset
					if j.CampaignJobID == newCampaignJobs[1].ID {
						jobs = append(jobs, j)
					}
				}
				return jobs
			},
			wantCreatedChangesetJobs: func(changesetJobs []*a8n.ChangesetJob, newCampaignJobs []*a8n.CampaignJob) (jobs []*a8n.ChangesetJob) {
				for _, j := range changesetJobs {
					// CampaignJob has no old counterpart, so new ChangesetJob should be created
					if j.CampaignJobID == newCampaignJobs[2].ID {
						jobs = append(jobs, j)
					}
				}
				return jobs
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup old CampaignPlan and CampaignJobs
			oldPlan := &a8n.CampaignPlan{CampaignType: "test", Arguments: `{}`}
			err = store.CreateCampaignPlan(ctx, oldPlan)
			if err != nil {
				t.Fatal(err)
			}

			oldCampaignJobs := tt.oldCampaignJobs(oldPlan.ID)
			oldCampaignJobsByID := make(map[int64]*a8n.CampaignJob)
			for _, j := range oldCampaignJobs {
				err := store.CreateCampaignJob(ctx, j)
				if err != nil {
					t.Fatal(err)
				}
				oldCampaignJobsByID[j.ID] = j
			}
			campaign := testCampaign(u.ID, oldPlan.ID)

			// Create Campaign
			svc := NewServiceWithClock(store, gitClient, nil, cf, clock)
			err = svc.CreateCampaign(ctx, campaign, false)
			if err != nil {
				t.Fatal(err)
			}

			// Create Changesets and update ChangesetJobs to look like they ran
			oldChangesets := fakeRunChangesetJobs(ctx, t, store, now, campaign, oldCampaignJobsByID)

			oldTime := now
			now = now.Add(5 * time.Second)

			// Create a new CampaignPlan
			newPlan := &a8n.CampaignPlan{CampaignType: "test", Arguments: `{}`}
			err = store.CreateCampaignPlan(ctx, newPlan)
			if err != nil {
				t.Fatal(err)
			}

			newCampaignJobs := tt.newCampaignJobs(newPlan.ID, oldCampaignJobs)

			// Create new CampaignJobs
			for _, j := range newCampaignJobs {
				err := store.CreateCampaignJob(ctx, j)
				if err != nil {
					t.Fatal(err)
				}
			}

			// Update the Campaign to use the new plan
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

			wantUnmodifiedChangesetJobs := tt.wantUnmodifiedChangesetJobs(newChangesetJobs, newCampaignJobs)
			for _, j := range wantUnmodifiedChangesetJobs {
				if j.StartedAt != oldTime {
					t.Fatalf("ChangesetJob StartedAt changed. want=%v, have=%v", oldTime, j.StartedAt)
				}
				if j.FinishedAt != oldTime {
					t.Fatalf("ChangesetJob FinishedAt changed. want=%v, have=%v", oldTime, j.FinishedAt)
				}
				if j.ChangesetID == 0 {
					t.Fatalf("ChangesetJob does not have ChangesetID")
				}
			}

			wantModifiedChangesetJobs := tt.wantModifiedChangesetJobs(newChangesetJobs, newCampaignJobs)
			for _, j := range wantModifiedChangesetJobs {
				if !j.StartedAt.IsZero() {
					t.Fatalf("ChangesetJob StartedAt not reset. have=%v", j.StartedAt)
				}
				if !j.FinishedAt.IsZero() {
					t.Fatalf("ChangesetJob FinishedAt not reset. have=%v", j.FinishedAt)
				}
				if j.ChangesetID == 0 {
					t.Fatalf("ChangesetJob does not have ChangesetID")
				}
			}

			wantCreatedChangesetJobs := tt.wantCreatedChangesetJobs(newChangesetJobs, newCampaignJobs)
			for _, j := range wantCreatedChangesetJobs {
				if !j.StartedAt.IsZero() {
					t.Fatalf("ChangesetJob StartedAt is set. have=%v", j.StartedAt)
				}

				if !j.FinishedAt.IsZero() {
					t.Fatalf("ChangesetJob FinishedAt is set. have=%v", j.FinishedAt)
				}
				if j.ChangesetID != 0 {
					t.Fatalf("ChangesetJob.ChangesetID is not 0")
				}
			}

			// Check that Changesets attached to the unmodified and modified
			// ChangesetJobs are still attached to Campaign.
			var wantAttachedChangesetIDs []int64
			for _, j := range append(wantUnmodifiedChangesetJobs, wantModifiedChangesetJobs...) {
				wantAttachedChangesetIDs = append(wantAttachedChangesetIDs, j.ChangesetID)
			}
			changesets, _, err := store.ListChangesets(ctx, ListChangesetsOpts{IDs: wantAttachedChangesetIDs})
			if err != nil {
				t.Fatal(err)
			}
			if have, want := len(changesets), len(wantAttachedChangesetIDs); have != want {
				t.Fatalf("wrong number of changesets. want=%d, have=%d", have, want)
			}
			for _, c := range changesets {
				if len(c.CampaignIDs) != 1 || c.CampaignIDs[0] != campaign.ID {
					t.Fatalf("changeset has wrong CampaignIDs. want=[%d], have=%v", campaign.ID, c.CampaignIDs)
				}
			}

			// Check that Changesets with RepoID == campaignJobWithoutChangesetJob.RepoID
			// are detached from Campaign.
			detachedCampaignJobs := tt.wantCampaignJobsWithoutChangesetJob(oldCampaignJobs)
			if len(detachedCampaignJobs) == 0 {
				return
			}

			wantDetachedChangesetIDs := make([]int64, 0, len(detachedCampaignJobs))
			for _, c := range oldChangesets {
				for _, job := range detachedCampaignJobs {
					if c.RepoID == job.RepoID {
						wantDetachedChangesetIDs = append(wantDetachedChangesetIDs, c.ID)
					}
				}
			}
			if len(wantDetachedChangesetIDs) != len(detachedCampaignJobs) {
				t.Fatalf("could not find old changeset to be detached")
			}

			changesets, _, err = store.ListChangesets(ctx, ListChangesetsOpts{IDs: wantDetachedChangesetIDs})
			if err != nil {
				t.Fatal(err)
			}
			for _, c := range changesets {
				if len(c.CampaignIDs) != 0 {
					t.Fatalf("old changeset still attached to campaign")
				}
				for _, changesetID := range campaign.ChangesetIDs {
					if changesetID == c.ID {
						t.Fatalf("old changeset still attached to campaign")
					}
				}
			}
		})
	}
}

// fakeRunChangesetJobs does what (&Service).RunChangesetJobs does on a
// database level, but doesn't talk to the codehost. It creates fake Changesets
// for the ChangesetJobs associated with the given Campaign and updates the
// ChangesetJobs so they appear to have run.
func fakeRunChangesetJobs(
	ctx context.Context,
	t *testing.T,
	store *Store,
	now time.Time,
	campaign *a8n.Campaign,
	campaignJobsByID map[int64]*a8n.CampaignJob,
) []*a8n.Changeset {
	jobs, _, err := store.ListChangesetJobs(ctx, ListChangesetJobsOpts{
		CampaignID: campaign.ID,
		Limit:      -1,
	})
	if err != nil {
		t.Fatal(err)
	}

	if have, want := len(jobs), len(campaignJobsByID); have != want {
		t.Fatalf("wrong number of changeset jobs. want=%d, have=%d", want, have)
	}

	cs := make([]*a8n.Changeset, 0, len(campaignJobsByID))
	for _, changesetJob := range jobs {
		campaignJob, ok := campaignJobsByID[changesetJob.CampaignJobID]
		if !ok {
			t.Fatal("no CampaignJob found for ChangesetJob")
		}

		changeset := testChangeset(campaignJob.RepoID, changesetJob.CampaignID, changesetJob.ID)
		err = store.CreateChangesets(ctx, changeset)
		if err != nil {
			t.Fatal(err)
		}

		cs = append(cs, changeset)

		changesetJob.ChangesetID = changeset.ID
		changesetJob.StartedAt = now
		changesetJob.FinishedAt = now

		err := store.UpdateChangesetJob(ctx, changesetJob)
		if err != nil {
			t.Fatal(err)
		}
	}
	return cs
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

func testChangeset(repoID int32, campaign int64, changesetJob int64) *a8n.Changeset {
	return &a8n.Changeset{
		RepoID:              repoID,
		CampaignIDs:         []int64{campaign},
		ExternalServiceType: "github",
		ExternalID:          fmt.Sprintf("ext-id-%d", changesetJob),
	}
}

type dummyGitserverClient struct {
	response    string
	responseErr error
}

func (d *dummyGitserverClient) CreateCommitFromPatch(ctx context.Context, req protocol.CreateCommitFromPatchRequest) (string, error) {
	return d.response, d.responseErr
}
