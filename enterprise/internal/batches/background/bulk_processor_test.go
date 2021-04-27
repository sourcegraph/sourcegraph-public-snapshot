package background

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/sources"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
)

func TestBulkProcessor(t *testing.T) {
	ctx := context.Background()
	db := dbtesting.GetDB(t)
	tx := dbtest.NewTx(t, db)
	bstore := store.New(tx)
	user := ct.CreateTestUser(t, db, true)
	repos, _ := ct.CreateTestRepos(t, ctx, db, 1)
	repo := repos[0]
	changeset := ct.CreateChangeset(t, ctx, bstore, ct.TestChangesetOpts{Repo: repo.ID})

	t.Run("Unknown job type", func(t *testing.T) {
		fake := &sources.FakeChangesetSource{}
		bp := &bulkProcessor{
			store:   bstore,
			sourcer: sources.NewFakeSourcer(nil, fake),
		}
		job := &types.ChangesetJob{
			JobType:     types.ChangesetJobTypeComment,
			ChangesetID: changeset.ID,
			UserID:      user.ID,
		}
		if err := bstore.CreateChangesetJob(ctx, job); err != nil {
			t.Fatal(err)
		}
		job.JobType = types.ChangesetJobType("UNKNOWN")
		err := bp.process(ctx, job)
		if err.Error() != `invalid job type "UNKNOWN"` {
			t.Fatalf("unexpected error returned %s", err)
		}
	})

	t.Run("Comment job", func(t *testing.T) {
		fake := &sources.FakeChangesetSource{}
		bp := &bulkProcessor{
			store:   bstore,
			sourcer: sources.NewFakeSourcer(nil, fake),
		}
		job := &types.ChangesetJob{
			JobType:     types.ChangesetJobTypeComment,
			ChangesetID: changeset.ID,
			UserID:      user.ID,
		}
		if err := bstore.CreateChangesetJob(ctx, job); err != nil {
			t.Fatal(err)
		}
		err := bp.process(ctx, job)
		if err != nil {
			t.Fatal(err)
		}
		if !fake.CreateCommentCalled {
			t.Fatal("expected CreateComment to be called but wasn't")
		}
	})
}
