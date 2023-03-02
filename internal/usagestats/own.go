package usagestats

import (
	"context"
	"database/sql"
	_ "embed"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

const (
	selectFileOwnersEventName   = "SelectFileOwnersSearch"
	fileHasOwnerEventName       = "FileHasOwnersSearch"
	ownershipPanelOpenEventName = "OwnershipPanelOpened"
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
		// At this poing we do not compute ReposCount.WithOwnership as this is really
		// computationally intensive (get all repos and query gitserver for each).
		// This will become very easy once we have versioned CODEOWNERS in the database.
	}
	featureFlagOn := false
	ff, err := db.FeatureFlags().GetFeatureFlag(ctx, "search-ownership")
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err != sql.ErrNoRows {
		featureFlagOn = ff.Bool.Value
	}
	stats.FeatureFlagOn = &featureFlagOn
	activity, err := db.EventLogs().OwnershipFeatureActivity(ctx, timeNow(),
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
