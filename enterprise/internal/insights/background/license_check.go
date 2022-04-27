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
	dashboardStore := store.NewDashboardStore(insightsdb)
	insightTx, err := insightStore.Transact(ctx)
	dashboardTx := dashboardStore.With(insightTx)
	if err != nil {
		return err
	}
	defer func() { err = insightTx.Done(err) }()

	licenseError := licensing.Check(licensing.FeatureCodeInsights)
	if licenseError == nil {
		err := insightTx.UnfreezeAllInsights(ctx)
		if err != nil {
			return errors.Wrap(err, "UnfreezeAllInsights")
		}
		return nil
	}

	log15.Info("No license found for Code Insights. Freezing insights for limited access mode", "error", licenseError.Error())

	globalUnfrozenInsightCount, totalUnfrozenInsightCount, err := insightTx.GetUnfrozenInsightCount(ctx)
	if err != nil {
		return errors.Wrap(err, "GetUnfrozenInsightCount")
	}
	// Insights need to be frozen if:
	// - more than 2 global insights are unfrozen
	// - any other insights are unfrozen
	shouldFreeze := globalUnfrozenInsightCount > 2 || totalUnfrozenInsightCount != globalUnfrozenInsightCount
	if shouldFreeze {
		err = insightTx.FreezeAllInsights(ctx)
		if err != nil {
			return errors.Wrap(err, "FreezeAllInsights")
		}
		err = insightTx.UnfreezeGlobalInsights(ctx, 2)
		if err != nil {
			return errors.Wrap(err, "UnfreezeGlobalInsights")
		}

		// Attach the unfrozen insights to the limited access mode dashboard
		dashboardId, err := dashboardTx.EnsureLimitedAccessModeDashboard(ctx)
		if err != nil {
			return errors.Wrap(err, "EnsureLimitedAccessModeDashboard")
		}
		insightUniqueIds, err := insightTx.GetUnfrozenInsightUniqueIds(ctx)
		if err != nil {
			return errors.Wrap(err, "GetUnfrozenInsightIds")
		}
		err = dashboardTx.AddViewsToDashboard(ctx, dashboardId, insightUniqueIds)
		if err != nil {
			return errors.Wrap(err, "AddViewsToDashboard")
		}
	}

	return nil
}
