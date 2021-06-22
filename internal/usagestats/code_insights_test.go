package usagestats

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/jsonc"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
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

	db := dbtesting.GetDB(t)

	_, err := db.Exec(`
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
		},
		{
			InsightType:      &searchInsightsType,
			Additions:        &twoInt,
			Edits:            &oneInt,
			Removals:         &zeroInt,
			Hovers:           &zeroInt,
			UICustomizations: &zeroInt,
			DataPointClicks:  &zeroInt,
		},
	}

	want := &types.CodeInsightsUsageStatistics{
		WeeklyUsageStatisticsByInsight: weeklyUsageStatisticsByInsight,
		WeeklyInsightsPageViews:        &twoInt,
		WeeklyInsightsUniquePageViews:  &oneInt,
		WeeklyInsightConfigureClick:    &zeroInt,
		WeeklyInsightAddMoreClick:      &zeroInt,
		WeekStart:                      weekStart,
		WeeklyInsightCreators:          &twoInt,
		WeeklyFirstTimeInsightCreators: &oneInt,
	}

	wantedWeeklyUsage := []types.AggregatedPingStats{
		{Name: "CodeInsightsSearchBasedCreationPageSubmitClick", TotalCount: 1, UniqueCount: 1},
	}

	want.WeeklyAggregatedUsage = wantedWeeklyUsage
	want.InsightTimeIntervals = []types.InsightTimeIntervalPing{}
	want.InsightOrgVisible = []types.OrgVisibleInsightPing{{Type: "search"}, {Type: "lang-stats"}}

	if diff := cmp.Diff(want, have); diff != "" {
		t.Fatal(diff)
	}
}

func TestWithCreationPings(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2021, 1, 28, 0, 0, 0, 0, time.UTC)

	db := dbtesting.GetDB(t)

	user1 := "420657f0-d443-4d16-ac7d-003d8cdc91ef"
	user2 := "55555555-5555-5555-5555-555555555555"

	_, err := db.Exec(`
		INSERT INTO event_logs
			(id, name, argument, url, user_id, anonymous_user_id, source, version, timestamp)
		VALUES
			(1, 'ViewInsights', '{}', '', 1, $2, 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			(2, 'ViewInsights', '{}', '', 1, $2, 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			(3, 'ViewCodeInsightsCreationPage', '{}', '', 1, $2, 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			(4, 'ViewCodeInsightsCreationPage', '{}', '', 1, $2, 'WEB', '3.23.0', $1::timestamp - interval '10 days'),
			(5, 'CodeInsightsExploreInsightExtensionsClick', '{}', '', 2, $3, 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			(6, 'CodeInsightsExploreInsightExtensionsClick', '{}', '', 2, $3, 'WEB', '3.23.0', $1::timestamp - interval '10 days'),
			(7, 'ViewCodeInsightsCreationPage', '{}', '', 2, $3, 'WEB', '3.23.0', $1::timestamp - interval '2 days'),
			(8, 'ViewCodeInsightsCreationPage', '{}', '', 2, $3, 'WEB', '3.23.0', $1::timestamp - interval '2 days')
	`, now, user1, user2)
	if err != nil {
		t.Fatal(err)
	}

	want := map[types.PingName]types.AggregatedPingStats{
		"CodeInsightsExploreInsightExtensionsClick": {Name: "CodeInsightsExploreInsightExtensionsClick", UniqueCount: 1, TotalCount: 1},
		"ViewCodeInsightsCreationPage":              {Name: "ViewCodeInsightsCreationPage", UniqueCount: 2, TotalCount: 3},
	}

	results, err := GetCreationViewUsage(ctx, db, func() time.Time {
		return now
	})
	if err != nil {
		t.Fatal(err)
	}

	// convert into map so we can reliably test for equality
	got := make(map[types.PingName]types.AggregatedPingStats)
	for _, v := range results {
		got[v.Name] = v
	}

	if !cmp.Equal(want, got) {
		t.Fatal(fmt.Sprintf("want: %v got: %v", want, got))
	}
}

func TestFilterSettingJson(t *testing.T) {
	var want map[string]json.RawMessage
	if err := jsonc.Unmarshal(insightAloneSettingStr, &want); err != nil {
		t.Fatal(err)
	}

	input := insightInlineSettingStr
	got, err := FilterSettingJson(input, "searchInsights.")
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("unexpected json map diff:%v", diff)
	}

	for key, val := range got {
		t.Logf("k: %v val: %v", key, val)
	}
}

func TestGetSearchInsights(t *testing.T) {
	db := dbtesting.GetDB(t)
	ctx := context.Background()
	_, err := db.Exec(`INSERT INTO orgs(id, name) VALUES (1, 'first-org'), (2, 'second-org');`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`

			INSERT INTO settings (id, org_id, contents, created_at, user_id, author_user_id)
			VALUES  (1, 1, $1, CURRENT_TIMESTAMP, NULL, NULL),
					(2, 2, $2, CURRENT_TIMESTAMP, NULL, NULL);`,
		insightSettingMulti, insightSettingSimple)

	if err != nil {
		t.Fatal(err)
	}

	step := 2
	want := []SearchInsight{
		{
			ID:           "searchInsights.insight.global.first",
			Title:        "my insight",
			Repositories: []string{"github.com/sourcegraph/sourcegraph"},
			Series: []TimeSeries{{
				Name:   "Redis",
				Stroke: "var(--oc-red-7)",
				Query:  "redis",
			}},
			Step:       Interval{Weeks: &step},
			Visibility: "",
		},
		{
			ID:           "searchInsights.insight.global.second",
			Title:        "my insight",
			Repositories: []string{"github.com/sourcegraph/sourcegraph"},
			Series: []TimeSeries{{
				Name:   "Redis",
				Stroke: "var(--oc-red-7)",
				Query:  "redis",
			}},
			Step:       Interval{Weeks: &step},
			Visibility: "",
		},
		{
			ID:           "searchInsights.insight.global.simple",
			Title:        "my insight",
			Repositories: []string{"github.com/sourcegraph/sourcegraph"},
			Series: []TimeSeries{{
				Name:   "Redis",
				Stroke: "var(--oc-red-7)",
				Query:  "redis",
			}},
			Step:       Interval{Weeks: &step},
			Visibility: "",
		},
	}

	got, err := GetSearchInsights(ctx, db, All)
	if err != nil {
		t.Fatal(err)
	}

	// Sorting the slices so that we can reliably compare them
	sort.Slice(got, func(i, j int) bool {
		return got[i].ID < got[j].ID
	})
	sort.Slice(want, func(i, j int) bool {
		return want[i].ID < want[j].ID

	})

	if !reflect.DeepEqual(want, got) {
		t.Fatalf("unexpected insights diff: %v", cmp.Diff(want, got))
	}
}

func TestGetLangStatsInsights(t *testing.T) {
	db := dbtesting.GetDB(t)
	ctx := context.Background()
	_, err := db.Exec(`INSERT INTO orgs(id, name) VALUES (1, 'first-org'), (2, 'second-org');`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`

			INSERT INTO settings (id, org_id, contents, created_at, user_id, author_user_id)
			VALUES  (1, 1, $1, CURRENT_TIMESTAMP, NULL, NULL)`,
		langStatsInsightSettingStr)
	if err != nil {
		t.Fatal(err)
	}

	want := []LangStatsInsight{
		{
			ID:             "codeStatsInsights.insight.global.lang1",
			Title:          "my insight",
			Repository:     "github.com/sourcegraph/sourcegraph",
			OtherThreshold: float32(0),
		},
	}

	got, err := GetLangStatsInsights(ctx, db, All)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(want, got) {
		t.Fatalf("unexpected insights diff: %v", cmp.Diff(want, got))
	}
}

const insightSettingSimple = `{"searchInsights.insight.global.simple": {
    "title": "my insight",
    "repositories": ["github.com/sourcegraph/sourcegraph"],
    "series": [
      {
        "name": "Redis",
        "query": "redis",
        "stroke": "var(--oc-red-7)"
      }
    ],
    "step": {
      "weeks": 2
    }
  }}`

const insightSettingMulti = `{"searchInsights.insight.global.first": {
    "title": "my insight",
    "repositories": ["github.com/sourcegraph/sourcegraph"],
    "series": [
      {
        "name": "Redis",
        "query": "redis",
        "stroke": "var(--oc-red-7)"
      }
    ],
    "step": {
      "weeks": 2
    }
  },
"searchInsights.insight.global.second": {
    "title": "my insight",
    "repositories": ["github.com/sourcegraph/sourcegraph"],
    "series": [
      {
        "name": "Redis",
        "query": "redis",
        "stroke": "var(--oc-red-7)"
      }
    ],
    "step": {
      "weeks": 2
    }
  }}`

const insightAloneSettingStr = `{"searchInsights.insight.global.myinsight": {
    "title": "my insight",
    "repositories": ["github.com/sourcegraph/sourcegraph"],
    "series": [
      {
        "name": "Redis",
        "query": "redis",
        "stroke": "var(--oc-red-7)"
      }
    ],
    "step": {
      "weeks": 2
    }
  }}`

const insightInlineSettingStr = `{"searchInsights.insight.global.myinsight": {
    "title": "my insight",
    "repositories": ["github.com/sourcegraph/sourcegraph"],
    "series": [
      {
        "name": "Redis",
        "query": "redis",
        "stroke": "var(--oc-red-7)"
      }
    ],
    "step": {
      "weeks": 2
    }
  },
  "codecov.insight.pie": true}`

const langStatsInsightSettingStr = `{"codeStatsInsights.insight.global.lang1": {
    "title": "my insight",
    "repository": "github.com/sourcegraph/sourcegraph",
  }}`
