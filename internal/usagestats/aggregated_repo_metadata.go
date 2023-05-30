package usagestats

import (
	"context"
	"database/sql"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func GetAggregatedRepoMetadataStats(ctx context.Context, db database.DB) (*types.RepoMetadataAggregatedStats, error) {
	now := time.Now().UTC()
	daily, err := db.EventLogs().AggregatedRepoMetadataEvents(ctx, now, database.Daily)
	if err != nil {
		return nil, err
	}

	weekly, err := db.EventLogs().AggregatedRepoMetadataEvents(ctx, now, database.Weekly)
	if err != nil {
		return nil, err
	}

	monthly, err := db.EventLogs().AggregatedRepoMetadataEvents(ctx, now, database.Monthly)
	if err != nil {
		return nil, err
	}

	summary, err := getAggregatedRepoMetadataSummary(ctx, db)
	if err != nil {
		return nil, err
	}

	return &types.RepoMetadataAggregatedStats{
		Summary: summary,
		Daily:   daily,
		Weekly:  weekly,
		Monthly: monthly,
	}, nil
}

func getAggregatedRepoMetadataSummary(ctx context.Context, db database.DB) (*types.RepoMetadataAggregatedSummary, error) {
	q := `
	SELECT
		COUNT(*) AS total_count,
		COUNT(DISTINCT repo_id) AS total_repos_count
	FROM repo_kvps
	`
	var summary types.RepoMetadataAggregatedSummary
	err := db.QueryRowContext(ctx, q).Scan(&summary.RepoMetadataCount, &summary.ReposWithMetadataCount)
	if err != nil {
		return nil, err
	}

	flag, err := db.FeatureFlags().GetFeatureFlag(ctx, "repository-metadata")
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
	}
	summary.IsEnabled = flag == nil || flag.Bool.Value

	return &summary, nil
}
