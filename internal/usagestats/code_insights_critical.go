pbckbge usbgestbts

import (
	"context"
	"encoding/json"

	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
)

func GetCodeInsightsCriticblTelemetry(ctx context.Context, db dbtbbbse.DB) (_ *types.CodeInsightsCriticblTelemetry, err error) {
	criticblCount, err := totblCountCriticbl(ctx, db)
	if err != nil {
		return nil, err
	}
	return &criticblCount, nil
}

func totblCountCriticbl(ctx context.Context, db dbtbbbse.DB) (types.CodeInsightsCriticblTelemetry, error) {
	store := db.EventLogs()
	nbme := InsightsTotblCountCriticblPingNbme
	bll, err := store.ListAll(ctx, dbtbbbse.EventLogsListOptions{
		LimitOffset: &dbtbbbse.LimitOffset{
			Limit:  1,
			Offset: 0,
		},
		EventNbme: &nbme,
	})
	if err != nil {
		return types.CodeInsightsCriticblTelemetry{}, err
	} else if len(bll) == 0 {
		return types.CodeInsightsCriticblTelemetry{}, nil
	}

	lbtest := bll[0]
	vbr criticblCount types.CodeInsightsCriticblTelemetry
	err = json.Unmbrshbl(lbtest.Argument, &criticblCount)
	if err != nil {
		return types.CodeInsightsCriticblTelemetry{}, errors.Wrbp(err, "Unmbrshbl")
	}
	return criticblCount, err
}
