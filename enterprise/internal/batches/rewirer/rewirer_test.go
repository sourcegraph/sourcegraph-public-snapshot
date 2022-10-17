package rewirer

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/global"
	bt "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestRewirer_Rewire(t *testing.T) {
	testBatchChangeID := int64(123)
	testChangesetSpecID := int64(512)
	testRepoID := api.RepoID(128)
	testRepo := &types.Repo{
		ID: testRepoID,
		ExternalRepo: api.ExternalRepoSpec{
			ServiceType: extsvc.TypeGitHub,
		},
	}
	unsupportedTestRepoID := api.RepoID(256)
	unsupportedTestRepo := &types.Repo{
		ID: unsupportedTestRepoID,
		ExternalRepo: api.ExternalRepoSpec{
			ServiceType: extsvc.TypeOther,
		},
	}
	testCases := []struct {
		name           string
		mappings       btypes.RewirerMappings
		wantChangesets []bt.ChangesetAssertions
		wantErr        error
	}{
		{
			name:           "empty mappings",
			mappings:       btypes.RewirerMappings{},
			wantChangesets: []bt.ChangesetAssertions{},
		},
		// NO CHANGESET SPEC
		{
			name: "no spec matching existing imported changeset",
			mappings: btypes.RewirerMappings{{
				Changeset: bt.BuildChangeset(bt.TestChangesetOpts{
					Repo:         testRepoID,
					BatchChanges: []btypes.BatchChangeAssoc{{BatchChangeID: testBatchChangeID}},

					// Imported changeset:
					OwnedByBatchChange: 0,
					CurrentSpec:        0,
				}),
				Repo: testRepo,
			}},
			wantChangesets: []bt.ChangesetAssertions{
				// No match, should be re-enqueued and detached from the batch change.
				assertResetReconcilerState(bt.ChangesetAssertions{
					Repo:       testRepoID,
					DetachFrom: []int64{testBatchChangeID},
				}),
			},
		},
		{
			name: "no spec matching existing unpublished branch changeset owned by this batch change",
			mappings: btypes.RewirerMappings{{
				Changeset: bt.BuildChangeset(bt.TestChangesetOpts{
					Repo:         testRepoID,
					BatchChanges: []btypes.BatchChangeAssoc{{BatchChangeID: testBatchChangeID}},

					// Owned unpublished branch changeset:
					PublicationState:   btypes.ChangesetPublicationStateUnpublished,
					OwnedByBatchChange: testBatchChangeID,
					CurrentSpec:        testChangesetSpecID,
				}),
				Repo: testRepo,
			}},
			wantChangesets: []bt.ChangesetAssertions{
				// No match, should be re-enqueued and detached from the batch change.
				assertResetReconcilerState(bt.ChangesetAssertions{
					PublicationState:   btypes.ChangesetPublicationStateUnpublished,
					OwnedByBatchChange: testBatchChangeID,
					CurrentSpec:        testChangesetSpecID,
					Repo:               testRepoID,
					DetachFrom:         []int64{testBatchChangeID},
				}),
			},
		},
		{
			name: "no spec matching existing published and open branch changeset owned by this batch change",
			mappings: btypes.RewirerMappings{{
				Changeset: bt.BuildChangeset(bt.TestChangesetOpts{
					Repo:         testRepoID,
					BatchChanges: []btypes.BatchChangeAssoc{{BatchChangeID: testBatchChangeID}},

					// Owned, published branch changeset:
					OwnedByBatchChange: testBatchChangeID,
					CurrentSpec:        testChangesetSpecID,
					PublicationState:   btypes.ChangesetPublicationStatePublished,
					ExternalState:      btypes.ChangesetExternalStateOpen,
					// Publication succeeded
					ReconcilerState: btypes.ReconcilerStateCompleted,
				}),
				Repo: testRepo,
			}},
			wantChangesets: []bt.ChangesetAssertions{
				// No match, should be re-enqueued and detached from the batch change.
				assertResetReconcilerState(bt.ChangesetAssertions{
					PublicationState:   btypes.ChangesetPublicationStatePublished,
					ExternalState:      btypes.ChangesetExternalStateOpen,
					OwnedByBatchChange: testBatchChangeID,
					CurrentSpec:        testChangesetSpecID,
					Repo:               testRepoID,
					// Current spec should have been made the previous spec.
					PreviousSpec: testChangesetSpecID,
					// The changeset should be closed on the code host.
					Closing: true,
					// And still attached to the batch change but archived
					ArchiveIn:  testBatchChangeID,
					AttachedTo: []int64{testBatchChangeID},
				}),
			},
		},
		{
			name: "no spec matching existing published and merged branch changeset owned by this batch change",
			mappings: btypes.RewirerMappings{{
				Changeset: bt.BuildChangeset(bt.TestChangesetOpts{
					Repo:         testRepoID,
					BatchChanges: []btypes.BatchChangeAssoc{{BatchChangeID: testBatchChangeID}},

					// Owned, published branch changeset:
					OwnedByBatchChange: testBatchChangeID,
					CurrentSpec:        testChangesetSpecID,
					PublicationState:   btypes.ChangesetPublicationStatePublished,
					ExternalState:      btypes.ChangesetExternalStateMerged,
					// Publication succeeded
					ReconcilerState: btypes.ReconcilerStateCompleted,
				}),
				Repo: testRepo,
			}},
			wantChangesets: []bt.ChangesetAssertions{
				// No match, should be re-enqueued and detached from the batch change.
				assertResetReconcilerState(bt.ChangesetAssertions{
					PublicationState:   btypes.ChangesetPublicationStatePublished,
					ExternalState:      btypes.ChangesetExternalStateMerged,
					OwnedByBatchChange: testBatchChangeID,
					CurrentSpec:        testChangesetSpecID,
					Repo:               testRepoID,
					// Current spec should have been made the previous spec.
					PreviousSpec: testChangesetSpecID,
					// The changeset should NOT be closed on the code host, since it's already merged
					Closing: false,
					// And still attached to the batch change but archived
					ArchiveIn:  testBatchChangeID,
					AttachedTo: []int64{testBatchChangeID},
				}),
			},
		},
		{
			name: "no spec matching existing published and closed branch changeset owned by this batch change",
			mappings: btypes.RewirerMappings{{
				Changeset: bt.BuildChangeset(bt.TestChangesetOpts{
					Repo:         testRepoID,
					BatchChanges: []btypes.BatchChangeAssoc{{BatchChangeID: testBatchChangeID}},

					// Owned, published branch changeset:
					OwnedByBatchChange: testBatchChangeID,
					CurrentSpec:        testChangesetSpecID,
					PublicationState:   btypes.ChangesetPublicationStatePublished,
					ExternalState:      btypes.ChangesetExternalStateClosed,
					// Publication succeeded
					ReconcilerState: btypes.ReconcilerStateCompleted,
				}),
				Repo: testRepo,
			}},
			wantChangesets: []bt.ChangesetAssertions{
				// No match, should be re-enqueued and detached from the batch change.
				assertResetReconcilerState(bt.ChangesetAssertions{
					PublicationState:   btypes.ChangesetPublicationStatePublished,
					ExternalState:      btypes.ChangesetExternalStateClosed,
					OwnedByBatchChange: testBatchChangeID,
					CurrentSpec:        testChangesetSpecID,
					Repo:               testRepoID,
					// Current spec should have been made the previous spec.
					PreviousSpec: testChangesetSpecID,
					// The changeset should NOT be closed on the code host, since it's already closed
					Closing: false,
					// And still attached to the batch change but archived
					ArchiveIn:  testBatchChangeID,
					AttachedTo: []int64{testBatchChangeID},
				}),
			},
		},
		{
			name: "no spec matching existing changeset, no repo perms",
			mappings: btypes.RewirerMappings{{
				Changeset: bt.BuildChangeset(bt.TestChangesetOpts{
					Repo:         0,
					BatchChanges: []btypes.BatchChangeAssoc{{BatchChangeID: testBatchChangeID}},
				}),
				// No access to repo.
				Repo: nil,
			}},
			// Nothing should be done.
			wantChangesets: []bt.ChangesetAssertions{},
		},
		// END NO CHANGESET SPEC
		// NO CHANGESET
		{
			name: "new importing spec",
			mappings: btypes.RewirerMappings{{
				ChangesetSpec: bt.BuildChangesetSpec(t, bt.TestSpecOpts{
					Repo: testRepoID,

					ExternalID: "123",
					Typ:        btypes.ChangesetSpecTypeExisting,
				}),
				Repo: testRepo,
			}},
			wantChangesets: []bt.ChangesetAssertions{assertResetReconcilerState(bt.ChangesetAssertions{
				Repo:       testRepoID,
				ExternalID: "123",
				// Imported changesets always start as unpublished and will be set to published once the import succeeded.
				PublicationState: btypes.ChangesetPublicationStateUnpublished,
				AttachedTo:       []int64{testBatchChangeID},
			})},
		},
		{
			name: "new branch spec",
			mappings: btypes.RewirerMappings{{
				ChangesetSpec: bt.BuildChangesetSpec(t, bt.TestSpecOpts{
					ID:   testChangesetSpecID,
					Repo: testRepoID,

					HeadRef: "refs/heads/test-branch",
					Typ:     btypes.ChangesetSpecTypeBranch,
				}),
				Repo: testRepo,
			}},
			wantChangesets: []bt.ChangesetAssertions{assertResetReconcilerState(bt.ChangesetAssertions{
				Repo:               testRepoID,
				PublicationState:   btypes.ChangesetPublicationStateUnpublished,
				AttachedTo:         []int64{testBatchChangeID},
				OwnedByBatchChange: testBatchChangeID,
				CurrentSpec:        testChangesetSpecID,
				// Diff stat is copied over from changeset spec
				DiffStat: bt.TestChangsetSpecDiffStat,
			})},
		},
		{
			name: "unsupported repo",
			mappings: btypes.RewirerMappings{{
				ChangesetSpec: bt.BuildChangesetSpec(t, bt.TestSpecOpts{
					Repo:       unsupportedTestRepoID,
					ExternalID: "123",
					Typ:        btypes.ChangesetSpecTypeExisting,
				}),
				RepoID: unsupportedTestRepoID,
				Repo:   unsupportedTestRepo,
			}},
			wantErr: &ErrRepoNotSupported{
				ServiceType: unsupportedTestRepo.ExternalRepo.ServiceType,
				RepoName:    string(unsupportedTestRepo.Name),
			},
		},
		{
			name: "inaccessible repo",
			mappings: btypes.RewirerMappings{{
				ChangesetSpec: bt.BuildChangesetSpec(t, bt.TestSpecOpts{
					Repo:       testRepoID,
					ExternalID: "123",
					Typ:        btypes.ChangesetSpecTypeExisting,
				}),
				RepoID: testRepoID,
				Repo:   nil,
			}},
			wantErr: &database.RepoNotFoundErr{ID: testRepoID},
		},
		// END NO CHANGESET
		// CHANGESET SPEC AND CHANGESET
		{
			name: "update importing spec: imported by other",
			mappings: btypes.RewirerMappings{{
				ChangesetSpec: bt.BuildChangesetSpec(t, bt.TestSpecOpts{
					Repo: testRepoID,

					ExternalID: "123",
					Typ:        btypes.ChangesetSpecTypeExisting,
				}),
				Changeset: bt.BuildChangeset(bt.TestChangesetOpts{
					Repo:       testRepoID,
					ExternalID: "123",
					// Already attached to another batch change
					BatchChanges: []btypes.BatchChangeAssoc{{BatchChangeID: testBatchChangeID + 1}},
				}),
				Repo: testRepo,
			}},
			wantChangesets: []bt.ChangesetAssertions{
				// Should not be reenqueued
				{
					Repo:       testRepoID,
					ExternalID: "123",
					// Now should be attached to both btypes.
					AttachedTo: []int64{testBatchChangeID + 1, testBatchChangeID},
				},
			},
		},
		{
			name: "update importing spec: failed before",
			mappings: btypes.RewirerMappings{{
				ChangesetSpec: bt.BuildChangesetSpec(t, bt.TestSpecOpts{
					Repo: testRepoID,

					ExternalID: "123",
					Typ:        btypes.ChangesetSpecTypeExisting,
				}),
				Changeset: bt.BuildChangeset(bt.TestChangesetOpts{
					Repo:       testRepoID,
					ExternalID: "123",
					// Already attached to another batch change
					BatchChanges:    []btypes.BatchChangeAssoc{{BatchChangeID: testBatchChangeID + 1}},
					ReconcilerState: btypes.ReconcilerStateFailed,
				}),
				Repo: testRepo,
			}},
			wantChangesets: []bt.ChangesetAssertions{assertResetReconcilerState(bt.ChangesetAssertions{
				Repo:       testRepoID,
				ExternalID: "123",
				// Now should be attached to both btypes.
				AttachedTo: []int64{testBatchChangeID + 1, testBatchChangeID},
			})},
		},
		{
			name: "update importing spec: created by other batch change",
			mappings: btypes.RewirerMappings{{
				ChangesetSpec: bt.BuildChangesetSpec(t, bt.TestSpecOpts{
					Repo: testRepoID,

					ExternalID: "123",
					Typ:        btypes.ChangesetSpecTypeExisting,
				}),
				Changeset: bt.BuildChangeset(bt.TestChangesetOpts{
					Repo:       testRepoID,
					ExternalID: "123",
					// Already attached to another batch change
					BatchChanges: []btypes.BatchChangeAssoc{{BatchChangeID: testBatchChangeID + 1}},
					// Other batch change created this changeset.
					OwnedByBatchChange: testBatchChangeID + 1,
				}),
				Repo: testRepo,
			}},
			wantChangesets: []bt.ChangesetAssertions{
				// Changeset owned by another batch change should not be retried.
				{
					Repo:               testRepoID,
					ExternalID:         "123",
					OwnedByBatchChange: testBatchChangeID + 1,
					// Now should be attached to both btypes.
					AttachedTo: []int64{testBatchChangeID + 1, testBatchChangeID},
				}},
		},
		{
			name: "update branch spec",
			mappings: btypes.RewirerMappings{{
				ChangesetSpec: bt.BuildChangesetSpec(t, bt.TestSpecOpts{
					ID:   testChangesetSpecID + 1,
					Repo: testRepoID,

					HeadRef: "refs/heads/test-branch",
					Typ:     btypes.ChangesetSpecTypeBranch,
				}),
				Changeset: bt.BuildChangeset(bt.TestChangesetOpts{
					Repo:               testRepoID,
					ExternalID:         "123",
					CurrentSpec:        testChangesetSpecID,
					BatchChanges:       []btypes.BatchChangeAssoc{{BatchChangeID: testBatchChangeID}},
					OwnedByBatchChange: testBatchChangeID,
					PublicationState:   btypes.ChangesetPublicationStatePublished,
					ReconcilerState:    btypes.ReconcilerStateCompleted,
				}),
				Repo: testRepo,
			}},
			wantChangesets: []bt.ChangesetAssertions{assertResetReconcilerState(bt.ChangesetAssertions{
				Repo:               testRepoID,
				ExternalID:         "123",
				OwnedByBatchChange: testBatchChangeID,
				AttachedTo:         []int64{testBatchChangeID},
				PublicationState:   btypes.ChangesetPublicationStatePublished,
				CurrentSpec:        testChangesetSpecID + 1,
				// The changeset was reconciled successfully before, so the previous spec should have been recorded.
				PreviousSpec: testChangesetSpecID,
				// Diff stat is copied over from changeset spec
				DiffStat: bt.TestChangsetSpecDiffStat,
			})},
		},
		{
			name: "update branch spec - failed before",
			mappings: btypes.RewirerMappings{{
				ChangesetSpec: bt.BuildChangesetSpec(t, bt.TestSpecOpts{
					ID:   testChangesetSpecID + 1,
					Repo: testRepoID,

					HeadRef: "refs/heads/test-branch",
					Typ:     btypes.ChangesetSpecTypeBranch,
				}),
				Changeset: bt.BuildChangeset(bt.TestChangesetOpts{
					Repo:               testRepoID,
					ExternalID:         "123",
					CurrentSpec:        testChangesetSpecID,
					BatchChanges:       []btypes.BatchChangeAssoc{{BatchChangeID: testBatchChangeID}},
					OwnedByBatchChange: testBatchChangeID,
					PublicationState:   btypes.ChangesetPublicationStatePublished,
					ReconcilerState:    btypes.ReconcilerStateFailed,
				}),
				Repo: testRepo,
			}},
			wantChangesets: []bt.ChangesetAssertions{assertResetReconcilerState(bt.ChangesetAssertions{
				Repo:               testRepoID,
				ExternalID:         "123",
				OwnedByBatchChange: testBatchChangeID,
				AttachedTo:         []int64{testBatchChangeID},
				PublicationState:   btypes.ChangesetPublicationStatePublished,
				CurrentSpec:        testChangesetSpecID + 1,
				// The changeset was not reconciled successfully before, so the previous spec should have remained unset.
				PreviousSpec: 0,
				// Diff stat is copied over from changeset spec
				DiffStat: bt.TestChangsetSpecDiffStat,
			})},
		},
		// END CHANGESET SPEC AND CHANGESET
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := New(tc.mappings, testBatchChangeID)

			changesets, err := r.Rewire()
			if err != nil && tc.wantErr == nil {
				t.Fatal(err)
			}
			if tc.wantErr != nil && err.Error() != tc.wantErr.Error() {
				t.Fatalf("incorrect error returned. want=%+v have=%+v", tc.wantErr, err)
			}
			if have, want := len(changesets), len(tc.wantChangesets); have != want {
				t.Fatalf("incorrect amount of changesets returned. want=%d have=%d", want, have)
			}
			for i, changeset := range changesets {
				bt.AssertChangeset(t, changeset, tc.wantChangesets[i])
			}
		})
	}
}

func assertResetReconcilerState(a bt.ChangesetAssertions) bt.ChangesetAssertions {
	a.ReconcilerState = global.DefaultReconcilerEnqueueState()
	a.NumFailures = 0
	a.NumResets = 0
	a.FailureMessage = nil
	a.SyncErrorMessage = nil
	return a
}
