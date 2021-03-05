package service

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/rewirer"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/batches"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// ErrApplyClosedBatchChange is returned by ApplyBatchChange when the campaign
// matched by the batch spec is already closed.
var ErrApplyClosedBatchChange = errors.New("existing batch change matched by batch spec is closed")

// ErrMatchingBatchChangeExists is returned by ApplyCampaign if a campaign matching the
// campaign spec already exists and FailIfExists was set.
var ErrMatchingBatchChangeExists = errors.New("a batch change matching the given batch spec already exists")

// TODO(campaigns-deprecation): this needs to be renamed, but cast to
// "EnsureCampaignFailed" if the applycampaign mutation was used.
//
// ErrEnsureBatchChangeFailed is returned by AppplyBatchChange when a
// ensureBatchChangeID is provided but a batch change with the name specified the
// batchSpec exists in the given namespace but has a different ID.
var ErrEnsureBatchChangeFailed = errors.New("a batch change in the given namespace and with the given name exists but does not match the given ID")

type ApplyBatchChangeOpts struct {
	BatchSpecRandID     string
	EnsureBatchChangeID int64

	// When FailIfBatchChangeExists is true, ApplyCampaign will fail if a Campaign
	// matching the given CampaignSpec already exists.
	FailIfBatchChangeExists bool
}

func (o ApplyBatchChangeOpts) String() string {
	return fmt.Sprintf(
		"BatchSpec %s, EnsureBatchChangeID %d",
		o.BatchSpecRandID,
		o.EnsureBatchChangeID,
	)
}

// ApplyBatchChange creates the BatchChange.
func (s *Service) ApplyBatchChange(ctx context.Context, opts ApplyBatchChangeOpts) (batchChange *batches.BatchChange, err error) {
	tr, ctx := trace.New(ctx, "Service.ApplyBatchChange", opts.String())
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	campaignSpec, err := s.store.GetBatchSpec(ctx, store.GetBatchSpecOpts{
		RandID: opts.BatchSpecRandID,
	})
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only site-admins or the creator of campaignSpec can apply
	// campaignSpec.
	if err := backend.CheckSiteAdminOrSameUser(ctx, campaignSpec.UserID); err != nil {
		return nil, err
	}

	batchChange, previousSpecID, err := s.ReconcileBatchChange(ctx, campaignSpec)
	if err != nil {
		return nil, err
	}

	if batchChange.ID != 0 && opts.FailIfBatchChangeExists {
		return nil, ErrMatchingBatchChangeExists
	}

	if opts.EnsureBatchChangeID != 0 && batchChange.ID != opts.EnsureBatchChangeID {
		return nil, ErrEnsureBatchChangeFailed
	}

	if batchChange.Closed() {
		return nil, ErrApplyClosedBatchChange
	}

	if previousSpecID == campaignSpec.ID {
		return batchChange, nil
	}

	// Before we write to the database in a transaction, we cancel all
	// currently enqueued/errored-and-retryable changesets the campaign might
	// have.
	// We do this so we don't continue to possibly create changesets on the
	// codehost while we're applying a new campaign spec.
	// This is blocking, because the changeset rows currently being processed by the
	// reconciler are locked.
	if err := s.store.CancelQueuedBatchChangeChangesets(ctx, batchChange.ID); err != nil {
		return batchChange, nil
	}

	tx, err := s.store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	if batchChange.ID == 0 {
		if err := tx.CreateBatchChange(ctx, batchChange); err != nil {
			return nil, err
		}
	} else {
		if err := tx.UpdateBatchChange(ctx, batchChange); err != nil {
			return nil, err
		}
	}

	// Now we need to wire up the ChangesetSpecs of the new CampaignSpec
	// correctly with the Changesets so that the reconciler can create/update
	// them.

	// Load the mapping between ChangesetSpecs and existing Changesets in the target campaign.
	mappings, err := tx.GetRewirerMappings(ctx, store.GetRewirerMappingsOpts{
		CampaignSpecID: batchChange.CampaignSpecID,
		CampaignID:     batchChange.ID,
	})
	if err != nil {
		return nil, err
	}
	if err := mappings.Hydrate(ctx, tx); err != nil {
		return nil, err
	}

	// And execute the mapping.
	changesets, err := rewirer.New(mappings, batchChange.ID).Rewire()
	if err != nil {
		return nil, err
	}

	// Upsert all changesets.
	for _, changeset := range changesets {
		if err := tx.UpsertChangeset(ctx, changeset); err != nil {
			return nil, err
		}
	}

	return batchChange, nil
}

func (s *Service) ReconcileBatchChange(ctx context.Context, campaignSpec *batches.BatchSpec) (campaign *batches.BatchChange, previousSpecID int64, err error) {
	campaign, err = s.GetBatchChangeMatchingBatchSpec(ctx, campaignSpec)
	if err != nil {
		return nil, 0, err
	}
	if campaign == nil {
		campaign = &batches.BatchChange{}
	} else {
		previousSpecID = campaign.CampaignSpecID
	}
	// Populate the campaign with the values from the campaign spec.
	campaign.CampaignSpecID = campaignSpec.ID
	campaign.NamespaceOrgID = campaignSpec.NamespaceOrgID
	campaign.NamespaceUserID = campaignSpec.NamespaceUserID
	campaign.Name = campaignSpec.Spec.Name
	a := actor.FromContext(ctx)
	if campaign.InitialApplierID == 0 {
		campaign.InitialApplierID = a.UID
	}
	campaign.LastApplierID = a.UID
	campaign.LastAppliedAt = s.clock()
	campaign.Description = campaignSpec.Spec.Description
	return campaign, previousSpecID, nil
}
