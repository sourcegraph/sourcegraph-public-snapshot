package campaigns

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/testing"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/schema"
)

func init() {
	dbtesting.DBNameSuffix = "campaignsenterpriserdb"
}

func TestServicePermissionLevels(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := backend.WithAuthzBypass(context.Background())
	dbtesting.SetupGlobalTestDB(t)

	now := time.Now().UTC().Truncate(time.Microsecond)

	store := NewStore(dbconn.Global)
	svc := NewService(store, nil)

	admin := createTestUser(ctx, t)
	if !admin.SiteAdmin {
		t.Fatalf("admin is not site admin")
	}

	user := createTestUser(ctx, t)
	if user.SiteAdmin {
		t.Fatalf("user cannot be site admin")
	}

	otherUser := createTestUser(ctx, t)
	if otherUser.SiteAdmin {
		t.Fatalf("user cannot be site admin")
	}

	var rs []*repos.Repo
	for i := 0; i < 4; i++ {
		rs = append(rs, testRepo(i, extsvc.TypeGitHub))
	}

	reposStore := repos.NewDBStore(dbconn.Global, sql.TxOptions{})
	err := reposStore.UpsertRepos(ctx, rs...)
	if err != nil {
		t.Fatal(err)
	}

	createTestData := func(t *testing.T, s *Store, svc *Service, author int32) (*campaigns.Campaign, []*campaigns.Patch, []*campaigns.ChangesetJob, *campaigns.Changeset) {
		patchSet := &campaigns.PatchSet{UserID: author}
		if err = s.CreatePatchSet(ctx, patchSet); err != nil {
			t.Fatal(err)
		}

		patches := make([]*campaigns.Patch, 0, len(rs))
		for _, repo := range rs {
			patch := testPatch(patchSet.ID, repo.ID, now)
			if err := s.CreatePatch(ctx, patch); err != nil {
				t.Fatal(err)
			}
			patches = append(patches, patch)
		}

		campaign := testCampaign(author, patchSet.ID)
		if err = s.CreateCampaign(ctx, campaign); err != nil {
			t.Fatal(err)
		}

		changesetJobs := make([]*campaigns.ChangesetJob, 0, len(patches))
		for _, p := range patches {
			job := &campaigns.ChangesetJob{CampaignID: campaign.ID, PatchID: p.ID}
			if err = s.CreateChangesetJob(ctx, job); err != nil {
				t.Fatal(err)
			}
			changesetJobs = append(changesetJobs, job)
		}

		changeset := testChangeset(patches[0].RepoID, changesetJobs[0].CampaignID, changesetJobs[0].ID, campaigns.ChangesetStateOpen)
		if err = s.CreateChangesets(ctx, changeset); err != nil {
			t.Fatal(err)
		}

		campaign.ChangesetIDs = append(campaign.ChangesetIDs, changeset.ID)
		if err := s.UpdateCampaign(ctx, campaign); err != nil {
			t.Fatal(err)
		}

		return campaign, patches, changesetJobs, changeset
	}

	assertAuthError := func(t *testing.T, err error) {
		t.Helper()

		if err == nil {
			t.Fatalf("expected error. got none")
		}
		if err != nil {
			if _, ok := err.(*backend.InsufficientAuthorizationError); !ok {
				t.Fatalf("wrong error: %s (%T)", err, err)
			}
		}
	}

	assertNoAuthError := func(t *testing.T, err error) {
		t.Helper()

		// Ignore other errors, we only want to check whether it's an auth error
		if _, ok := err.(*backend.InsufficientAuthorizationError); ok {
			t.Fatalf("got auth error")
		}
	}

	tests := []struct {
		name           string
		campaignAuthor int32
		currentUser    int32
		assertFunc     func(t *testing.T, err error)
	}{
		{
			name:           "unauthorized user",
			campaignAuthor: user.ID,
			currentUser:    otherUser.ID,
			assertFunc:     assertAuthError,
		},
		{
			name:           "campaign author",
			campaignAuthor: user.ID,
			currentUser:    user.ID,
			assertFunc:     assertNoAuthError,
		},

		{
			name:           "site-admin",
			campaignAuthor: user.ID,
			currentUser:    admin.ID,
			assertFunc:     assertNoAuthError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			campaign, patches, _, changeset := createTestData(t, store, svc, tc.campaignAuthor)
			// Fresh context.Background() because the previous one is wrapped in AuthzBypas
			currentUserCtx := actor.WithActor(context.Background(), actor.FromUser(tc.currentUser))

			t.Run("EnqueueChangesetJobs", func(t *testing.T) {
				_, err = svc.EnqueueChangesetJobs(currentUserCtx, campaign.ID)
				tc.assertFunc(t, err)
			})

			t.Run("RetryPublishCampaign", func(t *testing.T) {
				_, err = svc.RetryPublishCampaign(currentUserCtx, campaign.ID)
				tc.assertFunc(t, err)
			})

			t.Run("EnqueueChangesetJobForPatch", func(t *testing.T) {
				for _, p := range patches {
					err = svc.EnqueueChangesetJobForPatch(currentUserCtx, p.ID)
					tc.assertFunc(t, err)
				}
			})

			t.Run("UpdateCampaign", func(t *testing.T) {
				newName := "newName"
				_, _, err = svc.UpdateCampaign(currentUserCtx, UpdateCampaignArgs{
					Campaign: campaign.ID,
					Name:     &newName,
				})
				tc.assertFunc(t, err)
			})

			t.Run("EnqueueChangesetSync", func(t *testing.T) {
				err = svc.EnqueueChangesetSync(currentUserCtx, changeset.ID)
				tc.assertFunc(t, err)
			})

			t.Run("AddChangesetsToCampaign", func(t *testing.T) {
				_, err = svc.AddChangesetsToCampaign(currentUserCtx, campaign.ID, []int64{99999})
				tc.assertFunc(t, err)
			})

			t.Run("CloseCampaign", func(t *testing.T) {
				_, err = svc.CloseCampaign(currentUserCtx, campaign.ID, false)
				tc.assertFunc(t, err)
			})

			t.Run("DeleteCampaign", func(t *testing.T) {
				err = svc.DeleteCampaign(currentUserCtx, campaign.ID, false)
				tc.assertFunc(t, err)
			})
		})
	}
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
	cf := httpcli.NewExternalHTTPClientFactory()

	user := createTestUser(ctx, t)

	store := NewStoreWithClock(dbconn.Global, clock)

	reposStore := repos.NewDBStore(dbconn.Global, sql.TxOptions{})

	ext := &repos.ExternalService{
		Kind:        extsvc.TypeGitHub,
		DisplayName: "GitHub",
		Config: marshalJSON(t, &schema.GitHubConnection{
			Url:   "https://github.com",
			Token: "SECRETTOKEN",
		}),
	}
	if err := reposStore.UpsertExternalServices(ctx, ext); err != nil {
		t.Fatal(err)
	}

	var rs []*repos.Repo
	for i := 0; i < 4; i++ {
		r := testRepo(i, extsvc.TypeGitHub)
		r.Sources = map[string]*repos.SourceInfo{ext.URN(): {ID: ext.URN()}}

		rs = append(rs, r)
	}

	awsCodeCommitRepoID := 4
	{
		r := testRepo(awsCodeCommitRepoID, extsvc.TypeAWSCodeCommit)
		r.Sources = map[string]*repos.SourceInfo{ext.URN(): {ID: ext.URN()}}
		rs = append(rs, r)
	}

	err := reposStore.UpsertRepos(ctx, rs...)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("CreatePatchSetFromPatches", func(t *testing.T) {
		svc := NewServiceWithClock(store, nil, clock)

		input := []*campaigns.Patch{
			{RepoID: api.RepoID(rs[0].ID), Rev: "deadbeef", BaseRef: "refs/heads/master", Diff: "+-"},
			{RepoID: api.RepoID(rs[1].ID), Rev: "f00b4r", BaseRef: "refs/heads/master", Diff: "+-"},
		}

		// Filter out rs[0] in authzFilter
		db.MockAuthzFilter = func(ctx context.Context, repos []*types.Repo, p authz.Perms) ([]*types.Repo, error) {
			var filtered []*types.Repo
			for _, r := range repos {
				if r.ID == rs[0].ID {
					continue
				}
				filtered = append(filtered, r)
			}
			return filtered, nil
		}

		if _, err := svc.CreatePatchSetFromPatches(ctx, input, user.ID); !errcode.IsNotFound(err) {
			t.Fatalf("want not found error, got: %s", err)
		}

		// Now reset filter and try again
		db.MockAuthzFilter = nil

		patchSet, err := svc.CreatePatchSetFromPatches(ctx, input, user.ID)
		if err != nil {
			t.Fatal(err)
		}

		want := make([]*campaigns.Patch, 0, len(input))
		for _, in := range input {
			want = append(want, &campaigns.Patch{
				PatchSetID: patchSet.ID,
				RepoID:     in.RepoID,
				Rev:        in.Rev,
				BaseRef:    in.BaseRef,
				Diff:       in.Diff,
				CreatedAt:  now,
				UpdatedAt:  now,
			})
		}

		if _, err := store.GetPatchSet(ctx, GetPatchSetOpts{ID: patchSet.ID}); err != nil {
			t.Fatal(err)
		}

		have, _, err := store.ListPatches(ctx, ListPatchesOpts{PatchSetID: patchSet.ID})
		if err != nil {
			t.Fatal(err)
		}
		for _, p := range have {
			p.ID = 0 // ignore database ID when checking for expected output
		}
		if !cmp.Equal(have, want) {
			t.Error("have != want", cmp.Diff(have, want))
		}
	})

	t.Run("CreateCampaign", func(t *testing.T) {
		patchSet := &campaigns.PatchSet{UserID: user.ID}
		err = store.CreatePatchSet(ctx, patchSet)
		if err != nil {
			t.Fatal(err)
		}

		campaign := testCampaign(user.ID, patchSet.ID)
		svc := NewServiceWithClock(store, cf, clock)

		// Without Patches it should fail
		err = svc.CreateCampaign(ctx, campaign)
		if err != ErrNoPatches {
			t.Fatal("CreateCampaign did not produce expected error")
		}

		for _, repo := range rs {
			patch := testPatch(patchSet.ID, repo.ID, now)
			err := store.CreatePatch(ctx, patch)
			if err != nil {
				t.Fatal(err)
			}
		}

		// With Patches it should succeed
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

		// Validate no changeset jobs have been created yet.
		if len(haveJobs) != 0 {
			t.Errorf("wrong number of ChangesetJobs: %d. want=%d", len(haveJobs), 0)
		}
	})

	t.Run("DeleteCampaign", func(t *testing.T) {
		patchSet := &campaigns.PatchSet{UserID: user.ID}
		if err = store.CreatePatchSet(ctx, patchSet); err != nil {
			t.Fatal(err)
		}

		patch := testPatch(patchSet.ID, rs[0].ID, now)
		if err := store.CreatePatch(ctx, patch); err != nil {
			t.Fatal(err)
		}

		campaign := testCampaign(user.ID, patchSet.ID)

		svc := NewServiceWithClock(store, cf, clock)

		if err = svc.CreateCampaign(ctx, campaign); err != nil {
			t.Fatal(err)
		}

		// Create a processing changeset job.
		err = svc.EnqueueChangesetJobForPatch(ctx, patch.ID)
		if err != nil {
			t.Fatalf("Failed to create ChangesetJob: %s", err)
		}

		if err := svc.DeleteCampaign(ctx, campaign.ID, true); err != ErrDeleteProcessingCampaign {
			t.Fatalf("DeleteCampaign returned unexpected error: %s", err)
		}

		jobs, _, err := store.ListChangesetJobs(ctx, ListChangesetJobsOpts{
			CampaignID: campaign.ID,
			Limit:      -1,
		})
		if err != nil {
			t.Fatal(err)
		}

		if len(jobs) != 1 {
			t.Fatalf("wrong number of changeset jobs: %d", len(jobs))
		}

		for _, j := range jobs {
			j.Error = "failed"
			j.FinishedAt = clock()
			if err := store.UpdateChangesetJob(ctx, j); err != nil {
				t.Fatalf("updating changeset job failed: %s\n", err)
			}
		}

		// Now it should work, since the jobs failed to execute and campaign is
		// no longer processing.
		if err := svc.DeleteCampaign(ctx, campaign.ID, true); err != nil {
			t.Fatalf("campaign not deleted: %s", err)
		}
	})

	t.Run("CloseCampaign", func(t *testing.T) {
		patchSet := &campaigns.PatchSet{UserID: user.ID}
		if err = store.CreatePatchSet(ctx, patchSet); err != nil {
			t.Fatal(err)
		}

		patch := testPatch(patchSet.ID, rs[0].ID, now)
		if err := store.CreatePatch(ctx, patch); err != nil {
			t.Fatal(err)
		}

		campaign := testCampaign(user.ID, patchSet.ID)

		svc := NewServiceWithClock(store, cf, clock)

		if err = svc.CreateCampaign(ctx, campaign); err != nil {
			t.Fatal(err)
		}

		// Create a processing changeset job.
		err = svc.EnqueueChangesetJobForPatch(ctx, patch.ID)
		if err != nil {
			t.Fatalf("Failed to create ChangesetJob: %s", err)
		}

		if _, err = svc.CloseCampaign(ctx, campaign.ID, true); err != ErrCloseProcessingCampaign {
			t.Fatalf("CloseCampaign returned unexpected error: %s", err)
		}

		jobs, _, err := store.ListChangesetJobs(ctx, ListChangesetJobsOpts{
			CampaignID: campaign.ID,
			Limit:      -1,
		})
		if err != nil {
			t.Fatal(err)
		}

		if len(jobs) != 1 {
			t.Fatalf("wrong number of changeset jobs: %d", len(jobs))
		}

		for _, j := range jobs {
			j.Error = "failed"
			j.FinishedAt = clock()
			if err := store.UpdateChangesetJob(ctx, j); err != nil {
				t.Fatalf("updating changeset job failed: %s\n", err)
			}
		}

		// Now it should work, since the jobs failed to execute and campaign is
		// no longer processing.
		campaign, err = svc.CloseCampaign(ctx, campaign.ID, true)
		if err != nil {
			t.Fatalf("campaign not closed: %s", err)
		}
		if campaign.ClosedAt.IsZero() {
			t.Fatalf("campaign ClosedAt is zero")
		}
	})

	t.Run("CreateCampaignWithPatchSetAttachedToOtherCampaign", func(t *testing.T) {
		patchSet := &campaigns.PatchSet{UserID: user.ID}
		err = store.CreatePatchSet(ctx, patchSet)
		if err != nil {
			t.Fatal(err)
		}

		for _, repo := range rs {
			err := store.CreatePatch(ctx, testPatch(patchSet.ID, repo.ID, now))
			if err != nil {
				t.Fatal(err)
			}
		}

		campaign := testCampaign(user.ID, patchSet.ID)
		svc := NewServiceWithClock(store, cf, clock)

		err = svc.CreateCampaign(ctx, campaign)
		if err != nil {
			t.Fatal(err)
		}

		otherCampaign := testCampaign(user.ID, patchSet.ID)
		err = svc.CreateCampaign(ctx, otherCampaign)
		if err != ErrPatchSetDuplicate {
			t.Fatal("no error even though another campaign has same patch set")
		}
	})

	t.Run("EnqueueChangesetJobForPatch", func(t *testing.T) {
		patchSet := &campaigns.PatchSet{UserID: user.ID}
		err = store.CreatePatchSet(ctx, patchSet)
		if err != nil {
			t.Fatal(err)
		}

		patch := testPatch(patchSet.ID, rs[0].ID, now)
		err := store.CreatePatch(ctx, patch)
		if err != nil {
			t.Fatal(err)
		}

		campaign := testCampaign(user.ID, patchSet.ID)
		err = store.CreateCampaign(ctx, campaign)
		if err != nil {
			t.Fatal(err)
		}

		svc := NewServiceWithClock(store, cf, clock)

		// Filter out the repository in the authz filter
		db.MockAuthzFilter = func(ctx context.Context, repos []*types.Repo, p authz.Perms) ([]*types.Repo, error) {
			return []*types.Repo{}, nil
		}
		// should result in a not found error
		if err = svc.EnqueueChangesetJobForPatch(ctx, patch.ID); !errcode.IsNotFound(err) {
			t.Fatalf("want not found error, got: %s", err)
		}

		// Now reset the filter
		db.MockAuthzFilter = nil

		if err = svc.EnqueueChangesetJobForPatch(ctx, patch.ID); err != nil {
			t.Fatal(err)
		}
		haveJob, err := store.GetChangesetJob(ctx, GetChangesetJobOpts{
			CampaignID: campaign.ID,
			PatchID:    patch.ID,
		})
		if err != nil {
			t.Fatal(err)
		}
		if have, want := haveJob.PatchID, patch.ID; have != want {
			t.Fatalf("ChangesetJob has wrong PatchID. want=%d, have=%d", want, have)
		}

		// Try to create again, check that it's the same one
		err = svc.EnqueueChangesetJobForPatch(ctx, patch.ID)
		if err != nil {
			t.Fatal(err)
		}
		haveJob2, err := store.GetChangesetJob(ctx, GetChangesetJobOpts{
			CampaignID: campaign.ID,
			PatchID:    patch.ID,
		})
		if err != nil {
			t.Fatal(err)
		}

		if haveJob2.ID != haveJob.ID {
			t.Errorf("wrong changesetJob: %d. want=%d", haveJob2.ID, haveJob.ID)
		}

		// Error out the changeset job and verify that
		// EnqueueChangesetJobForPatch updates the job to force a retry.
		haveJob.Error = "ruh roh"
		haveJob.StartedAt = time.Now()
		haveJob.FinishedAt = time.Now()
		if err := store.UpdateChangesetJob(ctx, haveJob); err != nil {
			t.Fatal(err)
		}
		// Sanity check: did this result in the job being considered
		// unsuccessfully completed?
		if !haveJob.UnsuccessfullyCompleted() {
			t.Error("tried to error out the changesetJob and failed")
		}
		if err := svc.EnqueueChangesetJobForPatch(ctx, patch.ID); err != nil {
			t.Fatal(err)
		}
		haveJob3, err := store.GetChangesetJob(ctx, GetChangesetJobOpts{
			CampaignID: campaign.ID,
			PatchID:    patch.ID,
		})
		if err != nil {
			t.Fatal(err)
		}
		if haveJob3.ID != haveJob.ID {
			t.Errorf("wrong changesetJob: %d. want=%d", haveJob3.ID, haveJob.ID)
		}
		if haveJob3.Error != "" {
			t.Errorf("unexpected changesetJob Error value: %s. want=%s", haveJob3.Error, "")
		}
		if !haveJob3.StartedAt.IsZero() {
			t.Errorf("unexpected changesetJob StartedAt value: %v. want=%v", haveJob3.StartedAt, time.Time{})
		}
		if !haveJob3.FinishedAt.IsZero() {
			t.Errorf("unexpected changesetJob FinishedAt value: %v. want=%v", haveJob3.FinishedAt, time.Time{})
		}

		// Update repo to unsupported codehost type.
		awsPatch := testPatch(patchSet.ID, rs[awsCodeCommitRepoID].ID, now)
		if err := store.CreatePatch(ctx, awsPatch); err != nil {
			t.Fatal(err)
		}
		if wantErr, haveErr := ErrUnsupportedCodehost, svc.EnqueueChangesetJobForPatch(ctx, awsPatch.ID); wantErr != haveErr {
			t.Fatalf("got invalid error, want=%v have=%v", wantErr, haveErr)
		}
	})

	t.Run("EnqueueChangesetJobs", func(t *testing.T) {
		patchSet := &campaigns.PatchSet{UserID: user.ID}
		if err = store.CreatePatchSet(ctx, patchSet); err != nil {
			t.Fatal(err)
		}

		patches := make([]*campaigns.Patch, 0, len(rs))
		patchesByID := make(map[int64]*campaigns.Patch, len(rs))
		for _, r := range rs {
			patch := testPatch(patchSet.ID, r.ID, now)
			if err := store.CreatePatch(ctx, patch); err != nil {
				t.Fatal(err)
			}
			patches = append(patches, patch)
			patchesByID[patch.ID] = patch
		}

		campaign := testCampaign(user.ID, patchSet.ID)
		if err = store.CreateCampaign(ctx, campaign); err != nil {
			t.Fatal(err)
		}

		// Filter out one repository to make sure it's skipped
		filteredOutPatch := patches[len(patches)-2]
		db.MockAuthzFilter = func(ctx context.Context, repos []*types.Repo, p authz.Perms) ([]*types.Repo, error) {
			var filtered []*types.Repo
			for _, r := range repos {
				if r.ID == filteredOutPatch.RepoID {
					continue
				}
				filtered = append(filtered, r)
			}
			return filtered, nil
		}
		defer func() { db.MockAuthzFilter = nil }()

		svc := NewServiceWithClock(store, cf, clock)
		if _, err = svc.EnqueueChangesetJobs(ctx, campaign.ID); err != nil {
			t.Fatal(err)
		}

		haveJobs, _, err := store.ListChangesetJobs(ctx, ListChangesetJobsOpts{
			CampaignID: campaign.ID,
		})
		if err != nil {
			t.Fatal(err)
		}

		wantJobsCount := len(patches) - 2 // We filtered out one repository, and one is unsupported.
		if have, want := len(haveJobs), wantJobsCount; have != want {
			t.Fatal("wrong number of changeset jobs created")
		}

		for _, job := range haveJobs {
			if job.PatchID == filteredOutPatch.ID {
				continue
			}
			if _, ok := patchesByID[job.PatchID]; !ok {
				t.Fatalf("job %d has wrong patch id %d", job.ID, job.PatchID)
			}
		}

		// Try again, check that no new jobs have been created
		if _, err = svc.EnqueueChangesetJobs(ctx, campaign.ID); err != nil {
			t.Fatal(err)
		}
		haveJobs2, _, err := store.ListChangesetJobs(ctx, ListChangesetJobsOpts{
			CampaignID: campaign.ID,
		})
		if err != nil {
			t.Fatal(err)
		}

		if have, want := len(haveJobs2), len(haveJobs); have != want {
			t.Fatal("wrong number of changeset jobs created")
		}

		for _, job := range haveJobs2 {
			if _, ok := patchesByID[job.PatchID]; !ok {
				t.Fatalf("job %d has wrong patch id %d", job.ID, job.PatchID)
			}
		}

		// Now mark one ChangesetJob as failed and check that it's not reset
		failedJob := haveJobs2[0]
		failedJob.Error = "ruh roh"
		failedJob.StartedAt = time.Now()
		failedJob.FinishedAt = time.Now()
		if err := store.UpdateChangesetJob(ctx, failedJob); err != nil {
			t.Fatal(err)
		}

		// Try again, check that the job was not reset
		if _, err = svc.EnqueueChangesetJobs(ctx, campaign.ID); err != nil {
			t.Fatal(err)
		}
		haveJobs3, _, err := store.ListChangesetJobs(ctx, ListChangesetJobsOpts{
			CampaignID: campaign.ID,
		})
		if err != nil {
			t.Fatal(err)
		}

		if have, want := len(haveJobs3), len(haveJobs2); have != want {
			t.Fatal("wrong number of changeset jobs created")
		}

		found := false
		for _, job := range haveJobs3 {
			if job.ID == failedJob.ID {
				found = true

				if job.Error == "" {
					t.Fatalf("job was reset")
				}
			}
		}
		if !found {
			t.Fatalf("failed job not found")
		}
	})

	t.Run("GetCampaignStatus", func(t *testing.T) {
		// Make sure that user is an admin
		if !user.SiteAdmin {
			t.Fatalf("user is not an admin")
		}

		otherUser := createTestUser(ctx, t)
		if otherUser.SiteAdmin {
			t.Fatalf("otherUser is admin")
		}

		patchSet := &campaigns.PatchSet{UserID: otherUser.ID}
		if err = store.CreatePatchSet(ctx, patchSet); err != nil {
			t.Fatal(err)
		}

		patches := make([]*campaigns.Patch, 0, len(rs))
		for _, repo := range rs {
			patch := testPatch(patchSet.ID, repo.ID, now)
			if err := store.CreatePatch(ctx, patch); err != nil {
				t.Fatal(err)
			}
			patches = append(patches, patch)
		}

		campaign := testCampaign(otherUser.ID, patchSet.ID)
		if err = store.CreateCampaign(ctx, campaign); err != nil {
			t.Fatal(err)
		}

		changesetJobs := make([]*campaigns.ChangesetJob, 0, len(patches))
		for _, p := range patches {
			job := &campaigns.ChangesetJob{
				CampaignID: campaign.ID,
				PatchID:    p.ID,
				StartedAt:  clock(),
				FinishedAt: clock(),
				Error:      "error",
			}
			if err = store.CreateChangesetJob(ctx, job); err != nil {
				t.Fatal(err)
			}
			changesetJobs = append(changesetJobs, job)
		}

		// As site-admin
		userCtx := actor.WithActor(context.Background(), actor.FromUser(user.ID))
		svc := NewService(store, cf)
		status, err := svc.GetCampaignStatus(userCtx, campaign)
		if err != nil {
			t.Fatal(err)
		}

		if have, want := len(status.ProcessErrors), len(changesetJobs); have != want {
			t.Fatalf("wrong number of errors returned. want=%d, have=%d", want, have)
		}

		// As author of campaign and non-site-admin
		otherUserCtx := actor.WithActor(context.Background(), actor.FromUser(otherUser.ID))
		status, err = svc.GetCampaignStatus(otherUserCtx, campaign)
		if err != nil {
			t.Fatal(err)
		}

		if have, want := len(status.ProcessErrors), len(changesetJobs); have != want {
			t.Fatalf("wrong number of errors returned. want=%d, have=%d", want, have)
		}

		// As author of campaign and non-site-admin with filtered out repositories
		db.MockAuthzFilter = func(ctx context.Context, repos []*types.Repo, p authz.Perms) ([]*types.Repo, error) {
			var filtered []*types.Repo
			for _, r := range repos {
				// Filter out one repository
				if r.ID == patches[0].RepoID {
					continue
				}
				filtered = append(filtered, r)
			}
			return filtered, nil
		}
		defer func() { db.MockAuthzFilter = nil }()

		status, err = svc.GetCampaignStatus(otherUserCtx, campaign)
		if err != nil {
			t.Fatal(err)
		}

		// One less error
		if have, want := len(status.ProcessErrors), len(changesetJobs)-1; have != want {
			t.Fatalf("wrong number of errors returned. want=%d, have=%d", want, have)
		}

		// Change author of campaign to site-admin
		campaign.AuthorID = user.ID
		if err = store.UpdateCampaign(ctx, campaign); err != nil {
			t.Fatal(err)
		}

		// As non-author and non-site-admin
		status, err = svc.GetCampaignStatus(otherUserCtx, campaign)
		if err != nil {
			t.Fatal(err)
		}

		if have, want := len(status.ProcessErrors), 0; have != want {
			t.Fatalf("wrong number of errors returned. want=%d, have=%d", want, have)
		}
	})

	t.Run("UpdateCampaign", func(t *testing.T) {
		strPointer := func(s string) *string { return &s }
		subTests := []struct {
			name      string
			branch    *string
			createJob bool
			closed    bool
			process   bool
			err       string
		}{
			{
				name:      "published unprocessed campaign",
				createJob: true,
				err:       ErrUpdateProcessingCampaign.Error(),
			},
			{
				name:   "closed campaign",
				closed: true,
				err:    ErrUpdateClosedCampaign.Error(),
			},
			{
				name:   "change campaign branch",
				branch: strPointer("changed-branch"),
			},
			{
				name:      "change published campaign branch",
				branch:    strPointer("changed-branch"),
				createJob: true,
				process:   true,
				err:       ErrPublishedCampaignBranchChange.Error(),
			},
			{
				name:   "change campaign blank branch",
				branch: strPointer(""),
				err:    ErrCampaignBranchBlank.Error(),
			},
			{
				name:   "change campaign invalid branch",
				branch: strPointer("invalid-branch."),
				err:    ErrCampaignBranchInvalid.Error(),
			},
		}
		for _, tc := range subTests {
			t.Run(tc.name, func(t *testing.T) {
				if tc.err == "" {
					tc.err = "<nil>"
				}

				patchSet := &campaigns.PatchSet{UserID: user.ID}
				err = store.CreatePatchSet(ctx, patchSet)
				if err != nil {
					t.Fatal(err)
				}

				patch := testPatch(patchSet.ID, rs[0].ID, now)
				err := store.CreatePatch(ctx, patch)
				if err != nil {
					t.Fatal(err)
				}

				svc := NewServiceWithClock(store, cf, clock)
				campaign := testCampaign(user.ID, patchSet.ID)

				if tc.closed {
					campaign.ClosedAt = now
				}

				err = svc.CreateCampaign(ctx, campaign)
				if err != nil {
					t.Fatal(err)
				}

				haveJobs, _, err := store.ListChangesetJobs(ctx, ListChangesetJobsOpts{
					CampaignID: campaign.ID,
				})
				if err != nil {
					t.Fatal(err)
				}

				// sanity check
				if len(haveJobs) != 0 {
					t.Errorf("wrong number of ChangesetJobs: %d. want=%d", len(haveJobs), 0)
				}

				if tc.createJob {
					// Create a processing changeset job.
					err = svc.EnqueueChangesetJobForPatch(ctx, patch.ID)
					if err != nil {
						t.Fatalf("Failed to create ChangesetJob: %s", err)
					}

					// Refetch jobs
					haveJobs, _, err = store.ListChangesetJobs(ctx, ListChangesetJobsOpts{
						CampaignID: campaign.ID,
					})
					if err != nil {
						t.Fatal(err)
					}
				}

				// sanity checks
				wantJobs := 0
				if tc.createJob {
					wantJobs = 1
				}
				if len(haveJobs) != wantJobs {
					t.Errorf("wrong number of ChangesetJobs: %d. want=%d", len(haveJobs), wantJobs)
				}

				if wantJobs > 0 {
					if !haveJobs[0].StartedAt.IsZero() {
						t.Errorf("ChangesetJobs is not unprocessed. StartedAt=%v", haveJobs[0].StartedAt)
					}
				}

				if tc.process {
					patchesByID := map[int64]*campaigns.Patch{
						patch.ID: patch,
					}
					states := map[int64]campaigns.ChangesetState{
						patch.ID: campaigns.ChangesetStateOpen,
					}
					fakeRunChangesetJobs(ctx, t, store, now, campaign, patchesByID, states)
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

	t.Run("UpdateCampaignWithPatchSetAttachedToOtherCampaign", func(t *testing.T) {
		svc := NewServiceWithClock(store, cf, clock)

		patchSet := &campaigns.PatchSet{UserID: user.ID}
		err = store.CreatePatchSet(ctx, patchSet)
		if err != nil {
			t.Fatal(err)
		}

		for _, repo := range rs {
			err := store.CreatePatch(ctx, testPatch(patchSet.ID, repo.ID, now))
			if err != nil {
				t.Fatal(err)
			}
		}

		campaign := testCampaign(user.ID, patchSet.ID)
		err = svc.CreateCampaign(ctx, campaign)
		if err != nil {
			t.Fatal(err)
		}

		otherPatchSet := &campaigns.PatchSet{UserID: user.ID}
		err = store.CreatePatchSet(ctx, otherPatchSet)
		if err != nil {
			t.Fatal(err)
		}

		for _, repo := range rs {
			err := store.CreatePatch(ctx, testPatch(otherPatchSet.ID, repo.ID, now))
			if err != nil {
				t.Fatal(err)
			}
		}
		otherCampaign := testCampaign(user.ID, otherPatchSet.ID)
		err = svc.CreateCampaign(ctx, otherCampaign)
		if err != nil {
			t.Fatal(err)
		}

		args := UpdateCampaignArgs{Campaign: otherCampaign.ID, PatchSet: &patchSet.ID}
		_, _, err := svc.UpdateCampaign(ctx, args)
		if err != ErrPatchSetDuplicate {
			t.Fatal("no error even though another campaign has same patch set")
		}
	})

	t.Run("EnqueueChangesetSync", func(t *testing.T) {
		svc := NewServiceWithClock(store, cf, clock)

		campaign := testCampaign(user.ID, 0)
		if err = store.CreateCampaign(ctx, campaign); err != nil {
			t.Fatal(err)
		}

		changeset := testChangeset(rs[0].ID, campaign.ID, 0, campaigns.ChangesetStateOpen)
		if err = store.CreateChangesets(ctx, changeset); err != nil {
			t.Fatal(err)
		}

		campaign.ChangesetIDs = []int64{changeset.ID}
		if err = store.UpdateCampaign(ctx, campaign); err != nil {
			t.Fatal(err)
		}

		called := false
		repoupdater.MockEnqueueChangesetSync = func(ctx context.Context, ids []int64) error {
			if len(ids) != 1 && ids[0] != changeset.ID {
				t.Fatalf("MockEnqueueChangesetSync received wrong ids: %+v", ids)
			}
			called = true
			return nil
		}
		t.Cleanup(func() { repoupdater.MockEnqueueChangesetSync = nil })

		if err := svc.EnqueueChangesetSync(ctx, changeset.ID); err != nil {
			t.Fatal(err)
		}

		if !called {
			t.Fatal("MockEnqueueChangesetSync not called")
		}

		// Repo filtered out by authzFilter
		db.MockAuthzFilter = func(ctx context.Context, repos []*types.Repo, p authz.Perms) ([]*types.Repo, error) {
			return []*types.Repo{}, nil
		}
		t.Cleanup(func() { db.MockAuthzFilter = nil })

		// should result in a not found error
		if err := svc.EnqueueChangesetSync(ctx, changeset.ID); !errcode.IsNotFound(err) {
			t.Fatalf("expected not-found error but got %s", err)
		}
	})

	t.Run("AddChangesetsToCampaign", func(t *testing.T) {
		svc := NewServiceWithClock(store, cf, clock)

		campaign := testCampaign(user.ID, 0)
		if err = store.CreateCampaign(ctx, campaign); err != nil {
			t.Fatal(err)
		}

		changeset := testChangeset(rs[0].ID, 0, 98765, campaigns.ChangesetStateOpen)
		if err = store.CreateChangesets(ctx, changeset); err != nil {
			t.Fatal(err)
		}

		changeset2 := testChangeset(rs[1].ID, 0, 12345, campaigns.ChangesetStateOpen)
		if err = store.CreateChangesets(ctx, changeset2); err != nil {
			t.Fatal(err)
		}

		campaign, err = svc.AddChangesetsToCampaign(ctx, campaign.ID, []int64{changeset.ID})
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff([]int64{changeset.ID}, campaign.ChangesetIDs); diff != "" {
			t.Fatalf("campaign.ChangesetIDs is wrong: %s", diff)
		}

		changeset, err = store.GetChangeset(ctx, GetChangesetOpts{ID: changeset.ID})
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff([]int64{campaign.ID}, changeset.CampaignIDs); diff != "" {
			t.Fatalf("changeset.CampaignIDs is wrong: %s", diff)
		}

		// Repo filtered out by authzFilter
		db.MockAuthzFilter = func(ctx context.Context, repos []*types.Repo, p authz.Perms) ([]*types.Repo, error) {
			return []*types.Repo{}, nil
		}
		t.Cleanup(func() { db.MockAuthzFilter = nil })

		_, err = svc.AddChangesetsToCampaign(ctx, campaign.ID, []int64{changeset2.ID})
		if !errcode.IsNotFound(err) {
			t.Fatalf("expected not-found error but got %s", err)
		}
	})

	t.Run("CloseOpenChangesets", func(t *testing.T) {
		changeset1 := testChangeset(rs[0].ID, 0, 121314, campaigns.ChangesetStateOpen)
		changeset2 := testChangeset(rs[1].ID, 0, 141516, campaigns.ChangesetStateOpen)
		if err = store.CreateChangesets(ctx, changeset1, changeset2); err != nil {
			t.Fatal(err)
		}

		// Repo of changeset2 filtered out by authzFilter
		db.MockAuthzFilter = func(ctx context.Context, repos []*types.Repo, p authz.Perms) (filtered []*types.Repo, err error) {
			for _, r := range repos {
				if r.ID == changeset2.RepoID {
					continue
				}
				filtered = append(filtered, r)
			}
			return filtered, nil
		}
		t.Cleanup(func() { db.MockAuthzFilter = nil })

		fakeSource := &ct.FakeChangesetSource{Err: nil}
		sourcer := repos.NewFakeSourcer(nil, fakeSource)

		svc := NewServiceWithClock(store, cf, clock)
		svc.sourcer = sourcer

		// Try to close open changesets
		err := svc.CloseOpenChangesets(ctx, []*campaigns.Changeset{changeset1, changeset2})
		if err != nil {
			t.Fatal(err)
		}

		// Only changeset1 should be closed
		if have, want := len(fakeSource.ClosedChangesets), 1; have != want {
			t.Fatalf("ClosedChangesets has wrong length. want=%d, have=%d", want, have)
		}

		if have, want := fakeSource.ClosedChangesets[0].RepoID, changeset1.RepoID; have != want {
			t.Fatalf("wrong changesets closed. want=%d, have=%d", want, have)
		}
	})

	t.Run("RetryPublishCampaign", func(t *testing.T) {
		patchSet := &campaigns.PatchSet{UserID: user.ID}
		if err = store.CreatePatchSet(ctx, patchSet); err != nil {
			t.Fatal(err)
		}

		patches := make([]*campaigns.Patch, 0, len(rs))
		for _, repo := range rs {
			patch := testPatch(patchSet.ID, repo.ID, now)
			if err := store.CreatePatch(ctx, patch); err != nil {
				t.Fatal(err)
			}
			patches = append(patches, patch)
		}

		campaign := testCampaign(user.ID, patchSet.ID)
		if err = store.CreateCampaign(ctx, campaign); err != nil {
			t.Fatal(err)
		}

		changesetJobs := make([]*campaigns.ChangesetJob, 0, len(patches))
		for _, p := range patches {
			job := &campaigns.ChangesetJob{
				CampaignID: campaign.ID,
				PatchID:    p.ID,
				StartedAt:  clock(),
				FinishedAt: clock(),
				Error:      "error",
			}
			if err = store.CreateChangesetJob(ctx, job); err != nil {
				t.Fatal(err)
			}
			changesetJobs = append(changesetJobs, job)
		}

		// Repo of patches[0]/changesetJobs[0] filtered out by authzFilter
		db.MockAuthzFilter = func(ctx context.Context, repos []*types.Repo, p authz.Perms) (filtered []*types.Repo, err error) {
			for _, r := range repos {
				if r.ID == patches[0].RepoID {
					continue
				}
				filtered = append(filtered, r)
			}
			return filtered, nil
		}
		t.Cleanup(func() { db.MockAuthzFilter = nil })

		svc := NewServiceWithClock(store, cf, clock)
		_, err := svc.RetryPublishCampaign(ctx, campaign.ID)
		if err != nil {
			t.Fatal(err)
		}

		reloadedJobs, _, err := store.ListChangesetJobs(ctx, ListChangesetJobsOpts{
			CampaignID: campaign.ID,
			Limit:      -1,
		})
		if err != nil {
			t.Fatal(err)
		}

		if have, want := len(reloadedJobs), len(changesetJobs); have != want {
			t.Fatalf("wrong number of failed changeset jobs. want=%d, have=%d", want, have)
		}

		for _, j := range reloadedJobs {
			if j.PatchID == patches[0].ID {
				if !j.UnsuccessfullyCompleted() {
					t.Fatalf("ChangesetJob %d reset, but should have been filtered out", j.ID)
				}
				continue
			}

			if j.UnsuccessfullyCompleted() {
				t.Fatalf("ChangesetJob %d not reset", j.ID)
			}
		}
	})
}

type repoNames []string

type newPatchSpec struct {
	repo string

	modifiedDiff bool
	modifiedRev  bool
}

func TestService_UpdateCampaignWithNewPatchSetID(t *testing.T) {
	ctx := backend.WithAuthzBypass(context.Background())
	dbtesting.SetupGlobalTestDB(t)

	now := time.Now().UTC().Truncate(time.Microsecond)
	clock := func() time.Time {
		return now.UTC().Truncate(time.Microsecond)
	}

	cf := httpcli.NewExternalHTTPClientFactory()

	user := createTestUser(ctx, t)

	var rs []*repos.Repo
	for i := 0; i < 4; i++ {
		rs = append(rs, testRepo(i, extsvc.TypeGitHub))
	}
	rs = append(rs, testRepo(len(rs), extsvc.TypeAWSCodeCommit))

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

		campaignWithoutPatchSet bool
		campaignIsClosed        bool

		// Repositories for which we had Patches attached to the old PatchSet
		oldPatches repoNames

		// Repositories for which the ChangesetJob/Changeset have been
		// individually published
		individuallyPublished repoNames

		// Repositories for which the actor doesn't have repository permissions
		missingRepoPerms repoNames

		// Mapping of repository names to state of changesets after creating the campaign.
		// Default state is ChangesetStateOpen
		changesetStates map[string]campaigns.ChangesetState

		updatePatchSet, updateName, updateDescription bool
		newPatches                                    []newPatchSpec

		// Repositories for which we want no Changeset/ChangesetJob after update
		wantDetached repoNames
		// Repositories for which we want to keep Changeset/ChangesetJob unmodified
		wantUnmodified repoNames
		// Repositories for which we want to keep Changeset/ChangesetJob and update them
		wantModified repoNames
		// Repositories for which we want to create a new ChangesetJob (and thus a Changeset)
		wantCreated repoNames
		// An error to be thrown when attempting to do the update
		wantErr string
	}{
		{
			name:                    "campaign without patchset, no new patch set, name update",
			campaignWithoutPatchSet: true,
			updateName:              true,
		},
		{
			name:           "1 unmodified",
			updatePatchSet: true,
			oldPatches:     repoNames{"repo-0"},
			newPatches: []newPatchSpec{
				{repo: "repo-0"},
			},
			wantUnmodified: repoNames{"repo-0"},
		},
		{
			name:         "no new patch set but name update",
			updateName:   true,
			oldPatches:   repoNames{"repo-0"},
			wantModified: repoNames{"repo-0"},
		},
		{
			name:              "no new patch set but description update",
			updateDescription: true,
			oldPatches:        repoNames{"repo-0"},
			wantModified:      repoNames{"repo-0"},
		},
		{
			name:             "no new patch set but name update, missing repo permissions",
			updateName:       true,
			oldPatches:       repoNames{"repo-0", "repo-1"},
			missingRepoPerms: repoNames{"repo-1"},
			wantModified:     repoNames{"repo-0"},
			wantUnmodified:   repoNames{"repo-1"},
		},
		{
			name:              "no new patch set but description update, missing repo permissions",
			updateDescription: true,
			oldPatches:        repoNames{"repo-0", "repo-1"},
			missingRepoPerms:  repoNames{"repo-1"},
			wantModified:      repoNames{"repo-0"},
			wantUnmodified:    repoNames{"repo-1"},
		},
		{
			name:           "1 modified diff",
			updatePatchSet: true,
			oldPatches:     repoNames{"repo-0"},
			newPatches: []newPatchSpec{
				{repo: "repo-0", modifiedDiff: true},
			},
			wantModified: repoNames{"repo-0"},
		},
		{
			name:           "1 modified rev",
			updatePatchSet: true,
			oldPatches:     repoNames{"repo-0"},
			newPatches: []newPatchSpec{
				{repo: "repo-0", modifiedRev: true},
			},
			wantModified: repoNames{"repo-0"},
		},
		{
			name:           "1 detached, 1 unmodified, 1 modified, 1 new changeset",
			updatePatchSet: true,
			oldPatches:     repoNames{"repo-0", "repo-1", "repo-2"},
			newPatches: []newPatchSpec{
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
			name:            "1 modified diff for already merged changeset",
			updatePatchSet:  true,
			oldPatches:      repoNames{"repo-0"},
			changesetStates: map[string]campaigns.ChangesetState{"repo-0": campaigns.ChangesetStateMerged},
			newPatches: []newPatchSpec{
				{repo: "repo-0", modifiedDiff: true},
			},
			wantUnmodified: repoNames{"repo-0"},
		},
		{
			name:            "1 modified rev for already merged changeset",
			updatePatchSet:  true,
			oldPatches:      repoNames{"repo-0"},
			changesetStates: map[string]campaigns.ChangesetState{"repo-0": campaigns.ChangesetStateMerged},
			newPatches: []newPatchSpec{
				{repo: "repo-0", modifiedDiff: true},
			},
			wantUnmodified: repoNames{"repo-0"},
		},
		{
			name:            "1 modified diff for already closed changeset",
			updatePatchSet:  true,
			oldPatches:      repoNames{"repo-0"},
			changesetStates: map[string]campaigns.ChangesetState{"repo-0": campaigns.ChangesetStateClosed},
			newPatches: []newPatchSpec{
				{repo: "repo-0", modifiedDiff: true},
			},
			wantUnmodified: repoNames{"repo-0"},
		},
		{
			name:            "1 modified rev for already closed changeset",
			updatePatchSet:  true,
			oldPatches:      repoNames{"repo-0"},
			changesetStates: map[string]campaigns.ChangesetState{"repo-0": campaigns.ChangesetStateClosed},
			newPatches: []newPatchSpec{
				{repo: "repo-0", modifiedDiff: true},
			},
			wantUnmodified: repoNames{"repo-0"},
		},
		{
			name:                    "set patchset on campaign without patchset",
			updatePatchSet:          true,
			campaignWithoutPatchSet: true,
			newPatches: []newPatchSpec{
				{repo: "repo-0", modifiedDiff: true},
			},
			wantErr: "",
		},
		{
			name:             "update patchset on closed campaign",
			updatePatchSet:   true,
			campaignIsClosed: true,
			oldPatches:       repoNames{"repo-0"},
			changesetStates:  map[string]campaigns.ChangesetState{"repo-0": campaigns.ChangesetStateOpen},
			newPatches: []newPatchSpec{
				{repo: "repo-0", modifiedDiff: true},
			},
			wantErr: ErrUpdateClosedCampaign.Error(),
		},
		{
			name:            "1 unmodified merged, 1 new changeset",
			updatePatchSet:  true,
			oldPatches:      repoNames{"repo-0"},
			changesetStates: map[string]campaigns.ChangesetState{"repo-0": campaigns.ChangesetStateMerged},
			newPatches: []newPatchSpec{
				{repo: "repo-1"},
			},
			wantUnmodified: repoNames{"repo-0"},
			wantCreated:    repoNames{"repo-1"},
		},
		{
			name:            "1 unmodified closed, 1 new changeset",
			updatePatchSet:  true,
			oldPatches:      repoNames{"repo-0"},
			changesetStates: map[string]campaigns.ChangesetState{"repo-0": campaigns.ChangesetStateClosed},
			newPatches: []newPatchSpec{
				{repo: "repo-1"},
			},
			wantUnmodified: repoNames{"repo-0"},
			wantCreated:    repoNames{"repo-1"},
		},
		{
			name:           "1 modified but missing permissions",
			updatePatchSet: true,
			oldPatches:     repoNames{"repo-0"},
			newPatches: []newPatchSpec{
				{repo: "repo-0", modifiedDiff: true},
			},

			missingRepoPerms: repoNames{"repo-0"},
			wantErr:          (&db.RepoNotFoundErr{ID: rs[0].ID}).Error(),
		},
		{
			name:           "1 unmodified, 1 created but missing permissions for created",
			updatePatchSet: true,
			oldPatches:     repoNames{"repo-0", "repo-1"},
			newPatches: []newPatchSpec{
				{repo: "repo-1"},
			},
			missingRepoPerms: repoNames{"repo-1"},
			wantErr:          (&db.RepoNotFoundErr{ID: rs[1].ID}).Error(),
		},
		{
			name:           "1 detached, 1 created, missing repo permissions for deatched",
			updatePatchSet: true,
			oldPatches:     repoNames{"repo-0"},
			newPatches: []newPatchSpec{
				{repo: "repo-1", modifiedDiff: true},
			},
			missingRepoPerms: repoNames{"repo-0"},
			wantUnmodified:   repoNames{"repo-0"},
			wantCreated:      repoNames{"repo-1"},
		},
		{
			name:           "1 added on unsupported codehost",
			updatePatchSet: true,
			oldPatches:     repoNames{"repo-0"},
			newPatches: []newPatchSpec{
				{repo: "repo-0"},
				{repo: "repo-4"},
			},
			wantUnmodified: repoNames{"repo-0"},
			wantCreated:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewStoreWithClock(dbconn.Global, clock)
			svc := NewServiceWithClock(store, cf, clock)

			var (
				campaign    *campaigns.Campaign
				oldPatches  []*campaigns.Patch
				patchesByID map[int64]*campaigns.Patch

				changesetStateByPatchID map[int64]campaigns.ChangesetState

				oldChangesets []*campaigns.Changeset
			)

			patchesByID = make(map[int64]*campaigns.Patch)

			if tt.campaignWithoutPatchSet {
				campaign = testCampaign(user.ID, 0)
			} else {
				patchSet := &campaigns.PatchSet{UserID: user.ID}
				err = store.CreatePatchSet(ctx, patchSet)
				if err != nil {
					t.Fatal(err)
				}
				changesetStateByPatchID = make(map[int64]campaigns.ChangesetState)
				for _, repoName := range tt.oldPatches {
					repo, ok := reposByName[repoName]
					if !ok {
						t.Fatalf("unrecognized repo name: %s", repoName)
					}

					j := testPatch(patchSet.ID, repo.ID, now)
					err := store.CreatePatch(ctx, j)
					if err != nil {
						t.Fatal(err)
					}
					patchesByID[j.ID] = j
					oldPatches = append(oldPatches, j)

					if s, ok := tt.changesetStates[repoName]; ok {
						changesetStateByPatchID[j.ID] = s
					} else {
						changesetStateByPatchID[j.ID] = campaigns.ChangesetStateOpen
					}
				}
				campaign = testCampaign(user.ID, patchSet.ID)
			}

			if tt.campaignIsClosed {
				campaign.ClosedAt = now
			}
			err = svc.CreateCampaign(ctx, campaign)
			if err != nil {
				t.Fatal(err)
			}

			if !tt.campaignWithoutPatchSet {
				for _, p := range patchesByID {
					if err := svc.EnqueueChangesetJobForPatch(ctx, p.ID); err != nil {
						t.Fatal(err)
					}
				}
				// Create Changesets and update ChangesetJobs to look like they ran
				oldChangesets = fakeRunChangesetJobs(ctx, t, store, now, campaign, patchesByID, changesetStateByPatchID)
			}

			if len(tt.individuallyPublished) != 0 {
				toPublish := make(map[int64]*campaigns.Patch)
				for _, name := range tt.individuallyPublished {
					repo, ok := reposByName[name]
					if !ok {
						t.Errorf("unrecognized repo name: %s", name)
					}
					for _, j := range oldPatches {
						if j.RepoID == repo.ID {
							toPublish[j.ID] = j

							err = svc.EnqueueChangesetJobForPatch(ctx, j.ID)
							if err != nil {
								t.Fatalf("Failed to individually created ChangesetJob: %s", err)
							}
						}
					}
				}

				oldChangesets = fakeRunChangesetJobs(ctx, t, store, now, campaign, toPublish, changesetStateByPatchID)
			}

			oldTime := now
			now = now.Add(5 * time.Second)

			newPatchSet := &campaigns.PatchSet{UserID: user.ID}
			err = store.CreatePatchSet(ctx, newPatchSet)
			if err != nil {
				t.Fatal(err)
			}

			for _, spec := range tt.newPatches {
				r, ok := reposByName[spec.repo]
				if !ok {
					t.Fatalf("unrecognized repo name: %s", spec.repo)
				}

				j := testPatch(newPatchSet.ID, r.ID, now)

				if spec.modifiedDiff {
					j.Diff = j.Diff + "-modified"
				}

				if spec.modifiedRev {
					j.Rev = j.Rev + "-modified"
				}

				err := store.CreatePatch(ctx, j)
				if err != nil {
					t.Fatal(err)
				}

				patchesByID[j.ID] = j
			}

			// Filter out the repositories
			toFilter := map[string]struct{}{}
			for _, repoName := range tt.missingRepoPerms {
				toFilter[repoName] = struct{}{}
			}
			db.MockAuthzFilter = func(ctx context.Context, repos []*types.Repo, p authz.Perms) ([]*types.Repo, error) {
				var filtered []*types.Repo
				for _, r := range repos {
					if _, ok := toFilter[string(r.Name)]; ok {
						continue
					}
					filtered = append(filtered, r)
				}
				return filtered, nil
			}
			defer func() { db.MockAuthzFilter = nil }()

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
			if tt.updatePatchSet {
				args.PatchSet = &newPatchSet.ID
			}

			// We ignore the returned campaign here and load it from the
			// database again to make sure the changes are persisted
			_, detachedChangesets, err := svc.UpdateCampaign(ctx, args)

			if tt.wantErr != "" {
				if have, want := fmt.Sprint(err), tt.wantErr; have != want {
					t.Fatalf("error:\nhave: %q\nwant: %q", have, want)
				}
				return
			} else if err != nil {
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

			if args.PatchSet != nil && updatedCampaign.PatchSetID != *args.PatchSet {
				t.Fatalf("campaign PatchSetID not updated. want=%q, have=%q", newPatchSet.ID, updatedCampaign.PatchSetID)
			}

			newChangesetJobs, _, err := store.ListChangesetJobs(ctx, ListChangesetJobsOpts{
				CampaignID: campaign.ID,
				Limit:      -1,
			})
			if err != nil {
				t.Fatal(err)
			}

			var wantChangesetJobLen int
			if tt.updatePatchSet {
				if len(tt.individuallyPublished) != 0 {
					wantChangesetJobLen = len(tt.individuallyPublished)
				} else {
					wantChangesetJobLen = len(tt.wantCreated) + len(tt.wantUnmodified) + len(tt.wantModified)
				}
			} else {
				wantChangesetJobLen = len(oldPatches)
			}
			if len(newChangesetJobs) != wantChangesetJobLen {
				t.Fatalf("wrong number of new ChangesetJobs. want=%d, have=%d", wantChangesetJobLen, len(newChangesetJobs))
			}

			newChangesetJobsByRepo := map[string]*campaigns.ChangesetJob{}
			for _, c := range newChangesetJobs {
				patch, ok := patchesByID[c.PatchID]
				if !ok {
					t.Fatalf("ChangesetJob has invalid PatchID: %+v", c)
				}
				r, ok := reposByID[patch.RepoID]
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
			changesets, _, err := store.ListChangesets(ctx, ListChangesetsOpts{
				CampaignID: campaign.ID,
				IDs:        wantAttachedChangesetIDs,
			})
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
	patchesByID map[int64]*campaigns.Patch,
	changesetStatesByPatchID map[int64]campaigns.ChangesetState,
) []*campaigns.Changeset {
	jobs, _, err := store.ListChangesetJobs(ctx, ListChangesetJobsOpts{
		CampaignID: campaign.ID,
		Limit:      -1,
	})
	if err != nil {
		t.Fatal(err)
	}

	if have, want := len(jobs), len(patchesByID); have != want {
		t.Fatalf("wrong number of changeset jobs. want=%d, have=%d", want, have)
	}

	cs := make([]*campaigns.Changeset, 0, len(patchesByID))
	for _, changesetJob := range jobs {
		patch, ok := patchesByID[changesetJob.PatchID]
		if !ok {
			t.Fatal("no Patch found for ChangesetJob")
		}

		state, ok := changesetStatesByPatchID[patch.ID]
		if !ok {
			t.Fatal("no desired state found for Changeset")
		}
		changeset := testChangeset(patch.RepoID, changesetJob.CampaignID, changesetJob.ID, state)
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

func testPatch(patchSet int64, repo api.RepoID, t time.Time) *campaigns.Patch {
	return &campaigns.Patch{
		PatchSetID: patchSet,
		RepoID:     repo,
		Rev:        "deadbeef",
		BaseRef:    "refs/heads/master",
		Diff:       "cool diff",
	}
}

func testCampaign(user int32, patchSet int64) *campaigns.Campaign {
	c := &campaigns.Campaign{
		Name:            "Testing Campaign",
		Description:     "Testing Campaign",
		AuthorID:        user,
		NamespaceUserID: user,
		PatchSetID:      patchSet,
	}

	if patchSet != 0 {
		c.Branch = "test-branch"
	}

	return c
}

func testChangeset(repoID api.RepoID, campaign int64, changesetJob int64, state campaigns.ChangesetState) *campaigns.Changeset {
	changeset := &campaigns.Changeset{
		RepoID:              repoID,
		ExternalServiceType: extsvc.TypeGitHub,
		ExternalID:          fmt.Sprintf("ext-id-%d", changesetJob),
		Metadata:            &github.PullRequest{State: string(state)},
		ExternalState:       state,
	}

	if campaign != 0 {
		changeset.CampaignIDs = []int64{campaign}
	}

	return changeset
}
