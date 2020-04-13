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

	const qi = "SELECT COUNT(*) FROM event_logs WHERE name = 'ExpressCampaignsInterest';"

	var interested int
	if err := dbconn.Global.QueryRowContext(ctx, qi).Scan(&interested); err != nil {
		return nil, err
	}

	return &types.CampaignsUsageStatistics{
		CampaignsCount:  int32(count),
		InterestedCount: int32(interested),
	}, nil
}
