package service

import (
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// UiPublicationStates takes the publicationStates input from the
// applyBatchChange mutation, and applies the required validation and processing
// logic to calculate the eventual publication state for each changeset spec.
//
// External users must call Add() to add changeset spec random IDs to the
// struct, then process() must be called before publication states can be
// retrieved using get().
type UiPublicationStates struct {
	rand map[string]batches.PublishedValue
	id   map[int64]*btypes.ChangesetUiPublicationState
}

// Add adds a changeset spec random ID to the publication states.
func (ps *UiPublicationStates) Add(rand string, value batches.PublishedValue) error {
	if ps.rand == nil {
		ps.rand = map[string]batches.PublishedValue{rand: value}
		return nil
	}

	if _, ok := ps.rand[rand]; ok {
		return errors.Newf("duplicate changeset spec: %s", rand)
	}

	ps.rand[rand] = value
	return nil
}

func (ps *UiPublicationStates) get(id int64) *btypes.ChangesetUiPublicationState {
	if ps.id != nil {
		return ps.id[id]
	}
	return nil
}

// prepareAndValidate looks up the random changeset spec IDs, and ensures that
// the changeset specs are included in the current rewirer mappings and are
// eligible for a UI publication state.
func (ps *UiPublicationStates) prepareAndValidate(mappings btypes.RewirerMappings) error {
	// If there are no publication states -- which is the normal case -- there's
	// nothing to do here, and we can bail early.
	if len(ps.rand) == 0 {
		ps.id = nil
		return nil
	}

	// Fetch the changeset specs from the rewirer mappings and key them by
	// random ID, since that's the input we have.
	specs := map[string]*btypes.ChangesetSpec{}
	for _, mapping := range mappings {
		if mapping.ChangesetSpecID != 0 {
			specs[mapping.ChangesetSpec.RandID] = mapping.ChangesetSpec
		}
	}

	// Handle the specs. We'll drain ps.rand while we add entries to ps.id,
	// which means we can ensure that all the given changeset spec IDs mapped to
	// a changeset spec.
	var errs error
	ps.id = map[int64]*btypes.ChangesetUiPublicationState{}
	for rid, pv := range ps.rand {
		if spec, ok := specs[rid]; ok {
			if !spec.Published.Nil() {
				// If the changeset spec has an explicit published field, we cannot
				// override the publication state in the UI.
				errs = errors.Append(errs, errors.Newf("changeset spec %q has the published field set in its spec", rid))
			} else {
				ps.id[spec.ID] = btypes.ChangesetUiPublicationStateFromPublishedValue(pv)
				delete(ps.rand, spec.RandID)
			}
		}
	}

	// If there are any changeset spec IDs remaining, let's turn them into
	// errors.
	for rid := range ps.rand {
		errs = errors.Append(errs, errors.Newf("changeset spec %q not found", rid))
	}

	return errs
}
