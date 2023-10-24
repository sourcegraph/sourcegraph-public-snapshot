package usagestats

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestCodeInsightsUsageStatistics(t *testing.T) {
	ctx := context.Background()

	defer func() {
		timeNow = time.Now
	}()

	weekStart := time.Date(2021, 1, 25, 0, 0, 0, 0, time.UTC)
	now := time.Date(2021, 1, 28, 0, 0, 0, 0, time.UTC)
	mockTimeNow(now)

	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))

	_, err := db.ExecContext(context.Background(), `
		INSERT INTO event_logs
			(id, name, argument, url, user_id, anonymous_user_id, source, version, timestamp)
		VALUES
			(1, 'ViewInsights', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			(2, 'ViewInsights', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			(3, 'InsightAddition', '{"insightType": "searchInsights"}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			(4, 'InsightAddition', '{"insightType": "codeStatsInsights"}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			(5, 'InsightAddition', '{"insightType": "searchInsights"}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			(6, 'InsightEdit', '{"insightType": "searchInsights"}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', '3.23.0', $1::timestamp - interval '2 days'),
			(7, 'InsightAddition', '{"insightType": "codeStatsInsights"}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestamp - interval '8 days'),
			(8, 'CodeInsightsSearchBasedCreationPageSubmitClick', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestamp - interval '1 day')
	`, now)
	if err != nil {
		t.Fatal(err)
	}

	have, err := GetCodeInsightsUsageStatistics(ctx, db)
	if err != nil {
		t.Fatal(err)
	}

	zeroInt := int32(0)
	oneInt := int32(1)
	twoInt := int32(2)

	searchInsightsType := "searchInsights"
	codeStatsInsightsType := "codeStatsInsights"

	weeklyUsageStatisticsByInsight := []*types.InsightUsageStatistics{
		{
			InsightType:      &codeStatsInsightsType,
			Additions:        &oneInt,
			Edits:            &zeroInt,
			Removals:         &zeroInt,
			Hovers:           &zeroInt,
			UICustomizations: &zeroInt,
			DataPointClicks:  &zeroInt,
			FiltersChange:    &zeroInt,
		},
		{
			InsightType:      &searchInsightsType,
			Additions:        &twoInt,
			Edits:            &oneInt,
			Removals:         &zeroInt,
			Hovers:           &zeroInt,
			UICustomizations: &zeroInt,
			DataPointClicks:  &zeroInt,
			FiltersChange:    &zeroInt,
		},
	}

	want := &types.CodeInsightsUsageStatistics{
		WeeklyUsageStatisticsByInsight:               weeklyUsageStatisticsByInsight,
		WeeklyInsightsPageViews:                      &twoInt,
		WeeklyInsightsGetStartedPageViews:            &zeroInt,
		WeeklyInsightsUniquePageViews:                &oneInt,
		WeeklyInsightsGetStartedUniquePageViews:      &zeroInt,
		WeeklyInsightConfigureClick:                  &zeroInt,
		WeeklyInsightAddMoreClick:                    &zeroInt,
		WeekStart:                                    weekStart,
		WeeklyInsightCreators:                        &twoInt,
		WeeklyFirstTimeInsightCreators:               &oneInt,
		WeeklyGetStartedTabClickByTab:                []types.InsightGetStartedTabClickPing{},
		WeeklyGetStartedTabMoreClickByTab:            []types.InsightGetStartedTabClickPing{},
		TotalDashboardCount:                          &zeroInt,
		TotalOrgsWithDashboard:                       &zeroInt,
		WeeklyStandaloneDashboardClicks:              &zeroInt,
		WeeklyStandaloneInsightUniqueEditClicks:      &zeroInt,
		WeeklyStandaloneInsightUniquePageViews:       &zeroInt,
		WeeklyStandaloneInsightUniqueDashboardClicks: &zeroInt,
		WeeklyStandaloneInsightPageViews:             &zeroInt,
		WeeklyStandaloneEditClicks:                   &zeroInt,
		WeeklyGroupResultsOpenSection:                &zeroInt,
		WeeklyGroupResultsCollapseSection:            &zeroInt,
		WeeklyGroupResultsInfoIconHover:              &zeroInt,
		WeeklyDataExportClicks:                       &zeroInt,
	}

	wantedWeeklyUsage := []types.AggregatedPingStats{
		{Name: "CodeInsightsSearchBasedCreationPageSubmitClick", TotalCount: 1, UniqueCount: 1},
	}

	want.WeeklyAggregatedUsage = wantedWeeklyUsage

	want.WeeklyGroupResultsExpandedViewOpen = []types.GroupResultExpandedViewPing{}
	want.WeeklyGroupResultsExpandedViewCollapse = []types.GroupResultExpandedViewPing{}
	want.WeeklyGroupResultsChartBarHover = []types.GroupResultPing{}
	want.WeeklyGroupResultsChartBarClick = []types.GroupResultPing{}
	want.WeeklyGroupResultsAggregationModeClicked = []types.GroupResultPing{}
	want.WeeklyGroupResultsAggregationModeDisabledHover = []types.GroupResultPing{}

	if diff := cmp.Diff(want, have); diff != "" {
		t.Fatal(diff)
	}
}

func TestWithCreationPings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	now := time.Date(2021, 1, 28, 0, 0, 0, 0, time.UTC)

	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))

	user1 := "420657f0-d443-4d16-ac7d-003d8cdc91ef"
	user2 := "55555555-5555-5555-5555-555555555555"

	_, err := db.ExecContext(context.Background(), `
		INSERT INTO event_logs
			(id, name, argument, url, user_id, anonymous_user_id, source, version, timestamp)
		VALUES
			(1, 'ViewInsights', '{}', '', 1, $2, 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			(2, 'ViewInsights', '{}', '', 1, $2, 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			(3, 'ViewCodeInsightsCreationPage', '{}', '', 1, $2, 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			(4, 'ViewCodeInsightsCreationPage', '{}', '', 1, $2, 'WEB', '3.23.0', $1::timestamp - interval '10 days'),
			(5, 'ViewCodeInsightsCreationPage', '{}', '', 2, $3, 'WEB', '3.23.0', $1::timestamp - interval '2 days'),
			(6, 'ViewCodeInsightsCreationPage', '{}', '', 2, $3, 'WEB', '3.23.0', $1::timestamp - interval '2 days')
	`, now, user1, user2)
	if err != nil {
		t.Fatal(err)
	}

	want := map[types.PingName]types.AggregatedPingStats{
		"ViewCodeInsightsCreationPage": {Name: "ViewCodeInsightsCreationPage", UniqueCount: 2, TotalCount: 3},
	}

	stats := &types.CodeInsightsUsageStatistics{}
	err = getCreationViewUsage(ctx, db, stats, now)
	if err != nil {
		t.Fatal(err)
	}

	// convert into map so we can reliably test for equality
	got := make(map[types.PingName]types.AggregatedPingStats)
	for _, v := range stats.WeeklyAggregatedUsage {
		got[v.Name] = v
	}

	if !cmp.Equal(want, got) {
		t.Fatal(fmt.Sprintf("want: %v got: %v", want, got))
	}
}
