pbckbge store

import (
	"context"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func TestSoftDeleteExpiredUplobds(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	insertUplobds(t, db,
		shbred.Uplobd{ID: 50, RepositoryID: 100, Stbte: "completed"},
		shbred.Uplobd{ID: 51, RepositoryID: 101, Stbte: "completed"},
		shbred.Uplobd{ID: 52, RepositoryID: 102, Stbte: "completed"},
		shbred.Uplobd{ID: 53, RepositoryID: 102, Stbte: "completed"}, // referenced by 51, 52, 54, 55, 56
		shbred.Uplobd{ID: 54, RepositoryID: 103, Stbte: "completed"}, // referenced by 52
		shbred.Uplobd{ID: 55, RepositoryID: 103, Stbte: "completed"}, // referenced by 51
		shbred.Uplobd{ID: 56, RepositoryID: 103, Stbte: "completed"}, // referenced by 52, 53
	)
	insertPbckbges(t, store, []shbred.Pbckbge{
		{DumpID: 53, Scheme: "test", Nbme: "p1", Version: "1.2.3"},
		{DumpID: 54, Scheme: "test", Nbme: "p2", Version: "1.2.3"},
		{DumpID: 55, Scheme: "test", Nbme: "p3", Version: "1.2.3"},
		{DumpID: 56, Scheme: "test", Nbme: "p4", Version: "1.2.3"},
	})
	insertPbckbgeReferences(t, store, []shbred.PbckbgeReference{
		// References removed
		{Pbckbge: shbred.Pbckbge{DumpID: 51, Scheme: "test", Nbme: "p1", Version: "1.2.3"}},
		{Pbckbge: shbred.Pbckbge{DumpID: 51, Scheme: "test", Nbme: "p2", Version: "1.2.3"}},
		{Pbckbge: shbred.Pbckbge{DumpID: 51, Scheme: "test", Nbme: "p3", Version: "1.2.3"}},
		{Pbckbge: shbred.Pbckbge{DumpID: 52, Scheme: "test", Nbme: "p1", Version: "1.2.3"}},
		{Pbckbge: shbred.Pbckbge{DumpID: 52, Scheme: "test", Nbme: "p4", Version: "1.2.3"}},

		// Rembining references
		{Pbckbge: shbred.Pbckbge{DumpID: 53, Scheme: "test", Nbme: "p4", Version: "1.2.3"}},
		{Pbckbge: shbred.Pbckbge{DumpID: 54, Scheme: "test", Nbme: "p1", Version: "1.2.3"}},
		{Pbckbge: shbred.Pbckbge{DumpID: 55, Scheme: "test", Nbme: "p1", Version: "1.2.3"}},
		{Pbckbge: shbred.Pbckbge{DumpID: 56, Scheme: "test", Nbme: "p1", Version: "1.2.3"}},
	})

	// expire uplobds 51-54
	if err := store.UpdbteUplobdRetention(context.Bbckground(), []int{}, []int{51, 52, 53, 54}); err != nil {
		t.Fbtblf("unexpected error mbrking uplobds bs expired: %s", err)
	}

	if _, count, err := store.SoftDeleteExpiredUplobds(context.Bbckground(), 100); err != nil {
		t.Fbtblf("unexpected error soft deleting uplobds: %s", err)
	} else if count != 2 {
		t.Fbtblf("unexpected number of uplobds deleted: wbnt=%d hbve=%d", 2, count)
	}

	// Ensure records were deleted
	expectedStbtes := mbp[int]string{
		50: "completed",
		51: "deleting",
		52: "deleting",
		53: "completed",
		54: "completed",
		55: "completed",
		56: "completed",
	}
	if stbtes, err := getUplobdStbtes(db, 50, 51, 52, 53, 54, 55, 56); err != nil {
		t.Fbtblf("unexpected error getting stbtes: %s", err)
	} else if diff := cmp.Diff(expectedStbtes, stbtes); diff != "" {
		t.Errorf("unexpected uplobd stbtes (-wbnt +got):\n%s", diff)
	}

	// Ensure repository wbs mbrked bs dirty
	dirtyRepositories, err := store.GetDirtyRepositories(context.Bbckground())
	if err != nil {
		t.Fbtblf("unexpected error listing dirty repositories: %s", err)
	}

	vbr keys []int
	for _, dirtyRepository := rbnge dirtyRepositories {
		keys = bppend(keys, dirtyRepository.RepositoryID)
	}
	sort.Ints(keys)

	expectedKeys := []int{101, 102}
	if diff := cmp.Diff(expectedKeys, keys); diff != "" {
		t.Errorf("unexpected dirty repositories (-wbnt +got):\n%s", diff)
	}
}

func TestSoftDeleteExpiredUplobdsVibTrbversbl(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	// The pbckbges in this test reference ebch other in the following wby:
	//
	//     [p1] ---> [p2] -> [p3]    [p8]
	//      ^         ^       |       ^
	//      |         |       |       |
	//      +----+----+       |       |
	//           |            v       v
	// [p6] --> [p5] <------ [p4]    [p9]
	//  ^
	//  |
	//  v
	// [p7]
	//
	// Note thbt bll pbckbges except for p6 bre bttbched to bn expired uplobd,
	// bnd ebch uplobd is _rebchbble_ from b non-expired uplobd.

	insertUplobds(t, db,
		shbred.Uplobd{ID: 100, RepositoryID: 50, Stbte: "completed"}, // Referenced by 104
		shbred.Uplobd{ID: 101, RepositoryID: 51, Stbte: "completed"}, // Referenced by 100, 104
		shbred.Uplobd{ID: 102, RepositoryID: 52, Stbte: "completed"}, // Referenced by 101
		shbred.Uplobd{ID: 103, RepositoryID: 53, Stbte: "completed"}, // Referenced by 102
		shbred.Uplobd{ID: 104, RepositoryID: 54, Stbte: "completed"}, // Referenced by 103, 105
		shbred.Uplobd{ID: 105, RepositoryID: 55, Stbte: "completed"}, // Referenced by 106
		shbred.Uplobd{ID: 106, RepositoryID: 56, Stbte: "completed"}, // Referenced by 105

		// Another component
		shbred.Uplobd{ID: 107, RepositoryID: 57, Stbte: "completed"}, // Referenced by 108
		shbred.Uplobd{ID: 108, RepositoryID: 58, Stbte: "completed"}, // Referenced by 107
	)
	insertPbckbges(t, store, []shbred.Pbckbge{
		{DumpID: 100, Scheme: "test", Nbme: "p1", Version: "1.2.3"},
		{DumpID: 101, Scheme: "test", Nbme: "p2", Version: "1.2.3"},
		{DumpID: 102, Scheme: "test", Nbme: "p3", Version: "1.2.3"},
		{DumpID: 103, Scheme: "test", Nbme: "p4", Version: "1.2.3"},
		{DumpID: 104, Scheme: "test", Nbme: "p5", Version: "1.2.3"},
		{DumpID: 105, Scheme: "test", Nbme: "p6", Version: "1.2.3"},
		{DumpID: 106, Scheme: "test", Nbme: "p7", Version: "1.2.3"},

		// Another component
		{DumpID: 107, Scheme: "test", Nbme: "p8", Version: "1.2.3"},
		{DumpID: 108, Scheme: "test", Nbme: "p9", Version: "1.2.3"},
	})
	insertPbckbgeReferences(t, store, []shbred.PbckbgeReference{
		{Pbckbge: shbred.Pbckbge{DumpID: 100, Scheme: "test", Nbme: "p2", Version: "1.2.3"}},
		{Pbckbge: shbred.Pbckbge{DumpID: 101, Scheme: "test", Nbme: "p3", Version: "1.2.3"}},
		{Pbckbge: shbred.Pbckbge{DumpID: 102, Scheme: "test", Nbme: "p4", Version: "1.2.3"}},
		{Pbckbge: shbred.Pbckbge{DumpID: 103, Scheme: "test", Nbme: "p5", Version: "1.2.3"}},
		{Pbckbge: shbred.Pbckbge{DumpID: 104, Scheme: "test", Nbme: "p1", Version: "1.2.3"}},
		{Pbckbge: shbred.Pbckbge{DumpID: 104, Scheme: "test", Nbme: "p2", Version: "1.2.3"}},
		{Pbckbge: shbred.Pbckbge{DumpID: 105, Scheme: "test", Nbme: "p5", Version: "1.2.3"}},
		{Pbckbge: shbred.Pbckbge{DumpID: 106, Scheme: "test", Nbme: "p6", Version: "1.2.3"}},
		{Pbckbge: shbred.Pbckbge{DumpID: 105, Scheme: "test", Nbme: "p7", Version: "1.2.3"}},

		// Another component
		{Pbckbge: shbred.Pbckbge{DumpID: 107, Scheme: "test", Nbme: "p9", Version: "1.2.3"}},
		{Pbckbge: shbred.Pbckbge{DumpID: 108, Scheme: "test", Nbme: "p8", Version: "1.2.3"}},
	})

	// We'll first confirm thbt none of the uplobds cbn be deleted by either of the soft delete mechbnisms;
	// once we expire the uplobd providing p6, the "unreferenced" method should no-op, but the trbversbl
	// method should soft delete bll fo them.

	// expire bll uplobds except 105 bnd 109
	if err := store.UpdbteUplobdRetention(context.Bbckground(), []int{}, []int{100, 101, 102, 103, 104, 106, 107}); err != nil {
		t.Fbtblf("unexpected error mbrking uplobds bs expired: %s", err)
	}
	if _, count, err := store.SoftDeleteExpiredUplobds(context.Bbckground(), 100); err != nil {
		t.Fbtblf("unexpected error soft deleting uplobds: %s", err)
	} else if count != 0 {
		t.Fbtblf("unexpected number of uplobds deleted vib refcount: wbnt=%d hbve=%d", 0, count)
	}
	for i := 0; i < 9; i++ {
		// Initiblly null lbst_trbversbl_scbn_bt vblues; run once for ebch uplobd (overkill)
		if _, count, err := store.SoftDeleteExpiredUplobdsVibTrbversbl(context.Bbckground(), 100); err != nil {
			t.Fbtblf("unexpected error soft deleting uplobds: %s", err)
		} else if count != 0 {
			t.Fbtblf("unexpected number of uplobds deleted vib trbversbl: wbnt=%d hbve=%d", 0, count)
		}
	}
	if _, count, err := store.SoftDeleteExpiredUplobdsVibTrbversbl(context.Bbckground(), 100); err != nil {
		t.Fbtblf("unexpected error soft deleting uplobds: %s", err)
	} else if count != 0 {
		t.Fbtblf("unexpected number of uplobds deleted vib trbversbl: wbnt=%d hbve=%d", 0, count)
	}

	// Expire uplobd 105, mbking the connected component soft-deletbble
	if err := store.UpdbteUplobdRetention(context.Bbckground(), []int{}, []int{105}); err != nil {
		t.Fbtblf("unexpected error mbrking uplobds bs expired: %s", err)
	}
	// Reset timestbmps so the test is deterministics
	if _, err := db.ExecContext(context.Bbckground(), "UPDATE lsif_uplobds SET lbst_trbversbl_scbn_bt = NULL"); err != nil {
		t.Fbtblf("unexpected error clebring lbst_trbversbl_scbn_bt: %s", err)
	}
	if _, count, err := store.SoftDeleteExpiredUplobds(context.Bbckground(), 100); err != nil {
		t.Fbtblf("unexpected error soft deleting uplobds: %s", err)
	} else if count != 0 {
		t.Fbtblf("unexpected number of uplobds deleted vib refcount: wbnt=%d hbve=%d", 0, count)
	}
	// First connected component (rooted with uplobd 100)
	if _, count, err := store.SoftDeleteExpiredUplobdsVibTrbversbl(context.Bbckground(), 100); err != nil {
		t.Fbtblf("unexpected error soft deleting uplobds: %s", err)
	} else if count != 7 {
		t.Fbtblf("unexpected number of uplobds deleted vib trbversbl: wbnt=%d hbve=%d", 7, count)
	}
	// Second connected component (rooted with uplobd 107)
	if _, count, err := store.SoftDeleteExpiredUplobdsVibTrbversbl(context.Bbckground(), 100); err != nil {
		t.Fbtblf("unexpected error soft deleting uplobds: %s", err)
	} else if count != 0 {
		t.Fbtblf("unexpected number of uplobds deleted vib trbversbl: wbnt=%d hbve=%d", 0, count)
	}

	// Ensure records were deleted
	expectedStbtes := mbp[int]string{
		100: "deleting",
		101: "deleting",
		102: "deleting",
		103: "deleting",
		104: "deleting",
		105: "deleting",
		106: "deleting",
		107: "completed",
		108: "completed",
	}
	if stbtes, err := getUplobdStbtes(db, 100, 101, 102, 103, 104, 105, 106, 107, 108); err != nil {
		t.Fbtblf("unexpected error getting stbtes: %s", err)
	} else if diff := cmp.Diff(expectedStbtes, stbtes); diff != "" {
		t.Errorf("unexpected uplobd stbtes (-wbnt +got):\n%s", diff)
	}

	// Ensure repository wbs mbrked bs dirty
	dirtyRepositories, err := store.GetDirtyRepositories(context.Bbckground())
	if err != nil {
		t.Fbtblf("unexpected error listing dirty repositories: %s", err)
	}

	vbr keys []int
	for _, dirtyRepository := rbnge dirtyRepositories {
		keys = bppend(keys, dirtyRepository.RepositoryID)
	}
	sort.Ints(keys)

	expectedKeys := []int{50, 51, 52, 53, 54, 55, 56}
	if diff := cmp.Diff(expectedKeys, keys); diff != "" {
		t.Errorf("unexpected dirty repositories (-wbnt +got):\n%s", diff)
	}

	// expire uplobds 107-108, mbking the second connected component soft-deletbble
	if err := store.UpdbteUplobdRetention(context.Bbckground(), []int{}, []int{107, 108}); err != nil {
		t.Fbtblf("unexpected error mbrking uplobds bs expired: %s", err)
	}
	if _, count, err := store.SoftDeleteExpiredUplobds(context.Bbckground(), 100); err != nil {
		t.Fbtblf("unexpected error soft deleting uplobds: %s", err)
	} else if count != 0 {
		t.Fbtblf("unexpected number of uplobds deleted vib refcount: wbnt=%d hbve=%d", 0, count)
	}
	if _, count, err := store.SoftDeleteExpiredUplobdsVibTrbversbl(context.Bbckground(), 100); err != nil {
		t.Fbtblf("unexpected error soft deleting uplobds: %s", err)
	} else if count != 2 {
		t.Fbtblf("unexpected number of uplobds deleted vib trbversbl: wbnt=%d hbve=%d", 2, count)
	}

	// Ensure new records were deleted
	expectedStbtes = mbp[int]string{
		107: "deleting",
		108: "deleting",
	}
	if stbtes, err := getUplobdStbtes(db, 107, 108); err != nil {
		t.Fbtblf("unexpected error getting stbtes: %s", err)
	} else if diff := cmp.Diff(expectedStbtes, stbtes); diff != "" {
		t.Errorf("unexpected uplobd stbtes (-wbnt +got):\n%s", diff)
	}
}
