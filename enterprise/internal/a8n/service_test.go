package a8n

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
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
	cf := httpcli.NewExternalHTTPClientFactory()

	user := createTestUser(ctx, t)

	store := NewStoreWithClock(dbconn.Global, clock)

	var rs []*repos.Repo
	for i := 0; i < 4; i++ {
		rs = append(rs, testRepo(i, github.ServiceType))
	}

	reposStore := repos.NewDBStore(dbconn.Global, sql.TxOptions{})
	err := reposStore.UpsertRepos(ctx, rs...)
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

		plan, err := svc.CreateCampaignPlanFromPatches(ctx, patches, user.ID)
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
		plan := &a8n.CampaignPlan{CampaignType: "test", Arguments: `{}`, UserID: user.ID}
		err = store.CreateCampaignPlan(ctx, plan)
		if err != nil {
			t.Fatal(err)
		}

		campaign := testCampaign(user.ID, plan.ID)
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
		plan := &a8n.CampaignPlan{CampaignType: "test", Arguments: `{}`, UserID: user.ID}
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

		campaign := testCampaign(user.ID, plan.ID)

		svc := NewServiceWithClock(store, gitClient, nil, cf, clock)
		err = svc.CreateCampaign(ctx, campaign, true)
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

		if len(haveJobs) != 0 {
			t.Errorf("wrong number of ChangesetJobs: %d. want=%d", len(haveJobs), 0)
		}
	})

	t.Run("CreateChangesetJobForCampaignJob", func(t *testing.T) {
		plan := &a8n.CampaignPlan{CampaignType: "test", Arguments: `{}`, UserID: user.ID}
		err = store.CreateCampaignPlan(ctx, plan)
		if err != nil {
			t.Fatal(err)
		}

		campaignJob := testCampaignJob(plan.ID, rs[0].ID, now)
		err := store.CreateCampaignJob(ctx, campaignJob)
		if err != nil {
			t.Fatal(err)
		}

		campaign := testCampaign(user.ID, plan.ID)
		err = store.CreateCampaign(ctx, campaign)
		if err != nil {
			t.Fatal(err)
		}

		svc := NewServiceWithClock(store, gitClient, nil, cf, clock)
		err = svc.CreateChangesetJobForCampaignJob(ctx, campaignJob.ID)
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

		// Try to create again, check that it's the same one
		err = svc.CreateChangesetJobForCampaignJob(ctx, campaignJob.ID)
		if err != nil {
			t.Fatal(err)
		}
		haveJob2, err := store.GetChangesetJob(ctx, GetChangesetJobOpts{
			CampaignID:    campaign.ID,
			CampaignJobID: campaignJob.ID,
		})
		if err != nil {
			t.Fatal(err)
		}

		if haveJob2.ID != haveJob.ID {
			t.Errorf("wrong changesetJob: %d. want=%d", haveJob2.ID, haveJob.ID)
		}
	})

	t.Run("UpdateCampaignWithUnprocessedChangesetJobs", func(t *testing.T) {
		subTests := []struct {
			name  string
			draft bool
			err   string
		}{
			{
				name:  "published campaign",
				draft: false,
				err:   ErrUpdateProcessingCampaign.Error(),
			},
			{
				name:  "draft campaign",
				draft: true,
			},
		}
		for _, tc := range subTests {
			t.Run(tc.name, func(t *testing.T) {
				if tc.err == "" {
					tc.err = "<nil>"
				}

				plan := &a8n.CampaignPlan{CampaignType: "test", Arguments: `{}`, UserID: user.ID}
				err = store.CreateCampaignPlan(ctx, plan)
				if err != nil {
					t.Fatal(err)
				}

				campaignJob := testCampaignJob(plan.ID, rs[0].ID, now)
				err := store.CreateCampaignJob(ctx, campaignJob)
				if err != nil {
					t.Fatal(err)
				}

				svc := NewServiceWithClock(store, gitClient, nil, cf, clock)
				campaign := testCampaign(user.ID, plan.ID)

				err = svc.CreateCampaign(ctx, campaign, tc.draft)
				if err != nil {
					t.Fatal(err)
				}

				if !tc.draft {
					haveJobs, _, err := store.ListChangesetJobs(ctx, ListChangesetJobsOpts{
						CampaignID: campaign.ID,
					})
					if err != nil {
						t.Fatal(err)
					}

					// sanity checks
					if len(haveJobs) != 1 {
						t.Errorf("wrong number of ChangesetJobs: %d. want=%d", len(haveJobs), 1)
					}

					if !haveJobs[0].StartedAt.IsZero() {
						t.Errorf("ChangesetJobs is not unprocessed. StartedAt=%v", haveJobs[0].StartedAt)
					}
				}

				newName := "this is a new campaign name"
				args := UpdateCampaignArgs{Campaign: campaign.ID, Name: &newName}

				updatedCampaign, _, err := svc.UpdateCampaign(ctx, args)
				if have, want := fmt.Sprint(err), tc.err; have != want {
					t.Errorf("error:\nhave: %q\nwant: %q", have, want)
				}

				if tc.err != "<nil>" {
					return
				}

				if updatedCampaign.Name != newName {
					t.Errorf("Name not updated. want=%q, have=%q", newName, updatedCampaign.Name)
				}
			})
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
	cf := httpcli.NewExternalHTTPClientFactory()

	user := createTestUser(ctx, t)

	var rs []*repos.Repo
	for i := 0; i < 4; i++ {
		rs = append(rs, testRepo(i, github.ServiceType))
	}

	reposStore := repos.NewDBStore(dbconn.Global, sql.TxOptions{})
	err := reposStore.UpsertRepos(ctx, rs...)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name                                string
		draft                               bool
		manualCampaign                      bool
		args                                func(campaignID, newPlanID int64) UpdateCampaignArgs
		oldCampaignJobs                     func(currentPlanID int64) []*a8n.CampaignJob
		newCampaignJobs                     func(newPlanID int64, oldCampaignJobs []*a8n.CampaignJob) []*a8n.CampaignJob
		wantCampaignJobsWithoutChangesetJob func(oldCampaignJobs []*a8n.CampaignJob) []*a8n.CampaignJob
		wantUnmodifiedChangesetJobs         func(changesetJobs []*a8n.ChangesetJob, newCampaignJobs []*a8n.CampaignJob) []*a8n.ChangesetJob
		wantModifiedChangesetJobs           func(changesetJobs []*a8n.ChangesetJob, newCampaignJobs []*a8n.CampaignJob) (jobs []*a8n.ChangesetJob)
		wantCreatedChangesetJobs            func(changesetJobs []*a8n.ChangesetJob, newCampaignJobs []*a8n.CampaignJob) (jobs []*a8n.ChangesetJob)
	}{
		{
			name:           "manual campaign, no new plan, name update",
			manualCampaign: true,
			args: func(campaignID, planID int64) UpdateCampaignArgs {
				newName := "this is a new name"
				return UpdateCampaignArgs{Campaign: campaignID, Name: &newName}
			},
			oldCampaignJobs: func(plan int64) []*a8n.CampaignJob {
				return []*a8n.CampaignJob{}
			},
			newCampaignJobs: func(plan int64, oldCampaignJobs []*a8n.CampaignJob) []*a8n.CampaignJob {
				return []*a8n.CampaignJob{}
			},
		},
		{
			name: "1 unmodified",
			args: func(campaignID, planID int64) UpdateCampaignArgs {
				return UpdateCampaignArgs{Campaign: campaignID, Plan: &planID}
			},
			oldCampaignJobs: func(plan int64) []*a8n.CampaignJob {
				return []*a8n.CampaignJob{testCampaignJob(plan, rs[0].ID, now)}
			},
			newCampaignJobs: func(plan int64, oldCampaignJobs []*a8n.CampaignJob) []*a8n.CampaignJob {
				job := oldCampaignJobs[0].Clone()
				job.CampaignPlanID = plan
				return []*a8n.CampaignJob{job}
			},
			wantUnmodifiedChangesetJobs: func(changesetJobs []*a8n.ChangesetJob, newCampaignJobs []*a8n.CampaignJob) (jobs []*a8n.ChangesetJob) {
				// We only have 1 ChangesetJob and that should be unmodified
				return changesetJobs
			},
		},
		{
			name: "1 unmodified, no new plan but name update",
			args: func(campaignID, planID int64) UpdateCampaignArgs {
				newName := "this is a new name"
				return UpdateCampaignArgs{Campaign: campaignID, Name: &newName}
			},
			oldCampaignJobs: func(plan int64) []*a8n.CampaignJob {
				return []*a8n.CampaignJob{testCampaignJob(plan, rs[0].ID, now)}
			},
			newCampaignJobs: func(plan int64, oldCampaignJobs []*a8n.CampaignJob) []*a8n.CampaignJob {
				job := oldCampaignJobs[0].Clone()
				job.CampaignPlanID = plan
				return []*a8n.CampaignJob{job}
			},
			wantModifiedChangesetJobs: func(changesetJobs []*a8n.ChangesetJob, newCampaignJobs []*a8n.CampaignJob) (jobs []*a8n.ChangesetJob) {
				// We only have 1 ChangesetJob and that should be modified
				return changesetJobs
			},
		},
		{
			name: "1 unmodified, no new plan but description update",
			args: func(campaignID, planID int64) UpdateCampaignArgs {
				newDescription := "this is a new description"
				return UpdateCampaignArgs{Campaign: campaignID, Description: &newDescription}
			},
			oldCampaignJobs: func(plan int64) []*a8n.CampaignJob {
				return []*a8n.CampaignJob{testCampaignJob(plan, rs[0].ID, now)}
			},
			newCampaignJobs: func(plan int64, oldCampaignJobs []*a8n.CampaignJob) []*a8n.CampaignJob {
				job := oldCampaignJobs[0].Clone()
				job.CampaignPlanID = plan
				return []*a8n.CampaignJob{job}
			},
			wantModifiedChangesetJobs: func(changesetJobs []*a8n.ChangesetJob, newCampaignJobs []*a8n.CampaignJob) (jobs []*a8n.ChangesetJob) {
				// We only have 1 ChangesetJob and that should be modified
				return changesetJobs
			},
		},
		{
			name: "1 modified diff",
			args: func(campaignID, planID int64) UpdateCampaignArgs {
				return UpdateCampaignArgs{Campaign: campaignID, Plan: &planID}
			},
			oldCampaignJobs: func(plan int64) []*a8n.CampaignJob {
				return []*a8n.CampaignJob{testCampaignJob(plan, rs[0].ID, now)}
			},
			newCampaignJobs: func(plan int64, oldCampaignJobs []*a8n.CampaignJob) []*a8n.CampaignJob {
				job := oldCampaignJobs[0].Clone()
				job.CampaignPlanID = plan
				job.Diff = "different diff"
				return []*a8n.CampaignJob{job}
			},
			wantModifiedChangesetJobs: func(changesetJobs []*a8n.ChangesetJob, newCampaignJobs []*a8n.CampaignJob) (jobs []*a8n.ChangesetJob) {
				// We only have 1 ChangesetJob and that should be modified
				return changesetJobs
			},
		},
		{
			name: "1 modified rev",
			args: func(campaignID, planID int64) UpdateCampaignArgs {
				return UpdateCampaignArgs{Campaign: campaignID, Plan: &planID}
			},
			oldCampaignJobs: func(plan int64) []*a8n.CampaignJob {
				return []*a8n.CampaignJob{testCampaignJob(plan, rs[0].ID, now)}
			},
			newCampaignJobs: func(plan int64, oldCampaignJobs []*a8n.CampaignJob) []*a8n.CampaignJob {
				job := oldCampaignJobs[0].Clone()
				job.CampaignPlanID = plan
				job.Rev = "deadbeef23"
				return []*a8n.CampaignJob{job}
			},
			wantModifiedChangesetJobs: func(changesetJobs []*a8n.ChangesetJob, newCampaignJobs []*a8n.CampaignJob) (jobs []*a8n.ChangesetJob) {
				// We only have 1 ChangesetJob and that should be modified
				return changesetJobs
			},
		},
		{
			name: "1 unmodified, 1 modified, 1 new changeset",
			args: func(campaignID, planID int64) UpdateCampaignArgs {
				return UpdateCampaignArgs{Campaign: campaignID, Plan: &planID}
			},
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
		{
			name:  "draft campaign, 1 unmodified, 1 modified, 1 new changeset",
			draft: true,
			args: func(campaignID, planID int64) UpdateCampaignArgs {
				return UpdateCampaignArgs{Campaign: campaignID, Plan: &planID}
			},
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
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewStoreWithClock(dbconn.Global, clock)
			svc := NewServiceWithClock(store, gitClient, nil, cf, clock)

			var (
				campaign            *a8n.Campaign
				oldCampaignJobs     []*a8n.CampaignJob
				oldCampaignJobsByID map[int64]*a8n.CampaignJob
			)

			if tt.manualCampaign {
				campaign = testCampaign(user.ID, 0)
			} else {
				plan := &a8n.CampaignPlan{CampaignType: "comby", Arguments: `{}`, UserID: user.ID}
				err = store.CreateCampaignPlan(ctx, plan)
				if err != nil {
					t.Fatal(err)
				}

				oldCampaignJobs = tt.oldCampaignJobs(plan.ID)
				oldCampaignJobsByID = make(map[int64]*a8n.CampaignJob)
				for _, j := range oldCampaignJobs {
					err := store.CreateCampaignJob(ctx, j)
					if err != nil {
						t.Fatal(err)
					}
					oldCampaignJobsByID[j.ID] = j
				}
				campaign = testCampaign(user.ID, plan.ID)
			}

			err = svc.CreateCampaign(ctx, campaign, tt.draft)
			if err != nil {
				t.Fatal(err)
			}

			var oldChangesets []*a8n.Changeset
			if !tt.draft && !tt.manualCampaign {
				// Create Changesets and update ChangesetJobs to look like they ran
				oldChangesets = fakeRunChangesetJobs(ctx, t, store, now, campaign, oldCampaignJobsByID)
			}

			oldTime := now
			now = now.Add(5 * time.Second)

			newPlan := &a8n.CampaignPlan{CampaignType: "comby", Arguments: `{}`, UserID: user.ID}
			err = store.CreateCampaignPlan(ctx, newPlan)
			if err != nil {
				t.Fatal(err)
			}

			newCampaignJobs := tt.newCampaignJobs(newPlan.ID, oldCampaignJobs)

			for _, j := range newCampaignJobs {
				err := store.CreateCampaignJob(ctx, j)
				if err != nil {
					t.Fatal(err)
				}
			}

			// Update the Campaign
			args := tt.args(campaign.ID, newPlan.ID)
			updatedCampaign, detachedChangesets, err := svc.UpdateCampaign(ctx, args)
			if err != nil {
				t.Fatal(err)
			}

			if args.Name != nil && updatedCampaign.Name != *args.Name {
				t.Fatalf("campaign name not updated. want=%q, have=%q", *args.Name, updatedCampaign.Name)
			}
			if args.Description != nil && updatedCampaign.Description != *args.Description {
				t.Fatalf("campaign description not updated. want=%q, have=%q", *args.Description, updatedCampaign.Description)
			}

			if args.Plan != nil && updatedCampaign.CampaignPlanID != *args.Plan {
				t.Fatalf("campaign CampaignPlanID not updated. want=%q, have=%q", newPlan.ID, updatedCampaign.CampaignPlanID)
			}

			newChangesetJobs, _, err := store.ListChangesetJobs(ctx, ListChangesetJobsOpts{
				CampaignID: campaign.ID,
				Limit:      -1,
			})
			if err != nil {
				t.Fatal(err)
			}

			// When a campaign is created as a draft, we don't create
			// ChangesetJobs, which means we can return here after checking
			// that we haven't created ChangesetJobs
			if tt.draft {
				if len(newChangesetJobs) != 0 {
					t.Fatalf("changesetJobs created even though campaign is draft. have=%d", len(newChangesetJobs))
				}
				return
			}

			if len(newChangesetJobs) != len(newCampaignJobs) {
				t.Fatalf("wrong number of new ChangesetJobs. want=%d, have=%d", len(newCampaignJobs), len(newChangesetJobs))
			}

			var wantUnmodifiedChangesetJobs []*a8n.ChangesetJob
			if tt.wantUnmodifiedChangesetJobs != nil {
				wantUnmodifiedChangesetJobs = tt.wantUnmodifiedChangesetJobs(newChangesetJobs, newCampaignJobs)
			}
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

			var wantModifiedChangesetJobs []*a8n.ChangesetJob
			if tt.wantModifiedChangesetJobs != nil {
				wantModifiedChangesetJobs = tt.wantModifiedChangesetJobs(newChangesetJobs, newCampaignJobs)
			}
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

			var wantCreatedChangesetJobs []*a8n.ChangesetJob
			if tt.wantCreatedChangesetJobs != nil {
				wantCreatedChangesetJobs = tt.wantCreatedChangesetJobs(newChangesetJobs, newCampaignJobs)
			}
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
			var detachedCampaignJobs []*a8n.CampaignJob
			if tt.wantCampaignJobsWithoutChangesetJob != nil {
				detachedCampaignJobs = tt.wantCampaignJobsWithoutChangesetJob(oldCampaignJobs)
			}
			if len(detachedCampaignJobs) == 0 {
				return
			}

			wantIDs := make([]int64, 0, len(detachedCampaignJobs))
			for _, c := range oldChangesets {
				for _, job := range detachedCampaignJobs {
					if c.RepoID == job.RepoID {
						wantIDs = append(wantIDs, c.ID)
					}
				}
			}
			if len(wantIDs) != len(detachedCampaignJobs) {
				t.Fatalf("could not find old changeset to be detached")
			}

			haveIDs := make([]int64, 0, len(detachedChangesets))
			for _, c := range detachedChangesets {
				if len(c.CampaignIDs) != 0 {
					t.Fatalf("old changeset still attached to campaign")
				}
				for _, changesetID := range campaign.ChangesetIDs {
					if changesetID == c.ID {
						t.Fatalf("old changeset still attached to campaign")
					}
				}
				haveIDs = append(haveIDs, c.ID)
			}
			sort.Slice(wantIDs, func(i, j int) bool { return wantIDs[i] < wantIDs[j] })
			sort.Slice(haveIDs, func(i, j int) bool { return haveIDs[i] < haveIDs[j] })

			if diff := cmp.Diff(haveIDs, wantIDs); diff != "" {
				t.Fatal(diff)
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

var testUser = db.NewUser{
	Email:                 "thorsten@sourcegraph.com",
	Username:              "thorsten",
	DisplayName:           "thorsten",
	Password:              "1234",
	EmailVerificationCode: "foobar",
}

func createTestUser(ctx context.Context, t *testing.T) *types.User {
	t.Helper()
	user, err := db.Users.Create(ctx, testUser)
	if err != nil {
		t.Fatal(err)
	}
	return user
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
