package queryrunner

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hexops/autogold"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/compression"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/priority"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

// TestJobQueue tests that EnqueueJob and dequeueJob work mutually to transfer jobs to/from the
// database.
func TestJobQueue(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)
	mainAppDB := database.NewDB(logger, dbtest.NewDB(logger, t))

	ctx := actor.WithInternalActor(context.Background())

	workerBaseStore := basestore.NewWithHandle(mainAppDB.Handle())

	// Check we get no dequeued job first.
	recordID := 0
	job, err := dequeueJob(ctx, workerBaseStore, recordID)
	autogold.Want("0", (*Job)(nil)).Equal(t, job)
	autogold.Want("1", "expected 1 job to dequeue, found 0").Equal(t, fmt.Sprint(err))

	// Now enqueue two jobs.
	firstJobID, err := EnqueueJob(ctx, workerBaseStore, &Job{
		SearchJob: SearchJob{
			SeriesID:    "job 1",
			SearchQuery: "our search 1",
			PersistMode: string(store.RecordMode),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	secondJobID, err := EnqueueJob(ctx, workerBaseStore, &Job{
		SearchJob: SearchJob{
			SeriesID:    "job 2",
			SearchQuery: "our search 2",
			PersistMode: string(store.RecordMode)},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Check the information we care about got transferred properly.
	firstJob, err := dequeueJob(ctx, workerBaseStore, firstJobID)
	autogold.Want("2", &Job{
		SearchJob: SearchJob{
			SeriesID: "job 1", SearchQuery: "our search 1",
			PersistMode:     "record",
			DependentFrames: []time.Time{},
		},
		ID: 1,
	}).Equal(t, firstJob)
	autogold.Want("3", "<nil>").Equal(t, fmt.Sprint(err))
	secondJob, err := dequeueJob(ctx, workerBaseStore, secondJobID)
	autogold.Want("4", &Job{
		SearchJob: SearchJob{
			SeriesID: "job 2", SearchQuery: "our search 2",
			DependentFrames: []time.Time{},
			PersistMode:     "record",
		},
		ID: 2,
	}).Equal(t, secondJob)
	autogold.Want("5", "<nil>").Equal(t, fmt.Sprint(err))
}

func TestJobQueueDependencies(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)
	mainAppDB := database.NewDB(logger, dbtest.NewDB(logger, t))

	ctx := actor.WithInternalActor(context.Background())
	workerBaseStore := basestore.NewWithHandle(mainAppDB.Handle())

	t.Run("enqueue without dependencies, get none back", func(t *testing.T) {
		id, err := EnqueueJob(ctx, workerBaseStore, &Job{
			SearchJob: SearchJob{
				SeriesID:    "job 1",
				SearchQuery: "our search 1",
				PersistMode: string(store.RecordMode),
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		got, err := dequeueJob(ctx, workerBaseStore, id)
		if err != nil {
			t.Fatal(err)
		}
		autogold.Want("1", &Job{
			SearchJob: SearchJob{
				SeriesID: "job 1", SearchQuery: "our search 1",
				PersistMode:     "record",
				DependentFrames: []time.Time{},
			},
			ID: 1,
		}).Equal(t, got)
	})
	t.Run("enqueue with dependencies", func(t *testing.T) {
		now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		id, err := EnqueueJob(ctx, workerBaseStore, &Job{
			SearchJob: SearchJob{
				SeriesID:        "job 2",
				SearchQuery:     "our search 2",
				DependentFrames: []time.Time{now, now},
				PersistMode:     string(store.RecordMode),
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		got, err := dequeueJob(ctx, workerBaseStore, id)
		if err != nil {
			t.Fatal(err)
		}

		autogold.Equal(t, got, autogold.ExportedOnly())
	})
}

func TestQueryExecution_ToQueueJob(t *testing.T) {
	bTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("test to job with dependents", func(t *testing.T) {
		var exec compression.QueryExecution
		exec.RecordingTime = bTime
		exec.Revision = "asdf1234"
		exec.SharedRecordings = append(exec.SharedRecordings, bTime.Add(time.Hour*24))

		got := ToQueueJob(&exec, "series1", "sourcegraphquery1", priority.Cost(500), priority.Low)
		autogold.Equal(t, got, autogold.ExportedOnly())
	})
	t.Run("test to job without dependents", func(t *testing.T) {
		var exec compression.QueryExecution
		exec.RecordingTime = bTime
		exec.Revision = "asdf1234"

		got := ToQueueJob(&exec, "series1", "sourcegraphquery1", priority.Cost(500), priority.Low)
		autogold.Equal(t, got, autogold.ExportedOnly())
	})
}
