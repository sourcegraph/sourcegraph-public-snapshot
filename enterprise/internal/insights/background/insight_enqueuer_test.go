package background

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"

	"github.com/hexops/autogold"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/background/queryrunner"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

var testRealGlobalSettings = &api.Settings{ID: 1, Contents: `{
	"insights": [
		{
		  "title": "fmt usage",
		  "description": "errors.Errorf/fmt.Printf usage",
		  "series": [
			{
			  "label": "errors.Errorf",
			  "search": "errorf",
			},
			{
			  "label": "printf",
			  "search": "fmt.Printf",
			},
			{
			  "label": "duplicate",
			  "search": "errorf",
			}
		  ]
		},
		{
			"title": "gitserver usage",
			"description": "gitserver exec & close usage",
			"series": [
			  {
				"label": "exec",
				"search": "gitserver.Exec",
			  },
			  {
				"label": "close",
				"search": "gitserver.Close",
			  },
			  {
				"label": "duplicate",
				"search": "gitserver.Close",
			  }
			]
		  }
		]
	}
`}

// Test_discoverAndEnqueueInsights tests that insight discovery and job enqueueing works and
// adheres to a few properties:
//
// 1. Webhook insights are not enqueued (not yet supported.)
// 2. Duplicate insights are deduplicated / do not submit multiple jobs.
// 3. Jobs are scheduled not to all run at the same time.
//
func Test_discoverAndEnqueueInsights(t *testing.T) {
	// Setup the setting store and job enqueuer mocks.
	ctx := context.Background()
	var enqueued []*queryrunner.Job
	enqueueQueryRunnerJob := func(ctx context.Context, job *queryrunner.Job) error {
		enqueued = append(enqueued, job)
		return nil
	}

	// Create a fake clock so the times reported in our test data do not change and can be easily verified.
	now, err := time.Parse(time.RFC3339, "2020-03-01T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	clock := func() time.Time { return now }

	dataSeriesStore := store.NewMockDataSeriesStore()

	dataSeriesStore.GetDataSeriesFunc.SetDefaultReturn([]types.InsightSeries{
		{
			ID:                 1,
			SeriesID:           "series1",
			Query:              "query1",
			NextRecordingAfter: now.Add(-1 * time.Hour),
		},
		{
			ID:                 2,
			SeriesID:           "series2",
			Query:              "query2",
			NextRecordingAfter: now.Add(1 * time.Hour),
		},
	}, nil)

	if err := discoverAndEnqueueInsights(ctx, clock, dataSeriesStore, enqueueQueryRunnerJob); err != nil {
		t.Fatal(err)
	}

	// JSON marshal to keep times formatted nicely.
	enqueuedJSON, err := json.MarshalIndent(enqueued, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	autogold.Want("0", `[
  {
    "SeriesID": "series1",
    "SearchQuery": "query1 count:all",
    "RecordTime": null,
    "Cost": 500,
    "Priority": 10,
    "PersistMode": "record",
    "DependentFrames": null,
    "ID": 0,
    "State": "queued",
    "FailureMessage": null,
    "StartedAt": null,
    "FinishedAt": null,
    "ProcessAfter": null,
    "NumResets": 0,
    "NumFailures": 0,
    "ExecutionLogs": null
  },
  {
    "SeriesID": "series2",
    "SearchQuery": "query2 count:all",
    "RecordTime": null,
    "Cost": 500,
    "Priority": 10,
    "PersistMode": "record",
    "DependentFrames": null,
    "ID": 0,
    "State": "queued",
    "FailureMessage": null,
    "StartedAt": null,
    "FinishedAt": null,
    "ProcessAfter": null,
    "NumResets": 0,
    "NumFailures": 0,
    "ExecutionLogs": null
  },
  {
    "SeriesID": "series1",
    "SearchQuery": "query1 count:all",
    "RecordTime": null,
    "Cost": 500,
    "Priority": 10,
    "PersistMode": "snapshot",
    "DependentFrames": null,
    "ID": 0,
    "State": "queued",
    "FailureMessage": null,
    "StartedAt": null,
    "FinishedAt": null,
    "ProcessAfter": null,
    "NumResets": 0,
    "NumFailures": 0,
    "ExecutionLogs": null
  },
  {
    "SeriesID": "series2",
    "SearchQuery": "query2 count:all",
    "RecordTime": null,
    "Cost": 500,
    "Priority": 10,
    "PersistMode": "snapshot",
    "DependentFrames": null,
    "ID": 0,
    "State": "queued",
    "FailureMessage": null,
    "StartedAt": null,
    "FinishedAt": null,
    "ProcessAfter": null,
    "NumResets": 0,
    "NumFailures": 0,
    "ExecutionLogs": null
  }
]`).Equal(t, string(enqueuedJSON))
}
