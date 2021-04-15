package reconciler

import (
	"testing"

	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

func TestDetermineReconcilerPlan(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name           string
		previousSpec   ct.TestSpecOpts
		currentSpec    ct.TestSpecOpts
		changeset      ct.TestChangesetOpts
		wantOperations Operations
	}{
		{
			name:        "publish true",
			currentSpec: ct.TestSpecOpts{Published: true},
			changeset: ct.TestChangesetOpts{
				PublicationState: btypes.ChangesetPublicationStateUnpublished,
			},
			wantOperations: Operations{
				btypes.ReconcilerOperationPush,
				btypes.ReconcilerOperationPublish,
			},
		},
		{
			name:        "publish as draft",
			currentSpec: ct.TestSpecOpts{Published: "draft"},
			changeset: ct.TestChangesetOpts{
				PublicationState: btypes.ChangesetPublicationStateUnpublished,
			},
			wantOperations: Operations{btypes.ReconcilerOperationPush, btypes.ReconcilerOperationPublishDraft},
		},
		{
			name:        "publish false",
			currentSpec: ct.TestSpecOpts{Published: false},
			changeset: ct.TestChangesetOpts{
				PublicationState: btypes.ChangesetPublicationStateUnpublished,
			},
			wantOperations: Operations{},
		},
		{
			name:        "draft but unsupported",
			currentSpec: ct.TestSpecOpts{Published: "draft"},
			changeset: ct.TestChangesetOpts{
				ExternalServiceType: extsvc.TypeBitbucketServer,
				PublicationState:    btypes.ChangesetPublicationStateUnpublished,
			},
			// should be a noop
			wantOperations: Operations{},
		},
		{
			name:         "draft to publish true",
			previousSpec: ct.TestSpecOpts{Published: "draft"},
			currentSpec:  ct.TestSpecOpts{Published: true},
			changeset: ct.TestChangesetOpts{
				PublicationState: btypes.ChangesetPublicationStatePublished,
			},
			wantOperations: Operations{btypes.ReconcilerOperationUndraft},
		},
		{
			name:         "draft to publish true on unpublished changeset",
			previousSpec: ct.TestSpecOpts{Published: "draft"},
			currentSpec:  ct.TestSpecOpts{Published: true},
			changeset: ct.TestChangesetOpts{
				PublicationState: btypes.ChangesetPublicationStateUnpublished,
			},
			wantOperations: Operations{btypes.ReconcilerOperationPush, btypes.ReconcilerOperationPublish},
		},
		{
			name:         "title changed on published changeset",
			previousSpec: ct.TestSpecOpts{Published: true, Title: "Before"},
			currentSpec:  ct.TestSpecOpts{Published: true, Title: "After"},
			changeset: ct.TestChangesetOpts{
				PublicationState: btypes.ChangesetPublicationStatePublished,
			},
			wantOperations: Operations{btypes.ReconcilerOperationUpdate},
		},
		{
			name:         "commit diff changed on published changeset",
			previousSpec: ct.TestSpecOpts{Published: true, CommitDiff: "testDiff"},
			currentSpec:  ct.TestSpecOpts{Published: true, CommitDiff: "newTestDiff"},
			changeset: ct.TestChangesetOpts{
				PublicationState: btypes.ChangesetPublicationStatePublished,
			},
			wantOperations: Operations{
				btypes.ReconcilerOperationPush,
				btypes.ReconcilerOperationSleep,
				btypes.ReconcilerOperationSync,
			},
		},
		{
			name:         "commit message changed on published changeset",
			previousSpec: ct.TestSpecOpts{Published: true, CommitMessage: "old message"},
			currentSpec:  ct.TestSpecOpts{Published: true, CommitMessage: "new message"},
			changeset: ct.TestChangesetOpts{
				PublicationState: btypes.ChangesetPublicationStatePublished,
			},
			wantOperations: Operations{
				btypes.ReconcilerOperationPush,
				btypes.ReconcilerOperationSleep,
				btypes.ReconcilerOperationSync,
			},
		},
		{
			name:         "commit diff changed on merge changeset",
			previousSpec: ct.TestSpecOpts{Published: true, CommitDiff: "testDiff"},
			currentSpec:  ct.TestSpecOpts{Published: true, CommitDiff: "newTestDiff"},
			changeset: ct.TestChangesetOpts{
				PublicationState: btypes.ChangesetPublicationStatePublished,
				ExternalState:    btypes.ChangesetExternalStateMerged,
			},
			// should be a noop
			wantOperations: Operations{},
		},
		{
			name:         "changeset closed-and-detached will reopen",
			previousSpec: ct.TestSpecOpts{Published: true},
			currentSpec:  ct.TestSpecOpts{Published: true},
			changeset: ct.TestChangesetOpts{
				PublicationState:   btypes.ChangesetPublicationStatePublished,
				ExternalState:      btypes.ChangesetExternalStateClosed,
				OwnedByBatchChange: 1234,
				BatchChanges:       []btypes.BatchChangeAssoc{{BatchChangeID: 1234}},
			},
			wantOperations: Operations{
				btypes.ReconcilerOperationReopen,
			},
		},
		{
			name:         "closing",
			previousSpec: ct.TestSpecOpts{Published: true},
			currentSpec:  ct.TestSpecOpts{Published: true},
			changeset: ct.TestChangesetOpts{
				PublicationState:   btypes.ChangesetPublicationStatePublished,
				ExternalState:      btypes.ChangesetExternalStateOpen,
				OwnedByBatchChange: 1234,
				BatchChanges:       []btypes.BatchChangeAssoc{{BatchChangeID: 1234}},
				// Important bit:
				Closing: true,
			},
			wantOperations: Operations{
				btypes.ReconcilerOperationClose,
			},
		},
		{
			name:         "closing already-closed changeset",
			previousSpec: ct.TestSpecOpts{Published: true},
			currentSpec:  ct.TestSpecOpts{Published: true},
			changeset: ct.TestChangesetOpts{
				PublicationState:   btypes.ChangesetPublicationStatePublished,
				ExternalState:      btypes.ChangesetExternalStateClosed,
				OwnedByBatchChange: 1234,
				BatchChanges:       []btypes.BatchChangeAssoc{{BatchChangeID: 1234}},
				// Important bit:
				Closing: true,
			},
			wantOperations: Operations{
				// TODO: This should probably be a noop in the future
				btypes.ReconcilerOperationClose,
			},
		},
		{
			name:         "detaching",
			previousSpec: ct.TestSpecOpts{Published: true},
			currentSpec:  ct.TestSpecOpts{Published: true},
			changeset: ct.TestChangesetOpts{
				PublicationState:   btypes.ChangesetPublicationStatePublished,
				ExternalState:      btypes.ChangesetExternalStateOpen,
				OwnedByBatchChange: 1234,
				BatchChanges:       []btypes.BatchChangeAssoc{{BatchChangeID: 1234, Detach: true}},
			},
			wantOperations: Operations{
				btypes.ReconcilerOperationDetach,
			},
		},
		{
			name:         "detaching already-detached changeset",
			previousSpec: ct.TestSpecOpts{Published: true},
			currentSpec:  ct.TestSpecOpts{Published: true},
			changeset: ct.TestChangesetOpts{
				PublicationState:   btypes.ChangesetPublicationStatePublished,
				ExternalState:      btypes.ChangesetExternalStateClosed,
				OwnedByBatchChange: 1234,
				BatchChanges:       []btypes.BatchChangeAssoc{},
			},
			wantOperations: Operations{
				// Expect no operations.
			},
		},
		{
			name:        "detaching a failed publish changeset",
			currentSpec: ct.TestSpecOpts{Published: true},
			changeset: ct.TestChangesetOpts{
				PublicationState:   btypes.ChangesetPublicationStateUnpublished,
				ReconcilerState:    btypes.ReconcilerStateFailed,
				OwnedByBatchChange: 1234,
				BatchChanges:       []btypes.BatchChangeAssoc{{BatchChangeID: 1234, Detach: true}},
			},
			wantOperations: Operations{
				btypes.ReconcilerOperationDetach,
			},
		},
		{
			name: "detaching a failed importing changeset",
			changeset: ct.TestChangesetOpts{
				ExternalID:       "123",
				PublicationState: btypes.ChangesetPublicationStateUnpublished,
				ReconcilerState:  btypes.ReconcilerStateFailed,
				BatchChanges:     []btypes.BatchChangeAssoc{{BatchChangeID: 1234, Detach: true}},
			},
			wantOperations: Operations{
				btypes.ReconcilerOperationDetach,
			},
		},
		{
			name: "archiving",
			changeset: ct.TestChangesetOpts{
				PublicationState:   btypes.ChangesetPublicationStatePublished,
				ExternalState:      btypes.ChangesetExternalStateOpen,
				OwnedByBatchChange: 1234,
				BatchChanges:       []btypes.BatchChangeAssoc{{BatchChangeID: 1234, Archive: true}},
			},
			wantOperations: Operations{
				btypes.ReconcilerOperationArchive,
			},
		},
		{
			name: "archiving already-archived changeset",
			changeset: ct.TestChangesetOpts{
				PublicationState:   btypes.ChangesetPublicationStatePublished,
				ExternalState:      btypes.ChangesetExternalStateClosed,
				OwnedByBatchChange: 1234,
				BatchChanges: []btypes.BatchChangeAssoc{{
					BatchChangeID: 1234, Archive: true, IsArchived: true,
				}},
			},
			wantOperations: Operations{
				// Expect no operations.
			},
		},
		{
			name: "import changeset",
			changeset: ct.TestChangesetOpts{
				ExternalID:       "123",
				PublicationState: btypes.ChangesetPublicationStateUnpublished,
				ReconcilerState:  btypes.ReconcilerStateQueued,
				BatchChanges:     []btypes.BatchChangeAssoc{{BatchChangeID: 1234}},
			},
			wantOperations: Operations{
				btypes.ReconcilerOperationImport,
			},
		},
		{
			name: "detaching an importing changeset but remains imported by another",
			changeset: ct.TestChangesetOpts{
				ExternalID:       "123",
				PublicationState: btypes.ChangesetPublicationStateUnpublished,
				ReconcilerState:  btypes.ReconcilerStateQueued,
				BatchChanges:     []btypes.BatchChangeAssoc{{BatchChangeID: 1234, Detach: true}, {BatchChangeID: 2345}},
			},
			wantOperations: Operations{
				btypes.ReconcilerOperationDetach,
				btypes.ReconcilerOperationImport,
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			var previousSpec, currentSpec *btypes.ChangesetSpec
			if tc.previousSpec != (ct.TestSpecOpts{}) {
				previousSpec = ct.BuildChangesetSpec(t, tc.previousSpec)
			}

			if tc.currentSpec != (ct.TestSpecOpts{}) {
				currentSpec = ct.BuildChangesetSpec(t, tc.currentSpec)
			}

			cs := ct.BuildChangeset(tc.changeset)

			plan, err := DeterminePlan(previousSpec, currentSpec, cs)
			if err != nil {
				t.Fatal(err)
			}
			if have, want := plan.Ops, tc.wantOperations; !have.Equal(want) {
				t.Fatalf("incorrect plan determined, want=%v have=%v", want, have)
			}
		})
	}
}
