package reconciler

import (
	"testing"

	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/testing"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
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
				PublicationState: campaigns.ChangesetPublicationStateUnpublished,
			},
			wantOperations: Operations{
				campaigns.ReconcilerOperationPush,
				campaigns.ReconcilerOperationPublish,
			},
		},
		{
			name:        "publish as draft",
			currentSpec: ct.TestSpecOpts{Published: "draft"},
			changeset: ct.TestChangesetOpts{
				PublicationState: campaigns.ChangesetPublicationStateUnpublished,
			},
			wantOperations: Operations{campaigns.ReconcilerOperationPush, campaigns.ReconcilerOperationPublishDraft},
		},
		{
			name:        "publish false",
			currentSpec: ct.TestSpecOpts{Published: false},
			changeset: ct.TestChangesetOpts{
				PublicationState: campaigns.ChangesetPublicationStateUnpublished,
			},
			wantOperations: Operations{},
		},
		{
			name:        "draft but unsupported",
			currentSpec: ct.TestSpecOpts{Published: "draft"},
			changeset: ct.TestChangesetOpts{
				ExternalServiceType: extsvc.TypeBitbucketServer,
				PublicationState:    campaigns.ChangesetPublicationStateUnpublished,
			},
			// should be a noop
			wantOperations: Operations{},
		},
		{
			name:         "draft to publish true",
			previousSpec: ct.TestSpecOpts{Published: "draft"},
			currentSpec:  ct.TestSpecOpts{Published: true},
			changeset: ct.TestChangesetOpts{
				PublicationState: campaigns.ChangesetPublicationStatePublished,
			},
			wantOperations: Operations{campaigns.ReconcilerOperationUndraft},
		},
		{
			name:         "draft to publish true on unpublished changeset",
			previousSpec: ct.TestSpecOpts{Published: "draft"},
			currentSpec:  ct.TestSpecOpts{Published: true},
			changeset: ct.TestChangesetOpts{
				PublicationState: campaigns.ChangesetPublicationStateUnpublished,
			},
			wantOperations: Operations{campaigns.ReconcilerOperationPush, campaigns.ReconcilerOperationPublish},
		},
		{
			name:         "title changed on published changeset",
			previousSpec: ct.TestSpecOpts{Published: true, Title: "Before"},
			currentSpec:  ct.TestSpecOpts{Published: true, Title: "After"},
			changeset: ct.TestChangesetOpts{
				PublicationState: campaigns.ChangesetPublicationStatePublished,
			},
			wantOperations: Operations{campaigns.ReconcilerOperationUpdate},
		},
		{
			name:         "commit diff changed on published changeset",
			previousSpec: ct.TestSpecOpts{Published: true, CommitDiff: "testDiff"},
			currentSpec:  ct.TestSpecOpts{Published: true, CommitDiff: "newTestDiff"},
			changeset: ct.TestChangesetOpts{
				PublicationState: campaigns.ChangesetPublicationStatePublished,
			},
			wantOperations: Operations{
				campaigns.ReconcilerOperationPush,
				campaigns.ReconcilerOperationSleep,
				campaigns.ReconcilerOperationSync,
			},
		},
		{
			name:         "commit message changed on published changeset",
			previousSpec: ct.TestSpecOpts{Published: true, CommitMessage: "old message"},
			currentSpec:  ct.TestSpecOpts{Published: true, CommitMessage: "new message"},
			changeset: ct.TestChangesetOpts{
				PublicationState: campaigns.ChangesetPublicationStatePublished,
			},
			wantOperations: Operations{
				campaigns.ReconcilerOperationPush,
				campaigns.ReconcilerOperationSleep,
				campaigns.ReconcilerOperationSync,
			},
		},
		{
			name:         "commit diff changed on merge changeset",
			previousSpec: ct.TestSpecOpts{Published: true, CommitDiff: "testDiff"},
			currentSpec:  ct.TestSpecOpts{Published: true, CommitDiff: "newTestDiff"},
			changeset: ct.TestChangesetOpts{
				PublicationState: campaigns.ChangesetPublicationStatePublished,
				ExternalState:    campaigns.ChangesetExternalStateMerged,
			},
			// should be a noop
			wantOperations: Operations{},
		},
		{
			name:         "changeset closed-and-detached will reopen",
			previousSpec: ct.TestSpecOpts{Published: true},
			currentSpec:  ct.TestSpecOpts{Published: true},
			changeset: ct.TestChangesetOpts{
				PublicationState: campaigns.ChangesetPublicationStatePublished,
				ExternalState:    campaigns.ChangesetExternalStateClosed,
				OwnedByCampaign:  1234,
				CampaignIDs:      []int64{1234},
			},
			wantOperations: Operations{
				campaigns.ReconcilerOperationReopen,
			},
		},
		{
			name:         "closing",
			previousSpec: ct.TestSpecOpts{Published: true},
			currentSpec:  ct.TestSpecOpts{Published: true},
			changeset: ct.TestChangesetOpts{
				PublicationState: campaigns.ChangesetPublicationStatePublished,
				ExternalState:    campaigns.ChangesetExternalStateOpen,
				OwnedByCampaign:  1234,
				CampaignIDs:      []int64{1234},
				// Important bit:
				Closing: true,
			},
			wantOperations: Operations{
				campaigns.ReconcilerOperationClose,
			},
		},
		{
			name:         "closing already-closed changeset",
			previousSpec: ct.TestSpecOpts{Published: true},
			currentSpec:  ct.TestSpecOpts{Published: true},
			changeset: ct.TestChangesetOpts{
				PublicationState: campaigns.ChangesetPublicationStatePublished,
				ExternalState:    campaigns.ChangesetExternalStateClosed,
				OwnedByCampaign:  1234,
				CampaignIDs:      []int64{1234},
				// Important bit:
				Closing: true,
			},
			wantOperations: Operations{
				// TODO: This should probably be a noop in the future
				campaigns.ReconcilerOperationClose,
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			var previousSpec *campaigns.ChangesetSpec
			if tc.previousSpec != (ct.TestSpecOpts{}) {
				previousSpec = ct.BuildChangesetSpec(t, tc.previousSpec)
			}

			currentSpec := ct.BuildChangesetSpec(t, tc.currentSpec)
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
