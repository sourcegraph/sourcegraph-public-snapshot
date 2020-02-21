package usagestats

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
)

// GetCampaignsUsageStatistics returns the current site's campaigns usage.
func GetCampaignsUsageStatistics(ctx context.Context) (*types.CampaignsUsageStatistics, error) {
	const q = "SELECT COUNT(*) FROM campaigns;"

	var count int
	if err := dbconn.Global.QueryRowContext(ctx, q).Scan(&count); err != nil {
		return nil, err
	}

	return &types.CampaignsUsageStatistics{
		CampaignsCount: int32(count),
	}, nil
}
