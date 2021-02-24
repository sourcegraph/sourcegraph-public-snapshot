package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/reconciler"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/store"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/testing"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func TestServiceApplyCampaign(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := backend.WithAuthzBypass(context.Background())
	db := dbtesting.GetDB(t)

	admin := ct.CreateTestUser(t, db, true)
	adminCtx := actor.WithActor(context.Background(), actor.FromUser(admin.ID))

	user := ct.CreateTestUser(t, db, false)
	userCtx := actor.WithActor(context.Background(), actor.FromUser(user.ID))

	repos, _ := ct.CreateTestRepos(t, ctx, db, 4)

	now := timeutil.Now()
	clock := func() time.Time { return now }
	store := store.NewWithClock(db, clock)
	svc := New(store)

	t.Run("campaignSpec without changesetSpecs", func(t *testing.T) {
		t.Run("new campaign", func(t *testing.T) {
			ct.TruncateTables(t, db, "changeset_events", "changesets", "campaigns", "campaign_specs", "changeset_specs")
			campaignSpec := ct.CreateCampaignSpec(t, ctx, store, "campaign1", admin.ID)
			campaign, err := svc.ApplyCampaign(adminCtx, ApplyCampaignOpts{
				CampaignSpecRandID: campaignSpec.RandID,
			})
			if err != nil {
				t.Fatal(err)
			}

			if campaign.ID == 0 {
				t.Fatalf("campaign ID is 0")
			}

			want := &campaigns.Campaign{
				Name:             campaignSpec.Spec.Name,
				Description:      campaignSpec.Spec.Description,
				InitialApplierID: admin.ID,
				LastApplierID:    admin.ID,
				LastAppliedAt:    now,
				NamespaceUserID:  campaignSpec.NamespaceUserID,
				CampaignSpecID:   campaignSpec.ID,

				// Ignore these fields
				ID:        campaign.ID,
				UpdatedAt: campaign.UpdatedAt,
				CreatedAt: campaign.CreatedAt,
			}

			if diff := cmp.Diff(want, campaign); diff != "" {
				t.Fatalf("wrong spec fields (-want +got):\n%s", diff)
			}
		})

		t.Run("existing campaign", func(t *testing.T) {
			ct.TruncateTables(t, db, "changeset_events", "changesets", "campaigns", "campaign_specs", "changeset_specs")
			campaignSpec := ct.CreateCampaignSpec(t, ctx, store, "campaign2", admin.ID)
			campaign := ct.CreateCampaign(t, ctx, store, "campaign2", admin.ID, campaignSpec.ID)

			t.Run("apply same campaignSpec", func(t *testing.T) {
				campaign2, err := svc.ApplyCampaign(adminCtx, ApplyCampaignOpts{
					CampaignSpecRandID: campaignSpec.RandID,
				})
				if err != nil {
					t.Fatal(err)
				}

				if have, want := campaign2.ID, campaign.ID; have != want {
					t.Fatalf("campaign ID is wrong. want=%d, have=%d", want, have)
				}
			})

			t.Run("apply same campaignSpec with FailIfExists", func(t *testing.T) {
				_, err := svc.ApplyCampaign(ctx, ApplyCampaignOpts{
					CampaignSpecRandID:   campaignSpec.RandID,
					FailIfCampaignExists: true,
				})
				if err != ErrMatchingCampaignExists {
					t.Fatalf("unexpected error. want=%s, got=%s", ErrMatchingCampaignExists, err)
				}
			})

			t.Run("apply campaign spec with same name", func(t *testing.T) {
				campaignSpec2 := ct.CreateCampaignSpec(t, ctx, store, "campaign2", admin.ID)
				campaign2, err := svc.ApplyCampaign(adminCtx, ApplyCampaignOpts{
					CampaignSpecRandID: campaignSpec2.RandID,
				})
				if err != nil {
					t.Fatal(err)
				}

				if have, want := campaign2.ID, campaign.ID; have != want {
					t.Fatalf("campaign ID is wrong. want=%d, have=%d", want, have)
				}
			})

			t.Run("apply campaign spec with same name but different current user", func(t *testing.T) {
				campaignSpec := ct.CreateCampaignSpec(t, ctx, store, "created-by-user", user.ID)
				campaign := ct.CreateCampaign(t, ctx, store, "created-by-user", user.ID, campaignSpec.ID)

				if have, want := campaign.InitialApplierID, user.ID; have != want {
					t.Fatalf("campaign InitialApplierID is wrong. want=%d, have=%d", want, have)
				}

				if have, want := campaign.LastApplierID, user.ID; have != want {
					t.Fatalf("campaign LastApplierID is wrong. want=%d, have=%d", want, have)
				}

				campaignSpec2 := ct.CreateCampaignSpec(t, ctx, store, "created-by-user", user.ID)
				campaign2, err := svc.ApplyCampaign(adminCtx, ApplyCampaignOpts{
					CampaignSpecRandID: campaignSpec2.RandID,
				})
				if err != nil {
					t.Fatal(err)
				}

				if have, want := campaign2.ID, campaign.ID; have != want {
					t.Fatalf("campaign ID is wrong. want=%d, have=%d", want, have)
				}

				if have, want := campaign2.InitialApplierID, campaign.InitialApplierID; have != want {
					t.Fatalf("campaign InitialApplierID is wrong. want=%d, have=%d", want, have)
				}

				if have, want := campaign2.LastApplierID, admin.ID; have != want {
					t.Fatalf("campaign LastApplierID is wrong. want=%d, have=%d", want, have)
				}
			})

			t.Run("apply campaign spec with same name but different namespace", func(t *testing.T) {
				user2 := ct.CreateTestUser(t, db, false)
				campaignSpec2 := ct.CreateCampaignSpec(t, ctx, store, "campaign2", user2.ID)

				campaign2, err := svc.ApplyCampaign(adminCtx, ApplyCampaignOpts{
					CampaignSpecRandID: campaignSpec2.RandID,
				})
				if err != nil {
					t.Fatal(err)
				}

				if campaign2.ID == 0 {
					t.Fatalf("campaign2 ID is 0")
				}

				if campaign2.ID == campaign.ID {
					t.Fatalf("campaign IDs are the same, but want different")
				}
			})

			t.Run("campaign spec with same name and same ensureCampaignID", func(t *testing.T) {
				campaignSpec2 := ct.CreateCampaignSpec(t, ctx, store, "campaign2", admin.ID)

				campaign2, err := svc.ApplyCampaign(adminCtx, ApplyCampaignOpts{
					CampaignSpecRandID: campaignSpec2.RandID,
					EnsureCampaignID:   campaign.ID,
				})
				if err != nil {
					t.Fatal(err)
				}
				if have, want := campaign2.ID, campaign.ID; have != want {
					t.Fatalf("campaign has wrong ID. want=%d, have=%d", want, have)
				}
			})

			t.Run("campaign spec with same name but different ensureCampaignID", func(t *testing.T) {
				campaignSpec2 := ct.CreateCampaignSpec(t, ctx, store, "campaign2", admin.ID)

				_, err := svc.ApplyCampaign(adminCtx, ApplyCampaignOpts{
					CampaignSpecRandID: campaignSpec2.RandID,
					EnsureCampaignID:   campaign.ID + 999,
				})
				if err != ErrEnsureCampaignFailed {
					t.Fatalf("wrong error: %s", err)
				}
			})
		})
	})

	// These tests focus on changesetSpecs and wiring them up with changesets.
	// The applying/re-applying of a campaignSpec to an existing campaign is
	// covered in the tests above.
	t.Run("campaignSpec with changesetSpecs", func(t *testing.T) {
		t.Run("new campaign", func(t *testing.T) {
			ct.TruncateTables(t, db, "changeset_events", "changesets", "campaigns", "campaign_specs", "changeset_specs")
			campaignSpec := ct.CreateCampaignSpec(t, ctx, store, "campaign3", admin.ID)

			spec1 := ct.CreateChangesetSpec(t, ctx, store, ct.TestSpecOpts{
				User:         admin.ID,
				Repo:         repos[0].ID,
				CampaignSpec: campaignSpec.ID,
				ExternalID:   "1234",
			})

			spec2 := ct.CreateChangesetSpec(t, ctx, store, ct.TestSpecOpts{
				User:         admin.ID,
				Repo:         repos[1].ID,
				CampaignSpec: campaignSpec.ID,
				HeadRef:      "refs/heads/my-branch",
			})

			campaign, cs := applyAndListChangesets(adminCtx, t, svc, campaignSpec.RandID, 2)

			if have, want := campaign.Name, "campaign3"; have != want {
				t.Fatalf("wrong campaign name. want=%s, have=%s", want, have)
			}

			c1 := cs.Find(campaigns.WithExternalID(spec1.Spec.ExternalID))
			ct.AssertChangeset(t, c1, ct.ChangesetAssertions{
				Repo:             spec1.RepoID,
				ExternalID:       "1234",
				ReconcilerState:  campaigns.ReconcilerStateQueued,
				PublicationState: campaigns.ChangesetPublicationStateUnpublished,
				AttachedTo:       []int64{campaign.ID},
			})

			c2 := cs.Find(campaigns.WithCurrentSpecID(spec2.ID))
			ct.AssertChangeset(t, c2, ct.ChangesetAssertions{
				Repo:             spec2.RepoID,
				CurrentSpec:      spec2.ID,
				OwnedByCampaign:  campaign.ID,
				ReconcilerState:  campaigns.ReconcilerStateQueued,
				PublicationState: campaigns.ChangesetPublicationStateUnpublished,
				DiffStat:         ct.TestChangsetSpecDiffStat,
				AttachedTo:       []int64{campaign.ID},
			})
		})

		t.Run("campaign with changesets", func(t *testing.T) {
			ct.TruncateTables(t, db, "changeset_events", "changesets", "campaigns", "campaign_specs", "changeset_specs")
			// First we create a campaignSpec and apply it, so that we have
			// changesets and changesetSpecs in the database, wired up
			// correctly.
			campaignSpec1 := ct.CreateCampaignSpec(t, ctx, store, "campaign4", admin.ID)

			ct.CreateChangesetSpec(t, ctx, store, ct.TestSpecOpts{
				User:         admin.ID,
				Repo:         repos[0].ID,
				CampaignSpec: campaignSpec1.ID,
				ExternalID:   "1234",
			})

			ct.CreateChangesetSpec(t, ctx, store, ct.TestSpecOpts{
				User:         admin.ID,
				Repo:         repos[0].ID,
				CampaignSpec: campaignSpec1.ID,
				ExternalID:   "5678",
			})

			oldSpec3 := ct.CreateChangesetSpec(t, ctx, store, ct.TestSpecOpts{
				User:         admin.ID,
				Repo:         repos[1].ID,
				CampaignSpec: campaignSpec1.ID,
				HeadRef:      "refs/heads/repo-1-branch-1",
			})

			oldSpec4 := ct.CreateChangesetSpec(t, ctx, store, ct.TestSpecOpts{
				User:         admin.ID,
				Repo:         repos[2].ID,
				CampaignSpec: campaignSpec1.ID,
				HeadRef:      "refs/heads/repo-2-branch-1",
			})

			// Apply and expect 4 changesets
			oldCampaign, oldChangesets := applyAndListChangesets(adminCtx, t, svc, campaignSpec1.RandID, 4)

			// Now we create another campaign spec with the same campaign name
			// and namespace.
			campaignSpec2 := ct.CreateCampaignSpec(t, ctx, store, "campaign4", admin.ID)

			// Same
			spec1 := ct.CreateChangesetSpec(t, ctx, store, ct.TestSpecOpts{
				User:         admin.ID,
				Repo:         repos[0].ID,
				CampaignSpec: campaignSpec2.ID,
				ExternalID:   "1234",
			})

			// DIFFERENT: Track #9999 in repo[0]
			spec2 := ct.CreateChangesetSpec(t, ctx, store, ct.TestSpecOpts{
				User:         admin.ID,
				Repo:         repos[0].ID,
				CampaignSpec: campaignSpec2.ID,
				ExternalID:   "5678",
			})

			// Same
			spec3 := ct.CreateChangesetSpec(t, ctx, store, ct.TestSpecOpts{
				User:         admin.ID,
				Repo:         repos[1].ID,
				CampaignSpec: campaignSpec2.ID,
				HeadRef:      "refs/heads/repo-1-branch-1",
			})

			// DIFFERENT: branch changed in repo[2]
			spec4 := ct.CreateChangesetSpec(t, ctx, store, ct.TestSpecOpts{
				User:         admin.ID,
				Repo:         repos[2].ID,
				CampaignSpec: campaignSpec2.ID,
				HeadRef:      "refs/heads/repo-2-branch-2",
			})

			// NEW: repo[3]
			spec5 := ct.CreateChangesetSpec(t, ctx, store, ct.TestSpecOpts{
				User:         admin.ID,
				Repo:         repos[3].ID,
				CampaignSpec: campaignSpec2.ID,
				HeadRef:      "refs/heads/repo-3-branch-1",
			})

			// Before we apply the new campaign spec, we make the changeset we
			// expect to be closed to look "published", otherwise it won't be
			// closed.
			wantClosed := oldChangesets.Find(campaigns.WithCurrentSpecID(oldSpec4.ID))
			ct.SetChangesetPublished(t, ctx, store, wantClosed, "98765", oldSpec4.Spec.HeadRef)

			changeset3 := oldChangesets.Find(campaigns.WithCurrentSpecID(oldSpec3.ID))
			ct.SetChangesetPublished(t, ctx, store, changeset3, "12345", oldSpec3.Spec.HeadRef)

			// Apply and expect 6 changesets
			campaign, cs := applyAndListChangesets(adminCtx, t, svc, campaignSpec2.RandID, 6)

			if oldCampaign.ID != campaign.ID {
				t.Fatal("expected to update campaign, but got a new one")
			}

			// This changeset we want marked as "to be closed"
			ct.ReloadAndAssertChangeset(t, ctx, store, wantClosed, ct.ChangesetAssertions{
				Repo:         repos[2].ID,
				CurrentSpec:  oldSpec4.ID,
				PreviousSpec: oldSpec4.ID,
				ExternalID:   wantClosed.ExternalID,
				// It's still open, just _marked as to be closed_.
				ExternalState:    campaigns.ChangesetExternalStateOpen,
				ExternalBranch:   wantClosed.ExternalBranch,
				OwnedByCampaign:  campaign.ID,
				ReconcilerState:  campaigns.ReconcilerStateQueued,
				PublicationState: campaigns.ChangesetPublicationStatePublished,
				DiffStat:         ct.TestChangsetSpecDiffStat,
				DetachFrom:       []int64{campaign.ID},
				Closing:          true,
			})

			c1 := cs.Find(campaigns.WithExternalID(spec1.Spec.ExternalID))
			ct.AssertChangeset(t, c1, ct.ChangesetAssertions{
				Repo:             repos[0].ID,
				CurrentSpec:      0,
				PreviousSpec:     0,
				ExternalID:       "1234",
				ReconcilerState:  campaigns.ReconcilerStateQueued,
				PublicationState: campaigns.ChangesetPublicationStateUnpublished,
				AttachedTo:       []int64{campaign.ID},
			})

			c2 := cs.Find(campaigns.WithExternalID(spec2.Spec.ExternalID))
			ct.AssertChangeset(t, c2, ct.ChangesetAssertions{
				Repo:             repos[0].ID,
				CurrentSpec:      0,
				PreviousSpec:     0,
				ExternalID:       "5678",
				ReconcilerState:  campaigns.ReconcilerStateQueued,
				PublicationState: campaigns.ChangesetPublicationStateUnpublished,
				AttachedTo:       []int64{campaign.ID},
			})

			c3 := cs.Find(campaigns.WithCurrentSpecID(spec3.ID))
			ct.AssertChangeset(t, c3, ct.ChangesetAssertions{
				Repo:           repos[1].ID,
				CurrentSpec:    spec3.ID,
				ExternalID:     changeset3.ExternalID,
				ExternalBranch: changeset3.ExternalBranch,
				ExternalState:  campaigns.ChangesetExternalStateOpen,
				// Has a previous spec, because it succeeded publishing.
				PreviousSpec:     oldSpec3.ID,
				OwnedByCampaign:  campaign.ID,
				ReconcilerState:  campaigns.ReconcilerStateQueued,
				PublicationState: campaigns.ChangesetPublicationStatePublished,
				DiffStat:         ct.TestChangsetSpecDiffStat,
				AttachedTo:       []int64{campaign.ID},
			})

			c4 := cs.Find(campaigns.WithCurrentSpecID(spec4.ID))
			ct.AssertChangeset(t, c4, ct.ChangesetAssertions{
				Repo:             repos[2].ID,
				CurrentSpec:      spec4.ID,
				OwnedByCampaign:  campaign.ID,
				ReconcilerState:  campaigns.ReconcilerStateQueued,
				PublicationState: campaigns.ChangesetPublicationStateUnpublished,
				DiffStat:         ct.TestChangsetSpecDiffStat,
				AttachedTo:       []int64{campaign.ID},
			})

			c5 := cs.Find(campaigns.WithCurrentSpecID(spec5.ID))
			ct.AssertChangeset(t, c5, ct.ChangesetAssertions{
				Repo:             repos[3].ID,
				CurrentSpec:      spec5.ID,
				OwnedByCampaign:  campaign.ID,
				ReconcilerState:  campaigns.ReconcilerStateQueued,
				PublicationState: campaigns.ChangesetPublicationStateUnpublished,
				DiffStat:         ct.TestChangsetSpecDiffStat,
				AttachedTo:       []int64{campaign.ID},
			})
		})

		t.Run("campaign tracking changesets owned by another campaign", func(t *testing.T) {
			ct.TruncateTables(t, db, "changeset_events", "changesets", "campaigns", "campaign_specs", "changeset_specs")
			campaignSpec1 := ct.CreateCampaignSpec(t, ctx, store, "owner-campaign", admin.ID)

			oldSpec1 := ct.CreateChangesetSpec(t, ctx, store, ct.TestSpecOpts{
				User:         admin.ID,
				Repo:         repos[0].ID,
				CampaignSpec: campaignSpec1.ID,
				HeadRef:      "refs/heads/repo-0-branch-0",
			})

			ownerCampaign, ownerChangesets := applyAndListChangesets(adminCtx, t, svc, campaignSpec1.RandID, 1)

			// Now we update the changeset so it looks like it's been published
			// on the code host.
			c := ownerChangesets[0]
			ct.SetChangesetPublished(t, ctx, store, c, "88888", "refs/heads/repo-0-branch-0")

			// This other campaign tracks the changeset created by the first one
			campaignSpec2 := ct.CreateCampaignSpec(t, ctx, store, "tracking-campaign", admin.ID)
			ct.CreateChangesetSpec(t, ctx, store, ct.TestSpecOpts{
				User:         admin.ID,
				Repo:         c.RepoID,
				CampaignSpec: campaignSpec2.ID,
				ExternalID:   c.ExternalID,
			})

			trackingCampaign, trackedChangesets := applyAndListChangesets(adminCtx, t, svc, campaignSpec2.RandID, 1)
			// This should still point to the owner campaign
			c2 := trackedChangesets[0]
			trackedChangesetAssertions := ct.ChangesetAssertions{
				Repo:             c.RepoID,
				CurrentSpec:      oldSpec1.ID,
				OwnedByCampaign:  ownerCampaign.ID,
				ExternalBranch:   c.ExternalBranch,
				ExternalID:       c.ExternalID,
				ExternalState:    campaigns.ChangesetExternalStateOpen,
				ReconcilerState:  campaigns.ReconcilerStateCompleted,
				PublicationState: campaigns.ChangesetPublicationStatePublished,
				DiffStat:         ct.TestChangsetSpecDiffStat,
				AttachedTo:       []int64{ownerCampaign.ID, trackingCampaign.ID},
			}
			ct.AssertChangeset(t, c2, trackedChangesetAssertions)

			// Now try to apply a new spec that wants to modify the formerly tracked changeset.
			campaignSpec3 := ct.CreateCampaignSpec(t, ctx, store, "tracking-campaign", admin.ID)

			spec3 := ct.CreateChangesetSpec(t, ctx, store, ct.TestSpecOpts{
				User:         admin.ID,
				Repo:         repos[0].ID,
				CampaignSpec: campaignSpec3.ID,
				HeadRef:      "refs/heads/repo-0-branch-0",
			})
			// Apply again. This should have flagged the association as detach and it should not be closed, since the campaign is
			// not the owner.
			trackingCampaign, cs := applyAndListChangesets(adminCtx, t, svc, campaignSpec3.RandID, 2)

			trackedChangesetAssertions.Closing = false
			trackedChangesetAssertions.ReconcilerState = campaigns.ReconcilerStateQueued
			trackedChangesetAssertions.DetachFrom = []int64{trackingCampaign.ID}
			trackedChangesetAssertions.AttachedTo = []int64{ownerCampaign.ID}
			ct.ReloadAndAssertChangeset(t, ctx, store, c2, trackedChangesetAssertions)

			// But we do want to have a new changeset record that is going to create a new changeset on the code host.
			ct.ReloadAndAssertChangeset(t, ctx, store, cs[1], ct.ChangesetAssertions{
				Repo:             spec3.RepoID,
				CurrentSpec:      spec3.ID,
				OwnedByCampaign:  trackingCampaign.ID,
				ReconcilerState:  campaigns.ReconcilerStateQueued,
				PublicationState: campaigns.ChangesetPublicationStateUnpublished,
				DiffStat:         ct.TestChangsetSpecDiffStat,
				AttachedTo:       []int64{trackingCampaign.ID},
			})
		})

		t.Run("campaign with changeset that is unpublished", func(t *testing.T) {
			ct.TruncateTables(t, db, "changeset_events", "changesets", "campaigns", "campaign_specs", "changeset_specs")
			campaignSpec1 := ct.CreateCampaignSpec(t, ctx, store, "unpublished-changesets", admin.ID)

			ct.CreateChangesetSpec(t, ctx, store, ct.TestSpecOpts{
				User:         admin.ID,
				Repo:         repos[3].ID,
				CampaignSpec: campaignSpec1.ID,
				HeadRef:      "refs/heads/never-published",
			})

			// We apply the spec and expect 1 changeset
			applyAndListChangesets(adminCtx, t, svc, campaignSpec1.RandID, 1)

			// But the changeset was not published yet.
			// And now we apply a new spec without any changesets.
			campaignSpec2 := ct.CreateCampaignSpec(t, ctx, store, "unpublished-changesets", admin.ID)

			// That should close no changesets, but set the unpublished changesets to be detached when
			// the reconciler picks them up.
			applyAndListChangesets(adminCtx, t, svc, campaignSpec2.RandID, 1)
		})

		t.Run("campaign with changeset that wasn't processed before reapply", func(t *testing.T) {
			ct.TruncateTables(t, db, "changeset_events", "changesets", "campaigns", "campaign_specs", "changeset_specs")
			campaignSpec1 := ct.CreateCampaignSpec(t, ctx, store, "queued-changesets", admin.ID)

			specOpts := ct.TestSpecOpts{
				User:         admin.ID,
				Repo:         repos[3].ID,
				CampaignSpec: campaignSpec1.ID,
				Title:        "Spec1",
				HeadRef:      "refs/heads/queued",
				Published:    true,
			}
			spec1 := ct.CreateChangesetSpec(t, ctx, store, specOpts)

			// We apply the spec and expect 1 changeset
			campaign, changesets := applyAndListChangesets(adminCtx, t, svc, campaignSpec1.RandID, 1)

			// And publish it.
			ct.SetChangesetPublished(t, ctx, store, changesets[0], "123-queued", "refs/heads/queued")

			ct.ReloadAndAssertChangeset(t, ctx, store, changesets[0], ct.ChangesetAssertions{
				ReconcilerState:  campaigns.ReconcilerStateCompleted,
				PublicationState: campaigns.ChangesetPublicationStatePublished,
				ExternalBranch:   "refs/heads/queued",
				ExternalID:       "123-queued",
				ExternalState:    campaigns.ChangesetExternalStateOpen,
				Repo:             repos[3].ID,
				CurrentSpec:      spec1.ID,
				OwnedByCampaign:  campaign.ID,
				DiffStat:         ct.TestChangsetSpecDiffStat,
				AttachedTo:       []int64{campaign.ID},
			})

			// Apply again so that an update to the changeset is pending.
			campaignSpec2 := ct.CreateCampaignSpec(t, ctx, store, "queued-changesets", admin.ID)

			specOpts.CampaignSpec = campaignSpec2.ID
			specOpts.Title = "Spec2"
			spec2 := ct.CreateChangesetSpec(t, ctx, store, specOpts)

			// That should still want to publish the changeset
			_, changesets = applyAndListChangesets(adminCtx, t, svc, campaignSpec2.RandID, 1)

			ct.ReloadAndAssertChangeset(t, ctx, store, changesets[0], ct.ChangesetAssertions{
				ReconcilerState:  campaigns.ReconcilerStateQueued,
				PublicationState: campaigns.ChangesetPublicationStatePublished,
				ExternalBranch:   "refs/heads/queued",
				ExternalID:       "123-queued",
				ExternalState:    campaigns.ChangesetExternalStateOpen,
				Repo:             repos[3].ID,
				CurrentSpec:      spec2.ID,
				// Track the previous spec.
				PreviousSpec:    spec1.ID,
				OwnedByCampaign: campaign.ID,
				DiffStat:        ct.TestChangsetSpecDiffStat,
				AttachedTo:      []int64{campaign.ID},
			})

			// Make sure the reconciler wants to update this changeset.
			plan, err := reconciler.DeterminePlan(
				// changesets[0].PreviousSpecID
				spec1,
				// changesets[0].CurrentSpecID
				spec2,
				changesets[0],
			)
			if err != nil {
				t.Fatal(err)
			}
			if !plan.Ops.Equal(reconciler.Operations{campaigns.ReconcilerOperationUpdate}) {
				t.Fatalf("Got invalid reconciler operations: %q", plan.Ops.String())
			}

			// And now we apply a new spec before the reconciler could process the changeset.
			campaignSpec3 := ct.CreateCampaignSpec(t, ctx, store, "queued-changesets", admin.ID)

			// No change this time, just reapplying.
			specOpts.CampaignSpec = campaignSpec3.ID
			spec3 := ct.CreateChangesetSpec(t, ctx, store, specOpts)

			_, changesets = applyAndListChangesets(adminCtx, t, svc, campaignSpec3.RandID, 1)

			ct.ReloadAndAssertChangeset(t, ctx, store, changesets[0], ct.ChangesetAssertions{
				ReconcilerState:  campaigns.ReconcilerStateQueued,
				PublicationState: campaigns.ChangesetPublicationStatePublished,
				ExternalBranch:   "refs/heads/queued",
				ExternalID:       "123-queued",
				ExternalState:    campaigns.ChangesetExternalStateOpen,
				Repo:             repos[3].ID,
				CurrentSpec:      spec3.ID,
				// Still be pointing at the first spec, since the second was never applied.
				PreviousSpec:    spec1.ID,
				OwnedByCampaign: campaign.ID,
				DiffStat:        ct.TestChangsetSpecDiffStat,
				AttachedTo:      []int64{campaign.ID},
			})

			// Make sure the reconciler would still update this changeset.
			plan, err = reconciler.DeterminePlan(
				// changesets[0].PreviousSpecID
				spec1,
				// changesets[0].CurrentSpecID
				spec3,
				changesets[0],
			)
			if err != nil {
				t.Fatal(err)
			}
			if !plan.Ops.Equal(reconciler.Operations{campaigns.ReconcilerOperationUpdate}) {
				t.Fatalf("Got invalid reconciler operations: %q", plan.Ops.String())
			}

			// Now test that it still updates when this update failed.
			ct.SetChangesetFailed(t, ctx, store, changesets[0])

			campaignSpec4 := ct.CreateCampaignSpec(t, ctx, store, "queued-changesets", admin.ID)

			// No change this time, just reapplying.
			specOpts.CampaignSpec = campaignSpec4.ID
			spec4 := ct.CreateChangesetSpec(t, ctx, store, specOpts)

			_, changesets = applyAndListChangesets(adminCtx, t, svc, campaignSpec4.RandID, 1)

			ct.ReloadAndAssertChangeset(t, ctx, store, changesets[0], ct.ChangesetAssertions{
				ReconcilerState:  campaigns.ReconcilerStateQueued,
				PublicationState: campaigns.ChangesetPublicationStatePublished,
				ExternalBranch:   "refs/heads/queued",
				ExternalID:       "123-queued",
				ExternalState:    campaigns.ChangesetExternalStateOpen,
				Repo:             repos[3].ID,
				CurrentSpec:      spec4.ID,
				// Still be pointing at the first spec, since the second and third were never applied.
				PreviousSpec:    spec1.ID,
				OwnedByCampaign: campaign.ID,
				DiffStat:        ct.TestChangsetSpecDiffStat,
				AttachedTo:      []int64{campaign.ID},
			})

			// Make sure the reconciler would still update this changeset.
			plan, err = reconciler.DeterminePlan(
				// changesets[0].PreviousSpecID
				spec1,
				// changesets[0].CurrentSpecID
				spec4,
				changesets[0],
			)
			if err != nil {
				t.Fatal(err)
			}
			if !plan.Ops.Equal(reconciler.Operations{campaigns.ReconcilerOperationUpdate}) {
				t.Fatalf("Got invalid reconciler operations: %q", plan.Ops.String())
			}
		})

		t.Run("missing repository permissions", func(t *testing.T) {
			ct.TruncateTables(t, db, "changeset_events", "changesets", "campaigns", "campaign_specs", "changeset_specs")
			ct.MockRepoPermissions(t, db, user.ID, repos[0].ID, repos[2].ID, repos[3].ID)

			// NOTE: We cannot use a context that has authz bypassed.
			campaignSpec := ct.CreateCampaignSpec(t, userCtx, store, "missing-permissions", user.ID)

			ct.CreateChangesetSpec(t, userCtx, store, ct.TestSpecOpts{
				User:         user.ID,
				Repo:         repos[0].ID,
				CampaignSpec: campaignSpec.ID,
				ExternalID:   "1234",
			})

			ct.CreateChangesetSpec(t, userCtx, store, ct.TestSpecOpts{
				User:         user.ID,
				Repo:         repos[1].ID, // Not authorized to access this repository
				CampaignSpec: campaignSpec.ID,
				HeadRef:      "refs/heads/my-branch",
			})

			_, err := svc.ApplyCampaign(userCtx, ApplyCampaignOpts{
				CampaignSpecRandID: campaignSpec.RandID,
			})
			if err == nil {
				t.Fatal("expected error, but got none")
			}
			notFoundErr, ok := err.(*database.RepoNotFoundErr)
			if !ok {
				t.Fatalf("expected RepoNotFoundErr but got: %s", err)
			}
			if notFoundErr.ID != repos[1].ID {
				t.Fatalf("wrong repository ID in RepoNotFoundErr: %d", notFoundErr.ID)
			}
		})

		t.Run("campaign with errored changeset", func(t *testing.T) {
			ct.TruncateTables(t, db, "changeset_events", "changesets", "campaigns", "campaign_specs", "changeset_specs")
			campaignSpec1 := ct.CreateCampaignSpec(t, ctx, store, "errored-changeset-campaign", admin.ID)

			spec1Opts := ct.TestSpecOpts{
				User:         admin.ID,
				Repo:         repos[0].ID,
				CampaignSpec: campaignSpec1.ID,
				ExternalID:   "1234",
				Published:    true,
			}
			ct.CreateChangesetSpec(t, ctx, store, spec1Opts)

			spec2Opts := ct.TestSpecOpts{
				User:         admin.ID,
				Repo:         repos[1].ID,
				CampaignSpec: campaignSpec1.ID,
				HeadRef:      "refs/heads/repo-1-branch-1",
				Published:    true,
			}
			ct.CreateChangesetSpec(t, ctx, store, spec2Opts)

			_, oldChangesets := applyAndListChangesets(adminCtx, t, svc, campaignSpec1.RandID, 2)

			// Set the changesets to look like they failed in the reconciler
			for _, c := range oldChangesets {
				ct.SetChangesetFailed(t, ctx, store, c)
			}

			// Now we create another campaign spec with the same campaign name
			// and namespace.
			campaignSpec2 := ct.CreateCampaignSpec(t, ctx, store, "errored-changeset-campaign", admin.ID)
			spec1Opts.CampaignSpec = campaignSpec2.ID
			newSpec1 := ct.CreateChangesetSpec(t, ctx, store, spec1Opts)
			spec2Opts.CampaignSpec = campaignSpec2.ID
			newSpec2 := ct.CreateChangesetSpec(t, ctx, store, spec2Opts)

			campaign, cs := applyAndListChangesets(adminCtx, t, svc, campaignSpec2.RandID, 2)

			c1 := cs.Find(campaigns.WithExternalID(newSpec1.Spec.ExternalID))
			ct.ReloadAndAssertChangeset(t, ctx, store, c1, ct.ChangesetAssertions{
				Repo:             spec1Opts.Repo,
				ExternalID:       "1234",
				PublicationState: campaigns.ChangesetPublicationStateUnpublished,
				AttachedTo:       []int64{campaign.ID},

				ReconcilerState: campaigns.ReconcilerStateQueued,
				FailureMessage:  nil,
				NumFailures:     0,
			})

			c2 := cs.Find(campaigns.WithCurrentSpecID(newSpec2.ID))
			ct.AssertChangeset(t, c2, ct.ChangesetAssertions{
				Repo:        newSpec2.RepoID,
				CurrentSpec: newSpec2.ID,
				// An errored changeset doesn't get the specs rotated, to prevent https://github.com/sourcegraph/sourcegraph/issues/16041.
				PreviousSpec:     0,
				OwnedByCampaign:  campaign.ID,
				PublicationState: campaigns.ChangesetPublicationStateUnpublished,
				DiffStat:         ct.TestChangsetSpecDiffStat,
				AttachedTo:       []int64{campaign.ID},

				ReconcilerState: campaigns.ReconcilerStateQueued,
				FailureMessage:  nil,
				NumFailures:     0,
			})

			// Make sure the reconciler would still publish this changeset.
			plan, err := reconciler.DeterminePlan(
				// c2.previousSpec is 0
				nil,
				// c2.currentSpec is newSpec2
				newSpec2,
				c2,
			)
			if err != nil {
				t.Fatal(err)
			}
			if !plan.Ops.Equal(reconciler.Operations{campaigns.ReconcilerOperationPush, campaigns.ReconcilerOperationPublish}) {
				t.Fatalf("Got invalid reconciler operations: %q", plan.Ops.String())
			}
		})

		t.Run("closed and detached changeset not re-enqueued for close", func(t *testing.T) {
			ct.TruncateTables(t, db, "changeset_events", "changesets", "campaigns", "campaign_specs", "changeset_specs")
			campaignSpec1 := ct.CreateCampaignSpec(t, ctx, store, "detached-closed-changeset", admin.ID)

			specOpts := ct.TestSpecOpts{
				User:         admin.ID,
				Repo:         repos[0].ID,
				CampaignSpec: campaignSpec1.ID,
				HeadRef:      "refs/heads/detached-closed",
			}
			spec1 := ct.CreateChangesetSpec(t, ctx, store, specOpts)

			// STEP 1: We apply the spec and expect 1 changeset.
			campaign, changesets := applyAndListChangesets(adminCtx, t, svc, campaignSpec1.RandID, 1)

			// Now we update the changeset so it looks like it's been published
			// on the code host.
			c := changesets[0]
			ct.SetChangesetPublished(t, ctx, store, c, "995544", specOpts.HeadRef)

			assertions := ct.ChangesetAssertions{
				Repo:             c.RepoID,
				CurrentSpec:      spec1.ID,
				ExternalID:       c.ExternalID,
				ExternalBranch:   c.ExternalBranch,
				ExternalState:    campaigns.ChangesetExternalStateOpen,
				OwnedByCampaign:  campaign.ID,
				ReconcilerState:  campaigns.ReconcilerStateCompleted,
				PublicationState: campaigns.ChangesetPublicationStatePublished,
				DiffStat:         ct.TestChangsetSpecDiffStat,
				AttachedTo:       []int64{campaign.ID},
			}
			c = ct.ReloadAndAssertChangeset(t, ctx, store, c, assertions)

			// STEP 2: Now we apply a new spec without any changesets, but expect the changeset-to-be-detached to
			// be left in the campaign (the reconciler would detach it, if the executor picked up the changeset).
			campaignSpec2 := ct.CreateCampaignSpec(t, ctx, store, "detached-closed-changeset", admin.ID)
			applyAndListChangesets(adminCtx, t, svc, campaignSpec2.RandID, 1)

			// Our previously published changeset should be marked as "to be closed"
			assertions.Closing = true
			assertions.ReconcilerState = campaigns.ReconcilerStateQueued
			// And the previous spec is recorded, because the previous run finished with reconcilerState completed.
			assertions.PreviousSpec = spec1.ID
			assertions.DetachFrom = []int64{campaign.ID}
			assertions.AttachedTo = []int64{}
			c = ct.ReloadAndAssertChangeset(t, ctx, store, c, assertions)

			// Now we update the changeset to make it look closed.
			ct.SetChangesetClosed(t, ctx, store, c)
			assertions.Closing = false
			assertions.DetachFrom = []int64{}
			assertions.ReconcilerState = campaigns.ReconcilerStateCompleted
			assertions.ExternalState = campaigns.ChangesetExternalStateClosed
			c = ct.ReloadAndAssertChangeset(t, ctx, store, c, assertions)

			// STEP 3: We apply a new campaign spec and expect that the detached changeset record is not re-enqueued.
			campaignSpec3 := ct.CreateCampaignSpec(t, ctx, store, "detached-closed-changeset", admin.ID)

			applyAndListChangesets(adminCtx, t, svc, campaignSpec3.RandID, 0)

			// Assert that the changeset record is still completed and closed.
			ct.ReloadAndAssertChangeset(t, ctx, store, c, assertions)
		})

		t.Run("campaign with changeset that is detached and reattached", func(t *testing.T) {
			t.Run("changeset has been closed before re-attaching", func(t *testing.T) {
				ct.TruncateTables(t, db, "changeset_events", "changesets", "campaigns", "campaign_specs", "changeset_specs")
				campaignSpec1 := ct.CreateCampaignSpec(t, ctx, store, "detach-reattach-changeset", admin.ID)

				specOpts := ct.TestSpecOpts{
					User:         admin.ID,
					Repo:         repos[0].ID,
					CampaignSpec: campaignSpec1.ID,
					HeadRef:      "refs/heads/detached-reattached",
				}
				spec1 := ct.CreateChangesetSpec(t, ctx, store, specOpts)

				// STEP 1: We apply the spec and expect 1 changeset.
				campaign, changesets := applyAndListChangesets(adminCtx, t, svc, campaignSpec1.RandID, 1)

				// Now we update the changeset so it looks like it's been published
				// on the code host.
				c := changesets[0]
				ct.SetChangesetPublished(t, ctx, store, c, "995533", specOpts.HeadRef)

				assertions := ct.ChangesetAssertions{
					Repo:             c.RepoID,
					CurrentSpec:      spec1.ID,
					ExternalID:       c.ExternalID,
					ExternalBranch:   c.ExternalBranch,
					ExternalState:    campaigns.ChangesetExternalStateOpen,
					OwnedByCampaign:  campaign.ID,
					ReconcilerState:  campaigns.ReconcilerStateCompleted,
					PublicationState: campaigns.ChangesetPublicationStatePublished,
					DiffStat:         ct.TestChangsetSpecDiffStat,
					AttachedTo:       []int64{campaign.ID},
				}
				ct.ReloadAndAssertChangeset(t, ctx, store, c, assertions)

				// STEP 2: Now we apply a new spec without any changesets.
				campaignSpec2 := ct.CreateCampaignSpec(t, ctx, store, "detach-reattach-changeset", admin.ID)
				applyAndListChangesets(adminCtx, t, svc, campaignSpec2.RandID, 1)

				// Our previously published changeset should be marked as "to be closed"
				assertions.Closing = true
				assertions.DetachFrom = []int64{campaign.ID}
				assertions.AttachedTo = []int64{}
				assertions.ReconcilerState = campaigns.ReconcilerStateQueued
				// And the previous spec is recorded.
				assertions.PreviousSpec = spec1.ID
				c = ct.ReloadAndAssertChangeset(t, ctx, store, c, assertions)

				// Now we update the changeset to make it look closed.
				ct.SetChangesetClosed(t, ctx, store, c)
				assertions.Closing = false
				assertions.DetachFrom = []int64{}
				assertions.ReconcilerState = campaigns.ReconcilerStateCompleted
				assertions.ExternalState = campaigns.ChangesetExternalStateClosed
				ct.ReloadAndAssertChangeset(t, ctx, store, c, assertions)

				// STEP 3: We apply a new campaign spec with a changeset spec that
				// matches the old changeset and expect _the same changeset_ to be
				// re-attached.
				campaignSpec3 := ct.CreateCampaignSpec(t, ctx, store, "detach-reattach-changeset", admin.ID)

				specOpts.CampaignSpec = campaignSpec3.ID
				spec2 := ct.CreateChangesetSpec(t, ctx, store, specOpts)

				_, changesets = applyAndListChangesets(adminCtx, t, svc, campaignSpec3.RandID, 1)

				attachedChangeset := changesets[0]
				if have, want := attachedChangeset.ID, c.ID; have != want {
					t.Fatalf("attached changeset has wrong ID. want=%d, have=%d", want, have)
				}

				// Assert that the changeset has been updated to point to the new spec
				assertions.CurrentSpec = spec2.ID
				// Assert that the previous spec is still spec 1
				assertions.PreviousSpec = spec1.ID
				assertions.ReconcilerState = campaigns.ReconcilerStateQueued
				assertions.AttachedTo = []int64{campaign.ID}
				ct.AssertChangeset(t, attachedChangeset, assertions)
			})

			t.Run("changeset has failed closing before re-attaching", func(t *testing.T) {
				ct.TruncateTables(t, db, "changeset_events", "changesets", "campaigns", "campaign_specs", "changeset_specs")
				campaignSpec1 := ct.CreateCampaignSpec(t, ctx, store, "detach-reattach-failed-changeset", admin.ID)

				specOpts := ct.TestSpecOpts{
					User:         admin.ID,
					Repo:         repos[0].ID,
					CampaignSpec: campaignSpec1.ID,
					HeadRef:      "refs/heads/detached-reattach-failed",
				}
				spec1 := ct.CreateChangesetSpec(t, ctx, store, specOpts)

				// STEP 1: We apply the spec and expect 1 changeset.
				campaign, changesets := applyAndListChangesets(adminCtx, t, svc, campaignSpec1.RandID, 1)

				// Now we update the changeset so it looks like it's been published
				// on the code host.
				c := changesets[0]
				ct.SetChangesetPublished(t, ctx, store, c, "80022", specOpts.HeadRef)

				assertions := ct.ChangesetAssertions{
					Repo:             c.RepoID,
					CurrentSpec:      spec1.ID,
					ExternalID:       c.ExternalID,
					ExternalBranch:   c.ExternalBranch,
					ExternalState:    campaigns.ChangesetExternalStateOpen,
					OwnedByCampaign:  campaign.ID,
					ReconcilerState:  campaigns.ReconcilerStateCompleted,
					PublicationState: campaigns.ChangesetPublicationStatePublished,
					DiffStat:         ct.TestChangsetSpecDiffStat,
					AttachedTo:       []int64{campaign.ID},
				}
				ct.ReloadAndAssertChangeset(t, ctx, store, c, assertions)

				// STEP 2: Now we apply a new spec without any changesets.
				campaignSpec2 := ct.CreateCampaignSpec(t, ctx, store, "detach-reattach-failed-changeset", admin.ID)
				applyAndListChangesets(adminCtx, t, svc, campaignSpec2.RandID, 1)

				// Our previously published changeset should be marked as "to be closed"
				assertions.Closing = true
				assertions.DetachFrom = []int64{campaign.ID}
				assertions.AttachedTo = []int64{}
				assertions.ReconcilerState = campaigns.ReconcilerStateQueued
				// And the previous spec is recorded.
				assertions.PreviousSpec = spec1.ID
				c = ct.ReloadAndAssertChangeset(t, ctx, store, c, assertions)

				if len(c.Campaigns) != 1 {
					t.Fatal("Expected changeset to be still attached to campaign, but wasn't")
				}

				// Now we update the changeset to simulate that closing failed.
				ct.SetChangesetFailed(t, ctx, store, c)
				assertions.Closing = true
				assertions.ReconcilerState = campaigns.ReconcilerStateFailed
				assertions.ExternalState = campaigns.ChangesetExternalStateOpen

				// Side-effects of ct.setChangesetFailed.
				assertions.FailureMessage = c.FailureMessage
				assertions.NumFailures = c.NumFailures
				ct.ReloadAndAssertChangeset(t, ctx, store, c, assertions)

				// STEP 3: We apply a new campaign spec with a changeset spec that
				// matches the old changeset and expect _the same changeset_ to be
				// re-attached.
				campaignSpec3 := ct.CreateCampaignSpec(t, ctx, store, "detach-reattach-failed-changeset", admin.ID)

				specOpts.CampaignSpec = campaignSpec3.ID
				spec2 := ct.CreateChangesetSpec(t, ctx, store, specOpts)

				_, changesets = applyAndListChangesets(adminCtx, t, svc, campaignSpec3.RandID, 1)

				attachedChangeset := changesets[0]
				if have, want := attachedChangeset.ID, c.ID; have != want {
					t.Fatalf("attached changeset has wrong ID. want=%d, have=%d", want, have)
				}

				// Assert that the changeset has been updated to point to the new spec
				assertions.CurrentSpec = spec2.ID
				// Assert that the previous spec is still spec 1
				assertions.PreviousSpec = spec1.ID
				assertions.ReconcilerState = campaigns.ReconcilerStateQueued
				assertions.FailureMessage = nil
				assertions.NumFailures = 0
				assertions.DetachFrom = []int64{}
				assertions.AttachedTo = []int64{campaign.ID}
				ct.AssertChangeset(t, attachedChangeset, assertions)
			})

			t.Run("changeset has not been closed before re-attaching", func(t *testing.T) {
				ct.TruncateTables(t, db, "changeset_events", "changesets", "campaigns", "campaign_specs", "changeset_specs")
				// The difference to the previous test: we DON'T update the
				// changeset to make it look closed. We want to make sure that
				// we also pick up enqueued-to-be-closed changesets.

				campaignSpec1 := ct.CreateCampaignSpec(t, ctx, store, "detach-reattach-changeset-2", admin.ID)

				specOpts := ct.TestSpecOpts{
					User:         admin.ID,
					Repo:         repos[0].ID,
					CampaignSpec: campaignSpec1.ID,
					HeadRef:      "refs/heads/detached-reattached-2",
				}
				spec1 := ct.CreateChangesetSpec(t, ctx, store, specOpts)

				// STEP 1: We apply the spec and expect 1 changeset.
				campaign, changesets := applyAndListChangesets(adminCtx, t, svc, campaignSpec1.RandID, 1)

				c := changesets[0]
				ct.SetChangesetPublished(t, ctx, store, c, "449955", specOpts.HeadRef)

				assertions := ct.ChangesetAssertions{
					Repo:             c.RepoID,
					CurrentSpec:      spec1.ID,
					ExternalID:       c.ExternalID,
					ExternalBranch:   c.ExternalBranch,
					ExternalState:    campaigns.ChangesetExternalStateOpen,
					OwnedByCampaign:  campaign.ID,
					ReconcilerState:  campaigns.ReconcilerStateCompleted,
					PublicationState: campaigns.ChangesetPublicationStatePublished,
					DiffStat:         ct.TestChangsetSpecDiffStat,
					AttachedTo:       []int64{campaign.ID},
				}
				ct.ReloadAndAssertChangeset(t, ctx, store, c, assertions)

				// STEP 2: Now we apply a new spec without any changesets.
				campaignSpec2 := ct.CreateCampaignSpec(t, ctx, store, "detach-reattach-changeset-2", admin.ID)
				applyAndListChangesets(adminCtx, t, svc, campaignSpec2.RandID, 1)

				// Our previously published changeset should be marked as "to be closed"
				assertions.Closing = true
				assertions.DetachFrom = []int64{campaign.ID}
				assertions.AttachedTo = []int64{}
				assertions.ReconcilerState = campaigns.ReconcilerStateQueued
				// And the previous spec is recorded.
				assertions.PreviousSpec = spec1.ID
				ct.ReloadAndAssertChangeset(t, ctx, store, c, assertions)

				// STEP 3: We apply a new campaign spec with a changeset spec that
				// matches the old changeset and expect _the same changeset_ to be
				// re-attached.
				campaignSpec3 := ct.CreateCampaignSpec(t, ctx, store, "detach-reattach-changeset-2", admin.ID)

				specOpts.CampaignSpec = campaignSpec3.ID
				spec2 := ct.CreateChangesetSpec(t, ctx, store, specOpts)

				_, changesets = applyAndListChangesets(adminCtx, t, svc, campaignSpec3.RandID, 1)

				attachedChangeset := changesets[0]
				if have, want := attachedChangeset.ID, c.ID; have != want {
					t.Fatalf("attached changeset has wrong ID. want=%d, have=%d", want, have)
				}

				// Assert that the changeset has been updated to point to the new spec
				assertions.CurrentSpec = spec2.ID
				// Assert that the previous spec is still spec 1
				assertions.PreviousSpec = spec1.ID
				assertions.ReconcilerState = campaigns.ReconcilerStateQueued
				assertions.DetachFrom = []int64{}
				assertions.AttachedTo = []int64{campaign.ID}
				ct.AssertChangeset(t, attachedChangeset, assertions)
			})
		})
	})

	t.Run("applying to closed campaign", func(t *testing.T) {
		ct.TruncateTables(t, db, "changeset_events", "changesets", "campaigns", "campaign_specs", "changeset_specs")
		campaignSpec := ct.CreateCampaignSpec(t, ctx, store, "closed-campaign", admin.ID)
		campaign := ct.CreateCampaign(t, ctx, store, "closed-campaign", admin.ID, campaignSpec.ID)

		campaign.ClosedAt = time.Now()
		if err := store.UpdateCampaign(ctx, campaign); err != nil {
			t.Fatalf("failed to update campaign: %s", err)
		}

		_, err := svc.ApplyCampaign(adminCtx, ApplyCampaignOpts{
			CampaignSpecRandID: campaignSpec.RandID,
		})
		if err != ErrApplyClosedCampaign {
			t.Fatalf("ApplyCampaign returned unexpected error: %s", err)
		}
	})
}

func applyAndListChangesets(ctx context.Context, t *testing.T, svc *Service, campaignSpecRandID string, wantChangesets int) (*campaigns.Campaign, campaigns.Changesets) {
	t.Helper()

	campaign, err := svc.ApplyCampaign(ctx, ApplyCampaignOpts{
		CampaignSpecRandID: campaignSpecRandID,
	})
	if err != nil {
		t.Fatalf("failed to apply campaign: %s", err)
	}

	if campaign.ID == 0 {
		t.Fatalf("campaign ID is zero")
	}

	changesets, _, err := svc.store.ListChangesets(ctx, store.ListChangesetsOpts{CampaignID: campaign.ID})
	if err != nil {
		t.Fatal(err)
	}

	if have, want := len(changesets), wantChangesets; have != want {
		t.Fatalf("wrong number of changesets. want=%d, have=%d", want, have)
	}

	return campaign, changesets
}
