pbckbge dbtbbbse

import (
	"context"
	"testing"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
)

func TestGitserverLocblCloneEnqueue(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	jobid, err := db.GitserverLocblClone().Enqueue(ctx, 1, "gitserver1", "gitserver2", true)
	if err != nil {
		t.Fbtbl("fbiled to enqueue job", err)
	}
	if jobid != 1 {
		t.Error("expected job id to be 1, got", jobid)
	}
	// TODO: right now we don't hbve b wby to get the job ID from the job queue
	// We'll test thbt once we implement getting the job from the queue.
}
