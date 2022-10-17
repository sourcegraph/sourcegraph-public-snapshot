package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/reconciler"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	bt "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestServiceApplyBatchChange(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := actor.WithInternalActor(context.Background())
	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	admin := bt.CreateTestUser(t, db, true)
	adminCtx := actor.WithActor(context.Background(), actor.FromUser(admin.ID))

	user := bt.CreateTestUser(t, db, false)
	userCtx := actor.WithActor(context.Background(), actor.FromUser(user.ID))

	repos, _ := bt.CreateTestRepos(t, ctx, db, 4)

	now := timeutil.Now()
	clock := func() time.Time { return now }
	store := store.NewWithClock(db, &observation.TestContext, nil, clock)
	svc := New(store)

	t.Run("BatchSpec without changesetSpecs", func(t *testing.T) {
		t.Run("new batch change", func(t *testing.T) {
			bt.TruncateTables(t, db, "changeset_events", "changesets", "batch_changes", "batch_specs", "changeset_specs")
			batchSpec := bt.CreateBatchSpec(t, ctx, store, "batchchange1", admin.ID, 0)
			batchChange, err := svc.ApplyBatchChange(adminCtx, ApplyBatchChangeOpts{
				BatchSpecRandID: batchSpec.RandID,
			})
			if err != nil {
				t.Fatal(err)
			}

			if batchChange.ID == 0 {
				t.Fatalf("batch change ID is 0")
			}

			want := &btypes.BatchChange{
				Name:            batchSpec.Spec.Name,
				Description:     batchSpec.Spec.Description,
				CreatorID:       admin.ID,
				LastApplierID:   admin.ID,
				LastAppliedAt:   now,
				NamespaceUserID: batchSpec.NamespaceUserID,
				BatchSpecID:     batchSpec.ID,

				// Ignore these fields
				ID:        batchChange.ID,
				UpdatedAt: batchChange.UpdatedAt,
				CreatedAt: batchChange.CreatedAt,
			}

			if diff := cmp.Diff(want, batchChange); diff != "" {
				t.Fatalf("wrong spec fields (-want +got):\n%s", diff)
			}
		})

		t.Run("existing batch change", func(t *testing.T) {
			bt.TruncateTables(t, db, "changeset_events", "changesets", "batch_changes", "batch_specs", "changeset_specs")
			batchSpec := bt.CreateBatchSpec(t, ctx, store, "batchchange2", admin.ID, 0)
			batchChange := bt.CreateBatchChange(t, ctx, store, "batchchange2", admin.ID, batchSpec.ID)

			t.Run("apply same BatchSpec", func(t *testing.T) {
				batchChange2, err := svc.ApplyBatchChange(adminCtx, ApplyBatchChangeOpts{
					BatchSpecRandID: batchSpec.RandID,
				})
				if err != nil {
					t.Fatal(err)
				}

				if have, want := batchChange2.ID, batchChange.ID; have != want {
					t.Fatalf("batch change ID is wrong. want=%d, have=%d", want, have)
				}
			})

			t.Run("apply same BatchSpec with FailIfExists", func(t *testing.T) {
				_, err := svc.ApplyBatchChange(ctx, ApplyBatchChangeOpts{
					BatchSpecRandID:         batchSpec.RandID,
					FailIfBatchChangeExists: true,
				})
				if err != ErrMatchingBatchChangeExists {
					t.Fatalf("unexpected error. want=%s, got=%s", ErrMatchingBatchChangeExists, err)
				}
			})

			t.Run("apply batch spec with same name", func(t *testing.T) {
				batchSpec2 := bt.CreateBatchSpec(t, ctx, store, "batchchange2", admin.ID, 0)
				batchChange2, err := svc.ApplyBatchChange(adminCtx, ApplyBatchChangeOpts{
					BatchSpecRandID: batchSpec2.RandID,
				})
				if err != nil {
					t.Fatal(err)
				}

				if have, want := batchChange2.ID, batchChange.ID; have != want {
					t.Fatalf("batch change ID is wrong. want=%d, have=%d", want, have)
				}
			})

			t.Run("apply batch spec with same name but different current user", func(t *testing.T) {
				batchSpec := bt.CreateBatchSpec(t, ctx, store, "created-by-user", user.ID, 0)
				batchChange := bt.CreateBatchChange(t, ctx, store, "created-by-user", user.ID, batchSpec.ID)

				if have, want := batchChange.CreatorID, user.ID; have != want {
					t.Fatalf("batch change CreatorID is wrong. want=%d, have=%d", want, have)
				}

				if have, want := batchChange.LastApplierID, user.ID; have != want {
					t.Fatalf("batch change LastApplierID is wrong. want=%d, have=%d", want, have)
				}

				batchSpec2 := bt.CreateBatchSpec(t, ctx, store, "created-by-user", user.ID, 0)
				batchChange2, err := svc.ApplyBatchChange(adminCtx, ApplyBatchChangeOpts{
					BatchSpecRandID: batchSpec2.RandID,
				})
				if err != nil {
					t.Fatal(err)
				}

				if have, want := batchChange2.ID, batchChange.ID; have != want {
					t.Fatalf("batch change ID is wrong. want=%d, have=%d", want, have)
				}

				if have, want := batchChange2.CreatorID, batchChange.CreatorID; have != want {
					t.Fatalf("batch change CreatorID is wrong. want=%d, have=%d", want, have)
				}

				if have, want := batchChange2.LastApplierID, admin.ID; have != want {
					t.Fatalf("batch change LastApplierID is wrong. want=%d, have=%d", want, have)
				}
			})

			t.Run("apply batch spec with same name but different namespace", func(t *testing.T) {
				user2 := bt.CreateTestUser(t, db, false)
				batchSpec2 := bt.CreateBatchSpec(t, ctx, store, "batchchange2", user2.ID, 0)

				batchChange2, err := svc.ApplyBatchChange(adminCtx, ApplyBatchChangeOpts{
					BatchSpecRandID: batchSpec2.RandID,
				})
				if err != nil {
					t.Fatal(err)
				}

				if batchChange2.ID == 0 {
					t.Fatalf("batchChange2 ID is 0")
				}

				if batchChange2.ID == batchChange.ID {
					t.Fatalf("batch change IDs are the same, but want different")
				}
			})

			t.Run("batch spec with same name and same ensureBatchChangeID", func(t *testing.T) {
				batchSpec2 := bt.CreateBatchSpec(t, ctx, store, "batchchange2", admin.ID, 0)

				batchChange2, err := svc.ApplyBatchChange(adminCtx, ApplyBatchChangeOpts{
					BatchSpecRandID:     batchSpec2.RandID,
					EnsureBatchChangeID: batchChange.ID,
				})
				if err != nil {
					t.Fatal(err)
				}
				if have, want := batchChange2.ID, batchChange.ID; have != want {
					t.Fatalf("batch change has wrong ID. want=%d, have=%d", want, have)
				}
			})

			t.Run("batch spec with same name but different ensureBatchChangeID", func(t *testing.T) {
				batchSpec2 := bt.CreateBatchSpec(t, ctx, store, "batchchange2", admin.ID, 0)

				_, err := svc.ApplyBatchChange(adminCtx, ApplyBatchChangeOpts{
					BatchSpecRandID:     batchSpec2.RandID,
					EnsureBatchChangeID: batchChange.ID + 999,
				})
				if err != ErrEnsureBatchChangeFailed {
					t.Fatalf("wrong error: %s", err)
				}
			})
		})
	})

	// These tests focus on changesetSpecs and wiring them up with changesets.
	// The applying/re-applying of a batchSpec to an existing batch change is
	// covered in the tests above.
	t.Run("batchSpec with changesetSpecs", func(t *testing.T) {
		t.Run("new batch change", func(t *testing.T) {
			bt.TruncateTables(t, db, "changeset_events", "changesets", "batch_changes", "batch_specs", "changeset_specs")
			batchSpec := bt.CreateBatchSpec(t, ctx, store, "batchchange3", admin.ID, 0)

			spec1 := bt.CreateChangesetSpec(t, ctx, store, bt.TestSpecOpts{
				User:       admin.ID,
				Repo:       repos[0].ID,
				BatchSpec:  batchSpec.ID,
				ExternalID: "1234",
				Typ:        btypes.ChangesetSpecTypeExisting,
			})

			spec2 := bt.CreateChangesetSpec(t, ctx, store, bt.TestSpecOpts{
				User:      admin.ID,
				Repo:      repos[1].ID,
				BatchSpec: batchSpec.ID,
				HeadRef:   "refs/heads/my-branch",
				Typ:       btypes.ChangesetSpecTypeBranch,
			})

			batchChange, cs := applyAndListChangesets(adminCtx, t, svc, batchSpec.RandID, 2)

			if have, want := batchChange.Name, "batchchange3"; have != want {
				t.Fatalf("wrong batch change name. want=%s, have=%s", want, have)
			}

			c1 := cs.Find(btypes.WithExternalID(spec1.ExternalID))
			bt.AssertChangeset(t, c1, bt.ChangesetAssertions{
				Repo:             spec1.BaseRepoID,
				ExternalID:       "1234",
				ReconcilerState:  btypes.ReconcilerStateQueued,
				PublicationState: btypes.ChangesetPublicationStateUnpublished,
				AttachedTo:       []int64{batchChange.ID},
			})

			c2 := cs.Find(btypes.WithCurrentSpecID(spec2.ID))
			bt.AssertChangeset(t, c2, bt.ChangesetAssertions{
				Repo:               spec2.BaseRepoID,
				CurrentSpec:        spec2.ID,
				OwnedByBatchChange: batchChange.ID,
				ReconcilerState:    btypes.ReconcilerStateQueued,
				PublicationState:   btypes.ChangesetPublicationStateUnpublished,
				DiffStat:           bt.TestChangsetSpecDiffStat,
				AttachedTo:         []int64{batchChange.ID},
			})
		})

		t.Run("batch change with changesets", func(t *testing.T) {
			bt.TruncateTables(t, db, "changeset_events", "changesets", "batch_changes", "batch_specs", "changeset_specs")
			// First we create a batchSpec and apply it, so that we have
			// changesets and changesetSpecs in the database, wired up
			// correctly.
			batchSpec1 := bt.CreateBatchSpec(t, ctx, store, "batchchange4", admin.ID, 0)

			bt.CreateChangesetSpec(t, ctx, store, bt.TestSpecOpts{
				User:       admin.ID,
				Repo:       repos[0].ID,
				BatchSpec:  batchSpec1.ID,
				ExternalID: "1234",
				Typ:        btypes.ChangesetSpecTypeExisting,
			})

			bt.CreateChangesetSpec(t, ctx, store, bt.TestSpecOpts{
				User:       admin.ID,
				Repo:       repos[0].ID,
				BatchSpec:  batchSpec1.ID,
				ExternalID: "5678",
				Typ:        btypes.ChangesetSpecTypeExisting,
			})

			oldSpec3 := bt.CreateChangesetSpec(t, ctx, store, bt.TestSpecOpts{
				User:      admin.ID,
				Repo:      repos[1].ID,
				BatchSpec: batchSpec1.ID,
				HeadRef:   "refs/heads/repo-1-branch-1",
				Typ:       btypes.ChangesetSpecTypeBranch,
			})

			oldSpec4 := bt.CreateChangesetSpec(t, ctx, store, bt.TestSpecOpts{
				User:      admin.ID,
				Repo:      repos[2].ID,
				BatchSpec: batchSpec1.ID,
				HeadRef:   "refs/heads/repo-2-branch-1",
				Typ:       btypes.ChangesetSpecTypeBranch,
			})

			// Apply and expect 4 changesets
			oldBatchChange, oldChangesets := applyAndListChangesets(adminCtx, t, svc, batchSpec1.RandID, 4)

			// Now we create another batch spec with the same batch change name
			// and namespace.
			batchSpec2 := bt.CreateBatchSpec(t, ctx, store, "batchchange4", admin.ID, 0)

			// Same
			spec1 := bt.CreateChangesetSpec(t, ctx, store, bt.TestSpecOpts{
				User:       admin.ID,
				Repo:       repos[0].ID,
				BatchSpec:  batchSpec2.ID,
				ExternalID: "1234",
				Typ:        btypes.ChangesetSpecTypeExisting,
			})

			// DIFFERENT: Track #9999 in repo[0]
			spec2 := bt.CreateChangesetSpec(t, ctx, store, bt.TestSpecOpts{
				User:       admin.ID,
				Repo:       repos[0].ID,
				BatchSpec:  batchSpec2.ID,
				ExternalID: "5678",
				Typ:        btypes.ChangesetSpecTypeExisting,
			})

			// Same
			spec3 := bt.CreateChangesetSpec(t, ctx, store, bt.TestSpecOpts{
				User:      admin.ID,
				Repo:      repos[1].ID,
				BatchSpec: batchSpec2.ID,
				HeadRef:   "refs/heads/repo-1-branch-1",
				Typ:       btypes.ChangesetSpecTypeBranch,
			})

			// DIFFERENT: branch changed in repo[2]
			spec4 := bt.CreateChangesetSpec(t, ctx, store, bt.TestSpecOpts{
				User:      admin.ID,
				Repo:      repos[2].ID,
				BatchSpec: batchSpec2.ID,
				HeadRef:   "refs/heads/repo-2-branch-2",
				Typ:       btypes.ChangesetSpecTypeBranch,
			})

			// NEW: repo[3]
			spec5 := bt.CreateChangesetSpec(t, ctx, store, bt.TestSpecOpts{
				User:      admin.ID,
				Repo:      repos[3].ID,
				BatchSpec: batchSpec2.ID,
				HeadRef:   "refs/heads/repo-3-branch-1",
				Typ:       btypes.ChangesetSpecTypeBranch,
			})

			// Before we apply the new batch spec, we make the changeset we
			// expect to be closed to look "published", otherwise it won't be
			// closed.
			wantClosed := oldChangesets.Find(btypes.WithCurrentSpecID(oldSpec4.ID))
			bt.SetChangesetPublished(t, ctx, store, wantClosed, "98765", oldSpec4.HeadRef)

			changeset3 := oldChangesets.Find(btypes.WithCurrentSpecID(oldSpec3.ID))
			bt.SetChangesetPublished(t, ctx, store, changeset3, "12345", oldSpec3.HeadRef)

			// Apply and expect 6 changesets
			batchChange, cs := applyAndListChangesets(adminCtx, t, svc, batchSpec2.RandID, 6)

			if oldBatchChange.ID != batchChange.ID {
				t.Fatal("expected to update batch change, but got a new one")
			}

			// This changeset we want marked as "to be archived" and "to be closed"
			bt.ReloadAndAssertChangeset(t, ctx, store, wantClosed, bt.ChangesetAssertions{
				Repo:         repos[2].ID,
				CurrentSpec:  oldSpec4.ID,
				PreviousSpec: oldSpec4.ID,
				ExternalID:   wantClosed.ExternalID,
				// It's still open, just _marked as to be closed_.
				ExternalState:      btypes.ChangesetExternalStateOpen,
				ExternalBranch:     wantClosed.ExternalBranch,
				OwnedByBatchChange: batchChange.ID,
				ReconcilerState:    btypes.ReconcilerStateQueued,
				PublicationState:   btypes.ChangesetPublicationStatePublished,
				DiffStat:           bt.TestChangsetSpecDiffStat,
				AttachedTo:         []int64{batchChange.ID},
				ArchiveIn:          batchChange.ID,
				Closing:            true,
			})

			c1 := cs.Find(btypes.WithExternalID(spec1.ExternalID))
			bt.AssertChangeset(t, c1, bt.ChangesetAssertions{
				Repo:             repos[0].ID,
				CurrentSpec:      0,
				PreviousSpec:     0,
				ExternalID:       "1234",
				ReconcilerState:  btypes.ReconcilerStateQueued,
				PublicationState: btypes.ChangesetPublicationStateUnpublished,
				AttachedTo:       []int64{batchChange.ID},
			})

			c2 := cs.Find(btypes.WithExternalID(spec2.ExternalID))
			bt.AssertChangeset(t, c2, bt.ChangesetAssertions{
				Repo:             repos[0].ID,
				CurrentSpec:      0,
				PreviousSpec:     0,
				ExternalID:       "5678",
				ReconcilerState:  btypes.ReconcilerStateQueued,
				PublicationState: btypes.ChangesetPublicationStateUnpublished,
				AttachedTo:       []int64{batchChange.ID},
			})

			c3 := cs.Find(btypes.WithCurrentSpecID(spec3.ID))
			bt.AssertChangeset(t, c3, bt.ChangesetAssertions{
				Repo:           repos[1].ID,
				CurrentSpec:    spec3.ID,
				ExternalID:     changeset3.ExternalID,
				ExternalBranch: changeset3.ExternalBranch,
				ExternalState:  btypes.ChangesetExternalStateOpen,
				// Has a previous spec, because it succeeded publishing.
				PreviousSpec:       oldSpec3.ID,
				OwnedByBatchChange: batchChange.ID,
				ReconcilerState:    btypes.ReconcilerStateQueued,
				PublicationState:   btypes.ChangesetPublicationStatePublished,
				DiffStat:           bt.TestChangsetSpecDiffStat,
				AttachedTo:         []int64{batchChange.ID},
			})

			c4 := cs.Find(btypes.WithCurrentSpecID(spec4.ID))
			bt.AssertChangeset(t, c4, bt.ChangesetAssertions{
				Repo:               repos[2].ID,
				CurrentSpec:        spec4.ID,
				OwnedByBatchChange: batchChange.ID,
				ReconcilerState:    btypes.ReconcilerStateQueued,
				PublicationState:   btypes.ChangesetPublicationStateUnpublished,
				DiffStat:           bt.TestChangsetSpecDiffStat,
				AttachedTo:         []int64{batchChange.ID},
			})

			c5 := cs.Find(btypes.WithCurrentSpecID(spec5.ID))
			bt.AssertChangeset(t, c5, bt.ChangesetAssertions{
				Repo:               repos[3].ID,
				CurrentSpec:        spec5.ID,
				OwnedByBatchChange: batchChange.ID,
				ReconcilerState:    btypes.ReconcilerStateQueued,
				PublicationState:   btypes.ChangesetPublicationStateUnpublished,
				DiffStat:           bt.TestChangsetSpecDiffStat,
				AttachedTo:         []int64{batchChange.ID},
			})
		})

		t.Run("batch change tracking changesets owned by another batch change", func(t *testing.T) {
			bt.TruncateTables(t, db, "changeset_events", "changesets", "batch_changes", "batch_specs", "changeset_specs")
			batchSpec1 := bt.CreateBatchSpec(t, ctx, store, "owner-batch-change", admin.ID, 0)

			oldSpec1 := bt.CreateChangesetSpec(t, ctx, store, bt.TestSpecOpts{
				User:      admin.ID,
				Repo:      repos[0].ID,
				BatchSpec: batchSpec1.ID,
				HeadRef:   "refs/heads/repo-0-branch-0",
				Typ:       btypes.ChangesetSpecTypeBranch,
			})

			ownerBatchChange, ownerChangesets := applyAndListChangesets(adminCtx, t, svc, batchSpec1.RandID, 1)

			// Now we update the changeset so it looks like it's been published
			// on the code host.
			c := ownerChangesets[0]
			bt.SetChangesetPublished(t, ctx, store, c, "88888", "refs/heads/repo-0-branch-0")

			// This other batch change tracks the changeset created by the first one
			batchSpec2 := bt.CreateBatchSpec(t, ctx, store, "tracking-batch-change", admin.ID, 0)
			bt.CreateChangesetSpec(t, ctx, store, bt.TestSpecOpts{
				User:       admin.ID,
				Repo:       c.RepoID,
				BatchSpec:  batchSpec2.ID,
				ExternalID: c.ExternalID,
				Typ:        btypes.ChangesetSpecTypeExisting,
			})

			trackingBatchChange, trackedChangesets := applyAndListChangesets(adminCtx, t, svc, batchSpec2.RandID, 1)
			// This should still point to the owner batch change
			c2 := trackedChangesets[0]
			trackedChangesetAssertions := bt.ChangesetAssertions{
				Repo:               c.RepoID,
				CurrentSpec:        oldSpec1.ID,
				OwnedByBatchChange: ownerBatchChange.ID,
				ExternalBranch:     c.ExternalBranch,
				ExternalID:         c.ExternalID,
				ExternalState:      btypes.ChangesetExternalStateOpen,
				ReconcilerState:    btypes.ReconcilerStateCompleted,
				PublicationState:   btypes.ChangesetPublicationStatePublished,
				DiffStat:           bt.TestChangsetSpecDiffStat,
				AttachedTo:         []int64{ownerBatchChange.ID, trackingBatchChange.ID},
			}
			bt.AssertChangeset(t, c2, trackedChangesetAssertions)

			// Now try to apply a new spec that wants to modify the formerly tracked changeset.
			batchSpec3 := bt.CreateBatchSpec(t, ctx, store, "tracking-batch-change", admin.ID, 0)

			spec3 := bt.CreateChangesetSpec(t, ctx, store, bt.TestSpecOpts{
				User:      admin.ID,
				Repo:      repos[0].ID,
				BatchSpec: batchSpec3.ID,
				HeadRef:   "refs/heads/repo-0-branch-0",
				Typ:       btypes.ChangesetSpecTypeBranch,
			})
			// Apply again. This should have flagged the association as detach
			// and it should not be closed, since the batch change is not the
			// owner.
			trackingBatchChange, cs := applyAndListChangesets(adminCtx, t, svc, batchSpec3.RandID, 2)

			trackedChangesetAssertions.Closing = false
			trackedChangesetAssertions.ReconcilerState = btypes.ReconcilerStateQueued
			trackedChangesetAssertions.DetachFrom = []int64{trackingBatchChange.ID}
			trackedChangesetAssertions.AttachedTo = []int64{ownerBatchChange.ID}
			bt.ReloadAndAssertChangeset(t, ctx, store, c2, trackedChangesetAssertions)

			// But we do want to have a new changeset record that is going to create a new changeset on the code host.
			bt.ReloadAndAssertChangeset(t, ctx, store, cs[1], bt.ChangesetAssertions{
				Repo:               spec3.BaseRepoID,
				CurrentSpec:        spec3.ID,
				OwnedByBatchChange: trackingBatchChange.ID,
				ReconcilerState:    btypes.ReconcilerStateQueued,
				PublicationState:   btypes.ChangesetPublicationStateUnpublished,
				DiffStat:           bt.TestChangsetSpecDiffStat,
				AttachedTo:         []int64{trackingBatchChange.ID},
			})
		})

		t.Run("batch change with changeset that is unpublished", func(t *testing.T) {
			bt.TruncateTables(t, db, "changeset_events", "changesets", "batch_changes", "batch_specs", "changeset_specs")
			batchSpec1 := bt.CreateBatchSpec(t, ctx, store, "unpublished-changesets", admin.ID, 0)

			bt.CreateChangesetSpec(t, ctx, store, bt.TestSpecOpts{
				User:      admin.ID,
				Repo:      repos[3].ID,
				BatchSpec: batchSpec1.ID,
				HeadRef:   "refs/heads/never-published",
				Typ:       btypes.ChangesetSpecTypeBranch,
			})

			// We apply the spec and expect 1 changeset
			applyAndListChangesets(adminCtx, t, svc, batchSpec1.RandID, 1)

			// But the changeset was not published yet.
			// And now we apply a new spec without any changesets.
			batchSpec2 := bt.CreateBatchSpec(t, ctx, store, "unpublished-changesets", admin.ID, 0)

			// That should close no changesets, but set the unpublished changesets to be detached when
			// the reconciler picks them up.
			applyAndListChangesets(adminCtx, t, svc, batchSpec2.RandID, 1)
		})

		t.Run("batch change with changeset that wasn't processed before reapply", func(t *testing.T) {
			bt.TruncateTables(t, db, "changeset_events", "changesets", "batch_changes", "batch_specs", "changeset_specs")
			batchSpec1 := bt.CreateBatchSpec(t, ctx, store, "queued-changesets", admin.ID, 0)

			specOpts := bt.TestSpecOpts{
				User:      admin.ID,
				Repo:      repos[3].ID,
				BatchSpec: batchSpec1.ID,
				Title:     "Spec1",
				HeadRef:   "refs/heads/queued",
				Typ:       btypes.ChangesetSpecTypeBranch,
				Published: true,
			}
			spec1 := bt.CreateChangesetSpec(t, ctx, store, specOpts)

			// We apply the spec and expect 1 changeset
			batchChange, changesets := applyAndListChangesets(adminCtx, t, svc, batchSpec1.RandID, 1)

			// And publish it.
			bt.SetChangesetPublished(t, ctx, store, changesets[0], "123-queued", "refs/heads/queued")

			bt.ReloadAndAssertChangeset(t, ctx, store, changesets[0], bt.ChangesetAssertions{
				ReconcilerState:    btypes.ReconcilerStateCompleted,
				PublicationState:   btypes.ChangesetPublicationStatePublished,
				ExternalBranch:     "refs/heads/queued",
				ExternalID:         "123-queued",
				ExternalState:      btypes.ChangesetExternalStateOpen,
				Repo:               repos[3].ID,
				CurrentSpec:        spec1.ID,
				OwnedByBatchChange: batchChange.ID,
				DiffStat:           bt.TestChangsetSpecDiffStat,
				AttachedTo:         []int64{batchChange.ID},
			})

			// Apply again so that an update to the changeset is pending.
			batchSpec2 := bt.CreateBatchSpec(t, ctx, store, "queued-changesets", admin.ID, 0)

			specOpts.BatchSpec = batchSpec2.ID
			specOpts.Title = "Spec2"
			spec2 := bt.CreateChangesetSpec(t, ctx, store, specOpts)

			// That should still want to publish the changeset
			_, changesets = applyAndListChangesets(adminCtx, t, svc, batchSpec2.RandID, 1)

			bt.ReloadAndAssertChangeset(t, ctx, store, changesets[0], bt.ChangesetAssertions{
				ReconcilerState:  btypes.ReconcilerStateQueued,
				PublicationState: btypes.ChangesetPublicationStatePublished,
				ExternalBranch:   "refs/heads/queued",
				ExternalID:       "123-queued",
				ExternalState:    btypes.ChangesetExternalStateOpen,
				Repo:             repos[3].ID,
				CurrentSpec:      spec2.ID,
				// Track the previous spec.
				PreviousSpec:       spec1.ID,
				OwnedByBatchChange: batchChange.ID,
				DiffStat:           bt.TestChangsetSpecDiffStat,
				AttachedTo:         []int64{batchChange.ID},
			})

			// Make sure the reconciler wants to update this changeset.
			plan, err := reconciler.DeterminePlan(
				// changesets[0].PreviousSpecID
				spec1,
				// changesets[0].CurrentSpecID
				spec2,
				nil,
				changesets[0],
			)
			if err != nil {
				t.Fatal(err)
			}
			if !plan.Ops.Equal(reconciler.Operations{btypes.ReconcilerOperationUpdate}) {
				t.Fatalf("Got invalid reconciler operations: %q", plan.Ops.String())
			}

			// And now we apply a new spec before the reconciler could process the changeset.
			batchSpec3 := bt.CreateBatchSpec(t, ctx, store, "queued-changesets", admin.ID, 0)

			// No change this time, just reapplying.
			specOpts.BatchSpec = batchSpec3.ID
			spec3 := bt.CreateChangesetSpec(t, ctx, store, specOpts)

			_, changesets = applyAndListChangesets(adminCtx, t, svc, batchSpec3.RandID, 1)

			bt.ReloadAndAssertChangeset(t, ctx, store, changesets[0], bt.ChangesetAssertions{
				ReconcilerState:  btypes.ReconcilerStateQueued,
				PublicationState: btypes.ChangesetPublicationStatePublished,
				ExternalBranch:   "refs/heads/queued",
				ExternalID:       "123-queued",
				ExternalState:    btypes.ChangesetExternalStateOpen,
				Repo:             repos[3].ID,
				CurrentSpec:      spec3.ID,
				// Still be pointing at the first spec, since the second was never applied.
				PreviousSpec:       spec1.ID,
				OwnedByBatchChange: batchChange.ID,
				DiffStat:           bt.TestChangsetSpecDiffStat,
				AttachedTo:         []int64{batchChange.ID},
			})

			// Make sure the reconciler would still update this changeset.
			plan, err = reconciler.DeterminePlan(
				// changesets[0].PreviousSpecID
				spec1,
				// changesets[0].CurrentSpecID
				spec3,
				nil,
				changesets[0],
			)
			if err != nil {
				t.Fatal(err)
			}
			if !plan.Ops.Equal(reconciler.Operations{btypes.ReconcilerOperationUpdate}) {
				t.Fatalf("Got invalid reconciler operations: %q", plan.Ops.String())
			}

			// Now test that it still updates when this update failed.
			bt.SetChangesetFailed(t, ctx, store, changesets[0])

			batchSpec4 := bt.CreateBatchSpec(t, ctx, store, "queued-changesets", admin.ID, 0)

			// No change this time, just reapplying.
			specOpts.BatchSpec = batchSpec4.ID
			spec4 := bt.CreateChangesetSpec(t, ctx, store, specOpts)

			_, changesets = applyAndListChangesets(adminCtx, t, svc, batchSpec4.RandID, 1)

			bt.ReloadAndAssertChangeset(t, ctx, store, changesets[0], bt.ChangesetAssertions{
				ReconcilerState:  btypes.ReconcilerStateQueued,
				PublicationState: btypes.ChangesetPublicationStatePublished,
				ExternalBranch:   "refs/heads/queued",
				ExternalID:       "123-queued",
				ExternalState:    btypes.ChangesetExternalStateOpen,
				Repo:             repos[3].ID,
				CurrentSpec:      spec4.ID,
				// Still be pointing at the first spec, since the second and third were never applied.
				PreviousSpec:       spec1.ID,
				OwnedByBatchChange: batchChange.ID,
				DiffStat:           bt.TestChangsetSpecDiffStat,
				AttachedTo:         []int64{batchChange.ID},
			})

			// Make sure the reconciler would still update this changeset.
			plan, err = reconciler.DeterminePlan(
				// changesets[0].PreviousSpecID
				spec1,
				// changesets[0].CurrentSpecID
				spec4,
				nil,
				changesets[0],
			)
			if err != nil {
				t.Fatal(err)
			}
			if !plan.Ops.Equal(reconciler.Operations{btypes.ReconcilerOperationUpdate}) {
				t.Fatalf("Got invalid reconciler operations: %q", plan.Ops.String())
			}
		})

		t.Run("missing repository permissions", func(t *testing.T) {
			bt.TruncateTables(t, db, "changeset_events", "changesets", "batch_changes", "batch_specs", "changeset_specs")
			bt.MockRepoPermissions(t, db, user.ID, repos[0].ID, repos[2].ID, repos[3].ID)

			// NOTE: We cannot use a context with an internal actor.
			batchSpec := bt.CreateBatchSpec(t, userCtx, store, "missing-permissions", user.ID, 0)

			bt.CreateChangesetSpec(t, userCtx, store, bt.TestSpecOpts{
				User:       user.ID,
				Repo:       repos[0].ID,
				BatchSpec:  batchSpec.ID,
				ExternalID: "1234",
				Typ:        btypes.ChangesetSpecTypeExisting,
			})

			bt.CreateChangesetSpec(t, userCtx, store, bt.TestSpecOpts{
				User:      user.ID,
				Repo:      repos[1].ID, // Not authorized to access this repository
				BatchSpec: batchSpec.ID,
				HeadRef:   "refs/heads/my-branch",
				Typ:       btypes.ChangesetSpecTypeBranch,
			})

			_, err := svc.ApplyBatchChange(userCtx, ApplyBatchChangeOpts{
				BatchSpecRandID: batchSpec.RandID,
			})
			if err == nil {
				t.Fatal("expected error, but got none")
			}
			var e *database.RepoNotFoundErr
			if !errors.As(err, &e) {
				t.Fatalf("expected RepoNotFoundErr but got: %s", err)
			}
			if e.ID != repos[1].ID {
				t.Fatalf("wrong repository ID in RepoNotFoundErr: %d", e.ID)
			}
		})

		t.Run("batch change with errored changeset", func(t *testing.T) {
			bt.TruncateTables(t, db, "changeset_events", "changesets", "batch_changes", "batch_specs", "changeset_specs")
			batchSpec1 := bt.CreateBatchSpec(t, ctx, store, "errored-changeset-batch-change", admin.ID, 0)

			spec1Opts := bt.TestSpecOpts{
				User:       admin.ID,
				Repo:       repos[0].ID,
				BatchSpec:  batchSpec1.ID,
				ExternalID: "1234",
				Typ:        btypes.ChangesetSpecTypeExisting,
				Published:  true,
			}
			bt.CreateChangesetSpec(t, ctx, store, spec1Opts)

			spec2Opts := bt.TestSpecOpts{
				User:      admin.ID,
				Repo:      repos[1].ID,
				BatchSpec: batchSpec1.ID,
				HeadRef:   "refs/heads/repo-1-branch-1",
				Typ:       btypes.ChangesetSpecTypeBranch,
				Published: true,
			}
			bt.CreateChangesetSpec(t, ctx, store, spec2Opts)

			_, oldChangesets := applyAndListChangesets(adminCtx, t, svc, batchSpec1.RandID, 2)

			// Set the changesets to look like they failed in the reconciler
			for _, c := range oldChangesets {
				bt.SetChangesetFailed(t, ctx, store, c)
			}

			// Now we create another batch spec with the same batch change name
			// and namespace.
			batchSpec2 := bt.CreateBatchSpec(t, ctx, store, "errored-changeset-batch-change", admin.ID, 0)
			spec1Opts.BatchSpec = batchSpec2.ID
			newSpec1 := bt.CreateChangesetSpec(t, ctx, store, spec1Opts)
			spec2Opts.BatchSpec = batchSpec2.ID
			newSpec2 := bt.CreateChangesetSpec(t, ctx, store, spec2Opts)

			batchChange, cs := applyAndListChangesets(adminCtx, t, svc, batchSpec2.RandID, 2)

			c1 := cs.Find(btypes.WithExternalID(newSpec1.ExternalID))
			bt.ReloadAndAssertChangeset(t, ctx, store, c1, bt.ChangesetAssertions{
				Repo:             spec1Opts.Repo,
				ExternalID:       "1234",
				PublicationState: btypes.ChangesetPublicationStateUnpublished,
				AttachedTo:       []int64{batchChange.ID},

				ReconcilerState: btypes.ReconcilerStateQueued,
				FailureMessage:  nil,
				NumFailures:     0,
			})

			c2 := cs.Find(btypes.WithCurrentSpecID(newSpec2.ID))
			bt.AssertChangeset(t, c2, bt.ChangesetAssertions{
				Repo:        newSpec2.BaseRepoID,
				CurrentSpec: newSpec2.ID,
				// An errored changeset doesn't get the specs rotated, to prevent https://github.com/sourcegraph/sourcegraph/issues/16041.
				PreviousSpec:       0,
				OwnedByBatchChange: batchChange.ID,
				PublicationState:   btypes.ChangesetPublicationStateUnpublished,
				DiffStat:           bt.TestChangsetSpecDiffStat,
				AttachedTo:         []int64{batchChange.ID},

				ReconcilerState: btypes.ReconcilerStateQueued,
				FailureMessage:  nil,
				NumFailures:     0,
			})

			// Make sure the reconciler would still publish this changeset.
			plan, err := reconciler.DeterminePlan(
				// c2.previousSpec is 0
				nil,
				// c2.currentSpec is newSpec2
				newSpec2,
				nil,
				c2,
			)
			if err != nil {
				t.Fatal(err)
			}
			if !plan.Ops.Equal(reconciler.Operations{btypes.ReconcilerOperationPush, btypes.ReconcilerOperationPublish}) {
				t.Fatalf("Got invalid reconciler operations: %q", plan.Ops.String())
			}
		})

		t.Run("closed and archived changeset not re-enqueued for close", func(t *testing.T) {
			bt.TruncateTables(t, db, "changeset_events", "changesets", "batch_changes", "batch_specs", "changeset_specs")
			batchSpec1 := bt.CreateBatchSpec(t, ctx, store, "archived-closed-changeset", admin.ID, 0)

			specOpts := bt.TestSpecOpts{
				User:      admin.ID,
				Repo:      repos[0].ID,
				BatchSpec: batchSpec1.ID,
				Typ:       btypes.ChangesetSpecTypeBranch,
				HeadRef:   "refs/heads/archived-closed",
			}
			spec1 := bt.CreateChangesetSpec(t, ctx, store, specOpts)

			// STEP 1: We apply the spec and expect 1 changeset.
			batchChange, changesets := applyAndListChangesets(adminCtx, t, svc, batchSpec1.RandID, 1)

			// Now we update the changeset so it looks like it's been published
			// on the code host.
			c := changesets[0]
			bt.SetChangesetPublished(t, ctx, store, c, "995544", specOpts.HeadRef)

			assertions := bt.ChangesetAssertions{
				Repo:               c.RepoID,
				CurrentSpec:        spec1.ID,
				ExternalID:         c.ExternalID,
				ExternalBranch:     c.ExternalBranch,
				ExternalState:      btypes.ChangesetExternalStateOpen,
				OwnedByBatchChange: batchChange.ID,
				ReconcilerState:    btypes.ReconcilerStateCompleted,
				PublicationState:   btypes.ChangesetPublicationStatePublished,
				DiffStat:           bt.TestChangsetSpecDiffStat,
				AttachedTo:         []int64{batchChange.ID},
			}
			c = bt.ReloadAndAssertChangeset(t, ctx, store, c, assertions)

			// STEP 2: Now we apply a new spec without any changesets, but expect the changeset-to-be-archived to
			// be left in the batch change (the reconciler would detach it, if the executor picked up the changeset).
			batchSpec2 := bt.CreateBatchSpec(t, ctx, store, "archived-closed-changeset", admin.ID, 0)
			applyAndListChangesets(adminCtx, t, svc, batchSpec2.RandID, 1)

			// Our previously published changeset should be marked as "to be
			// archived" and "to be closed"
			assertions.ArchiveIn = batchChange.ID
			assertions.AttachedTo = []int64{batchChange.ID}
			assertions.Closing = true
			assertions.ReconcilerState = btypes.ReconcilerStateQueued
			// And the previous spec is recorded, because the previous run finished with reconcilerState completed.
			assertions.PreviousSpec = spec1.ID
			c = bt.ReloadAndAssertChangeset(t, ctx, store, c, assertions)

			// Now we update the changeset to make it look closed and archived.
			bt.SetChangesetClosed(t, ctx, store, c)
			assertions.Closing = false
			assertions.ReconcilerState = btypes.ReconcilerStateCompleted
			assertions.ArchivedInOwnerBatchChange = true
			assertions.ArchiveIn = 0
			assertions.ExternalState = btypes.ChangesetExternalStateClosed
			c = bt.ReloadAndAssertChangeset(t, ctx, store, c, assertions)

			// STEP 3: We apply a new batch spec and expect that the archived changeset record is not re-enqueued.
			batchSpec3 := bt.CreateBatchSpec(t, ctx, store, "archived-closed-changeset", admin.ID, 0)

			// 1 changeset that's archived
			applyAndListChangesets(adminCtx, t, svc, batchSpec3.RandID, 1)

			// Assert that the changeset record is still archived and closed.
			bt.ReloadAndAssertChangeset(t, ctx, store, c, assertions)
		})

		t.Run("batch change with changeset that is archived and reattached", func(t *testing.T) {
			t.Run("changeset has been closed before re-attaching", func(t *testing.T) {
				bt.TruncateTables(t, db, "changeset_events", "changesets", "batch_changes", "batch_specs", "changeset_specs")
				batchSpec1 := bt.CreateBatchSpec(t, ctx, store, "detach-reattach-changeset", admin.ID, 0)

				specOpts := bt.TestSpecOpts{
					User:      admin.ID,
					Repo:      repos[0].ID,
					BatchSpec: batchSpec1.ID,
					HeadRef:   "refs/heads/archived-reattached",
					Typ:       btypes.ChangesetSpecTypeBranch,
				}
				spec1 := bt.CreateChangesetSpec(t, ctx, store, specOpts)

				// STEP 1: We apply the spec and expect 1 changeset.
				batchChange, changesets := applyAndListChangesets(adminCtx, t, svc, batchSpec1.RandID, 1)

				// Now we update the changeset so it looks like it's been published
				// on the code host.
				c := changesets[0]
				bt.SetChangesetPublished(t, ctx, store, c, "995533", specOpts.HeadRef)

				assertions := bt.ChangesetAssertions{
					Repo:               c.RepoID,
					CurrentSpec:        spec1.ID,
					ExternalID:         c.ExternalID,
					ExternalBranch:     c.ExternalBranch,
					ExternalState:      btypes.ChangesetExternalStateOpen,
					OwnedByBatchChange: batchChange.ID,
					ReconcilerState:    btypes.ReconcilerStateCompleted,
					PublicationState:   btypes.ChangesetPublicationStatePublished,
					DiffStat:           bt.TestChangsetSpecDiffStat,
					AttachedTo:         []int64{batchChange.ID},
				}
				bt.ReloadAndAssertChangeset(t, ctx, store, c, assertions)

				// STEP 2: Now we apply a new spec without any changesets.
				batchSpec2 := bt.CreateBatchSpec(t, ctx, store, "detach-reattach-changeset", admin.ID, 0)
				applyAndListChangesets(adminCtx, t, svc, batchSpec2.RandID, 1)

				// Our previously published changeset should be marked as "to
				// be archived" and "to be closed"
				assertions.Closing = true
				assertions.ArchiveIn = batchChange.ID
				assertions.AttachedTo = []int64{batchChange.ID}
				assertions.ReconcilerState = btypes.ReconcilerStateQueued
				// And the previous spec is recorded.
				assertions.PreviousSpec = spec1.ID
				c = bt.ReloadAndAssertChangeset(t, ctx, store, c, assertions)

				// Now we update the changeset to make it look closed.
				bt.SetChangesetClosed(t, ctx, store, c)
				assertions.Closing = false
				assertions.ArchiveIn = 0
				assertions.ReconcilerState = btypes.ReconcilerStateCompleted
				assertions.ExternalState = btypes.ChangesetExternalStateClosed
				bt.ReloadAndAssertChangeset(t, ctx, store, c, assertions)

				// STEP 3: We apply a new batch spec with a changeset spec that
				// matches the old changeset and expect _the same changeset_ to be
				// re-attached.
				batchSpec3 := bt.CreateBatchSpec(t, ctx, store, "detach-reattach-changeset", admin.ID, 0)

				specOpts.BatchSpec = batchSpec3.ID
				spec2 := bt.CreateChangesetSpec(t, ctx, store, specOpts)

				_, changesets = applyAndListChangesets(adminCtx, t, svc, batchSpec3.RandID, 1)

				attachedChangeset := changesets[0]
				if have, want := attachedChangeset.ID, c.ID; have != want {
					t.Fatalf("attached changeset has wrong ID. want=%d, have=%d", want, have)
				}

				// Assert that the changeset has been updated to point to the new spec
				assertions.CurrentSpec = spec2.ID
				// Assert that the previous spec is still spec 1
				assertions.PreviousSpec = spec1.ID
				assertions.ReconcilerState = btypes.ReconcilerStateQueued
				// Assert that it's not archived anymore:
				assertions.ArchiveIn = 0
				assertions.AttachedTo = []int64{batchChange.ID}
				bt.AssertChangeset(t, attachedChangeset, assertions)
			})

			t.Run("changeset has failed closing before re-attaching", func(t *testing.T) {
				bt.TruncateTables(t, db, "changeset_events", "changesets", "batch_changes", "batch_specs", "changeset_specs")
				batchSpec1 := bt.CreateBatchSpec(t, ctx, store, "detach-reattach-failed-changeset", admin.ID, 0)

				specOpts := bt.TestSpecOpts{
					User:      admin.ID,
					Repo:      repos[0].ID,
					BatchSpec: batchSpec1.ID,
					HeadRef:   "refs/heads/detached-reattach-failed",
					Typ:       btypes.ChangesetSpecTypeBranch,
				}
				spec1 := bt.CreateChangesetSpec(t, ctx, store, specOpts)

				// STEP 1: We apply the spec and expect 1 changeset.
				batchChange, changesets := applyAndListChangesets(adminCtx, t, svc, batchSpec1.RandID, 1)

				// Now we update the changeset so it looks like it's been published
				// on the code host.
				c := changesets[0]
				bt.SetChangesetPublished(t, ctx, store, c, "80022", specOpts.HeadRef)

				assertions := bt.ChangesetAssertions{
					Repo:               c.RepoID,
					CurrentSpec:        spec1.ID,
					ExternalID:         c.ExternalID,
					ExternalBranch:     c.ExternalBranch,
					ExternalState:      btypes.ChangesetExternalStateOpen,
					OwnedByBatchChange: batchChange.ID,
					ReconcilerState:    btypes.ReconcilerStateCompleted,
					PublicationState:   btypes.ChangesetPublicationStatePublished,
					DiffStat:           bt.TestChangsetSpecDiffStat,
					AttachedTo:         []int64{batchChange.ID},
				}
				bt.ReloadAndAssertChangeset(t, ctx, store, c, assertions)

				// STEP 2: Now we apply a new spec without any changesets.
				batchSpec2 := bt.CreateBatchSpec(t, ctx, store, "detach-reattach-failed-changeset", admin.ID, 0)
				applyAndListChangesets(adminCtx, t, svc, batchSpec2.RandID, 1)

				// Our previously published changeset should be marked as "to
				// be archived" and "to be closed"
				assertions.Closing = true
				assertions.ArchiveIn = batchChange.ID
				assertions.AttachedTo = []int64{batchChange.ID}
				assertions.ReconcilerState = btypes.ReconcilerStateQueued
				// And the previous spec is recorded.
				assertions.PreviousSpec = spec1.ID
				c = bt.ReloadAndAssertChangeset(t, ctx, store, c, assertions)

				if len(c.BatchChanges) != 1 {
					t.Fatal("Expected changeset to be still attached to batch change, but wasn't")
				}

				// Now we update the changeset to simulate that closing failed.
				bt.SetChangesetFailed(t, ctx, store, c)
				assertions.Closing = true
				assertions.ReconcilerState = btypes.ReconcilerStateFailed
				assertions.ExternalState = btypes.ChangesetExternalStateOpen

				// Side-effects of bt.setChangesetFailed.
				assertions.FailureMessage = c.FailureMessage
				assertions.NumFailures = c.NumFailures
				bt.ReloadAndAssertChangeset(t, ctx, store, c, assertions)

				// STEP 3: We apply a new batch spec with a changeset spec that
				// matches the old changeset and expect _the same changeset_ to be
				// re-attached.
				batchSpec3 := bt.CreateBatchSpec(t, ctx, store, "detach-reattach-failed-changeset", admin.ID, 0)

				specOpts.BatchSpec = batchSpec3.ID
				spec2 := bt.CreateChangesetSpec(t, ctx, store, specOpts)

				_, changesets = applyAndListChangesets(adminCtx, t, svc, batchSpec3.RandID, 1)

				attachedChangeset := changesets[0]
				if have, want := attachedChangeset.ID, c.ID; have != want {
					t.Fatalf("attached changeset has wrong ID. want=%d, have=%d", want, have)
				}

				// Assert that the changeset has been updated to point to the new spec
				assertions.CurrentSpec = spec2.ID
				// Assert that the previous spec is still spec 1
				assertions.PreviousSpec = spec1.ID
				assertions.ReconcilerState = btypes.ReconcilerStateQueued
				assertions.FailureMessage = nil
				assertions.NumFailures = 0
				assertions.DetachFrom = []int64{}
				assertions.AttachedTo = []int64{batchChange.ID}
				assertions.ArchiveIn = 0
				bt.AssertChangeset(t, attachedChangeset, assertions)
			})

			t.Run("changeset has not been closed before re-attaching", func(t *testing.T) {
				bt.TruncateTables(t, db, "changeset_events", "changesets", "batch_changes", "batch_specs", "changeset_specs")
				// The difference to the previous test: we DON'T update the
				// changeset to make it look closed. We want to make sure that
				// we also pick up enqueued-to-be-closed changesets.

				batchSpec1 := bt.CreateBatchSpec(t, ctx, store, "detach-reattach-changeset-2", admin.ID, 0)

				specOpts := bt.TestSpecOpts{
					User:      admin.ID,
					Repo:      repos[0].ID,
					BatchSpec: batchSpec1.ID,
					HeadRef:   "refs/heads/detached-reattached-2",
					Typ:       btypes.ChangesetSpecTypeBranch,
				}
				spec1 := bt.CreateChangesetSpec(t, ctx, store, specOpts)

				// STEP 1: We apply the spec and expect 1 changeset.
				batchChange, changesets := applyAndListChangesets(adminCtx, t, svc, batchSpec1.RandID, 1)

				c := changesets[0]
				bt.SetChangesetPublished(t, ctx, store, c, "449955", specOpts.HeadRef)

				assertions := bt.ChangesetAssertions{
					Repo:               c.RepoID,
					CurrentSpec:        spec1.ID,
					ExternalID:         c.ExternalID,
					ExternalBranch:     c.ExternalBranch,
					ExternalState:      btypes.ChangesetExternalStateOpen,
					OwnedByBatchChange: batchChange.ID,
					ReconcilerState:    btypes.ReconcilerStateCompleted,
					PublicationState:   btypes.ChangesetPublicationStatePublished,
					DiffStat:           bt.TestChangsetSpecDiffStat,
					AttachedTo:         []int64{batchChange.ID},
				}
				bt.ReloadAndAssertChangeset(t, ctx, store, c, assertions)

				// STEP 2: Now we apply a new spec without any changesets.
				batchChange2 := bt.CreateBatchSpec(t, ctx, store, "detach-reattach-changeset-2", admin.ID, 0)
				applyAndListChangesets(adminCtx, t, svc, batchChange2.RandID, 1)

				// Our previously published changeset should be marked as "to
				// be archived" and "to be closed"
				assertions.Closing = true
				assertions.ArchiveIn = batchChange.ID
				assertions.AttachedTo = []int64{batchChange.ID}
				assertions.ReconcilerState = btypes.ReconcilerStateQueued
				// And the previous spec is recorded.
				assertions.PreviousSpec = spec1.ID
				bt.ReloadAndAssertChangeset(t, ctx, store, c, assertions)

				// STEP 3: We apply a new batch spec with a changeset spec that
				// matches the old changeset and expect _the same changeset_ to be
				// re-attached.
				batchSpec3 := bt.CreateBatchSpec(t, ctx, store, "detach-reattach-changeset-2", admin.ID, 0)

				specOpts.BatchSpec = batchSpec3.ID
				spec2 := bt.CreateChangesetSpec(t, ctx, store, specOpts)

				_, changesets = applyAndListChangesets(adminCtx, t, svc, batchSpec3.RandID, 1)

				attachedChangeset := changesets[0]
				if have, want := attachedChangeset.ID, c.ID; have != want {
					t.Fatalf("attached changeset has wrong ID. want=%d, have=%d", want, have)
				}

				// Assert that the changeset has been updated to point to the new spec
				assertions.CurrentSpec = spec2.ID
				// Assert that the previous spec is still spec 1
				assertions.PreviousSpec = spec1.ID
				assertions.ReconcilerState = btypes.ReconcilerStateQueued
				assertions.DetachFrom = []int64{}
				assertions.AttachedTo = []int64{batchChange.ID}
				assertions.ArchiveIn = 0
				bt.AssertChangeset(t, attachedChangeset, assertions)
			})
		})

		t.Run("invalid changeset specs", func(t *testing.T) {
			bt.TruncateTables(t, db, "changeset_events", "changesets", "batch_changes", "batch_specs", "changeset_specs")
			batchSpec := bt.CreateBatchSpec(t, ctx, store, "batchchange-invalid-specs", admin.ID, 0)

			// Both specs here have the same HeadRef in the same repository
			bt.CreateChangesetSpec(t, ctx, store, bt.TestSpecOpts{
				User:      admin.ID,
				Repo:      repos[0].ID,
				BatchSpec: batchSpec.ID,
				HeadRef:   "refs/heads/my-branch",
				Typ:       btypes.ChangesetSpecTypeBranch,
			})

			bt.CreateChangesetSpec(t, ctx, store, bt.TestSpecOpts{
				User:      admin.ID,
				Repo:      repos[0].ID,
				BatchSpec: batchSpec.ID,
				HeadRef:   "refs/heads/my-branch",
				Typ:       btypes.ChangesetSpecTypeBranch,
			})

			_, err := svc.ApplyBatchChange(adminCtx, ApplyBatchChangeOpts{
				BatchSpecRandID: batchSpec.RandID,
			})
			if err == nil {
				t.Fatal("expected error, but got none")
			}

			if !strings.Contains(err.Error(), "Validating changeset specs resulted in an error") {
				t.Fatalf("wrong error: %s", err)
			}
		})
	})

	t.Run("applying to closed batch change", func(t *testing.T) {
		bt.TruncateTables(t, db, "changeset_events", "changesets", "batch_changes", "batch_specs", "changeset_specs")
		batchSpec := bt.CreateBatchSpec(t, ctx, store, "closed-batch-change", admin.ID, 0)
		batchChange := bt.CreateBatchChange(t, ctx, store, "closed-batch-change", admin.ID, batchSpec.ID)

		batchChange.ClosedAt = time.Now()
		if err := store.UpdateBatchChange(ctx, batchChange); err != nil {
			t.Fatalf("failed to update batch change: %s", err)
		}

		_, err := svc.ApplyBatchChange(adminCtx, ApplyBatchChangeOpts{
			BatchSpecRandID: batchSpec.RandID,
		})
		if err != ErrApplyClosedBatchChange {
			t.Fatalf("ApplyBatchChange returned unexpected error: %s", err)
		}
	})
}

func applyAndListChangesets(ctx context.Context, t *testing.T, svc *Service, batchSpecRandID string, wantChangesets int) (*btypes.BatchChange, btypes.Changesets) {
	t.Helper()

	batchChange, err := svc.ApplyBatchChange(ctx, ApplyBatchChangeOpts{
		BatchSpecRandID: batchSpecRandID,
	})
	if err != nil {
		t.Fatalf("failed to apply batch change: %s", err)
	}

	if batchChange.ID == 0 {
		t.Fatalf("batch change ID is zero")
	}

	changesets, _, err := svc.store.ListChangesets(ctx, store.ListChangesetsOpts{
		BatchChangeID:   batchChange.ID,
		IncludeArchived: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if have, want := len(changesets), wantChangesets; have != want {
		t.Fatalf("wrong number of changesets. want=%d, have=%d", want, have)
	}

	return batchChange, changesets
}
