package service

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/hashicorp/go-multierror"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/lib/batches"
)

// UiPublicationStates takes the publicationStates input from the
// applyBatchChange mutation, and applies the required validation and processing
// logic.
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

type ListChangesetSpeccer interface {
	ListChangesetSpecs(context.Context, store.ListChangesetSpecsOpts) (btypes.ChangesetSpecs, int64, error)
}

var _ ListChangesetSpeccer = &store.Store{}

// prepareAndValidate looks up the random changeset spec IDs, and ensures that the
// changeset specs are attached to the batch spec and are eligible for a UI
// publication state.
func (ps *UiPublicationStates) prepareAndValidate(ctx context.Context, s ListChangesetSpeccer, batchSpecID int64) error {
	// If there are no publication states -- which is the normal case -- there's
	// nothing to do here, and we can bail early.
	if len(ps.rand) == 0 {
		ps.id = nil
		return nil
	}

	// Fetch the changeset specs, being careful to only look at the current
	// batch spec.
	randIDs := make([]string, 0, len(ps.rand))
	for id := range ps.rand {
		randIDs = append(randIDs, id)
	}

	specs, _, err := s.ListChangesetSpecs(ctx, store.ListChangesetSpecsOpts{
		BatchSpecID: batchSpecID,
		RandIDs:     randIDs,
	})
	if err != nil {
		return err
	}

	// Handle the specs. We'll drain ps.rand while we add entries to ps.id,
	// which means we can ensure that all the given changeset spec IDs mapped to
	// a changeset spec.
	ps.id = map[int64]*btypes.ChangesetUiPublicationState{}
	var errs *multierror.Error
	for _, spec := range specs {
		if !spec.Spec.Published.Nil() {
			// If the changeset spec has an explicit published field, we cannot
			// override the publication state in the UI.
			errs = multierror.Append(errs, errors.Newf("changeset spec %q has the published field set in its spec", spec.RandID))
		} else {
			ps.id[spec.ID] = btypes.ChangesetUiPublicationStateFromPublishedValue(ps.rand[spec.RandID])
			delete(ps.rand, spec.RandID)
		}
	}

	// If there are any changeset spec IDs remaining, let's turn them into
	// errors.
	for id := range ps.rand {
		errs = multierror.Append(errs, errors.Newf("changeset spec %q not found", id))
	}

	return errs.ErrorOrNil()
}
