package campaigns

// This file contains methods that exist solely to migrate campaigns and
// changesets lingering from before specs were added in Sourcegraph 3.19 into
// the new world.
//
// It should be removed in or after Sourcegraph 3.21.

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
)

func (s *Store) deletePreSpecCampaign(ctx context.Context, id int64) error {
	return s.Store.Exec(ctx, sqlf.Sprintf(
		"DELETE FROM campaigns_old WHERE id = %s",
		id,
	))
}

// preSpecCampaignColumns are used by the campaign related Store methods to
// insert, update and query old campaigns.
var preSpecCampaignColumns = []*sqlf.Query{
	sqlf.Sprintf("campaigns_old.id"),
	sqlf.Sprintf("campaigns_old.name"),
	sqlf.Sprintf("campaigns_old.description"),
	sqlf.Sprintf("campaigns_old.initial_applier_id"),
	sqlf.Sprintf("campaigns_old.last_applier_id"),
	sqlf.Sprintf("campaigns_old.last_applied_at"),
	sqlf.Sprintf("campaigns_old.namespace_user_id"),
	sqlf.Sprintf("campaigns_old.namespace_org_id"),
	sqlf.Sprintf("campaigns_old.created_at"),
	sqlf.Sprintf("campaigns_old.updated_at"),
	sqlf.Sprintf("campaigns_old.changeset_ids"),
	sqlf.Sprintf("campaigns_old.closed_at"),
	sqlf.Sprintf("campaigns_old.campaign_spec_id"),
}

func (s *Store) listPreSpecCampaigns(ctx context.Context) ([]*campaigns.Campaign, error) {
	cs := []*campaigns.Campaign{}
	q := sqlf.Sprintf(
		listPreSpecCampaignsFmtstr,
		sqlf.Join(preSpecCampaignColumns, ", "),
	)
	if err := s.query(ctx, q, func(sc scanner) error {
		var c campaigns.Campaign
		if err := scanPreSpecCampaign(&c, sc); err != nil {
			return err
		}
		cs = append(cs, &c)
		return nil
	}); err != nil {
		return nil, err
	}

	return cs, nil
}

const listPreSpecCampaignsFmtstr = `
-- source: enterprise/internal/campaigns/spec_migration_campaigns.go:listPreSpecCampaigns
SELECT %s FROM campaigns_old
`

func scanPreSpecCampaign(c *campaigns.Campaign, s scanner) error {
	return s.Scan(
		&c.ID,
		&c.Name,
		&dbutil.NullString{S: &c.Description},
		&c.InitialApplierID,
		&dbutil.NullInt32{N: &c.LastApplierID},
		&dbutil.NullTime{Time: &c.LastAppliedAt},
		&dbutil.NullInt32{N: &c.NamespaceUserID},
		&dbutil.NullInt32{N: &c.NamespaceOrgID},
		&c.CreatedAt,
		&c.UpdatedAt,
		&dbutil.JSONInt64Set{Set: &c.ChangesetIDs},
		&dbutil.NullTime{Time: &c.ClosedAt},
		&dbutil.NullInt64{N: &c.CampaignSpecID},
	)
}

func (s *Store) updateCampaignID(ctx context.Context, c *campaigns.Campaign, id int64) error {
	q := sqlf.Sprintf(
		updateCampaignIDFmtstr,
		id,
		c.ID,
		sqlf.Join(campaignColumns, ", "),
	)

	return s.query(ctx, q, func(sc scanner) error {
		return scanCampaign(c, sc)
	})
}

const updateCampaignIDFmtstr = `
-- source: enterprise/internal/campaigns/spec_migration_campaigns.go:updateCampaignID
UPDATE
	campaigns
SET
	id = %s
WHERE
	id = %s
RETURNING
	%s
`
