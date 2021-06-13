package background

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/sources"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestBulkProcessor(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := dbtest.NewDB(t, "")
	tx := dbtest.NewTx(t, db)
	bstore := store.New(tx, nil)
	user := ct.CreateTestUser(t, db, true)
	repos, _ := ct.CreateTestRepos(t, ctx, db, 1)
	repo := repos[0]
	batchSpec := ct.CreateBatchSpec(t, ctx, bstore, "test-bulk", user.ID)
	batchChange := ct.CreateBatchChange(t, ctx, bstore, "test-bulk", user.ID, batchSpec.ID)
	changeset := ct.CreateChangeset(t, ctx, bstore, ct.TestChangesetOpts{Repo: repo.ID, BatchChanges: []types.BatchChangeAssoc{{BatchChangeID: batchChange.ID}}})

	t.Run("Unknown job type", func(t *testing.T) {
		fake := &sources.FakeChangesetSource{}
		bp := &bulkProcessor{
			tx:      bstore,
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
			tx:      bstore,
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

	t.Run("Detach job", func(t *testing.T) {
		fake := &sources.FakeChangesetSource{}
		bp := &bulkProcessor{
			tx:      bstore,
			sourcer: sources.NewFakeSourcer(nil, fake),
		}
		job := &types.ChangesetJob{
			JobType:       types.ChangesetJobTypeDetach,
			ChangesetID:   changeset.ID,
			UserID:        user.ID,
			BatchChangeID: batchChange.ID,
		}
		if err := bstore.CreateChangesetJob(ctx, job); err != nil {
			t.Fatal(err)
		}
		err := bp.process(ctx, job)
		if err != nil {
			t.Fatal(err)
		}
		ch, err := bstore.GetChangesetByID(ctx, changeset.ID)
		if err != nil {
			t.Fatal(err)
		}
		if len(ch.BatchChanges) != 1 {
			t.Fatalf("invalid batch changes associated, expected one, got=%+v", ch.BatchChanges)
		}
		if !ch.BatchChanges[0].Detach {
			t.Fatal("not marked as to be detached")
		}
		if ch.ReconcilerState != btypes.ReconcilerStateQueued {
			t.Fatalf("invalid reconciler state, got=%q", ch.ReconcilerState)
		}
	})

	t.Run("Reenqueue job", func(t *testing.T) {
		fake := &sources.FakeChangesetSource{}
		bp := &bulkProcessor{
			tx:      bstore,
			sourcer: sources.NewFakeSourcer(nil, fake),
		}
		job := &types.ChangesetJob{
			JobType:     types.ChangesetJobTypeReenqueue,
			ChangesetID: changeset.ID,
			UserID:      user.ID,
		}
		changeset.ReconcilerState = btypes.ReconcilerStateFailed
		if err := bstore.UpdateChangeset(ctx, changeset); err != nil {
			t.Fatal(err)
		}
		if err := bstore.CreateChangesetJob(ctx, job); err != nil {
			t.Fatal(err)
		}
		err := bp.process(ctx, job)
		if err != nil {
			t.Fatal(err)
		}
		changeset, err = bstore.GetChangesetByID(ctx, changeset.ID)
		if err != nil {
			t.Fatal(err)
		}
		if have, want := changeset.ReconcilerState, btypes.ReconcilerStateQueued; have != want {
			t.Fatalf("unexpected reconciler state, have=%q want=%q", have, want)
		}
	})

	t.Run("Merge job", func(t *testing.T) {
		fake := &sources.FakeChangesetSource{}
		bp := &bulkProcessor{
			tx:      bstore,
			sourcer: sources.NewFakeSourcer(nil, fake),
		}
		job := &types.ChangesetJob{
			JobType:     types.ChangesetJobTypeMerge,
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
		if !fake.MergeChangesetCalled {
			t.Fatal("expected MergeChangeset to be called but wasn't")
		}
	})
}
