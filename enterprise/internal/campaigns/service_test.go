package campaigns

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
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

func init() {
	dbtesting.DBNameSuffix = "campaignsenterpriserdb"
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
		svc := NewServiceWithClock(store, nil, nil, clock)

		const patch = `diff f f
--- f
+++ f
@@ -1,1 +1,2 @@
+x
 y
`
		patches := []campaigns.CampaignPlanPatch{
			{Repo: api.RepoID(rs[0].ID), BaseRevision: "deadbeef", BaseRef: "refs/heads/master", Patch: patch},
			{Repo: api.RepoID(rs[1].ID), BaseRevision: "f00b4r", BaseRef: "refs/heads/master", Patch: patch},
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
		wantJobs := make([]*campaigns.CampaignJob, len(patches))
		for i, patch := range patches {
			wantJobs[i] = &campaigns.CampaignJob{
				CampaignPlanID: plan.ID,
				RepoID:         patch.Repo,
				Rev:            patch.BaseRevision,
				BaseRef:        patch.BaseRef,
				Diff:           patch.Patch,
				CreatedAt:      now,
				UpdatedAt:      now,
			}
		}
		if !cmp.Equal(jobs, wantJobs) {
			t.Error("jobs != wantJobs", cmp.Diff(jobs, wantJobs))
		}
	})

	t.Run("CreateCampaign", func(t *testing.T) {
		plan := &campaigns.CampaignPlan{UserID: user.ID}
		err = store.CreateCampaignPlan(ctx, plan)
		if err != nil {
			t.Fatal(err)
		}

		campaign := testCampaign(user.ID, plan.ID)
		svc := NewServiceWithClock(store, gitClient, cf, clock)

		// Without CampaignJobs it should fail
		err = svc.CreateCampaign(ctx, campaign, false)
		if err != ErrNoCampaignJobs {
			t.Fatal("CreateCampaign did not produce expected error")
		}

		campaignJobs := make([]*campaigns.CampaignJob, 0, len(rs))
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
		plan := &campaigns.CampaignPlan{UserID: user.ID}
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

		svc := NewServiceWithClock(store, gitClient, cf, clock)
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

	t.Run("CreateCampaignWithPlanAttachedToOtherCampaign", func(t *testing.T) {
		plan := &campaigns.CampaignPlan{UserID: user.ID}
		err = store.CreateCampaignPlan(ctx, plan)
		if err != nil {
			t.Fatal(err)
		}

		for _, repo := range rs {
			err := store.CreateCampaignJob(ctx, testCampaignJob(plan.ID, repo.ID, now))
			if err != nil {
				t.Fatal(err)
			}
		}

		campaign := testCampaign(user.ID, plan.ID)
		svc := NewServiceWithClock(store, gitClient, cf, clock)

		err = svc.CreateCampaign(ctx, campaign, false)
		if err != nil {
			t.Fatal(err)
		}

		otherCampaign := testCampaign(user.ID, plan.ID)
		err = svc.CreateCampaign(ctx, otherCampaign, false)
		if err != ErrCampaignPlanDuplicate {
			t.Fatal("no error even though another campaign has same plan")
		}
	})

	t.Run("CreateChangesetJobForCampaignJob", func(t *testing.T) {
		plan := &campaigns.CampaignPlan{UserID: user.ID}
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

		svc := NewServiceWithClock(store, gitClient, cf, clock)
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

	t.Run("UpdateCampaign", func(t *testing.T) {
		strPointer := func(s string) *string { return &s }
		subTests := []struct {
			name    string
			branch  *string
			draft   bool
			process bool
			err     string
		}{
			{
				name:  "published unprocessed campaign",
				draft: false,
				err:   ErrUpdateProcessingCampaign.Error(),
			},
			{
				name:  "draft campaign",
				draft: true,
			},
			{
				name:    "change campaign branch",
				branch:  strPointer("changed-branch"),
				draft:   true,
				process: true,
			},
			{
				name:    "change published campaign branch",
				branch:  strPointer("changed-branch"),
				process: true,
				err:     ErrPublishedCampaignBranchChange.Error(),
			},
			{
				name:    "change campaign blank branch",
				branch:  strPointer(""),
				draft:   true,
				process: true,
				err:     ErrCampaignBranchBlank.Error(),
			},
		}
		for _, tc := range subTests {
			t.Run(tc.name, func(t *testing.T) {
				if tc.err == "" {
					tc.err = "<nil>"
				}

				plan := &campaigns.CampaignPlan{UserID: user.ID}
				err = store.CreateCampaignPlan(ctx, plan)
				if err != nil {
					t.Fatal(err)
				}

				campaignJob := testCampaignJob(plan.ID, rs[0].ID, now)
				err := store.CreateCampaignJob(ctx, campaignJob)
				if err != nil {
					t.Fatal(err)
				}

				svc := NewServiceWithClock(store, gitClient, cf, clock)
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

					if tc.process {
						campaignJobsByID := map[int64]*campaigns.CampaignJob{
							campaignJob.ID: campaignJob,
						}
						states := map[int64]campaigns.ChangesetState{
							campaignJob.ID: campaigns.ChangesetStateOpen,
						}
						fakeRunChangesetJobs(ctx, t, store, now, campaign, campaignJobsByID, states)
					}
				}

				newName := "this is a new campaign name"
				args := UpdateCampaignArgs{Campaign: campaign.ID, Name: &newName, Branch: tc.branch}

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

				if tc.branch != nil && updatedCampaign.Branch != *tc.branch {
					t.Errorf("Branch not updated. want=%q, have %q", updatedCampaign.Branch, *tc.branch)
				}
			})
		}
	})

	t.Run("UpdateCampaignWithPlanAttachedToOtherCampaign", func(t *testing.T) {
		svc := NewServiceWithClock(store, gitClient, cf, clock)

		plan := &campaigns.CampaignPlan{UserID: user.ID}
		err = store.CreateCampaignPlan(ctx, plan)
		if err != nil {
			t.Fatal(err)
		}

		for _, repo := range rs {
			err := store.CreateCampaignJob(ctx, testCampaignJob(plan.ID, repo.ID, now))
			if err != nil {
				t.Fatal(err)
			}
		}

		campaign := testCampaign(user.ID, plan.ID)
		err = svc.CreateCampaign(ctx, campaign, false)
		if err != nil {
			t.Fatal(err)
		}

		otherPlan := &campaigns.CampaignPlan{UserID: user.ID}
		err = store.CreateCampaignPlan(ctx, otherPlan)
		if err != nil {
			t.Fatal(err)
		}

		for _, repo := range rs {
			err := store.CreateCampaignJob(ctx, testCampaignJob(otherPlan.ID, repo.ID, now))
			if err != nil {
				t.Fatal(err)
			}
		}
		otherCampaign := testCampaign(user.ID, otherPlan.ID)
		err = svc.CreateCampaign(ctx, otherCampaign, false)
		if err != nil {
			t.Fatal(err)
		}

		args := UpdateCampaignArgs{Campaign: otherCampaign.ID, Plan: &plan.ID}
		_, _, err := svc.UpdateCampaign(ctx, args)
		if err != ErrCampaignPlanDuplicate {
			t.Fatal("no error even though another campaign has same plan")
		}
	})

}

type repoNames []string

type newCampaignJobSpec struct {
	repo string

	modifiedDiff bool
	modifiedRev  bool
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

	reposByID := make(map[api.RepoID]*repos.Repo, len(rs))
	reposByName := make(map[string]*repos.Repo, len(rs))
	for _, r := range rs {
		reposByID[r.ID] = r
		reposByName[r.Name] = r
	}

	tests := []struct {
		name string

		campaignIsDraft  bool
		campaignIsManual bool

		// Repositories for which we had CampaignJobs attached to the old CampaignPlan
		oldCampaignJobs repoNames

		// Repositories for which the ChangesetJob/Changeset have been
		// individually published while Campaign was in draft mode
		individuallyPublished repoNames

		// Mapping of repository names to state of changesets after creating the campaign.
		// Default state is ChangesetStateOpen
		changesetStates map[string]campaigns.ChangesetState

		updatePlan, updateName, updateDescription bool
		newCampaignJobs                           []newCampaignJobSpec

		// Repositories for which we want no Changeset/ChangesetJob after update
		wantDetached repoNames
		// Repositories for which we want to keep Changeset/ChangesetJob unmodified
		wantUnmodified repoNames
		// Repositories for which we want to keep Changeset/ChangesetJob and update them
		wantModified repoNames
		// Repositories for which we want to create a new ChangesetJob (and thus a Changeset)
		wantCreated repoNames
	}{
		{
			name:             "manual campaign, no new plan, name update",
			campaignIsManual: true,
			updateName:       true,
		},
		{
			name:            "1 unmodified",
			updatePlan:      true,
			oldCampaignJobs: repoNames{"repo-0"},
			newCampaignJobs: []newCampaignJobSpec{
				{repo: "repo-0"},
			},
			wantUnmodified: repoNames{"repo-0"},
		},
		{
			name:            "no new plan but name update",
			updateName:      true,
			oldCampaignJobs: repoNames{"repo-0"},
			wantModified:    repoNames{"repo-0"},
		},
		{
			name:              "no new plan but description update",
			updateDescription: true,
			oldCampaignJobs:   repoNames{"repo-0"},
			wantModified:      repoNames{"repo-0"},
		},
		{
			name:            "1 modified diff",
			updatePlan:      true,
			oldCampaignJobs: repoNames{"repo-0"},
			newCampaignJobs: []newCampaignJobSpec{
				{repo: "repo-0", modifiedDiff: true},
			},
			wantModified: repoNames{"repo-0"},
		},
		{
			name:            "1 modified rev",
			updatePlan:      true,
			oldCampaignJobs: repoNames{"repo-0"},
			newCampaignJobs: []newCampaignJobSpec{
				{repo: "repo-0", modifiedRev: true},
			},
			wantModified: repoNames{"repo-0"},
		},
		{
			name:            "1 detached, 1 unmodified, 1 modified, 1 new changeset",
			updatePlan:      true,
			oldCampaignJobs: repoNames{"repo-0", "repo-1", "repo-2"},
			newCampaignJobs: []newCampaignJobSpec{
				{repo: "repo-0"},
				{repo: "repo-1", modifiedDiff: true},
				{repo: "repo-3"},
			},
			wantDetached:   repoNames{"repo-2"},
			wantUnmodified: repoNames{"repo-0"},
			wantModified:   repoNames{"repo-1"},
			wantCreated:    repoNames{"repo-3"},
		},
		{
			name:            "draft campaign, 1 unmodified, 1 modified, 1 new changeset",
			campaignIsDraft: true,
			updatePlan:      true,
			oldCampaignJobs: repoNames{"repo-0", "repo-1", "repo-2"},
			newCampaignJobs: []newCampaignJobSpec{
				{repo: "repo-0"},
				{repo: "repo-1", modifiedDiff: true},
				{repo: "repo-3"},
			},
		},
		{
			name:                  "draft campaign, 1 published unmodified, 1 modified, 1 detached, 1 new changeset",
			campaignIsDraft:       true,
			updatePlan:            true,
			oldCampaignJobs:       repoNames{"repo-0", "repo-1", "repo-2"},
			individuallyPublished: repoNames{"repo-0"},
			newCampaignJobs: []newCampaignJobSpec{
				{repo: "repo-0"},
				{repo: "repo-1", modifiedDiff: true},
				{repo: "repo-3"},
			},
			wantUnmodified: repoNames{"repo-0"},
		},
		{
			name:                  "draft campaign, 1 published unmodified, 1 published modified, 1 detached, 1 new changeset",
			campaignIsDraft:       true,
			updatePlan:            true,
			oldCampaignJobs:       repoNames{"repo-0", "repo-1", "repo-2"},
			individuallyPublished: repoNames{"repo-0", "repo-1"},
			newCampaignJobs: []newCampaignJobSpec{
				{repo: "repo-0"},
				{repo: "repo-1", modifiedDiff: true},
				{repo: "repo-3"},
			},
			wantUnmodified: repoNames{"repo-0"},
			wantModified:   repoNames{"repo-1"},
		},
		{
			name:                  "draft campaign, 1 published unmodified, 1 published modified, 1 published detached, 1 new changeset",
			campaignIsDraft:       true,
			updatePlan:            true,
			oldCampaignJobs:       repoNames{"repo-0", "repo-1", "repo-2"},
			individuallyPublished: repoNames{"repo-0", "repo-1", "repo-2"},
			newCampaignJobs: []newCampaignJobSpec{
				{repo: "repo-0"},
				{repo: "repo-1", modifiedDiff: true},
				{repo: "repo-3"},
			},
			wantUnmodified: repoNames{"repo-0"},
			wantModified:   repoNames{"repo-1"},
			wantDetached:   repoNames{"repo-2"},
		},
		{
			name:            "1 modified diff for already merged changeset",
			updatePlan:      true,
			oldCampaignJobs: repoNames{"repo-0"},
			changesetStates: map[string]campaigns.ChangesetState{"repo-0": campaigns.ChangesetStateMerged},
			newCampaignJobs: []newCampaignJobSpec{
				{repo: "repo-0", modifiedDiff: true},
			},
			wantUnmodified: repoNames{"repo-0"},
		},
		{
			name:            "1 modified rev for already merged changeset",
			updatePlan:      true,
			oldCampaignJobs: repoNames{"repo-0"},
			changesetStates: map[string]campaigns.ChangesetState{"repo-0": campaigns.ChangesetStateMerged},
			newCampaignJobs: []newCampaignJobSpec{
				{repo: "repo-0", modifiedDiff: true},
			},
			wantUnmodified: repoNames{"repo-0"},
		},
		{
			name:            "1 modified diff for already closed changeset",
			updatePlan:      true,
			oldCampaignJobs: repoNames{"repo-0"},
			changesetStates: map[string]campaigns.ChangesetState{"repo-0": campaigns.ChangesetStateClosed},
			newCampaignJobs: []newCampaignJobSpec{
				{repo: "repo-0", modifiedDiff: true},
			},
			wantUnmodified: repoNames{"repo-0"},
		},
		{
			name:            "1 modified rev for already closed changeset",
			updatePlan:      true,
			oldCampaignJobs: repoNames{"repo-0"},
			changesetStates: map[string]campaigns.ChangesetState{"repo-0": campaigns.ChangesetStateClosed},
			newCampaignJobs: []newCampaignJobSpec{
				{repo: "repo-0", modifiedDiff: true},
			},
			wantUnmodified: repoNames{"repo-0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewStoreWithClock(dbconn.Global, clock)
			svc := NewServiceWithClock(store, gitClient, cf, clock)

			var (
				campaign         *campaigns.Campaign
				oldCampaignJobs  []*campaigns.CampaignJob
				newCampaignJobs  []*campaigns.CampaignJob
				campaignJobsByID map[int64]*campaigns.CampaignJob

				changesetStateByCampaignJobID map[int64]campaigns.ChangesetState

				oldChangesets []*campaigns.Changeset
			)

			if tt.campaignIsManual {
				campaign = testCampaign(user.ID, 0)
			} else {
				plan := &campaigns.CampaignPlan{UserID: user.ID}
				err = store.CreateCampaignPlan(ctx, plan)
				if err != nil {
					t.Fatal(err)
				}

				campaignJobsByID = make(map[int64]*campaigns.CampaignJob)
				changesetStateByCampaignJobID = make(map[int64]campaigns.ChangesetState)
				for _, repoName := range tt.oldCampaignJobs {
					repo, ok := reposByName[repoName]
					if !ok {
						t.Fatalf("unrecognized repo name: %s", repoName)
					}

					j := testCampaignJob(plan.ID, repo.ID, now)
					err := store.CreateCampaignJob(ctx, j)
					if err != nil {
						t.Fatal(err)
					}
					campaignJobsByID[j.ID] = j
					oldCampaignJobs = append(oldCampaignJobs, j)

					if s, ok := tt.changesetStates[repoName]; ok {
						changesetStateByCampaignJobID[j.ID] = s
					} else {
						changesetStateByCampaignJobID[j.ID] = campaigns.ChangesetStateOpen
					}
				}
				campaign = testCampaign(user.ID, plan.ID)
			}

			err = svc.CreateCampaign(ctx, campaign, tt.campaignIsDraft)
			if err != nil {
				t.Fatal(err)
			}

			if !tt.campaignIsDraft && !tt.campaignIsManual {
				// Create Changesets and update ChangesetJobs to look like they ran
				oldChangesets = fakeRunChangesetJobs(ctx, t, store, now, campaign, campaignJobsByID, changesetStateByCampaignJobID)
			}

			if tt.campaignIsDraft && len(tt.individuallyPublished) != 0 {
				toPublish := make(map[int64]*campaigns.CampaignJob)
				for _, name := range tt.individuallyPublished {
					repo, ok := reposByName[name]
					if !ok {
						t.Errorf("unrecognized repo name: %s", name)
					}
					for _, j := range oldCampaignJobs {
						if j.RepoID == repo.ID {
							toPublish[j.ID] = j

							err = svc.CreateChangesetJobForCampaignJob(ctx, j.ID)
							if err != nil {
								t.Fatalf("Failed to individually created ChangesetJob: %s", err)
							}
						}
					}
				}

				oldChangesets = fakeRunChangesetJobs(ctx, t, store, now, campaign, toPublish, changesetStateByCampaignJobID)
			}

			oldTime := now
			now = now.Add(5 * time.Second)

			newPlan := &campaigns.CampaignPlan{UserID: user.ID}
			err = store.CreateCampaignPlan(ctx, newPlan)
			if err != nil {
				t.Fatal(err)
			}

			for _, spec := range tt.newCampaignJobs {
				r, ok := reposByName[spec.repo]
				if !ok {
					t.Fatalf("unrecognized repo name: %s", spec.repo)
				}

				j := testCampaignJob(newPlan.ID, r.ID, now)

				if spec.modifiedDiff {
					j.Diff = j.Diff + "-modified"
				}

				if spec.modifiedRev {
					j.Rev = j.Rev + "-modified"
				}

				err := store.CreateCampaignJob(ctx, j)
				if err != nil {
					t.Fatal(err)
				}

				newCampaignJobs = append(newCampaignJobs, j)
				campaignJobsByID[j.ID] = j
			}

			// Update the Campaign
			args := UpdateCampaignArgs{Campaign: campaign.ID}
			if tt.updateName {
				newName := "new campaign Name"
				args.Name = &newName
			}
			if tt.updateDescription {
				newDescription := "new campaign description"
				args.Description = &newDescription
			}
			if tt.updatePlan {
				args.Plan = &newPlan.ID
			}

			// We ignore the returned campaign here and load it from the
			// database again to make sure the changes are persisted
			_, detachedChangesets, err := svc.UpdateCampaign(ctx, args)
			if err != nil {
				t.Fatal(err)
			}

			updatedCampaign, err := store.GetCampaign(ctx, GetCampaignOpts{ID: campaign.ID})
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
			if tt.campaignIsDraft && len(tt.individuallyPublished) == 0 {
				if have, want := len(newChangesetJobs), len(tt.individuallyPublished); have != want {
					t.Fatalf("changesetJobs created even though campaign is draft. have=%d, want=%d", have, want)
				}
				return
			}

			var wantChangesetJobLen int
			if tt.updatePlan {
				if len(tt.individuallyPublished) != 0 {
					wantChangesetJobLen = len(tt.individuallyPublished)
				} else {
					wantChangesetJobLen = len(newCampaignJobs)
				}
			} else {
				wantChangesetJobLen = len(oldCampaignJobs)
			}
			if len(newChangesetJobs) != wantChangesetJobLen {
				t.Fatalf("wrong number of new ChangesetJobs. want=%d, have=%d", wantChangesetJobLen, len(newChangesetJobs))
			}

			newChangesetJobsByRepo := map[string]*campaigns.ChangesetJob{}
			for _, c := range newChangesetJobs {
				campaignJob, ok := campaignJobsByID[c.CampaignJobID]
				if !ok {
					t.Fatalf("ChangesetJob has invalid CampaignJobID: %+v", c)
				}
				r, ok := reposByID[campaignJob.RepoID]
				if !ok {
					t.Fatalf("ChangesetJob has invalid RepoID: %v", c)
				}
				if c.ChangesetID != 0 && c.Branch == "" {
					t.Fatalf("Finished ChangesetJob is missing branch")
				}
				newChangesetJobsByRepo[r.Name] = c
			}

			wantUnmodifiedChangesetJobs := findChangesetJobsByRepoName(t, newChangesetJobsByRepo, tt.wantUnmodified)
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

			wantModifiedChangesetJobs := findChangesetJobsByRepoName(t, newChangesetJobsByRepo, tt.wantModified)
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

			wantCreatedChangesetJobs := findChangesetJobsByRepoName(t, newChangesetJobsByRepo, tt.wantCreated)
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
				t.Fatalf("wrong number of changesets. want=%d, have=%d", want, have)
			}
			for _, c := range changesets {
				if len(c.CampaignIDs) != 1 || c.CampaignIDs[0] != campaign.ID {
					t.Fatalf("changeset has wrong CampaignIDs. want=[%d], have=%v", campaign.ID, c.CampaignIDs)
				}
			}

			// Check that Changesets with RepoID == reposByName[wantDetached].ID
			// are detached from Campaign.
			wantIDs := make([]int64, 0, len(tt.wantDetached))
			for _, repoName := range tt.wantDetached {
				r, ok := reposByName[repoName]
				if !ok {
					t.Fatalf("unrecognized repo name: %s", repoName)
				}

				for _, c := range oldChangesets {
					if c.RepoID == r.ID {
						wantIDs = append(wantIDs, c.ID)
					}
				}
			}
			if len(wantIDs) != len(tt.wantDetached) {
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

			if diff := cmp.Diff(wantIDs, haveIDs); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func findChangesetJobsByRepoName(
	t *testing.T,
	jobsByRepo map[string]*campaigns.ChangesetJob,
	names repoNames,
) []*campaigns.ChangesetJob {
	t.Helper()

	var cs []*campaigns.ChangesetJob

	for _, n := range names {
		c, ok := jobsByRepo[n]
		if !ok {
			t.Fatalf("could not find ChangesetJob belonging to repo with name %s", n)
		}
		cs = append(cs, c)
	}

	if want, have := len(names), len(cs); want != have {
		t.Fatalf("could not find all ChangesetJobs. want=%d, have=%d", want, have)
	}
	return cs
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
	campaign *campaigns.Campaign,
	campaignJobsByID map[int64]*campaigns.CampaignJob,
	changesetStatesByCampaignJobID map[int64]campaigns.ChangesetState,
) []*campaigns.Changeset {
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

	cs := make([]*campaigns.Changeset, 0, len(campaignJobsByID))
	for _, changesetJob := range jobs {
		campaignJob, ok := campaignJobsByID[changesetJob.CampaignJobID]
		if !ok {
			t.Fatal("no CampaignJob found for ChangesetJob")
		}

		state, ok := changesetStatesByCampaignJobID[campaignJob.ID]
		if !ok {
			t.Fatal("no desired state found for Changeset")
		}
		changeset := testChangeset(campaignJob.RepoID, changesetJob.CampaignID, changesetJob.ID, state)
		err = store.CreateChangesets(ctx, changeset)
		if err != nil {
			t.Fatal(err)
		}

		cs = append(cs, changeset)

		changesetJob.ChangesetID = changeset.ID
		changesetJob.Branch = campaign.Branch
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

var createTestUser = func() func(context.Context, *testing.T) *types.User {
	count := 0

	return func(ctx context.Context, t *testing.T) *types.User {
		t.Helper()

		u := testUser
		u.Username = fmt.Sprintf("%s-%d", u.Username, count)

		user, err := db.Users.Create(ctx, u)
		if err != nil {
			t.Fatal(err)
		}

		count += 1

		return user
	}
}()

func testCampaignJob(plan int64, repo api.RepoID, t time.Time) *campaigns.CampaignJob {
	return &campaigns.CampaignJob{
		CampaignPlanID: plan,
		RepoID:         api.RepoID(repo),
		Rev:            "deadbeef",
		BaseRef:        "refs/heads/master",
		Diff:           "cool diff",
	}
}

func testCampaign(user int32, plan int64) *campaigns.Campaign {
	c := &campaigns.Campaign{
		Name:            "Testing Campaign",
		Description:     "Testing Campaign",
		AuthorID:        user,
		NamespaceUserID: user,
		CampaignPlanID:  plan,
	}

	if plan != 0 {
		c.Branch = "test-branch"
	}

	return c
}

func testChangeset(repoID api.RepoID, campaign int64, changesetJob int64, state campaigns.ChangesetState) *campaigns.Changeset {
	pr := &github.PullRequest{State: string(state)}
	return &campaigns.Changeset{
		RepoID:              repoID,
		CampaignIDs:         []int64{campaign},
		ExternalServiceType: "github",
		ExternalID:          fmt.Sprintf("ext-id-%d", changesetJob),
		Metadata:            pr,
		ExternalState:       state,
	}
}

type dummyGitserverClient struct {
	response    string
	responseErr error
}

func (d *dummyGitserverClient) CreateCommitFromPatch(ctx context.Context, req protocol.CreateCommitFromPatchRequest) (string, error) {
	return d.response, d.responseErr
}
