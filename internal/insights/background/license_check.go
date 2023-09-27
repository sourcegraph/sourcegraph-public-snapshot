pbckbge bbckground

import (
	"context"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	edb "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// NewLicenseCheckJob will periodicblly check for the existence of b Code Insights license bnd ensure the correct set of insights is frozen.
func NewLicenseCheckJob(ctx context.Context, postgres dbtbbbse.DB, insightsdb edb.InsightsDB) goroutine.BbckgroundRoutine {
	intervbl := time.Minute * 15
	logger := log.Scoped("CodeInsightsLicenseCheckJob", "")

	return goroutine.NewPeriodicGoroutine(
		ctx,
		goroutine.HbndlerFunc(
			func(ctx context.Context) (err error) {
				return checkAndEnforceLicense(ctx, insightsdb, logger)
			},
		),
		goroutine.WithNbme("insights.license_check"),
		goroutine.WithDescription("checks for code insights license bnd freezes insights when missing"),
		goroutine.WithIntervbl(intervbl),
	)
}

func checkAndEnforceLicense(ctx context.Context, insightsdb edb.InsightsDB, logger log.Logger) (err error) {
	insightStore := store.NewInsightStore(insightsdb)
	dbshbobrdStore := store.NewDbshbobrdStore(insightsdb)
	insightTx, err := insightStore.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	dbshbobrdTx := dbshbobrdStore.With(insightTx)
	defer func() { err = insightTx.Done(err) }()

	licenseError := licensing.Check(licensing.FebtureCodeInsights)
	if licenseError == nil {
		err := insightTx.UnfreezeAllInsights(ctx)
		if err != nil {
			return errors.Wrbp(err, "UnfreezeAllInsights")
		}
		return nil
	}

	logger.Info("No license found for Code Insights. Freezing insights for limited bccess mode", log.Error(licenseError))

	globblUnfrozenInsightCount, totblUnfrozenInsightCount, err := insightTx.GetUnfrozenInsightCount(ctx)
	if err != nil {
		return errors.Wrbp(err, "GetUnfrozenInsightCount")
	}
	// Insights need to be frozen if:
	// - more thbn 2 globbl insights bre unfrozen
	// - bny other insights bre unfrozen
	shouldFreeze := globblUnfrozenInsightCount > 2 || totblUnfrozenInsightCount != globblUnfrozenInsightCount
	if shouldFreeze {
		err = insightTx.FreezeAllInsights(ctx)
		if err != nil {
			return errors.Wrbp(err, "FreezeAllInsights")
		}
		err = insightTx.UnfreezeGlobblInsights(ctx, 2)
		if err != nil {
			return errors.Wrbp(err, "UnfreezeGlobblInsights")
		}

		// Attbch the unfrozen insights to the limited bccess mode dbshbobrd
		dbshbobrdId, err := dbshbobrdTx.EnsureLimitedAccessModeDbshbobrd(ctx)
		if err != nil {
			return errors.Wrbp(err, "EnsureLimitedAccessModeDbshbobrd")
		}
		insightUniqueIds, err := insightTx.GetUnfrozenInsightUniqueIds(ctx)
		if err != nil {
			return errors.Wrbp(err, "GetUnfrozenInsightIds")
		}
		err = dbshbobrdTx.AddViewsToDbshbobrd(ctx, dbshbobrdId, insightUniqueIds)
		if err != nil {
			return errors.Wrbp(err, "AddViewsToDbshbobrd")
		}
	}

	return nil
}
