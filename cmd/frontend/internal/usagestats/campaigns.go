package usagestats

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
)

// GetCampaignsUsageStatistics returns the current site's campaigns usage.
func GetCampaignsUsageStatistics(ctx context.Context) (*types.CampaignsUsageStatistics, error) {
	const campaignQuery = `
WITH changeset_summary AS (
    SELECT changesets.id changeset_id, cj.id job_id, changesets.external_state external_state FROM changesets
    LEFT JOIN changeset_jobs cj ON changesets.id = cj.changeset_id
)
SELECT (SELECT COUNT(*) from campaigns) as campaign_count,
       (SELECT COUNT(*) FROM changeset_summary WHERE job_id is not null) AS generated_changesets,
       (SELECT COUNT(*) FROM changeset_summary WHERE job_id is not null and external_state = 'MERGED') AS generated_changesets_merged,
       (SELECT COUNT(*) FROM changeset_summary WHERE job_id is null) AS manual_changesets,
       (SELECT COUNT(*) FROM changeset_summary WHERE job_id is null and external_state = 'MERGED') AS manual_changesets_merged;
`
	var (
		campaignCount             int
		generatedChangesets       int
		generatedChangesetsMerged int
		manualChangesets          int
		manualChangesetsMerged    int
	)

	if err := dbconn.Global.QueryRowContext(ctx, campaignQuery).Scan(
		&campaignCount,
		&generatedChangesets,
		&generatedChangesetsMerged,
		&manualChangesets,
		&manualChangesetsMerged,
	); err != nil {
		return nil, err
	}

	return &types.CampaignsUsageStatistics{
		CampaignsCount:               int32(campaignCount),
		GeneratedChangesetCount:      int32(generatedChangesets),
		GeneratedChangesetMergeCount: int32(generatedChangesetsMerged),
		ManualChangesetCount:         int32(manualChangesets),
		ManualChangesetMergeCount:    int32(manualChangesetsMerged),
	}, nil
}
