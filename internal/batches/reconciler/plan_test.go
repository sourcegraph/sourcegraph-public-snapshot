package reconciler

import (
	"testing"

	bt "github.com/sourcegraph/sourcegraph/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestDetermineReconcilerPlan(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name           string
		previousSpec   *bt.TestSpecOpts
		currentSpec    *bt.TestSpecOpts
		changeset      bt.TestChangesetOpts
		wantOperations Operations
	}{
		{
			name:        "publish true",
			currentSpec: &bt.TestSpecOpts{Published: true},
			changeset: bt.TestChangesetOpts{
				PublicationState: btypes.ChangesetPublicationStateUnpublished,
			},
			wantOperations: Operations{
				btypes.ReconcilerOperationPush,
				btypes.ReconcilerOperationPublish,
			},
		},
		{
			name:        "publish as draft",
			currentSpec: &bt.TestSpecOpts{Published: "draft"},
			changeset: bt.TestChangesetOpts{
				PublicationState: btypes.ChangesetPublicationStateUnpublished,
			},
			wantOperations: Operations{btypes.ReconcilerOperationPush, btypes.ReconcilerOperationPublishDraft},
		},
		{
			name:        "publish false",
			currentSpec: &bt.TestSpecOpts{Published: false},
			changeset: bt.TestChangesetOpts{
				PublicationState: btypes.ChangesetPublicationStateUnpublished,
			},
			wantOperations: Operations{},
		},
		{
			name:        "draft but unsupported",
			currentSpec: &bt.TestSpecOpts{Published: "draft"},
			changeset: bt.TestChangesetOpts{
				ExternalServiceType: extsvc.TypeBitbucketServer,
				PublicationState:    btypes.ChangesetPublicationStateUnpublished,
			},
			// should be a noop
			wantOperations: Operations{},
		},
		{
			name:         "draft to publish true",
			previousSpec: &bt.TestSpecOpts{Published: "draft"},
			currentSpec:  &bt.TestSpecOpts{Published: true},
			changeset: bt.TestChangesetOpts{
				PublicationState: btypes.ChangesetPublicationStatePublished,
			},
			wantOperations: Operations{btypes.ReconcilerOperationUndraft},
		},
		{
			name:         "draft to publish true on unpublished changeset",
			previousSpec: &bt.TestSpecOpts{Published: "draft"},
			currentSpec:  &bt.TestSpecOpts{Published: true},
			changeset: bt.TestChangesetOpts{
				PublicationState: btypes.ChangesetPublicationStateUnpublished,
			},
			wantOperations: Operations{btypes.ReconcilerOperationPush, btypes.ReconcilerOperationPublish},
		},
		{
			name:        "publish nil; no ui state",
			currentSpec: &bt.TestSpecOpts{Published: nil},
			changeset: bt.TestChangesetOpts{
				PublicationState: btypes.ChangesetPublicationStateUnpublished,
			},
			wantOperations: Operations{},
		},
		{
			name:        "publish nil; unpublished ui state",
			currentSpec: &bt.TestSpecOpts{Published: nil},
			changeset: bt.TestChangesetOpts{
				PublicationState:   btypes.ChangesetPublicationStateUnpublished,
				UiPublicationState: pointers.Ptr(btypes.ChangesetUiPublicationStateUnpublished),
			},
			wantOperations: Operations{},
		},
		{
			name:        "publish nil; draft ui state",
			currentSpec: &bt.TestSpecOpts{Published: nil},
			changeset: bt.TestChangesetOpts{
				PublicationState:   btypes.ChangesetPublicationStateUnpublished,
				UiPublicationState: pointers.Ptr(btypes.ChangesetUiPublicationStateDraft),
			},
			wantOperations: Operations{btypes.ReconcilerOperationPush, btypes.ReconcilerOperationPublishDraft},
		},
		{
			name:        "publish nil; draft ui state; unsupported code host",
			currentSpec: &bt.TestSpecOpts{Published: nil},
			changeset: bt.TestChangesetOpts{
				ExternalServiceType: extsvc.TypeBitbucketServer,
				PublicationState:    btypes.ChangesetPublicationStateUnpublished,
				UiPublicationState:  pointers.Ptr(btypes.ChangesetUiPublicationStateDraft),
			},
			// Cannot draft on an unsupported code host, so this is a no-op.
			wantOperations: Operations{},
		},
		{
			name:        "publish nil; published ui state",
			currentSpec: &bt.TestSpecOpts{Published: nil},
			changeset: bt.TestChangesetOpts{
				PublicationState:   btypes.ChangesetPublicationStateUnpublished,
				UiPublicationState: pointers.Ptr(btypes.ChangesetUiPublicationStatePublished),
			},
			wantOperations: Operations{btypes.ReconcilerOperationPush, btypes.ReconcilerOperationPublish},
		},
		{
			name:         "publish draft to publish nil; ui state published",
			previousSpec: &bt.TestSpecOpts{Published: "draft"},
			currentSpec:  &bt.TestSpecOpts{Published: nil},
			changeset: bt.TestChangesetOpts{
				PublicationState:   btypes.ChangesetPublicationStatePublished,
				UiPublicationState: pointers.Ptr(btypes.ChangesetUiPublicationStatePublished),
			},
			wantOperations: Operations{btypes.ReconcilerOperationUndraft},
		},
		{
			name:         "publish draft to publish nil; ui state draft",
			previousSpec: &bt.TestSpecOpts{Published: "draft"},
			currentSpec:  &bt.TestSpecOpts{Published: nil},
			changeset: bt.TestChangesetOpts{
				PublicationState:   btypes.ChangesetPublicationStatePublished,
				UiPublicationState: pointers.Ptr(btypes.ChangesetUiPublicationStateDraft),
			},
			// No change to the actual state, so this is a no-op.
			wantOperations: Operations{},
		},
		{
			name:         "publish draft to publish nil; ui state unpublished",
			previousSpec: &bt.TestSpecOpts{Published: "draft"},
			currentSpec:  &bt.TestSpecOpts{Published: nil},
			changeset: bt.TestChangesetOpts{
				PublicationState:   btypes.ChangesetPublicationStatePublished,
				UiPublicationState: pointers.Ptr(btypes.ChangesetUiPublicationStateUnpublished),
			},
			// We can't unscramble an egg, nor can we unpublish a published
			// changeset, so this is a no-op.
			wantOperations: Operations{},
		},
		{
			name:         "publish draft to publish nil; ui state nil",
			previousSpec: &bt.TestSpecOpts{Published: "draft"},
			currentSpec:  &bt.TestSpecOpts{Published: nil},
			changeset: bt.TestChangesetOpts{
				PublicationState:   btypes.ChangesetPublicationStatePublished,
				UiPublicationState: nil,
			},
			// We can't unscramble an egg, nor can we unpublish a published
			// changeset, so this is a no-op.
			wantOperations: Operations{},
		},
		{
			name:         "published to publish nil; ui state nil",
			previousSpec: &bt.TestSpecOpts{Published: true},
			currentSpec:  &bt.TestSpecOpts{Published: nil},
			changeset: bt.TestChangesetOpts{
				PublicationState:   btypes.ChangesetPublicationStatePublished,
				UiPublicationState: nil,
			},
			// We can't unscramble an egg, nor can we unpublish a published
			// changeset, so this is a no-op.
			wantOperations: Operations{},
		},
		{
			name:        "ui published draft to ui published published",
			currentSpec: &bt.TestSpecOpts{Published: nil},
			changeset: bt.TestChangesetOpts{
				PublicationState:   btypes.ChangesetPublicationStatePublished,
				ExternalState:      btypes.ChangesetExternalStateDraft,
				UiPublicationState: &btypes.ChangesetUiPublicationStatePublished,
			},
			wantOperations: Operations{btypes.ReconcilerOperationUndraft},
		},
		{
			name:        "ui published published to ui published draft",
			currentSpec: &bt.TestSpecOpts{Published: nil},
			changeset: bt.TestChangesetOpts{
				PublicationState:   btypes.ChangesetPublicationStatePublished,
				ExternalState:      btypes.ChangesetExternalStateOpen,
				UiPublicationState: &btypes.ChangesetUiPublicationStateDraft,
			},
			// We expect a no-op here.
			wantOperations: Operations{},
		},
		{
			name:         "publishing read-only changeset",
			previousSpec: &bt.TestSpecOpts{Published: true},
			currentSpec:  &bt.TestSpecOpts{Published: true},
			changeset: bt.TestChangesetOpts{
				PublicationState:   btypes.ChangesetPublicationStatePublished,
				ExternalState:      btypes.ChangesetExternalStateReadOnly,
				UiPublicationState: &btypes.ChangesetUiPublicationStateDraft,
			},
			// We expect a no-op here.
			wantOperations: Operations{},
		},
		{
			name:         "title changed on published changeset",
			previousSpec: &bt.TestSpecOpts{Published: true, Title: "Before"},
			currentSpec:  &bt.TestSpecOpts{Published: true, Title: "After"},
			changeset: bt.TestChangesetOpts{
				PublicationState: btypes.ChangesetPublicationStatePublished,
			},
			wantOperations: Operations{btypes.ReconcilerOperationUpdate},
		},
		{
			name:         "title changed on read-only changeset",
			previousSpec: &bt.TestSpecOpts{Published: true, Title: "Before"},
			currentSpec:  &bt.TestSpecOpts{Published: true, Title: "After"},
			changeset: bt.TestChangesetOpts{
				PublicationState: btypes.ChangesetPublicationStatePublished,
				ExternalState:    btypes.ChangesetExternalStateReadOnly,
			},
			// We expect a no-op here.
			wantOperations: Operations{},
		},
		{
			name:         "commit diff changed on published changeset",
			previousSpec: &bt.TestSpecOpts{Published: true, CommitDiff: []byte("testDiff")},
			currentSpec:  &bt.TestSpecOpts{Published: true, CommitDiff: []byte("newTestDiff")},
			changeset: bt.TestChangesetOpts{
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
			previousSpec: &bt.TestSpecOpts{Published: true, CommitMessage: "old message"},
			currentSpec:  &bt.TestSpecOpts{Published: true, CommitMessage: "new message"},
			changeset: bt.TestChangesetOpts{
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
			previousSpec: &bt.TestSpecOpts{Published: true, CommitDiff: []byte("testDiff")},
			currentSpec:  &bt.TestSpecOpts{Published: true, CommitDiff: []byte("newTestDiff")},
			changeset: bt.TestChangesetOpts{
				PublicationState: btypes.ChangesetPublicationStatePublished,
				ExternalState:    btypes.ChangesetExternalStateMerged,
			},
			// should be a noop
			wantOperations: Operations{},
		},
		{
			name:         "changeset closed-and-detached will reopen",
			previousSpec: &bt.TestSpecOpts{Published: true},
			currentSpec:  &bt.TestSpecOpts{Published: true},
			changeset: bt.TestChangesetOpts{
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
			previousSpec: &bt.TestSpecOpts{Published: true},
			currentSpec:  &bt.TestSpecOpts{Published: true},
			changeset: bt.TestChangesetOpts{
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
			previousSpec: &bt.TestSpecOpts{Published: true},
			currentSpec:  &bt.TestSpecOpts{Published: true},
			changeset: bt.TestChangesetOpts{
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
			name:         "closing read-only changeset",
			previousSpec: &bt.TestSpecOpts{Published: true},
			currentSpec:  &bt.TestSpecOpts{Published: true},
			changeset: bt.TestChangesetOpts{
				PublicationState:   btypes.ChangesetPublicationStatePublished,
				ExternalState:      btypes.ChangesetExternalStateReadOnly,
				OwnedByBatchChange: 1234,
				BatchChanges:       []btypes.BatchChangeAssoc{{BatchChangeID: 1234}},
				// Important bit:
				Closing: true,
			},
			// should be a noop
			wantOperations: Operations{},
		},
		{
			name:         "detaching",
			previousSpec: &bt.TestSpecOpts{Published: true},
			currentSpec:  &bt.TestSpecOpts{Published: true},
			changeset: bt.TestChangesetOpts{
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
			previousSpec: &bt.TestSpecOpts{Published: true},
			currentSpec:  &bt.TestSpecOpts{Published: true},
			changeset: bt.TestChangesetOpts{
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
			currentSpec: &bt.TestSpecOpts{Published: true},
			changeset: bt.TestChangesetOpts{
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
			changeset: bt.TestChangesetOpts{
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
			changeset: bt.TestChangesetOpts{
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
			changeset: bt.TestChangesetOpts{
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
			changeset: bt.TestChangesetOpts{
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
			changeset: bt.TestChangesetOpts{
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
			if tc.previousSpec != nil {
				tc.previousSpec.Typ = btypes.ChangesetSpecTypeBranch
				previousSpec = bt.BuildChangesetSpec(t, *tc.previousSpec)
			}

			if tc.currentSpec != nil {
				tc.currentSpec.Typ = btypes.ChangesetSpecTypeBranch
				currentSpec = bt.BuildChangesetSpec(t, *tc.currentSpec)
			}

			cs := bt.BuildChangeset(tc.changeset)

			plan, err := DeterminePlan(previousSpec, currentSpec, nil, cs)
			if err != nil {
				t.Fatal(err)
			}
			if have, want := plan.Ops, tc.wantOperations; !have.Equal(want) {
				t.Fatalf("incorrect plan determined, want=%v have=%v", want, have)
			}
		})
	}
}
