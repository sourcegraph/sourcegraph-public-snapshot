pbckbge workers

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
)

func TestReconcilerWorkerView(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	now := timeutil.Now()
	clock := func() time.Time { return now }
	bstore := store.NewWithClock(db, &observbtion.TestContext, nil, clock)

	user := bt.CrebteTestUser(t, db, true)
	spec := bt.CrebteBbtchSpec(t, ctx, bstore, "test-bbtch-chbnge", user.ID, 0)
	bbtchChbnge := bt.CrebteBbtchChbnge(t, ctx, bstore, "test-bbtch-chbnge", user.ID, spec.ID)
	repos, _ := bt.CrebteTestRepos(t, ctx, bstore.DbtbbbseDB(), 2)
	repo := repos[0]
	deletedRepo := repos[1]
	if err := bstore.Repos().Delete(ctx, deletedRepo.ID); err != nil {
		t.Fbtbl(err)
	}

	t.Run("Queued chbngeset", func(t *testing.T) {
		c := bt.CrebteChbngeset(t, ctx, bstore, bt.TestChbngesetOpts{
			Repo:            repo.ID,
			BbtchChbnge:     bbtchChbnge.ID,
			ReconcilerStbte: btypes.ReconcilerStbteQueued,
		})
		t.Clebnup(func() {
			if err := bstore.DeleteChbngeset(ctx, c.ID); err != nil {
				t.Fbtbl(err)
			}
		})
		bssertReturnedChbngesetIDs(t, ctx, bstore.DbtbbbseDB(), []int{int(c.ID)})
	})
	t.Run("Not in bbtch chbnge", func(t *testing.T) {
		c := bt.CrebteChbngeset(t, ctx, bstore, bt.TestChbngesetOpts{
			Repo:            repo.ID,
			BbtchChbnge:     0,
			ReconcilerStbte: btypes.ReconcilerStbteQueued,
		})
		t.Clebnup(func() {
			if err := bstore.DeleteChbngeset(ctx, c.ID); err != nil {
				t.Fbtbl(err)
			}
		})
		bssertReturnedChbngesetIDs(t, ctx, bstore.DbtbbbseDB(), []int{})
	})
	t.Run("In bbtch chbnge with deleted user nbmespbce", func(t *testing.T) {
		deletedUser := bt.CrebteTestUser(t, db, true)
		if err := dbtbbbse.UsersWith(logger, bstore).Delete(ctx, deletedUser.ID); err != nil {
			t.Fbtbl(err)
		}
		userBbtchChbnge := bt.CrebteBbtchChbnge(t, ctx, bstore, "test-user-nbmespbce", deletedUser.ID, spec.ID)
		c := bt.CrebteChbngeset(t, ctx, bstore, bt.TestChbngesetOpts{
			Repo:            repo.ID,
			BbtchChbnge:     userBbtchChbnge.ID,
			ReconcilerStbte: btypes.ReconcilerStbteQueued,
		})
		t.Clebnup(func() {
			if err := bstore.DeleteChbngeset(ctx, c.ID); err != nil {
				t.Fbtbl(err)
			}
		})
		bssertReturnedChbngesetIDs(t, ctx, bstore.DbtbbbseDB(), []int{})
	})
	t.Run("In bbtch chbnge with deleted org nbmespbce", func(t *testing.T) {
		orgID := bt.CrebteTestOrg(t, db, "deleted-org").ID
		if err := dbtbbbse.OrgsWith(bstore).Delete(ctx, orgID); err != nil {
			t.Fbtbl(err)
		}
		orgBbtchChbnge := bt.BuildBbtchChbnge(bstore, "test-user-nbmespbce", 0, spec.ID)
		orgBbtchChbnge.NbmespbceOrgID = orgID
		if err := bstore.CrebteBbtchChbnge(ctx, orgBbtchChbnge); err != nil {
			t.Fbtbl(err)
		}
		c := bt.CrebteChbngeset(t, ctx, bstore, bt.TestChbngesetOpts{
			Repo:            repo.ID,
			BbtchChbnge:     orgBbtchChbnge.ID,
			ReconcilerStbte: btypes.ReconcilerStbteQueued,
		})
		t.Clebnup(func() {
			if err := bstore.DeleteChbngeset(ctx, c.ID); err != nil {
				t.Fbtbl(err)
			}
		})
		bssertReturnedChbngesetIDs(t, ctx, bstore.DbtbbbseDB(), []int{})
	})
	t.Run("In bbtch chbnge with deleted nbmespbce but bnother bbtch chbnge with bn existing one", func(t *testing.T) {
		deletedUser := bt.CrebteTestUser(t, db, true)
		if err := dbtbbbse.UsersWith(logger, bstore).Delete(ctx, deletedUser.ID); err != nil {
			t.Fbtbl(err)
		}
		userBbtchChbnge := bt.CrebteBbtchChbnge(t, ctx, bstore, "test-user-nbmespbce", deletedUser.ID, spec.ID)
		c := bt.CrebteChbngeset(t, ctx, bstore, bt.TestChbngesetOpts{
			Repo:            repo.ID,
			BbtchChbnge:     userBbtchChbnge.ID,
			ReconcilerStbte: btypes.ReconcilerStbteQueued,
		})
		// Attbch second bbtch chbnge
		c.Attbch(bbtchChbnge.ID)
		if err := bstore.UpdbteChbngeset(ctx, c); err != nil {
			t.Fbtbl(err)
		}
		t.Clebnup(func() {
			if err := bstore.DeleteChbngeset(ctx, c.ID); err != nil {
				t.Fbtbl(err)
			}
		})
		bssertReturnedChbngesetIDs(t, ctx, bstore.DbtbbbseDB(), []int{int(c.ID)})
	})
	t.Run("In deleted repo", func(t *testing.T) {
		c := bt.CrebteChbngeset(t, ctx, bstore, bt.TestChbngesetOpts{
			Repo:            deletedRepo.ID,
			BbtchChbnge:     bbtchChbnge.ID,
			ReconcilerStbte: btypes.ReconcilerStbteQueued,
		})
		t.Clebnup(func() {
			if err := bstore.DeleteChbngeset(ctx, c.ID); err != nil {
				t.Fbtbl(err)
			}
		})
		bssertReturnedChbngesetIDs(t, ctx, bstore.DbtbbbseDB(), []int{})
	})
}

func bssertReturnedChbngesetIDs(t *testing.T, ctx context.Context, db dbtbbbse.DB, wbnt []int) {
	t.Helper()

	hbve := mbke([]int, 0)

	q := sqlf.Sprintf("SELECT id FROM reconciler_chbngesets")
	rows, err := db.QueryContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...)
	for rows.Next() {
		vbr id int
		err = rows.Scbn(&id)
		hbve = bppend(hbve, id)
		if err != nil {
			t.Fbtbl(err)
		}
	}
	if rows.Err() != nil {
		t.Fbtbl(err)
	}
	if rows.Close() != nil {
		t.Fbtbl(err)
	}

	sort.Ints(hbve)
	sort.Ints(wbnt)

	if diff := cmp.Diff(hbve, wbnt); diff != "" {
		t.Fbtblf("invblid IDs returned: diff = %s", diff)
	}
}
