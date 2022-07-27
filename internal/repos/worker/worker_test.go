package webhookbuilder

import (
	"context"
	"fmt"
	"testing"

	"github.com/hexops/autogold"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestJobQueue(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)
	ctx := actor.WithInternalActor(context.Background())

	mainAppDB := database.NewDB(logger, dbtest.NewDB(logger, t))
	workerBaseStore := basestore.NewWithHandle(mainAppDB.Handle())

	type testCase struct {
		extSvcKind string
	}

	tc := testCase{
		extSvcKind: "GITHUB",
	}

	t.Run(tc.extSvcKind, func(t *testing.T) {
		recordID := 0
		job, err := dequeueJob(ctx, workerBaseStore, recordID)
		autogold.Want("0", (*Job)(nil)).Equal(t, job)
		autogold.Want("1", "expected 1 job to dequeue, found 0").Equal(t, fmt.Sprint(err))

		firstJobID, err := EnqueueJob(ctx, workerBaseStore, &Job{
			RepoID:     1,
			RepoName:   "repo 1",
			ExtSvcKind: tc.extSvcKind,
		})
		if err != nil {
			t.Fatal(err)
		}

		secondJobID, err := EnqueueJob(ctx, workerBaseStore, &Job{
			RepoID:     2,
			RepoName:   "repo 2",
			ExtSvcKind: tc.extSvcKind,
		})
		if err != nil {
			t.Fatal(err)
		}

		firstJob, err := dequeueJob(ctx, workerBaseStore, firstJobID)
		autogold.Want("2", &Job{
			RepoID:     1,
			RepoName:   "repo 1",
			ExtSvcKind: tc.extSvcKind,
		}).Equal(t, firstJob)
		autogold.Want("3", "<nil>").Equal(t, fmt.Sprint(err))

		secondJob, err := dequeueJob(ctx, workerBaseStore, secondJobID)
		autogold.Want("4", &Job{
			RepoID:     2,
			RepoName:   "repo 2",
			ExtSvcKind: tc.extSvcKind,
		}).Equal(t, secondJob)
		autogold.Want("5", "<nil>").Equal(t, fmt.Sprint(err))
	})
}
