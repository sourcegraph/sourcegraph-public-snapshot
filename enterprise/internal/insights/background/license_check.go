package background

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewLicenseCheckJob will periodically check for the existence of a Code Insights license and ensure the correct set of insights is frozen.
func NewLicenseCheckJob(ctx context.Context, postgres dbutil.DB, insightsdb dbutil.DB) goroutine.BackgroundRoutine {
	interval := time.Minute * 15

	return goroutine.NewPeriodicGoroutine(ctx, interval,
		goroutine.NewHandlerWithErrorMessage("insights_license_check", func(ctx context.Context) (err error) {
			return checkAndEnforceLicense(ctx, insightsdb)
		}))
}

func checkAndEnforceLicense(ctx context.Context, insightsdb dbutil.DB) (err error) {
	insightStore := store.NewInsightStore(insightsdb)
	tx, err := insightStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	info, err := licensing.GetConfiguredProductLicenseInfo()
	if err != nil {
		return errors.Wrap(err, "GetConfiguredProductLicenseInfo")
	}

	if info.Plan().HasFeature(licensing.FeatureCodeInsights) {
		err := tx.UnfreezeAllInsights(ctx)
		if err != nil {
			return errors.Wrap(err, "UnfreezeAllInsights")
		}
		return nil
	}

	log15.Info("No license found for Code Insights. Freezing insights for limited access mode.")
	globalUnfrozenInsightCount, totalUnfrozenInsightCount, err := tx.GetUnfrozenInsightCount(ctx)
	if err != nil {
		return errors.Wrap(err, "GetUnfrozenInsightCount")
	}
	// Insights need to be frozen if:
	// - more than 2 global insights are unfrozen
	// - any other insights are unfrozen
	shouldFreeze := globalUnfrozenInsightCount > 2 || totalUnfrozenInsightCount != globalUnfrozenInsightCount
	if shouldFreeze {
		err = tx.FreezeAllInsights(ctx)
		if err != nil {
			return errors.Wrap(err, "FreezeAllInsights")
		}
		err = tx.UnfreezeGlobalInsights(ctx, 2)
		if err != nil {
			return errors.Wrap(err, "UnfreezeGlobalInsights")
		}
	}

	return nil
}
