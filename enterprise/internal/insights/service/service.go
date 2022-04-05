package service

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/database"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

func IsJustInTime(repositories []string) bool {
	return len(repositories) != 0
}

func CreateCodeMonitorInsight(ctx context.Context, insightsDB dbutil.DB, monitor *database.Monitor) error {
	insightStore := store.NewInsightStore(insightsDB)

	series, err := insightStore.CreateSeries(ctx, types.InsightSeries{
		SeriesID:                   CodeMonitorInsightSeriesID(monitor),
		Query:                      "N/A",
		GeneratedFromCaptureGroups: false,
		JustInTime:                 false,
		GenerationMethod:           types.CodeMonitor,
		SampleIntervalUnit:         string(types.Month), // this shouldn't be necessary but I can't see a clear way around this atm
		SampleIntervalValue:        0,
	})
	if err != nil {
		return err
	}

	view, err := insightStore.CreateView(ctx, types.InsightView{
		Title:            monitor.Description,
		UniqueID:         fmt.Sprintf("code-monitor-%d-view", monitor.ID),
		PresentationType: types.Line,
	}, []store.InsightViewGrant{store.UserGrant(int(monitor.UserID))})
	if err != nil {
		return err
	}

	err = insightStore.AttachSeriesToView(ctx, series, view, types.InsightViewSeriesMetadata{
		Label:  "results",
		Stroke: "blue",
	})
	if err != nil {
		return err
	}

	_, err = insightStore.StampBackfill(ctx, series)
	if err != nil {
		return err
	}

	return nil
}

func CodeMonitorInsightSeriesID(monitor *database.Monitor) string {
	return fmt.Sprintf("code-monitor-%d", monitor.ID)
}

func CodeMonitorInsightViewID(monitor *database.Monitor) string {
	return fmt.Sprintf("code-monitor-%d-view", monitor.ID)
}
