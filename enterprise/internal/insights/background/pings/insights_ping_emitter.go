package pings

import (
	"context"
	"encoding/json"
	"time"

	"github.com/inconshreveable/log15"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/usagestats"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewInsightsPingEmitterJob will emit pings from Code Insights that involve enterprise features such as querying
// directly against the code insights database.
func NewInsightsPingEmitterJob(ctx context.Context, base database.DB, insights edb.InsightsDB) goroutine.BackgroundRoutine {
	interval := time.Minute * 60
	e := InsightsPingEmitter{
		postgresDb: base,
		insightsDb: insights,
	}

	return goroutine.NewPeriodicGoroutine(ctx, interval,
		goroutine.NewHandlerWithErrorMessage("insights_pings_emitter", e.emit))
}

type InsightsPingEmitter struct {
	postgresDb database.DB
	insightsDb edb.InsightsDB
}

func (e *InsightsPingEmitter) emit(ctx context.Context) error {
	log15.Info("Emitting Code Insights Pings")

	type emitter func(ctx context.Context) error
	var emitters = map[string]emitter{
		"emitInsightTotalCounts":      e.emitInsightTotalCounts,
		"emitIntervalCounts":          e.emitIntervalCounts,
		"emitOrgVisibleInsightCounts": e.emitOrgVisibleInsightCounts,
		"emitTotalOrgsWithDashboard":  e.emitTotalOrgsWithDashboard,
		"emitTotalDashboards":         e.emitTotalDashboards,
		"emitInsightsPerDashboard":    e.emitInsightsPerDashboard,
		"emitTotalCountCritical":      e.emitTotalCountCritical,
	}
	hasError := false
	for name, delegate := range emitters {
		err := delegate(ctx)
		if err != nil {
			log15.Error(errors.Wrap(err, name).Error())
			hasError = true
		}
	}
	if hasError {
		log15.Error("Code Insights ping emitter encountered errors. Errors were skipped")
	}

	return nil
}

func (e *InsightsPingEmitter) emitInsightTotalCounts(ctx context.Context) error {
	var counts types.InsightTotalCounts
	byViewType, err := e.GetTotalCountByViewType(ctx)
	if err != nil {
		return errors.Wrap(err, "GetTotalCountByViewType")
	}
	counts.ViewCounts = byViewType

	bySeriesType, err := e.GetTotalCountBySeriesType(ctx)
	if err != nil {
		return errors.Wrap(err, "GetTotalCountBySeriesType")
	}
	counts.SeriesCounts = bySeriesType

	byViewSeriesType, err := e.GetTotalCountByViewSeriesType(ctx)
	if err != nil {
		return errors.Wrap(err, "GetTotalCountByViewSeriesType")
	}
	counts.ViewSeriesCounts = byViewSeriesType

	marshal, err := json.Marshal(counts)
	if err != nil {
		return errors.Wrap(err, "Marshal")
	}

	err = e.SaveEvent(ctx, usagestats.InsightsTotalCountPingName, marshal)
	if err != nil {
		return errors.Wrap(err, "SaveEvent")
	}
	return nil
}

func (e *InsightsPingEmitter) emitTotalCountCritical(ctx context.Context) error {
	var arg types.CodeInsightsCriticalTelemetry
	count, err := e.GetTotalCountCritical(ctx)
	if err != nil {
		return errors.Wrap(err, "GetTotalCountCritical")
	}
	arg.TotalInsights = int32(count)

	marshal, err := json.Marshal(arg)
	if err != nil {
		return errors.Wrap(err, "Marshal")
	}

	err = e.SaveEvent(ctx, usagestats.InsightsTotalCountCriticalPingName, marshal)
	if err != nil {
		return errors.Wrap(err, "SaveEvent")
	}
	return nil
}

func (e *InsightsPingEmitter) emitIntervalCounts(ctx context.Context) error {
	counts, err := e.GetIntervalCounts(ctx)
	if err != nil {
		return errors.Wrap(err, "GetIntervalCounts")
	}

	marshal, err := json.Marshal(counts)
	if err != nil {
		return errors.Wrap(err, "Marshal")
	}

	err = e.SaveEvent(ctx, usagestats.InsightsIntervalCountsPingName, marshal)
	if err != nil {
		return errors.Wrap(err, "SaveEvent")
	}
	return nil
}

func (e *InsightsPingEmitter) emitOrgVisibleInsightCounts(ctx context.Context) error {
	counts, err := e.GetOrgVisibleInsightCounts(ctx)
	if err != nil {
		return errors.Wrap(err, "GetOrgVisibleInsightCounts")
	}

	marshal, err := json.Marshal(counts)
	if err != nil {
		return errors.Wrap(err, "Marshal")
	}

	err = e.SaveEvent(ctx, usagestats.InsightsOrgVisibleInsightsPingName, marshal)
	if err != nil {
		return errors.Wrap(err, "SaveEvent")
	}
	return nil
}

func (e *InsightsPingEmitter) emitTotalOrgsWithDashboard(ctx context.Context) error {
	counts, err := e.GetTotalOrgsWithDashboard(ctx)
	if err != nil {
		return errors.Wrap(err, "GetTotalOrgsWithDashboard")
	}

	marshal, err := json.Marshal(counts)
	if err != nil {
		return errors.Wrap(err, "Marshal")
	}

	err = e.SaveEvent(ctx, usagestats.InsightsTotalOrgsWithDashboardPingName, marshal)
	if err != nil {
		return errors.Wrap(err, "SaveEvent")
	}
	return nil
}

func (e *InsightsPingEmitter) emitTotalDashboards(ctx context.Context) error {
	counts, err := e.GetTotalDashboards(ctx)
	if err != nil {
		return errors.Wrap(err, "GetTotalDashboards")
	}

	marshal, err := json.Marshal(counts)
	if err != nil {
		return errors.Wrap(err, "Marshal")
	}

	err = e.SaveEvent(ctx, usagestats.InsightsDashboardTotalCountPingName, marshal)
	if err != nil {
		return errors.Wrap(err, "SaveEvent")
	}
	return nil
}

func (e *InsightsPingEmitter) emitInsightsPerDashboard(ctx context.Context) error {
	counts, err := e.GetInsightsPerDashboard(ctx)
	if err != nil {
		return errors.Wrap(err, "GetInsightsPerDashboard")
	}

	marshal, err := json.Marshal(counts)
	if err != nil {
		return errors.Wrap(err, "Marshal")
	}

	err = e.SaveEvent(ctx, usagestats.InsightsPerDashboardPingName, marshal)
	if err != nil {
		return errors.Wrap(err, "SaveEvent")
	}
	return nil
}

func (e *InsightsPingEmitter) SaveEvent(ctx context.Context, name string, argument json.RawMessage) error {
	store := e.postgresDb.EventLogs()

	err := store.Insert(ctx, &database.Event{
		Name:            name,
		UserID:          0,
		AnonymousUserID: "backend",
		Argument:        argument,
		Timestamp:       time.Now(),
		Source:          "BACKEND",
	})
	if err != nil {
		return err
	}
	return nil
}
