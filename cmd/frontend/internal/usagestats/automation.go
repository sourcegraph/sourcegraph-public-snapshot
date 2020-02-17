package usagestats

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
)

// GetAutomationUsageStatistics returns the current site's automation usage.
func GetAutomationUsageStatistics(ctx context.Context) (*types.AutomationUsageStatistics, error) {
	const q = "SELECT COUNT(*) FROM campaigns;"

	var count int
	if err := dbconn.Global.QueryRowContext(ctx, q).Scan(&count); err != nil {
		return nil, err
	}

	return &types.AutomationUsageStatistics{
		CampaignsCount: count,
	}, nil
}
