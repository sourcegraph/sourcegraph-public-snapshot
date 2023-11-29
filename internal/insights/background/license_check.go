package background

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	edb "github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewLicenseCheckJob will periodically check for the existence of a Code Insights license and ensure the correct set of insights is frozen.
func NewLicenseCheckJob(ctx context.Context, postgres database.DB, insightsdb edb.InsightsDB) goroutine.BackgroundRoutine {
	interval := time.Minute * 15
	logger := log.Scoped("CodeInsightsLicenseCheckJob")

	return goroutine.NewPeriodicGoroutine(
		ctx,
		goroutine.HandlerFunc(
			func(ctx context.Context) (err error) {
				return checkAndEnforceLicense(ctx, insightsdb, logger)
			},
		),
		goroutine.WithName("insights.license_check"),
		goroutine.WithDescription("checks for code insights license and freezes insights when missing"),
		goroutine.WithInterval(interval),
	)
}

func checkAndEnforceLicense(ctx context.Context, insightsdb edb.InsightsDB, logger log.Logger) (err error) {
	insightStore := store.NewInsightStore(insightsdb)
	dashboardStore := store.NewDashboardStore(insightsdb)
	insightTx, err := insightStore.Transact(ctx)
	if err != nil {
		return err
	}
	dashboardTx := dashboardStore.With(insightTx)
	defer func() { err = insightTx.Done(err) }()

	licenseError := licensing.Check(licensing.FeatureCodeInsights)
	if licenseError == nil {
		err := insightTx.UnfreezeAllInsights(ctx)
		if err != nil {
			return errors.Wrap(err, "UnfreezeAllInsights")
		}
		return nil
	}

	logger.Info("No license found for Code Insights. Freezing insights for limited access mode", log.Error(licenseError))

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
