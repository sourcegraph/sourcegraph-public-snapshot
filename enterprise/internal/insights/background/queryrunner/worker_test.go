package queryrunner

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"

	"github.com/hexops/autogold"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
)

func init() {
	dbtesting.DBNameSuffix = "codeinsightsbackendqueryrunner"
}

// TestJobQueue tests that EnqueueJob and dequeueJob work mutually to transfer jobs to/from the
// database.
func TestJobQueue(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	//t.Parallel() // TODO: dbtesting.GetDB is not parallel-safe, yuck.

	ctx := actor.WithInternalActor(context.Background())

	mainAppDB := dbtesting.GetDB(t)
	workerBaseStore := basestore.NewWithDB(mainAppDB, sql.TxOptions{})

	// Check we get no dequeued job first.
	recordID := 0
	job, err := dequeueJob(ctx, workerBaseStore, recordID)
	autogold.Want("0", (*Job)(nil)).Equal(t, job)
	autogold.Want("1", "expected 1 job to dequeue, found 0").Equal(t, fmt.Sprint(err))

	// Now enqueue two jobs.
	firstJobID, err := EnqueueJob(ctx, workerBaseStore, &Job{
		SeriesID:    "job 1",
		SearchQuery: "our search 1",
		PersistMode: string(store.RecordMode),
	})
	if err != nil {
		t.Fatal(err)
	}
	secondJobID, err := EnqueueJob(ctx, workerBaseStore, &Job{
		SeriesID:    "job 2",
		SearchQuery: "our search 2",
		PersistMode: string(store.RecordMode),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Check the information we care about got transferred properly.
	firstJob, err := dequeueJob(ctx, workerBaseStore, firstJobID)
	autogold.Want("2", &Job{
		SeriesID: "job 1", SearchQuery: "our search 1",
		PersistMode:     "record",
		DependentFrames: []time.Time{},
		ID:              1,
	}).Equal(t, firstJob)
	autogold.Want("3", "<nil>").Equal(t, fmt.Sprint(err))
	secondJob, err := dequeueJob(ctx, workerBaseStore, secondJobID)
	autogold.Want("4", &Job{
		SeriesID: "job 2", SearchQuery: "our search 2",
		DependentFrames: []time.Time{},
		ID:              2,
		PersistMode:     "record",
	}).Equal(t, secondJob)
	autogold.Want("5", "<nil>").Equal(t, fmt.Sprint(err))
}

func TestJobQueueDependencies(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := actor.WithInternalActor(context.Background())
	mainAppDB := dbtesting.GetDB(t)
	workerBaseStore := basestore.NewWithDB(mainAppDB, sql.TxOptions{})

	t.Run("enqueue without dependencies, get none back", func(t *testing.T) {
		id, err := EnqueueJob(ctx, workerBaseStore, &Job{
			SeriesID:    "job 1",
			SearchQuery: "our search 1",
			PersistMode: string(store.RecordMode),
		})
		if err != nil {
			t.Fatal(err)
		}

		got, err := dequeueJob(ctx, workerBaseStore, id)
		if err != nil {
			t.Fatal(err)
		}
		autogold.Want("1", &Job{
			SeriesID: "job 1", SearchQuery: "our search 1",
			PersistMode:     "record",
			DependentFrames: []time.Time{},
			ID:              1,
		}).Equal(t, got)
	})
	t.Run("enqueue with dependencies", func(t *testing.T) {
		now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		id, err := EnqueueJob(ctx, workerBaseStore, &Job{
			SeriesID:        "job 2",
			SearchQuery:     "our search 2",
			DependentFrames: []time.Time{now, now},
			PersistMode:     string(store.RecordMode),
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
