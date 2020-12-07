package campaigns

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/testing"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func TestServiceApplyCampaign(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := backend.WithAuthzBypass(context.Background())
	dbtesting.SetupGlobalTestDB(t)

	admin := createTestUser(t, true)
	adminCtx := actor.WithActor(context.Background(), actor.FromUser(admin.ID))

	user := createTestUser(t, false)
	userCtx := actor.WithActor(context.Background(), actor.FromUser(user.ID))

	repos, _ := ct.CreateTestRepos(t, ctx, dbconn.Global, 4)

	now := timeutil.Now()
	clock := func() time.Time { return now }
	store := NewStoreWithClock(dbconn.Global, clock)
	svc := NewService(store, httpcli.NewExternalHTTPClientFactory())

	t.Run("campaignSpec without changesetSpecs", func(t *testing.T) {
		t.Run("new campaign", func(t *testing.T) {
			campaignSpec := createCampaignSpec(t, ctx, store, "campaign1", admin.ID)
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
			campaignSpec := createCampaignSpec(t, ctx, store, "campaign2", admin.ID)
			campaign := createCampaign(t, ctx, store, "campaign2", admin.ID, campaignSpec.ID)

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
				campaignSpec2 := createCampaignSpec(t, ctx, store, "campaign2", admin.ID)
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
				campaignSpec := createCampaignSpec(t, ctx, store, "created-by-user", user.ID)
				campaign := createCampaign(t, ctx, store, "created-by-user", user.ID, campaignSpec.ID)

				if have, want := campaign.InitialApplierID, user.ID; have != want {
					t.Fatalf("campaign InitialApplierID is wrong. want=%d, have=%d", want, have)
				}

				if have, want := campaign.LastApplierID, user.ID; have != want {
					t.Fatalf("campaign LastApplierID is wrong. want=%d, have=%d", want, have)
				}

				campaignSpec2 := createCampaignSpec(t, ctx, store, "created-by-user", user.ID)
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
				user2 := createTestUser(t, false)
				campaignSpec2 := createCampaignSpec(t, ctx, store, "campaign2", user2.ID)

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
				campaignSpec2 := createCampaignSpec(t, ctx, store, "campaign2", admin.ID)

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
				campaignSpec2 := createCampaignSpec(t, ctx, store, "campaign2", admin.ID)

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
			campaignSpec := createCampaignSpec(t, ctx, store, "campaign3", admin.ID)

			spec1 := createChangesetSpec(t, ctx, store, testSpecOpts{
				user:         admin.ID,
				repo:         repos[0].ID,
				campaignSpec: campaignSpec.ID,
				externalID:   "1234",
			})

			spec2 := createChangesetSpec(t, ctx, store, testSpecOpts{
				user:         admin.ID,
				repo:         repos[1].ID,
				campaignSpec: campaignSpec.ID,
				headRef:      "refs/heads/my-branch",
			})

			campaign, cs := applyAndListChangesets(adminCtx, t, svc, campaignSpec.RandID, 2)

			if have, want := campaign.Name, "campaign3"; have != want {
				t.Fatalf("wrong campaign name. want=%s, have=%s", want, have)
			}

			c1 := cs.Find(campaigns.WithExternalID(spec1.Spec.ExternalID))
			assertChangeset(t, c1, changesetAssertions{
				repo:             spec1.RepoID,
				externalID:       "1234",
				unsynced:         true,
				reconcilerState:  campaigns.ReconcilerStateQueued,
				publicationState: campaigns.ChangesetPublicationStatePublished,
			})

			c2 := cs.Find(campaigns.WithCurrentSpecID(spec2.ID))
			assertChangeset(t, c2, changesetAssertions{
				repo:             spec2.RepoID,
				currentSpec:      spec2.ID,
				ownedByCampaign:  campaign.ID,
				reconcilerState:  campaigns.ReconcilerStateQueued,
				publicationState: campaigns.ChangesetPublicationStateUnpublished,
				diffStat:         testChangsetSpecDiffStat,
			})
		})

		t.Run("campaign with changesets", func(t *testing.T) {
			// First we create a campaignSpec and apply it, so that we have
			// changesets and changesetSpecs in the database, wired up
			// correctly.
			campaignSpec1 := createCampaignSpec(t, ctx, store, "campaign4", admin.ID)

			createChangesetSpec(t, ctx, store, testSpecOpts{
				user:         admin.ID,
				repo:         repos[0].ID,
				campaignSpec: campaignSpec1.ID,
				externalID:   "1234",
			})

			createChangesetSpec(t, ctx, store, testSpecOpts{
				user:         admin.ID,
				repo:         repos[0].ID,
				campaignSpec: campaignSpec1.ID,
				externalID:   "5678",
			})

			oldSpec3 := createChangesetSpec(t, ctx, store, testSpecOpts{
				user:         admin.ID,
				repo:         repos[1].ID,
				campaignSpec: campaignSpec1.ID,
				headRef:      "refs/heads/repo-1-branch-1",
			})

			oldSpec4 := createChangesetSpec(t, ctx, store, testSpecOpts{
				user:         admin.ID,
				repo:         repos[2].ID,
				campaignSpec: campaignSpec1.ID,
				headRef:      "refs/heads/repo-2-branch-1",
			})

			// Apply and expect 4 changesets
			_, oldChangesets := applyAndListChangesets(adminCtx, t, svc, campaignSpec1.RandID, 4)

			// Now we create another campaign spec with the same campaign name
			// and namespace.
			campaignSpec2 := createCampaignSpec(t, ctx, store, "campaign4", admin.ID)

			// Same
			spec1 := createChangesetSpec(t, ctx, store, testSpecOpts{
				user:         admin.ID,
				repo:         repos[0].ID,
				campaignSpec: campaignSpec2.ID,
				externalID:   "1234",
			})

			// DIFFERENT: Track #9999 in repo[0]
			spec2 := createChangesetSpec(t, ctx, store, testSpecOpts{
				user:         admin.ID,
				repo:         repos[0].ID,
				campaignSpec: campaignSpec2.ID,
				externalID:   "5678",
			})

			// Same
			spec3 := createChangesetSpec(t, ctx, store, testSpecOpts{
				user:         admin.ID,
				repo:         repos[1].ID,
				campaignSpec: campaignSpec2.ID,
				headRef:      "refs/heads/repo-1-branch-1",
			})

			// DIFFERENT: branch changed in repo[2]
			spec4 := createChangesetSpec(t, ctx, store, testSpecOpts{
				user:         admin.ID,
				repo:         repos[2].ID,
				campaignSpec: campaignSpec2.ID,
				headRef:      "refs/heads/repo-2-branch-2",
			})

			// NEW: repo[3]
			spec5 := createChangesetSpec(t, ctx, store, testSpecOpts{
				user:         admin.ID,
				repo:         repos[3].ID,
				campaignSpec: campaignSpec2.ID,
				headRef:      "refs/heads/repo-3-branch-1",
			})

			// Before we apply the new campaign spec, we make the changeset we
			// expect to be closed to look "published", otherwise it won't be
			// closed.
			wantClosed := oldChangesets.Find(campaigns.WithCurrentSpecID(oldSpec4.ID))
			setChangesetPublished(t, ctx, store, wantClosed, "98765", oldSpec4.Spec.HeadRef)

			changeset3 := oldChangesets.Find(campaigns.WithCurrentSpecID(oldSpec3.ID))
			setChangesetPublished(t, ctx, store, changeset3, "12345", oldSpec3.Spec.HeadRef)

			// Apply and expect 5 changesets
			campaign, cs := applyAndListChangesets(adminCtx, t, svc, campaignSpec2.RandID, 5)

			// This changeset we want marked as "to be closed"
			reloadAndAssertChangeset(t, ctx, store, wantClosed, changesetAssertions{
				repo:         repos[2].ID,
				currentSpec:  oldSpec4.ID,
				previousSpec: oldSpec4.ID,
				externalID:   wantClosed.ExternalID,
				// It's still open, just _marked as to be closed_.
				externalState:    campaigns.ChangesetExternalStateOpen,
				externalBranch:   wantClosed.ExternalBranch,
				ownedByCampaign:  campaign.ID,
				reconcilerState:  campaigns.ReconcilerStateQueued,
				publicationState: campaigns.ChangesetPublicationStatePublished,
				diffStat:         testChangsetSpecDiffStat,
				closing:          true,
			})

			c1 := cs.Find(campaigns.WithExternalID(spec1.Spec.ExternalID))
			assertChangeset(t, c1, changesetAssertions{
				repo:             repos[0].ID,
				currentSpec:      0,
				previousSpec:     0,
				externalID:       "1234",
				unsynced:         true,
				reconcilerState:  campaigns.ReconcilerStateQueued,
				publicationState: campaigns.ChangesetPublicationStatePublished,
			})

			c2 := cs.Find(campaigns.WithExternalID(spec2.Spec.ExternalID))
			assertChangeset(t, c2, changesetAssertions{
				repo:             repos[0].ID,
				currentSpec:      0,
				previousSpec:     0,
				externalID:       "5678",
				unsynced:         true,
				reconcilerState:  campaigns.ReconcilerStateQueued,
				publicationState: campaigns.ChangesetPublicationStatePublished,
			})

			c3 := cs.Find(campaigns.WithCurrentSpecID(spec3.ID))
			assertChangeset(t, c3, changesetAssertions{
				repo:           repos[1].ID,
				currentSpec:    spec3.ID,
				externalID:     changeset3.ExternalID,
				externalBranch: changeset3.ExternalBranch,
				externalState:  campaigns.ChangesetExternalStateOpen,
				// Has a previous spec, because it succeeded publishing.
				previousSpec:     oldSpec3.ID,
				ownedByCampaign:  campaign.ID,
				reconcilerState:  campaigns.ReconcilerStateQueued,
				publicationState: campaigns.ChangesetPublicationStatePublished,
				diffStat:         testChangsetSpecDiffStat,
			})

			c4 := cs.Find(campaigns.WithCurrentSpecID(spec4.ID))
			assertChangeset(t, c4, changesetAssertions{
				repo:             repos[2].ID,
				currentSpec:      spec4.ID,
				ownedByCampaign:  campaign.ID,
				reconcilerState:  campaigns.ReconcilerStateQueued,
				publicationState: campaigns.ChangesetPublicationStateUnpublished,
				diffStat:         testChangsetSpecDiffStat,
			})

			c5 := cs.Find(campaigns.WithCurrentSpecID(spec5.ID))
			assertChangeset(t, c5, changesetAssertions{
				repo:             repos[3].ID,
				currentSpec:      spec5.ID,
				ownedByCampaign:  campaign.ID,
				reconcilerState:  campaigns.ReconcilerStateQueued,
				publicationState: campaigns.ChangesetPublicationStateUnpublished,
				diffStat:         testChangsetSpecDiffStat,
			})
		})

		t.Run("campaign tracking changesets owned by another campaign", func(t *testing.T) {
			campaignSpec1 := createCampaignSpec(t, ctx, store, "owner-campaign", admin.ID)

			oldSpec1 := createChangesetSpec(t, ctx, store, testSpecOpts{
				user:         admin.ID,
				repo:         repos[0].ID,
				campaignSpec: campaignSpec1.ID,
				headRef:      "refs/heads/repo-0-branch-0",
			})

			ownerCampaign, ownerChangesets := applyAndListChangesets(adminCtx, t, svc, campaignSpec1.RandID, 1)

			// Now we update the changeset so it looks like it's been published
			// on the code host.
			c := ownerChangesets[0]
			setChangesetPublished(t, ctx, store, c, "88888", "refs/heads/repo-0-branch-0")

			// This other campaign tracks the changeset created by the first one
			campaignSpec2 := createCampaignSpec(t, ctx, store, "tracking-campaign", admin.ID)
			createChangesetSpec(t, ctx, store, testSpecOpts{
				user:         admin.ID,
				repo:         c.RepoID,
				campaignSpec: campaignSpec2.ID,
				externalID:   c.ExternalID,
			})

			_, trackedChangesets := applyAndListChangesets(adminCtx, t, svc, campaignSpec2.RandID, 1)
			// This should still point to the owner campaign
			c2 := trackedChangesets[0]
			trackedChangesetAssertions := changesetAssertions{
				repo:             c.RepoID,
				currentSpec:      oldSpec1.ID,
				ownedByCampaign:  ownerCampaign.ID,
				externalBranch:   c.ExternalBranch,
				externalID:       c.ExternalID,
				externalState:    campaigns.ChangesetExternalStateOpen,
				reconcilerState:  campaigns.ReconcilerStateCompleted,
				publicationState: campaigns.ChangesetPublicationStatePublished,
				diffStat:         testChangsetSpecDiffStat,
			}
			assertChangeset(t, c2, trackedChangesetAssertions)

			// Now try to apply a new spec that wants to modify the formerly tracked changeset.
			campaignSpec3 := createCampaignSpec(t, ctx, store, "tracking-campaign", admin.ID)

			spec3 := createChangesetSpec(t, ctx, store, testSpecOpts{
				user:         admin.ID,
				repo:         repos[0].ID,
				campaignSpec: campaignSpec3.ID,
				headRef:      "refs/heads/repo-0-branch-0",
			})
			// Apply again. This should have detached the tracked changeset but it should not be closed, since the campaign is
			// not the owner.
			trackingCampaign, cs := applyAndListChangesets(adminCtx, t, svc, campaignSpec3.RandID, 1)

			trackedChangesetAssertions.closing = false
			reloadAndAssertChangeset(t, ctx, store, c2, trackedChangesetAssertions)

			// But we do want to have a new changeset record that is going to create a new changeset on the code host.
			reloadAndAssertChangeset(t, ctx, store, cs[0], changesetAssertions{
				repo:             spec3.RepoID,
				currentSpec:      spec3.ID,
				ownedByCampaign:  trackingCampaign.ID,
				reconcilerState:  campaigns.ReconcilerStateQueued,
				publicationState: campaigns.ChangesetPublicationStateUnpublished,
				diffStat:         testChangsetSpecDiffStat,
			})
		})

		t.Run("campaign with changeset that is unpublished", func(t *testing.T) {
			campaignSpec1 := createCampaignSpec(t, ctx, store, "unpublished-changesets", admin.ID)

			createChangesetSpec(t, ctx, store, testSpecOpts{
				user:         admin.ID,
				repo:         repos[3].ID,
				campaignSpec: campaignSpec1.ID,
				headRef:      "refs/heads/never-published",
			})

			// We apply the spec and expect 1 changeset
			applyAndListChangesets(adminCtx, t, svc, campaignSpec1.RandID, 1)

			// But the changeset was not published yet.
			// And now we apply a new spec without any changesets.
			campaignSpec2 := createCampaignSpec(t, ctx, store, "unpublished-changesets", admin.ID)

			// That should close no changesets, but leave the campaign with 0 changesets,
			// and the unpublished changesets should be detached
			applyAndListChangesets(adminCtx, t, svc, campaignSpec2.RandID, 0)
		})

		t.Run("campaign with changeset that wasn't processed before reapply", func(t *testing.T) {
			campaignSpec1 := createCampaignSpec(t, ctx, store, "queued-changesets", admin.ID)

			specOpts := testSpecOpts{
				user:         admin.ID,
				repo:         repos[3].ID,
				campaignSpec: campaignSpec1.ID,
				title:        "Spec1",
				headRef:      "refs/heads/queued",
				published:    true,
			}
			spec1 := createChangesetSpec(t, ctx, store, specOpts)

			// We apply the spec and expect 1 changeset
			campaign, changesets := applyAndListChangesets(adminCtx, t, svc, campaignSpec1.RandID, 1)

			// And publish it.
			setChangesetPublished(t, ctx, store, changesets[0], "123-queued", "refs/heads/queued")

			reloadAndAssertChangeset(t, ctx, store, changesets[0], changesetAssertions{
				reconcilerState:  campaigns.ReconcilerStateCompleted,
				publicationState: campaigns.ChangesetPublicationStatePublished,
				externalBranch:   "refs/heads/queued",
				externalID:       "123-queued",
				externalState:    campaigns.ChangesetExternalStateOpen,
				repo:             repos[3].ID,
				currentSpec:      spec1.ID,
				ownedByCampaign:  campaign.ID,
				diffStat:         testChangsetSpecDiffStat,
			})

			// Apply again so that an update to the changeset is pending.
			campaignSpec2 := createCampaignSpec(t, ctx, store, "queued-changesets", admin.ID)

			specOpts.campaignSpec = campaignSpec2.ID
			specOpts.title = "Spec2"
			spec2 := createChangesetSpec(t, ctx, store, specOpts)

			// That should still want to publish the changeset
			_, changesets = applyAndListChangesets(adminCtx, t, svc, campaignSpec2.RandID, 1)

			reloadAndAssertChangeset(t, ctx, store, changesets[0], changesetAssertions{
				reconcilerState:  campaigns.ReconcilerStateQueued,
				publicationState: campaigns.ChangesetPublicationStatePublished,
				externalBranch:   "refs/heads/queued",
				externalID:       "123-queued",
				externalState:    campaigns.ChangesetExternalStateOpen,
				repo:             repos[3].ID,
				currentSpec:      spec2.ID,
				// Track the previous spec.
				previousSpec:    spec1.ID,
				ownedByCampaign: campaign.ID,
				diffStat:        testChangsetSpecDiffStat,
			})

			// Make sure the reconciler wants to update this changeset.
			plan, err := DetermineReconcilerPlan(
				// changesets[0].PreviousSpecID
				spec1,
				// changesets[0].CurrentSpecID
				spec2,
				changesets[0],
			)
			if err != nil {
				t.Fatal(err)
			}
			if !plan.Ops.Equal(ReconcilerOperations{campaigns.ReconcilerOperationUpdate}) {
				t.Fatalf("Got invalid reconciler operations: %q", plan.Ops.String())
			}

			// And now we apply a new spec before the reconciler could process the changeset.
			campaignSpec3 := createCampaignSpec(t, ctx, store, "queued-changesets", admin.ID)

			// No change this time, just reapplying.
			specOpts.campaignSpec = campaignSpec3.ID
			spec3 := createChangesetSpec(t, ctx, store, specOpts)

			_, changesets = applyAndListChangesets(adminCtx, t, svc, campaignSpec3.RandID, 1)

			reloadAndAssertChangeset(t, ctx, store, changesets[0], changesetAssertions{
				reconcilerState:  campaigns.ReconcilerStateQueued,
				publicationState: campaigns.ChangesetPublicationStatePublished,
				externalBranch:   "refs/heads/queued",
				externalID:       "123-queued",
				externalState:    campaigns.ChangesetExternalStateOpen,
				repo:             repos[3].ID,
				currentSpec:      spec3.ID,
				// Still be pointing at the first spec, since the second was never applied.
				previousSpec:    spec1.ID,
				ownedByCampaign: campaign.ID,
				diffStat:        testChangsetSpecDiffStat,
			})

			// Make sure the reconciler would still update this changeset.
			plan, err = DetermineReconcilerPlan(
				// changesets[0].PreviousSpecID
				spec1,
				// changesets[0].CurrentSpecID
				spec3,
				changesets[0],
			)
			if err != nil {
				t.Fatal(err)
			}
			if !plan.Ops.Equal(ReconcilerOperations{campaigns.ReconcilerOperationUpdate}) {
				t.Fatalf("Got invalid reconciler operations: %q", plan.Ops.String())
			}

			// Now test that it still updates when this update failed.
			setChangesetFailed(t, ctx, store, changesets[0])

			campaignSpec4 := createCampaignSpec(t, ctx, store, "queued-changesets", admin.ID)

			// No change this time, just reapplying.
			specOpts.campaignSpec = campaignSpec4.ID
			spec4 := createChangesetSpec(t, ctx, store, specOpts)

			_, changesets = applyAndListChangesets(adminCtx, t, svc, campaignSpec4.RandID, 1)

			reloadAndAssertChangeset(t, ctx, store, changesets[0], changesetAssertions{
				reconcilerState:  campaigns.ReconcilerStateQueued,
				publicationState: campaigns.ChangesetPublicationStatePublished,
				externalBranch:   "refs/heads/queued",
				externalID:       "123-queued",
				externalState:    campaigns.ChangesetExternalStateOpen,
				repo:             repos[3].ID,
				currentSpec:      spec4.ID,
				// Still be pointing at the first spec, since the second and third were never applied.
				previousSpec:    spec1.ID,
				ownedByCampaign: campaign.ID,
				diffStat:        testChangsetSpecDiffStat,
			})

			// Make sure the reconciler would still update this changeset.
			plan, err = DetermineReconcilerPlan(
				// changesets[0].PreviousSpecID
				spec1,
				// changesets[0].CurrentSpecID
				spec4,
				changesets[0],
			)
			if err != nil {
				t.Fatal(err)
			}
			if !plan.Ops.Equal(ReconcilerOperations{campaigns.ReconcilerOperationUpdate}) {
				t.Fatalf("Got invalid reconciler operations: %q", plan.Ops.String())
			}
		})

		t.Run("missing repository permissions", func(t *testing.T) {
			ct.MockRepoPermissions(t, user.ID, repos[0].ID, repos[2].ID, repos[3].ID)

			// NOTE: We cannot use a context that has authz bypassed.
			campaignSpec := createCampaignSpec(t, userCtx, store, "missing-permissions", user.ID)

			createChangesetSpec(t, userCtx, store, testSpecOpts{
				user:         user.ID,
				repo:         repos[0].ID,
				campaignSpec: campaignSpec.ID,
				externalID:   "1234",
			})

			createChangesetSpec(t, userCtx, store, testSpecOpts{
				user:         user.ID,
				repo:         repos[1].ID, // Not authorized to access this repository
				campaignSpec: campaignSpec.ID,
				headRef:      "refs/heads/my-branch",
			})

			_, err := svc.ApplyCampaign(userCtx, ApplyCampaignOpts{
				CampaignSpecRandID: campaignSpec.RandID,
			})
			if err == nil {
				t.Fatal("expected error, but got none")
			}
			notFoundErr, ok := err.(*db.RepoNotFoundErr)
			if !ok {
				t.Fatalf("expected RepoNotFoundErr but got: %s", err)
			}
			if notFoundErr.ID != repos[1].ID {
				t.Fatalf("wrong repository ID in RepoNotFoundErr: %d", notFoundErr.ID)
			}
		})

		t.Run("campaign with errored changeset", func(t *testing.T) {
			campaignSpec1 := createCampaignSpec(t, ctx, store, "errored-changeset-campaign", admin.ID)

			spec1Opts := testSpecOpts{
				user:         admin.ID,
				repo:         repos[0].ID,
				campaignSpec: campaignSpec1.ID,
				externalID:   "1234",
				published:    true,
			}
			createChangesetSpec(t, ctx, store, spec1Opts)

			spec2Opts := testSpecOpts{
				user:         admin.ID,
				repo:         repos[1].ID,
				campaignSpec: campaignSpec1.ID,
				headRef:      "refs/heads/repo-1-branch-1",
				published:    true,
			}
			createChangesetSpec(t, ctx, store, spec2Opts)

			_, oldChangesets := applyAndListChangesets(adminCtx, t, svc, campaignSpec1.RandID, 2)

			// Set the changesets to look like they failed in the reconciler
			for _, c := range oldChangesets {
				setChangesetFailed(t, ctx, store, c)
			}

			// Now we create another campaign spec with the same campaign name
			// and namespace.
			campaignSpec2 := createCampaignSpec(t, ctx, store, "errored-changeset-campaign", admin.ID)
			spec1Opts.campaignSpec = campaignSpec2.ID
			newSpec1 := createChangesetSpec(t, ctx, store, spec1Opts)
			spec2Opts.campaignSpec = campaignSpec2.ID
			newSpec2 := createChangesetSpec(t, ctx, store, spec2Opts)

			campaign, cs := applyAndListChangesets(adminCtx, t, svc, campaignSpec2.RandID, 2)

			c1 := cs.Find(campaigns.WithExternalID(newSpec1.Spec.ExternalID))
			reloadAndAssertChangeset(t, ctx, store, c1, changesetAssertions{
				repo:             spec1Opts.repo,
				externalID:       "1234",
				unsynced:         true,
				publicationState: campaigns.ChangesetPublicationStatePublished,

				reconcilerState: campaigns.ReconcilerStateQueued,
				failureMessage:  nil,
				numFailures:     0,
			})

			c2 := cs.Find(campaigns.WithCurrentSpecID(newSpec2.ID))
			assertChangeset(t, c2, changesetAssertions{
				repo:        newSpec2.RepoID,
				currentSpec: newSpec2.ID,
				// An errored changeset doesn't get the specs rotated, to prevent https://github.com/sourcegraph/sourcegraph/issues/16041.
				previousSpec:     0,
				ownedByCampaign:  campaign.ID,
				publicationState: campaigns.ChangesetPublicationStateUnpublished,
				diffStat:         testChangsetSpecDiffStat,

				reconcilerState: campaigns.ReconcilerStateQueued,
				failureMessage:  nil,
				numFailures:     0,
			})

			// Make sure the reconciler would still publish this changeset.
			plan, err := DetermineReconcilerPlan(
				// c2.previousSpec is 0
				nil,
				// c2.currentSpec is newSpec2
				newSpec2,
				c2,
			)
			if err != nil {
				t.Fatal(err)
			}
			if !plan.Ops.Equal(ReconcilerOperations{campaigns.ReconcilerOperationPush, campaigns.ReconcilerOperationPublish}) {
				t.Fatalf("Got invalid reconciler operations: %q", plan.Ops.String())
			}
		})

		t.Run("closed and detached changeset not re-enqueued for close", func(t *testing.T) {
			campaignSpec1 := createCampaignSpec(t, ctx, store, "detached-closed-changeset", admin.ID)

			specOpts := testSpecOpts{
				user:         admin.ID,
				repo:         repos[0].ID,
				campaignSpec: campaignSpec1.ID,
				headRef:      "refs/heads/detached-closed",
			}
			spec1 := createChangesetSpec(t, ctx, store, specOpts)

			// STEP 1: We apply the spec and expect 1 changeset.
			campaign, changesets := applyAndListChangesets(adminCtx, t, svc, campaignSpec1.RandID, 1)

			// Now we update the changeset so it looks like it's been published
			// on the code host.
			c := changesets[0]
			setChangesetPublished(t, ctx, store, c, "995544", specOpts.headRef)

			assertions := changesetAssertions{
				repo:             c.RepoID,
				currentSpec:      spec1.ID,
				externalID:       c.ExternalID,
				externalBranch:   c.ExternalBranch,
				externalState:    campaigns.ChangesetExternalStateOpen,
				ownedByCampaign:  campaign.ID,
				reconcilerState:  campaigns.ReconcilerStateCompleted,
				publicationState: campaigns.ChangesetPublicationStatePublished,
				diffStat:         testChangsetSpecDiffStat,
			}
			c = reloadAndAssertChangeset(t, ctx, store, c, assertions)

			// STEP 2: Now we apply a new spec without any changesets.
			campaignSpec2 := createCampaignSpec(t, ctx, store, "detached-closed-changeset", admin.ID)
			applyAndListChangesets(adminCtx, t, svc, campaignSpec2.RandID, 0)

			// Our previously published changeset should be marked as "to be closed"
			assertions.closing = true
			assertions.reconcilerState = campaigns.ReconcilerStateQueued
			// And the previous spec is recorded, because the previous run finished with reconcilerState completed.
			assertions.previousSpec = spec1.ID
			c = reloadAndAssertChangeset(t, ctx, store, c, assertions)

			// Now we update the changeset to make it look closed.
			setChangesetClosed(t, ctx, store, c)
			assertions.closing = false
			assertions.reconcilerState = campaigns.ReconcilerStateCompleted
			assertions.externalState = campaigns.ChangesetExternalStateClosed
			c = reloadAndAssertChangeset(t, ctx, store, c, assertions)

			// STEP 3: We apply a new campaign spec and expect that the detached changeset record is not re-enqueued.
			campaignSpec3 := createCampaignSpec(t, ctx, store, "detached-closed-changeset", admin.ID)

			applyAndListChangesets(adminCtx, t, svc, campaignSpec3.RandID, 0)

			// Assert that the changeset record is still completed and closed.
			reloadAndAssertChangeset(t, ctx, store, c, assertions)
		})

		t.Run("campaign with changeset that is detached and reattached", func(t *testing.T) {
			t.Run("changeset has been closed before re-attaching", func(t *testing.T) {
				campaignSpec1 := createCampaignSpec(t, ctx, store, "detach-reattach-changeset", admin.ID)

				specOpts := testSpecOpts{
					user:         admin.ID,
					repo:         repos[0].ID,
					campaignSpec: campaignSpec1.ID,
					headRef:      "refs/heads/detached-reattached",
				}
				spec1 := createChangesetSpec(t, ctx, store, specOpts)

				// STEP 1: We apply the spec and expect 1 changeset.
				campaign, changesets := applyAndListChangesets(adminCtx, t, svc, campaignSpec1.RandID, 1)

				// Now we update the changeset so it looks like it's been published
				// on the code host.
				c := changesets[0]
				setChangesetPublished(t, ctx, store, c, "995533", specOpts.headRef)

				assertions := changesetAssertions{
					repo:             c.RepoID,
					currentSpec:      spec1.ID,
					externalID:       c.ExternalID,
					externalBranch:   c.ExternalBranch,
					externalState:    campaigns.ChangesetExternalStateOpen,
					ownedByCampaign:  campaign.ID,
					reconcilerState:  campaigns.ReconcilerStateCompleted,
					publicationState: campaigns.ChangesetPublicationStatePublished,
					diffStat:         testChangsetSpecDiffStat,
				}
				reloadAndAssertChangeset(t, ctx, store, c, assertions)

				// STEP 2: Now we apply a new spec without any changesets.
				campaignSpec2 := createCampaignSpec(t, ctx, store, "detach-reattach-changeset", admin.ID)
				applyAndListChangesets(adminCtx, t, svc, campaignSpec2.RandID, 0)

				// Our previously published changeset should be marked as "to be closed"
				assertions.closing = true
				assertions.reconcilerState = campaigns.ReconcilerStateQueued
				// And the previous spec is recorded.
				assertions.previousSpec = spec1.ID
				c = reloadAndAssertChangeset(t, ctx, store, c, assertions)

				// Now we update the changeset to make it look closed.
				setChangesetClosed(t, ctx, store, c)
				assertions.closing = false
				assertions.reconcilerState = campaigns.ReconcilerStateCompleted
				assertions.externalState = campaigns.ChangesetExternalStateClosed
				reloadAndAssertChangeset(t, ctx, store, c, assertions)

				// STEP 3: We apply a new campaign spec with a changeset spec that
				// matches the old changeset and expect _the same changeset_ to be
				// re-attached.
				campaignSpec3 := createCampaignSpec(t, ctx, store, "detach-reattach-changeset", admin.ID)

				specOpts.campaignSpec = campaignSpec3.ID
				spec2 := createChangesetSpec(t, ctx, store, specOpts)

				campaign, changesets = applyAndListChangesets(adminCtx, t, svc, campaignSpec3.RandID, 1)

				attachedChangeset := changesets[0]
				if have, want := attachedChangeset.ID, c.ID; have != want {
					t.Fatalf("attached changeset has wrong ID. want=%d, have=%d", want, have)
				}

				// Assert that the changeset has been updated to point to the new spec
				assertions.currentSpec = spec2.ID
				// Assert that the previous spec is still spec 1
				assertions.previousSpec = spec1.ID
				assertions.reconcilerState = campaigns.ReconcilerStateQueued
				assertChangeset(t, attachedChangeset, assertions)
			})

			t.Run("changeset has failed closing before re-attaching", func(t *testing.T) {
				campaignSpec1 := createCampaignSpec(t, ctx, store, "detach-reattach-failed-changeset", admin.ID)

				specOpts := testSpecOpts{
					user:         admin.ID,
					repo:         repos[0].ID,
					campaignSpec: campaignSpec1.ID,
					headRef:      "refs/heads/detached-reattach-failed",
				}
				spec1 := createChangesetSpec(t, ctx, store, specOpts)

				// STEP 1: We apply the spec and expect 1 changeset.
				campaign, changesets := applyAndListChangesets(adminCtx, t, svc, campaignSpec1.RandID, 1)

				// Now we update the changeset so it looks like it's been published
				// on the code host.
				c := changesets[0]
				setChangesetPublished(t, ctx, store, c, "80022", specOpts.headRef)

				assertions := changesetAssertions{
					repo:             c.RepoID,
					currentSpec:      spec1.ID,
					externalID:       c.ExternalID,
					externalBranch:   c.ExternalBranch,
					externalState:    campaigns.ChangesetExternalStateOpen,
					ownedByCampaign:  campaign.ID,
					reconcilerState:  campaigns.ReconcilerStateCompleted,
					publicationState: campaigns.ChangesetPublicationStatePublished,
					diffStat:         testChangsetSpecDiffStat,
				}
				reloadAndAssertChangeset(t, ctx, store, c, assertions)

				// STEP 2: Now we apply a new spec without any changesets.
				campaignSpec2 := createCampaignSpec(t, ctx, store, "detach-reattach-failed-changeset", admin.ID)
				applyAndListChangesets(adminCtx, t, svc, campaignSpec2.RandID, 0)

				// Our previously published changeset should be marked as "to be closed"
				assertions.closing = true
				assertions.reconcilerState = campaigns.ReconcilerStateQueued
				// And the previous spec is recorded.
				assertions.previousSpec = spec1.ID
				c = reloadAndAssertChangeset(t, ctx, store, c, assertions)

				if len(c.CampaignIDs) != 0 {
					t.Fatal("Expected changeset to be detached from campaign, but wasn't")
				}

				// Now we update the changeset to simulate that closing failed.
				setChangesetFailed(t, ctx, store, c)
				assertions.closing = true
				assertions.reconcilerState = campaigns.ReconcilerStateErrored
				assertions.externalState = campaigns.ChangesetExternalStateOpen

				// Side-effects of setChangesetFailed.
				assertions.failureMessage = c.FailureMessage
				assertions.numFailures = c.NumFailures
				reloadAndAssertChangeset(t, ctx, store, c, assertions)

				// STEP 3: We apply a new campaign spec with a changeset spec that
				// matches the old changeset and expect _the same changeset_ to be
				// re-attached.
				campaignSpec3 := createCampaignSpec(t, ctx, store, "detach-reattach-failed-changeset", admin.ID)

				specOpts.campaignSpec = campaignSpec3.ID
				spec2 := createChangesetSpec(t, ctx, store, specOpts)

				_, changesets = applyAndListChangesets(adminCtx, t, svc, campaignSpec3.RandID, 1)

				attachedChangeset := changesets[0]
				if have, want := attachedChangeset.ID, c.ID; have != want {
					t.Fatalf("attached changeset has wrong ID. want=%d, have=%d", want, have)
				}

				// Assert that the changeset has been updated to point to the new spec
				assertions.currentSpec = spec2.ID
				// Assert that the previous spec is still spec 1
				assertions.previousSpec = spec1.ID
				assertions.reconcilerState = campaigns.ReconcilerStateQueued
				assertions.failureMessage = nil
				assertions.numFailures = 0
				assertChangeset(t, attachedChangeset, assertions)
			})

			t.Run("changeset has not been closed before re-attaching", func(t *testing.T) {
				// The difference to the previous test: we DON'T update the
				// changeset to make it look closed. We want to make sure that
				// we also pick up enqueued-to-be-closed changesets.

				campaignSpec1 := createCampaignSpec(t, ctx, store, "detach-reattach-changeset-2", admin.ID)

				specOpts := testSpecOpts{
					user:         admin.ID,
					repo:         repos[0].ID,
					campaignSpec: campaignSpec1.ID,
					headRef:      "refs/heads/detached-reattached-2",
				}
				spec1 := createChangesetSpec(t, ctx, store, specOpts)

				// STEP 1: We apply the spec and expect 1 changeset.
				campaign, changesets := applyAndListChangesets(adminCtx, t, svc, campaignSpec1.RandID, 1)

				c := changesets[0]
				setChangesetPublished(t, ctx, store, c, "449955", specOpts.headRef)

				assertions := changesetAssertions{
					repo:             c.RepoID,
					currentSpec:      spec1.ID,
					externalID:       c.ExternalID,
					externalBranch:   c.ExternalBranch,
					externalState:    campaigns.ChangesetExternalStateOpen,
					ownedByCampaign:  campaign.ID,
					reconcilerState:  campaigns.ReconcilerStateCompleted,
					publicationState: campaigns.ChangesetPublicationStatePublished,
					diffStat:         testChangsetSpecDiffStat,
				}
				reloadAndAssertChangeset(t, ctx, store, c, assertions)

				// STEP 2: Now we apply a new spec without any changesets.
				campaignSpec2 := createCampaignSpec(t, ctx, store, "detach-reattach-changeset-2", admin.ID)
				applyAndListChangesets(adminCtx, t, svc, campaignSpec2.RandID, 0)

				// Our previously published changeset should be marked as "to be closed"
				assertions.closing = true
				assertions.reconcilerState = campaigns.ReconcilerStateQueued
				// And the previous spec is recorded.
				assertions.previousSpec = spec1.ID
				reloadAndAssertChangeset(t, ctx, store, c, assertions)

				// STEP 3: We apply a new campaign spec with a changeset spec that
				// matches the old changeset and expect _the same changeset_ to be
				// re-attached.
				campaignSpec3 := createCampaignSpec(t, ctx, store, "detach-reattach-changeset-2", admin.ID)

				specOpts.campaignSpec = campaignSpec3.ID
				spec2 := createChangesetSpec(t, ctx, store, specOpts)

				campaign, changesets = applyAndListChangesets(adminCtx, t, svc, campaignSpec3.RandID, 1)

				attachedChangeset := changesets[0]
				if have, want := attachedChangeset.ID, c.ID; have != want {
					t.Fatalf("attached changeset has wrong ID. want=%d, have=%d", want, have)
				}

				// Assert that the changeset has been updated to point to the new spec
				assertions.currentSpec = spec2.ID
				// Assert that the previous spec is still spec 1
				assertions.previousSpec = spec1.ID
				assertions.reconcilerState = campaigns.ReconcilerStateQueued
				assertChangeset(t, attachedChangeset, assertions)
			})
		})
	})

	t.Run("applying to closed campaign", func(t *testing.T) {
		campaignSpec := createCampaignSpec(t, ctx, store, "closed-campaign", admin.ID)
		campaign := createCampaign(t, ctx, store, "closed-campaign", admin.ID, campaignSpec.ID)

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
