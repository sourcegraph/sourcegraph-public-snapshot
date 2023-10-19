package workers

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/batches/store"
	bt "github.com/sourcegraph/sourcegraph/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func TestReconcilerWorkerView(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(t))

	now := timeutil.Now()
	clock := func() time.Time { return now }
	bstore := store.NewWithClock(db, &observation.TestContext, nil, clock)

	user := bt.CreateTestUser(t, db, true)
	spec := bt.CreateBatchSpec(t, ctx, bstore, "test-batch-change", user.ID, 0)
	batchChange := bt.CreateBatchChange(t, ctx, bstore, "test-batch-change", user.ID, spec.ID)
	repos, _ := bt.CreateTestRepos(t, ctx, bstore.DatabaseDB(), 2)
	repo := repos[0]
	deletedRepo := repos[1]
	if err := bstore.Repos().Delete(ctx, deletedRepo.ID); err != nil {
		t.Fatal(err)
	}

	t.Run("Queued changeset", func(t *testing.T) {
		c := bt.CreateChangeset(t, ctx, bstore, bt.TestChangesetOpts{
			Repo:            repo.ID,
			BatchChange:     batchChange.ID,
			ReconcilerState: btypes.ReconcilerStateQueued,
		})
		t.Cleanup(func() {
			if err := bstore.DeleteChangeset(ctx, c.ID); err != nil {
				t.Fatal(err)
			}
		})
		assertReturnedChangesetIDs(t, ctx, bstore.DatabaseDB(), []int{int(c.ID)})
	})
	t.Run("Not in batch change", func(t *testing.T) {
		c := bt.CreateChangeset(t, ctx, bstore, bt.TestChangesetOpts{
			Repo:            repo.ID,
			BatchChange:     0,
			ReconcilerState: btypes.ReconcilerStateQueued,
		})
		t.Cleanup(func() {
			if err := bstore.DeleteChangeset(ctx, c.ID); err != nil {
				t.Fatal(err)
			}
		})
		assertReturnedChangesetIDs(t, ctx, bstore.DatabaseDB(), []int{})
	})
	t.Run("In batch change with deleted user namespace", func(t *testing.T) {
		deletedUser := bt.CreateTestUser(t, db, true)
		if err := database.UsersWith(logger, bstore).Delete(ctx, deletedUser.ID); err != nil {
			t.Fatal(err)
		}
		userBatchChange := bt.CreateBatchChange(t, ctx, bstore, "test-user-namespace", deletedUser.ID, spec.ID)
		c := bt.CreateChangeset(t, ctx, bstore, bt.TestChangesetOpts{
			Repo:            repo.ID,
			BatchChange:     userBatchChange.ID,
			ReconcilerState: btypes.ReconcilerStateQueued,
		})
		t.Cleanup(func() {
			if err := bstore.DeleteChangeset(ctx, c.ID); err != nil {
				t.Fatal(err)
			}
		})
		assertReturnedChangesetIDs(t, ctx, bstore.DatabaseDB(), []int{})
	})
	t.Run("In batch change with deleted org namespace", func(t *testing.T) {
		orgID := bt.CreateTestOrg(t, db, "deleted-org").ID
		if err := database.OrgsWith(bstore).Delete(ctx, orgID); err != nil {
			t.Fatal(err)
		}
		orgBatchChange := bt.BuildBatchChange(bstore, "test-user-namespace", 0, spec.ID)
		orgBatchChange.NamespaceOrgID = orgID
		if err := bstore.CreateBatchChange(ctx, orgBatchChange); err != nil {
			t.Fatal(err)
		}
		c := bt.CreateChangeset(t, ctx, bstore, bt.TestChangesetOpts{
			Repo:            repo.ID,
			BatchChange:     orgBatchChange.ID,
			ReconcilerState: btypes.ReconcilerStateQueued,
		})
		t.Cleanup(func() {
			if err := bstore.DeleteChangeset(ctx, c.ID); err != nil {
				t.Fatal(err)
			}
		})
		assertReturnedChangesetIDs(t, ctx, bstore.DatabaseDB(), []int{})
	})
	t.Run("In batch change with deleted namespace but another batch change with an existing one", func(t *testing.T) {
		deletedUser := bt.CreateTestUser(t, db, true)
		if err := database.UsersWith(logger, bstore).Delete(ctx, deletedUser.ID); err != nil {
			t.Fatal(err)
		}
		userBatchChange := bt.CreateBatchChange(t, ctx, bstore, "test-user-namespace", deletedUser.ID, spec.ID)
		c := bt.CreateChangeset(t, ctx, bstore, bt.TestChangesetOpts{
			Repo:            repo.ID,
			BatchChange:     userBatchChange.ID,
			ReconcilerState: btypes.ReconcilerStateQueued,
		})
		// Attach second batch change
		c.Attach(batchChange.ID)
		if err := bstore.UpdateChangeset(ctx, c); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := bstore.DeleteChangeset(ctx, c.ID); err != nil {
				t.Fatal(err)
			}
		})
		assertReturnedChangesetIDs(t, ctx, bstore.DatabaseDB(), []int{int(c.ID)})
	})
	t.Run("In deleted repo", func(t *testing.T) {
		c := bt.CreateChangeset(t, ctx, bstore, bt.TestChangesetOpts{
			Repo:            deletedRepo.ID,
			BatchChange:     batchChange.ID,
			ReconcilerState: btypes.ReconcilerStateQueued,
		})
		t.Cleanup(func() {
			if err := bstore.DeleteChangeset(ctx, c.ID); err != nil {
				t.Fatal(err)
			}
		})
		assertReturnedChangesetIDs(t, ctx, bstore.DatabaseDB(), []int{})
	})
}

func assertReturnedChangesetIDs(t *testing.T, ctx context.Context, db database.DB, want []int) {
	t.Helper()

	have := make([]int, 0)

	q := sqlf.Sprintf("SELECT id FROM reconciler_changesets")
	rows, err := db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		have = append(have, id)
		if err != nil {
			t.Fatal(err)
		}
	}
	if rows.Err() != nil {
		t.Fatal(err)
	}
	if rows.Close() != nil {
		t.Fatal(err)
	}

	sort.Ints(have)
	sort.Ints(want)

	if diff := cmp.Diff(have, want); diff != "" {
		t.Fatalf("invalid IDs returned: diff = %s", diff)
	}
}
