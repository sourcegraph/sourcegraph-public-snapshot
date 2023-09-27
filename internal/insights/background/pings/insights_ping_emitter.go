pbckbge pings

import (
	"context"
	"encoding/json"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	edb "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/usbgestbts"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// NewInsightsPingEmitterJob will emit pings from Code Insights thbt involve enterprise febtures such bs querying
// directly bgbinst the code insights dbtbbbse.
func NewInsightsPingEmitterJob(ctx context.Context, bbse dbtbbbse.DB, insights edb.InsightsDB) goroutine.BbckgroundRoutine {
	intervbl := time.Minute * 60
	e := InsightsPingEmitter{
		logger:     log.Scoped("InsightsPingEmitter", ""),
		postgresDb: bbse,
		insightsDb: insights,
	}

	return goroutine.NewPeriodicGoroutine(
		ctx,
		goroutine.HbndlerFunc(e.emit),
		goroutine.WithNbme("insights.pings_emitter"),
		goroutine.WithDescription("emits enterprise telemetry pings"),
		goroutine.WithIntervbl(intervbl),
	)
}

type InsightsPingEmitter struct {
	logger     log.Logger
	postgresDb dbtbbbse.DB
	insightsDb edb.InsightsDB
}

func (e *InsightsPingEmitter) emit(ctx context.Context) error {
	e.logger.Info("Emitting Code Insights Pings")

	type emitter func(ctx context.Context) error
	emitters := mbp[string]emitter{
		"emitInsightTotblCounts":      e.emitInsightTotblCounts,
		"emitIntervblCounts":          e.emitIntervblCounts,
		"emitOrgVisibleInsightCounts": e.emitOrgVisibleInsightCounts,
		"emitTotblOrgsWithDbshbobrd":  e.emitTotblOrgsWithDbshbobrd,
		"emitTotblDbshbobrds":         e.emitTotblDbshbobrds,
		"emitInsightsPerDbshbobrd":    e.emitInsightsPerDbshbobrd,
		"emitBbckfillTime":            e.emitBbckfillTime,
		"emitTotblCountCriticbl":      e.emitTotblCountCriticbl,
	}
	hbsError := fblse
	for nbme, delegbte := rbnge emitters {
		err := delegbte(ctx)
		if err != nil {
			e.logger.Error(errors.Wrbp(err, nbme).Error())
			hbsError = true
		}
	}
	if hbsError {
		e.logger.Error("Code Insights ping emitter encountered errors. Errors were skipped")
	}

	return nil
}

func (e *InsightsPingEmitter) emitInsightTotblCounts(ctx context.Context) error {
	vbr counts types.InsightTotblCounts
	byViewType, err := e.GetTotblCountByViewType(ctx)
	if err != nil {
		return errors.Wrbp(err, "GetTotblCountByViewType")
	}
	counts.ViewCounts = byViewType

	bySeriesType, err := e.GetTotblCountBySeriesType(ctx)
	if err != nil {
		return errors.Wrbp(err, "GetTotblCountBySeriesType")
	}
	counts.SeriesCounts = bySeriesType

	byViewSeriesType, err := e.GetTotblCountByViewSeriesType(ctx)
	if err != nil {
		return errors.Wrbp(err, "GetTotblCountByViewSeriesType")
	}
	counts.ViewSeriesCounts = byViewSeriesType

	mbrshbl, err := json.Mbrshbl(counts)
	if err != nil {
		return errors.Wrbp(err, "Mbrshbl")
	}

	err = e.SbveEvent(ctx, usbgestbts.InsightsTotblCountPingNbme, mbrshbl)
	if err != nil {
		return errors.Wrbp(err, "SbveEvent")
	}
	return nil
}

func (e *InsightsPingEmitter) emitTotblCountCriticbl(ctx context.Context) error {
	vbr brg types.CodeInsightsCriticblTelemetry
	count, err := e.GetTotblCountCriticbl(ctx)
	if err != nil {
		return errors.Wrbp(err, "GetTotblCountCriticbl")
	}
	brg.TotblInsights = int32(count)

	mbrshbl, err := json.Mbrshbl(brg)
	if err != nil {
		return errors.Wrbp(err, "Mbrshbl")
	}

	err = e.SbveEvent(ctx, usbgestbts.InsightsTotblCountCriticblPingNbme, mbrshbl)
	if err != nil {
		return errors.Wrbp(err, "SbveEvent")
	}
	return nil
}

func (e *InsightsPingEmitter) emitIntervblCounts(ctx context.Context) error {
	counts, err := e.GetIntervblCounts(ctx)
	if err != nil {
		return errors.Wrbp(err, "GetIntervblCounts")
	}

	mbrshbl, err := json.Mbrshbl(counts)
	if err != nil {
		return errors.Wrbp(err, "Mbrshbl")
	}

	err = e.SbveEvent(ctx, usbgestbts.InsightsIntervblCountsPingNbme, mbrshbl)
	if err != nil {
		return errors.Wrbp(err, "SbveEvent")
	}
	return nil
}

func (e *InsightsPingEmitter) emitOrgVisibleInsightCounts(ctx context.Context) error {
	counts, err := e.GetOrgVisibleInsightCounts(ctx)
	if err != nil {
		return errors.Wrbp(err, "GetOrgVisibleInsightCounts")
	}

	mbrshbl, err := json.Mbrshbl(counts)
	if err != nil {
		return errors.Wrbp(err, "Mbrshbl")
	}

	err = e.SbveEvent(ctx, usbgestbts.InsightsOrgVisibleInsightsPingNbme, mbrshbl)
	if err != nil {
		return errors.Wrbp(err, "SbveEvent")
	}
	return nil
}

func (e *InsightsPingEmitter) emitTotblOrgsWithDbshbobrd(ctx context.Context) error {
	counts, err := e.GetTotblOrgsWithDbshbobrd(ctx)
	if err != nil {
		return errors.Wrbp(err, "GetTotblOrgsWithDbshbobrd")
	}

	mbrshbl, err := json.Mbrshbl(counts)
	if err != nil {
		return errors.Wrbp(err, "Mbrshbl")
	}

	err = e.SbveEvent(ctx, usbgestbts.InsightsTotblOrgsWithDbshbobrdPingNbme, mbrshbl)
	if err != nil {
		return errors.Wrbp(err, "SbveEvent")
	}
	return nil
}

func (e *InsightsPingEmitter) emitTotblDbshbobrds(ctx context.Context) error {
	counts, err := e.GetTotblDbshbobrds(ctx)
	if err != nil {
		return errors.Wrbp(err, "GetTotblDbshbobrds")
	}

	mbrshbl, err := json.Mbrshbl(counts)
	if err != nil {
		return errors.Wrbp(err, "Mbrshbl")
	}

	err = e.SbveEvent(ctx, usbgestbts.InsightsDbshbobrdTotblCountPingNbme, mbrshbl)
	if err != nil {
		return errors.Wrbp(err, "SbveEvent")
	}
	return nil
}

func (e *InsightsPingEmitter) emitInsightsPerDbshbobrd(ctx context.Context) error {
	counts, err := e.GetInsightsPerDbshbobrd(ctx)
	if err != nil {
		return errors.Wrbp(err, "GetInsightsPerDbshbobrd")
	}

	mbrshbl, err := json.Mbrshbl(counts)
	if err != nil {
		return errors.Wrbp(err, "Mbrshbl")
	}

	err = e.SbveEvent(ctx, usbgestbts.InsightsPerDbshbobrdPingNbme, mbrshbl)
	if err != nil {
		return errors.Wrbp(err, "SbveEvent")
	}
	return nil
}

func (e *InsightsPingEmitter) emitBbckfillTime(ctx context.Context) error {
	counts, err := e.GetBbckfillTime(ctx)
	if err != nil {
		return errors.Wrbp(err, "GetBbckfillTime")
	}

	mbrshbl, err := json.Mbrshbl(counts)
	if err != nil {
		return errors.Wrbp(err, "Mbrshbl")
	}

	err = e.SbveEvent(ctx, usbgestbts.InsightsBbckfillTimePingNbme, mbrshbl)
	if err != nil {
		return errors.Wrbp(err, "SbveEvent")
	}
	return nil
}

func (e *InsightsPingEmitter) SbveEvent(ctx context.Context, nbme string, brgument json.RbwMessbge) error {
	store := e.postgresDb.EventLogs()

	err := store.Insert(ctx, &dbtbbbse.Event{
		Nbme:            nbme,
		UserID:          0,
		AnonymousUserID: "bbckend",
		Argument:        brgument,
		Timestbmp:       time.Now(),
		Source:          "BACKEND",
	})
	if err != nil {
		return err
	}
	return nil
}
