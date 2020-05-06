package usagestats

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
)

// GetCampaignsUsageStatistics returns the current site's campaigns usage.
func GetCampaignsUsageStatistics(ctx context.Context) (*types.CampaignsUsageStatistics, error) {
	const campaignQuery = `
SELECT
    (SELECT COUNT(*) from campaigns) AS campaigns_count,
    COUNT(*) FILTER (WHERE changeset_jobs.changeset_id IS NOT NULL) AS created_changesets,
    COUNT(*) FILTER (WHERE changeset_jobs.changeset_id IS NOT NULL AND external_state = 'MERGED') AS created_changesets_merged,
    COUNT(*) FILTER (WHERE changeset_jobs.changeset_id IS NULL) AS manual_changesets,
    COUNT(*) FILTER (WHERE changeset_jobs.changeset_id IS NULL AND external_state = 'MERGED') AS manual_changesets_merged
    FROM changesets
    LEFT JOIN changeset_jobs ON changesets.id = changeset_jobs.changeset_id
    WHERE campaign_ids::text <> ''
    AND campaign_ids::text <> '{}';
`
	var (
		campaignsCount          int
		createdChangesets       int
		createdChangesetsMerged int
		manualChangesets        int
		manualChangesetsMerged  int
	)

	if err := dbconn.Global.QueryRowContext(ctx, campaignQuery).Scan(
		&campaignsCount,
		&createdChangesets,
		&createdChangesetsMerged,
		&manualChangesets,
		&manualChangesetsMerged,
	); err != nil {
		return nil, err
	}

	return &types.CampaignsUsageStatistics{
		CampaignsCount:               int32(campaignsCount),
		CreatedChangesetsCount:       int32(createdChangesets),
		CreatedChangesetsMergedCount: int32(createdChangesetsMerged),
		ManualChangesetsCount:        int32(manualChangesets),
		ManualChangesetsMergedCount:  int32(manualChangesetsMerged),
	}, nil
}
