package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	stesting "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/sources/testing"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	bt "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestServicePermissionLevels(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := actor.WithInternalActor(context.Background())
	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	s := store.New(db, &observation.TestContext, nil)
	svc := New(s)

	admin := bt.CreateTestUser(t, db, true)
	user := bt.CreateTestUser(t, db, false)
	otherUser := bt.CreateTestUser(t, db, false)

	repo, _ := bt.CreateTestRepo(t, ctx, db)

	createTestData := func(t *testing.T, s *store.Store, author int32) (*btypes.BatchChange, *btypes.Changeset, *btypes.BatchSpec) {
		spec := testBatchSpec(author)
		if err := s.CreateBatchSpec(ctx, spec); err != nil {
			t.Fatal(err)
		}

		batchChange := testBatchChange(author, spec)
		if err := s.CreateBatchChange(ctx, batchChange); err != nil {
			t.Fatal(err)
		}

		changeset := testChangeset(repo.ID, batchChange.ID, btypes.ChangesetExternalStateOpen)
		if err := s.CreateChangeset(ctx, changeset); err != nil {
			t.Fatal(err)
		}

		return batchChange, changeset, spec
	}

	tests := []struct {
		name              string
		batchChangeAuthor int32
		currentUser       int32
		assertFunc        func(t *testing.T, err error)
	}{
		{
			name:              "unauthorized user",
			batchChangeAuthor: user.ID,
			currentUser:       otherUser.ID,
			assertFunc:        assertAuthError,
		},
		{
			name:              "batch change author",
			batchChangeAuthor: user.ID,
			currentUser:       user.ID,
			assertFunc:        assertNoAuthError,
		},

		{
			name:              "site-admin",
			batchChangeAuthor: user.ID,
			currentUser:       admin.ID,
			assertFunc:        assertNoAuthError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			batchChange, changeset, batchSpec := createTestData(t, s, tc.batchChangeAuthor)
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

			t.Run("CloseBatchChange", func(t *testing.T) {
				_, err := svc.CloseBatchChange(currentUserCtx, batchChange.ID, false)
				tc.assertFunc(t, err)
			})

			t.Run("DeleteBatchChange", func(t *testing.T) {
				err := svc.DeleteBatchChange(currentUserCtx, batchChange.ID)
				tc.assertFunc(t, err)
			})

			t.Run("MoveBatchChange", func(t *testing.T) {
				_, err := svc.MoveBatchChange(currentUserCtx, MoveBatchChangeOpts{
					BatchChangeID: batchChange.ID,
					NewName:       "foobar2",
				})
				tc.assertFunc(t, err)
			})

			t.Run("ApplyBatchChange", func(t *testing.T) {
				_, err := svc.ApplyBatchChange(currentUserCtx, ApplyBatchChangeOpts{
					BatchSpecRandID: batchSpec.RandID,
				})
				tc.assertFunc(t, err)
			})

			t.Run("CreateChangesetJobs", func(t *testing.T) {
				_, err := svc.CreateChangesetJobs(currentUserCtx, batchChange.ID, []int64{changeset.ID}, btypes.ChangesetJobTypeComment, btypes.ChangesetJobCommentPayload{Message: "test"}, store.ListChangesetsOpts{})
				tc.assertFunc(t, err)
			})

			t.Run("ExecuteBatchSpec", func(t *testing.T) {
				_, err := svc.ExecuteBatchSpec(currentUserCtx, ExecuteBatchSpecOpts{
					BatchSpecRandID: batchSpec.RandID,
				})
				tc.assertFunc(t, err)
			})

			t.Run("ReplaceBatchSpecInput", func(t *testing.T) {
				_, err := svc.ReplaceBatchSpecInput(currentUserCtx, ReplaceBatchSpecInputOpts{
					BatchSpecRandID: batchSpec.RandID,
					RawSpec:         bt.TestRawBatchSpecYAML,
				})
				tc.assertFunc(t, err)
			})

			t.Run("UpsertBatchSpecInput", func(t *testing.T) {
				_, err := svc.UpsertBatchSpecInput(currentUserCtx, UpsertBatchSpecInputOpts{
					RawSpec:         bt.TestRawBatchSpecYAML,
					NamespaceUserID: tc.batchChangeAuthor,
				})
				tc.assertFunc(t, err)
			})

			t.Run("CreateBatchSpecFromRaw", func(t *testing.T) {
				_, err := svc.CreateBatchSpecFromRaw(currentUserCtx, CreateBatchSpecFromRawOpts{
					RawSpec:         bt.TestRawBatchSpecYAML,
					NamespaceUserID: tc.batchChangeAuthor,
					BatchChange:     batchChange.ID,
				})
				tc.assertFunc(t, err)
			})

			t.Run("CancelBatchSpec", func(t *testing.T) {
				_, err := svc.CancelBatchSpec(currentUserCtx, CancelBatchSpecOpts{
					BatchSpecRandID: batchSpec.RandID,
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

	logger := logtest.Scoped(t)
	ctx := actor.WithInternalActor(context.Background())
	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	admin := bt.CreateTestUser(t, db, true)
	user := bt.CreateTestUser(t, db, false)
	user2 := bt.CreateTestUser(t, db, false)

	adminCtx := actor.WithActor(context.Background(), actor.FromUser(admin.ID))
	userCtx := actor.WithActor(context.Background(), actor.FromUser(user.ID))
	user2Ctx := actor.WithActor(context.Background(), actor.FromUser(user2.ID))

	now := timeutil.Now()
	clock := func() time.Time { return now }

	s := store.NewWithClock(db, &observation.TestContext, nil, clock)
	rs, _ := bt.CreateTestRepos(t, ctx, db, 4)

	fakeSource := &stesting.FakeChangesetSource{}
	sourcer := stesting.NewFakeSourcer(nil, fakeSource)

	svc := New(s)
	svc.sourcer = sourcer

	t.Run("DeleteBatchChange", func(t *testing.T) {
		spec := testBatchSpec(admin.ID)
		if err := s.CreateBatchSpec(ctx, spec); err != nil {
			t.Fatal(err)
		}

		batchChange := testBatchChange(admin.ID, spec)
		if err := s.CreateBatchChange(ctx, batchChange); err != nil {
			t.Fatal(err)
		}
		if err := svc.DeleteBatchChange(ctx, batchChange.ID); err != nil {
			t.Fatalf("batch change not deleted: %s", err)
		}

		_, err := s.GetBatchChange(ctx, store.GetBatchChangeOpts{ID: batchChange.ID})
		if err != nil && err != store.ErrNoResults {
			t.Fatalf("want batch change to be deleted, but was not: %e", err)
		}
	})

	t.Run("CloseBatchChange", func(t *testing.T) {
		createBatchChange := func(t *testing.T) *btypes.BatchChange {
			t.Helper()

			spec := testBatchSpec(admin.ID)
			if err := s.CreateBatchSpec(ctx, spec); err != nil {
				t.Fatal(err)
			}

			batchChange := testBatchChange(admin.ID, spec)
			if err := s.CreateBatchChange(ctx, batchChange); err != nil {
				t.Fatal(err)
			}
			return batchChange
		}

		closeConfirm := func(t *testing.T, c *btypes.BatchChange, closeChangesets bool) {
			t.Helper()

			closedBatchChange, err := svc.CloseBatchChange(adminCtx, c.ID, closeChangesets)
			if err != nil {
				t.Fatalf("batch change not closed: %s", err)
			}
			if !closedBatchChange.ClosedAt.Equal(now) {
				t.Fatalf("batch change ClosedAt is zero")
			}

			if !closeChangesets {
				return
			}

			cs, _, err := s.ListChangesets(ctx, store.ListChangesetsOpts{
				OwnedByBatchChangeID: c.ID,
			})
			if err != nil {
				t.Fatalf("listing changesets failed: %s", err)
			}
			for _, c := range cs {
				if !c.Closing {
					t.Errorf("changeset should be Closing, but is not")
				}

				if have, want := c.ReconcilerState, btypes.ReconcilerStateQueued; have != want {
					t.Errorf("changeset ReconcilerState wrong. want=%s, have=%s", want, have)
				}
			}
		}

		t.Run("no changesets", func(t *testing.T) {
			batchChange := createBatchChange(t)
			closeConfirm(t, batchChange, false)
		})

		t.Run("changesets", func(t *testing.T) {
			batchChange := createBatchChange(t)

			changeset1 := testChangeset(rs[0].ID, batchChange.ID, btypes.ChangesetExternalStateOpen)
			changeset1.ReconcilerState = btypes.ReconcilerStateCompleted
			if err := s.CreateChangeset(ctx, changeset1); err != nil {
				t.Fatal(err)
			}

			changeset2 := testChangeset(rs[1].ID, batchChange.ID, btypes.ChangesetExternalStateOpen)
			changeset2.ReconcilerState = btypes.ReconcilerStateCompleted
			if err := s.CreateChangeset(ctx, changeset2); err != nil {
				t.Fatal(err)
			}

			closeConfirm(t, batchChange, true)
		})
	})

	t.Run("EnqueueChangesetSync", func(t *testing.T) {
		spec := testBatchSpec(user.ID)
		if err := s.CreateBatchSpec(ctx, spec); err != nil {
			t.Fatal(err)
		}

		batchChange := testBatchChange(user.ID, spec)
		if err := s.CreateBatchChange(ctx, batchChange); err != nil {
			t.Fatal(err)
		}

		changeset := testChangeset(rs[0].ID, batchChange.ID, btypes.ChangesetExternalStateOpen)
		if err := s.CreateChangeset(ctx, changeset); err != nil {
			t.Fatal(err)
		}

		called := false
		repoupdater.MockEnqueueChangesetSync = func(_ context.Context, ids []int64) error {
			if len(ids) != 1 && ids[0] != changeset.ID {
				t.Fatalf("MockEnqueueChangesetSync received wrong ids: %+v", ids)
			}
			called = true
			return nil
		}
		t.Cleanup(func() { repoupdater.MockEnqueueChangesetSync = nil })

		if err := svc.EnqueueChangesetSync(userCtx, changeset.ID); err != nil {
			t.Fatal(err)
		}

		if !called {
			t.Fatal("MockEnqueueChangesetSync not called")
		}

		// rs[0] is filtered out
		bt.MockRepoPermissions(t, db, user.ID, rs[1].ID, rs[2].ID, rs[3].ID)

		// should result in a not found error
		if err := svc.EnqueueChangesetSync(userCtx, changeset.ID); !errcode.IsNotFound(err) {
			t.Fatalf("expected not-found error but got %v", err)
		}
	})

	t.Run("ReenqueueChangeset", func(t *testing.T) {
		spec := testBatchSpec(user.ID)
		if err := s.CreateBatchSpec(ctx, spec); err != nil {
			t.Fatal(err)
		}

		batchChange := testBatchChange(user.ID, spec)
		if err := s.CreateBatchChange(ctx, batchChange); err != nil {
			t.Fatal(err)
		}

		changeset := testChangeset(rs[0].ID, batchChange.ID, btypes.ChangesetExternalStateOpen)
		if err := s.CreateChangeset(ctx, changeset); err != nil {
			t.Fatal(err)
		}

		bt.SetChangesetFailed(t, ctx, s, changeset)

		if _, _, err := svc.ReenqueueChangeset(userCtx, changeset.ID); err != nil {
			t.Fatal(err)
		}

		bt.ReloadAndAssertChangeset(t, ctx, s, changeset, bt.ChangesetAssertions{
			Repo:          rs[0].ID,
			ExternalState: btypes.ChangesetExternalStateOpen,
			ExternalID:    "ext-id-5",
			AttachedTo:    []int64{batchChange.ID},

			// The important fields:
			ReconcilerState: btypes.ReconcilerStateQueued,
			NumResets:       0,
			NumFailures:     0,
			FailureMessage:  nil,
		})

		// rs[0] is filtered out
		bt.MockRepoPermissions(t, db, user.ID, rs[1].ID, rs[2].ID, rs[3].ID)

		// should result in a not found error
		if _, _, err := svc.ReenqueueChangeset(userCtx, changeset.ID); !errcode.IsNotFound(err) {
			t.Fatalf("expected not-found error but got %v", err)
		}
	})

	t.Run("CreateBatchSpec", func(t *testing.T) {
		changesetSpecs := make([]*btypes.ChangesetSpec, 0, len(rs))
		changesetSpecRandIDs := make([]string, 0, len(rs))
		for _, r := range rs {
			cs := &btypes.ChangesetSpec{BaseRepoID: r.ID, UserID: admin.ID, ExternalID: "123"}
			if err := s.CreateChangesetSpec(ctx, cs); err != nil {
				t.Fatal(err)
			}
			changesetSpecs = append(changesetSpecs, cs)
			changesetSpecRandIDs = append(changesetSpecRandIDs, cs.RandID)
		}

		t.Run("success", func(t *testing.T) {
			opts := CreateBatchSpecOpts{
				NamespaceUserID:      admin.ID,
				RawSpec:              bt.TestRawBatchSpec,
				ChangesetSpecRandIDs: changesetSpecRandIDs,
			}

			spec, err := svc.CreateBatchSpec(adminCtx, opts)
			if err != nil {
				t.Fatal(err)
			}

			if spec.ID == 0 {
				t.Fatalf("BatchSpec ID is 0")
			}

			if have, want := spec.UserID, admin.ID; have != want {
				t.Fatalf("UserID is %d, want %d", have, want)
			}

			var wantFields *batcheslib.BatchSpec
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

				if have, want := cs2.BatchSpecID, spec.ID; have != want {
					t.Fatalf("changesetSpec has wrong BatchSpec. want=%d, have=%d", want, have)
				}
			}
		})

		t.Run("success with YAML raw spec", func(t *testing.T) {
			opts := CreateBatchSpecOpts{
				NamespaceUserID: admin.ID,
				RawSpec:         bt.TestRawBatchSpecYAML,
			}

			spec, err := svc.CreateBatchSpec(adminCtx, opts)
			if err != nil {
				t.Fatal(err)
			}

			if spec.ID == 0 {
				t.Fatalf("BatchSpec ID is 0")
			}

			var wantFields *batcheslib.BatchSpec
			if err := json.Unmarshal([]byte(bt.TestRawBatchSpec), &wantFields); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(wantFields, spec.Spec); diff != "" {
				t.Fatalf("wrong spec fields (-want +got):\n%s", diff)
			}
		})

		t.Run("missing repository permissions", func(t *testing.T) {
			bt.MockRepoPermissions(t, db, user.ID)

			opts := CreateBatchSpecOpts{
				NamespaceUserID:      user.ID,
				RawSpec:              bt.TestRawBatchSpec,
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
				RawSpec:              bt.TestRawBatchSpec,
				ChangesetSpecRandIDs: containsInvalidID,
			}

			if _, err := svc.CreateBatchSpec(adminCtx, opts); !errcode.IsNotFound(err) {
				t.Fatalf("expected not-found error but got %s", err)
			}
		})

		t.Run("namespace user is not admin and not creator", func(t *testing.T) {
			opts := CreateBatchSpecOpts{
				NamespaceUserID: admin.ID,
				RawSpec:         bt.TestRawBatchSpecYAML,
			}

			_, err := svc.CreateBatchSpec(userCtx, opts)
			if !errcode.IsUnauthorized(err) {
				t.Fatalf("expected unauthorized error but got %s", err)
			}

			// Try again as admin
			opts.NamespaceUserID = user.ID

			_, err = svc.CreateBatchSpec(adminCtx, opts)
			if err != nil {
				t.Fatalf("expected no error but got %s", err)
			}
		})

		t.Run("missing access to namespace org", func(t *testing.T) {
			orgID := bt.InsertTestOrg(t, db, "test-org")

			opts := CreateBatchSpecOpts{
				NamespaceOrgID:       orgID,
				RawSpec:              bt.TestRawBatchSpec,
				ChangesetSpecRandIDs: changesetSpecRandIDs,
			}

			_, err := svc.CreateBatchSpec(userCtx, opts)
			if have, want := err, backend.ErrNotAnOrgMember; have != want {
				t.Fatalf("expected %s error but got %s", want, have)
			}

			// Create org membership and try again
			if _, err := db.OrgMembers().Create(ctx, orgID, user.ID); err != nil {
				t.Fatal(err)
			}

			_, err = svc.CreateBatchSpec(userCtx, opts)
			if err != nil {
				t.Fatalf("expected no error but got %s", err)
			}
		})

		t.Run("no side-effects if no changeset spec IDs are given", func(t *testing.T) {
			// We already have ChangesetSpecs in the database. Here we
			// want to make sure that the new BatchSpec is created,
			// without accidently attaching the existing ChangesetSpecs.
			opts := CreateBatchSpecOpts{
				NamespaceUserID:      admin.ID,
				RawSpec:              bt.TestRawBatchSpec,
				ChangesetSpecRandIDs: []string{},
			}

			spec, err := svc.CreateBatchSpec(adminCtx, opts)
			if err != nil {
				t.Fatal(err)
			}

			countOpts := store.CountChangesetSpecsOpts{BatchSpecID: spec.ID}
			count, err := s.CountChangesetSpecs(adminCtx, countOpts)
			if err != nil {
				return
			}
			if count != 0 {
				t.Fatalf("want no changeset specs attached to batch spec, but have %d", count)
			}
		})
	})

	t.Run("CreateChangesetSpec", func(t *testing.T) {
		repo := rs[0]
		rawSpec := bt.NewRawChangesetSpecGitBranch(graphqlbackend.MarshalRepositoryID(repo.ID), "d34db33f")

		t.Run("success", func(t *testing.T) {
			spec, err := svc.CreateChangesetSpec(ctx, rawSpec, admin.ID)
			if err != nil {
				t.Fatal(err)
			}

			if spec.ID == 0 {
				t.Fatalf("ChangesetSpec ID is 0")
			}

			want := &btypes.ChangesetSpec{
				ID:   5,
				Type: btypes.ChangesetSpecTypeBranch,
				Diff: []byte(`diff --git INSTALL.md INSTALL.md
index e5af166..d44c3fc 100644
--- INSTALL.md
+++ INSTALL.md
@@ -3,10 +3,10 @@
 Line 1
 Line 2
 Line 3
-Line 4
+This is cool: Line 4
 Line 5
 Line 6
-Line 7
-Line 8
+Another Line 7
+Foobar Line 8
 Line 9
 Line 10
`),
				DiffStatAdded:     1,
				DiffStatChanged:   2,
				DiffStatDeleted:   1,
				BaseRepoID:        1,
				UserID:            1,
				BaseRev:           "d34db33f",
				BaseRef:           "refs/heads/master",
				HeadRef:           "refs/heads/my-branch",
				Title:             "the title",
				Body:              "the body of the PR",
				Published:         batcheslib.PublishedValue{Val: false},
				CommitMessage:     "git commit message\n\nand some more content in a second paragraph.",
				CommitAuthorName:  "Mary McButtons",
				CommitAuthorEmail: "mary@example.com",
			}

			if diff := cmp.Diff(want, spec, cmpopts.IgnoreFields(btypes.ChangesetSpec{}, "CreatedAt", "UpdatedAt", "RandID")); diff != "" {
				t.Fatalf("wrong spec fields (-want +got):\n%s", diff)
			}

			wantDiffStat := *bt.ChangesetSpecDiffStat
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
			wantErr := "4 errors occurred:\n\t* Must validate one and only one schema (oneOf)\n\t* baseRepository is required\n\t* externalID is required\n\t* Additional property externalComputer is not allowed"
			if diff := cmp.Diff(wantErr, haveErr); diff != "" {
				t.Fatalf("unexpected error (-want +got):\n%s", diff)
			}
		})

		t.Run("missing repository permissions", func(t *testing.T) {
			bt.MockRepoPermissions(t, db, user.ID, rs[1].ID, rs[2].ID, rs[3].ID)

			_, err := svc.CreateChangesetSpec(userCtx, rawSpec, admin.ID)
			if !errcode.IsNotFound(err) {
				t.Fatalf("expected not-found error but got %v", err)
			}
		})
	})

	t.Run("ApplyBatchChange", func(t *testing.T) {
		// See TestServiceApplyBatchChange
	})

	t.Run("MoveBatchChange", func(t *testing.T) {
		createBatchChange := func(t *testing.T, name string, authorID, userID, orgID int32) *btypes.BatchChange {
			t.Helper()

			spec := &btypes.BatchSpec{
				UserID:          authorID,
				NamespaceUserID: userID,
				NamespaceOrgID:  orgID,
			}

			if err := s.CreateBatchSpec(ctx, spec); err != nil {
				t.Fatal(err)
			}

			c := &btypes.BatchChange{
				CreatorID:       authorID,
				NamespaceUserID: userID,
				NamespaceOrgID:  orgID,
				Name:            name,
				LastApplierID:   authorID,
				LastAppliedAt:   time.Now(),
				BatchSpecID:     spec.ID,
			}

			if err := s.CreateBatchChange(ctx, c); err != nil {
				t.Fatal(err)
			}

			return c
		}

		t.Run("new name", func(t *testing.T) {
			batchChange := createBatchChange(t, "old-name", admin.ID, admin.ID, 0)

			opts := MoveBatchChangeOpts{BatchChangeID: batchChange.ID, NewName: "new-name"}
			moved, err := svc.MoveBatchChange(ctx, opts)
			if err != nil {
				t.Fatal(err)
			}

			if have, want := moved.Name, opts.NewName; have != want {
				t.Fatalf("wrong name. want=%q, have=%q", want, have)
			}
		})

		t.Run("new user namespace", func(t *testing.T) {
			batchChange := createBatchChange(t, "old-name", admin.ID, admin.ID, 0)

			user2 := bt.CreateTestUser(t, db, false)

			opts := MoveBatchChangeOpts{BatchChangeID: batchChange.ID, NewNamespaceUserID: user2.ID}
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
			batchChange := createBatchChange(t, "old-name", user.ID, user.ID, 0)

			user2 := bt.CreateTestUser(t, db, false)

			opts := MoveBatchChangeOpts{BatchChangeID: batchChange.ID, NewNamespaceUserID: user2.ID}

			_, err := svc.MoveBatchChange(userCtx, opts)
			if !errcode.IsUnauthorized(err) {
				t.Fatalf("expected unauthorized error but got %s", err)
			}
		})

		t.Run("new org namespace", func(t *testing.T) {
			batchChange := createBatchChange(t, "old-name-1", admin.ID, admin.ID, 0)

			orgID := bt.InsertTestOrg(t, db, "org")

			opts := MoveBatchChangeOpts{BatchChangeID: batchChange.ID, NewNamespaceOrgID: orgID}
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
			batchChange := createBatchChange(t, "old-name-2", user.ID, user.ID, 0)

			orgID := bt.InsertTestOrg(t, db, "org-no-access")

			opts := MoveBatchChangeOpts{BatchChangeID: batchChange.ID, NewNamespaceOrgID: orgID}

			_, err := svc.MoveBatchChange(userCtx, opts)
			if have, want := err, backend.ErrNotAnOrgMember; !errors.Is(have, want) {
				t.Fatalf("expected %s error but got %s", want, have)
			}
		})
	})

	t.Run("GetBatchChangeMatchingBatchSpec", func(t *testing.T) {
		batchSpec := bt.CreateBatchSpec(t, ctx, s, "matching-batch-spec", admin.ID, 0)

		haveBatchChange, err := svc.GetBatchChangeMatchingBatchSpec(ctx, batchSpec)
		if err != nil {
			t.Fatalf("unexpected error: %s\n", err)
		}
		if haveBatchChange != nil {
			t.Fatalf("expected batch change to be nil, but is not: %+v\n", haveBatchChange)
		}

		matchingBatchChange := &btypes.BatchChange{
			Name:            batchSpec.Spec.Name,
			Description:     batchSpec.Spec.Description,
			CreatorID:       admin.ID,
			NamespaceOrgID:  batchSpec.NamespaceOrgID,
			NamespaceUserID: batchSpec.NamespaceUserID,
			BatchSpecID:     batchSpec.ID,
			LastApplierID:   admin.ID,
			LastAppliedAt:   time.Now(),
		}
		if err := s.CreateBatchChange(ctx, matchingBatchChange); err != nil {
			t.Fatalf("failed to create batch change: %s\n", err)
		}

		t.Run("BatchChangeID is not provided", func(t *testing.T) {
			haveBatchChange, err = svc.GetBatchChangeMatchingBatchSpec(ctx, batchSpec)
			if err != nil {
				t.Fatalf("unexpected error: %s\n", err)
			}
			if haveBatchChange == nil {
				t.Fatalf("expected to have matching batch change, but got nil")
			}

			if diff := cmp.Diff(matchingBatchChange, haveBatchChange); diff != "" {
				t.Fatalf("wrong batch change was matched (-want +got):\n%s", diff)
			}
		})

		t.Run("BatchChangeID is provided", func(t *testing.T) {
			batchSpec2 := bt.CreateBatchSpec(t, ctx, s, "matching-batch-spec", admin.ID, matchingBatchChange.ID)
			haveBatchChange, err = svc.GetBatchChangeMatchingBatchSpec(ctx, batchSpec2)
			if err != nil {
				t.Fatalf("unexpected error: %s\n", err)
			}
			if haveBatchChange == nil {
				t.Fatalf("expected to have matching batch change, but got nil")
			}

			if diff := cmp.Diff(matchingBatchChange, haveBatchChange); diff != "" {
				t.Fatalf("wrong batch change was matched (-want +got):\n%s", diff)
			}
		})
	})

	t.Run("GetNewestBatchSpec", func(t *testing.T) {
		older := bt.CreateBatchSpec(t, ctx, s, "superseding", user.ID, 0)
		newer := bt.CreateBatchSpec(t, ctx, s, "superseding", user.ID, 0)

		for name, in := range map[string]*btypes.BatchSpec{
			"older": older,
			"newer": newer,
		} {
			t.Run(name, func(t *testing.T) {
				have, err := svc.GetNewestBatchSpec(ctx, s, in, user.ID)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if diff := cmp.Diff(newer, have); diff != "" {
					t.Errorf("unexpected newer batch spec (-want +have):\n%s", diff)
				}
			})
		}

		t.Run("different user", func(t *testing.T) {
			have, err := svc.GetNewestBatchSpec(ctx, s, older, admin.ID)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if have != nil {
				t.Errorf("unexpected non-nil batch spec: %+v", have)
			}
		})
	})

	t.Run("FetchUsernameForBitbucketServerToken", func(t *testing.T) {
		fakeSource := &stesting.FakeChangesetSource{Username: "my-bbs-username"}
		sourcer := stesting.NewFakeSourcer(nil, fakeSource)

		// Create a fresh service for this test as to not mess with state
		// possibly used by other tests.
		testSvc := New(s)
		testSvc.sourcer = sourcer

		rs, _ := bt.CreateBbsTestRepos(t, ctx, db, 1)
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

	t.Run("ValidateAuthenticator", func(t *testing.T) {
		t.Run("valid", func(t *testing.T) {
			fakeSource.AuthenticatorIsValid = true
			fakeSource.ValidateAuthenticatorCalled = false
			if err := svc.ValidateAuthenticator(
				ctx,
				"https://github.com/",
				extsvc.TypeGitHub,
				&auth.OAuthBearerToken{Token: "test123"},
			); err != nil {
				t.Fatal(err)
			}
			if !fakeSource.ValidateAuthenticatorCalled {
				t.Fatal("ValidateAuthenticator on Source not called")
			}
		})
		t.Run("invalid", func(t *testing.T) {
			fakeSource.AuthenticatorIsValid = false
			fakeSource.ValidateAuthenticatorCalled = false
			if err := svc.ValidateAuthenticator(
				ctx,
				"https://github.com/",
				extsvc.TypeGitHub,
				&auth.OAuthBearerToken{Token: "test123"},
			); err == nil {
				t.Fatal("unexpected nil-error returned from ValidateAuthenticator")
			}
			if !fakeSource.ValidateAuthenticatorCalled {
				t.Fatal("ValidateAuthenticator on Source not called")
			}
		})
	})

	t.Run("CreateChangesetJobs", func(t *testing.T) {
		spec := testBatchSpec(admin.ID)
		if err := s.CreateBatchSpec(ctx, spec); err != nil {
			t.Fatal(err)
		}

		batchChange := testBatchChange(admin.ID, spec)
		if err := s.CreateBatchChange(ctx, batchChange); err != nil {
			t.Fatal(err)
		}

		t.Run("creates jobs", func(t *testing.T) {
			changeset1 := bt.CreateChangeset(t, ctx, s, bt.TestChangesetOpts{
				Repo:             rs[0].ID,
				PublicationState: btypes.ChangesetPublicationStatePublished,
				BatchChange:      batchChange.ID,
			})
			changeset2 := bt.CreateChangeset(t, ctx, s, bt.TestChangesetOpts{
				Repo:             rs[1].ID,
				PublicationState: btypes.ChangesetPublicationStatePublished,
				BatchChange:      batchChange.ID,
			})
			bulkOperationID, err := svc.CreateChangesetJobs(
				adminCtx,
				batchChange.ID,
				[]int64{changeset1.ID, changeset2.ID},
				btypes.ChangesetJobTypeComment,
				btypes.ChangesetJobCommentPayload{Message: "test"},
				store.ListChangesetsOpts{},
			)
			if err != nil {
				t.Fatal(err)
			}
			// Validate the bulk operation exists.
			if _, err = s.GetBulkOperation(ctx, store.GetBulkOperationOpts{ID: bulkOperationID}); err != nil {
				t.Fatal(err)
			}
		})

		t.Run("changeset not found", func(t *testing.T) {
			changeset := bt.CreateChangeset(t, ctx, s, bt.TestChangesetOpts{
				Repo:        rs[0].ID,
				BatchChange: batchChange.ID,
			})
			_, err := svc.CreateChangesetJobs(
				adminCtx,
				batchChange.ID,
				[]int64{changeset.ID},
				btypes.ChangesetJobTypeComment,
				btypes.ChangesetJobCommentPayload{Message: "test"},
				store.ListChangesetsOpts{
					ReconcilerStates: []btypes.ReconcilerState{btypes.ReconcilerStateCompleted},
				},
			)
			if err != ErrChangesetsForJobNotFound {
				t.Fatalf("wrong error. want=%s, got=%s", ErrChangesetsForJobNotFound, err)
			}
		})

		t.Run("DetachChangesets", func(t *testing.T) {
			spec := testBatchSpec(admin.ID)
			if err := s.CreateBatchSpec(ctx, spec); err != nil {
				t.Fatal(err)
			}

			batchChange := testBatchChange(admin.ID, spec)
			if err := s.CreateBatchChange(ctx, batchChange); err != nil {
				t.Fatal(err)
			}
			t.Run("attached changeset", func(t *testing.T) {
				changeset := bt.CreateChangeset(t, ctx, s, bt.TestChangesetOpts{
					Repo:             rs[0].ID,
					ReconcilerState:  btypes.ReconcilerStateCompleted,
					PublicationState: btypes.ChangesetPublicationStatePublished,
					ExternalState:    btypes.ChangesetExternalStateOpen,
					BatchChange:      batchChange.ID,
					IsArchived:       false,
				})
				_, err := svc.CreateChangesetJobs(ctx, batchChange.ID, []int64{changeset.ID}, btypes.ChangesetJobTypeDetach, btypes.ChangesetJobDetachPayload{}, store.ListChangesetsOpts{OnlyArchived: true})
				if err != ErrChangesetsForJobNotFound {
					t.Fatalf("wrong error. want=%s, got=%s", ErrChangesetsForJobNotFound, err)
				}
			})
			t.Run("detached changeset", func(t *testing.T) {
				detachedChangeset := bt.CreateChangeset(t, ctx, s, bt.TestChangesetOpts{
					Repo:             rs[2].ID,
					ReconcilerState:  btypes.ReconcilerStateCompleted,
					PublicationState: btypes.ChangesetPublicationStatePublished,
					ExternalState:    btypes.ChangesetExternalStateOpen,
					BatchChanges:     []btypes.BatchChangeAssoc{},
				})
				_, err := svc.CreateChangesetJobs(ctx, batchChange.ID, []int64{detachedChangeset.ID}, btypes.ChangesetJobTypeDetach, btypes.ChangesetJobDetachPayload{}, store.ListChangesetsOpts{OnlyArchived: true})
				if err != ErrChangesetsForJobNotFound {
					t.Fatalf("wrong error. want=%s, got=%s", ErrChangesetsForJobNotFound, err)
				}
			})
		})

		t.Run("MergeChangesets", func(t *testing.T) {
			spec := testBatchSpec(admin.ID)
			if err := s.CreateBatchSpec(ctx, spec); err != nil {
				t.Fatal(err)
			}

			batchChange := testBatchChange(admin.ID, spec)
			if err := s.CreateBatchChange(ctx, batchChange); err != nil {
				t.Fatal(err)
			}
			published := btypes.ChangesetPublicationStatePublished
			openState := btypes.ChangesetExternalStateOpen
			t.Run("open changeset", func(t *testing.T) {
				changeset := bt.CreateChangeset(t, adminCtx, s, bt.TestChangesetOpts{
					Repo:             rs[0].ID,
					ReconcilerState:  btypes.ReconcilerStateCompleted,
					ExternalState:    btypes.ChangesetExternalStateOpen,
					PublicationState: btypes.ChangesetPublicationStatePublished,
					BatchChange:      batchChange.ID,
					IsArchived:       false,
				})
				_, err := svc.CreateChangesetJobs(
					adminCtx,
					batchChange.ID,
					[]int64{changeset.ID},
					btypes.ChangesetJobTypeMerge,
					btypes.ChangesetJobMergePayload{Squash: true},
					store.ListChangesetsOpts{
						PublicationState: &published,
						ReconcilerStates: []btypes.ReconcilerState{btypes.ReconcilerStateCompleted},
						ExternalStates:   []btypes.ChangesetExternalState{openState},
					},
				)
				if err != nil {
					t.Fatal(err)
				}
			})
			t.Run("closed changeset", func(t *testing.T) {
				closedChangeset := bt.CreateChangeset(t, adminCtx, s, bt.TestChangesetOpts{
					Repo:             rs[0].ID,
					ReconcilerState:  btypes.ReconcilerStateCompleted,
					ExternalState:    btypes.ChangesetExternalStateClosed,
					PublicationState: btypes.ChangesetPublicationStatePublished,
					BatchChange:      batchChange.ID,
				})
				_, err := svc.CreateChangesetJobs(
					adminCtx,
					batchChange.ID,
					[]int64{closedChangeset.ID},
					btypes.ChangesetJobTypeMerge,
					btypes.ChangesetJobMergePayload{},
					store.ListChangesetsOpts{
						PublicationState: &published,
						ReconcilerStates: []btypes.ReconcilerState{btypes.ReconcilerStateCompleted},
						ExternalStates:   []btypes.ChangesetExternalState{openState},
					},
				)
				if err != ErrChangesetsForJobNotFound {
					t.Fatalf("wrong error. want=%s, got=%s", ErrChangesetsForJobNotFound, err)
				}
			})
		})

		t.Run("CloseChangesets", func(t *testing.T) {
			spec := testBatchSpec(admin.ID)
			if err := s.CreateBatchSpec(ctx, spec); err != nil {
				t.Fatal(err)
			}

			batchChange := testBatchChange(admin.ID, spec)
			if err := s.CreateBatchChange(ctx, batchChange); err != nil {
				t.Fatal(err)
			}
			published := btypes.ChangesetPublicationStatePublished
			openState := btypes.ChangesetExternalStateOpen
			t.Run("open changeset", func(t *testing.T) {
				changeset := bt.CreateChangeset(t, adminCtx, s, bt.TestChangesetOpts{
					Repo:             rs[0].ID,
					ReconcilerState:  btypes.ReconcilerStateCompleted,
					ExternalState:    btypes.ChangesetExternalStateOpen,
					PublicationState: btypes.ChangesetPublicationStatePublished,
					BatchChange:      batchChange.ID,
					IsArchived:       false,
				})
				_, err := svc.CreateChangesetJobs(
					adminCtx,
					batchChange.ID,
					[]int64{changeset.ID},
					btypes.ChangesetJobTypeClose,
					btypes.ChangesetJobClosePayload{},
					store.ListChangesetsOpts{
						PublicationState: &published,
						ReconcilerStates: []btypes.ReconcilerState{btypes.ReconcilerStateCompleted},
						ExternalStates:   []btypes.ChangesetExternalState{openState},
					},
				)
				if err != nil {
					t.Fatal(err)
				}
			})
			t.Run("closed changeset", func(t *testing.T) {
				closedChangeset := bt.CreateChangeset(t, adminCtx, s, bt.TestChangesetOpts{
					Repo:             rs[0].ID,
					ReconcilerState:  btypes.ReconcilerStateCompleted,
					ExternalState:    btypes.ChangesetExternalStateClosed,
					PublicationState: btypes.ChangesetPublicationStatePublished,
					BatchChange:      batchChange.ID,
				})
				_, err := svc.CreateChangesetJobs(
					adminCtx,
					batchChange.ID,
					[]int64{closedChangeset.ID},
					btypes.ChangesetJobTypeClose,
					btypes.ChangesetJobClosePayload{},
					store.ListChangesetsOpts{
						PublicationState: &published,
						ReconcilerStates: []btypes.ReconcilerState{btypes.ReconcilerStateCompleted},
						ExternalStates:   []btypes.ChangesetExternalState{openState},
					},
				)
				if err != ErrChangesetsForJobNotFound {
					t.Fatalf("wrong error. want=%s, got=%s", ErrChangesetsForJobNotFound, err)
				}
			})
		})

		t.Run("PublishChangesets", func(t *testing.T) {
			spec := testBatchSpec(admin.ID)
			if err := s.CreateBatchSpec(ctx, spec); err != nil {
				t.Fatal(err)
			}

			batchChange := testBatchChange(admin.ID, spec)
			if err := s.CreateBatchChange(ctx, batchChange); err != nil {
				t.Fatal(err)
			}

			changeset := bt.CreateChangeset(t, adminCtx, s, bt.TestChangesetOpts{
				Repo:             rs[0].ID,
				ReconcilerState:  btypes.ReconcilerStateCompleted,
				PublicationState: btypes.ChangesetPublicationStateUnpublished,
				BatchChange:      batchChange.ID,
			})

			_, err := svc.CreateChangesetJobs(
				adminCtx,
				batchChange.ID,
				[]int64{changeset.ID},
				btypes.ChangesetJobTypePublish,
				btypes.ChangesetJobPublishPayload{Draft: true},
				store.ListChangesetsOpts{},
			)
			if err != nil {
				t.Fatal(err)
			}

		})
	})

	t.Run("ExecuteBatchSpec", func(t *testing.T) {
		adminCtx := actor.WithActor(ctx, actor.FromUser(admin.ID))
		t.Run("success", func(t *testing.T) {
			spec := testBatchSpec(admin.ID)
			if err := s.CreateBatchSpec(ctx, spec); err != nil {
				t.Fatal(err)
			}

			// Simulate successful resolution.
			job := &btypes.BatchSpecResolutionJob{
				State:       btypes.BatchSpecResolutionJobStateCompleted,
				BatchSpecID: spec.ID,
				InitiatorID: admin.ID,
			}

			if err := s.CreateBatchSpecResolutionJob(ctx, job); err != nil {
				t.Fatal(err)
			}

			var workspaceIDs []int64
			for _, repo := range rs {
				ws := &btypes.BatchSpecWorkspace{
					BatchSpecID: spec.ID,
					RepoID:      repo.ID,
				}
				if err := s.CreateBatchSpecWorkspace(ctx, ws); err != nil {
					t.Fatal(err)
				}
				workspaceIDs = append(workspaceIDs, ws.ID)
			}

			// Execute BatchSpec by creating execution jobs
			if _, err := svc.ExecuteBatchSpec(adminCtx, ExecuteBatchSpecOpts{BatchSpecRandID: spec.RandID}); err != nil {
				t.Fatal(err)
			}

			jobs, err := s.ListBatchSpecWorkspaceExecutionJobs(ctx, store.ListBatchSpecWorkspaceExecutionJobsOpts{
				BatchSpecWorkspaceIDs: workspaceIDs,
			})
			if err != nil {
				t.Fatal(err)
			}

			if len(jobs) != len(rs) {
				t.Fatalf("wrong number of execution jobs created. want=%d, have=%d", len(rs), len(jobs))
			}
		})
		t.Run("resolution not completed", func(t *testing.T) {
			spec := testBatchSpec(admin.ID)
			if err := s.CreateBatchSpec(ctx, spec); err != nil {
				t.Fatal(err)
			}

			job := &btypes.BatchSpecResolutionJob{
				State:       btypes.BatchSpecResolutionJobStateQueued,
				BatchSpecID: spec.ID,
				InitiatorID: admin.ID,
			}

			if err := s.CreateBatchSpecResolutionJob(ctx, job); err != nil {
				t.Fatal(err)
			}

			// Execute BatchSpec by creating execution jobs
			_, err := svc.ExecuteBatchSpec(adminCtx, ExecuteBatchSpecOpts{BatchSpecRandID: spec.RandID})
			if !errors.Is(err, ErrBatchSpecResolutionIncomplete) {
				t.Fatalf("error has wrong type: %T", err)
			}
		})

		t.Run("resolution failed", func(t *testing.T) {
			spec := testBatchSpec(admin.ID)
			if err := s.CreateBatchSpec(ctx, spec); err != nil {
				t.Fatal(err)
			}

			failureMessage := "cat ate the homework"
			job := &btypes.BatchSpecResolutionJob{
				State:          btypes.BatchSpecResolutionJobStateFailed,
				FailureMessage: &failureMessage,
				BatchSpecID:    spec.ID,
				InitiatorID:    admin.ID,
			}

			if err := s.CreateBatchSpecResolutionJob(ctx, job); err != nil {
				t.Fatal(err)
			}

			// Execute BatchSpec by creating execution jobs
			_, err := svc.ExecuteBatchSpec(adminCtx, ExecuteBatchSpecOpts{BatchSpecRandID: spec.RandID})
			if !errors.HasType(err, ErrBatchSpecResolutionErrored{}) {
				t.Fatalf("error has wrong type: %T", err)
			}
		})

		t.Run("ignored/unsupported workspace", func(t *testing.T) {
			spec := testBatchSpec(admin.ID)
			if err := s.CreateBatchSpec(ctx, spec); err != nil {
				t.Fatal(err)
			}

			// Simulate successful resolution.
			job := &btypes.BatchSpecResolutionJob{
				State:       btypes.BatchSpecResolutionJobStateCompleted,
				BatchSpecID: spec.ID,
				InitiatorID: admin.ID,
			}

			if err := s.CreateBatchSpecResolutionJob(ctx, job); err != nil {
				t.Fatal(err)
			}

			ignoredWorkspace := &btypes.BatchSpecWorkspace{
				BatchSpecID: spec.ID,
				RepoID:      rs[0].ID,
				Ignored:     true,
			}

			unsupportedWorkspace := &btypes.BatchSpecWorkspace{
				BatchSpecID: spec.ID,
				RepoID:      rs[0].ID,
				Unsupported: true,
			}
			if err := s.CreateBatchSpecWorkspace(ctx, ignoredWorkspace, unsupportedWorkspace); err != nil {
				t.Fatal(err)
			}

			if _, err := svc.ExecuteBatchSpec(adminCtx, ExecuteBatchSpecOpts{BatchSpecRandID: spec.RandID}); err != nil {
				t.Fatal(err)
			}

			ids := []int64{ignoredWorkspace.ID, unsupportedWorkspace.ID}
			jobs, err := s.ListBatchSpecWorkspaceExecutionJobs(ctx, store.ListBatchSpecWorkspaceExecutionJobsOpts{
				BatchSpecWorkspaceIDs: ids,
			})
			if err != nil {
				t.Fatal(err)
			}

			if len(jobs) != 0 {
				t.Fatalf("wrong number of execution jobs created. want=%d, have=%d", len(rs), len(jobs))
			}

			for _, workspaceID := range ids {
				reloaded, err := s.GetBatchSpecWorkspace(ctx, store.GetBatchSpecWorkspaceOpts{ID: workspaceID})
				if err != nil {
					t.Fatal(err)
				}
				if !reloaded.Skipped {
					t.Fatalf("workspace not marked as skipped")
				}
			}
		})
	})

	t.Run("CancelBatchSpec", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			spec := testBatchSpec(admin.ID)
			spec.CreatedFromRaw = true
			if err := s.CreateBatchSpec(ctx, spec); err != nil {
				t.Fatal(err)
			}

			// Simulate successful resolution.
			job := &btypes.BatchSpecResolutionJob{
				State:       btypes.BatchSpecResolutionJobStateCompleted,
				BatchSpecID: spec.ID,
				InitiatorID: admin.ID,
			}

			if err := s.CreateBatchSpecResolutionJob(ctx, job); err != nil {
				t.Fatal(err)
			}

			var jobIDs []int64
			for _, repo := range rs {
				ws := &btypes.BatchSpecWorkspace{BatchSpecID: spec.ID, RepoID: repo.ID}
				if err := s.CreateBatchSpecWorkspace(ctx, ws); err != nil {
					t.Fatal(err)
				}

				job := &btypes.BatchSpecWorkspaceExecutionJob{
					BatchSpecWorkspaceID: ws.ID,
					UserID:               user.ID,
				}
				if err := bt.CreateBatchSpecWorkspaceExecutionJob(ctx, s, store.ScanBatchSpecWorkspaceExecutionJob, job); err != nil {
					t.Fatal(err)
				}

				job.State = btypes.BatchSpecWorkspaceExecutionJobStateProcessing
				job.StartedAt = time.Now()
				bt.UpdateJobState(t, ctx, s, job)

				jobIDs = append(jobIDs, job.ID)
			}

			if _, err := svc.CancelBatchSpec(ctx, CancelBatchSpecOpts{BatchSpecRandID: spec.RandID}); err != nil {
				t.Fatal(err)
			}

			jobs, err := s.ListBatchSpecWorkspaceExecutionJobs(ctx, store.ListBatchSpecWorkspaceExecutionJobsOpts{
				IDs: jobIDs,
			})
			if err != nil {
				t.Fatal(err)
			}

			if len(jobs) != len(rs) {
				t.Fatalf("wrong number of execution jobs created. want=%d, have=%d", len(rs), len(jobs))
			}

			var canceled int
			for _, j := range jobs {
				if j.Cancel {
					canceled += 1
				}
			}
			if canceled != len(jobs) {
				t.Fatalf("not all jobs were canceled. jobs=%d, canceled=%d", len(jobs), canceled)
			}
		})

		t.Run("already completed", func(t *testing.T) {
			spec := testBatchSpec(admin.ID)
			spec.CreatedFromRaw = true
			if err := s.CreateBatchSpec(ctx, spec); err != nil {
				t.Fatal(err)
			}

			resolutionJob := &btypes.BatchSpecResolutionJob{
				BatchSpecID: spec.ID,
				InitiatorID: admin.ID,
			}
			if err := s.CreateBatchSpecResolutionJob(ctx, resolutionJob); err != nil {
				t.Fatal(err)
			}

			ws := &btypes.BatchSpecWorkspace{BatchSpecID: spec.ID, RepoID: rs[0].ID}
			if err := s.CreateBatchSpecWorkspace(ctx, ws); err != nil {
				t.Fatal(err)
			}

			job := &btypes.BatchSpecWorkspaceExecutionJob{
				BatchSpecWorkspaceID: ws.ID,
				UserID:               user.ID,
			}
			if err := bt.CreateBatchSpecWorkspaceExecutionJob(ctx, s, store.ScanBatchSpecWorkspaceExecutionJob, job); err != nil {
				t.Fatal(err)
			}

			job.State = btypes.BatchSpecWorkspaceExecutionJobStateCompleted
			job.StartedAt = time.Now()
			job.FinishedAt = time.Now()
			bt.UpdateJobState(t, ctx, s, job)

			_, err := svc.CancelBatchSpec(ctx, CancelBatchSpecOpts{BatchSpecRandID: spec.RandID})
			if !errors.Is(err, ErrBatchSpecNotCancelable) {
				t.Fatalf("error has wrong type: %T", err)
			}
		})
	})

	t.Run("ReplaceBatchSpecInput", func(t *testing.T) {
		createBatchSpecWithWorkspaces := func(t *testing.T) *btypes.BatchSpec {
			t.Helper()
			spec := testBatchSpec(admin.ID)
			if err := s.CreateBatchSpec(ctx, spec); err != nil {
				t.Fatal(err)
			}

			job := &btypes.BatchSpecResolutionJob{
				State:       btypes.BatchSpecResolutionJobStateCompleted,
				BatchSpecID: spec.ID,
				InitiatorID: admin.ID,
			}

			if err := s.CreateBatchSpecResolutionJob(ctx, job); err != nil {
				t.Fatal(err)
			}

			for _, repo := range rs {
				ws := &btypes.BatchSpecWorkspace{BatchSpecID: spec.ID, RepoID: repo.ID}
				if err := s.CreateBatchSpecWorkspace(ctx, ws); err != nil {
					t.Fatal(err)
				}
			}
			return spec
		}

		createBatchSpecWithWorkspacesAndChangesetSpecs := func(t *testing.T) *btypes.BatchSpec {
			t.Helper()

			spec := createBatchSpecWithWorkspaces(t)

			for _, r := range rs {
				bt.CreateChangesetSpec(t, ctx, s, bt.TestSpecOpts{
					BatchSpec: spec.ID,
					Repo:      r.ID,
					Typ:       btypes.ChangesetSpecTypeBranch,
				})
			}

			return spec
		}

		t.Run("success", func(t *testing.T) {
			spec := createBatchSpecWithWorkspaces(t)

			newSpec, err := svc.ReplaceBatchSpecInput(ctx, ReplaceBatchSpecInputOpts{
				BatchSpecRandID: spec.RandID,
				RawSpec:         bt.TestRawBatchSpecYAML,
			})
			if err != nil {
				t.Fatal(err)
			}

			if newSpec.ID == spec.ID {
				t.Fatalf("new batch spec has same ID as old one: %d", newSpec.ID)
			}

			if newSpec.RandID != spec.RandID {
				t.Fatalf("new batch spec has different RandID. new=%s, old=%s", newSpec.RandID, spec.RandID)
			}
			if newSpec.UserID != spec.UserID {
				t.Fatalf("new batch spec has different UserID. new=%d, old=%d", newSpec.UserID, spec.UserID)
			}
			if newSpec.NamespaceUserID != spec.NamespaceUserID {
				t.Fatalf("new batch spec has different NamespaceUserID. new=%d, old=%d", newSpec.NamespaceUserID, spec.NamespaceUserID)
			}
			if newSpec.NamespaceOrgID != spec.NamespaceOrgID {
				t.Fatalf("new batch spec has different NamespaceOrgID. new=%d, old=%d", newSpec.NamespaceOrgID, spec.NamespaceOrgID)
			}

			if !newSpec.CreatedFromRaw {
				t.Fatalf("new batch spec not createdFromRaw: %t", newSpec.CreatedFromRaw)
			}

			resolutionJob, err := s.GetBatchSpecResolutionJob(ctx, store.GetBatchSpecResolutionJobOpts{
				BatchSpecID: newSpec.ID,
			})
			if err != nil {
				t.Fatal(err)
			}
			if want, have := btypes.BatchSpecResolutionJobStateQueued, resolutionJob.State; have != want {
				t.Fatalf("resolution job has wrong state. want=%s, have=%s", want, have)
			}

			// Assert that old batch spec is deleted
			_, err = s.GetBatchSpec(ctx, store.GetBatchSpecOpts{ID: spec.ID})
			if err != store.ErrNoResults {
				t.Fatalf("unexpected error: %s", err)
			}
		})

		t.Run("batchSpec already has changeset specs", func(t *testing.T) {
			assertNoChangesetSpecs := func(t *testing.T, batchSpecID int64) {
				t.Helper()
				specs, _, err := s.ListChangesetSpecs(ctx, store.ListChangesetSpecsOpts{
					BatchSpecID: batchSpecID,
				})
				if err != nil {
					t.Fatal(err)
				}
				if len(specs) != 0 {
					t.Fatalf("wrong number of changeset specs attached to batch spec %d: %d", batchSpecID, len(specs))
				}
			}

			spec := createBatchSpecWithWorkspacesAndChangesetSpecs(t)

			newSpec, err := svc.ReplaceBatchSpecInput(ctx, ReplaceBatchSpecInputOpts{
				BatchSpecRandID: spec.RandID,
				RawSpec:         bt.TestRawBatchSpecYAML,
			})
			if err != nil {
				t.Fatal(err)
			}

			assertNoChangesetSpecs(t, newSpec.ID)
			assertNoChangesetSpecs(t, spec.ID)
		})

		t.Run("mount error", func(t *testing.T) {
			spec := createBatchSpecWithWorkspacesAndChangesetSpecs(t)

			_, err := svc.ReplaceBatchSpecInput(adminCtx, ReplaceBatchSpecInputOpts{
				BatchSpecRandID: spec.RandID,
				RawSpec: `
name: test-spec
description: A test spec
steps:
  - run: /tmp/sample.sh
    container: alpine:3
    mount:
      - path: /some/path/sample.sh
        mountpoint: /tmp/sample.sh
changesetTemplate:
  title: Test Mount
  body: Test a mounted path
  branch: test
  commit:
    message: Test
`,
			})
			assert.Equal(t, "mounts are not allowed for server-side processing", err.Error())
		})
	})

	t.Run("CreateBatchSpecFromRaw", func(t *testing.T) {
		t.Run("batch change isn't owned by non-admin user", func(t *testing.T) {
			spec := testBatchSpec(user.ID)
			if err := s.CreateBatchSpec(ctx, spec); err != nil {
				t.Fatal(err)
			}

			batchChange := testBatchChange(user.ID, spec)
			if err := s.CreateBatchChange(ctx, batchChange); err != nil {
				t.Fatal(err)
			}

			_, err := svc.CreateBatchSpecFromRaw(user2Ctx, CreateBatchSpecFromRawOpts{
				RawSpec:         bt.TestRawBatchSpecYAML,
				NamespaceUserID: user2.ID,
				BatchChange:     batchChange.ID,
			})

			assert.Equal(t, "must be authenticated as the authorized user or as an admin (must be site admin)", err.Error())
		})

		t.Run("success - without batch change ID", func(t *testing.T) {
			newSpec, err := svc.CreateBatchSpecFromRaw(adminCtx, CreateBatchSpecFromRawOpts{
				RawSpec:         bt.TestRawBatchSpecYAML,
				NamespaceUserID: admin.ID,
			})
			if err != nil {
				t.Fatal(err)
			}

			if !newSpec.CreatedFromRaw {
				t.Fatalf("batchSpec not createdFromRaw: %t", newSpec.CreatedFromRaw)
			}

			resolutionJob, err := s.GetBatchSpecResolutionJob(ctx, store.GetBatchSpecResolutionJobOpts{
				BatchSpecID: newSpec.ID,
			})
			if err != nil {
				t.Fatal(err)
			}
			if want, have := btypes.BatchSpecResolutionJobStateQueued, resolutionJob.State; have != want {
				t.Fatalf("resolution job has wrong state. want=%s, have=%s", want, have)
			}
		})

		t.Run("success - with batch change ID", func(t *testing.T) {
			spec := testBatchSpec(admin.ID)
			if err := s.CreateBatchSpec(ctx, spec); err != nil {
				t.Fatal(err)
			}

			batchChange := testBatchChange(admin.ID, spec)
			if err := s.CreateBatchChange(ctx, batchChange); err != nil {
				t.Fatal(err)
			}

			newSpec, err := svc.CreateBatchSpecFromRaw(adminCtx, CreateBatchSpecFromRawOpts{
				RawSpec:         bt.TestRawBatchSpecYAML,
				NamespaceUserID: admin.ID,
				BatchChange:     batchChange.ID,
			})
			if err != nil {
				t.Fatal(err)
			}

			if !newSpec.CreatedFromRaw {
				t.Fatalf("batchSpec not createdFromRaw: %t", newSpec.CreatedFromRaw)
			}

			resolutionJob, err := s.GetBatchSpecResolutionJob(ctx, store.GetBatchSpecResolutionJobOpts{
				BatchSpecID: newSpec.ID,
			})
			if err != nil {
				t.Fatal(err)
			}
			if want, have := btypes.BatchSpecResolutionJobStateQueued, resolutionJob.State; have != want {
				t.Fatalf("resolution job has wrong state. want=%s, have=%s", want, have)
			}
		})

		t.Run("validation error", func(t *testing.T) {
			rawSpec := batcheslib.BatchSpec{
				Name:        "test-batch-change",
				Description: "only importing",
				ImportChangesets: []batcheslib.ImportChangeset{
					{Repository: string(rs[0].Name), ExternalIDs: []any{true, false}},
				},
			}

			marshaledRawSpec, err := json.Marshal(rawSpec)
			if err != nil {
				t.Fatal(err)
			}

			_, err = svc.CreateBatchSpecFromRaw(adminCtx, CreateBatchSpecFromRawOpts{
				RawSpec:         string(marshaledRawSpec),
				NamespaceUserID: admin.ID,
			})
			if err == nil {
				t.Fatalf("expected error but got none")
			}
			if !strings.Contains(err.Error(), "Invalid type. Expected: string, given: boolean") {
				t.Fatalf("wrong error message: %s", err)
			}
		})

		t.Run("mount error", func(t *testing.T) {
			_, err := svc.CreateBatchSpecFromRaw(adminCtx, CreateBatchSpecFromRawOpts{
				RawSpec: `
name: test-spec
description: A test spec
steps:
  - run: /tmp/sample.sh
    container: alpine:3
    mount:
      - path: /some/path/sample.sh
        mountpoint: /tmp/sample.sh
changesetTemplate:
  title: Test Mount
  body: Test a mounted path
  branch: test
  commit:
    message: Test
`,
				NamespaceUserID: admin.ID,
			})
			assert.Equal(t, "mounts are not allowed for server-side processing", err.Error())
		})
	})

	t.Run("UpsertBatchSpecInput", func(t *testing.T) {
		adminCtx := actor.WithActor(ctx, actor.FromUser(admin.ID))
		t.Run("new spec", func(t *testing.T) {
			newSpec, err := svc.UpsertBatchSpecInput(adminCtx, UpsertBatchSpecInputOpts{
				RawSpec:         bt.TestRawBatchSpecYAML,
				NamespaceUserID: admin.ID,
			})
			assert.Nil(t, err)
			assert.True(t, newSpec.CreatedFromRaw)

			resolutionJob, err := s.GetBatchSpecResolutionJob(ctx, store.GetBatchSpecResolutionJobOpts{
				BatchSpecID: newSpec.ID,
			})
			assert.Nil(t, err)
			assert.Equal(t, btypes.BatchSpecResolutionJobStateQueued, resolutionJob.State)
		})

		t.Run("replaced spec", func(t *testing.T) {
			oldSpec, err := svc.UpsertBatchSpecInput(adminCtx, UpsertBatchSpecInputOpts{
				RawSpec:         bt.TestRawBatchSpecYAML,
				NamespaceUserID: admin.ID,
			})
			assert.Nil(t, err)
			assert.True(t, oldSpec.CreatedFromRaw)

			newSpec, err := svc.UpsertBatchSpecInput(adminCtx, UpsertBatchSpecInputOpts{
				RawSpec:         bt.TestRawBatchSpecYAML,
				NamespaceUserID: admin.ID,
			})
			assert.Nil(t, err)
			assert.True(t, newSpec.CreatedFromRaw)
			assert.Equal(t, oldSpec.RandID, newSpec.RandID)
			assert.Equal(t, oldSpec.NamespaceUserID, newSpec.NamespaceUserID)
			assert.Equal(t, oldSpec.NamespaceOrgID, newSpec.NamespaceOrgID)

			// Check that the replaced batch spec was really deleted.
			_, err = s.GetBatchSpec(ctx, store.GetBatchSpecOpts{
				ID: oldSpec.ID,
			})
			assert.Equal(t, store.ErrNoResults, err)
		})

		t.Run("mount error", func(t *testing.T) {
			_, err := svc.UpsertBatchSpecInput(adminCtx, UpsertBatchSpecInputOpts{
				RawSpec: `
name: test-spec
description: A test spec
steps:
  - run: /tmp/sample.sh
    container: alpine:3
    mount:
      - path: /some/path/sample.sh
        mountpoint: /tmp/sample.sh
changesetTemplate:
  title: Test Mount
  body: Test a mounted path
  branch: test
  commit:
    message: Test
`,
				NamespaceUserID: admin.ID,
			})
			assert.Equal(t, "mounts are not allowed for server-side processing", err.Error())
		})
	})

	t.Run("ValidateChangesetSpecs", func(t *testing.T) {
		batchSpec := bt.CreateBatchSpec(t, ctx, s, "matching-batch-spec", admin.ID, 0)
		conflictingRef := "refs/heads/conflicting-head-ref"
		for _, opts := range []bt.TestSpecOpts{
			{HeadRef: conflictingRef, Typ: btypes.ChangesetSpecTypeBranch, Repo: rs[0].ID, BatchSpec: batchSpec.ID},
			{HeadRef: conflictingRef, Typ: btypes.ChangesetSpecTypeBranch, Repo: rs[1].ID, BatchSpec: batchSpec.ID},
			{HeadRef: conflictingRef, Typ: btypes.ChangesetSpecTypeBranch, Repo: rs[1].ID, BatchSpec: batchSpec.ID},
			{HeadRef: conflictingRef + "-2", Typ: btypes.ChangesetSpecTypeBranch, Repo: rs[2].ID, BatchSpec: batchSpec.ID},
			{HeadRef: conflictingRef + "-2", Typ: btypes.ChangesetSpecTypeBranch, Repo: rs[2].ID, BatchSpec: batchSpec.ID},
			{HeadRef: conflictingRef + "-2", Typ: btypes.ChangesetSpecTypeBranch, Repo: rs[2].ID, BatchSpec: batchSpec.ID},
		} {
			bt.CreateChangesetSpec(t, ctx, s, opts)
		}
		err := svc.ValidateChangesetSpecs(ctx, batchSpec.ID)
		if err == nil {
			t.Fatal("expected error, but got none")
		}

		want := `2 errors when validating changeset specs:
* 2 changeset specs in repo-1-2 use the same branch: refs/heads/conflicting-head-ref
* 3 changeset specs in repo-1-3 use the same branch: refs/heads/conflicting-head-ref-2
`
		if diff := cmp.Diff(want, err.Error()); diff != "" {
			t.Fatalf("wrong error message: %s", diff)
		}
	})

	t.Run("ComputeBatchSpecState", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			spec := testBatchSpec(admin.ID)
			spec.CreatedFromRaw = true
			if err := s.CreateBatchSpec(ctx, spec); err != nil {
				t.Fatal(err)
			}

			job := &btypes.BatchSpecResolutionJob{
				BatchSpecID: spec.ID,
				InitiatorID: admin.ID,
			}
			if err := s.CreateBatchSpecResolutionJob(ctx, job); err != nil {
				t.Fatal(err)
			}

			startedAt := clock()
			for _, repo := range rs {
				ws := &btypes.BatchSpecWorkspace{BatchSpecID: spec.ID, RepoID: repo.ID}
				if err := s.CreateBatchSpecWorkspace(ctx, ws); err != nil {
					t.Fatal(err)
				}

				job := &btypes.BatchSpecWorkspaceExecutionJob{BatchSpecWorkspaceID: ws.ID, UserID: user.ID}
				if err := bt.CreateBatchSpecWorkspaceExecutionJob(ctx, s, store.ScanBatchSpecWorkspaceExecutionJob, job); err != nil {
					t.Fatal(err)
				}

				job.State = btypes.BatchSpecWorkspaceExecutionJobStateProcessing
				job.StartedAt = startedAt
				bt.UpdateJobState(t, ctx, s, job)
			}

			have, err := svc.LoadBatchSpecStats(ctx, spec)
			if err != nil {
				t.Fatal(err)
			}
			want := btypes.BatchSpecStats{
				Workspaces: len(rs),
				Executions: len(rs),
				Processing: len(rs),
				StartedAt:  startedAt,
			}

			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatalf("wrong stats: %s", diff)
			}
		})
	})

	t.Run("RetryBatchSpecWorkspaces", func(t *testing.T) {
		failureMessage := "this failed"

		t.Run("success", func(t *testing.T) {
			spec := testBatchSpec(admin.ID)
			if err := s.CreateBatchSpec(ctx, spec); err != nil {
				t.Fatal(err)
			}

			changesetSpec1 := bt.CreateChangesetSpec(t, ctx, s, bt.TestSpecOpts{
				Repo:      rs[2].ID,
				BatchSpec: spec.ID,
				HeadRef:   "refs/heads/my-spec",
				Typ:       btypes.ChangesetSpecTypeBranch,
			})

			changesetSpec2 := bt.CreateChangesetSpec(t, ctx, s, bt.TestSpecOpts{
				Repo:      rs[2].ID,
				BatchSpec: spec.ID,
				HeadRef:   "refs/heads/my-spec-2",
				Typ:       btypes.ChangesetSpecTypeBranch,
			})

			var workspaceIDs []int64
			for i, repo := range rs {
				ws := &btypes.BatchSpecWorkspace{
					BatchSpecID: spec.ID,
					RepoID:      repo.ID,
				}
				// This workspace has the completed job and resulted in 2 changesetspecs
				if i == 2 {
					ws.ChangesetSpecIDs = append(ws.ChangesetSpecIDs, changesetSpec1.ID, changesetSpec2.ID)
				}

				if err := s.CreateBatchSpecWorkspace(ctx, ws); err != nil {
					t.Fatal(err)
				}
				workspaceIDs = append(workspaceIDs, ws.ID)
			}

			failedJob := &btypes.BatchSpecWorkspaceExecutionJob{
				BatchSpecWorkspaceID: workspaceIDs[0],
				State:                btypes.BatchSpecWorkspaceExecutionJobStateFailed,
				StartedAt:            time.Now(),
				FinishedAt:           time.Now(),
				FailureMessage:       &failureMessage,
			}
			createJob(t, s, failedJob)

			completedJob := &btypes.BatchSpecWorkspaceExecutionJob{
				BatchSpecWorkspaceID: workspaceIDs[2],
				State:                btypes.BatchSpecWorkspaceExecutionJobStateCompleted,
				StartedAt:            time.Now(),
				FinishedAt:           time.Now(),
			}
			createJob(t, s, completedJob)

			jobs := []*btypes.BatchSpecWorkspaceExecutionJob{failedJob, completedJob}

			// RETRY
			if err := svc.RetryBatchSpecWorkspaces(ctx, workspaceIDs); err != nil {
				t.Fatal(err)
			}

			assertJobsDeleted(t, s, jobs)
			assertChangesetSpecsDeleted(t, s, []*btypes.ChangesetSpec{changesetSpec1, changesetSpec2})
			assertJobsCreatedFor(t, s, []int64{workspaceIDs[0], workspaceIDs[1], workspaceIDs[2]})
		})

		t.Run("batch spec already applied", func(t *testing.T) {
			spec := testBatchSpec(admin.ID)
			if err := s.CreateBatchSpec(ctx, spec); err != nil {
				t.Fatal(err)
			}

			batchChange := testBatchChange(spec.UserID, spec)
			if err := s.CreateBatchChange(ctx, batchChange); err != nil {
				t.Fatal(err)
			}

			ws := &btypes.BatchSpecWorkspace{
				BatchSpecID: spec.ID,
				RepoID:      rs[0].ID,
			}

			if err := s.CreateBatchSpecWorkspace(ctx, ws); err != nil {
				t.Fatal(err)
			}

			failedJob := &btypes.BatchSpecWorkspaceExecutionJob{
				BatchSpecWorkspaceID: ws.ID,
				State:                btypes.BatchSpecWorkspaceExecutionJobStateFailed,
				StartedAt:            time.Now(),
				FinishedAt:           time.Now(),
				FailureMessage:       &failureMessage,
			}
			createJob(t, s, failedJob)

			// RETRY
			err := svc.RetryBatchSpecWorkspaces(ctx, []int64{ws.ID})
			if err == nil {
				t.Fatal("no error")
			}
			if err.Error() != "batch spec already applied" {
				t.Fatalf("wrong error: %s", err)
			}
		})

		t.Run("batch spec associated with draft batch change", func(t *testing.T) {
			spec := testBatchSpec(admin.ID)
			if err := s.CreateBatchSpec(ctx, spec); err != nil {
				t.Fatal(err)
			}

			// Associate with draft batch change
			batchChange := testDraftBatchChange(spec.UserID, spec)
			if err := s.CreateBatchChange(ctx, batchChange); err != nil {
				t.Fatal(err)
			}

			ws := &btypes.BatchSpecWorkspace{
				BatchSpecID: spec.ID,
				RepoID:      rs[0].ID,
			}

			if err := s.CreateBatchSpecWorkspace(ctx, ws); err != nil {
				t.Fatal(err)
			}

			failedJob := &btypes.BatchSpecWorkspaceExecutionJob{
				BatchSpecWorkspaceID: ws.ID,
				State:                btypes.BatchSpecWorkspaceExecutionJobStateFailed,
				StartedAt:            time.Now(),
				FinishedAt:           time.Now(),
				FailureMessage:       &failureMessage,
			}
			createJob(t, s, failedJob)

			// RETRY
			err := svc.RetryBatchSpecWorkspaces(ctx, []int64{ws.ID})
			if err != nil {
				t.Fatal("unexpected error")
			}

			assertJobsDeleted(t, s, []*btypes.BatchSpecWorkspaceExecutionJob{failedJob})
			assertJobsCreatedFor(t, s, []int64{ws.ID})
		})

		t.Run("job not retryable", func(t *testing.T) {
			spec := testBatchSpec(admin.ID)
			if err := s.CreateBatchSpec(ctx, spec); err != nil {
				t.Fatal(err)
			}

			ws := &btypes.BatchSpecWorkspace{
				BatchSpecID: spec.ID,
				RepoID:      rs[0].ID,
			}

			if err := s.CreateBatchSpecWorkspace(ctx, ws); err != nil {
				t.Fatal(err)
			}

			queuedJob := &btypes.BatchSpecWorkspaceExecutionJob{
				BatchSpecWorkspaceID: ws.ID,
				State:                btypes.BatchSpecWorkspaceExecutionJobStateQueued,
			}
			createJob(t, s, queuedJob)

			// RETRY
			err := svc.RetryBatchSpecWorkspaces(ctx, []int64{ws.ID})
			if err == nil {
				t.Fatal("no error")
			}
			if !strings.Contains(err.Error(), "not retryable") {
				t.Fatalf("wrong error: %s", err)
			}
		})

		t.Run("user is not namespace user and not admin", func(t *testing.T) {
			// admin owns batch spec
			spec := testBatchSpec(admin.ID)
			if err := s.CreateBatchSpec(ctx, spec); err != nil {
				t.Fatal(err)
			}

			ws := testWorkspace(spec.ID, rs[0].ID)
			if err := s.CreateBatchSpecWorkspace(ctx, ws); err != nil {
				t.Fatal(err)
			}

			queuedJob := &btypes.BatchSpecWorkspaceExecutionJob{
				BatchSpecWorkspaceID: ws.ID,
				State:                btypes.BatchSpecWorkspaceExecutionJobStateQueued,
			}
			createJob(t, s, queuedJob)

			// userCtx uses user as actor
			err := svc.RetryBatchSpecWorkspaces(userCtx, []int64{ws.ID})
			assertAuthError(t, err)
		})
	})

	t.Run("RetryBatchSpecExecution", func(t *testing.T) {
		failureMessage := "this failed"

		createSpec := func(t *testing.T) *btypes.BatchSpec {
			t.Helper()

			spec := testBatchSpec(admin.ID)
			spec.CreatedFromRaw = true
			if err := s.CreateBatchSpec(ctx, spec); err != nil {
				t.Fatal(err)
			}
			job := &btypes.BatchSpecResolutionJob{
				BatchSpecID: spec.ID,
				State:       btypes.BatchSpecResolutionJobStateCompleted,
				InitiatorID: admin.ID,
			}
			if err := s.CreateBatchSpecResolutionJob(ctx, job); err != nil {
				t.Fatal(err)
			}

			return spec
		}

		t.Run("success", func(t *testing.T) {
			spec := createSpec(t)

			changesetSpec1 := bt.CreateChangesetSpec(t, ctx, s, bt.TestSpecOpts{
				Repo:      rs[2].ID,
				BatchSpec: spec.ID,
				HeadRef:   "refs/heads/my-spec",
				Typ:       btypes.ChangesetSpecTypeBranch,
			})

			changesetSpec2 := bt.CreateChangesetSpec(t, ctx, s, bt.TestSpecOpts{
				Repo:      rs[2].ID,
				BatchSpec: spec.ID,
				HeadRef:   "refs/heads/my-spec-2",
				Typ:       btypes.ChangesetSpecTypeBranch,
			})

			var workspaceIDs []int64
			for i, repo := range rs {
				ws := &btypes.BatchSpecWorkspace{
					BatchSpecID: spec.ID,
					RepoID:      repo.ID,
				}
				// This workspace has the completed job and resulted in 2 changesetspecs
				if i == 2 {
					ws.ChangesetSpecIDs = append(ws.ChangesetSpecIDs, changesetSpec1.ID, changesetSpec2.ID)
				}

				if err := s.CreateBatchSpecWorkspace(ctx, ws); err != nil {
					t.Fatal(err)
				}
				workspaceIDs = append(workspaceIDs, ws.ID)
			}

			failedJob := &btypes.BatchSpecWorkspaceExecutionJob{
				BatchSpecWorkspaceID: workspaceIDs[0],
				State:                btypes.BatchSpecWorkspaceExecutionJobStateFailed,
				StartedAt:            time.Now(),
				FinishedAt:           time.Now(),
				FailureMessage:       &failureMessage,
			}
			createJob(t, s, failedJob)

			completedJob1 := &btypes.BatchSpecWorkspaceExecutionJob{
				BatchSpecWorkspaceID: workspaceIDs[1],
				State:                btypes.BatchSpecWorkspaceExecutionJobStateCompleted,
				StartedAt:            time.Now(),
				FinishedAt:           time.Now(),
			}
			createJob(t, s, completedJob1)

			completedJob2 := &btypes.BatchSpecWorkspaceExecutionJob{
				BatchSpecWorkspaceID: workspaceIDs[2],
				State:                btypes.BatchSpecWorkspaceExecutionJobStateCompleted,
				StartedAt:            time.Now(),
				FinishedAt:           time.Now(),
			}
			createJob(t, s, completedJob2)

			// RETRY
			if err := svc.RetryBatchSpecExecution(ctx, RetryBatchSpecExecutionOpts{BatchSpecRandID: spec.RandID}); err != nil {
				t.Fatal(err)
			}

			// Completed jobs should not be retried
			assertJobsDeleted(t, s, []*btypes.BatchSpecWorkspaceExecutionJob{failedJob})
			assertChangesetSpecsNotDeleted(t, s, []*btypes.ChangesetSpec{changesetSpec1, changesetSpec2})
			assertJobsCreatedFor(t, s, []int64{workspaceIDs[0], workspaceIDs[1], workspaceIDs[2]})
		})

		t.Run("success with IncludeCompleted", func(t *testing.T) {
			spec := createSpec(t)

			changesetSpec1 := bt.CreateChangesetSpec(t, ctx, s, bt.TestSpecOpts{
				Repo:      rs[2].ID,
				BatchSpec: spec.ID,
				HeadRef:   "refs/heads/my-spec",
				Typ:       btypes.ChangesetSpecTypeBranch,
			})

			changesetSpec2 := bt.CreateChangesetSpec(t, ctx, s, bt.TestSpecOpts{
				Repo:      rs[2].ID,
				BatchSpec: spec.ID,
				HeadRef:   "refs/heads/my-spec-2",
				Typ:       btypes.ChangesetSpecTypeBranch,
			})

			var workspaceIDs []int64
			for i, repo := range rs {
				ws := &btypes.BatchSpecWorkspace{
					BatchSpecID: spec.ID,
					RepoID:      repo.ID,
				}
				// This workspace has the completed job and resulted in 2 changesetspecs
				if i == 2 {
					ws.ChangesetSpecIDs = append(ws.ChangesetSpecIDs, changesetSpec1.ID, changesetSpec2.ID)
				}

				if err := s.CreateBatchSpecWorkspace(ctx, ws); err != nil {
					t.Fatal(err)
				}
				workspaceIDs = append(workspaceIDs, ws.ID)
			}

			failedJob := &btypes.BatchSpecWorkspaceExecutionJob{
				BatchSpecWorkspaceID: workspaceIDs[0],
				State:                btypes.BatchSpecWorkspaceExecutionJobStateFailed,
				StartedAt:            time.Now(),
				FinishedAt:           time.Now(),
				FailureMessage:       &failureMessage,
			}
			createJob(t, s, failedJob)

			completedJob1 := &btypes.BatchSpecWorkspaceExecutionJob{
				BatchSpecWorkspaceID: workspaceIDs[1],
				State:                btypes.BatchSpecWorkspaceExecutionJobStateCompleted,
				StartedAt:            time.Now(),
				FinishedAt:           time.Now(),
			}
			createJob(t, s, completedJob1)

			completedJob2 := &btypes.BatchSpecWorkspaceExecutionJob{
				BatchSpecWorkspaceID: workspaceIDs[2],
				State:                btypes.BatchSpecWorkspaceExecutionJobStateCompleted,
				StartedAt:            time.Now(),
				FinishedAt:           time.Now(),
			}
			createJob(t, s, completedJob2)

			// RETRY
			opts := RetryBatchSpecExecutionOpts{BatchSpecRandID: spec.RandID, IncludeCompleted: true}
			if err := svc.RetryBatchSpecExecution(ctx, opts); err != nil {
				t.Fatal(err)
			}

			// Queued job should not be deleted
			assertJobsDeleted(t, s, []*btypes.BatchSpecWorkspaceExecutionJob{
				failedJob,
				completedJob1,
				completedJob2,
			})
			assertChangesetSpecsDeleted(t, s, []*btypes.ChangesetSpec{changesetSpec1, changesetSpec2})
			assertJobsCreatedFor(t, s, []int64{workspaceIDs[0], workspaceIDs[1], workspaceIDs[2]})
		})

		t.Run("batch spec already applied", func(t *testing.T) {
			spec := createSpec(t)

			batchChange := testBatchChange(spec.UserID, spec)
			if err := s.CreateBatchChange(ctx, batchChange); err != nil {
				t.Fatal(err)
			}

			ws := &btypes.BatchSpecWorkspace{
				BatchSpecID: spec.ID,
				RepoID:      rs[0].ID,
			}

			if err := s.CreateBatchSpecWorkspace(ctx, ws); err != nil {
				t.Fatal(err)
			}

			failedJob := &btypes.BatchSpecWorkspaceExecutionJob{
				BatchSpecWorkspaceID: ws.ID,
				State:                btypes.BatchSpecWorkspaceExecutionJobStateFailed,
				StartedAt:            time.Now(),
				FinishedAt:           time.Now(),
				FailureMessage:       &failureMessage,
			}
			createJob(t, s, failedJob)

			// RETRY
			err := svc.RetryBatchSpecExecution(ctx, RetryBatchSpecExecutionOpts{BatchSpecRandID: spec.RandID})
			if err == nil {
				t.Fatal("no error")
			}
			if err.Error() != "batch spec already applied" {
				t.Fatalf("wrong error: %s", err)
			}
		})

		t.Run("batch spec associated with draft batch change", func(t *testing.T) {
			spec := testBatchSpec(admin.ID)
			if err := s.CreateBatchSpec(ctx, spec); err != nil {
				t.Fatal(err)
			}

			// Associate with draft batch change
			batchChange := testDraftBatchChange(spec.UserID, spec)
			if err := s.CreateBatchChange(ctx, batchChange); err != nil {
				t.Fatal(err)
			}

			ws := &btypes.BatchSpecWorkspace{
				BatchSpecID: spec.ID,
				RepoID:      rs[0].ID,
			}

			if err := s.CreateBatchSpecWorkspace(ctx, ws); err != nil {
				t.Fatal(err)
			}

			failedJob := &btypes.BatchSpecWorkspaceExecutionJob{
				BatchSpecWorkspaceID: ws.ID,
				State:                btypes.BatchSpecWorkspaceExecutionJobStateFailed,
				StartedAt:            time.Now(),
				FinishedAt:           time.Now(),
				FailureMessage:       &failureMessage,
			}
			createJob(t, s, failedJob)

			// RETRY
			opts := RetryBatchSpecExecutionOpts{BatchSpecRandID: spec.RandID, IncludeCompleted: true}
			if err := svc.RetryBatchSpecExecution(ctx, opts); err != nil {
				t.Fatal(err)
			}

			// Queued job should not be deleted
			assertJobsDeleted(t, s, []*btypes.BatchSpecWorkspaceExecutionJob{
				failedJob,
			})
			assertJobsCreatedFor(t, s, []int64{ws.ID})
		})

		t.Run("user is not namespace user and not admin", func(t *testing.T) {
			// admin owns batch spec
			spec := createSpec(t)

			ws := testWorkspace(spec.ID, rs[0].ID)
			if err := s.CreateBatchSpecWorkspace(ctx, ws); err != nil {
				t.Fatal(err)
			}

			queuedJob := &btypes.BatchSpecWorkspaceExecutionJob{
				BatchSpecWorkspaceID: ws.ID,
				State:                btypes.BatchSpecWorkspaceExecutionJobStateQueued,
			}
			createJob(t, s, queuedJob)

			// userCtx uses user as actor
			err := svc.RetryBatchSpecExecution(userCtx, RetryBatchSpecExecutionOpts{BatchSpecRandID: spec.RandID})
			assertAuthError(t, err)
		})

		t.Run("batch spec not in final state", func(t *testing.T) {
			spec := createSpec(t)

			ws := &btypes.BatchSpecWorkspace{
				BatchSpecID: spec.ID,
				RepoID:      rs[0].ID,
			}

			if err := s.CreateBatchSpecWorkspace(ctx, ws); err != nil {
				t.Fatal(err)
			}

			queuedJob := &btypes.BatchSpecWorkspaceExecutionJob{
				BatchSpecWorkspaceID: ws.ID,
				State:                btypes.BatchSpecWorkspaceExecutionJobStateQueued,
			}
			createJob(t, s, queuedJob)

			// RETRY
			err := svc.RetryBatchSpecExecution(ctx, RetryBatchSpecExecutionOpts{BatchSpecRandID: spec.RandID})
			if err == nil {
				t.Fatal("no error")
			}
			if !errors.Is(err, ErrRetryNonFinal) {
				t.Fatalf("wrong error: %s", err)
			}
		})
	})

	t.Run("GetAvailableBulkOperations", func(t *testing.T) {
		spec := testBatchSpec(admin.ID)
		if err := s.CreateBatchSpec(ctx, spec); err != nil {
			t.Fatal(err)
		}

		batchChange := testBatchChange(admin.ID, spec)
		if err := s.CreateBatchChange(ctx, batchChange); err != nil {
			t.Fatal(err)
		}

		t.Run("failed changesets", func(t *testing.T) {
			changeset := bt.CreateChangeset(t, ctx, s, bt.TestChangesetOpts{
				Repo:               rs[0].ID,
				PublicationState:   btypes.ChangesetPublicationStatePublished,
				BatchChange:        batchChange.ID,
				ReconcilerState:    btypes.ReconcilerStateFailed,
				OwnedByBatchChange: batchChange.ID,
			})

			bulkOperations, err := svc.GetAvailableBulkOperations(ctx, GetAvailableBulkOperationsOpts{
				Changesets: []int64{
					changeset.ID,
				},
				BatchChange: batchChange.ID,
			})

			if err != nil {
				t.Fatal(err)
			}

			expectedBulkOperations := []string{"REENQUEUE", "PUBLISH", "CLOSE"}
			if !assert.ElementsMatch(t, expectedBulkOperations, bulkOperations) {
				t.Errorf("wrong bulk operation type returned. want=%q, have=%q", expectedBulkOperations, bulkOperations)
			}
		})

		t.Run("archived changesets", func(t *testing.T) {
			changeset := bt.CreateChangeset(t, ctx, s, bt.TestChangesetOpts{
				Repo:               rs[0].ID,
				PublicationState:   btypes.ChangesetPublicationStatePublished,
				BatchChange:        batchChange.ID,
				OwnedByBatchChange: batchChange.ID,

				// archived changeset
				IsArchived: true,
			})

			bulkOperations, err := svc.GetAvailableBulkOperations(ctx, GetAvailableBulkOperationsOpts{
				Changesets: []int64{
					changeset.ID,
				},
				BatchChange: batchChange.ID,
			})

			if err != nil {
				t.Fatal(err)
			}

			expectedBulkOperations := []string{"DETACH"}
			if !assert.ElementsMatch(t, expectedBulkOperations, bulkOperations) {
				t.Errorf("wrong bulk operation type returned. want=%q, have=%q", expectedBulkOperations, bulkOperations)
			}
		})

		t.Run("unpublished changesets", func(t *testing.T) {
			changeset := bt.CreateChangeset(t, ctx, s, bt.TestChangesetOpts{
				Repo:               rs[0].ID,
				PublicationState:   btypes.ChangesetPublicationStateUnpublished,
				BatchChange:        batchChange.ID,
				OwnedByBatchChange: batchChange.ID,
			})

			bulkOperations, err := svc.GetAvailableBulkOperations(ctx, GetAvailableBulkOperationsOpts{
				Changesets: []int64{
					changeset.ID,
				},
				BatchChange: batchChange.ID,
			})

			if err != nil {
				t.Fatal(err)
			}

			expectedBulkOperations := []string{"PUBLISH"}
			if !assert.ElementsMatch(t, expectedBulkOperations, bulkOperations) {
				t.Errorf("wrong bulk operation type returned. want=%q, have=%q", expectedBulkOperations, bulkOperations)
			}
		})

		t.Run("draft changesets", func(t *testing.T) {
			changeset := bt.CreateChangeset(t, ctx, s, bt.TestChangesetOpts{
				Repo:               rs[0].ID,
				PublicationState:   btypes.ChangesetPublicationStatePublished,
				BatchChange:        batchChange.ID,
				ExternalState:      btypes.ChangesetExternalStateDraft,
				OwnedByBatchChange: batchChange.ID,
			})

			bulkOperations, err := svc.GetAvailableBulkOperations(ctx, GetAvailableBulkOperationsOpts{
				Changesets: []int64{
					changeset.ID,
				},
				BatchChange: batchChange.ID,
			})

			if err != nil {
				t.Fatal(err)
			}

			expectedBulkOperations := []string{"CLOSE", "COMMENT", "PUBLISH"}
			if !assert.ElementsMatch(t, expectedBulkOperations, bulkOperations) {
				t.Errorf("wrong bulk operation type returned. want=%q, have=%q", expectedBulkOperations, bulkOperations)
			}
		})

		t.Run("open changesets", func(t *testing.T) {
			changeset := bt.CreateChangeset(t, ctx, s, bt.TestChangesetOpts{
				Repo:               rs[0].ID,
				PublicationState:   btypes.ChangesetPublicationStatePublished,
				BatchChange:        batchChange.ID,
				OwnedByBatchChange: batchChange.ID,
				ExternalState:      btypes.ChangesetExternalStateOpen,
			})

			bulkOperations, err := svc.GetAvailableBulkOperations(ctx, GetAvailableBulkOperationsOpts{
				Changesets: []int64{
					changeset.ID,
				},
				BatchChange: batchChange.ID,
			})

			if err != nil {
				t.Fatal(err)
			}

			expectedBulkOperations := []string{"CLOSE", "COMMENT", "MERGE", "PUBLISH"}
			if !assert.ElementsMatch(t, expectedBulkOperations, bulkOperations) {
				t.Errorf("wrong bulk operation type returned. want=%q, have=%q", expectedBulkOperations, bulkOperations)
			}
		})

		t.Run("closed changesets", func(t *testing.T) {
			changeset := bt.CreateChangeset(t, ctx, s, bt.TestChangesetOpts{
				Repo:               rs[0].ID,
				PublicationState:   btypes.ChangesetPublicationStatePublished,
				BatchChange:        batchChange.ID,
				OwnedByBatchChange: batchChange.ID,
				ExternalState:      btypes.ChangesetExternalStateClosed,
			})

			bulkOperations, err := svc.GetAvailableBulkOperations(ctx, GetAvailableBulkOperationsOpts{
				Changesets: []int64{
					changeset.ID,
				},
				BatchChange: batchChange.ID,
			})

			if err != nil {
				t.Fatal(err)
			}

			expectedBulkOperations := []string{"COMMENT", "PUBLISH"}
			if !assert.ElementsMatch(t, expectedBulkOperations, bulkOperations) {
				t.Errorf("wrong bulk operation type returned. want=%q, have=%q", expectedBulkOperations, bulkOperations)
			}
		})

		t.Run("merged changesets", func(t *testing.T) {
			changeset := bt.CreateChangeset(t, ctx, s, bt.TestChangesetOpts{
				Repo:               rs[0].ID,
				PublicationState:   btypes.ChangesetPublicationStatePublished,
				BatchChange:        batchChange.ID,
				OwnedByBatchChange: batchChange.ID,
				ExternalState:      btypes.ChangesetExternalStateMerged,
			})

			bulkOperations, err := svc.GetAvailableBulkOperations(ctx, GetAvailableBulkOperationsOpts{
				Changesets: []int64{
					changeset.ID,
				},
				BatchChange: batchChange.ID,
			})

			if err != nil {
				t.Fatal(err)
			}

			expectedBulkOperations := []string{"COMMENT", "PUBLISH"}
			if !assert.ElementsMatch(t, expectedBulkOperations, bulkOperations) {
				t.Errorf("wrong bulk operation type returned. want=%q, have=%q", expectedBulkOperations, bulkOperations)
			}
		})

		t.Run("read-only changesets", func(t *testing.T) {
			changeset := bt.CreateChangeset(t, ctx, s, bt.TestChangesetOpts{
				Repo:               rs[0].ID,
				PublicationState:   btypes.ChangesetPublicationStatePublished,
				ExternalState:      btypes.ChangesetExternalStateReadOnly,
				BatchChange:        batchChange.ID,
				OwnedByBatchChange: batchChange.ID,
			})

			bulkOperations, err := svc.GetAvailableBulkOperations(ctx, GetAvailableBulkOperationsOpts{
				Changesets: []int64{
					changeset.ID,
				},
				BatchChange: batchChange.ID,
			})

			assert.NoError(t, err)
			assert.Empty(t, bulkOperations)
		})

		t.Run("imported changesets", func(t *testing.T) {
			changeset := bt.CreateChangeset(t, ctx, s, bt.TestChangesetOpts{
				Repo:               rs[0].ID,
				PublicationState:   btypes.ChangesetPublicationStatePublished,
				ExternalState:      btypes.ChangesetExternalStateOpen,
				OwnedByBatchChange: 0,
			})

			bulkOperations, err := svc.GetAvailableBulkOperations(ctx, GetAvailableBulkOperationsOpts{
				Changesets: []int64{
					changeset.ID,
				},
				BatchChange: batchChange.ID,
			})

			assert.NoError(t, err)
			expectedBulkOperations := []string{"COMMENT", "CLOSE", "MERGE"}
			if !assert.ElementsMatch(t, expectedBulkOperations, bulkOperations) {
				t.Errorf("wrong bulk operation type returned. want=%q, have=%q", expectedBulkOperations, bulkOperations)
			}
		})

		t.Run("draft, archived and failed changesets with no common bulk operation", func(t *testing.T) {
			failedChangeset := bt.CreateChangeset(t, ctx, s, bt.TestChangesetOpts{
				Repo:               rs[0].ID,
				PublicationState:   btypes.ChangesetPublicationStatePublished,
				BatchChange:        batchChange.ID,
				OwnedByBatchChange: batchChange.ID,
				ReconcilerState:    btypes.ReconcilerStateFailed,
			})

			archivedChangeset := bt.CreateChangeset(t, ctx, s, bt.TestChangesetOpts{
				Repo:               rs[0].ID,
				PublicationState:   btypes.ChangesetPublicationStatePublished,
				BatchChange:        batchChange.ID,
				OwnedByBatchChange: batchChange.ID,

				// archived changeset
				IsArchived: true,
			})

			draftChangeset := bt.CreateChangeset(t, ctx, s, bt.TestChangesetOpts{
				Repo:               rs[0].ID,
				PublicationState:   btypes.ChangesetPublicationStatePublished,
				BatchChange:        batchChange.ID,
				OwnedByBatchChange: batchChange.ID,
				ExternalState:      btypes.ChangesetExternalStateDraft,
			})

			bulkOperations, err := svc.GetAvailableBulkOperations(ctx, GetAvailableBulkOperationsOpts{
				Changesets: []int64{
					failedChangeset.ID,
					archivedChangeset.ID,
					draftChangeset.ID,
				},
				BatchChange: batchChange.ID,
			})

			if err != nil {
				t.Fatal(err)
			}

			expectedBulkOperations := []string{}
			if !assert.ElementsMatch(t, expectedBulkOperations, bulkOperations) {
				t.Errorf("wrong bulk operation type returned. want=%q, have=%q", expectedBulkOperations, bulkOperations)
			}
		})

		t.Run("draft, closed and merged changesets with a common bulk operation", func(t *testing.T) {
			draftChangeset := bt.CreateChangeset(t, ctx, s, bt.TestChangesetOpts{
				Repo:               rs[0].ID,
				PublicationState:   btypes.ChangesetPublicationStatePublished,
				BatchChange:        batchChange.ID,
				OwnedByBatchChange: batchChange.ID,
				ExternalState:      btypes.ChangesetExternalStateDraft,
			})

			closedChangeset := bt.CreateChangeset(t, ctx, s, bt.TestChangesetOpts{
				Repo:               rs[0].ID,
				PublicationState:   btypes.ChangesetPublicationStatePublished,
				BatchChange:        batchChange.ID,
				OwnedByBatchChange: batchChange.ID,
				ExternalState:      btypes.ChangesetExternalStateClosed,
			})

			mergedChangeset := bt.CreateChangeset(t, ctx, s, bt.TestChangesetOpts{
				Repo:               rs[0].ID,
				PublicationState:   btypes.ChangesetPublicationStatePublished,
				BatchChange:        batchChange.ID,
				OwnedByBatchChange: batchChange.ID,
				ExternalState:      btypes.ChangesetExternalStateMerged,
			})

			bulkOperations, err := svc.GetAvailableBulkOperations(ctx, GetAvailableBulkOperationsOpts{
				Changesets: []int64{
					closedChangeset.ID,
					mergedChangeset.ID,
					draftChangeset.ID,
				},
				BatchChange: batchChange.ID,
			})

			if err != nil {
				t.Fatal(err)
			}

			expectedBulkOperations := []string{"COMMENT", "PUBLISH"}
			if !assert.ElementsMatch(t, expectedBulkOperations, bulkOperations) {
				t.Errorf("wrong bulk operation type returned. want=%q, have=%q", expectedBulkOperations, bulkOperations)
			}
		})
	})

	t.Run("UpsertEmptyBatchChange", func(t *testing.T) {
		t.Run("creates new batch change if it is non-existent", func(t *testing.T) {
			name := "random-bc-name"

			// verify that the batch change doesn't exist
			_, err := s.GetBatchChange(ctx, store.GetBatchChangeOpts{
				Name:            name,
				NamespaceUserID: user.ID,
			})

			if err != store.ErrNoResults {
				t.Fatalf("batch change %s should not exist", name)
			}

			batchChange, err := svc.UpsertEmptyBatchChange(ctx, UpsertEmptyBatchChangeOpts{
				Name:            name,
				NamespaceUserID: user.ID,
			})
			if err != nil {
				t.Fatal(err)
			}

			if batchChange.ID == 0 {
				t.Fatalf("BatchChange ID is 0")
			}

			if have, want := batchChange.NamespaceUserID, user.ID; have != want {
				t.Fatalf("UserID is %d, want %d", have, want)
			}
		})

		t.Run("returns existing Batch Change", func(t *testing.T) {
			spec := &btypes.BatchSpec{
				UserID:          user.ID,
				NamespaceUserID: user.ID,
				NamespaceOrgID:  0,
			}
			if err := s.CreateBatchSpec(ctx, spec); err != nil {
				t.Fatal(err)
			}

			bc := testBatchChange(user.ID, spec)
			if err := s.CreateBatchChange(ctx, bc); err != nil {
				t.Fatal(err)
			}

			haveBatchChange, err := svc.UpsertEmptyBatchChange(ctx, UpsertEmptyBatchChangeOpts{
				Name:            bc.Name,
				NamespaceUserID: user.ID,
			})
			if err != nil {
				t.Fatal(err)
			}

			if haveBatchChange == nil {
				t.Fatal("expected to have matching batch change, but got nil")
			}

			if haveBatchChange.ID == 0 {
				t.Fatal("BatchChange ID is 0")
			}

			if haveBatchChange.ID != bc.ID {
				t.Fatal("expected same ID for batch change")
			}

			if haveBatchChange.BatchSpecID == bc.BatchSpecID {
				t.Fatal("expected different spec ID for batch change")
			}
		})
	})
}

func createJob(t *testing.T, s *store.Store, job *btypes.BatchSpecWorkspaceExecutionJob) {
	t.Helper()

	if job.UserID == 0 {
		job.UserID = 1
	}

	clone := *job

	if err := bt.CreateBatchSpecWorkspaceExecutionJob(context.Background(), s, store.ScanBatchSpecWorkspaceExecutionJob, job); err != nil {
		t.Fatal(err)
	}

	job.State = clone.State
	job.Cancel = clone.Cancel
	job.WorkerHostname = clone.WorkerHostname
	job.StartedAt = clone.StartedAt
	job.FinishedAt = clone.FinishedAt
	job.FailureMessage = clone.FailureMessage

	bt.UpdateJobState(t, context.Background(), s, job)
}

func assertJobsDeleted(t *testing.T, s *store.Store, jobs []*btypes.BatchSpecWorkspaceExecutionJob) {
	t.Helper()

	jobIDs := make([]int64, len(jobs))
	for i, j := range jobs {
		jobIDs[i] = j.ID
	}
	old, err := s.ListBatchSpecWorkspaceExecutionJobs(context.Background(), store.ListBatchSpecWorkspaceExecutionJobsOpts{
		IDs: jobIDs,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(old) != 0 {
		t.Fatal("old jobs not deleted")
	}
}

func assertJobsCreatedFor(t *testing.T, s *store.Store, workspaceIDs []int64) {
	t.Helper()

	idMap := make(map[int64]struct{}, len(workspaceIDs))
	for _, id := range workspaceIDs {
		idMap[id] = struct{}{}
	}
	jobs, err := s.ListBatchSpecWorkspaceExecutionJobs(context.Background(), store.ListBatchSpecWorkspaceExecutionJobsOpts{
		BatchSpecWorkspaceIDs: workspaceIDs,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != len(workspaceIDs) {
		t.Fatal("jobs not created")
	}
	for _, job := range jobs {
		if _, ok := idMap[job.BatchSpecWorkspaceID]; !ok {
			t.Fatalf("job created for wrong workspace")
		}
	}
}

func assertChangesetSpecsDeleted(t *testing.T, s *store.Store, specs []*btypes.ChangesetSpec) {
	t.Helper()

	ids := make([]int64, len(specs))
	for i, j := range specs {
		ids[i] = j.ID
	}
	old, _, err := s.ListChangesetSpecs(context.Background(), store.ListChangesetSpecsOpts{
		IDs: ids,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(old) != 0 {
		t.Fatal("specs not deleted")
	}
}

func assertChangesetSpecsNotDeleted(t *testing.T, s *store.Store, specs []*btypes.ChangesetSpec) {
	t.Helper()

	ids := make([]int64, len(specs))
	for i, j := range specs {
		ids[i] = j.ID
	}
	have, _, err := s.ListChangesetSpecs(context.Background(), store.ListChangesetSpecsOpts{
		IDs: ids,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(have) != len(ids) {
		t.Fatalf("wrong number of changeset specs. want=%d, have=%d", len(ids), len(have))
	}
	haveIDs := make([]int64, len(have))
	for i, j := range have {
		haveIDs[i] = j.ID
	}

	if diff := cmp.Diff(ids, haveIDs); diff != "" {
		t.Fatalf("wrong changeset specs exist: %s", diff)
	}
}

func testBatchChange(user int32, spec *btypes.BatchSpec) *btypes.BatchChange {
	c := &btypes.BatchChange{
		Name:            fmt.Sprintf("test-batch-change-%d", time.Now().UnixMicro()),
		CreatorID:       user,
		NamespaceUserID: user,
		BatchSpecID:     spec.ID,
		LastApplierID:   user,
		LastAppliedAt:   time.Now(),
	}

	return c
}

func testDraftBatchChange(user int32, spec *btypes.BatchSpec) *btypes.BatchChange {
	bc := testBatchChange(user, spec)
	bc.LastAppliedAt = time.Time{}
	bc.CreatorID = 0
	bc.LastApplierID = 0
	return bc
}

func testBatchSpec(user int32) *btypes.BatchSpec {
	return &btypes.BatchSpec{
		Spec:            &batcheslib.BatchSpec{},
		UserID:          user,
		NamespaceUserID: user,
	}
}

//nolint:unparam // unparam complains that `extState` always has same value across call-sites, but that's OK
func testChangeset(repoID api.RepoID, batchChange int64, extState btypes.ChangesetExternalState) *btypes.Changeset {
	changeset := &btypes.Changeset{
		RepoID:              repoID,
		ExternalServiceType: extsvc.TypeGitHub,
		ExternalID:          fmt.Sprintf("ext-id-%d", batchChange),
		Metadata:            &github.PullRequest{State: string(extState), CreatedAt: time.Now()},
		ExternalState:       extState,
	}

	if batchChange != 0 {
		changeset.BatchChanges = []btypes.BatchChangeAssoc{{BatchChangeID: batchChange}}
	}

	return changeset
}

func testWorkspace(batchSpecID int64, repoID api.RepoID) *btypes.BatchSpecWorkspace {
	return &btypes.BatchSpecWorkspace{
		BatchSpecID: batchSpecID,
		RepoID:      repoID,
	}
}

func assertAuthError(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error. got none")
	}
	if err != nil {
		if !errors.HasType(err, &backend.InsufficientAuthorizationError{}) {
			t.Fatalf("wrong error: %s (%T)", err, err)
		}
	}
}

func assertNoAuthError(t *testing.T, err error) {
	t.Helper()

	// Ignore other errors, we only want to check whether it's an auth error
	if errors.HasType(err, &backend.InsufficientAuthorizationError{}) {
		t.Fatalf("got auth error")
	}
}
