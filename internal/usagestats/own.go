package usagestats

import (
	"context"
	_ "embed"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

var (
	ownUsageStatsQuery string = `SELECT
		COUNT(*) AS total,
		COUNT(*) FILTER (WHERE o.id IS NOT NULL) AS with_ingested_ownership
		FROM repo AS r
		LEFT JOIN codeowners AS o
		ON r.id = o.repo_id;`
)

const (
	selectFileOwnersEventName   = "select:file.owners"
	fileHasOwnerEventName       = "file:has.owner"
	ownershipPanelOpenEventName = "ownershipPanelOpen"
)

func GetOwnershipUsageStats(ctx context.Context, db database.DB) (*types.OwnershipUsageStatistics, error) {
	var stats types.OwnershipUsageStatistics
	var totalReposCount int32
	if err := db.QueryRowContext(ctx, `SELECT total FROM repo_statistics`).Scan(&totalReposCount); err != nil {
		return nil, err
	}
	var ingestedOwnershipReposCount int32
	if err := db.QueryRowContext(ctx, `SELECT COUNT(DISTINCT repo_id) FROM codeowners`).Scan(&ingestedOwnershipReposCount); err != nil {
		return nil, err
	}
	stats.ReposCount = &types.OwnershipUsageReposCounts{
		Total:                 &totalReposCount,
		WithIngestedOwnership: &ingestedOwnershipReposCount,
	}
	fos ,err := db.FeatureFlags().GetOverridesForFlag(ctx, "search-ownership")
	if err != nil {
		return nil, err
	}
	featureFlagOn := false
	for _, fo := range fos {
		if fo.Value {
			featureFlagOn = true
			break
		}
	}
	stats.FeatureFlagOn = &featureFlagOn
	// At this poing we do not compute ReposCount.WithOwnership as this is really
	// computationally intensive (get all repos and query gitserver for each).
	// This will become very easy once we have versioned CODEOWNERS in the database.
	var reposCount types.OwnershipUsageReposCounts
	if err := db.QueryRowContext(ctx, ownUsageStatsQuery).Scan(
		&reposCount.Total,
		&reposCount.WithIngestedOwnership,
	); err != nil {
		return nil, err
	}
	stats.ReposCount = &reposCount
	activity, err := db.EventLogs().OwnershipFeatureActivity(ctx, time.Now(),
		selectFileOwnersEventName,
		fileHasOwnerEventName,
		ownershipPanelOpenEventName)
	if err != nil {
		return nil, err
	}
	stats.SelectFileOwnersSearch = activity[selectFileOwnersEventName]
	stats.FileHasOwnersSearch = activity[fileHasOwnerEventName]
	stats.OwnershipPanelOpened = activity[ownershipPanelOpenEventName]
	return &stats, nil
}
