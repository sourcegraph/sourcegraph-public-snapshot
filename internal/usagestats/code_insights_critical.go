package usagestats

import (
	"context"
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

func GetCodeInsightsCriticalTelemetry(ctx context.Context, db database.DB) (_ *types.CodeInsightsCriticalTelemetry, err error) {
	criticalCount, err := totalCountCritical(ctx, db)
	if err != nil {
		return nil, err
	}
	return &criticalCount, nil
}

func totalCountCritical(ctx context.Context, db database.DB) (types.CodeInsightsCriticalTelemetry, error) {
	store := db.EventLogs()
	name := InsightsTotalCountCriticalPingName
	all, err := store.ListAll(ctx, database.EventLogsListOptions{
		LimitOffset: &database.LimitOffset{
			Limit:  1,
			Offset: 0,
		},
		EventName: &name,
	})
	if err != nil {
		return types.CodeInsightsCriticalTelemetry{}, err
	} else if len(all) == 0 {
		return types.CodeInsightsCriticalTelemetry{}, nil
	}

	latest := all[0]
	var criticalCount types.CodeInsightsCriticalTelemetry
	err = json.Unmarshal(latest.Argument, &criticalCount)
	if err != nil {
		return types.CodeInsightsCriticalTelemetry{}, errors.Wrap(err, "Unmarshal")
	}
	return criticalCount, err
}
