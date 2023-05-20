package usagestats

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/cody"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func GetAggregatedCodyStats(ctx context.Context, db database.DB) (*types.CodyAggregatedStats, error) {
	stats, err := db.EventLogs().AggregatedCodyStats(ctx, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	flag := cody.IsCodyEnabled(ctx)
	stats.IsEnabled = flag
	return stats, nil
}
