package campaigns

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// ErrApplyClosedCampaign is returned by ApplyCampaign when the campaign
// matched by the campaign spec is already closed.
var ErrApplyClosedCampaign = errors.New("existing campaign matched by campaign spec is closed")

// ErrMatchingCampaignExists is returned by ApplyCampaign if a campaign matching the
// campaign spec already exists and FailIfExists was set.
var ErrMatchingCampaignExists = errors.New("a campaign matching the given campaign spec already exists")

type ApplyCampaignOpts struct {
	CampaignSpecRandID string
	EnsureCampaignID   int64

	// When FailIfCampaignExists is true, ApplyCampaign will fail if a Campaign
	// matching the given CampaignSpec already exists.
	FailIfCampaignExists bool
}

func (o ApplyCampaignOpts) String() string {
	return fmt.Sprintf(
		"CampaignSpec %s, EnsureCampaignID %d",
		o.CampaignSpecRandID,
		o.EnsureCampaignID,
	)
}

// ApplyCampaign creates the CampaignSpec.
func (s *Service) ApplyCampaign(ctx context.Context, opts ApplyCampaignOpts) (campaign *campaigns.Campaign, err error) {
	tr, ctx := trace.New(ctx, "Service.ApplyCampaign", opts.String())
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	campaignSpec, err := s.store.GetCampaignSpec(ctx, GetCampaignSpecOpts{
		RandID: opts.CampaignSpecRandID,
	})
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only site-admins or the creator of campaignSpec can apply
	// campaignSpec.
	if err := backend.CheckSiteAdminOrSameUser(ctx, campaignSpec.UserID); err != nil {
		return nil, err
	}

	campaign, previousSpecID, err := s.ReconcileCampaign(ctx, campaignSpec)
	if err != nil {
		return nil, err
	}

	if campaign.ID != 0 && opts.FailIfCampaignExists {
		return nil, ErrMatchingCampaignExists
	}

	if opts.EnsureCampaignID != 0 && campaign.ID != opts.EnsureCampaignID {
		return nil, ErrEnsureCampaignFailed
	}

	if campaign.Closed() {
		return nil, ErrApplyClosedCampaign
	}

	if previousSpecID == campaignSpec.ID {
		return campaign, nil
	}

	// Before we write to the database in a transaction, we cancel all
	// currently enqueued/errored-and-retryable changesets the campaign might
	// have.
	// We do this so we don't continue to possibly create changesets on the
	// codehost while we're applying a new campaign spec.
	// This is blocking, because the changeset rows currently being processed by the
	// reconciler are locked.
	if err := s.store.CancelQueuedCampaignChangesets(ctx, campaign.ID); err != nil {
		return campaign, nil
	}

	tx, err := s.store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	if campaign.ID == 0 {
		if err := tx.CreateCampaign(ctx, campaign); err != nil {
			return nil, err
		}
	} else {
		if err := tx.UpdateCampaign(ctx, campaign); err != nil {
			return nil, err
		}
	}

	rstore := repos.NewDBStore(tx.DB(), sql.TxOptions{})

	// Now we need to wire up the ChangesetSpecs of the new CampaignSpec
	// correctly with the Changesets so that the reconciler can create/update
	// them.

	// Load the mapping between ChangesetSpecs and existing Changesets in the target campaign.
	mappings, err := tx.GetRewirerMappings(ctx, GetRewirerMappingsOpts{
		CampaignSpecID: campaign.CampaignSpecID,
		CampaignID:     campaign.ID,
	})
	if err != nil {
		return nil, err
	}
	if err := mappings.Hydrate(ctx, tx); err != nil {
		return nil, err
	}

	// And execute the mapping.
	rewirer := NewChangesetRewirer(mappings, campaign, rstore)
	changesets, err := rewirer.Rewire(ctx)
	if err != nil {
		return nil, err
	}

	// Upsert all changesets.
	for _, changeset := range changesets {
		if err := tx.UpsertChangeset(ctx, changeset); err != nil {
			return nil, err
		}
	}

	return campaign, nil
}

func (s *Service) ReconcileCampaign(ctx context.Context, campaignSpec *campaigns.CampaignSpec) (campaign *campaigns.Campaign, previousSpecID int64, err error) {
	campaign, err = s.GetCampaignMatchingCampaignSpec(ctx, campaignSpec)
	if err != nil {
		return nil, 0, err
	}
	if campaign == nil {
		campaign = &campaigns.Campaign{}
	} else {
		previousSpecID = campaign.CampaignSpecID
	}
	// Populate the campaign with the values from the campaign spec.
	campaign.CampaignSpecID = campaignSpec.ID
	campaign.NamespaceOrgID = campaignSpec.NamespaceOrgID
	campaign.NamespaceUserID = campaignSpec.NamespaceUserID
	campaign.Name = campaignSpec.Spec.Name
	actor := actor.FromContext(ctx)
	if campaign.InitialApplierID == 0 {
		campaign.InitialApplierID = actor.UID
	}
	campaign.LastApplierID = actor.UID
	campaign.LastAppliedAt = s.clock()
	campaign.Description = campaignSpec.Spec.Description
	return campaign, previousSpecID, nil
}
