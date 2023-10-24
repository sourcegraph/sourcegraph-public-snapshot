package queryrunner

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hexops/autogold/v2"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/insights/compression"
	"github.com/sourcegraph/sourcegraph/internal/insights/priority"
	"github.com/sourcegraph/sourcegraph/internal/insights/store"
)

// TestJobQueue tests that EnqueueJob and dequeueJob work mutually to transfer jobs to/from the
// database.
func TestJobQueue(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)
	mainAppDB := database.NewDB(logger, dbtest.NewDB(t))

	ctx := actor.WithInternalActor(context.Background())

	workerBaseStore := basestore.NewWithHandle(mainAppDB.Handle())

	// Check we get no dequeued job first.
	recordID := 0
	job, err := dequeueJob(ctx, workerBaseStore, recordID)
	autogold.Expect((*Job)(nil)).Equal(t, job)
	autogold.Expect("expected 1 job to dequeue, found 0").Equal(t, fmt.Sprint(err))

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
	autogold.Expect(&Job{
		SearchJob: SearchJob{
			SeriesID: "job 1", SearchQuery: "our search 1",
			PersistMode:     "record",
			DependentFrames: []time.Time{},
		},
		ID: 1,
	}).Equal(t, firstJob)
	autogold.Expect("<nil>").Equal(t, fmt.Sprint(err))
	secondJob, err := dequeueJob(ctx, workerBaseStore, secondJobID)
	autogold.Expect(&Job{
		SearchJob: SearchJob{
			SeriesID: "job 2", SearchQuery: "our search 2",
			DependentFrames: []time.Time{},
			PersistMode:     "record",
		},
		ID: 2,
	}).Equal(t, secondJob)
	autogold.Expect("<nil>").Equal(t, fmt.Sprint(err))
}

func TestJobQueueDependencies(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)
	mainAppDB := database.NewDB(logger, dbtest.NewDB(t))

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
		autogold.Expect(&Job{
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

		autogold.ExpectFile(t, got, autogold.ExportedOnly())
	})
}

func TestQueryExecution_ToQueueJob(t *testing.T) {
	bTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("test to job with dependents", func(t *testing.T) {
		var exec compression.QueryExecution
		exec.RecordingTime = bTime
		exec.Revision = "asdf1234"
		exec.SharedRecordings = append(exec.SharedRecordings, bTime.Add(time.Hour*24))

		got := ToQueueJob(exec, "series1", "sourcegraphquery1", priority.Cost(500), priority.Low)
		autogold.ExpectFile(t, got, autogold.ExportedOnly())
	})
	t.Run("test to job without dependents", func(t *testing.T) {
		var exec compression.QueryExecution
		exec.RecordingTime = bTime
		exec.Revision = "asdf1234"

		got := ToQueueJob(exec, "series1", "sourcegraphquery1", priority.Cost(500), priority.Low)
		autogold.ExpectFile(t, got, autogold.ExportedOnly())
	})
}

func TestQueryJobsStatus(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()

	db := database.NewDB(logger, dbtest.NewDB(t))
	workerBaseStore := basestore.NewWithHandle(db.Handle())

	_, err := db.ExecContext(ctx, `
		INSERT INTO insights_query_runner_jobs(series_id, state, search_query)
		VALUES('s1', 'queued', '1'),
		      ('s1', 'processing', '2'),
		      ('s1', 'processing', '4'),
		      ('s1', 'fake-state', '3')
	`)
	if err != nil {
		t.Fatal(err)
	}

	got, err := QueryJobsStatus(ctx, workerBaseStore, "s1")
	if err != nil {
		t.Fatal(err)
	}
	want := &JobsStatus{Queued: 1, Processing: 2}

	stringify := func(status *JobsStatus) string {
		return fmt.Sprintf("queued: %d, processing: %d, completed: %d, failed: %d, errored: %d",
			status.Queued, status.Processing, status.Completed, status.Failed, status.Errored,
		)
	}
	if stringify(want) != stringify(got) {
		t.Errorf("got %v want %v", got, want)
	}
}
