package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/batches"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func TestServicePermissionLevels(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := backend.WithAuthzBypass(context.Background())
	db := dbtesting.GetDB(t)

	s := store.New(db)
	svc := New(s)

	admin := ct.CreateTestUser(t, db, true)
	user := ct.CreateTestUser(t, db, false)
	otherUser := ct.CreateTestUser(t, db, false)

	rs, _ := ct.CreateTestRepos(t, ctx, db, 1)

	createTestData := func(t *testing.T, s *store.Store, svc *Service, author int32) (*batches.BatchChange, *batches.Changeset, *batches.BatchSpec) {
		spec := testCampaignSpec(author)
		if err := s.CreateBatchSpec(ctx, spec); err != nil {
			t.Fatal(err)
		}

		campaign := testCampaign(author, spec)
		if err := s.CreateBatchChange(ctx, campaign); err != nil {
			t.Fatal(err)
		}

		changeset := testChangeset(rs[0].ID, campaign.ID, batches.ChangesetExternalStateOpen)
		if err := s.CreateChangeset(ctx, changeset); err != nil {
			t.Fatal(err)
		}

		return campaign, changeset, spec
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
			campaign, changeset, campaignSpec := createTestData(t, s, svc, tc.campaignAuthor)
			// Fresh context.Background() because the previous one is wrapped in AuthzBypas
			currentUserCtx := actor.WithActor(context.Background(), actor.FromUser(tc.currentUser))

			t.Run("EnqueueChangesetSync", func(t *testing.T) {
				// The cases that don't result in auth errors will fall through
				// to call repoupdater.EnqueueChangesetSync, so we need to
				// ensure we mock that call to avoid unexpected network calls.
				repoupdater.MockEnqueueChangesetSync = func(ctx context.Context, ids []int64) error {
					return nil
				}
				t.Cleanup(func() { repoupdater.MockEnqueueChangesetSync = nil })

				err := svc.EnqueueChangesetSync(currentUserCtx, changeset.ID)
				tc.assertFunc(t, err)
			})

			t.Run("ReenqueueChangeset", func(t *testing.T) {
				_, _, err := svc.ReenqueueChangeset(currentUserCtx, changeset.ID)
				tc.assertFunc(t, err)
			})

			t.Run("CloseCampaign", func(t *testing.T) {
				_, err := svc.CloseBatchChange(currentUserCtx, campaign.ID, false)
				tc.assertFunc(t, err)
			})

			t.Run("DeleteCampaign", func(t *testing.T) {
				err := svc.DeleteBatchChange(currentUserCtx, campaign.ID)
				tc.assertFunc(t, err)
			})

			t.Run("MoveCampaign", func(t *testing.T) {
				_, err := svc.MoveBatchChange(currentUserCtx, MoveBatchChangeOpts{
					BatchChangeID: campaign.ID,
					NewName:       "foobar2",
				})
				tc.assertFunc(t, err)
			})

			t.Run("ApplyCampaign", func(t *testing.T) {
				_, err := svc.ApplyBatchChange(currentUserCtx, ApplyBatchChangeOpts{
					BatchSpecRandID: campaignSpec.RandID,
				})
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
	db := dbtesting.GetDB(t)

	admin := ct.CreateTestUser(t, db, true)
	user := ct.CreateTestUser(t, db, false)

	now := timeutil.Now()
	clock := func() time.Time { return now }

	s := store.NewWithClock(db, clock)
	rs, _ := ct.CreateTestRepos(t, ctx, db, 4)

	fakeSource := &ct.FakeChangesetSource{}
	sourcer := repos.NewFakeSourcer(nil, fakeSource)

	svc := New(s)
	svc.sourcer = sourcer

	t.Run("DeleteCampaign", func(t *testing.T) {
		spec := testCampaignSpec(admin.ID)
		if err := s.CreateBatchSpec(ctx, spec); err != nil {
			t.Fatal(err)
		}

		campaign := testCampaign(admin.ID, spec)
		if err := s.CreateBatchChange(ctx, campaign); err != nil {
			t.Fatal(err)
		}
		if err := svc.DeleteBatchChange(ctx, campaign.ID); err != nil {
			t.Fatalf("campaign not deleted: %s", err)
		}

		_, err := s.GetBatchChange(ctx, store.CountBatchChangeOpts{ID: campaign.ID})
		if err != nil && err != store.ErrNoResults {
			t.Fatalf("want campaign to be deleted, but was not: %e", err)
		}
	})

	t.Run("CloseCampaign", func(t *testing.T) {
		createCampaign := func(t *testing.T) *batches.BatchChange {
			t.Helper()

			spec := testCampaignSpec(admin.ID)
			if err := s.CreateBatchSpec(ctx, spec); err != nil {
				t.Fatal(err)
			}

			campaign := testCampaign(admin.ID, spec)
			if err := s.CreateBatchChange(ctx, campaign); err != nil {
				t.Fatal(err)
			}
			return campaign
		}

		adminCtx := actor.WithActor(context.Background(), actor.FromUser(admin.ID))

		closeConfirm := func(t *testing.T, c *batches.BatchChange, closeChangesets bool) {
			t.Helper()

			closedCampaign, err := svc.CloseBatchChange(adminCtx, c.ID, closeChangesets)
			if err != nil {
				t.Fatalf("campaign not closed: %s", err)
			}
			if !closedCampaign.ClosedAt.Equal(now) {
				t.Fatalf("campaign ClosedAt is zero")
			}

			if !closeChangesets {
				return
			}

			cs, _, err := s.ListChangesets(ctx, store.ListChangesetsOpts{
				OwnedByCampaignID: c.ID,
			})
			if err != nil {
				t.Fatalf("listing changesets failed: %s", err)
			}
			for _, c := range cs {
				if !c.Closing {
					t.Errorf("changeset should be Closing, but is not")
				}

				if have, want := c.ReconcilerState, batches.ReconcilerStateQueued; have != want {
					t.Errorf("changeset ReconcilerState wrong. want=%s, have=%s", want, have)
				}
			}
		}

		t.Run("no changesets", func(t *testing.T) {
			campaign := createCampaign(t)
			closeConfirm(t, campaign, false)
		})

		t.Run("changesets", func(t *testing.T) {
			campaign := createCampaign(t)

			changeset1 := testChangeset(rs[0].ID, campaign.ID, batches.ChangesetExternalStateOpen)
			changeset1.ReconcilerState = batches.ReconcilerStateCompleted
			if err := s.CreateChangeset(ctx, changeset1); err != nil {
				t.Fatal(err)
			}

			changeset2 := testChangeset(rs[1].ID, campaign.ID, batches.ChangesetExternalStateOpen)
			changeset2.ReconcilerState = batches.ReconcilerStateCompleted
			if err := s.CreateChangeset(ctx, changeset2); err != nil {
				t.Fatal(err)
			}

			closeConfirm(t, campaign, true)
		})
	})

	t.Run("EnqueueChangesetSync", func(t *testing.T) {
		spec := testCampaignSpec(admin.ID)
		if err := s.CreateBatchSpec(ctx, spec); err != nil {
			t.Fatal(err)
		}

		campaign := testCampaign(admin.ID, spec)
		if err := s.CreateBatchChange(ctx, campaign); err != nil {
			t.Fatal(err)
		}

		changeset := testChangeset(rs[0].ID, campaign.ID, batches.ChangesetExternalStateOpen)
		if err := s.CreateChangeset(ctx, changeset); err != nil {
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

		// rs[0] is filtered out
		ct.MockRepoPermissions(t, db, user.ID, rs[1].ID, rs[2].ID, rs[3].ID)

		// should result in a not found error
		if err := svc.EnqueueChangesetSync(ctx, changeset.ID); !errcode.IsNotFound(err) {
			t.Fatalf("expected not-found error but got %v", err)
		}
	})

	t.Run("ReenqueueChangeset", func(t *testing.T) {
		spec := testCampaignSpec(admin.ID)
		if err := s.CreateBatchSpec(ctx, spec); err != nil {
			t.Fatal(err)
		}

		campaign := testCampaign(admin.ID, spec)
		if err := s.CreateBatchChange(ctx, campaign); err != nil {
			t.Fatal(err)
		}

		changeset := testChangeset(rs[0].ID, campaign.ID, batches.ChangesetExternalStateOpen)
		if err := s.CreateChangeset(ctx, changeset); err != nil {
			t.Fatal(err)
		}

		ct.SetChangesetFailed(t, ctx, s, changeset)

		if _, _, err := svc.ReenqueueChangeset(ctx, changeset.ID); err != nil {
			t.Fatal(err)
		}

		ct.ReloadAndAssertChangeset(t, ctx, s, changeset, ct.ChangesetAssertions{
			Repo:          rs[0].ID,
			ExternalState: batches.ChangesetExternalStateOpen,
			ExternalID:    "ext-id-5",
			AttachedTo:    []int64{campaign.ID},

			// The important fields:
			ReconcilerState: batches.ReconcilerStateQueued,
			NumResets:       0,
			NumFailures:     0,
			FailureMessage:  nil,
		})

		// rs[0] is filtered out
		ct.MockRepoPermissions(t, db, user.ID, rs[1].ID, rs[2].ID, rs[3].ID)

		// should result in a not found error
		if _, _, err := svc.ReenqueueChangeset(ctx, changeset.ID); !errcode.IsNotFound(err) {
			t.Fatalf("expected not-found error but got %v", err)
		}
	})

	t.Run("CreateCampaignSpec", func(t *testing.T) {
		changesetSpecs := make([]*batches.ChangesetSpec, 0, len(rs))
		changesetSpecRandIDs := make([]string, 0, len(rs))
		for _, r := range rs {
			cs := &batches.ChangesetSpec{RepoID: r.ID, UserID: admin.ID}
			if err := s.CreateChangesetSpec(ctx, cs); err != nil {
				t.Fatal(err)
			}
			changesetSpecs = append(changesetSpecs, cs)
			changesetSpecRandIDs = append(changesetSpecRandIDs, cs.RandID)
		}

		adminCtx := actor.WithActor(context.Background(), actor.FromUser(admin.ID))
		userCtx := actor.WithActor(context.Background(), actor.FromUser(user.ID))

		t.Run("success", func(t *testing.T) {
			opts := CreateBatchSpecOpts{
				NamespaceUserID:      admin.ID,
				RawSpec:              ct.TestRawBatchSpec,
				ChangesetSpecRandIDs: changesetSpecRandIDs,
			}

			spec, err := svc.CreateBatchSpec(adminCtx, opts)
			if err != nil {
				t.Fatal(err)
			}

			if spec.ID == 0 {
				t.Fatalf("CampaignSpec ID is 0")
			}

			if have, want := spec.UserID, admin.ID; have != want {
				t.Fatalf("UserID is %d, want %d", have, want)
			}

			var wantFields batches.BatchSpecFields
			if err := json.Unmarshal([]byte(spec.RawSpec), &wantFields); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(wantFields, spec.Spec); diff != "" {
				t.Fatalf("wrong spec fields (-want +got):\n%s", diff)
			}

			for _, cs := range changesetSpecs {
				cs2, err := s.GetChangesetSpec(ctx, store.GetChangesetSpecOpts{ID: cs.ID})
				if err != nil {
					t.Fatal(err)
				}

				if have, want := cs2.CampaignSpecID, spec.ID; have != want {
					t.Fatalf("changesetSpec has wrong CampaignSpecID. want=%d, have=%d", want, have)
				}
			}
		})

		t.Run("success with YAML raw spec", func(t *testing.T) {
			opts := CreateBatchSpecOpts{
				NamespaceUserID: admin.ID,
				RawSpec:         ct.TestRawBatchSpecYAML,
			}

			spec, err := svc.CreateBatchSpec(adminCtx, opts)
			if err != nil {
				t.Fatal(err)
			}

			if spec.ID == 0 {
				t.Fatalf("CampaignSpec ID is 0")
			}

			var wantFields batches.BatchSpecFields
			if err := json.Unmarshal([]byte(ct.TestRawBatchSpec), &wantFields); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(wantFields, spec.Spec); diff != "" {
				t.Fatalf("wrong spec fields (-want +got):\n%s", diff)
			}
		})

		t.Run("missing repository permissions", func(t *testing.T) {
			ct.MockRepoPermissions(t, db, user.ID)

			opts := CreateBatchSpecOpts{
				NamespaceUserID:      user.ID,
				RawSpec:              ct.TestRawBatchSpec,
				ChangesetSpecRandIDs: changesetSpecRandIDs,
			}

			if _, err := svc.CreateBatchSpec(userCtx, opts); !errcode.IsNotFound(err) {
				t.Fatalf("expected not-found error but got %s", err)
			}
		})

		t.Run("invalid changesetspec id", func(t *testing.T) {
			containsInvalidID := []string{changesetSpecRandIDs[0], "foobar"}
			opts := CreateBatchSpecOpts{
				NamespaceUserID:      admin.ID,
				RawSpec:              ct.TestRawBatchSpec,
				ChangesetSpecRandIDs: containsInvalidID,
			}

			if _, err := svc.CreateBatchSpec(adminCtx, opts); !errcode.IsNotFound(err) {
				t.Fatalf("expected not-found error but got %s", err)
			}
		})

		t.Run("namespace user is not admin and not creator", func(t *testing.T) {
			userCtx := actor.WithActor(context.Background(), actor.FromUser(user.ID))

			opts := CreateBatchSpecOpts{
				NamespaceUserID: admin.ID,
				RawSpec:         ct.TestRawBatchSpecYAML,
			}

			_, err := svc.CreateBatchSpec(userCtx, opts)
			if !errcode.IsUnauthorized(err) {
				t.Fatalf("expected unauthorized error but got %s", err)
			}

			// Try again as admin
			adminCtx := actor.WithActor(context.Background(), actor.FromUser(admin.ID))

			opts.NamespaceUserID = user.ID

			_, err = svc.CreateBatchSpec(adminCtx, opts)
			if err != nil {
				t.Fatalf("expected no error but got %s", err)
			}
		})

		t.Run("missing access to namespace org", func(t *testing.T) {
			orgID := ct.InsertTestOrg(t, db, "test-org")

			opts := CreateBatchSpecOpts{
				NamespaceOrgID:       orgID,
				RawSpec:              ct.TestRawBatchSpec,
				ChangesetSpecRandIDs: changesetSpecRandIDs,
			}

			userCtx := actor.WithActor(context.Background(), actor.FromUser(user.ID))

			_, err := svc.CreateBatchSpec(userCtx, opts)
			if have, want := err, backend.ErrNotAnOrgMember; have != want {
				t.Fatalf("expected %s error but got %s", want, have)
			}

			// Create org membership and try again
			if _, err := database.OrgMembers(db).Create(ctx, orgID, user.ID); err != nil {
				t.Fatal(err)
			}

			_, err = svc.CreateBatchSpec(userCtx, opts)
			if err != nil {
				t.Fatalf("expected no error but got %s", err)
			}
		})

		t.Run("no side-effects if no changeset spec IDs are given", func(t *testing.T) {
			// We already have ChangesetSpecs in the database. Here we
			// want to make sure that the new CampaignSpec is created,
			// without accidently attaching the existing ChangesetSpecs.
			opts := CreateBatchSpecOpts{
				NamespaceUserID:      admin.ID,
				RawSpec:              ct.TestRawBatchSpec,
				ChangesetSpecRandIDs: []string{},
			}

			spec, err := svc.CreateBatchSpec(adminCtx, opts)
			if err != nil {
				t.Fatal(err)
			}

			countOpts := store.CountChangesetSpecsOpts{CampaignSpecID: spec.ID}
			count, err := s.CountChangesetSpecs(adminCtx, countOpts)
			if err != nil {
				return
			}
			if count != 0 {
				t.Fatalf("want no changeset specs attached to campaign spec, but have %d", count)
			}
		})
	})

	t.Run("CreateChangesetSpec", func(t *testing.T) {
		repo := rs[0]
		rawSpec := ct.NewRawChangesetSpecGitBranch(graphqlbackend.MarshalRepositoryID(repo.ID), "d34db33f")

		t.Run("success", func(t *testing.T) {
			spec, err := svc.CreateChangesetSpec(ctx, rawSpec, admin.ID)
			if err != nil {
				t.Fatal(err)
			}

			if spec.ID == 0 {
				t.Fatalf("ChangesetSpec ID is 0")
			}

			wantFields := &batches.ChangesetSpecDescription{}
			if err := json.Unmarshal([]byte(spec.RawSpec), wantFields); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(wantFields, spec.Spec); diff != "" {
				t.Fatalf("wrong spec fields (-want +got):\n%s", diff)
			}

			wantDiffStat := *ct.ChangesetSpecDiffStat
			if diff := cmp.Diff(wantDiffStat, spec.DiffStat()); diff != "" {
				t.Fatalf("wrong diff stat (-want +got):\n%s", diff)
			}
		})

		t.Run("invalid raw spec", func(t *testing.T) {
			invalidRaw := `{"externalComputer": "beepboop"}`
			_, err := svc.CreateChangesetSpec(ctx, invalidRaw, admin.ID)
			if err == nil {
				t.Fatal("expected error but got nil")
			}

			haveErr := fmt.Sprintf("%v", err)
			wantErr := "4 errors occurred:\n\t* Must validate one and only one schema (oneOf)\n\t* baseRepository is required\n\t* externalID is required\n\t* Additional property externalComputer is not allowed\n\n"
			if diff := cmp.Diff(wantErr, haveErr); diff != "" {
				t.Fatalf("unexpected error (-want +got):\n%s", diff)
			}
		})

		t.Run("missing repository permissions", func(t *testing.T) {
			ct.MockRepoPermissions(t, db, user.ID, rs[1].ID, rs[2].ID, rs[3].ID)

			_, err := svc.CreateChangesetSpec(ctx, rawSpec, admin.ID)
			if !errcode.IsNotFound(err) {
				t.Fatalf("expected not-found error but got %v", err)
			}
		})
	})

	t.Run("ApplyCampaign", func(t *testing.T) {
		// See TestServiceApplyCampaign
	})

	t.Run("MoveCampaign", func(t *testing.T) {
		createBatchChange := func(t *testing.T, name string, authorID, userID, orgID int32) *batches.BatchChange {
			t.Helper()

			spec := &batches.BatchSpec{
				UserID:          authorID,
				NamespaceUserID: userID,
				NamespaceOrgID:  orgID,
			}

			if err := s.CreateBatchSpec(ctx, spec); err != nil {
				t.Fatal(err)
			}

			c := &batches.BatchChange{
				InitialApplierID: authorID,
				NamespaceUserID:  userID,
				NamespaceOrgID:   orgID,
				Name:             name,
				LastApplierID:    authorID,
				LastAppliedAt:    time.Now(),
				BatchSpecID:      spec.ID,
			}

			if err := s.CreateBatchChange(ctx, c); err != nil {
				t.Fatal(err)
			}

			return c
		}

		t.Run("new name", func(t *testing.T) {
			campaign := createBatchChange(t, "old-name", admin.ID, admin.ID, 0)

			opts := MoveBatchChangeOpts{BatchChangeID: campaign.ID, NewName: "new-name"}
			moved, err := svc.MoveBatchChange(ctx, opts)
			if err != nil {
				t.Fatal(err)
			}

			if have, want := moved.Name, opts.NewName; have != want {
				t.Fatalf("wrong name. want=%q, have=%q", want, have)
			}
		})

		t.Run("new user namespace", func(t *testing.T) {
			campaign := createBatchChange(t, "old-name", admin.ID, admin.ID, 0)

			user2 := ct.CreateTestUser(t, db, false)

			opts := MoveBatchChangeOpts{BatchChangeID: campaign.ID, NewNamespaceUserID: user2.ID}
			moved, err := svc.MoveBatchChange(ctx, opts)
			if err != nil {
				t.Fatal(err)
			}

			if have, want := moved.NamespaceUserID, opts.NewNamespaceUserID; have != want {
				t.Fatalf("wrong NamespaceUserID. want=%d, have=%d", want, have)
			}

			if have, want := moved.NamespaceOrgID, opts.NewNamespaceOrgID; have != want {
				t.Fatalf("wrong NamespaceOrgID. want=%d, have=%d", want, have)
			}
		})

		t.Run("new user namespace but current user is not admin", func(t *testing.T) {
			campaign := createBatchChange(t, "old-name", user.ID, user.ID, 0)

			user2 := ct.CreateTestUser(t, db, false)

			opts := MoveBatchChangeOpts{BatchChangeID: campaign.ID, NewNamespaceUserID: user2.ID}

			userCtx := actor.WithActor(context.Background(), actor.FromUser(user.ID))
			_, err := svc.MoveBatchChange(userCtx, opts)
			if !errcode.IsUnauthorized(err) {
				t.Fatalf("expected unauthorized error but got %s", err)
			}
		})

		t.Run("new org namespace", func(t *testing.T) {
			campaign := createBatchChange(t, "old-name", admin.ID, admin.ID, 0)

			orgID := ct.InsertTestOrg(t, db, "org")

			opts := MoveBatchChangeOpts{BatchChangeID: campaign.ID, NewNamespaceOrgID: orgID}
			moved, err := svc.MoveBatchChange(ctx, opts)
			if err != nil {
				t.Fatal(err)
			}

			if have, want := moved.NamespaceUserID, opts.NewNamespaceUserID; have != want {
				t.Fatalf("wrong NamespaceUserID. want=%d, have=%d", want, have)
			}

			if have, want := moved.NamespaceOrgID, opts.NewNamespaceOrgID; have != want {
				t.Fatalf("wrong NamespaceOrgID. want=%d, have=%d", want, have)
			}
		})

		t.Run("new org namespace but current user is missing access", func(t *testing.T) {
			campaign := createBatchChange(t, "old-name", user.ID, user.ID, 0)

			orgID := ct.InsertTestOrg(t, db, "org-no-access")

			opts := MoveBatchChangeOpts{BatchChangeID: campaign.ID, NewNamespaceOrgID: orgID}

			userCtx := actor.WithActor(context.Background(), actor.FromUser(user.ID))
			_, err := svc.MoveBatchChange(userCtx, opts)
			if have, want := err, backend.ErrNotAnOrgMember; have != want {
				t.Fatalf("expected %s error but got %s", want, have)
			}
		})
	})

	t.Run("GetBatchChangeMatchingBatchSpec", func(t *testing.T) {
		campaignSpec := ct.CreateBatchSpec(t, ctx, s, "matching-campaign-spec", admin.ID)

		haveCampaign, err := svc.GetBatchChangeMatchingBatchSpec(ctx, campaignSpec)
		if err != nil {
			t.Fatalf("unexpected error: %s\n", err)
		}
		if haveCampaign != nil {
			t.Fatalf("expected campaign to be nil, but is not: %+v\n", haveCampaign)
		}

		matchingCampaign := &batches.BatchChange{
			Name:             campaignSpec.Spec.Name,
			Description:      campaignSpec.Spec.Description,
			InitialApplierID: admin.ID,
			NamespaceOrgID:   campaignSpec.NamespaceOrgID,
			NamespaceUserID:  campaignSpec.NamespaceUserID,
			BatchSpecID:      campaignSpec.ID,
			LastApplierID:    admin.ID,
			LastAppliedAt:    time.Now(),
		}
		if err := s.CreateBatchChange(ctx, matchingCampaign); err != nil {
			t.Fatalf("failed to create campaign: %s\n", err)
		}

		haveCampaign, err = svc.GetBatchChangeMatchingBatchSpec(ctx, campaignSpec)
		if err != nil {
			t.Fatalf("unexpected error: %s\n", err)
		}
		if haveCampaign == nil {
			t.Fatalf("expected to have matching campaign, but got nil")
		}

		if diff := cmp.Diff(matchingCampaign, haveCampaign); diff != "" {
			t.Fatalf("wrong campaign was matched (-want +got):\n%s", diff)
		}
	})

	t.Run("GetNewestCampaignSpec", func(t *testing.T) {
		older := ct.CreateBatchSpec(t, ctx, s, "superseding", user.ID)
		newer := ct.CreateBatchSpec(t, ctx, s, "superseding", user.ID)

		for name, in := range map[string]*batches.BatchSpec{
			"older": older,
			"newer": newer,
		} {
			t.Run(name, func(t *testing.T) {
				have, err := svc.GetNewestBatchSpec(ctx, s, in, user.ID)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if diff := cmp.Diff(newer, have); diff != "" {
					t.Errorf("unexpected newer campaign spec (-want +have):\n%s", diff)
				}
			})
		}

		t.Run("different user", func(t *testing.T) {
			have, err := svc.GetNewestBatchSpec(ctx, s, older, admin.ID)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if have != nil {
				t.Errorf("unexpected non-nil campaign spec: %+v", have)
			}
		})
	})

	t.Run("FetchUsernameForBitbucketServerToken", func(t *testing.T) {
		fakeSource := &ct.FakeChangesetSource{Username: "my-bbs-username"}
		sourcer := repos.NewFakeSourcer(nil, fakeSource)

		// Create a fresh service for this test as to not mess with state
		// possibly used by other tests.
		testSvc := New(s)
		testSvc.sourcer = sourcer

		rs, _ := ct.CreateBbsTestRepos(t, ctx, db, 1)
		repo := rs[0]

		url := repo.ExternalRepo.ServiceID
		extType := repo.ExternalRepo.ServiceType

		username, err := testSvc.FetchUsernameForBitbucketServerToken(ctx, url, extType, "my-token")
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		if !fakeSource.AuthenticatedUsernameCalled {
			t.Errorf("service didn't call AuthenticatedUsername")
		}

		if have, want := username, fakeSource.Username; have != want {
			t.Errorf("wrong username returned. want=%q, have=%q", want, have)
		}
	})
}

func testCampaign(user int32, spec *batches.BatchSpec) *batches.BatchChange {
	c := &batches.BatchChange{
		Name:             "test-campaign",
		InitialApplierID: user,
		NamespaceUserID:  user,
		BatchSpecID:      spec.ID,
		LastApplierID:    user,
		LastAppliedAt:    time.Now(),
	}

	return c
}

func testCampaignSpec(user int32) *batches.BatchSpec {
	return &batches.BatchSpec{
		UserID:          user,
		NamespaceUserID: user,
	}
}

func testChangeset(repoID api.RepoID, campaign int64, extState batches.ChangesetExternalState) *batches.Changeset {
	changeset := &batches.Changeset{
		RepoID:              repoID,
		ExternalServiceType: extsvc.TypeGitHub,
		ExternalID:          fmt.Sprintf("ext-id-%d", campaign),
		Metadata:            &github.PullRequest{State: string(extState), CreatedAt: time.Now()},
		ExternalState:       extState,
	}

	if campaign != 0 {
		changeset.BatchChanges = []batches.BatchChangeAssoc{{BatchChangeID: campaign}}
	}

	return changeset
}
