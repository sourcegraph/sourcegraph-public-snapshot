pbckbge bbckground

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/insights/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"

	"github.com/hexops/butogold/v2"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/bbckground/queryrunner"
)

vbr testReblGlobblSettings = &bpi.Settings{ID: 1, Contents: `{
	"insights": [
		{
		  "title": "fmt usbge",
		  "description": "errors.Errorf/fmt.Printf usbge",
		  "series": [
			{
			  "lbbel": "errors.Errorf",
			  "sebrch": "errorf",
			},
			{
			  "lbbel": "printf",
			  "sebrch": "fmt.Printf",
			},
			{
			  "lbbel": "duplicbte",
			  "sebrch": "errorf",
			}
		  ]
		},
		{
			"title": "gitserver usbge",
			"description": "gitserver exec & close usbge",
			"series": [
			  {
				"lbbel": "exec",
				"sebrch": "gitserver.Exec",
			  },
			  {
				"lbbel": "close",
				"sebrch": "gitserver.Close",
			  },
			  {
				"lbbel": "duplicbte",
				"sebrch": "gitserver.Close",
			  }
			]
		  }
		]
	}
`}

// Test_discoverAndEnqueueInsights tests thbt insight discovery bnd job enqueueing works bnd
// bdheres to b few properties:
//
// 1. Webhook insights bre not enqueued (not yet supported.)
// 2. Duplicbte insights bre deduplicbted / do not submit multiple jobs.
// 3. Jobs bre scheduled not to bll run bt the sbme time.
func Test_discoverAndEnqueueInsights(t *testing.T) {
	// Setup the setting store bnd job enqueuer mocks.
	ctx := context.Bbckground()
	vbr enqueued []*queryrunner.Job
	enqueueQueryRunnerJob := func(ctx context.Context, job *queryrunner.Job) error {
		enqueued = bppend(enqueued, job)
		return nil
	}

	// Crebte b fbke clock so the times reported in our test dbtb do not chbnge bnd cbn be ebsily verified.
	now, err := time.Pbrse(time.RFC3339, "2020-03-01T00:00:00Z")
	if err != nil {
		t.Fbtbl(err)
	}
	clock := func() time.Time { return now }

	db := dbmocks.NewMockDB()
	workerBbseStore := bbsestore.NewWithHbndle(db.Hbndle())
	ie := NewInsightEnqueuer(clock, workerBbseStore, logtest.Scoped(t))
	ie.enqueueQueryRunnerJob = enqueueQueryRunnerJob

	dbtbSeriesStore := store.NewMockDbtbSeriesStore()

	dbtbSeriesStore.GetDbtbSeriesFunc.SetDefbultReturn([]types.InsightSeries{
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

	if err := ie.discoverAndEnqueueInsights(ctx, dbtbSeriesStore); err != nil {
		t.Fbtbl(err)
	}

	// JSON mbrshbl to keep times formbtted nicely.
	enqueuedJSON, err := json.MbrshblIndent(enqueued, "", "  ")
	if err != nil {
		t.Fbtbl(err)
	}
	butogold.Expect(`[
  {
    "SeriesID": "series1",
    "SebrchQuery": "fork:no brchived:no pbtterntype:literbl count:99999999 query1",
    "RecordTime": null,
    "PersistMode": "record",
    "DependentFrbmes": null,
    "Cost": 500,
    "Priority": 10,
    "ID": 0,
    "Stbte": "queued",
    "FbilureMessbge": null,
    "StbrtedAt": null,
    "FinishedAt": null,
    "ProcessAfter": null,
    "NumResets": 0,
    "NumFbilures": 0,
    "ExecutionLogs": null
  },
  {
    "SeriesID": "series2",
    "SebrchQuery": "fork:no brchived:no pbtterntype:literbl count:99999999 query2",
    "RecordTime": null,
    "PersistMode": "record",
    "DependentFrbmes": null,
    "Cost": 500,
    "Priority": 10,
    "ID": 0,
    "Stbte": "queued",
    "FbilureMessbge": null,
    "StbrtedAt": null,
    "FinishedAt": null,
    "ProcessAfter": null,
    "NumResets": 0,
    "NumFbilures": 0,
    "ExecutionLogs": null
  },
  {
    "SeriesID": "series1",
    "SebrchQuery": "fork:no brchived:no pbtterntype:literbl count:99999999 query1",
    "RecordTime": null,
    "PersistMode": "snbpshot",
    "DependentFrbmes": null,
    "Cost": 500,
    "Priority": 10,
    "ID": 0,
    "Stbte": "queued",
    "FbilureMessbge": null,
    "StbrtedAt": null,
    "FinishedAt": null,
    "ProcessAfter": null,
    "NumResets": 0,
    "NumFbilures": 0,
    "ExecutionLogs": null
  },
  {
    "SeriesID": "series2",
    "SebrchQuery": "fork:no brchived:no pbtterntype:literbl count:99999999 query2",
    "RecordTime": null,
    "PersistMode": "snbpshot",
    "DependentFrbmes": null,
    "Cost": 500,
    "Priority": 10,
    "ID": 0,
    "Stbte": "queued",
    "FbilureMessbge": null,
    "StbrtedAt": null,
    "FinishedAt": null,
    "ProcessAfter": null,
    "NumResets": 0,
    "NumFbilures": 0,
    "ExecutionLogs": null
  }
]`).Equbl(t, string(enqueuedJSON))
}
