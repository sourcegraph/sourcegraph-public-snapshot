package usagestats

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
)

// GetCampaignsUsageStatistics returns the current site's campaigns usage.
func GetCampaignsUsageStatistics(ctx context.Context) (*types.CampaignsUsageStatistics, error) {
	stats := types.CampaignsUsageStatistics{}

	const changesetCountsQuery = `
SELECT
    (SELECT COUNT(*) FROM campaigns) AS campaigns_count,
    COUNT(*)                        FILTER (WHERE owned_by_campaign_id IS NOT NULL AND publication_state = 'PUBLISHED') AS action_changesets,
    COALESCE(SUM(diff_stat_added)   FILTER (WHERE owned_by_campaign_id IS NOT NULL AND publication_state = 'PUBLISHED'), 0) AS action_changesets_diff_stat_added_sum,
    COALESCE(SUM(diff_stat_changed) FILTER (WHERE owned_by_campaign_id IS NOT NULL AND publication_state = 'PUBLISHED'), 0) AS action_changesets_diff_stat_changed_sum,
    COALESCE(SUM(diff_stat_deleted) FILTER (WHERE owned_by_campaign_id IS NOT NULL AND publication_state = 'PUBLISHED'), 0) AS action_changesets_diff_stat_deleted_sum,
    COUNT(*)                        FILTER (WHERE owned_by_campaign_id IS NOT NULL AND publication_state = 'PUBLISHED' AND external_state = 'MERGED') AS action_changesets_merged,
    COALESCE(SUM(diff_stat_added)   FILTER (WHERE owned_by_campaign_id IS NOT NULL AND publication_state = 'PUBLISHED' AND external_state = 'MERGED'), 0) AS action_changesets_merged_diff_stat_added_sum,
    COALESCE(SUM(diff_stat_changed) FILTER (WHERE owned_by_campaign_id IS NOT NULL AND publication_state = 'PUBLISHED' AND external_state = 'MERGED'), 0) AS action_changesets_merged_diff_stat_changed_sum,
    COALESCE(SUM(diff_stat_deleted) FILTER (WHERE owned_by_campaign_id IS NOT NULL AND publication_state = 'PUBLISHED' AND external_state = 'MERGED'), 0) AS action_changesets_merged_diff_stat_deleted_sum,
    COUNT(*) FILTER (WHERE added_to_campaign) AS manual_changesets,
    COUNT(*) FILTER (WHERE added_to_campaign AND external_state = 'MERGED') AS manual_changesets_merged
FROM changesets;
`
	if err := dbconn.Global.QueryRowContext(ctx, changesetCountsQuery).Scan(
		&stats.CampaignsCount,
		&stats.ActionChangesetsCount,
		&stats.ActionChangesetsDiffStatAddedSum,
		&stats.ActionChangesetsDiffStatChangedSum,
		&stats.ActionChangesetsDiffStatDeletedSum,
		&stats.ActionChangesetsMergedCount,
		&stats.ActionChangesetsMergedDiffStatAddedSum,
		&stats.ActionChangesetsMergedDiffStatChangedSum,
		&stats.ActionChangesetsMergedDiffStatDeletedSum,
		&stats.ManualChangesetsCount,
		&stats.ManualChangesetsMergedCount,
	); err != nil {
		return nil, err
	}

	const specsCountsQuery = `
SELECT
    COUNT(*) AS campaign_specs_created,
    COALESCE(SUM(jsonb_array_length(argument->'changeset_spec_rand_ids')), 0) AS changeset_specs_created_count
FROM event_logs
WHERE name = 'CampaignSpecCreated';
`

	err := dbconn.Global.QueryRowContext(ctx, specsCountsQuery).Scan(
		&stats.CampaignSpecsCreatedCount,
		&stats.ChangesetSpecsCreatedCount,
	)

	return &stats, err
}
