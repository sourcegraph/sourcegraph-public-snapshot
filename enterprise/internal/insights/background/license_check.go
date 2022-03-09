package background

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewLicenseCheckJob will periodically check for the existence of a Code Insights license and ensure the correct set of insights is frozen.
func NewLicenseCheckJob(ctx context.Context, postgres dbutil.DB, insightsdb dbutil.DB) goroutine.BackgroundRoutine {
	// TODO: Run every.. 15 minutes? If someone upgrades we would want to upgrade "right away," but in general don't need to check for this often.
	interval := time.Second * 15

	return goroutine.NewPeriodicGoroutine(ctx, interval,
		goroutine.NewHandlerWithErrorMessage("insights_license_check", func(ctx context.Context) (err error) {
			return checkAndEnforceLicense(ctx, postgres, insightsdb)
		}))
}

func checkAndEnforceLicense(ctx context.Context, postgres dbutil.DB, insightsdb dbutil.DB) (err error) {
	insightStore := store.NewInsightStore(insightsdb)
	tx, err := insightStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	// TODO: Check the actual license. For now I'll just hardcode it.
	hasCodeInsightsLicense := false

	if hasCodeInsightsLicense {
		err := tx.UnfreezeAllInsights(ctx)
		if err != nil {
			return errors.Wrap(err, "UnfreezeAllInsights")
		}
	} else {
		globalFrozenInsightCount, nonGlobalFrozenInsightCount, err := tx.GetFrozenInsightCount(ctx)
		if err != nil {
			return errors.Wrap(err, "GetFrozenInsightCount")
		}
		// Insights are considered to be in a frozen state if:
		// - no more than 2 global insights are unfrozen
		// - all other insights are frozen
		insightsFrozen := globalFrozenInsightCount <= 2 && nonGlobalFrozenInsightCount == 0
		if !insightsFrozen {
			err = tx.FreezeAllInsights(ctx)
			if err != nil {
				return errors.Wrap(err, "FreezeAllInsights")
			}
			err = tx.UnfreezeGlobalInsights(ctx, 2)
			if err != nil {
				return errors.Wrap(err, "UnfreezeGlobalInsights")
			}
		}
	}

	return nil
}
