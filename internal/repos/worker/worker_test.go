package webhookbuilder

import (
	"context"
	"errors"
	"testing"

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
		if err != nil && err.Error() != "expected 1 job to dequeue, found 0" {
			t.Fatal(err)
		}
		assertEqual(t, nil, nil, job)

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
		assertEqual(t, err, &Job{
			RepoID:     1,
			RepoName:   "repo 1",
			ExtSvcKind: tc.extSvcKind,
		}, firstJob)

		secondJob, err := dequeueJob(ctx, workerBaseStore, secondJobID)
		assertEqual(t, err, &Job{
			RepoID:     2,
			RepoName:   "repo 2",
			ExtSvcKind: tc.extSvcKind,
		}, secondJob)
	})
}

func assertEqual(t *testing.T, err error, want *Job, have *Job) {
	t.Helper()

	if err != nil {
		t.Fatal(err)
	}

	if have == nil {
		if have != nil {
			t.Fatal(errors.New("expected nil job, got non-nil job"))
		}
		return
	}

	if want.RepoID != have.RepoID ||
		want.RepoName != have.RepoName ||
		want.ExtSvcKind != have.ExtSvcKind {
		t.Fatal(errors.New("have, want not the same"))
	}
}
