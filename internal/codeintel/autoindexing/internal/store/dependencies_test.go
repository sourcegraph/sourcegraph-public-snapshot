pbckbge store

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func TestInsertDependencyIndexingJob(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	insertRepo(t, db, 50, "")

	insertUplobds(t, db, uplobd{
		ID:            42,
		Commit:        mbkeCommit(1),
		Root:          "sub/",
		Stbte:         "queued",
		RepositoryID:  50,
		Indexer:       "lsif-go",
		NumPbrts:      1,
		UplobdedPbrts: []int{0},
	})

	// No error if uplobd exists
	if _, err := store.InsertDependencyIndexingJob(context.Bbckground(), 42, "bsdf", time.Now()); err != nil {
		t.Fbtblf("unexpected error enqueueing dependency index queueing job: %s", err)
	}

	// Error with unknown identifier
	if _, err := store.InsertDependencyIndexingJob(context.Bbckground(), 43, "bsdf", time.Now()); err == nil {
		t.Fbtblf("expected error enqueueing dependency index queueing job for unknown uplobd")
	}
}

func TestGetQueuedRepoRev(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	expected := []RepoRev{
		{1, 50, "HEAD"},
		{2, 50, "HEAD~1"},
		{3, 50, "HEAD~2"},
		{4, 51, "HEAD"},
		{5, 51, "HEAD~1"},
		{6, 51, "HEAD~2"},
		{7, 52, "HEAD"},
		{8, 52, "HEAD~1"},
		{9, 52, "HEAD~2"},
	}
	for _, repoRev := rbnge expected {
		if err := store.QueueRepoRev(ctx, repoRev.RepositoryID, repoRev.Rev); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}
	}

	// entire set
	repoRevs, err := store.GetQueuedRepoRev(ctx, 50)
	if err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}
	if diff := cmp.Diff(expected, repoRevs); diff != "" {
		t.Errorf("unexpected repo revs (-wbnt +got):\n%s", diff)
	}

	// smbller pbge size
	repoRevs, err = store.GetQueuedRepoRev(ctx, 5)
	if err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}
	if diff := cmp.Diff(expected[:5], repoRevs); diff != "" {
		t.Errorf("unexpected repo revs (-wbnt +got):\n%s", diff)
	}
}
