package database

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestGitserverLocalCloneEnqueue(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	jobid, err := db.GitserverLocalClone().Enqueue(ctx, 1, "gitserver1", "gitserver2", true)
	if err != nil {
		t.Fatal("failed to enqueue job", err)
	}
	if jobid != 1 {
		t.Error("expected job id to be 1, got", jobid)
	}
	// TODO: right now we don't have a way to get the job ID from the job queue
	// We'll test that once we implement getting the job from the queue.
}
