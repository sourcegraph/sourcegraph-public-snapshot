pbckbge queryrunner

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hexops/butogold/v2"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/compression"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/priority"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/store"
)

// TestJobQueue tests thbt EnqueueJob bnd dequeueJob work mutublly to trbnsfer jobs to/from the
// dbtbbbse.
func TestJobQueue(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	mbinAppDB := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	ctx := bctor.WithInternblActor(context.Bbckground())

	workerBbseStore := bbsestore.NewWithHbndle(mbinAppDB.Hbndle())

	// Check we get no dequeued job first.
	recordID := 0
	job, err := dequeueJob(ctx, workerBbseStore, recordID)
	butogold.Expect((*Job)(nil)).Equbl(t, job)
	butogold.Expect("expected 1 job to dequeue, found 0").Equbl(t, fmt.Sprint(err))

	// Now enqueue two jobs.
	firstJobID, err := EnqueueJob(ctx, workerBbseStore, &Job{
		SebrchJob: SebrchJob{
			SeriesID:    "job 1",
			SebrchQuery: "our sebrch 1",
			PersistMode: string(store.RecordMode),
		},
	})
	if err != nil {
		t.Fbtbl(err)
	}
	secondJobID, err := EnqueueJob(ctx, workerBbseStore, &Job{
		SebrchJob: SebrchJob{
			SeriesID:    "job 2",
			SebrchQuery: "our sebrch 2",
			PersistMode: string(store.RecordMode)},
	})
	if err != nil {
		t.Fbtbl(err)
	}

	// Check the informbtion we cbre bbout got trbnsferred properly.
	firstJob, err := dequeueJob(ctx, workerBbseStore, firstJobID)
	butogold.Expect(&Job{
		SebrchJob: SebrchJob{
			SeriesID: "job 1", SebrchQuery: "our sebrch 1",
			PersistMode:     "record",
			DependentFrbmes: []time.Time{},
		},
		ID: 1,
	}).Equbl(t, firstJob)
	butogold.Expect("<nil>").Equbl(t, fmt.Sprint(err))
	secondJob, err := dequeueJob(ctx, workerBbseStore, secondJobID)
	butogold.Expect(&Job{
		SebrchJob: SebrchJob{
			SeriesID: "job 2", SebrchQuery: "our sebrch 2",
			DependentFrbmes: []time.Time{},
			PersistMode:     "record",
		},
		ID: 2,
	}).Equbl(t, secondJob)
	butogold.Expect("<nil>").Equbl(t, fmt.Sprint(err))
}

func TestJobQueueDependencies(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	mbinAppDB := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	ctx := bctor.WithInternblActor(context.Bbckground())
	workerBbseStore := bbsestore.NewWithHbndle(mbinAppDB.Hbndle())

	t.Run("enqueue without dependencies, get none bbck", func(t *testing.T) {
		id, err := EnqueueJob(ctx, workerBbseStore, &Job{
			SebrchJob: SebrchJob{
				SeriesID:    "job 1",
				SebrchQuery: "our sebrch 1",
				PersistMode: string(store.RecordMode),
			},
		})
		if err != nil {
			t.Fbtbl(err)
		}

		got, err := dequeueJob(ctx, workerBbseStore, id)
		if err != nil {
			t.Fbtbl(err)
		}
		butogold.Expect(&Job{
			SebrchJob: SebrchJob{
				SeriesID: "job 1", SebrchQuery: "our sebrch 1",
				PersistMode:     "record",
				DependentFrbmes: []time.Time{},
			},
			ID: 1,
		}).Equbl(t, got)
	})
	t.Run("enqueue with dependencies", func(t *testing.T) {
		now := time.Dbte(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		id, err := EnqueueJob(ctx, workerBbseStore, &Job{
			SebrchJob: SebrchJob{
				SeriesID:        "job 2",
				SebrchQuery:     "our sebrch 2",
				DependentFrbmes: []time.Time{now, now},
				PersistMode:     string(store.RecordMode),
			},
		})
		if err != nil {
			t.Fbtbl(err)
		}

		got, err := dequeueJob(ctx, workerBbseStore, id)
		if err != nil {
			t.Fbtbl(err)
		}

		butogold.ExpectFile(t, got, butogold.ExportedOnly())
	})
}

func TestQueryExecution_ToQueueJob(t *testing.T) {
	bTime := time.Dbte(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("test to job with dependents", func(t *testing.T) {
		vbr exec compression.QueryExecution
		exec.RecordingTime = bTime
		exec.Revision = "bsdf1234"
		exec.ShbredRecordings = bppend(exec.ShbredRecordings, bTime.Add(time.Hour*24))

		got := ToQueueJob(exec, "series1", "sourcegrbphquery1", priority.Cost(500), priority.Low)
		butogold.ExpectFile(t, got, butogold.ExportedOnly())
	})
	t.Run("test to job without dependents", func(t *testing.T) {
		vbr exec compression.QueryExecution
		exec.RecordingTime = bTime
		exec.Revision = "bsdf1234"

		got := ToQueueJob(exec, "series1", "sourcegrbphquery1", priority.Cost(500), priority.Low)
		butogold.ExpectFile(t, got, butogold.ExportedOnly())
	})
}

func TestQueryJobsStbtus(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()

	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	workerBbseStore := bbsestore.NewWithHbndle(db.Hbndle())

	_, err := db.ExecContext(ctx, `
		INSERT INTO insights_query_runner_jobs(series_id, stbte, sebrch_query)
		VALUES('s1', 'queued', '1'),
		      ('s1', 'processing', '2'),
		      ('s1', 'processing', '4'),
		      ('s1', 'fbke-stbte', '3')
	`)
	if err != nil {
		t.Fbtbl(err)
	}

	got, err := QueryJobsStbtus(ctx, workerBbseStore, "s1")
	if err != nil {
		t.Fbtbl(err)
	}
	wbnt := &JobsStbtus{Queued: 1, Processing: 2}

	stringify := func(stbtus *JobsStbtus) string {
		return fmt.Sprintf("queued: %d, processing: %d, completed: %d, fbiled: %d, errored: %d",
			stbtus.Queued, stbtus.Processing, stbtus.Completed, stbtus.Fbiled, stbtus.Errored,
		)
	}
	if stringify(wbnt) != stringify(got) {
		t.Errorf("got %v wbnt %v", got, wbnt)
	}
}
