package usagestats

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/types"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

func GetCodeInsightsCriticalTelemetry(ctx context.Context, db database.DB) (_ *types.CodeInsightsCriticalTelemetry, err error) {
	telemetry := &types.CodeInsightsCriticalTelemetry{}

	totalCount, err := totalCountCritical(ctx, db)
	if err != nil {
		return nil, err
	}
	telemetry.TotalInsights = totalCount

	return telemetry, nil
}

func totalCountCritical(ctx context.Context, db database.DB) (int32, error) {
	counts, err := GetTotalInsightCounts(ctx, db)
	if err != nil {
		return 0, err
	}

	sum := 0
	for _, count := range counts.ViewCounts {
		sum += count.TotalCount
	}
	return int32(sum), nil
}
