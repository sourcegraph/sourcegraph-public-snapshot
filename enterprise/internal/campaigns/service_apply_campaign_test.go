package campaigns

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/testing"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

func TestServiceApplyCampaign(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := backend.WithAuthzBypass(context.Background())
	dbtesting.SetupGlobalTestDB(t)

	admin := createTestUser(ctx, t)
	if !admin.SiteAdmin {
		t.Fatal("admin is not a site-admin")
	}
	adminCtx := actor.WithActor(context.Background(), actor.FromUser(admin.ID))

	user := createTestUser(ctx, t)
	if user.SiteAdmin {
		t.Fatal("user is admin, want non-admin")
	}

	repos, _ := createTestRepos(t, ctx, dbconn.Global, 4)

	now := time.Now().UTC().Truncate(time.Microsecond)
	clock := func() time.Time {
		return now.UTC().Truncate(time.Microsecond)
	}
	store := NewStoreWithClock(dbconn.Global, clock)
	svc := NewService(store, httpcli.NewExternalHTTPClientFactory())

	// The diff stat that corresponds to what's stored in the spec.
	diffStat := &diff.Stat{
		Added:   10,
		Changed: 5,
		Deleted: 2,
	}

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
				ChangesetIDs:     []int64{},
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
				user2 := createTestUser(ctx, t)
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
				diffStat:         diffStat,
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

			// Before we apply the new campaign spec, we set up the assertion
			// for changesets to be closed.
			wantClosed := oldChangesets.Find(campaigns.WithCurrentSpecID(oldSpec4.ID))
			// We need to make it look "published", otherwise it won't be closed.
			setChangesetPublished(t, ctx, store, wantClosed, oldSpec4.Spec.HeadRef, "98765")

			verifyClosed := assertChangesetsClose(t, admin.ID, wantClosed)

			// Apply and expect 5 changesets
			campaign, cs := applyAndListChangesets(adminCtx, t, svc, campaignSpec2.RandID, 5)

			verifyClosed()

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
				repo:             repos[1].ID,
				currentSpec:      spec3.ID,
				previousSpec:     oldSpec3.ID,
				ownedByCampaign:  campaign.ID,
				reconcilerState:  campaigns.ReconcilerStateQueued,
				publicationState: campaigns.ChangesetPublicationStateUnpublished,
				diffStat:         diffStat,
			})

			c4 := cs.Find(campaigns.WithCurrentSpecID(spec4.ID))
			assertChangeset(t, c4, changesetAssertions{
				repo:             repos[2].ID,
				currentSpec:      spec4.ID,
				ownedByCampaign:  campaign.ID,
				reconcilerState:  campaigns.ReconcilerStateQueued,
				publicationState: campaigns.ChangesetPublicationStateUnpublished,
				diffStat:         diffStat,
			})

			c5 := cs.Find(campaigns.WithCurrentSpecID(spec5.ID))
			assertChangeset(t, c5, changesetAssertions{
				repo:             repos[3].ID,
				currentSpec:      spec5.ID,
				ownedByCampaign:  campaign.ID,
				reconcilerState:  campaigns.ReconcilerStateQueued,
				publicationState: campaigns.ChangesetPublicationStateUnpublished,
				diffStat:         diffStat,
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
			setChangesetPublished(t, ctx, store, c, "refs/heads/repo-0-branch-0", "88888")

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
			assertChangeset(t, c2, changesetAssertions{
				repo:             c.RepoID,
				currentSpec:      oldSpec1.ID,
				ownedByCampaign:  ownerCampaign.ID,
				externalBranch:   c.ExternalBranch,
				externalID:       c.ExternalID,
				reconcilerState:  campaigns.ReconcilerStateCompleted,
				publicationState: campaigns.ChangesetPublicationStatePublished,
				diffStat:         diffStat,
			})

			// Now we stop tracking it in the second campaign
			campaignSpec3 := createCampaignSpec(t, ctx, store, "tracking-campaign", admin.ID)

			// Campaign should have 0 changesets after applying, but the
			// tracked changeset should not be closed, since the campaign is
			// not the owner.
			//
			verifyClosed := assertChangesetsClose(t, admin.ID)
			applyAndListChangesets(adminCtx, t, svc, campaignSpec3.RandID, 0)
			verifyClosed()
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
			_, changesets := applyAndListChangesets(adminCtx, t, svc, campaignSpec1.RandID, 1)

			// But the changeset was not published yet.
			// And now we apply a new spec without any changesets.
			campaignSpec2 := createCampaignSpec(t, ctx, store, "unpublished-changesets", admin.ID)

			// That should close no changesets, but leave the campaign with 0 changesets
			verifyClosed := assertChangesetsClose(t, admin.ID)
			applyAndListChangesets(adminCtx, t, svc, campaignSpec2.RandID, 0)
			verifyClosed()

			// And the unpublished changesets should be deleted
			toBeDeleted := changesets[0]
			_, err := store.GetChangeset(ctx, GetChangesetOpts{ID: toBeDeleted.ID})
			if err != ErrNoResults {
				t.Fatalf("expected changeset to be deleted but was not")
			}
		})

		t.Run("missing repository permissions", func(t *testing.T) {
			// Single repository filtered out by authzFilter
			ct.AuthzFilterRepos(t, repos[1].ID)

			campaignSpec := createCampaignSpec(t, ctx, store, "missing-permissions", admin.ID)

			createChangesetSpec(t, ctx, store, testSpecOpts{
				user:         admin.ID,
				repo:         repos[0].ID,
				campaignSpec: campaignSpec.ID,
				externalID:   "1234",
			})

			createChangesetSpec(t, ctx, store, testSpecOpts{
				user:         admin.ID,
				repo:         repos[1].ID, // Filtered out by authzFilter
				campaignSpec: campaignSpec.ID,
				headRef:      "refs/heads/my-branch",
			})

			_, err := svc.ApplyCampaign(adminCtx, ApplyCampaignOpts{
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

type changesetAssertions struct {
	repo             api.RepoID
	currentSpec      int64
	previousSpec     int64
	ownedByCampaign  int64
	reconcilerState  campaigns.ReconcilerState
	publicationState campaigns.ChangesetPublicationState
	externalID       string
	externalBranch   string
	diffStat         *diff.Stat
	unsynced         bool

	title string
	body  string

	failureMessage *string
}

func assertChangeset(t *testing.T, c *campaigns.Changeset, a changesetAssertions) {
	t.Helper()

	if c == nil {
		t.Fatalf("changeset is nil")
	}

	if have, want := c.RepoID, a.repo; have != want {
		t.Fatalf("changeset RepoID wrong. want=%d, have=%d", want, have)
	}

	if have, want := c.CurrentSpecID, a.currentSpec; have != want {
		t.Fatalf("changeset CurrentSpecID wrong. want=%d, have=%d", want, have)
	}

	if have, want := c.PreviousSpecID, a.previousSpec; have != want {
		t.Fatalf("changeset PreviousSpecID wrong. want=%d, have=%d", want, have)
	}

	if have, want := c.OwnedByCampaignID, a.ownedByCampaign; have != want {
		t.Fatalf("changeset OwnedByCampaignID wrong. want=%d, have=%d", want, have)
	}

	if have, want := c.ReconcilerState, a.reconcilerState; have != want {
		t.Fatalf("changeset ReconcilerState wrong. want=%s, have=%s", want, have)
	}

	if have, want := c.PublicationState, a.publicationState; have != want {
		t.Fatalf("changeset PublicationState wrong. want=%s, have=%s", want, have)
	}

	if have, want := c.ExternalID, a.externalID; have != want {
		t.Fatalf("changeset ExternalID wrong. want=%s, have=%s", want, have)
	}

	if have, want := c.ExternalBranch, a.externalBranch; have != want {
		t.Fatalf("changeset ExternalBranch wrong. want=%s, have=%s", want, have)
	}

	if want, have := a.failureMessage, c.FailureMessage; want == nil && have != nil {
		t.Fatalf("expected no failure message, but have=%q", *have)
	}

	if diff := cmp.Diff(a.diffStat, c.DiffStat()); diff != "" {
		t.Fatalf("changeset DiffStat wrong. (-want +got):\n%s", diff)
	}

	if diff := cmp.Diff(a.unsynced, c.Unsynced); diff != "" {
		t.Fatalf("changeset Unsynced wrong. (-want +got):\n%s", diff)
	}

	if want := c.FailureMessage; want != nil {
		if c.FailureMessage == nil {
			t.Fatalf("expected failure message %q but have none", *want)
		}
		if want, have := *a.failureMessage, *c.FailureMessage; have != want {
			t.Fatalf("wrong failure message. want=%q, have=%q", want, have)
		}
	}

	if have, want := c.ExternalBranch, a.externalBranch; have != want {
		t.Fatalf("changeset ExternalBranch wrong. want=%s, have=%s", want, have)
	}

	if want := a.title; want != "" {
		have, err := c.Title()
		if err != nil {
			t.Fatalf("changeset.Title failed: %s", err)
		}

		if have != want {
			t.Fatalf("changeset Title wrong. want=%s, have=%s", want, have)
		}
	}

	if want := a.body; want != "" {
		have, err := c.Body()
		if err != nil {
			t.Fatalf("changeset.Body failed: %s", err)
		}

		if have != want {
			t.Fatalf("changeset Body wrong. want=%s, have=%s", want, have)
		}
	}
}

func applyAndListChangesets(ctx context.Context, t *testing.T, svc *Service, campaignSpecRandID string, wantChangesets int) (*campaigns.Campaign, campaigns.Changesets) {
	campaign, err := svc.ApplyCampaign(ctx, ApplyCampaignOpts{
		CampaignSpecRandID: campaignSpecRandID,
	})
	if err != nil {
		t.Fatalf("failed to apply campaign: %s", err)
	}

	if campaign.ID == 0 {
		t.Fatalf("campaign ID is zero")
	}

	changesets, _, err := svc.store.ListChangesets(ctx, ListChangesetsOpts{CampaignID: campaign.ID})
	if err != nil {
		t.Fatal(err)
	}

	if have, want := len(changesets), wantChangesets; have != want {
		t.Fatalf("wrong number of changesets. want=%d, have=%d", want, have)
	}

	return campaign, changesets
}

func assertChangesetsClose(t *testing.T, wantActor int32, want ...*campaigns.Changeset) (verify func()) {
	t.Helper()

	closedCalled := false

	mockApplyCampaignCloseChangesets = func(ctx context.Context, toClose campaigns.Changesets) {
		closedCalled = true

		if a := actor.FromContext(ctx); a.UID != wantActor {
			t.Fatalf("wrong actor in context. want=%d, have=%d", wantActor, a.UID)
		}

		if have, want := len(toClose), len(want); have != want {
			t.Fatalf("closing wrong number of changesets. want=%d, have=%d", want, have)
		}
		closedByID := map[int64]bool{}
		for _, c := range toClose {
			closedByID[c.ID] = true
		}
		for _, c := range want {
			if _, ok := closedByID[c.ID]; !ok {
				t.Fatalf("expected changeset %d to be closed but was not", c.ID)
			}
		}
	}

	t.Cleanup(func() { mockApplyCampaignCloseChangesets = nil })

	verify = func() {
		if !closedCalled {
			t.Fatalf("expected CloseOpenChangesets to be called but was not")
		}
	}

	return verify
}

func setChangesetPublished(t *testing.T, ctx context.Context, s *Store, c *campaigns.Changeset, externalID, externalBranch string) {
	t.Helper()

	c.ExternalBranch = externalBranch
	c.ExternalID = externalID
	c.PublicationState = campaigns.ChangesetPublicationStatePublished
	c.ReconcilerState = campaigns.ReconcilerStateCompleted
	c.Unsynced = false

	if err := s.UpdateChangeset(ctx, c); err != nil {
		t.Fatalf("failed to update changeset: %s", err)
	}
}

type testSpecOpts struct {
	user         int32
	repo         api.RepoID
	campaignSpec int64

	// If this is non-blank, the changesetSpec will be an import/track spec for
	// the changeset with the given externalID in the given repo.
	externalID string

	// If this is set, the changesetSpec will be a "create commit on this
	// branch" changeset spec.
	headRef string

	// If this is set along with headRef, the changesetSpec will have published
	// set.
	published bool

	title         string
	body          string
	commitMessage string
	commitDiff    string
}

func createChangesetSpec(
	t *testing.T,
	ctx context.Context,
	store *Store,
	opts testSpecOpts,
) *campaigns.ChangesetSpec {
	t.Helper()

	spec := &campaigns.ChangesetSpec{
		UserID:         opts.user,
		RepoID:         opts.repo,
		CampaignSpecID: opts.campaignSpec,
		Spec: &campaigns.ChangesetSpecDescription{
			BaseRepository: graphqlbackend.MarshalRepositoryID(opts.repo),

			ExternalID: opts.externalID,
			HeadRef:    opts.headRef,
			Published:  opts.published,

			Title: opts.title,
			Body:  opts.body,

			Commits: []campaigns.GitCommitDescription{
				{
					Message: opts.commitMessage,
					Diff:    opts.commitDiff,
				},
			},
		},
		DiffStatAdded:   10,
		DiffStatChanged: 5,
		DiffStatDeleted: 2,
	}

	if err := store.CreateChangesetSpec(ctx, spec); err != nil {
		t.Fatal(err)
	}

	return spec
}

func createCampaignSpec(t *testing.T, ctx context.Context, store *Store, name string, userID int32) *campaigns.CampaignSpec {
	t.Helper()

	s := buildCampaignSpec(t, name, userID)
	if err := store.CreateCampaignSpec(ctx, s); err != nil {
		t.Fatal(err)
	}

	return s
}

func createCampaign(t *testing.T, ctx context.Context, store *Store, name string, userID int32, spec int64) *campaigns.Campaign {
	t.Helper()

	c := buildCampaign(t, store.Clock()(), name, userID, spec)

	if err := store.CreateCampaign(ctx, c); err != nil {
		t.Fatal(err)
	}

	return c
}

func buildChangesetSpec(
	t *testing.T,
	opts testSpecOpts,
) *campaigns.ChangesetSpec {
	t.Helper()

	spec := &campaigns.ChangesetSpec{
		UserID:         opts.user,
		RepoID:         opts.repo,
		CampaignSpecID: opts.campaignSpec,
		Spec: &campaigns.ChangesetSpecDescription{
			BaseRepository: graphqlbackend.MarshalRepositoryID(opts.repo),

			ExternalID: opts.externalID,
			HeadRef:    opts.headRef,
			Published:  opts.published,

			Title: opts.title,
			Body:  opts.body,

			Commits: []campaigns.GitCommitDescription{
				{
					Message: opts.commitMessage,
					Diff:    opts.commitDiff,
				},
			},
		},
		DiffStatAdded:   10,
		DiffStatChanged: 5,
		DiffStatDeleted: 2,
	}

	return spec
}

func buildCampaignSpec(t *testing.T, name string, userID int32) *campaigns.CampaignSpec {
	t.Helper()

	s := &campaigns.CampaignSpec{
		UserID:          userID,
		NamespaceUserID: userID,
		Spec: campaigns.CampaignSpecFields{
			Name:        name,
			Description: "the description",
			ChangesetTemplate: campaigns.ChangesetTemplate{
				Branch: "branch-name",
			},
		},
	}

	return s
}

func buildCampaign(t *testing.T, now time.Time, name string, userID int32, spec int64) *campaigns.Campaign {
	t.Helper()

	c := &campaigns.Campaign{
		InitialApplierID: userID,
		LastApplierID:    userID,
		LastAppliedAt:    now,
		NamespaceUserID:  userID,
		CampaignSpecID:   spec,
		Name:             name,
		Description:      "campaign description",
	}

	return c
}
