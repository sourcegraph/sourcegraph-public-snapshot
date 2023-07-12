package reconciler

import (
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/lib/batches"
)

// publicationStateCalculator calculates the desired publication state based on
// the published field of a changeset spec and the UI publication state of the
// changeset, if any.
type publicationStateCalculator struct {
	spec batches.PublishedValue
	ui   *btypes.ChangesetUiPublicationState
}

func calculatePublicationState(specPublished batches.PublishedValue, uiPublished *btypes.ChangesetUiPublicationState) *publicationStateCalculator {
	return &publicationStateCalculator{
		spec: specPublished,
		ui:   uiPublished,
	}
}

func (c *publicationStateCalculator) IsPublished() bool {
	return c.spec.True() || (c.spec.Nil() && c.ui != nil && *c.ui == btypes.ChangesetUiPublicationStatePublished)
}

func (c *publicationStateCalculator) IsDraft() bool {
	return c.spec.Draft() || (c.spec.Nil() && c.ui != nil && *c.ui == btypes.ChangesetUiPublicationStateDraft)
}

func (c *publicationStateCalculator) IsUnpublished() bool {
	return c.spec.False() || (c.spec.Nil() && (c.ui == nil || *c.ui == btypes.ChangesetUiPublicationStateUnpublished))
}
